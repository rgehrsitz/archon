package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/logging"
)

// Manager handles plugin discovery, loading, and lifecycle management
type Manager struct {
	logger             logging.Logger
	permissionManager  *PermissionManager
	installations      map[string]*PluginInstallation
	pluginsDir         string
}

// NewManager creates a new plugin manager
func NewManager(logger logging.Logger, pluginsDir string) *Manager {
	return &Manager{
		logger:            logger,
		permissionManager: NewPermissionManager(),
		installations:     make(map[string]*PluginInstallation),
		pluginsDir:        pluginsDir,
	}
}

// DiscoverPlugins scans for plugins in the plugins directory
func (m *Manager) DiscoverPlugins(ctx context.Context) ([]*PluginInstallation, errors.Envelope) {
	m.logger.Info().Str("dir", m.pluginsDir).Msg("Discovering plugins in directory")

	if _, err := os.Stat(m.pluginsDir); os.IsNotExist(err) {
		m.logger.Debug().Str("dir", m.pluginsDir).Msg("Plugins directory does not exist, creating")
		if err := os.MkdirAll(m.pluginsDir, 0755); err != nil {
			return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create plugins directory", err)
		}
	}

	var installations []*PluginInstallation
	
	err := filepath.Walk(m.pluginsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for manifest.json files
		if info.Name() == "manifest.json" {
			installation, envelope := m.loadPluginFromPath(filepath.Dir(path))
			if envelope.Code != "" {
				m.logger.Warn().Str("path", path).Str("error", envelope.Message).Msg("Failed to load plugin")
				return nil // Continue walking
			}
			
			if installation != nil {
				installations = append(installations, installation)
				m.installations[installation.Manifest.ID] = installation
			}
		}

		return nil
	})

	if err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to scan plugins directory", err)
	}

	m.logger.Info().Int("count", len(installations)).Msg("Plugin discovery complete")
	return installations, errors.Envelope{}
}

// InstallPlugin installs a plugin from a directory path
func (m *Manager) InstallPlugin(ctx context.Context, sourcePath string) (*PluginInstallation, errors.Envelope) {
	m.logger.Info().Str("path", sourcePath).Msg("Installing plugin")

	// Load and validate the plugin manifest
	installation, envelope := m.loadPluginFromPath(sourcePath)
	if envelope.Code != "" {
		return nil, envelope
	}

	// Check for conflicts with existing plugins
	if existing, exists := m.installations[installation.Manifest.ID]; exists {
		if existing.Enabled {
			return nil, errors.New(errors.ErrValidationFailure, 
				fmt.Sprintf("Plugin %s is already installed and enabled", installation.Manifest.ID))
		}
	}

	// Copy plugin to plugins directory if not already there
	targetPath := filepath.Join(m.pluginsDir, installation.Manifest.ID)
	if sourcePath != targetPath {
		if err := m.copyPlugin(sourcePath, targetPath); err != nil {
			return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to copy plugin", err)
		}
		installation.Path = targetPath
	}

	// Register permissions
	m.permissionManager.DeclarePermissions(installation.Manifest.ID, installation.Manifest.Permissions)

	// Store installation
	installation.Enabled = true
	installation.InstalledAt = time.Now()
	m.installations[installation.Manifest.ID] = installation

	m.logger.Info().Str("id", installation.Manifest.ID).Str("version", installation.Manifest.Version).Msg("Plugin installed successfully")
	return installation, errors.Envelope{}
}

// UninstallPlugin removes a plugin
func (m *Manager) UninstallPlugin(ctx context.Context, pluginID string) errors.Envelope {
	m.logger.Info().Str("id", pluginID).Msg("Uninstalling plugin")

	installation, exists := m.installations[pluginID]
	if !exists {
		return errors.New(errors.ErrNotFound, "Plugin not found: "+pluginID)
	}

	// Remove plugin directory
	if err := os.RemoveAll(installation.Path); err != nil {
		m.logger.Warn().Str("path", installation.Path).Err(err).Msg("Failed to remove plugin directory")
	}

	// Remove from installations
	delete(m.installations, pluginID)

	m.logger.Info().Str("id", pluginID).Msg("Plugin uninstalled successfully")
	return errors.Envelope{}
}

// EnablePlugin enables a plugin
func (m *Manager) EnablePlugin(ctx context.Context, pluginID string) errors.Envelope {
	installation, exists := m.installations[pluginID]
	if !exists {
		return errors.New(errors.ErrNotFound, "Plugin not found: "+pluginID)
	}

	installation.Enabled = true
	m.logger.Info().Str("id", pluginID).Msg("Plugin enabled")
	return errors.Envelope{}
}

// DisablePlugin disables a plugin
func (m *Manager) DisablePlugin(ctx context.Context, pluginID string) errors.Envelope {
	installation, exists := m.installations[pluginID]
	if !exists {
		return errors.New(errors.ErrNotFound, "Plugin not found: "+pluginID)
	}

	installation.Enabled = false
	m.logger.Info().Str("id", pluginID).Msg("Plugin disabled")
	return errors.Envelope{}
}

// GetPlugin returns a plugin by ID
func (m *Manager) GetPlugin(pluginID string) (*PluginInstallation, bool) {
	installation, exists := m.installations[pluginID]
	return installation, exists
}

// GetAllPlugins returns all installed plugins
func (m *Manager) GetAllPlugins() []*PluginInstallation {
	var plugins []*PluginInstallation
	for _, installation := range m.installations {
		plugins = append(plugins, installation)
	}
	return plugins
}

// GetEnabledPlugins returns only enabled plugins
func (m *Manager) GetEnabledPlugins() []*PluginInstallation {
	var plugins []*PluginInstallation
	for _, installation := range m.installations {
		if installation.Enabled {
			plugins = append(plugins, installation)
		}
	}
	return plugins
}

// GetPermissionManager returns the permission manager
func (m *Manager) GetPermissionManager() *PermissionManager {
	return m.permissionManager
}

// ValidateManifest validates a plugin manifest
func (m *Manager) ValidateManifest(manifest *PluginManifest) []string {
	var errors []string

	// Required fields
	if manifest.ID == "" {
		errors = append(errors, "Missing required field: id")
	} else if !m.isValidPluginID(manifest.ID) {
		errors = append(errors, "Invalid plugin ID format")
	}

	if manifest.Name == "" {
		errors = append(errors, "Missing required field: name")
	}

	if manifest.Version == "" {
		errors = append(errors, "Missing required field: version")
	}

	if manifest.Type == "" {
		errors = append(errors, "Missing required field: type")
	} else if !m.isValidPluginType(manifest.Type) {
		errors = append(errors, "Invalid plugin type")
	}

	if manifest.EntryPoint == "" {
		errors = append(errors, "Missing required field: entryPoint")
	}

	// Validate permissions
	permErrors := m.permissionManager.ValidatePermissions(manifest.Permissions)
	errors = append(errors, permErrors...)

	return errors
}

// loadPluginFromPath loads a plugin from a directory path
func (m *Manager) loadPluginFromPath(pluginPath string) (*PluginInstallation, errors.Envelope) {
	manifestPath := filepath.Join(pluginPath, "manifest.json")
	
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, errors.WrapError(errors.ErrNotFound, "Failed to read manifest.json", err)
	}

	var manifest PluginManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, errors.WrapError(errors.ErrValidationFailure, "Invalid manifest.json format", err)
	}

	// Validate manifest
	validationErrors := m.ValidateManifest(&manifest)
	if len(validationErrors) > 0 {
		return nil, errors.New(errors.ErrValidationFailure, 
			"Manifest validation failed: "+strings.Join(validationErrors, ", "))
	}

	installation := &PluginInstallation{
		Manifest:    manifest,
		Path:        pluginPath,
		InstalledAt: time.Now(),
		Enabled:     false,
		Source:      "local",
	}

	return installation, errors.Envelope{}
}

// copyPlugin copies a plugin directory to the target location
func (m *Manager) copyPlugin(src, dst string) error {
	// Simple implementation - in production might want more robust copying
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}

// isValidPluginID validates plugin ID format
func (m *Manager) isValidPluginID(id string) bool {
	// Must be reverse domain notation with lowercase letters, numbers, dots, and hyphens
	if len(id) == 0 || !strings.Contains(id, ".") {
		return false
	}
	
	if strings.HasPrefix(id, ".") || strings.HasSuffix(id, ".") || strings.Contains(id, "..") {
		return false
	}

	for _, char := range id {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '.' || char == '-') {
			return false
		}
	}

	return true
}

// isValidPluginType validates plugin type
func (m *Manager) isValidPluginType(pluginType PluginType) bool {
	switch pluginType {
	case PluginTypeImporter, PluginTypeExporter, PluginTypeTransformer, PluginTypeValidator,
		 PluginTypePanel, PluginTypeProvider, PluginTypeAttachmentProcessor,
		 PluginTypeConflictResolver, PluginTypeSearchIndexer, PluginTypeUIContrib:
		return true
	default:
		return false
	}
}