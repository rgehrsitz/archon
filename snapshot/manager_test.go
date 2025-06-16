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

	// Create ConfigVault
	vault := storage.NewConfigVault()
	if err := vault.Load(dir); err != nil {
		t.Fatalf("Failed to load project: %v", err)
	}

	return vault, dir
}

func TestNewManager(t *testing.T) {
	vault := storage.NewConfigVault()
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

	// Modify tree
	component := model.NewComponent("test", "Test", "device")
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
