package api

import (
	"context"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/git"
)

// DiffService provides access to repository diffs
type DiffService struct {
	projectService *ProjectService
}

func NewDiffService(projectService *ProjectService) *DiffService {
	return &DiffService{projectService: projectService}
}

// Diff returns a basic file-level diff between two refs (commit hashes or tags)
func (s *DiffService) Diff(ctx context.Context, refA, refB string) (*git.Diff, errors.Envelope) {
	if s.projectService == nil || s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	repo, err := s.getRepository()
	if err.Code != "" {
		return nil, err
	}
	defer repo.Close()

	return repo.GetDiff(ctx, refA, refB)
}

// Helper method to get a repository instance
func (s *DiffService) getRepository() (git.Repository, errors.Envelope) {
	if s.projectService.currentPath == "" {
		return nil, errors.New(errors.ErrInvalidPath, "No current project path")
	}

	config := git.RepositoryConfig{Path: s.projectService.currentPath}
	repo, err := git.NewRepository(config)
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create repository", err)
	}
	return repo, errors.Envelope{}
}
