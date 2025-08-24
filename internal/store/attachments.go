package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/git"
	"github.com/rgehrsitz/archon/internal/types"
)

// Attachments management: content-addressed storage with LFS integration
const (
	DefaultLFSThresholdBytes = 1 * 1024 * 1024 // 1MB
	AttachmentDirName        = "attachments"
	HashPrefixLength         = 2 // Use first 2 chars for directory sharding
)

// AttachmentStore manages content-addressed file storage with LFS integration
type AttachmentStore struct {
	basePath      string
	lfsThreshold  int64
	attachmentDir string
	gitRepo       git.Repository // Optional Git repository for LFS
}

// NewAttachmentStore creates a new attachment store
func NewAttachmentStore(basePath string) *AttachmentStore {
	return &AttachmentStore{
		basePath:      basePath,
		lfsThreshold:  DefaultLFSThresholdBytes,
		attachmentDir: filepath.Join(basePath, AttachmentDirName),
	}
}

// WithGitRepository configures the attachment store to use Git LFS for large files
func (as *AttachmentStore) WithGitRepository(repo git.Repository) *AttachmentStore {
	as.gitRepo = repo
	return as
}

// SetLFSThreshold changes the size threshold for LFS (default 1MB)
func (as *AttachmentStore) SetLFSThreshold(bytes int64) {
	as.lfsThreshold = bytes
}

// GetBasePath returns the base path for the attachment store
func (as *AttachmentStore) GetBasePath() string {
	return as.basePath
}

// AttachmentInfo contains metadata about a stored attachment
type AttachmentInfo struct {
	Hash       string
	Size       int64
	Path       string
	IsLFS      bool
	StoredAt   time.Time
	RefCount   int // Number of nodes referencing this attachment
}

// Store saves a file as a content-addressed attachment
func (as *AttachmentStore) Store(reader io.Reader, filename string) (*types.Attachment, error) {
	// Create a temporary file to read and hash the content
	tempFile, err := os.CreateTemp("", "archon-attachment-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Hash while copying to temp file
	hash := sha256.New()
	multiWriter := io.MultiWriter(tempFile, hash)
	
	size, err := io.Copy(multiWriter, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	hashStr := hex.EncodeToString(hash.Sum(nil))
	
	// Create attachment metadata
	attachment := &types.Attachment{
		Type:     "attachment",
		Hash:     hashStr,
		Filename: filename,
		Size:     size,
	}

	// Check if attachment already exists
	attachmentPath := as.getAttachmentPath(hashStr)
	if _, err := os.Stat(attachmentPath); err == nil {
		// File already exists, no need to store again (deduplication)
		return attachment, nil
	}

	// Ensure attachment directory exists
	attachmentDir := filepath.Dir(attachmentPath)
	if err := os.MkdirAll(attachmentDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create attachment directory: %w", err)
	}

	// Copy from temp file to final location
	if _, err := tempFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to rewind temp file: %w", err)
	}

	finalFile, err := os.Create(attachmentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create attachment file: %w", err)
	}
	defer finalFile.Close()

	if _, err := io.Copy(finalFile, tempFile); err != nil {
		os.Remove(attachmentPath) // Cleanup on failure
		return nil, fmt.Errorf("failed to copy attachment: %w", err)
	}

	// Handle LFS tracking for large files
	if size >= as.lfsThreshold && as.gitRepo != nil {
		if err := as.setupLFSTracking(); err != nil {
			// Log warning but don't fail - file is still stored locally
			// TODO: Add proper logging when available
		}
	}

	return attachment, nil
}

// Retrieve reads an attachment by its hash
func (as *AttachmentStore) Retrieve(hash string) (io.ReadCloser, error) {
	if !isValidHash(hash) {
		return nil, errors.New(errors.ErrInvalidInput, "invalid hash format")
	}

	attachmentPath := as.getAttachmentPath(hash)
	
	file, err := os.Open(attachmentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrNotFound, "attachment not found: "+hash)
		}
		return nil, fmt.Errorf("failed to open attachment: %w", err)
	}

	return file, nil
}

// GetInfo returns metadata about an attachment
func (as *AttachmentStore) GetInfo(hash string) (*AttachmentInfo, error) {
	if !isValidHash(hash) {
		return nil, errors.New(errors.ErrInvalidInput, "invalid hash format")
	}

	attachmentPath := as.getAttachmentPath(hash)
	
	stat, err := os.Stat(attachmentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrNotFound, "attachment not found: "+hash)
		}
		return nil, fmt.Errorf("failed to stat attachment: %w", err)
	}

	return &AttachmentInfo{
		Hash:     hash,
		Size:     stat.Size(),
		Path:     attachmentPath,
		IsLFS:    stat.Size() >= as.lfsThreshold,
		StoredAt: stat.ModTime(),
		RefCount: 0, // TODO: Implement reference counting
	}, nil
}

// Delete removes an attachment from storage
func (as *AttachmentStore) Delete(hash string) error {
	if !isValidHash(hash) {
		return errors.New(errors.ErrInvalidInput, "invalid hash format")
	}

	attachmentPath := as.getAttachmentPath(hash)
	
	if err := os.Remove(attachmentPath); err != nil {
		if os.IsNotExist(err) {
			return errors.New(errors.ErrNotFound, "attachment not found: "+hash)
		}
		return fmt.Errorf("failed to remove attachment: %w", err)
	}

	// Try to remove empty directory (ignore errors)
	parentDir := filepath.Dir(attachmentPath)
	os.Remove(parentDir)

	return nil
}

// List returns all stored attachments
func (as *AttachmentStore) List() ([]*AttachmentInfo, error) {
	var attachments []*AttachmentInfo

	if _, err := os.Stat(as.attachmentDir); os.IsNotExist(err) {
		return attachments, nil // No attachments directory yet
	}

	err := filepath.Walk(as.attachmentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Extract hash from path - filename should be the hash
			hash := info.Name()

			if isValidHash(hash) {
				attachments = append(attachments, &AttachmentInfo{
					Hash:     hash,
					Size:     info.Size(),
					Path:     path,
					IsLFS:    info.Size() >= as.lfsThreshold,
					StoredAt: info.ModTime(),
					RefCount: 0, // TODO: Implement reference counting
				})
			}
		}

		return nil
	})

	return attachments, err
}

// Verify checks the integrity of an attachment
func (as *AttachmentStore) Verify(hash string) error {
	if !isValidHash(hash) {
		return errors.New(errors.ErrInvalidInput, "invalid hash format")
	}

	file, err := as.Retrieve(hash)
	if err != nil {
		return err
	}
	defer file.Close()

	// Recompute hash
	computedHash := sha256.New()
	if _, err := io.Copy(computedHash, file); err != nil {
		return fmt.Errorf("failed to read attachment for verification: %w", err)
	}

	computedHashStr := hex.EncodeToString(computedHash.Sum(nil))
	if computedHashStr != hash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", hash, computedHashStr)
	}

	return nil
}

// getAttachmentPath returns the file system path for an attachment hash
func (as *AttachmentStore) getAttachmentPath(hash string) string {
	// Shard by first 2 characters: attachments/ab/abcdef123...
	prefix := hash[:HashPrefixLength]
	return filepath.Join(as.attachmentDir, prefix, hash)
}

// isValidHash checks if a hash string is a valid SHA-256 hex string
func isValidHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	
	// Check if all characters are valid hex
	for _, r := range hash {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	
	return true
}

// setupLFSTracking ensures Git LFS is initialized and tracking attachments
func (as *AttachmentStore) setupLFSTracking() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if LFS is already enabled
	enabled, env := as.gitRepo.IsLFSEnabled(ctx)
	if env.Code != "" {
		return fmt.Errorf("failed to check LFS status: %s", env.Message)
	}

	// Initialize LFS if not already enabled
	if !enabled {
		if env := as.gitRepo.InitLFS(ctx); env.Code != "" {
			return fmt.Errorf("failed to initialize Git LFS: %s", env.Message)
		}
	}

	// Track attachments pattern
	attachmentPattern := AttachmentDirName + "/**"
	if env := as.gitRepo.TrackLFSPattern(ctx, attachmentPattern); env.Code != "" {
		return fmt.Errorf("failed to track LFS pattern %s: %s", attachmentPattern, env.Message)
	}

	return nil
}

// IsLFSFile returns true if the given file size should use LFS
func (as *AttachmentStore) IsLFSFile(size int64) bool {
	return size >= as.lfsThreshold
}

// ValidateAttachmentReference checks if an attachment reference is valid
func (as *AttachmentStore) ValidateAttachmentReference(attachment *types.Attachment) error {
	if attachment == nil {
		return errors.New(errors.ErrInvalidInput, "attachment reference is nil")
	}

	// Validate attachment type
	if attachment.Type != "attachment" {
		return errors.New(errors.ErrInvalidInput, "invalid attachment type: "+attachment.Type)
	}

	// Validate hash
	if !isValidHash(attachment.Hash) {
		return errors.New(errors.ErrInvalidInput, "invalid attachment hash: "+attachment.Hash)
	}

	// Validate filename
	if attachment.Filename == "" {
		return errors.New(errors.ErrInvalidInput, "attachment filename cannot be empty")
	}

	// Validate size
	if attachment.Size < 0 {
		return errors.New(errors.ErrInvalidInput, "attachment size cannot be negative")
	}

	// Check if attachment exists
	info, err := as.GetInfo(attachment.Hash)
	if err != nil {
		return fmt.Errorf("attachment not found: %w", err)
	}

	// Verify size matches
	if info.Size != attachment.Size {
		return fmt.Errorf("attachment size mismatch: expected %d, got %d", 
			attachment.Size, info.Size)
	}

	return nil
}

// ValidateNodeAttachments validates all attachment references in a node's properties
func (as *AttachmentStore) ValidateNodeAttachments(node *types.Node) []error {
	var validationErrors []error

	if node.Properties == nil {
		return validationErrors
	}

	for propKey, property := range node.Properties {
		// Check if this property is an attachment
		if property.TypeHint == "attachment" {
			// Try to parse the value as an attachment
			if attachmentData, ok := property.Value.(map[string]interface{}); ok {
				attachment := &types.Attachment{}
				
				// Extract fields from the map
				if typeVal, exists := attachmentData["type"]; exists {
					if typeStr, ok := typeVal.(string); ok {
						attachment.Type = typeStr
					}
				}
				if hashVal, exists := attachmentData["hash"]; exists {
					if hashStr, ok := hashVal.(string); ok {
						attachment.Hash = hashStr
					}
				}
				if filenameVal, exists := attachmentData["filename"]; exists {
					if filenameStr, ok := filenameVal.(string); ok {
						attachment.Filename = filenameStr
					}
				}
				if sizeVal, exists := attachmentData["size"]; exists {
					switch s := sizeVal.(type) {
					case float64:
						attachment.Size = int64(s)
					case int64:
						attachment.Size = s
					case int:
						attachment.Size = int64(s)
					}
				}

				// Validate the attachment reference
				if err := as.ValidateAttachmentReference(attachment); err != nil {
					validationErrors = append(validationErrors, 
						fmt.Errorf("property %s: %w", propKey, err))
				}
			} else {
				validationErrors = append(validationErrors, 
					fmt.Errorf("property %s: attachment value is not a valid object", propKey))
			}
		}
	}

	return validationErrors
}

// GetReferencedHashes extracts all attachment hashes referenced by a node
func (as *AttachmentStore) GetReferencedHashes(node *types.Node) []string {
	var hashes []string

	if node.Properties == nil {
		return hashes
	}

	for _, property := range node.Properties {
		if property.TypeHint == "attachment" {
			if attachmentData, ok := property.Value.(map[string]interface{}); ok {
				if hashVal, exists := attachmentData["hash"]; exists {
					if hashStr, ok := hashVal.(string); ok && isValidHash(hashStr) {
						hashes = append(hashes, hashStr)
					}
				}
			}
		}
	}

	return hashes
}

// GarbageCollect removes unreferenced attachments
func (as *AttachmentStore) GarbageCollect(referencedHashes []string) (int, error) {
	// Get all stored attachments
	allAttachments, err := as.List()
	if err != nil {
		return 0, fmt.Errorf("failed to list stored attachments: %w", err)
	}

	// Create a set of referenced hashes for quick lookup
	referenced := make(map[string]bool)
	for _, hash := range referencedHashes {
		referenced[hash] = true
	}

	// Find unreferenced attachments and delete them
	var deleted int
	for _, attachment := range allAttachments {
		if !referenced[attachment.Hash] {
			if err := as.Delete(attachment.Hash); err != nil {
				// Log the error but continue with other deletions
				// TODO: Add proper logging when available
				continue
			}
			deleted++
		}
	}

	return deleted, nil
}

// GarbageCollectProject scans all nodes in a project and removes unreferenced attachments
func (as *AttachmentStore) GarbageCollectProject(projectPath string) (int, error) {
	// Create a loader to scan all project nodes
	loader := NewLoader(projectPath)

	// Get all node files
	nodeFiles, err := loader.ListNodeFiles()
	if err != nil {
		return 0, fmt.Errorf("failed to list project nodes: %w", err)
	}

	// Collect all referenced attachment hashes
	var allReferencedHashes []string
	for _, nodeFile := range nodeFiles {
		// Extract node ID from filename
		nodeID := strings.TrimSuffix(filepath.Base(nodeFile), ".json")
		
		// Load the node
		node, err := loader.LoadNode(nodeID)
		if err != nil {
			// Skip nodes that can't be loaded (they might be corrupted)
			continue
		}

		// Get referenced hashes from this node
		hashes := as.GetReferencedHashes(node)
		allReferencedHashes = append(allReferencedHashes, hashes...)
	}

	// Remove duplicates from referenced hashes
	uniqueHashes := make(map[string]bool)
	for _, hash := range allReferencedHashes {
		uniqueHashes[hash] = true
	}
	
	var referencedHashes []string
	for hash := range uniqueHashes {
		referencedHashes = append(referencedHashes, hash)
	}

	// Perform garbage collection
	return as.GarbageCollect(referencedHashes)
}
