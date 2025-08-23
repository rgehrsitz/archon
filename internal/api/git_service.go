package api

import (
	"context"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/git"
	"github.com/rgehrsitz/archon/internal/logging"
)

// GitService provides Wails-bound Git operations for Archon projects
type GitService struct {
	projectService *ProjectService
}

// NewGitService creates a new Git service
func NewGitService(projectService *ProjectService) *GitService {
	return &GitService{
		projectService: projectService,
	}
}

// Status returns the Git repository status for the current project
func (s *GitService) Status(ctx context.Context) (*git.Status, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	repo, err := s.getRepository()
	if err.Code != "" {
		return nil, err
	}
	defer repo.Close()

	return repo.Status(ctx)
}

// GetCurrentBranch returns the current Git branch
func (s *GitService) GetCurrentBranch(ctx context.Context) (string, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return "", errors.New(errors.ErrNoProject, "No project is currently open")
	}

	repo, err := s.getRepository()
	if err.Code != "" {
		return "", err
	}
	defer repo.Close()

	return repo.GetCurrentBranch(ctx)
}

// InitializeRepository initializes a Git repository for the current project
func (s *GitService) InitializeRepository(ctx context.Context) errors.Envelope {
	if s.projectService.currentProject == nil {
		return errors.New(errors.ErrNoProject, "No project is currently open")
	}
	if s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only; writes are disabled")
	}

	repo, err := s.getRepository()
	if err.Code != "" {
		return err
	}
	defer repo.Close()

	// Initialize Git repository
	initErr := repo.Init(ctx)
	if initErr.Code != "" {
		return initErr
	}

	// Log the initialization
	logging.Log().Info().
		Str("project_path", s.projectService.currentPath).
		Msg("Git repository initialized for project")

	return errors.Envelope{}
}

// SetRemoteURL sets a Git remote URL
func (s *GitService) SetRemoteURL(ctx context.Context, remote, url string) errors.Envelope {
	if s.projectService.currentProject == nil {
		return errors.New(errors.ErrNoProject, "No project is currently open")
	}
	if s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only; writes are disabled")
	}

	repo, err := s.getRepository()
	if err.Code != "" {
		return err
	}
	defer repo.Close()

	return repo.SetRemoteURL(remote, url)
}

// GetRemoteURL gets a Git remote URL
func (s *GitService) GetRemoteURL(ctx context.Context, remote string) (string, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return "", errors.New(errors.ErrNoProject, "No project is currently open")
	}

	repo, err := s.getRepository()
	if err.Code != "" {
		return "", err
	}
	defer repo.Close()

	return repo.GetRemoteURL(remote)
}

// IsRepository checks if the current project is a Git repository
func (s *GitService) IsRepository(ctx context.Context) (bool, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return false, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	return git.IsValidRepository(s.projectService.currentPath), errors.Envelope{}
}

// InitializeLFS initializes Git LFS for the current project
func (s *GitService) InitializeLFS(ctx context.Context) errors.Envelope {
	if s.projectService.currentProject == nil {
		return errors.New(errors.ErrNoProject, "No project is currently open")
	}
	if s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only; writes are disabled")
	}

	repo, err := s.getRepository()
	if err.Code != "" {
		return err
	}
	defer repo.Close()

	return repo.InitLFS(ctx)
}

// TrackLFSPattern adds a pattern to Git LFS tracking
func (s *GitService) TrackLFSPattern(ctx context.Context, pattern string) errors.Envelope {
	if s.projectService.currentProject == nil {
		return errors.New(errors.ErrNoProject, "No project is currently open")
	}
	if s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only; writes are disabled")
	}

	repo, err := s.getRepository()
	if err.Code != "" {
		return err
	}
	defer repo.Close()

	return repo.TrackLFSPattern(ctx, pattern)
}

// Helper method to get a repository instance
func (s *GitService) getRepository() (git.Repository, errors.Envelope) {
	if s.projectService.currentPath == "" {
		return nil, errors.New(errors.ErrInvalidPath, "No current project path")
	}

	config := git.RepositoryConfig{
		Path: s.projectService.currentPath,
		// Use default CLI/go-git preferences for now
	}

	repo, err := git.NewRepository(config)
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create repository", err)
	}

	return repo, errors.Envelope{}
}
