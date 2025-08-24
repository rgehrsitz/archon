package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/id"
	"github.com/rgehrsitz/archon/internal/types"
)

func TestGarbageCollect(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Create some test attachments
	attachment1, err := store.Store(strings.NewReader("content 1"), "file1.txt")
	if err != nil {
		t.Fatalf("Failed to store attachment 1: %v", err)
	}

	attachment2, err := store.Store(strings.NewReader("content 2"), "file2.txt")
	if err != nil {
		t.Fatalf("Failed to store attachment 2: %v", err)
	}

	attachment3, err := store.Store(strings.NewReader("content 3"), "file3.txt")
	if err != nil {
		t.Fatalf("Failed to store attachment 3: %v", err)
	}

	// Test GarbageCollect with some referenced hashes
	referencedHashes := []string{attachment1.Hash, attachment3.Hash}
	deleted, err := store.GarbageCollect(referencedHashes)
	if err != nil {
		t.Fatalf("GarbageCollect failed: %v", err)
	}

	// Should have deleted 1 attachment (attachment2)
	if deleted != 1 {
		t.Errorf("Expected 1 deleted attachment, got %d", deleted)
	}

	// Verify attachment2 is gone
	_, err = store.GetInfo(attachment2.Hash)
	if err == nil {
		t.Error("Expected attachment2 to be deleted")
	}

	// Verify attachment1 and attachment3 still exist
	_, err = store.GetInfo(attachment1.Hash)
	if err != nil {
		t.Errorf("Expected attachment1 to still exist: %v", err)
	}

	_, err = store.GetInfo(attachment3.Hash)
	if err != nil {
		t.Errorf("Expected attachment3 to still exist: %v", err)
	}
}

func TestGarbageCollectEmptyReferences(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Create some test attachments
	_, err := store.Store(strings.NewReader("content 1"), "file1.txt")
	if err != nil {
		t.Fatalf("Failed to store attachment 1: %v", err)
	}

	_, err = store.Store(strings.NewReader("content 2"), "file2.txt")
	if err != nil {
		t.Fatalf("Failed to store attachment 2: %v", err)
	}

	// Test GarbageCollect with no referenced hashes (should delete all)
	deleted, err := store.GarbageCollect([]string{})
	if err != nil {
		t.Fatalf("GarbageCollect failed: %v", err)
	}

	// Should have deleted 2 attachments
	if deleted != 2 {
		t.Errorf("Expected 2 deleted attachments, got %d", deleted)
	}

	// Verify no attachments remain
	attachments, err := store.List()
	if err != nil {
		t.Fatalf("Failed to list attachments: %v", err)
	}
	if len(attachments) != 0 {
		t.Errorf("Expected no attachments to remain, got %d", len(attachments))
	}
}

func TestGarbageCollectProject(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a project structure
	if err := os.MkdirAll(filepath.Join(tempDir, "nodes"), 0o755); err != nil {
		t.Fatalf("Failed to create nodes directory: %v", err)
	}

	// Create project.json
	project := &types.Project{
		RootID:        id.NewV7(),
		SchemaVersion: types.CurrentSchemaVersion,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	loader := NewLoader(tempDir)
	if err := loader.SaveProject(project); err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}

	// Create attachment store
	store := NewAttachmentStore(tempDir)

	// Create some test attachments
	attachment1, err := store.Store(strings.NewReader("referenced content"), "ref.txt")
	if err != nil {
		t.Fatalf("Failed to store referenced attachment: %v", err)
	}

	attachment2, err := store.Store(strings.NewReader("unreferenced content"), "unref.txt")
	if err != nil {
		t.Fatalf("Failed to store unreferenced attachment: %v", err)
	}

	// Create nodes - one with attachment reference, one without
	rootNode := &types.Node{
		ID:        project.RootID,
		Name:      "Root",
		Properties: map[string]types.Property{
			"document": {
				TypeHint: "attachment",
				Value: map[string]interface{}{
					"type":     "attachment",
					"hash":     attachment1.Hash,
					"filename": attachment1.Filename,
					"size":     float64(attachment1.Size),
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	childNode := &types.Node{
		ID:        id.NewV7(),
		Name:      "Child",
		Properties: map[string]types.Property{
			"description": {
				TypeHint: "string",
				Value:    "Just a regular property",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save the nodes
	if err := loader.SaveNode(rootNode); err != nil {
		t.Fatalf("Failed to save root node: %v", err)
	}
	if err := loader.SaveNode(childNode); err != nil {
		t.Fatalf("Failed to save child node: %v", err)
	}

	// Run garbage collection on the project
	deleted, err := store.GarbageCollectProject(tempDir)
	if err != nil {
		t.Fatalf("GarbageCollectProject failed: %v", err)
	}

	// Should have deleted 1 attachment (attachment2 - unreferenced)
	if deleted != 1 {
		t.Errorf("Expected 1 deleted attachment, got %d", deleted)
	}

	// Verify referenced attachment still exists
	_, err = store.GetInfo(attachment1.Hash)
	if err != nil {
		t.Errorf("Expected referenced attachment to still exist: %v", err)
	}

	// Verify unreferenced attachment is gone
	_, err = store.GetInfo(attachment2.Hash)
	if err == nil {
		t.Error("Expected unreferenced attachment to be deleted")
	}
}

func TestGarbageCollectProjectNoNodes(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Create some test attachments in a project with no nodes
	_, err := store.Store(strings.NewReader("orphaned content 1"), "orphan1.txt")
	if err != nil {
		t.Fatalf("Failed to store attachment 1: %v", err)
	}

	_, err = store.Store(strings.NewReader("orphaned content 2"), "orphan2.txt")
	if err != nil {
		t.Fatalf("Failed to store attachment 2: %v", err)
	}

	// Run garbage collection (should delete all attachments since no nodes exist)
	deleted, err := store.GarbageCollectProject(tempDir)
	if err != nil {
		t.Fatalf("GarbageCollectProject failed: %v", err)
	}

	// Should have deleted all attachments
	if deleted != 2 {
		t.Errorf("Expected 2 deleted attachments, got %d", deleted)
	}

	// Verify no attachments remain
	attachments, err := store.List()
	if err != nil {
		t.Fatalf("Failed to list attachments: %v", err)
	}
	if len(attachments) != 0 {
		t.Errorf("Expected no attachments to remain, got %d", len(attachments))
	}
}

func TestGarbageCollectProjectCorruptedNodes(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create nodes directory
	nodesDir := filepath.Join(tempDir, "nodes")
	if err := os.MkdirAll(nodesDir, 0o755); err != nil {
		t.Fatalf("Failed to create nodes directory: %v", err)
	}

	// Create a valid node
	validNodeID := id.NewV7()
	validNode := &types.Node{
		ID:        validNodeID,
		Name:      "Valid Node",
		Properties: map[string]types.Property{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	loader := NewLoader(tempDir)
	if err := loader.SaveNode(validNode); err != nil {
		t.Fatalf("Failed to save valid node: %v", err)
	}

	// Create a corrupted node file
	corruptedNodeID := id.NewV7()
	corruptedPath := filepath.Join(nodesDir, corruptedNodeID+".json")
	if err := os.WriteFile(corruptedPath, []byte("invalid json content"), 0o644); err != nil {
		t.Fatalf("Failed to create corrupted node file: %v", err)
	}

	// Create attachment store and add some attachments
	store := NewAttachmentStore(tempDir)
	_, err := store.Store(strings.NewReader("test content"), "test.txt")
	if err != nil {
		t.Fatalf("Failed to store attachment: %v", err)
	}

	// Run garbage collection (should handle corrupted nodes gracefully)
	deleted, err := store.GarbageCollectProject(tempDir)
	if err != nil {
		t.Fatalf("GarbageCollectProject failed: %v", err)
	}

	// Should have deleted the attachment since no valid nodes reference it
	if deleted != 1 {
		t.Errorf("Expected 1 deleted attachment, got %d", deleted)
	}
}

func TestGarbageCollectWithDuplicateReferences(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Create test attachments
	attachment1, err := store.Store(strings.NewReader("content 1"), "file1.txt")
	if err != nil {
		t.Fatalf("Failed to store attachment 1: %v", err)
	}

	_, err = store.Store(strings.NewReader("content 2"), "file2.txt")
	if err != nil {
		t.Fatalf("Failed to store attachment 2: %v", err)
	}

	// Test GarbageCollect with duplicate references (should handle gracefully)
	referencedHashes := []string{
		attachment1.Hash,
		attachment1.Hash, // duplicate
		attachment1.Hash, // another duplicate
	}
	deleted, err := store.GarbageCollect(referencedHashes)
	if err != nil {
		t.Fatalf("GarbageCollect failed: %v", err)
	}

	// Should have deleted 1 attachment (attachment2)
	if deleted != 1 {
		t.Errorf("Expected 1 deleted attachment, got %d", deleted)
	}

	// Verify attachment1 still exists
	_, err = store.GetInfo(attachment1.Hash)
	if err != nil {
		t.Errorf("Expected attachment1 to still exist: %v", err)
	}
}

func TestGarbageCollectNoAttachments(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Test garbage collection with no stored attachments
	deleted, err := store.GarbageCollect([]string{"fake-hash"})
	if err != nil {
		t.Fatalf("GarbageCollect failed: %v", err)
	}

	// Should have deleted 0 attachments
	if deleted != 0 {
		t.Errorf("Expected 0 deleted attachments, got %d", deleted)
	}
}