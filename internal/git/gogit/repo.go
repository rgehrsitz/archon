package gogit

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
	// Stub implementation using go-git
	// TODO: Implement using go-git worktree status
	return &Status{
		Branch:  "main",
		IsClean: true,
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
	// Stub implementation - will be expanded
	return []Commit{}, errors.Envelope{}
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
	// Stub implementation - will be expanded
	return &Diff{
		From:  from,
		To:    to,
		Files: []FileDiff{},
	}, errors.Envelope{}
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
	From  string     `json:"from"`
	To    string     `json:"to"`
	Files []FileDiff `json:"files"`
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
