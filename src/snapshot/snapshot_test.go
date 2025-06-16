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
	snapshot, err := manager.CreateSnapshot(components, "", "Test snapshot")
	if err != nil {
		t.Errorf("CreateSnapshot() error = %v", err)
		return
	}
	if snapshot.ID == "" {
		t.Error("CreateSnapshot() returned snapshot with empty ID")
	}
	if snapshot.Tag != "" {
		t.Errorf("Tag = %s, want empty string", snapshot.Tag)
	}
	if snapshot.Description != "Test snapshot" {
		t.Errorf("Description = %s, want 'Test snapshot'", snapshot.Description)
	}

	// Verify snapshot was saved to disk
	filePath := filepath.Join(snapshotsDir, snapshot.ID+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Snapshot file was not created")
	}

	// Test creating a snapshot with a tag
	snapshot2, err := manager.CreateSnapshot(components, "v1.0", "Tagged snapshot")
	if err != nil {
		t.Errorf("CreateSnapshot() with tag error = %v", err)
		return
	}
	if snapshot2.Tag != "v1.0" {
		t.Errorf("Tag = %s, want 'v1.0'", snapshot2.Tag)
	}

	// Test creating a snapshot with duplicate tag
	_, err = manager.CreateSnapshot(components, "v1.0", "Duplicate tag")
	if err == nil {
		t.Error("CreateSnapshot() with duplicate tag should return error")
	}

	// Test creating a snapshot with invalid tag
	_, err = manager.CreateSnapshot(components, "invalid tag", "Invalid tag")
	if err == nil {
		t.Error("CreateSnapshot() with invalid tag should return error")
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
	snapshot, _ := manager.CreateSnapshot(components, "v1.0", "Test snapshot")

	// Test getting snapshot by ID
	retrieved, err := manager.GetSnapshot(snapshot.ID)
	if err != nil {
		t.Errorf("GetSnapshot() error = %v", err)
		return
	}
	if retrieved.ID != snapshot.ID {
		t.Errorf("Retrieved ID = %s, want %s", retrieved.ID, snapshot.ID)
	}

	// Test getting snapshot by tag
	taggedSnapshot, err := manager.GetSnapshotByTag("v1.0")
	if err != nil {
		t.Errorf("GetSnapshotByTag() error = %v", err)
		return
	}
	if taggedSnapshot.ID != snapshot.ID {
		t.Errorf("Tagged snapshot ID = %s, want %s", taggedSnapshot.ID, snapshot.ID)
	}

	// Test getting non-existent snapshot
	_, err = manager.GetSnapshot("nonexistent")
	if err == nil {
		t.Error("GetSnapshot() with non-existent ID should return error")
	}

	// Test getting snapshot with non-existent tag
	_, err = manager.GetSnapshotByTag("nonexistent")
	if err == nil {
		t.Error("GetSnapshotByTag() with non-existent tag should return error")
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
	snapshot1, _ := manager.CreateSnapshot(components, "v1.0", "First snapshot")
	
	// Wait a moment to ensure different timestamps
	time.Sleep(10 * time.Millisecond)
	
	// Create second snapshot
	snapshot2, _ := manager.CreateSnapshot(components, "v1.1", "Second snapshot")

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

func TestSnapshotManager_UpdateTag(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	snapshotsDir := filepath.Join(tempDir, "snapshots")
	manager, _ := NewSnapshotManager(snapshotsDir)

	// Create test snapshots
	components := []byte(`[{"id":"comp1","name":"Component 1","type":"device"}]`)
	snapshot1, _ := manager.CreateSnapshot(components, "v1.0", "First snapshot")
	snapshot2, _ := manager.CreateSnapshot(components, "v1.1", "Second snapshot")

	// Test updating tag
	err = manager.UpdateTag(snapshot1.ID, "v2.0")
	if err != nil {
		t.Errorf("UpdateTag() error = %v", err)
		return
	}

	// Verify tag was updated
	updated, _ := manager.GetSnapshot(snapshot1.ID)
	if updated.Tag != "v2.0" {
		t.Errorf("Updated tag = %s, want 'v2.0'", updated.Tag)
	}

	// Test updating to existing tag
	err = manager.UpdateTag(snapshot2.ID, "v2.0")
	if err == nil {
		t.Error("UpdateTag() with duplicate tag should return error")
	}

	// Test updating tag of non-existent snapshot
	err = manager.UpdateTag("nonexistent", "v3.0")
	if err == nil {
		t.Error("UpdateTag() with non-existent ID should return error")
	}

	// Test updating to invalid tag
	err = manager.UpdateTag(snapshot1.ID, "invalid tag")
	if err == nil {
		t.Error("UpdateTag() with invalid tag should return error")
	}

	// Test removing tag
	err = manager.UpdateTag(snapshot1.ID, "")
	if err != nil {
		t.Errorf("UpdateTag() to remove tag error = %v", err)
		return
	}

	// Verify tag was removed
	updated, _ = manager.GetSnapshot(snapshot1.ID)
	if updated.Tag != "" {
		t.Errorf("Tag should be empty, got %s", updated.Tag)
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
	snapshot, _ := manager.CreateSnapshot(components, "v1.0", "Test snapshot")

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

func TestIsValidTag(t *testing.T) {
	tests := []struct {
		tag      string
		expected bool
	}{
		{"v1.0", true},
		{"release-2.3.4", true},
		{"test_tag", true},
		{"123", true},
		{"", false},
		{"invalid tag", false},
		{"invalid/tag", false},
		{"veryLongTagThatExceedsSixtyFourCharactersLimitAndShouldBeRejectedByValidation", false},
	}

	for _, test := range tests {
		result := isValidTag(test.tag)
		if result != test.expected {
			t.Errorf("isValidTag(%q) = %v, want %v", test.tag, result, test.expected)
		}
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
	snapshot := Snapshot{
		ID:          "test-id",
		Tag:         "v1.0",
		Description: "Test snapshot",
		Timestamp:   time.Now().UTC(),
		Components:  json.RawMessage(`[{"id":"comp1","name":"Component 1","type":"device"}]`),
		Metadata:    map[string]string{"key": "value"},
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

	if loaded.Tag != snapshot.Tag {
		t.Errorf("Loaded tag = %s, want %s", loaded.Tag, snapshot.Tag)
	}

	// Verify tag index was populated
	taggedSnapshot, err := manager.GetSnapshotByTag("v1.0")
	if err != nil {
		t.Errorf("GetSnapshotByTag() error = %v", err)
		return
	}

	if taggedSnapshot.ID != snapshot.ID {
		t.Errorf("Tagged snapshot ID = %s, want %s", taggedSnapshot.ID, snapshot.ID)
	}
}
