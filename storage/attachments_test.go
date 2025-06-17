package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAddAndRemoveAttachment(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	// Setup project
	vault, err := NewConfigVault(tempDir)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := vault.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Create a dummy file to attach
	dummyPath := filepath.Join(tempDir, "dummy.txt")
	dummyContent := []byte("hello world")
	if err := os.WriteFile(dummyPath, dummyContent, 0o644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}
	// Add attachment
	if err := vault.AddAttachment(dummyPath); err != nil {
		t.Fatalf("AddAttachment failed: %v", err)
	}

	attachments, err := vault.LoadAttachments()
	if err != nil {
		t.Fatalf("LoadAttachments failed: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("Expected 1 attachment, got %d", len(attachments))
	}
	att := attachments[0]
	if att.Name != "dummy.txt" {
		t.Errorf("Attachment name = %s, want dummy.txt", att.Name)
	}
	if att.Size != int64(len(dummyContent)) {
		t.Errorf("Attachment size = %d, want %d", att.Size, len(dummyContent))
	}
	if time.Since(att.CreatedAt) > time.Minute {
		t.Errorf("Attachment CreatedAt too old: %v", att.CreatedAt)
	}

	// Remove attachment
	if err := vault.RemoveAttachment("dummy.txt"); err != nil {
		t.Fatalf("RemoveAttachment failed: %v", err)
	}
	attachments, err = vault.LoadAttachments()
	if err != nil {
		t.Fatalf("LoadAttachments after remove failed: %v", err)
	}
	if len(attachments) != 0 {
		t.Errorf("Expected 0 attachments after removal, got %d", len(attachments))
	}
}

func TestAttachmentErrors(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	vault, err := NewConfigVault(tempDir)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}
	if err := vault.Initialize("Test Project"); err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}
	// Try to add a non-existent file
	if err := vault.AddAttachment(filepath.Join(tempDir, "does_not_exist.txt")); err == nil {
		t.Error("Expected error when adding non-existent file, got nil")
	}

	// Try to remove a non-existent attachment
	if err := vault.RemoveAttachment("notfound.txt"); err == nil {
		t.Error("Expected error when removing non-existent attachment, got nil")
	}
}
