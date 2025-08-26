package api

import (
	"context"
	"path/filepath"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/git"
	"github.com/rgehrsitz/archon/internal/index"
	"github.com/rgehrsitz/archon/internal/logging"
	"github.com/rgehrsitz/archon/internal/migrate"
	"github.com/rgehrsitz/archon/internal/store"
	"github.com/rgehrsitz/archon/internal/types"
)

// ProjectService provides Wails-bound project operations
type ProjectService struct {
	currentProject *store.ProjectStore
	currentPath    string
	readOnly       bool
	ctx            context.Context
}

// NewProjectService creates a new project service
func NewProjectService() *ProjectService {
	return &ProjectService{
		ctx: context.Background(),
	}
}

// SetContext sets the context for the service (called by Wails during initialization)
func (s *ProjectService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// CreateProject creates a new Archon project at the specified path
func (s *ProjectService) CreateProject(path string, settings map[string]any) (*types.Project, errors.Envelope) {
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
	s.readOnly = false

	// Initialize logging for this project based on environment; non-fatal if it fails
	if err := logging.InitializeFromEnvironment(cleanPath); err != nil {
		// Wrap but do not fail project creation; frontend can still operate
		// Optionally: surface as a warning envelope in future
	}
	
	return project, errors.Envelope{}
}

// OpenProject opens an existing Archon project
func (s *ProjectService) OpenProject(path string) (*types.Project, errors.Envelope) {
	// Clean and validate path
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.WrapError(errors.ErrInvalidPath, "Invalid project path", err)
	}
	
	// Create project store (index, etc.)
	projectStore, err := store.NewProjectStore(cleanPath)
	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to initialize project store", err)
	}

	// Load project directly to inspect schema without validation blocking migration of legacy versions
	loader := store.NewLoader(cleanPath)
	project, loadErr := loader.LoadProject()
	if loadErr != nil {
		if envelope, ok := loadErr.(errors.Envelope); ok {
			return nil, envelope
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to load project", loadErr)
	}

	// Determine read-only vs migration path
	s.readOnly = false
	if project.SchemaVersion > types.CurrentSchemaVersion {
		// Newer project than app: read-only per ADR-007. Validate via store.
		s.readOnly = true
		opened, storeErr := projectStore.OpenProject()
		if storeErr != nil {
			if envelope, ok := storeErr.(errors.Envelope); ok {
				return nil, envelope
			}
			return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to open project", storeErr)
		}
		project = opened
	} else if project.SchemaVersion < types.CurrentSchemaVersion {
		// Older project: perform backup and run forward migrations before validation
		if _, err := migrate.CreateBackup(cleanPath); err != nil {
			return nil, errors.WrapError(errors.ErrMigrationFailure, "Failed to create pre-migration backup", err)
		}
		if err := migrate.Run(cleanPath, project.SchemaVersion, types.CurrentSchemaVersion); err != nil {
			return nil, errors.WrapError(errors.ErrMigrationFailure, "Migration failed", err)
		}
		// Reload via ProjectStore to validate
		opened, storeErr := projectStore.OpenProject()
		if storeErr != nil {
			if envelope, ok := storeErr.(errors.Envelope); ok {
				return nil, envelope
			}
			return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to reload project post-migration", storeErr)
		}
		project = opened
		// Verify schema version matches target after migration
		if project.SchemaVersion != types.CurrentSchemaVersion {
			return nil, errors.New(errors.ErrMigrationFailure, "Post-migration schema version mismatch")
		}
	} else {
		// Equal schema: open normally (validates)
		opened, storeErr := projectStore.OpenProject()
		if storeErr != nil {
			if envelope, ok := storeErr.(errors.Envelope); ok {
				return nil, envelope
			}
			return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to open project", storeErr)
		}
		project = opened
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
func (s *ProjectService) CloseProject() errors.Envelope {
	s.currentProject = nil
	s.currentPath = ""
	s.readOnly = false
	return errors.Envelope{}
}

// GetProjectInfo returns information about the current project
func (s *ProjectService) GetProjectInfo() (map[string]any, errors.Envelope) {
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
	info["readOnly"] = s.readOnly
	
	return info, errors.Envelope{}
}

// UpdateProjectSettings updates the current project's settings
func (s *ProjectService) UpdateProjectSettings(settings map[string]any) errors.Envelope {
	if s.currentProject == nil {
		return errors.New(errors.ErrProjectNotFound, "No project currently open")
	}
	if s.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
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
func (s *ProjectService) ProjectExists(path string) (bool, errors.Envelope) {
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
func (s *ProjectService) GetCurrentProjectPath() (string, errors.Envelope) {
	if s.currentProject == nil {
		return "", errors.New(errors.ErrProjectNotFound, "No project currently open")
	}
	
	return s.currentPath, errors.Envelope{}
}

// IsProjectOpen returns true if a project is currently open
func (s *ProjectService) IsProjectOpen() bool {
	return s.currentProject != nil
}

// Helper method to get current project store (for internal use by other services)
func (s *ProjectService) GetCurrentProject() (*store.ProjectStore, string) {
	return s.currentProject, s.currentPath
}

// getNodeStore returns the node store for the current project (for plugin service)
func (s *ProjectService) getNodeStore() (*store.NodeStore, errors.Envelope) {
	if s.currentProject == nil {
		return nil, errors.New(errors.ErrProjectNotFound, "No project currently open")
	}
	
	return store.NewNodeStore(s.currentPath, s.currentProject.IndexManager), errors.Envelope{}
}

// getGitRepository returns the git repository for the current project (for plugin service)
func (s *ProjectService) getGitRepository() (git.Repository, errors.Envelope) {
	if s.currentProject == nil {
		return nil, errors.New(errors.ErrProjectNotFound, "No project currently open")
	}
	
	config := git.RepositoryConfig{
		Path: s.currentPath,
	}
	
	repo, err := git.NewRepository(config)
	if err != nil {
		return nil, errors.WrapError(errors.ErrGitFailure, "Failed to open repository", err)
	}
	
	return repo, errors.Envelope{}
}

// getIndexManager returns the index manager for the current project (for plugin service)
func (s *ProjectService) getIndexManager() (*index.Manager, errors.Envelope) {
	if s.currentProject == nil {
		return nil, errors.New(errors.ErrProjectNotFound, "No project currently open")
	}
	
	return s.currentProject.IndexManager, errors.Envelope{}
}
