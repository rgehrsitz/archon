package gogit

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	godiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rgehrsitz/archon/internal/errors"
)

// Repository implements fast read operations using go-git library
type Repository struct {
	path string
	repo *git.Repository
}

// NewRepository creates a new go-git repository instance
func NewRepository(path string) (*Repository, error) {
	// Clean path to ensure it's absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// Open repository
	repo, err := git.PlainOpen(absPath)
	if err != nil {
		return nil, err
	}

	return &Repository{
		path: absPath,
		repo: repo,
	}, nil
}

// Stub implementations for interface compliance - will expand incrementally

func (r *Repository) Status(ctx context.Context) (*Status, errors.Envelope) {
	if r.repo == nil {
		return &Status{IsClean: true}, errors.Envelope{}
	}

	// Determine branch
	branch := ""
	if head, err := r.repo.Head(); err == nil {
		if head.Name().IsBranch() {
			branch = head.Name().Short()
		}
	}

	wt, err := r.repo.Worktree()
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to open worktree", err)
	}
	st, err := wt.Status()
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get worktree status", err)
	}

	var staged, modified, untracked, conflicted []string
	isClean := true
	for path, fs := range st {
		// Untracked (do not mark dirty; Git allows checkout with untracked files if not overwritten)
		if fs.Worktree == git.Untracked {
			untracked = append(untracked, path)
			continue
		}
		// Conflicted
		if fs.Staging == git.UpdatedButUnmerged || fs.Worktree == git.UpdatedButUnmerged {
			conflicted = append(conflicted, path)
			isClean = false
			continue
		}
		// Staged vs modified
		if fs.Staging != git.Unmodified {
			staged = append(staged, path)
			isClean = false
		}
		if fs.Worktree != git.Unmodified {
			modified = append(modified, path)
			isClean = false
		}
	}

	return &Status{
		Branch:          branch,
		IsClean:         isClean,
		AheadBy:         0,
		BehindBy:        0,
		StagedFiles:     staged,
		ModifiedFiles:   modified,
		UntrackedFiles:  untracked,
		ConflictedFiles: conflicted,
	}, errors.Envelope{}
}

func (r *Repository) GetCurrentBranch(ctx context.Context) (string, errors.Envelope) {
	head, err := r.repo.Head()
	if err != nil {
		return "", errors.WrapError(errors.ErrStorageFailure, "Failed to get current branch", err)
	}

	if head.Name().IsBranch() {
		return head.Name().Short(), errors.Envelope{}
	}

	// Detached HEAD - return empty string
	return "", errors.Envelope{}
}

func (r *Repository) GetCommitHistory(ctx context.Context, limit int) ([]Commit, errors.Envelope) {
	if r.repo == nil {
		return []Commit{}, errors.Envelope{}
	}
	if limit <= 0 {
		limit = 100
	}

	head, err := r.repo.Head()
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get HEAD", err)
	}

	iter, err := r.repo.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get commit log", err)
	}
	defer iter.Close()

	commits := make([]Commit, 0, limit)
	count := 0
	for count < limit {
		c, err := iter.Next()
		if err != nil {
			break
		}
		hash := c.Hash.String()
		short := hash
		if len(short) > 8 {
			short = short[:8]
		}
		commits = append(commits, Commit{
			Hash:      hash,
			ShortHash: short,
			Message:   strings.TrimSpace(c.Message),
			Author:    Author{Name: c.Author.Name, Email: c.Author.Email},
		})
		count++
	}
	return commits, errors.Envelope{}
}

func (r *Repository) ListTags(ctx context.Context) ([]Tag, errors.Envelope) {
	if r.repo == nil {
		return []Tag{}, errors.Envelope{}
	}

	iter, err := r.repo.Tags()
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to list tags", err)
	}

	var tags []Tag
	_ = iter.ForEach(func(ref *plumbing.Reference) error {
		if ref == nil {
			return nil
		}
		name := ref.Name().Short()
		isSnapshot := strings.HasPrefix(name, "snapshot-")

		// Default values
		hash := ref.Hash().String()
		var msg string
		var date time.Time

		// Try annotated tag first
		if tagObj, err := r.repo.TagObject(ref.Hash()); err == nil && tagObj != nil {
			msg = tagObj.Message
			// Resolve to the target commit if possible for a stable commit hash
			if commit, err := r.repo.CommitObject(tagObj.Target); err == nil && commit != nil {
				hash = commit.Hash.String()
				date = commit.Author.When
			} else {
				// Fallback to tagger date if commit couldn't be resolved
				date = tagObj.Tagger.When
			}
		} else {
			// Lightweight tag: ref points directly to commit
			if commit, err := r.repo.CommitObject(ref.Hash()); err == nil && commit != nil {
				date = commit.Author.When
			}
		}

		tags = append(tags, Tag{
			Name:       name,
			Hash:       hash,
			Message:    strings.TrimSpace(msg),
			Date:       date,
			IsSnapshot: isSnapshot,
		})
		return nil
	})

	return tags, errors.Envelope{}
}

func (r *Repository) GetDiff(ctx context.Context, from, to string) (*Diff, errors.Envelope) {
	if r.repo == nil {
		return &Diff{From: from, To: to, Files: []FileDiff{}}, errors.Envelope{}
	}

	// Resolve refs to commits
	fromCommit, err := r.resolveToCommit(from)
	if err != nil {
		return nil, errors.WrapError(errors.ErrNotFound, "Failed to resolve 'from' ref", err)
	}
	toCommit, err := r.resolveToCommit(to)
	if err != nil {
		return nil, errors.WrapError(errors.ErrNotFound, "Failed to resolve 'to' ref", err)
	}

	fromTree, err := fromCommit.Tree()
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get 'from' tree", err)
	}
	toTree, err := toCommit.Tree()
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get 'to' tree", err)
	}

	patch, err := fromTree.Patch(toTree)
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to compute patch", err)
	}

	files := make([]FileDiff, 0, len(patch.FilePatches()))
	totalAdds := 0
	totalDels := 0

	for _, fp := range patch.FilePatches() {
		var fromPath, toPath string
		fObj, tObj := fp.Files()
		if fObj != nil {
			fromPath = fObj.Path()
		}
		if tObj != nil {
			toPath = tObj.Path()
		}

		// Determine status
		var status FileStatus
		switch {
		case fObj == nil && tObj != nil:
			status = FileStatusAdded
		case fObj != nil && tObj == nil:
			status = FileStatusDeleted
		default:
			status = FileStatusModified
		}
		// Heuristic rename detection
		oldPath := ""
		if fromPath != "" && toPath != "" && fromPath != toPath {
			oldPath = fromPath
			status = FileStatusRenamed
		}

		// Count additions/deletions from chunks
		adds := 0
		dels := 0
		for _, ch := range fp.Chunks() {
			switch ch.Type() {
			case godiff.Add:
				adds += strings.Count(ch.Content(), "\n")
			case godiff.Delete:
				dels += strings.Count(ch.Content(), "\n")
			}
		}
		totalAdds += adds
		totalDels += dels

		path := toPath
		if status == FileStatusDeleted {
			path = fromPath
		}
		files = append(files, FileDiff{
			Path:      path,
			OldPath:   oldPath,
			Status:    status,
			Additions: adds,
			Deletions: dels,
		})
	}

	diff := &Diff{
		From:  from,
		To:    to,
		Files: files,
		Summary: DiffSummary{
			FilesChanged: len(files),
			Additions:    totalAdds,
			Deletions:    totalDels,
		},
	}
	return diff, errors.Envelope{}
}

func (r *Repository) GetRemoteURL(remote string) (string, errors.Envelope) {
	remotes, err := r.repo.Remotes()
	if err != nil {
		return "", errors.WrapError(errors.ErrStorageFailure, "Failed to get remotes", err)
	}

	for _, rem := range remotes {
		if rem.Config().Name == remote {
			if len(rem.Config().URLs) > 0 {
				return rem.Config().URLs[0], errors.Envelope{}
			}
		}
	}

	return "", errors.New(errors.ErrNotFound, "Remote not found")
}

func (r *Repository) Close() error {
	// No cleanup needed for go-git
	return nil
}

// Basic type definitions to match CLI types (temporary)
type Status struct {
	Branch          string   `json:"branch"`
	IsClean         bool     `json:"isClean"`
	AheadBy         int      `json:"aheadBy"`
	BehindBy        int      `json:"behindBy"`
	StagedFiles     []string `json:"stagedFiles"`
	ModifiedFiles   []string `json:"modifiedFiles"`
	UntrackedFiles  []string `json:"untrackedFiles"`
	ConflictedFiles []string `json:"conflictedFiles"`
}

type Commit struct {
	Hash      string `json:"hash"`
	ShortHash string `json:"shortHash"`
	Message   string `json:"message"`
	Author    Author `json:"author"`
}

type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Tag struct {
	Name       string    `json:"name"`
	Hash       string    `json:"hash"`
	Message    string    `json:"message,omitempty"`
	Date       time.Time `json:"date"`
	IsSnapshot bool      `json:"isSnapshot"`
}

type Diff struct {
	From    string      `json:"from"`
	To      string      `json:"to"`
	Files   []FileDiff  `json:"files"`
	Summary DiffSummary `json:"summary"`
}

type FileDiff struct {
	Path      string     `json:"path"`
	OldPath   string     `json:"oldPath,omitempty"`
	Status    FileStatus `json:"status"`
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
}

type FileStatus string

const (
	FileStatusAdded    FileStatus = "added"
	FileStatusModified FileStatus = "modified"
	FileStatusDeleted  FileStatus = "deleted"
	FileStatusRenamed  FileStatus = "renamed"
	FileStatusCopied   FileStatus = "copied"
)

// DiffSummary mirrors the public git.DiffSummary
type DiffSummary struct {
	FilesChanged int `json:"filesChanged"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
}

// resolveToCommit attempts to resolve a ref (hash, branch, tag) to a commit
func (r *Repository) resolveToCommit(ref string) (*object.Commit, error) {
	if ref == "" {
		// Default to HEAD
		h, err := r.repo.Head()
		if err != nil {
			return nil, err
		}
		return r.repo.CommitObject(h.Hash())
	}

	// Try generic revision resolution first
	if hash, err := r.repo.ResolveRevision(plumbing.Revision(ref)); err == nil && hash != nil {
		if c, err := r.repo.CommitObject(*hash); err == nil {
			return c, nil
		}
		if tagObj, err := r.repo.TagObject(*hash); err == nil && tagObj != nil {
			// Dereference annotated tag
			if c, err := r.repo.CommitObject(tagObj.Target); err == nil {
				return c, nil
			}
		}
	}

	// Try explicit tag name
	// 1) short tag name
	if refIter, err := r.repo.Tags(); err == nil && refIter != nil {
		var found *object.Commit
		_ = refIter.ForEach(func(tagRef *plumbing.Reference) error {
			if tagRef.Name().Short() != ref {
				return nil
			}
			// Resolve tag to commit
			if tagObj, err := r.repo.TagObject(tagRef.Hash()); err == nil && tagObj != nil {
				if c, err := r.repo.CommitObject(tagObj.Target); err == nil {
					found = c
					return nil
				}
			}
			if c, err := r.repo.CommitObject(tagRef.Hash()); err == nil {
				found = c
				return nil
			}
			return nil
		})
		if found != nil {
			return found, nil
		}
	}

	// Try as a full reference name (e.g., refs/tags/<name>)
	if refName := plumbing.ReferenceName(ref); refName.IsTag() || refName.IsBranch() {
		if tagRef, err := r.repo.Reference(refName, true); err == nil && tagRef != nil {
			if c, err := r.repo.CommitObject(tagRef.Hash()); err == nil {
				return c, nil
			}
			if tagObj, err := r.repo.TagObject(tagRef.Hash()); err == nil && tagObj != nil {
				if c, err := r.repo.CommitObject(tagObj.Target); err == nil {
					return c, nil
				}
			}
		}
	}

	// Finally, try treating it as a direct commit hash
	if h := plumbing.NewHash(ref); !h.IsZero() {
		if c, err := r.repo.CommitObject(h); err == nil {
			return c, nil
		}
	}

	return nil, errors.New(errors.ErrNotFound, "Unable to resolve ref to commit")
}
