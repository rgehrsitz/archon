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

	// Create test project file
	root := model.NewComponent("root", "Root", "system")
	components := []*model.Component{root}

	data, err := json.Marshal(components)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	projectFile := filepath.Join(dir, "project.json")
	if err := os.WriteFile(projectFile, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	return dir
}

func TestNewConfigVault(t *testing.T) {
	vault := NewConfigVault()
	if vault == nil {
		t.Error("NewConfigVault() returned nil")
	}
	if vault.tree == nil {
		t.Error("tree not initialized")
	}
	if vault.rootPath != "" {
		t.Error("rootPath should be empty")
	}
}

func TestLoad(t *testing.T) {
	// Setup test project
	dir := setupTestProject(t)
	defer os.RemoveAll(dir)

	vault := NewConfigVault()

	// Test loading valid project
	err := vault.Load(dir)
	if err != nil {
		t.Errorf("Load() error = %v", err)
	}
	if vault.rootPath != dir {
		t.Errorf("rootPath = %v, want %v", vault.rootPath, dir)
	}
	if vault.tree == nil {
		t.Error("tree not loaded")
	}

	// Test loading non-existent project
	err = vault.Load("nonexistent")
	if err == nil {
		t.Error("Expected error when loading non-existent project")
	}
}

func TestSave(t *testing.T) {
	// Setup test project
	dir := setupTestProject(t)
	defer os.RemoveAll(dir)

	vault := NewConfigVault()

	// Load project
	err := vault.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Modify tree
	component := model.NewComponent("test", "Test", "device")
	vault.tree.AddComponent(component)

	// Save changes
	err = vault.Save()
	if err != nil {
		t.Errorf("Save() error = %v", err)
	}

	// Verify changes were saved
	vault2 := NewConfigVault()
	err = vault2.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check if component exists in loaded tree
	_, exists := vault2.tree.Components["test"]
	if !exists {
		t.Error("Component not found after save and reload")
	}
}

func TestGetComponentTree(t *testing.T) {
	vault := NewConfigVault()

	// Test getting tree without loaded project
	tree, err := vault.GetComponentTree()
	if err == nil {
		t.Error("Expected error when getting tree without loaded project")
	}
	if tree != nil {
		t.Error("Expected nil tree when getting tree without loaded project")
	}

	// Setup test project
	dir := setupTestProject(t)
	defer os.RemoveAll(dir)

	// Load project
	err = vault.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test getting tree
	tree, err = vault.GetComponentTree()
	if err != nil {
		t.Errorf("GetComponentTree() error = %v", err)
	}
	if tree == nil {
		t.Error("GetComponentTree() returned nil")
	}
}

func TestUpdateComponent(t *testing.T) {
	// Setup test project
	dir := setupTestProject(t)
	defer os.RemoveAll(dir)

	vault := NewConfigVault()

	// Load project
	err := vault.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test updating component
	component := model.NewComponent("test", "Test", "device")
	err = vault.UpdateComponent(component)
	if err != nil {
		t.Errorf("UpdateComponent() error = %v", err)
	}

	// Verify component was updated
	_, exists := vault.tree.Components["test"]
	if !exists {
		t.Error("Component not found after update")
	}
}

func TestDeleteComponent(t *testing.T) {
	// Setup test project
	dir := setupTestProject(t)
	defer os.RemoveAll(dir)

	vault := NewConfigVault()

	// Load project
	err := vault.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Add test component
	component := model.NewComponent("test", "Test", "device")
	vault.tree.AddComponent(component)

	// Test deleting component
	err = vault.DeleteComponent("test")
	if err != nil {
		t.Errorf("DeleteComponent() error = %v", err)
	}

	// Verify component was deleted
	_, exists := vault.tree.Components["test"]
	if exists {
		t.Error("Component still exists after delete")
	}
}
