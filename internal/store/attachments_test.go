package store

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAttachmentStore_Store(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Test storing a small file
	content := "Hello, world!"
	reader := strings.NewReader(content)
	filename := "hello.txt"

	attachment, err := store.Store(reader, filename)
	if err != nil {
		t.Fatalf("failed to store attachment: %v", err)
	}

	// Verify attachment metadata
	if attachment.Type != "attachment" {
		t.Errorf("expected type 'attachment', got '%s'", attachment.Type)
	}
	if attachment.Filename != filename {
		t.Errorf("expected filename '%s', got '%s'", filename, attachment.Filename)
	}
	if attachment.Size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), attachment.Size)
	}
	if attachment.Hash == "" {
		t.Error("expected non-empty hash")
	}

	// Verify file was created with correct path structure
	expectedPath := store.getAttachmentPath(attachment.Hash)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("attachment file was not created at expected path: %s", expectedPath)
	}
}

func TestAttachmentStore_StoreDeduplication(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	content := "Duplicate content"
	
	// Store the same content twice
	attachment1, err := store.Store(strings.NewReader(content), "file1.txt")
	if err != nil {
		t.Fatalf("failed to store first attachment: %v", err)
	}

	attachment2, err := store.Store(strings.NewReader(content), "file2.txt")
	if err != nil {
		t.Fatalf("failed to store second attachment: %v", err)
	}

	// Should have same hash (deduplication)
	if attachment1.Hash != attachment2.Hash {
		t.Errorf("expected same hash for identical content, got %s vs %s", 
			attachment1.Hash, attachment2.Hash)
	}

	// Only one file should exist on disk
	attachmentPath := store.getAttachmentPath(attachment1.Hash)
	if _, err := os.Stat(attachmentPath); os.IsNotExist(err) {
		t.Error("attachment file should exist after deduplication")
	}
}

func TestAttachmentStore_Retrieve(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Store an attachment first
	originalContent := "Content to retrieve"
	attachment, err := store.Store(strings.NewReader(originalContent), "test.txt")
	if err != nil {
		t.Fatalf("failed to store attachment: %v", err)
	}

	// Retrieve it back
	reader, err := store.Retrieve(attachment.Hash)
	if err != nil {
		t.Fatalf("failed to retrieve attachment: %v", err)
	}
	defer reader.Close()

	// Read the content
	retrievedContent, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read retrieved content: %v", err)
	}

	if string(retrievedContent) != originalContent {
		t.Errorf("retrieved content doesn't match: expected '%s', got '%s'", 
			originalContent, string(retrievedContent))
	}
}

func TestAttachmentStore_RetrieveNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Try to retrieve non-existent attachment
	_, err := store.Retrieve("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	if err == nil {
		t.Error("expected error when retrieving non-existent attachment")
	}
	
	// Should be a "not found" error
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestAttachmentStore_GetInfo(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Store an attachment
	content := "Info test content"
	attachment, err := store.Store(strings.NewReader(content), "info.txt")
	if err != nil {
		t.Fatalf("failed to store attachment: %v", err)
	}

	// Get info
	info, err := store.GetInfo(attachment.Hash)
	if err != nil {
		t.Fatalf("failed to get attachment info: %v", err)
	}

	// Verify info
	if info.Hash != attachment.Hash {
		t.Errorf("expected hash '%s', got '%s'", attachment.Hash, info.Hash)
	}
	if info.Size != attachment.Size {
		t.Errorf("expected size %d, got %d", attachment.Size, info.Size)
	}
	if info.IsLFS != (attachment.Size >= DefaultLFSThresholdBytes) {
		t.Errorf("expected IsLFS=%t for size %d", attachment.Size >= DefaultLFSThresholdBytes, attachment.Size)
	}
}

func TestAttachmentStore_Delete(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Store an attachment
	attachment, err := store.Store(strings.NewReader("Delete me"), "delete.txt")
	if err != nil {
		t.Fatalf("failed to store attachment: %v", err)
	}

	// Verify it exists
	attachmentPath := store.getAttachmentPath(attachment.Hash)
	if _, err := os.Stat(attachmentPath); os.IsNotExist(err) {
		t.Fatal("attachment should exist before deletion")
	}

	// Delete it
	if err := store.Delete(attachment.Hash); err != nil {
		t.Fatalf("failed to delete attachment: %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(attachmentPath); !os.IsNotExist(err) {
		t.Error("attachment should not exist after deletion")
	}

	// Try to delete again (should fail)
	if err := store.Delete(attachment.Hash); err == nil {
		t.Error("expected error when deleting non-existent attachment")
	}
}

func TestAttachmentStore_List(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Initially should be empty
	attachments, err := store.List()
	if err != nil {
		t.Fatalf("failed to list attachments: %v", err)
	}
	if len(attachments) != 0 {
		t.Errorf("expected 0 attachments, got %d", len(attachments))
	}

	// Store some attachments
	testFiles := []struct {
		content  string
		filename string
	}{
		{"Content 1", "file1.txt"},
		{"Content 2", "file2.txt"},
		{"Content 3", "file3.txt"},
	}

	var expectedHashes []string
	for _, tf := range testFiles {
		attachment, err := store.Store(strings.NewReader(tf.content), tf.filename)
		if err != nil {
			t.Fatalf("failed to store attachment %s: %v", tf.filename, err)
		}
		expectedHashes = append(expectedHashes, attachment.Hash)
	}

	// List attachments
	attachments, err = store.List()
	if err != nil {
		t.Fatalf("failed to list attachments: %v", err)
	}

	if len(attachments) != len(testFiles) {
		t.Errorf("expected %d attachments, got %d", len(testFiles), len(attachments))
	}

	// Verify all expected hashes are present
	foundHashes := make(map[string]bool)
	for _, att := range attachments {
		foundHashes[att.Hash] = true
	}

	for _, expectedHash := range expectedHashes {
		if !foundHashes[expectedHash] {
			t.Errorf("expected hash %s not found in list", expectedHash)
		}
	}
}

func TestAttachmentStore_Verify(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Store an attachment
	content := "Verify this content"
	attachment, err := store.Store(strings.NewReader(content), "verify.txt")
	if err != nil {
		t.Fatalf("failed to store attachment: %v", err)
	}

	// Verify should succeed
	if err := store.Verify(attachment.Hash); err != nil {
		t.Errorf("verification failed: %v", err)
	}

	// Corrupt the file
	attachmentPath := store.getAttachmentPath(attachment.Hash)
	if err := os.WriteFile(attachmentPath, []byte("corrupted content"), 0o644); err != nil {
		t.Fatalf("failed to corrupt file: %v", err)
	}

	// Verification should now fail
	if err := store.Verify(attachment.Hash); err == nil {
		t.Error("expected verification to fail for corrupted file")
	}
}

func TestAttachmentStore_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Create content larger than LFS threshold
	largeContent := strings.Repeat("A", int(DefaultLFSThresholdBytes+1000))
	
	attachment, err := store.Store(strings.NewReader(largeContent), "large.txt")
	if err != nil {
		t.Fatalf("failed to store large attachment: %v", err)
	}

	// Verify size
	if attachment.Size != int64(len(largeContent)) {
		t.Errorf("expected size %d, got %d", len(largeContent), attachment.Size)
	}

	// Verify info shows IsLFS=true
	info, err := store.GetInfo(attachment.Hash)
	if err != nil {
		t.Fatalf("failed to get info for large file: %v", err)
	}
	if !info.IsLFS {
		t.Error("expected large file to be marked as LFS")
	}

	// Should still be retrievable
	reader, err := store.Retrieve(attachment.Hash)
	if err != nil {
		t.Fatalf("failed to retrieve large attachment: %v", err)
	}
	defer reader.Close()

	retrievedContent, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read large attachment: %v", err)
	}

	if len(retrievedContent) != len(largeContent) {
		t.Errorf("retrieved large content size mismatch: expected %d, got %d", 
			len(largeContent), len(retrievedContent))
	}
}

func TestAttachmentStore_PathSharding(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Store an attachment
	attachment, err := store.Store(strings.NewReader("sharding test"), "shard.txt")
	if err != nil {
		t.Fatalf("failed to store attachment: %v", err)
	}

	// Verify path structure: attachments/XX/full-hash
	expectedPrefix := attachment.Hash[:HashPrefixLength]
	expectedPath := filepath.Join(store.attachmentDir, expectedPrefix, attachment.Hash)
	
	actualPath := store.getAttachmentPath(attachment.Hash)
	if actualPath != expectedPath {
		t.Errorf("expected path '%s', got '%s'", expectedPath, actualPath)
	}

	// Verify file exists at sharded path
	if _, err := os.Stat(actualPath); os.IsNotExist(err) {
		t.Error("attachment file should exist at sharded path")
	}
}

func TestIsValidHash(t *testing.T) {
	tests := []struct {
		hash  string
		valid bool
	}{
		{"abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", true},
		{"ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789", true},
		{"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", true},
		{"", false},                                                                    // empty
		{"short", false},                                                               // too short
		{"abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789x", false}, // too long
		{"ghijkl0123456789abcdef0123456789abcdef0123456789abcdef0123456789", false},  // invalid hex chars
		{"abcdef0123456789abcdef0123456789abcdef0123456789abcdef012345678", false},   // too short by 1
	}

	for _, tt := range tests {
		t.Run(tt.hash, func(t *testing.T) {
			if isValidHash(tt.hash) != tt.valid {
				t.Errorf("isValidHash(%q) = %t, want %t", tt.hash, !tt.valid, tt.valid)
			}
		})
	}
}

func TestAttachmentStore_BinaryContent(t *testing.T) {
	tempDir := t.TempDir()
	store := NewAttachmentStore(tempDir)

	// Create binary content
	binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD, 0x80, 0x7F}
	
	attachment, err := store.Store(bytes.NewReader(binaryContent), "binary.dat")
	if err != nil {
		t.Fatalf("failed to store binary attachment: %v", err)
	}

	// Retrieve and verify
	reader, err := store.Retrieve(attachment.Hash)
	if err != nil {
		t.Fatalf("failed to retrieve binary attachment: %v", err)
	}
	defer reader.Close()

	retrievedContent, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read binary content: %v", err)
	}

	if !bytes.Equal(binaryContent, retrievedContent) {
		t.Error("binary content mismatch after store/retrieve cycle")
	}
}