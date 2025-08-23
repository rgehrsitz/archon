package api

import (
	"context"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/git"
	"github.com/rgehrsitz/archon/internal/logging"
	"github.com/rgehrsitz/archon/internal/snapshot"
)

// SnapshotService provides Wails-bound snapshot operations
type SnapshotService struct {
	projectService *ProjectService
}

// NewSnapshotService creates a new snapshot service
func NewSnapshotService(projectService *ProjectService) *SnapshotService {
	return &SnapshotService{
		projectService: projectService,
	}
}

// CreateSnapshotRequest represents the frontend request to create a snapshot
type CreateSnapshotRequest struct {
	Name        string            `json:"name"`        // Required: unique name
	Message     string            `json:"message"`     // Required: commit message  
	Description string            `json:"description"` // Optional: user description
	Labels      map[string]string `json:"labels"`      // Optional: user labels
}

// Create creates a new snapshot with commit + tag pair
func (s *SnapshotService) Create(ctx context.Context, req CreateSnapshotRequest) (*snapshot.Snapshot, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}
	if s.projectService.readOnly {
		return nil, errors.New(errors.ErrSchemaVersion, "Project is opened read-only; writes are disabled")
	}

	manager, err := s.getSnapshotManager()
	if err.Code != "" {
		return nil, err
	}
	defer manager.Close()

	// Convert frontend request to internal request
	createReq := snapshot.CreateRequest{
		Name:        req.Name,
		Message:     req.Message,
		Description: req.Description,
		Labels:      req.Labels,
		// Author will be determined by Git config or defaults
	}

	snap, createErr := manager.Create(ctx, createReq)
	if createErr != nil {
		logging.Log().Error().
			Err(createErr).
			Str("snapshot_name", req.Name).
			Msg("Failed to create snapshot")
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create snapshot", createErr)
	}

	return snap, errors.Envelope{}
}

// List returns all snapshots in chronological order (most recent first)
func (s *SnapshotService) List(ctx context.Context) ([]snapshot.Snapshot, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	manager, err := s.getSnapshotManager()
	if err.Code != "" {
		return nil, err
	}
	defer manager.Close()

	snapshots, listErr := manager.List(ctx)
	if listErr != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to list snapshots", listErr)
	}

	return snapshots, errors.Envelope{}
}

// Get retrieves a specific snapshot by name
func (s *SnapshotService) Get(ctx context.Context, name string) (*snapshot.Snapshot, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	manager, err := s.getSnapshotManager()
	if err.Code != "" {
		return nil, err
	}
	defer manager.Close()

	snap, getErr := manager.Get(ctx, name)
	if getErr != nil {
		return nil, errors.WrapError(errors.ErrNotFound, "Snapshot not found", getErr)
	}

	return snap, errors.Envelope{}
}

// Restore restores the project to a snapshot state
func (s *SnapshotService) Restore(ctx context.Context, name string) errors.Envelope {
	if s.projectService.currentProject == nil {
		return errors.New(errors.ErrNoProject, "No project is currently open")
	}
	if s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only; writes are disabled")
	}

	manager, err := s.getSnapshotManager()
	if err.Code != "" {
		return err
	}
	defer manager.Close()

	if restoreErr := manager.Restore(ctx, name); restoreErr != nil {
		logging.Log().Error().
			Err(restoreErr).
			Str("snapshot_name", name).
			Msg("Failed to restore snapshot")
		return errors.WrapError(errors.ErrStorageFailure, "Failed to restore snapshot", restoreErr)
	}

	return errors.Envelope{}
}

// Delete removes a snapshot (preserves Git tag for history)
func (s *SnapshotService) Delete(ctx context.Context, name string) errors.Envelope {
	if s.projectService.currentProject == nil {
		return errors.New(errors.ErrNoProject, "No project is currently open")
	}
	if s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only; writes are disabled")
	}

	manager, err := s.getSnapshotManager()
	if err.Code != "" {
		return err
	}
	defer manager.Close()

	if deleteErr := manager.Delete(ctx, name); deleteErr != nil {
		return errors.WrapError(errors.ErrStorageFailure, "Failed to delete snapshot", deleteErr)
	}

	return errors.Envelope{}
}

// GetSnapshotHistory returns Git commit history for comparison
func (s *SnapshotService) GetSnapshotHistory(ctx context.Context, limit int) ([]git.Commit, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	repo, err := s.getRepository()
	if err.Code != "" {
		return nil, err
	}
	defer repo.Close()

	return repo.GetCommitHistory(ctx, limit)
}

// CompareSnapshots compares two snapshots (placeholder for semantic diff)
func (s *SnapshotService) CompareSnapshots(ctx context.Context, fromSnapshot, toSnapshot string) (map[string]interface{}, errors.Envelope) {
	if s.projectService.currentProject == nil {
		return nil, errors.New(errors.ErrNoProject, "No project is currently open")
	}

	// Placeholder implementation - will integrate with semantic diff engine
	comparison := map[string]interface{}{
		"from":    fromSnapshot,
		"to":      toSnapshot,
		"changes": []map[string]interface{}{},
		"summary": map[string]interface{}{
			"files_changed": 0,
			"additions":     0,
			"deletions":     0,
		},
	}

	return comparison, errors.Envelope{}
}

// Helper methods

func (s *SnapshotService) getSnapshotManager() (*snapshot.Manager, errors.Envelope) {
	if s.projectService.currentPath == "" {
		return nil, errors.New(errors.ErrInvalidPath, "No current project path")
	}

	manager, err := snapshot.NewManager(s.projectService.currentPath)
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create snapshot manager", err)
	}

	return manager, errors.Envelope{}
}

func (s *SnapshotService) getRepository() (git.Repository, errors.Envelope) {
	if s.projectService.currentPath == "" {
		return nil, errors.New(errors.ErrInvalidPath, "No current project path")
	}

	config := git.RepositoryConfig{
		Path: s.projectService.currentPath,
	}

	repo, err := git.NewRepository(config)
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create repository", err)
	}

	return repo, errors.Envelope{}
}
