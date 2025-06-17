package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rgehrsitz/archon/model"
)

func setupTestProject(t *testing.T) string {
	// Create temporary directory
	dir, err := os.MkdirTemp("", "archon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test project structure
	// Create components.json
	root := model.NewComponent("root", "Root", "system")
	components := []*model.Component{root}

	data, err := json.Marshal(components)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	componentsFile := filepath.Join(dir, ComponentsFile)
	if err := os.WriteFile(componentsFile, data, 0o644); err != nil {
		t.Fatalf("Failed to write components file: %v", err)
	}

	// Create archon.json
	config := ProjectConfig{
		Version: "1.0",
		Name:    "Test Project",
	}
	configData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configFile := filepath.Join(dir, ConfigFile)
	if err := os.WriteFile(configFile, configData, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create attachments directory
	if err := os.MkdirAll(filepath.Join(dir, AttachmentsDir), 0o755); err != nil {
		t.Fatalf("Failed to create attachments dir: %v", err)
	}

	return dir
}

func TestNewConfigVault(t *testing.T) {
	// Test in-memory vault
	vault, err := NewConfigVault("")
	if err != nil {
		t.Fatalf("NewConfigVault() returned error: %v", err)
	}
	if vault == nil {
		t.Error("NewConfigVault() returned nil")
	}
	if vault.tree == nil {
		t.Error("tree not initialized")
	}
	if vault.rootPath != "" {
		t.Error("rootPath should be empty for in-memory vault")
	}

	// Test with path
	vault2, err := NewConfigVault("/test/path")
	if err != nil {
		t.Fatalf("NewConfigVault() with path returned error: %v", err)
	}
	if vault2.rootPath == "" {
		t.Error("rootPath should be set when path provided")
	}
}

func TestLoad(t *testing.T) {
	// Setup test project
	dir := setupTestProject(t)
	defer os.RemoveAll(dir)

	vault, err := NewConfigVault("")
	if err != nil {
		t.Fatalf("NewConfigVault() returned error: %v", err)
	}

	// Test loading valid project
	if err = vault.Load(dir); err != nil {
		t.Errorf("Load() error = %v", err)
	}
	if vault.rootPath != dir {
		t.Errorf("rootPath = %v, want %v", vault.rootPath, dir)
	}
	if vault.tree == nil {
		t.Error("tree not loaded")
	}

	// Test loading non-existent project
	if err = vault.Load("/non/existent/path"); err == nil {
		t.Error("Load() should fail for non-existent path")
	}
}

func TestSave(t *testing.T) {
	// Setup test project
	dir := setupTestProject(t)
	defer os.RemoveAll(dir)

	vault, err := NewConfigVault("")
	if err != nil {
		t.Fatalf("NewConfigVault() returned error: %v", err)
	}

	// Load project first
	if err = vault.Load(dir); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Add a component
	component := model.NewComponent("test-comp", "Test Component", "test")
	if err = vault.AddComponent(component); err != nil {
		t.Errorf("AddComponent() error = %v", err)
	}

	// Verify save worked by loading again
	vault2, err := NewConfigVault("")
	if err != nil {
		t.Fatalf("NewConfigVault() returned error: %v", err)
	}
	if err = vault2.Load(dir); err != nil {
		t.Errorf("Load() after save error = %v", err)
	}

	// Check component exists
	if _, err = vault2.GetComponent("test-comp"); err != nil {
		t.Errorf("GetComponent() after save error = %v", err)
	}
}

func TestInMemoryOperations(t *testing.T) {
	vault, err := NewConfigVault("")
	if err != nil {
		t.Fatalf("NewConfigVault() returned error: %v", err)
	}

	// Test InitializeInMemory
	root := model.NewComponent("root", "Root", "system")
	components := []*model.Component{root}

	if err = vault.InitializeInMemory(components); err != nil {
		t.Errorf("InitializeInMemory() error = %v", err)
	}

	// Test getting component tree
	tree, err := vault.GetComponentTree()
	if err != nil {
		t.Errorf("GetComponentTree() error = %v", err)
	}
	if tree == nil {
		t.Error("GetComponentTree() returned nil")
	}

	// Test adding component
	component := model.NewComponent("test-comp", "Test Component", "test")
	if err = vault.AddComponent(component); err != nil {
		t.Errorf("AddComponent() error = %v", err)
	}

	// Test getting component
	retrieved, err := vault.GetComponent("test-comp")
	if err != nil {
		t.Errorf("GetComponent() error = %v", err)
	}
	if retrieved.ID != component.ID {
		t.Errorf("GetComponent() returned wrong component")
	}

	// Test updating component
	component.Name = "Updated Name"
	if err = vault.UpdateComponent(component); err != nil {
		t.Errorf("UpdateComponent() error = %v", err)
	}

	// Test deleting component
	if err = vault.DeleteComponent("test-comp"); err != nil {
		t.Errorf("DeleteComponent() error = %v", err)
	}

	// Verify deletion
	if _, err = vault.GetComponent("test-comp"); err == nil {
		t.Error("GetComponent() should fail after deletion")
	}
}

func TestProjectOperations(t *testing.T) {
	// Test creating new project
	dir, err := os.MkdirTemp("", "archon-project-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	vault, err := NewConfigVault(dir)
	if err != nil {
		t.Fatalf("NewConfigVault() returned error: %v", err)
	}

	// Test Initialize
	if err = vault.Initialize("Test Project"); err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	// Verify files were created
	if _, err := os.Stat(filepath.Join(dir, ConfigFile)); err != nil {
		t.Errorf("Config file not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ComponentsFile)); err != nil {
		t.Errorf("Components file not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, AttachmentsDir)); err != nil {
		t.Errorf("Attachments directory not created: %v", err)
	}

	// Test getting config
	config := vault.GetConfig()
	if config.Name != "Test Project" {
		t.Errorf("Config name = %v, want %v", config.Name, "Test Project")
	}
}
