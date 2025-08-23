package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
)

// Repository implements basic Git operations using system CLI
// This is a minimal implementation to establish the foundation
type Repository struct {
	config Config
	gitCmd string
}

// NewRepository creates a new CLI-based Git repository
func NewRepository(config Config) (*Repository, error) {
	gitCmd := "git"
	if config.GitPath != "" {
		gitCmd = config.GitPath
	}

	return &Repository{
		config: config,
		gitCmd: gitCmd,
	}, nil
}

// Basic repository operations for now - we'll expand these incrementally

func (r *Repository) Init(ctx context.Context) errors.Envelope {
	if err := os.MkdirAll(r.config.Path, 0o755); err != nil {
		return r.wrapError(err, errors.ErrStorageFailure, "Failed to create repository directory")
	}

	_, err := r.exec(ctx, "init")
	return r.wrapError(err, errors.ErrStorageFailure, "Failed to initialize Git repository")
}

func (r *Repository) GetRemoteURL(remote string) (string, errors.Envelope) {
	lines, err := r.execLines(context.Background(), "remote", "get-url", remote)
	if err != nil {
		return "", r.wrapError(err, errors.ErrNotFound, "Remote not found")
	}
	if len(lines) == 0 {
		return "", errors.New(errors.ErrNotFound, "Remote URL not found")
	}
	return lines[0], errors.Envelope{}
}

func (r *Repository) SetRemoteURL(remote, url string) errors.Envelope {
	// Try to set existing remote first
	_, err := r.exec(context.Background(), "remote", "set-url", remote, url)
	if err != nil {
		// If remote doesn't exist, add it
		_, err = r.exec(context.Background(), "remote", "add", remote, url)
	}
	return r.wrapError(err, errors.ErrStorageFailure, "Failed to set remote URL")
}

func (r *Repository) IsRepository() bool {
	if r.config.Path == "" {
		return false
	}
	gitDir := filepath.Join(r.config.Path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	// .git can be either a directory (normal repo) or a file (worktree)
	return info.IsDir() || info.Mode().IsRegular()
}

// Placeholder methods that will be implemented incrementally
func (r *Repository) GetCurrentBranch(ctx context.Context) (string, errors.Envelope) {
	lines, err := r.execLines(ctx, "branch", "--show-current")
	if err != nil {
		return "", r.wrapError(err, errors.ErrStorageFailure, "Failed to get current branch")
	}
	if len(lines) == 0 {
		return "", errors.Envelope{} // Detached HEAD or no commits
	}
	return lines[0], errors.Envelope{}
}

func (r *Repository) Clone(ctx context.Context, url, path string) errors.Envelope {
	_, err := r.exec(ctx, "clone", url, path)
	if err == nil {
		// Update our config path
		r.config.Path = path
	}
	return r.wrapError(err, errors.ErrStorageFailure, "Failed to clone repository")
}

func (r *Repository) Fetch(ctx context.Context, remote string) errors.Envelope {
	_, err := r.exec(ctx, "fetch", remote)
	return r.wrapError(err, errors.ErrStorageFailure, "Failed to fetch from remote")
}

func (r *Repository) Pull(ctx context.Context, remote, branch string) errors.Envelope {
	var args []string
	if remote != "" && branch != "" {
		args = []string{"pull", remote, branch}
	} else {
		args = []string{"pull"}
	}
	_, err := r.exec(ctx, args...)
	return r.wrapError(err, errors.ErrStorageFailure, "Failed to pull from remote")
}

func (r *Repository) Push(ctx context.Context, remote, branch string) errors.Envelope {
	var args []string
	if remote != "" && branch != "" {
		args = []string{"push", remote, branch}
	} else {
		args = []string{"push"}
	}
	_, err := r.exec(ctx, args...)
	return r.wrapError(err, errors.ErrStorageFailure, "Failed to push to remote")
}

// Checkout operations
func (r *Repository) Checkout(ctx context.Context, ref string) errors.Envelope {
	_, err := r.exec(ctx, "checkout", ref)
	return r.wrapError(err, errors.ErrStorageFailure, "Failed to checkout reference")
}

// LFS operations
func (r *Repository) InitLFS(ctx context.Context) errors.Envelope {
	_, err := r.exec(ctx, "lfs", "install")
	return r.wrapError(err, errors.ErrStorageFailure, "Failed to initialize Git LFS")
}

func (r *Repository) IsLFSEnabled(ctx context.Context) (bool, errors.Envelope) {
	// Check if LFS is installed and configured
	_, err := r.exec(ctx, "lfs", "env")
	return err == nil, errors.Envelope{}
}

func (r *Repository) TrackLFSPattern(ctx context.Context, pattern string) errors.Envelope {
	_, err := r.exec(ctx, "lfs", "track", pattern)
	if err != nil {
		return r.wrapError(err, errors.ErrStorageFailure, "Failed to track LFS pattern")
	}

	// Add .gitattributes to staging if it was modified
	gitattributesPath := filepath.Join(r.config.Path, ".gitattributes")
	if _, statErr := os.Stat(gitattributesPath); statErr == nil {
		r.exec(ctx, "add", ".gitattributes")
	}

	return errors.Envelope{}
}

// Stub implementations for interface compliance - will expand incrementally

func (r *Repository) Status(ctx context.Context) (*Status, errors.Envelope) {
	// Stub implementation - will be expanded
	return &Status{
		Branch:  "main",
		IsClean: true,
	}, errors.Envelope{}
}

func (r *Repository) GetCommitHistory(ctx context.Context, limit int) ([]Commit, errors.Envelope) {
	// Stub implementation - will be expanded  
	return []Commit{}, errors.Envelope{}
}

func (r *Repository) Add(ctx context.Context, paths []string) errors.Envelope {
	if len(paths) == 0 {
		return errors.New(errors.ErrInvalidInput, "No paths specified")
	}
	args := append([]string{"add"}, paths...)
	_, err := r.exec(ctx, args...)
	return r.wrapError(err, errors.ErrStorageFailure, "Failed to add files")
}

func (r *Repository) Commit(ctx context.Context, message string, author *Author) (*Commit, errors.Envelope) {
	if message == "" {
		return nil, errors.New(errors.ErrInvalidInput, "Commit message cannot be empty")
	}
	
	// Set author if provided
	args := []string{"commit", "-m", message}
	if author != nil && author.Name != "" && author.Email != "" {
		args = append([]string{"commit", "--author", author.Name + " <" + author.Email + ">", "-m", message}, args[2:]...)
	}
	
	_, err := r.exec(ctx, args...)
	if err != nil {
		return nil, r.wrapError(err, errors.ErrStorageFailure, "Failed to create commit")
	}
	
	// Get the latest commit hash
	hashOutput, err := r.exec(ctx, "rev-parse", "HEAD")
	if err != nil {
		return nil, r.wrapError(err, errors.ErrStorageFailure, "Failed to get commit hash")
	}
	
	hash := strings.TrimSpace(string(hashOutput))
	shortHash := hash
	if len(hash) > 8 {
		shortHash = hash[:8]
	}
	
	// Build commit object
	commit := &Commit{
		Hash:      hash,
		ShortHash: shortHash,
		Message:   message,
	}
	
	if author != nil {
		commit.Author = *author
	}
	
	return commit, errors.Envelope{}
}

func (r *Repository) CreateTag(ctx context.Context, name, message string) errors.Envelope {
	var args []string
	if message != "" {
		args = []string{"tag", "-a", name, "-m", message}
	} else {
		args = []string{"tag", name}
	}
	_, err := r.exec(ctx, args...)
	return r.wrapError(err, errors.ErrStorageFailure, "Failed to create tag")
}

func (r *Repository) ListTags(ctx context.Context) ([]Tag, errors.Envelope) {
	fmt.Printf("DEBUG: ListTags called\n")
	// Always use the simpler approach for maximum compatibility
	simpleOutput, simpleErr := r.exec(ctx, "tag", "-l")
	fmt.Printf("DEBUG: exec returned, err: %v\n", simpleErr)
	if simpleErr != nil {
		return nil, r.wrapError(simpleErr, errors.ErrStorageFailure, "Failed to list tags")
	}
	
	outputStr := strings.TrimSpace(string(simpleOutput))
	// DEBUG: Log what we actually got
	fmt.Printf("DEBUG: Git tag output: '%s' (length: %d)\n", outputStr, len(outputStr))
	
	if outputStr == "" {
		fmt.Printf("DEBUG: Output was empty, returning empty slice\n")
		return []Tag{}, errors.Envelope{}
	}
	
	// Create basic tags from simple listing
	var tags []Tag
	lines := strings.Split(outputStr, "\n")
	fmt.Printf("DEBUG: Split into %d lines\n", len(lines))
	for i, line := range lines {
		line = strings.TrimSpace(line)
		fmt.Printf("DEBUG: Line %d: '%s' (length: %d)\n", i, line, len(line))
		if line == "" {
			continue
		}
		tag := Tag{
			Name:       line,
			IsSnapshot: strings.HasPrefix(line, "snapshot-"),
		}
		tags = append(tags, tag)
		fmt.Printf("DEBUG: Added tag: %s (isSnapshot: %v)\n", tag.Name, tag.IsSnapshot)
	}
	fmt.Printf("DEBUG: Returning %d tags\n", len(tags))
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

func (r *Repository) Close() error {
	// No cleanup needed for CLI implementation
	return nil
}

// Basic type definitions to avoid import cycle - these will match the main git types

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