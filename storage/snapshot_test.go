package storage

import (
	"os"
	"testing"
	"time"
)

func TestCreateSnapshot(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup vault
	vault, err := NewConfigVault(tempDir)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := vault.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Create a component and save
	component := Component{ID: "c1", Name: "Root", Type: "rack"}
	if err := SaveComponents(vault.GetComponentsPath(), []Component{component}); err != nil {
		t.Fatalf("Failed to save components: %v", err)
	}

	// Create snapshot
	tag := "v1.0"
	msg := "First snapshot"
	author := "test@example.com"
	snapID, err := vault.CreateSnapshot(tag, msg, author)
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}
	if snapID == "" {
		t.Error("Expected non-empty snapshot ID")
	}

	// Try creating duplicate tag (should fail)
	_, err = vault.CreateSnapshot(tag, "duplicate", author)
	if err == nil {
		t.Error("Expected error for duplicate tag, got nil")
	}

	// Load tags index and verify
	idx, err := vault.loadTagsIndex()
	if err != nil {
		t.Fatalf("Failed to load tags index: %v", err)
	}
	if idx[tag] != snapID {
		t.Errorf("Tag index incorrect: expected %s, got %s", snapID, idx[tag])
	}

	// Load the snapshot and verify
	snap, err := vault.LoadSnapshot(snapID)
	if err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}
	if snap.ID != snapID {
		t.Errorf("Snapshot ID mismatch: expected %s, got %s", snapID, snap.ID)
	}
	if snap.Tag != tag {
		t.Errorf("Snapshot tag mismatch: expected %s, got %s", tag, snap.Tag)
	}
	if snap.Message != msg {
		t.Errorf("Snapshot message mismatch: expected %s, got %s", msg, snap.Message)
	}
	if snap.Author != author {
		t.Errorf("Snapshot author mismatch: expected %s, got %s", author, snap.Author)
	}
}

func TestListSnapshots(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vault, err := NewConfigVault(tempDir)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := vault.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Create multiple snapshots
	for i := 0; i < 3; i++ {
		component := Component{ID: "c1", Name: "Root", Type: "rack"}
		if err := SaveComponents(vault.GetComponentsPath(), []Component{component}); err != nil {
			t.Fatalf("Failed to save components: %v", err)
		}
		_, err := vault.CreateSnapshot("", "snapshot message", "test@example.com")
		if err != nil {
			t.Fatalf("CreateSnapshot failed: %v", err)
		}
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	// List snapshots
	snapshots, err := vault.ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}
	if len(snapshots) != 3 {
		t.Errorf("Expected 3 snapshots, got %d", len(snapshots))
	}
}

func TestSnapshotErrors(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vault, err := NewConfigVault(tempDir)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := vault.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Try to load non-existent snapshot
	_, err = vault.LoadSnapshot("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent snapshot, got nil")
	}
}

func TestEmptySnapshotList(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	vault, err := NewConfigVault(tempDir)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := vault.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// List snapshots when none exist
	snapshots, err := vault.ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}
	if len(snapshots) != 0 {
		t.Errorf("Expected 0 snapshots, got %d", len(snapshots))
	}
}
