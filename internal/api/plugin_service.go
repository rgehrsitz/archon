package api

import (
	"context"
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/git"
	"github.com/rgehrsitz/archon/internal/index"
	"github.com/rgehrsitz/archon/internal/logging"
	"github.com/rgehrsitz/archon/internal/plugins"
	"github.com/rgehrsitz/archon/internal/store"
)

// PluginService provides Wails-bound plugin operations
type PluginService struct {
	logger        logging.Logger
	projectService *ProjectService
	pluginManager  *plugins.Manager
	hostService    *plugins.HostService
}

// NewPluginService creates a new plugin service
func NewPluginService(logger logging.Logger, projectService *ProjectService) *PluginService {
	return &PluginService{
		logger:         logger,
		projectService: projectService,
	}
}

// InitializePluginSystem initializes the plugin system with the current project
func (s *PluginService) InitializePluginSystem(ctx context.Context) errors.Envelope {
	s.logger.Info().Msg("Initializing plugin system")

	// Get the current project
	projectStore, projectPath := s.projectService.GetCurrentProject()
	if projectStore == nil {
		return errors.New(errors.ErrValidationFailure, "No project is currently open")
	}

	// Set up plugins directory
	pluginsDir := filepath.Join(projectPath, ".archon", "plugins")

	// Create plugin manager
	s.pluginManager = plugins.NewManager(s.logger, pluginsDir)

	// Get required services from project service
	nodeStore, gitRepo, indexManager, envelope := s.getProjectServices()
	if envelope.Code != "" {
		return envelope
	}

	// Create host service
	s.hostService = plugins.NewHostService(
		s.logger,
		nodeStore,
		gitRepo,
		indexManager,
		s.pluginManager.GetPermissionManager(),
	)

	// Discover existing plugins
	_, envelope = s.pluginManager.DiscoverPlugins(ctx)
	if envelope.Code != "" {
		s.logger.Warn().Str("error", envelope.Message).Msg("Failed to discover plugins")
		// Don't return error - plugin system should still work
	}

	s.logger.Info().Msg("Plugin system initialized successfully")
	return errors.Envelope{}
}

// GetPlugins returns all installed plugins
func (s *PluginService) GetPlugins(ctx context.Context) ([]*plugins.PluginInstallation, errors.Envelope) {
	if err := s.ensureInitialized(); err.Code != "" {
		return nil, err
	}

	plugins := s.pluginManager.GetAllPlugins()
	return plugins, errors.Envelope{}
}

// GetEnabledPlugins returns only enabled plugins
func (s *PluginService) GetEnabledPlugins(ctx context.Context) ([]*plugins.PluginInstallation, errors.Envelope) {
	if err := s.ensureInitialized(); err.Code != "" {
		return nil, err
	}

	enabledPlugins := s.pluginManager.GetEnabledPlugins()
	return enabledPlugins, errors.Envelope{}
}

// InstallPlugin installs a plugin from a directory path
func (s *PluginService) InstallPlugin(ctx context.Context, sourcePath string) (*plugins.PluginInstallation, errors.Envelope) {
	if err := s.ensureInitialized(); err.Code != "" {
		return nil, err
	}

	return s.pluginManager.InstallPlugin(ctx, sourcePath)
}

// UninstallPlugin removes a plugin
func (s *PluginService) UninstallPlugin(ctx context.Context, pluginID string) errors.Envelope {
	if err := s.ensureInitialized(); err.Code != "" {
		return err
	}

	return s.pluginManager.UninstallPlugin(ctx, pluginID)
}

// EnablePlugin enables a plugin
func (s *PluginService) EnablePlugin(ctx context.Context, pluginID string) errors.Envelope {
	if err := s.ensureInitialized(); err.Code != "" {
		return err
	}

	return s.pluginManager.EnablePlugin(ctx, pluginID)
}

// DisablePlugin disables a plugin
func (s *PluginService) DisablePlugin(ctx context.Context, pluginID string) errors.Envelope {
	if err := s.ensureInitialized(); err.Code != "" {
		return err
	}

	return s.pluginManager.DisablePlugin(ctx, pluginID)
}

// GetPluginPermissions returns the permissions for a plugin
func (s *PluginService) GetPluginPermissions(ctx context.Context, pluginID string) ([]*plugins.PluginPermissionGrant, errors.Envelope) {
	if err := s.ensureInitialized(); err.Code != "" {
		return nil, err
	}

	permissions := s.pluginManager.GetPermissionManager().GetGrantedPermissions(pluginID)
	return permissions, errors.Envelope{}
}

// GrantPermission grants a permission to a plugin
func (s *PluginService) GrantPermission(ctx context.Context, pluginID string, permission string, temporary bool, durationMs int) errors.Envelope {
	if err := s.ensureInitialized(); err.Code != "" {
		return err
	}

	// Convert duration from milliseconds to time.Duration
	var duration time.Duration
	if durationMs > 0 {
		duration = time.Duration(durationMs) * time.Millisecond
	}

	return s.pluginManager.GetPermissionManager().GrantPermission(
		pluginID, 
		plugins.Permission(permission), 
		temporary, 
		duration,
	)
}

// RevokePermission revokes a permission from a plugin
func (s *PluginService) RevokePermission(ctx context.Context, pluginID string, permission string) errors.Envelope {
	if err := s.ensureInitialized(); err.Code != "" {
		return err
	}

	s.pluginManager.GetPermissionManager().RevokePermission(pluginID, plugins.Permission(permission))
	return errors.Envelope{}
}

// PluginGetNode allows a plugin to get a node (host service method)
func (s *PluginService) PluginGetNode(ctx context.Context, pluginID string, nodeID string) (*plugins.NodeData, errors.Envelope) {
	if err := s.ensureInitialized(); err.Code != "" {
		return nil, err
	}

	return s.hostService.GetNode(ctx, pluginID, nodeID)
}

// PluginListChildren allows a plugin to list node children
func (s *PluginService) PluginListChildren(ctx context.Context, pluginID string, nodeID string) ([]string, errors.Envelope) {
	if err := s.ensureInitialized(); err.Code != "" {
		return nil, err
	}

	return s.hostService.ListChildren(ctx, pluginID, nodeID)
}

// PluginQuery allows a plugin to search nodes
func (s *PluginService) PluginQuery(ctx context.Context, pluginID string, query string, limit int) ([]*plugins.NodeData, errors.Envelope) {
	if err := s.ensureInitialized(); err.Code != "" {
		return nil, err
	}

	return s.hostService.Query(ctx, pluginID, query, limit)
}

// PluginApplyMutations allows a plugin to apply mutations
func (s *PluginService) PluginApplyMutations(ctx context.Context, pluginID string, mutations []*plugins.Mutation) errors.Envelope {
	if s.projectService != nil && s.projectService.readOnly {
		return errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
	}
	
	if err := s.ensureInitialized(); err.Code != "" {
		return err
	}

	return s.hostService.ApplyMutations(ctx, pluginID, mutations)
}

// PluginCommit allows a plugin to create a commit
func (s *PluginService) PluginCommit(ctx context.Context, pluginID string, message string) (string, errors.Envelope) {
	if s.projectService != nil && s.projectService.readOnly {
		return "", errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
	}
	
	if err := s.ensureInitialized(); err.Code != "" {
		return "", err
	}

	return s.hostService.Commit(ctx, pluginID, message)
}

// PluginSnapshot allows a plugin to create a snapshot
func (s *PluginService) PluginSnapshot(ctx context.Context, pluginID string, message string) (string, errors.Envelope) {
	if s.projectService != nil && s.projectService.readOnly {
		return "", errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
	}
	
	if err := s.ensureInitialized(); err.Code != "" {
		return "", err
	}

	return s.hostService.Snapshot(ctx, pluginID, message)
}

// PluginIndexPut allows a plugin to add content to the search index
func (s *PluginService) PluginIndexPut(ctx context.Context, pluginID string, nodeID string, content string) errors.Envelope {
    if s.projectService != nil && s.projectService.readOnly {
        return errors.New(errors.ErrSchemaVersion, "Project is opened read-only due to newer schema; writes are disabled")
    }
    if err := s.ensureInitialized(); err.Code != "" {
        return err
    }

    return s.hostService.IndexPut(ctx, pluginID, nodeID, content)
}

// ValidatePluginManifest validates a plugin manifest
func (s *PluginService) ValidatePluginManifest(ctx context.Context, manifestData string) ([]string, errors.Envelope) {
	if err := s.ensureInitialized(); err.Code != "" {
		return nil, err
	}

	var manifest plugins.PluginManifest
	if err := json.Unmarshal([]byte(manifestData), &manifest); err != nil {
		return nil, errors.WrapError(errors.ErrValidationFailure, "Invalid JSON format", err)
	}

	validationErrors := s.pluginManager.ValidateManifest(&manifest)
	return validationErrors, errors.Envelope{}
}

// ensureInitialized checks that the plugin system is initialized
func (s *PluginService) ensureInitialized() errors.Envelope {
	if s.pluginManager == nil || s.hostService == nil {
		return errors.New(errors.ErrValidationFailure, "Plugin system not initialized")
	}
	return errors.Envelope{}
}

// getProjectServices retrieves the required services from the current project
func (s *PluginService) getProjectServices() (*store.NodeStore, git.Repository, *index.Manager, errors.Envelope) {
	nodeStore, err := s.projectService.getNodeStore()
	if err.Code != "" {
		return nil, nil, nil, err
	}

	gitRepo, envelope := s.projectService.getGitRepository()
	if envelope.Code != "" {
		return nil, nil, nil, envelope
	}

	indexManager, envelope := s.projectService.getIndexManager()
	if envelope.Code != "" {
		return nil, nil, nil, envelope
	}

	return nodeStore, gitRepo, indexManager, errors.Envelope{}
}