package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/archon/backend/model"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Valid path",
			path:    "test_project",
			wantErr: false,
		},
		{
			name:    "Empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("New() returned nil, want non-nil")
			}
			if !tt.wantErr && got.Path == "" {
				t.Errorf("New() returned project with empty path")
			}
		})
	}
}

func TestProject_Initialize(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name      string
		path      string
		projectName string
		preCreate bool // Whether to pre-create the project
		wantErr   bool
	}{
		{
			name:      "New project",
			path:      filepath.Join(tempDir, "new_project"),
			projectName: "Test Project",
			preCreate: false,
			wantErr:   false,
		},
		{
			name:      "Existing project",
			path:      filepath.Join(tempDir, "existing_project"),
			projectName: "Existing Project",
			preCreate: true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pre-create project if needed
			if tt.preCreate {
				p, _ := New(tt.path)
				if err := p.Initialize("Pre-existing"); err != nil {
					t.Fatalf("Failed to pre-create project: %v", err)
				}
			}

			// Test Initialize
			p, err := New(tt.path)
			if err != nil {
				t.Fatalf("New() failed: %v", err)
			}

			err = p.Initialize(tt.projectName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify project structure
			if _, err := os.Stat(filepath.Join(tt.path, ConfigFile)); err != nil {
				t.Errorf("Initialize() failed to create config file: %v", err)
			}

			if _, err := os.Stat(filepath.Join(tt.path, ComponentsFile)); err != nil {
				t.Errorf("Initialize() failed to create components file: %v", err)
			}

			if _, err := os.Stat(filepath.Join(tt.path, AttachmentsDir)); err != nil {
				t.Errorf("Initialize() failed to create attachments directory: %v", err)
			}

			// Verify config content
			configData, err := os.ReadFile(filepath.Join(tt.path, ConfigFile))
			if err != nil {
				t.Fatalf("Failed to read config file: %v", err)
			}

			var config ProjectConfig
			if err := json.Unmarshal(configData, &config); err != nil {
				t.Fatalf("Failed to parse config file: %v", err)
			}

			if config.Name != tt.projectName {
				t.Errorf("Config name = %s, want %s", config.Name, tt.projectName)
			}

			if config.Version != "1.0" {
				t.Errorf("Config version = %s, want 1.0", config.Version)
			}
		})
	}
}

func TestOpen(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a valid project
	validPath := filepath.Join(tempDir, "valid_project")
	validProject, _ := New(validPath)
	if err := validProject.Initialize("Valid Project"); err != nil {
		t.Fatalf("Failed to create valid project: %v", err)
	}

	// Create an invalid project (missing components.json)
	invalidPath := filepath.Join(tempDir, "invalid_project")
	if err := os.MkdirAll(invalidPath, 0755); err != nil {
		t.Fatalf("Failed to create invalid project directory: %v", err)
	}
	// Create only config file but not components.json
	config := ProjectConfig{Name: "Invalid", Version: "1.0"}
	configData, _ := json.MarshalIndent(config, "", "  ")
	if err := os.WriteFile(filepath.Join(invalidPath, ConfigFile), configData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test cases
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Valid project",
			path:    validPath,
			wantErr: false,
		},
		{
			name:    "Non-existent project",
			path:    filepath.Join(tempDir, "nonexistent"),
			wantErr: true,
		},
		{
			name:    "Invalid project structure",
			path:    invalidPath,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Open(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("Open() returned nil, want non-nil")
			}
			if !tt.wantErr && got.Config.Name == "" {
				t.Errorf("Open() returned project with empty name")
			}
		})
	}
}

func TestProject_SaveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a project
	projectPath := filepath.Join(tempDir, "test_project")
	project, _ := New(projectPath)
	if err := project.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Modify config and save
	project.Config.Description = "Updated description"
	project.Config.Metadata["key"] = "value"

	if err := project.SaveConfig(); err != nil {
		t.Errorf("SaveConfig() error = %v", err)
		return
	}

	// Verify saved config
	configData, err := os.ReadFile(filepath.Join(projectPath, ConfigFile))
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var savedConfig ProjectConfig
	if err := json.Unmarshal(configData, &savedConfig); err != nil {
		t.Fatalf("Failed to parse config file: %v", err)
	}

	if savedConfig.Description != "Updated description" {
		t.Errorf("Saved description = %s, want 'Updated description'", savedConfig.Description)
	}

	if savedConfig.Metadata["key"] != "value" {
		t.Errorf("Saved metadata value = %s, want 'value'", savedConfig.Metadata["key"])
	}
}

func TestProject_GetPaths(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a project
	projectPath := filepath.Join(tempDir, "test_project")
	project, _ := New(projectPath)
	if err := project.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Test GetComponentsPath
	componentsPath := project.GetComponentsPath()
	expectedComponentsPath := filepath.Join(projectPath, ComponentsFile)
	if componentsPath != expectedComponentsPath {
		t.Errorf("GetComponentsPath() = %s, want %s", componentsPath, expectedComponentsPath)
	}

	// Test GetAttachmentsPath
	attachmentsPath := project.GetAttachmentsPath()
	expectedAttachmentsPath := filepath.Join(projectPath, AttachmentsDir)
	if attachmentsPath != expectedAttachmentsPath {
		t.Errorf("GetAttachmentsPath() = %s, want %s", attachmentsPath, expectedAttachmentsPath)
	}
}

func TestProject_LoadSaveComponents(t *testing.T) {
	// Create a temporary project
	project, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if err := project.Initialize("test-project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Create test components
	components := []*model.Component{
		model.NewComponent("comp1", "Component 1", "type1"),
		model.NewComponent("comp2", "Component 2", "type2"),
	}

	// Save components
	if err := project.SaveComponents(components); err != nil {
		t.Fatalf("Failed to save components: %v", err)
	}

	// Load components
	loadedComponents, err := project.LoadComponents()
	if err != nil {
		t.Fatalf("Failed to load components: %v", err)
	}

	// Verify components
	if len(loadedComponents) != len(components) {
		t.Errorf("Expected %d components, got %d", len(components), len(loadedComponents))
	}

	for i, c := range components {
		if loadedComponents[i].ID != c.ID {
			t.Errorf("Component %d: expected ID %s, got %s", i, c.ID, loadedComponents[i].ID)
		}
	}
}

func TestProject_UpdateComponent(t *testing.T) {
	// Create a temporary project
	project, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if err := project.Initialize("test-project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Create initial component
	component := model.NewComponent("comp1", "Component 1", "type1")
	if err := project.SaveComponents([]*model.Component{component}); err != nil {
		t.Fatalf("Failed to save initial component: %v", err)
	}

	// Update component
	component.Name = "Updated Component"
	if err := project.UpdateComponent(component); err != nil {
		t.Fatalf("Failed to update component: %v", err)
	}

	// Load and verify
	components, err := project.LoadComponents()
	if err != nil {
		t.Fatalf("Failed to load components: %v", err)
	}

	if len(components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(components))
	}

	if components[0].Name != "Updated Component" {
		t.Errorf("Expected name 'Updated Component', got '%s'", components[0].Name)
	}
}

func TestProject_DeleteComponent(t *testing.T) {
	// Create a temporary project
	project, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if err := project.Initialize("test-project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Create test components
	components := []*model.Component{
		model.NewComponent("comp1", "Component 1", "type1"),
		model.NewComponent("comp2", "Component 2", "type2"),
	}

	if err := project.SaveComponents(components); err != nil {
		t.Fatalf("Failed to save components: %v", err)
	}

	// Delete a component
	if err := project.DeleteComponent("comp1"); err != nil {
		t.Fatalf("Failed to delete component: %v", err)
	}

	// Load and verify
	loadedComponents, err := project.LoadComponents()
	if err != nil {
		t.Fatalf("Failed to load components: %v", err)
	}

	if len(loadedComponents) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(loadedComponents))
	}

	if loadedComponents[0].ID != "comp2" {
		t.Errorf("Expected remaining component ID 'comp2', got '%s'", loadedComponents[0].ID)
	}

	// Try to delete non-existent component
	if err := project.DeleteComponent("nonexistent"); err != model.ErrComponentNotFound {
		t.Errorf("Expected ErrComponentNotFound, got %v", err)
	}
}

func TestProject_StateManagement(t *testing.T) {
	// Create a temporary project
	project, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	if err := project.Initialize("test-project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Test initial state
	state := project.GetState()
	if state.ChangeCount != 0 {
		t.Errorf("Expected initial change count 0, got %d", state.ChangeCount)
	}
	if state.LastSnapshot != "" {
		t.Errorf("Expected empty last snapshot, got %s", state.LastSnapshot)
	}

	// Test unsaved changes
	if project.HasUnsavedChanges() {
		t.Error("Expected no unsaved changes initially")
	}

	// Make some changes
	component := model.NewComponent("comp1", "Component 1", "type1")
	if err := project.SaveComponents([]*model.Component{component}); err != nil {
		t.Fatalf("Failed to save component: %v", err)
	}

	if !project.HasUnsavedChanges() {
		t.Error("Expected unsaved changes after saving component")
	}

	// Test snapshot tracking
	snapshotID := "test-snapshot-1"
	if err := project.UpdateLastSnapshot(snapshotID); err != nil {
		t.Fatalf("Failed to update last snapshot: %v", err)
	}

	if got := project.GetLastSnapshot(); got != snapshotID {
		t.Errorf("Expected last snapshot %s, got %s", snapshotID, got)
	}

	// Test reset change count
	project.ResetChangeCount()
	if project.HasUnsavedChanges() {
		t.Error("Expected no unsaved changes after reset")
	}

	// Verify state after changes
	state = project.GetState()
	if state.ChangeCount != 0 {
		t.Errorf("Expected change count 0 after reset, got %d", state.ChangeCount)
	}
	if state.LastSnapshot != snapshotID {
		t.Errorf("Expected last snapshot %s, got %s", snapshotID, state.LastSnapshot)
	}
}
