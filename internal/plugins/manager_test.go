package plugins

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rgehrsitz/archon/internal/logging"
)

func TestPluginManager_ValidateManifest(t *testing.T) {
	logger := logging.NewTestLogger()
	manager := NewManager(logger, "/tmp/test-plugins")

	tests := []struct {
		name      string
		manifest  PluginManifest
		wantValid bool
	}{
		{
			name: "valid manifest",
			manifest: PluginManifest{
				ID:          "com.example.test",
				Name:        "Test Plugin",
				Version:     "1.0.0",
				Type:        PluginTypeImporter,
				EntryPoint:  "index.js",
				Permissions: []Permission{PermissionReadRepo},
			},
			wantValid: true,
		},
		{
			name: "missing ID",
			manifest: PluginManifest{
				Name:        "Test Plugin",
				Version:     "1.0.0",
				Type:        PluginTypeImporter,
				EntryPoint:  "index.js",
				Permissions: []Permission{PermissionReadRepo},
			},
			wantValid: false,
		},
		{
			name: "invalid plugin type",
			manifest: PluginManifest{
				ID:          "com.example.test",
				Name:        "Test Plugin", 
				Version:     "1.0.0",
				Type:        PluginType("InvalidType"),
				EntryPoint:  "index.js",
				Permissions: []Permission{PermissionReadRepo},
			},
			wantValid: false,
		},
		{
			name: "invalid permission",
			manifest: PluginManifest{
				ID:          "com.example.test",
				Name:        "Test Plugin",
				Version:     "1.0.0", 
				Type:        PluginTypeImporter,
				EntryPoint:  "index.js",
				Permissions: []Permission{"invalidPermission"},
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := manager.ValidateManifest(&tt.manifest)
			valid := len(errors) == 0
			
			if valid != tt.wantValid {
				t.Errorf("ValidateManifest() valid = %v, want %v, errors = %v", valid, tt.wantValid, errors)
			}
		})
	}
}

func TestPluginManager_DiscoverPlugins(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	logger := logging.NewTestLogger()
	manager := NewManager(logger, tmpDir)

	// Create a test plugin
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("Failed to create plugin directory: %v", err)
	}

	manifest := PluginManifest{
		ID:          "com.test.plugin",
		Name:        "Test Plugin",
		Version:     "1.0.0",
		Type:        PluginTypeImporter,
		EntryPoint:  "index.js",
		Permissions: []Permission{PermissionReadRepo},
	}

	manifestData, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}

	manifestPath := filepath.Join(pluginDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Discover plugins
	ctx := context.Background()
	plugins, envelope := manager.DiscoverPlugins(ctx)

	if envelope.Code != "" {
		t.Fatalf("DiscoverPlugins failed: %v", envelope.Message)
	}

	if len(plugins) != 1 {
		t.Fatalf("Expected 1 plugin, got %d", len(plugins))
	}

	plugin := plugins[0]
	if plugin.Manifest.ID != "com.test.plugin" {
		t.Errorf("Expected plugin ID 'com.test.plugin', got '%s'", plugin.Manifest.ID)
	}

	if plugin.Manifest.Name != "Test Plugin" {
		t.Errorf("Expected plugin name 'Test Plugin', got '%s'", plugin.Manifest.Name)
	}
}

func TestPluginManager_InstallPlugin(t *testing.T) {
	// Create temporary directories for test
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	pluginsDir := filepath.Join(tmpDir, "plugins")

	logger := logging.NewTestLogger()
	manager := NewManager(logger, pluginsDir)

	// Create source plugin
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	manifest := PluginManifest{
		ID:          "com.test.install",
		Name:        "Install Test Plugin",
		Version:     "1.0.0",
		Type:        PluginTypeImporter,
		EntryPoint:  "index.js",
		Permissions: []Permission{PermissionReadRepo},
	}

	manifestData, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}

	manifestPath := filepath.Join(sourceDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Create plugin entry point
	entryPath := filepath.Join(sourceDir, "index.js")
	if err := os.WriteFile(entryPath, []byte("// Test plugin"), 0644); err != nil {
		t.Fatalf("Failed to write entry point: %v", err)
	}

	// Install plugin
	ctx := context.Background()
	installation, envelope := manager.InstallPlugin(ctx, sourceDir)

	if envelope.Code != "" {
		t.Fatalf("InstallPlugin failed: %v", envelope.Message)
	}

	if installation.Manifest.ID != "com.test.install" {
		t.Errorf("Expected plugin ID 'com.test.install', got '%s'", installation.Manifest.ID)
	}

	if !installation.Enabled {
		t.Error("Expected plugin to be enabled after installation")
	}

	// Verify plugin files were copied
	targetManifest := filepath.Join(pluginsDir, "com.test.install", "manifest.json")
	if _, err := os.Stat(targetManifest); os.IsNotExist(err) {
		t.Error("Manifest file not found in target directory")
	}

	targetEntry := filepath.Join(pluginsDir, "com.test.install", "index.js")
	if _, err := os.Stat(targetEntry); os.IsNotExist(err) {
		t.Error("Entry point file not found in target directory")
	}
}