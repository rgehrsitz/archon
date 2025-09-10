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
func (s *ProjectService) OpenProject(path string) *types.Project {
	// Clean and validate path
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		logging.GetLogger().Error().Err(err).Str("path", path).Msg("Invalid project path in OpenProject")
		return nil
	}
	
	logging.GetLogger().Info().Str("path", cleanPath).Msg("Opening project")
	
	// Create project store (index, etc.)
	projectStore, err := store.NewProjectStore(cleanPath)
	if err != nil {
		logging.GetLogger().Error().Err(err).Str("path", cleanPath).Msg("Failed to initialize project store in OpenProject")
		return nil
	}

	// Load project directly to inspect schema without validation blocking migration of legacy versions
	loader := store.NewLoader(cleanPath)
	project, loadErr := loader.LoadProject()
	if loadErr != nil {
		logging.GetLogger().Error().Err(loadErr).Str("path", cleanPath).Msg("Failed to load project in OpenProject")
		return nil
	}

	// Determine read-only vs migration path
	s.readOnly = false
	if project.SchemaVersion > types.CurrentSchemaVersion {
		// Newer project than app: read-only per ADR-007. Validate via store.
		logging.GetLogger().Info().Str("path", cleanPath).Int("project_version", project.SchemaVersion).Int("app_version", types.CurrentSchemaVersion).Msg("Opening newer project in read-only mode")
		s.readOnly = true
		opened, storeErr := projectStore.OpenProject()
		if storeErr != nil {
			logging.GetLogger().Error().Err(storeErr).Str("path", cleanPath).Msg("Failed to open newer project")
			return nil
		}
		project = opened
	} else if project.SchemaVersion < types.CurrentSchemaVersion {
		// Older project: perform backup and run forward migrations before validation
		logging.GetLogger().Info().Str("path", cleanPath).Int("project_version", project.SchemaVersion).Int("app_version", types.CurrentSchemaVersion).Msg("Migrating older project")
		if _, err := migrate.CreateBackup(cleanPath); err != nil {
			logging.GetLogger().Error().Err(err).Str("path", cleanPath).Msg("Failed to create pre-migration backup")
			return nil
		}
		if err := migrate.Run(cleanPath, project.SchemaVersion, types.CurrentSchemaVersion); err != nil {
			logging.GetLogger().Error().Err(err).Str("path", cleanPath).Msg("Migration failed")
			return nil
		}
		// Reload via ProjectStore to validate
		opened, storeErr := projectStore.OpenProject()
		if storeErr != nil {
			logging.GetLogger().Error().Err(storeErr).Str("path", cleanPath).Msg("Failed to reload project post-migration")
			return nil
		}
		project = opened
		// Verify schema version matches target after migration
		if project.SchemaVersion != types.CurrentSchemaVersion {
			logging.GetLogger().Error().Str("path", cleanPath).Int("expected", types.CurrentSchemaVersion).Int("actual", project.SchemaVersion).Msg("Post-migration schema version mismatch")
			return nil
		}
	} else {
		// Equal schema: open normally (validates)
		logging.GetLogger().Info().Str("path", cleanPath).Msg("Opening current version project")
		opened, storeErr := projectStore.OpenProject()
		if storeErr != nil {
			logging.GetLogger().Error().Err(storeErr).Str("path", cleanPath).Msg("Failed to open project")
			return nil
		}
		project = opened
	}

	// Set as current project
	s.currentProject = projectStore
	s.currentPath = cleanPath

	// Initialize logging for this project based on environment; non-fatal if it fails
	if err := logging.InitializeFromEnvironment(cleanPath); err != nil {
		logging.GetLogger().Warn().Err(err).Str("path", cleanPath).Msg("Failed to initialize project-specific logging")
	}
	
	logging.GetLogger().Info().Str("path", cleanPath).Str("rootId", project.RootID).Int("schemaVersion", project.SchemaVersion).Msg("Project opened successfully")

	// After a successful open, ensure index health. If unhealthy, schedule a background rebuild.
	if s.currentProject != nil && s.currentProject.IndexManager != nil && !s.readOnly {
		if err := s.currentProject.IndexManager.Health(); err != nil {
			logging.GetLogger().Warn().Err(err).Msg("Search index unhealthy; scheduling background rebuild")
			go func(ps *ProjectService) {
				idx := NewIndexService(ps)
				if env := idx.Rebuild(context.Background()); env.Code != "" {
					logging.GetLogger().Error().Str("code", env.Code).Str("message", env.Message).Interface("details", env.Details).Msg("Background index rebuild failed")
				} else {
					logging.GetLogger().Info().Msg("Background index rebuild completed successfully")
				}
			}(s)
		}
	}

	return project
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
func (s *ProjectService) ProjectExists(path string) bool {
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		logging.GetLogger().Error().Err(err).Str("path", path).Msg("Invalid project path in ProjectExists")
		return false
	}
	
	projectStore, err := store.NewProjectStore(cleanPath)
	if err != nil {
		logging.GetLogger().Error().Err(err).Str("path", cleanPath).Msg("Failed to initialize project store in ProjectExists")
		return false
	}
	defer projectStore.Close()
	
	exists := projectStore.ProjectExists()
	logging.GetLogger().Debug().Str("path", cleanPath).Bool("exists", exists).Msg("Project exists check")
	return exists
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
