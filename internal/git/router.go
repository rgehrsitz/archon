package git

import (
	"context"
	"os"
	"path/filepath"
	"slices"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/git/cli"
	"github.com/rgehrsitz/archon/internal/git/gogit"
	"github.com/rgehrsitz/archon/internal/logging"
)

// repositoryRouter implements Repository interface using hybrid CLI/go-git approach
type repositoryRouter struct {
	config    RepositoryConfig
	cliRepo   *cli.Repository
	goGitRepo *gogit.Repository
	path      string
}

// newRepositoryRouter creates a new hybrid repository implementation
func newRepositoryRouter(config RepositoryConfig) (Repository, error) {
	if config.Path == "" {
		return nil, errors.New(errors.ErrInvalidInput, "Repository path cannot be empty")
	}

	// Set defaults for operation preferences if not specified
	if config.PreferCLI == nil {
		// Porcelain operations that need credentials, LFS, etc.
		config.PreferCLI = []string{
			"clone", "fetch", "pull", "push", "initlfs", "tracklfs",
		}
	}
	if config.PreferGoGit == nil {
		// Fast read operations
		config.PreferGoGit = []string{
			"status", "history", "branch", "diff", "tags",
		}
	}

	router := &repositoryRouter{
		config: config,
		path:   config.Path,
	}

	// Initialize CLI repository
	cliRepo, err := cli.NewRepository(cli.Config{
		Path:    config.Path,
		GitPath: config.GitPath,
	})
	if err != nil {
		return nil, err
	}
	router.cliRepo = cliRepo

	// Initialize go-git repository if path exists and is a repo
	if IsValidRepository(config.Path) {
		goGitRepo, err := gogit.NewRepository(config.Path)
		if err != nil {
			logging.Log().Warn().
				Err(err).
				Str("path", config.Path).
				Msg("Failed to initialize go-git repository, falling back to CLI only")
		} else {
			router.goGitRepo = goGitRepo
		}
	}

	return router, nil
}

// Helper method to determine which implementation to use for an operation
func (r *repositoryRouter) shouldUseCLI(operation string) bool {
	if slices.Contains(r.config.PreferCLI, operation) {
		return true
	}
	if slices.Contains(r.config.PreferGoGit, operation) {
		return false
	}
	// Default: use CLI for safety
	return true
}

// Repository management methods

func (r *repositoryRouter) IsRepository() bool {
	return IsValidRepository(r.path)
}

func (r *repositoryRouter) Init(ctx context.Context) errors.Envelope {
	env := r.cliRepo.Init(ctx)
	if env.Code == "" {
		// Re-initialize go-git repo after successful init
		if goGitRepo, err := gogit.NewRepository(r.path); err == nil {
			r.goGitRepo = goGitRepo
		}
	}
	return env
}

func (r *repositoryRouter) GetRemoteURL(remote string) (string, errors.Envelope) {
	if r.shouldUseCLI("remote") || r.goGitRepo == nil {
		return r.cliRepo.GetRemoteURL(remote)
	}
	return r.goGitRepo.GetRemoteURL(remote)
}

func (r *repositoryRouter) SetRemoteURL(remote, url string) errors.Envelope {
	return r.cliRepo.SetRemoteURL(remote, url)
}

// Status and introspection methods (prefer go-git for speed)

func (r *repositoryRouter) Status(ctx context.Context) (*Status, errors.Envelope) {
	if r.shouldUseCLI("status") || r.goGitRepo == nil {
		status, env := r.cliRepo.Status(ctx)
		if env.Code != "" {
			return nil, env
		}
		return convertCLIStatus(status), errors.Envelope{}
	}
	status, env := r.goGitRepo.Status(ctx)
	if env.Code != "" {
		return nil, env
	}
	return convertGoGitStatus(status), errors.Envelope{}
}

func (r *repositoryRouter) GetCurrentBranch(ctx context.Context) (string, errors.Envelope) {
	if r.shouldUseCLI("branch") || r.goGitRepo == nil {
		return r.cliRepo.GetCurrentBranch(ctx)
	}
	return r.goGitRepo.GetCurrentBranch(ctx)
}

func (r *repositoryRouter) GetCommitHistory(ctx context.Context, limit int) ([]Commit, errors.Envelope) {
	if r.shouldUseCLI("history") || r.goGitRepo == nil {
		commits, env := r.cliRepo.GetCommitHistory(ctx, limit)
		if env.Code != "" {
			return nil, env
		}
		return convertCLICommits(commits), errors.Envelope{}
	}
	commits, env := r.goGitRepo.GetCommitHistory(ctx, limit)
	if env.Code != "" {
		return nil, env
	}
	return convertGoGitCommits(commits), errors.Envelope{}
}

// Porcelain operations (always use CLI for correctness)

func (r *repositoryRouter) Clone(ctx context.Context, url, path string) errors.Envelope {
	env := r.cliRepo.Clone(ctx, url, path)
	if env.Code == "" {
		// Update our path and re-initialize go-git repo
		r.path = path
		if goGitRepo, err := gogit.NewRepository(path); err == nil {
			r.goGitRepo = goGitRepo
		}
	}
	return env
}

func (r *repositoryRouter) Fetch(ctx context.Context, remote string) errors.Envelope {
	return r.cliRepo.Fetch(ctx, remote)
}

func (r *repositoryRouter) Pull(ctx context.Context, remote, branch string) errors.Envelope {
	return r.cliRepo.Pull(ctx, remote, branch)
}

func (r *repositoryRouter) Push(ctx context.Context, remote, branch string) errors.Envelope {
	return r.cliRepo.Push(ctx, remote, branch)
}

// Commit and tagging methods

func (r *repositoryRouter) Add(ctx context.Context, paths []string) errors.Envelope {
	return r.cliRepo.Add(ctx, paths)
}

func (r *repositoryRouter) Commit(ctx context.Context, message string, author *Author) (*Commit, errors.Envelope) {
	if author == nil && r.config.DefaultAuthor != nil {
		author = r.config.DefaultAuthor
	}
	commit, env := r.cliRepo.Commit(ctx, message, convertAuthorToCLI(author))
	if env.Code != "" {
		return nil, env
	}
	return convertCLICommit(commit), errors.Envelope{}
}

func (r *repositoryRouter) CreateTag(ctx context.Context, name, message string) errors.Envelope {
	return r.cliRepo.CreateTag(ctx, name, message)
}

func (r *repositoryRouter) ListTags(ctx context.Context) ([]Tag, errors.Envelope) {
	if r.shouldUseCLI("tags") || r.goGitRepo == nil {
		tags, env := r.cliRepo.ListTags(ctx)
		if env.Code != "" {
			return nil, env
		}
		return convertCLITags(tags), errors.Envelope{}
	}
	tags, env := r.goGitRepo.ListTags(ctx)
	if env.Code != "" {
		return nil, env
	}
	return convertGoGitTags(tags), errors.Envelope{}
}

// Branch and checkout operations (always use CLI for safety)

func (r *repositoryRouter) Checkout(ctx context.Context, ref string) errors.Envelope {
	return r.cliRepo.Checkout(ctx, ref)
}

// LFS support (always use CLI)

func (r *repositoryRouter) InitLFS(ctx context.Context) errors.Envelope {
	return r.cliRepo.InitLFS(ctx)
}

func (r *repositoryRouter) IsLFSEnabled(ctx context.Context) (bool, errors.Envelope) {
	return r.cliRepo.IsLFSEnabled(ctx)
}

func (r *repositoryRouter) TrackLFSPattern(ctx context.Context, pattern string) errors.Envelope {
	return r.cliRepo.TrackLFSPattern(ctx, pattern)
}

// Diff operations

func (r *repositoryRouter) GetDiff(ctx context.Context, from, to string) (*Diff, errors.Envelope) {
	if r.shouldUseCLI("diff") || r.goGitRepo == nil {
		diff, env := r.cliRepo.GetDiff(ctx, from, to)
		if env.Code != "" {
			return nil, env
		}
		return convertCLIDiff(diff), errors.Envelope{}
	}
	diff, env := r.goGitRepo.GetDiff(ctx, from, to)
	if env.Code != "" {
		return nil, env
	}
	return convertGoGitDiff(diff), errors.Envelope{}
}

// Cleanup

func (r *repositoryRouter) Close() error {
	var err error
	if r.cliRepo != nil {
		if closeErr := r.cliRepo.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if r.goGitRepo != nil {
		if closeErr := r.goGitRepo.Close(); closeErr != nil {
			err = closeErr
		}
	}
	return err
}

// IsValidRepository implementation
func IsValidRepository(path string) bool {
	if path == "" {
		return false
	}
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	// .git can be either a directory (normal repo) or a file (worktree)
	return info.IsDir() || info.Mode().IsRegular()
}
