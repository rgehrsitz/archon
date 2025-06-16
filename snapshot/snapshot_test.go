package snapshot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSnapshotManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	snapshotsDir := filepath.Join(tempDir, "snapshots")

	// Test with valid directory
	manager, err := NewSnapshotManager(snapshotsDir)
	if err != nil {
		t.Errorf("NewSnapshotManager() error = %v", err)
		return
	}
	if manager == nil {
		t.Error("NewSnapshotManager() returned nil")
		return
	}
	if manager.SnapshotsDir != snapshotsDir {
		t.Errorf("SnapshotsDir = %s, want %s", manager.SnapshotsDir, snapshotsDir)
	}

	// Test with empty directory path
	_, err = NewSnapshotManager("")
	if err == nil {
		t.Error("NewSnapshotManager() with empty path should return error")
	}

	// Verify snapshots directory was created
	if _, err := os.Stat(snapshotsDir); os.IsNotExist(err) {
		t.Error("Snapshots directory was not created")
	}
}

func TestSnapshotManager_CreateSnapshot(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	snapshotsDir := filepath.Join(tempDir, "snapshots")
	manager, _ := NewSnapshotManager(snapshotsDir)

	// Sample component data
	components := []byte(`[{"id":"comp1","name":"Component 1","type":"device"}]`)

	// Test creating a snapshot without a tag
	snapshot, err := manager.CreateSnapshot(components, "Test snapshot")
	if err != nil {
		t.Errorf("CreateSnapshot() error = %v", err)
		return
	}
	if snapshot.ID == "" {
		t.Error("CreateSnapshot() returned snapshot with empty ID")
	}

	// Verify snapshot was saved to disk
	filePath := filepath.Join(snapshotsDir, snapshot.ID+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Snapshot file was not created")
	}

	// Test creating a snapshot with duplicate tag
	_, err = manager.CreateSnapshotWithTag(components, "Duplicate tag", "v1.0")
	if err != nil {
		t.Errorf("CreateSnapshotWithTag() first call error = %v", err)
	}
	
	_, err = manager.CreateSnapshotWithTag(components, "Duplicate tag", "v1.0")
	if err == nil {
		t.Error("CreateSnapshotWithTag() with duplicate tag should return error")
	}

	// Test creating a snapshot with invalid tag
	_, err = manager.CreateSnapshotWithTag(components, "Invalid tag", "invalid tag with spaces")
	if err == nil {
		t.Error("CreateSnapshotWithTag() with invalid tag should return error")
	}
}

func TestSnapshotManager_GetSnapshot(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	snapshotsDir := filepath.Join(tempDir, "snapshots")
	manager, _ := NewSnapshotManager(snapshotsDir)

	// Create a test snapshot
	components := []byte(`[{"id":"comp1","name":"Component 1","type":"device"}]`)
	snapshot, _ := manager.CreateSnapshot(components, "Test snapshot")

	// Test getting snapshot by ID
	retrieved, err := manager.GetSnapshot(snapshot.ID)
	if err != nil {
		t.Errorf("GetSnapshot() error = %v", err)
		return
	}
	if retrieved.ID != snapshot.ID {
		t.Errorf("Retrieved ID = %s, want %s", retrieved.ID, snapshot.ID)
	}

	// Test getting non-existent snapshot
	_, err = manager.GetSnapshot("nonexistent")
	if err == nil {
		t.Error("GetSnapshot() with non-existent ID should return error")
	}
}

func TestSnapshotManager_ListSnapshots(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	snapshotsDir := filepath.Join(tempDir, "snapshots")
	manager, _ := NewSnapshotManager(snapshotsDir)

	// Create test snapshots with different timestamps
	components := []byte(`[{"id":"comp1","name":"Component 1","type":"device"}]`)
	
	// Create first snapshot
	snapshot1, _ := manager.CreateSnapshot(components, "First snapshot")
	
	// Wait a moment to ensure different timestamps
	time.Sleep(10 * time.Millisecond)
	
	// Create second snapshot
	snapshot2, _ := manager.CreateSnapshot(components, "Second snapshot")

	// List snapshots
	snapshots := manager.ListSnapshots()
	
	// Verify snapshots are returned in correct order (newest first)
	if len(snapshots) != 2 {
		t.Errorf("ListSnapshots() returned %d snapshots, want 2", len(snapshots))
		return
	}
	
	if snapshots[0].ID != snapshot2.ID {
		t.Errorf("First snapshot in list has ID %s, want %s", snapshots[0].ID, snapshot2.ID)
	}
	
	if snapshots[1].ID != snapshot1.ID {
		t.Errorf("Second snapshot in list has ID %s, want %s", snapshots[1].ID, snapshot1.ID)
	}
}

func TestSnapshotManager_DeleteSnapshot(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	snapshotsDir := filepath.Join(tempDir, "snapshots")
	manager, _ := NewSnapshotManager(snapshotsDir)

	// Create a test snapshot
	components := []byte(`[{"id":"comp1","name":"Component 1","type":"device"}]`)
	snapshot, _ := manager.CreateSnapshot(components, "Test snapshot")

	// Delete snapshot
	err = manager.DeleteSnapshot(snapshot.ID)
	if err != nil {
		t.Errorf("DeleteSnapshot() error = %v", err)
		return
	}

	// Verify snapshot was deleted
	_, err = manager.GetSnapshot(snapshot.ID)
	if err == nil {
		t.Error("GetSnapshot() should return error after deletion")
	}

	// Verify snapshot file was deleted
	filePath := filepath.Join(snapshotsDir, snapshot.ID+".json")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Snapshot file was not deleted")
	}

	// Test deleting non-existent snapshot
	err = manager.DeleteSnapshot("nonexistent")
	if err == nil {
		t.Error("DeleteSnapshot() with non-existent ID should return error")
	}
}

func TestLoadSnapshots(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	snapshotsDir := filepath.Join(tempDir, "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		t.Fatalf("Failed to create snapshots directory: %v", err)
	}

	// Create a test snapshot file manually
	snapshot := SnapshotData{ // Changed from Snapshot to SnapshotData
		ID:          "test-id",
		Message:     "Test snapshot", // Changed from Description to Message
		Timestamp:   time.Now().UTC(),
		Author:      "test-author", // Added Author field
		Tree:        json.RawMessage(`[{\"id\":\"comp1\",\"name\":\"Component 1\",\"type\":\"device\"}]`), // Changed from Components to Tree
		// Tag:         "v1.0", // Field "Tag" does not exist in SnapshotData, commented out
		// Metadata:    map[string]string{"key": "value"}, // Field "Metadata" does not exist in SnapshotData, commented out
	}

	data, _ := json.MarshalIndent(snapshot, "", "  ")
	filePath := filepath.Join(snapshotsDir, snapshot.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("Failed to write test snapshot file: %v", err)
	}

	// Create a non-JSON file to test filtering
	nonJsonPath := filepath.Join(snapshotsDir, "not-a-snapshot.txt")
	if err := os.WriteFile(nonJsonPath, []byte("not a snapshot"), 0644); err != nil {
		t.Fatalf("Failed to write non-JSON file: %v", err)
	}

	// Create a subdirectory to test filtering
	subDir := filepath.Join(snapshotsDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Initialize manager to load snapshots
	manager, err := NewSnapshotManager(snapshotsDir)
	if err != nil {
		t.Errorf("NewSnapshotManager() error = %v", err)
		return
	}

	// Verify snapshot was loaded
	loaded, err := manager.GetSnapshot("test-id")
	if err != nil {
		t.Errorf("GetSnapshot() error = %v", err)
		return
	}

	if loaded.ID != snapshot.ID {
		t.Errorf("Loaded ID = %s, want %s", loaded.ID, snapshot.ID)
	}
}
