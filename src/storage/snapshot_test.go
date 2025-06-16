package storage

import (
	"os"
	"testing"
)

func TestCreateSnapshotAndTag(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup project
	project, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	if err := project.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Create a component and save
	component := Component{ID: "c1", Name: "Root", Type: "rack"}
	if err := SaveComponents(project.GetComponentsPath(), []Component{component}); err != nil {
		t.Fatalf("Failed to save components: %v", err)
	}

	// Create a snapshot with a tag
	tag := "v1"
	msg := "Initial commit"
	author := "tester"
	snapID, err := project.CreateSnapshot(tag, msg, author)
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}
	if snapID == "" {
		t.Error("Expected non-empty snapshot ID")
	}

	// Try to create another snapshot with the same tag (should fail)
	_, err = project.CreateSnapshot(tag, "duplicate", author)
	if err == nil {
		t.Error("Expected error for duplicate tag, got nil")
	}

	// Check that the tag index maps the tag to the correct snapshot ID
	idx, err := project.loadTagsIndex()
	if err != nil {
		t.Fatalf("Failed to load tags index: %v", err)
	}
	if idx[tag] != snapID {
		t.Errorf("Tag index mismatch: got %s, want %s", idx[tag], snapID)
	}

	// Load the snapshot by ID
	snap, err := project.LoadSnapshot(snapID)
	if err != nil {
		t.Fatalf("Failed to load snapshot: %v", err)
	}
	if snap.Tag != tag {
		t.Errorf("Snapshot tag = %s, want %s", snap.Tag, tag)
	}
	if snap.Message != msg {
		t.Errorf("Snapshot message = %s, want %s", snap.Message, msg)
	}
	if snap.Author != author {
		t.Errorf("Snapshot author = %s, want %s", snap.Author, author)
	}
}

func TestAutoSnapshot_After5Changes(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon_test_auto_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	project, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	if err := project.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}
	// Add and save 5 different components using SaveComponentsWithProject
	for i := 0; i < 5; i++ {
		c := Component{ID: string(rune('a'+i)), Name: "C", Type: "t"}
		if err := SaveComponentsWithProject(project, []Component{c}); err != nil {
			t.Fatalf("Failed SaveComponentsWithProject: %v", err)
		}
	}
	snaps, err := project.ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}
	found := false
	for _, s := range snaps {
		if s.Tag != "" && len(s.Tag) > 5 && s.Tag[:5] == "auto-" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Auto-snapshot not found after 5 changes")
	}
}

func TestAutoSnapshot_OnConfigChange(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon_test_auto_cfg_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	project, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	if err := project.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}
	project.Config.Description = "desc"
	if err := project.SaveConfig(); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	snaps, err := project.ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}
	found := false
	for _, s := range snaps {
		if s.Tag != "" && len(s.Tag) > 5 && s.Tag[:5] == "auto-" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Auto-snapshot not found after config change")
	}
}

func TestAutoSnapshot_OnImportStub(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon_test_auto_imp_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	project, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	if err := project.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}
	if err := project.ImportConfigWithAutoSnapshot("dummy"); err != nil {
		t.Fatalf("ImportConfigWithAutoSnapshot failed: %v", err)
	}
	snaps, err := project.ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}
	found := false
	for _, s := range snaps {
		if s.Tag != "" && len(s.Tag) > 5 && s.Tag[:5] == "auto-" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Auto-snapshot not found after import stub")
	}
}
