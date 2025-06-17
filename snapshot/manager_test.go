package snapshot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rgehrsitz/archon/model"
	"github.com/rgehrsitz/archon/storage"
)

func setupTestProject(t *testing.T) (*storage.ConfigVault, string) {
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

	componentsFile := filepath.Join(dir, "components.json")
	if err := os.WriteFile(componentsFile, data, 0o644); err != nil {
		t.Fatalf("Failed to write components file: %v", err)
	}

	// Create archon.json
	config := storage.ProjectConfig{
		Version: "1.0",
		Name:    "Test Project",
	}
	configData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configFile := filepath.Join(dir, "archon.json")
	if err := os.WriteFile(configFile, configData, 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create attachments directory
	if err := os.MkdirAll(filepath.Join(dir, "attachments"), 0o755); err != nil {
		t.Fatalf("Failed to create attachments dir: %v", err)
	}

	// Create ConfigVault
	vault, err := storage.NewConfigVault("")
	if err != nil {
		t.Fatalf("Failed to create ConfigVault: %v", err)
	}
	if err := vault.Load(dir); err != nil {
		t.Fatalf("Failed to load project: %v", err)
	}

	return vault, dir
}

func TestNewManager(t *testing.T) {
	vault, err := storage.NewConfigVault("")
	if err != nil {
		t.Fatalf("Failed to create ConfigVault: %v", err)
	}
	manager := NewManager(vault)

	if manager == nil {
		t.Error("NewManager() returned nil")
	}
	if manager.configVault != vault {
		t.Error("configVault not properly set")
	}
	if manager.snapshots == nil {
		t.Error("snapshots slice not initialized")
	}
	if len(manager.snapshots) != 0 {
		t.Error("snapshots slice should be empty")
	}
}

func TestCreate(t *testing.T) {
	vault, dir := setupTestProject(t)
	defer os.RemoveAll(dir)

	manager := NewManager(vault)

	// Create snapshot
	snapshot, err := manager.Create("test snapshot")
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	// Verify snapshot
	if snapshot == nil {
		t.Error("Create() returned nil snapshot")
	}
	if snapshot.ID != "snapshot-1" {
		t.Errorf("ID = %v, want snapshot-1", snapshot.ID)
	}
	if snapshot.Message != "test snapshot" {
		t.Errorf("Message = %v, want test snapshot", snapshot.Message)
	}
	if snapshot.Author != "system" {
		t.Errorf("Author = %v, want system", snapshot.Author)
	}
	if snapshot.Tree == nil {
		t.Error("Tree data not set")
	}

	// Verify snapshot was saved
	snapshots, err := manager.List()
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(snapshots) != 1 {
		t.Errorf("Expected 1 snapshot, got %d", len(snapshots))
	}
}

func TestList(t *testing.T) {
	vault, dir := setupTestProject(t)
	defer os.RemoveAll(dir)

	manager := NewManager(vault)

	// Create multiple snapshots
	snapshots := []string{
		"first snapshot",
		"second snapshot",
		"third snapshot",
	}

	for _, msg := range snapshots {
		_, err := manager.Create(msg)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	// List snapshots
	list, err := manager.List()
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	// Verify snapshots
	if len(list) != len(snapshots) {
		t.Errorf("Expected %d snapshots, got %d", len(snapshots), len(list))
	}

	for i, snapshot := range list {
		if snapshot.Message != snapshots[i] {
			t.Errorf("Snapshot %d message = %v, want %v", i, snapshot.Message, snapshots[i])
		}
	}
}

func TestGet(t *testing.T) {
	vault, dir := setupTestProject(t)
	defer os.RemoveAll(dir)

	manager := NewManager(vault)

	// Create snapshot
	_, err := manager.Create("test snapshot")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Get snapshot
	snapshot, err := manager.Get("snapshot-1")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if snapshot == nil {
		t.Error("Get() returned nil snapshot")
	}
	if snapshot.Message != "test snapshot" {
		t.Errorf("Message = %v, want test snapshot", snapshot.Message)
	}

	// Get non-existent snapshot
	_, err = manager.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent snapshot")
	}
}

func TestRestore(t *testing.T) {
	vault, dir := setupTestProject(t)
	defer os.RemoveAll(dir)

	manager := NewManager(vault)

	// Create initial snapshot
	_, err := manager.Create("initial state")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	// Modify tree - first add a component, then modify it
	component := model.NewComponent("test", "Test", "device")
	if err := vault.AddComponent(component); err != nil {
		t.Fatalf("AddComponent() error = %v", err)
	}

	// Now update the component
	component.Name = "Updated Test"
	if err := vault.UpdateComponent(component); err != nil {
		t.Fatalf("UpdateComponent() error = %v", err)
	}

	// Create second snapshot
	_, err = manager.Create("modified state")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Restore to first snapshot
	err = manager.Restore("snapshot-1")
	if err != nil {
		t.Errorf("Restore() error = %v", err)
	}

	// TODO: Verify tree was restored correctly
	// This requires implementing tree restoration in the ConfigVault
}
