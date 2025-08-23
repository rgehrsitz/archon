package git

import (
	"context"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
)

// Repository provides Git operations for an Archon project
// Following ADR-010: Hybrid CLI/go-git approach
type Repository interface {
	// Repository management
	IsRepository() bool
	Init(ctx context.Context) errors.Envelope
	GetRemoteURL(remote string) (string, errors.Envelope)
	SetRemoteURL(remote, url string) errors.Envelope
	
	// Status and introspection (prefer go-git for speed)
	Status(ctx context.Context) (*Status, errors.Envelope)
	GetCurrentBranch(ctx context.Context) (string, errors.Envelope)
	GetCommitHistory(ctx context.Context, limit int) ([]Commit, errors.Envelope)
	
	// Porcelain operations (use CLI for correctness)
	Clone(ctx context.Context, url, path string) errors.Envelope
	Fetch(ctx context.Context, remote string) errors.Envelope
	Pull(ctx context.Context, remote, branch string) errors.Envelope
	Push(ctx context.Context, remote, branch string) errors.Envelope
	
	// Commit and tagging
	Add(ctx context.Context, paths []string) errors.Envelope
	Commit(ctx context.Context, message string, author *Author) (*Commit, errors.Envelope)
	CreateTag(ctx context.Context, name, message string) errors.Envelope
	ListTags(ctx context.Context) ([]Tag, errors.Envelope)
	
	// LFS support
	InitLFS(ctx context.Context) errors.Envelope
	IsLFSEnabled(ctx context.Context) (bool, errors.Envelope)
	TrackLFSPattern(ctx context.Context, pattern string) errors.Envelope
	
	// Diff and merge (start simple, expand later for semantic diff)
	GetDiff(ctx context.Context, from, to string) (*Diff, errors.Envelope)
	
	// Cleanup
	Close() error
}

// Status represents the current repository state
type Status struct {
	Branch          string   `json:"branch"`
	IsClean         bool     `json:"isClean"`
	AheadBy         int      `json:"aheadBy"`
	BehindBy        int      `json:"behindBy"`
	StagedFiles     []string `json:"stagedFiles"`
	ModifiedFiles   []string `json:"modifiedFiles"`
	UntrackedFiles  []string `json:"untrackedFiles"`
	ConflictedFiles []string `json:"conflictedFiles"`
	LastCommit      *Commit  `json:"lastCommit,omitempty"`
}

// Commit represents a Git commit
type Commit struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"shortHash"`
	Message   string    `json:"message"`
	Author    Author    `json:"author"`
	Date      time.Time `json:"date"`
	Parents   []string  `json:"parents,omitempty"`
}

// Author represents commit author information
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Tag represents a Git tag
type Tag struct {
	Name       string    `json:"name"`
	Hash       string    `json:"hash"`
	Message    string    `json:"message,omitempty"`
	Date       time.Time `json:"date"`
	IsSnapshot bool      `json:"isSnapshot"` // True for Archon-created snapshot tags
}

// Diff represents changes between two commits
type Diff struct {
	From        string      `json:"from"`
	To          string      `json:"to"`
	Files       []FileDiff  `json:"files"`
	Summary     DiffSummary `json:"summary"`
}

// FileDiff represents changes to a single file
type FileDiff struct {
	Path      string     `json:"path"`
	OldPath   string     `json:"oldPath,omitempty"` // For renames
	Status    FileStatus `json:"status"`            // Added, Modified, Deleted, Renamed
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
}

// FileStatus represents the type of change to a file
type FileStatus string

const (
	FileStatusAdded    FileStatus = "added"
	FileStatusModified FileStatus = "modified"
	FileStatusDeleted  FileStatus = "deleted"
	FileStatusRenamed  FileStatus = "renamed"
	FileStatusCopied   FileStatus = "copied"
)

// DiffSummary provides aggregate statistics about a diff
type DiffSummary struct {
	FilesChanged int `json:"filesChanged"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
}

// Remote represents a Git remote
type Remote struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	FetchURL string `json:"fetchUrl,omitempty"`
	PushURL  string `json:"pushUrl,omitempty"`
}

// RepositoryConfig holds configuration for Git operations
type RepositoryConfig struct {
	// Path to the repository root
	Path string
	
	// Prefer CLI for these operations (default: porcelain operations)
	PreferCLI []string
	
	// Prefer go-git for these operations (default: read operations)
	PreferGoGit []string
	
	// Git executable path (if not in PATH)
	GitPath string
	
	// Default author for commits
	DefaultAuthor *Author
}

// NewRepository creates a new Repository instance for the given path
// Implementation will be router that delegates to CLI or go-git adapters
func NewRepository(config RepositoryConfig) (Repository, error) {
	return newRepositoryRouter(config)
}
