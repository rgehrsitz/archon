package api

import (
	"context"
	"path/filepath"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/logging"
	"github.com/rgehrsitz/archon/internal/store"
	"github.com/rgehrsitz/archon/internal/types"
)

// ProjectService provides Wails-bound project operations
type ProjectService struct {
	currentProject *store.ProjectStore
	currentPath    string
}

// NewProjectService creates a new project service
func NewProjectService() *ProjectService {
	return &ProjectService{}
}

// CreateProject creates a new Archon project at the specified path
func (s *ProjectService) CreateProject(ctx context.Context, path string, settings map[string]any) (*types.Project, errors.Envelope) {
	// Clean and validate path
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.WrapError(errors.ErrInvalidPath, "Invalid project path", err)
	}
	
	// Create project store
	projectStore, err := store.NewProjectStore(cleanPath)
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to initialize project store", err)
	}
	
	// Create the project
	project, storeErr := projectStore.CreateProject(settings)
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create project", storeErr)
	}
	
	// Set as current project
	s.currentProject = projectStore
	s.currentPath = cleanPath

	// Initialize logging for this project based on environment; non-fatal if it fails
	if err := logging.InitializeFromEnvironment(cleanPath); err != nil {
		// Wrap but do not fail project creation; frontend can still operate
		// Optionally: surface as a warning envelope in future
	}
	
	return project, errors.Envelope{}
}

// OpenProject opens an existing Archon project
func (s *ProjectService) OpenProject(ctx context.Context, path string) (*types.Project, errors.Envelope) {
	// Clean and validate path
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.WrapError(errors.ErrInvalidPath, "Invalid project path", err)
	}
	
	// Create project store
	projectStore, err := store.NewProjectStore(cleanPath)
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to initialize project store", err)
	}
	
	// Open the project
	project, storeErr := projectStore.OpenProject()
	if storeErr != nil {
		if envelope, ok := storeErr.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to open project", storeErr)
	}
	
	// Set as current project
	s.currentProject = projectStore
	s.currentPath = cleanPath

	// Initialize logging for this project based on environment; non-fatal if it fails
	if err := logging.InitializeFromEnvironment(cleanPath); err != nil {
		// Wrap but do not fail project creation; frontend can still operate
		// Optionally: surface as a warning envelope in future
	}
	
	return project, errors.Envelope{}
}

// CloseProject closes the current project
func (s *ProjectService) CloseProject(ctx context.Context) errors.Envelope {
	s.currentProject = nil
	s.currentPath = ""
	return errors.Envelope{}
}

// GetProjectInfo returns information about the current project
func (s *ProjectService) GetProjectInfo(ctx context.Context) (map[string]any, errors.Envelope) {
	if s.currentProject == nil {
		return nil, errors.New(errors.ErrProjectNotFound, "No project currently open")
	}
	
	info, err := s.currentProject.GetProjectInfo()
	if err != nil {
		if envelope, ok := err.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to get project info", err)
	}
	
	// Add current path to info
	info["currentPath"] = s.currentPath
	
	return info, errors.Envelope{}
}

// UpdateProjectSettings updates the current project's settings
func (s *ProjectService) UpdateProjectSettings(ctx context.Context, settings map[string]any) errors.Envelope {
	if s.currentProject == nil {
		return errors.New(errors.ErrProjectNotFound, "No project currently open")
	}
	
	err := s.currentProject.UpdateProjectSettings(settings)
	if err != nil {
		if envelope, ok := err.(errors.Envelope); ok {
			return envelope
		}
		return errors.WrapError(errors.ErrStorageFailure, "Failed to update project settings", err)
	}
	
	return errors.Envelope{}
}

// ProjectExists checks if a project exists at the given path
func (s *ProjectService) ProjectExists(ctx context.Context, path string) (bool, errors.Envelope) {
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		return false, errors.WrapError(errors.ErrInvalidPath, "Invalid project path", err)
	}
	
	projectStore, err := store.NewProjectStore(cleanPath)
	if err != nil {
		return false, errors.WrapError(errors.ErrStorageFailure, "Failed to initialize project store", err)
	}
	defer projectStore.Close()
	return projectStore.ProjectExists(), errors.Envelope{}
}

// GetCurrentProjectPath returns the path of the currently open project
func (s *ProjectService) GetCurrentProjectPath(ctx context.Context) (string, errors.Envelope) {
	if s.currentProject == nil {
		return "", errors.New(errors.ErrProjectNotFound, "No project currently open")
	}
	
	return s.currentPath, errors.Envelope{}
}

// IsProjectOpen returns true if a project is currently open
func (s *ProjectService) IsProjectOpen(ctx context.Context) bool {
	return s.currentProject != nil
}

// Helper method to get current project store (for internal use by other services)
func (s *ProjectService) GetCurrentProject() (*store.ProjectStore, string) {
	return s.currentProject, s.currentPath
}
