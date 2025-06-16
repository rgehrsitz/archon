// Package snapshot provides functionality for managing Archon snapshots.
package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Snapshot errors
var (
	ErrInvalidSnapshot = errors.New("invalid snapshot")
	ErrSnapshotExists  = errors.New("snapshot already exists")
	ErrSnapshotNotFound = errors.New("snapshot not found")
	ErrDuplicateTag    = errors.New("duplicate tag")
	ErrInvalidTag      = errors.New("invalid tag")
)

// SnapshotData represents a versioned state of the configuration
type SnapshotData struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Tag       string    `json:"tag,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author"`
	Tree      []byte    `json:"tree"` // Serialized component tree
}

// SnapshotManager handles the creation, retrieval, and management of snapshots
type SnapshotManager struct {
	SnapshotsDir string
	snapshots    map[string]*SnapshotData
	tagIndex     map[string]string // tag -> snapshot ID mapping
}

// NewSnapshotManager creates a new snapshot manager for the given directory
func NewSnapshotManager(snapshotsDir string) (*SnapshotManager, error) {
	if snapshotsDir == "" {
		return nil, errors.New("snapshots directory path cannot be empty")
	}

	// Create snapshots directory if it doesn't exist
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	manager := &SnapshotManager{
		SnapshotsDir: snapshotsDir,
		snapshots:    make(map[string]*SnapshotData),
		tagIndex:     make(map[string]string),
	}

	// Load existing snapshots
	if err := manager.loadSnapshots(); err != nil {
		return nil, err
	}

	return manager, nil
}

// loadSnapshots reads all snapshot files from the snapshots directory
func (m *SnapshotManager) loadSnapshots() error {
	files, err := os.ReadDir(m.SnapshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist yet
		}
		return fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue // Skip directories and non-JSON files
		}

		filePath := filepath.Join(m.SnapshotsDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read snapshot file %s: %w", file.Name(), err)
		}

		var snapshot SnapshotData
		if err := json.Unmarshal(data, &snapshot); err != nil {
			return fmt.Errorf("failed to parse snapshot file %s: %w", file.Name(), err)
		}

		m.snapshots[snapshot.ID] = &snapshot
		
		// Build tag index
		if snapshot.Tag != "" {
			m.tagIndex[snapshot.Tag] = snapshot.ID
		}
	}

	return nil
}

// CreateSnapshot creates a new snapshot with the given components data
func (m *SnapshotManager) CreateSnapshot(components []byte, description string) (*SnapshotData, error) {
	// Create snapshot
	id := uuid.New().String()
	snapshot := &SnapshotData{
		ID:        id,
		Message:   description,
		Timestamp: time.Now().UTC(),
		Author:    "system",
		Tree:      components,
	}

	// Save snapshot to file
	if err := m.saveSnapshot(snapshot); err != nil {
		return nil, err
	}

	// Update in-memory indexes
	m.snapshots[snapshot.ID] = snapshot

	return snapshot, nil
}

// CreateSnapshotWithTag creates a new snapshot with a tag and validation
func (m *SnapshotManager) CreateSnapshotWithTag(components []byte, description, tag string) (*SnapshotData, error) {
	// Validate tag if provided
	if tag != "" {
		if err := m.validateTag(tag); err != nil {
			return nil, err
		}
	}

	// Create snapshot
	id := uuid.New().String()
	snapshot := &SnapshotData{
		ID:        id,
		Message:   description,
		Tag:       tag,
		Timestamp: time.Now().UTC(),
		Author:    "system",
		Tree:      components,
	}

	// Save snapshot to file
	if err := m.saveSnapshot(snapshot); err != nil {
		return nil, err
	}

	// Update in-memory indexes
	m.snapshots[snapshot.ID] = snapshot
	if tag != "" {
		m.tagIndex[tag] = snapshot.ID
	}

	return snapshot, nil
}

// validateTag validates a snapshot tag
func (m *SnapshotManager) validateTag(tag string) error {
	if tag == "" {
		return ErrInvalidTag
	}

	// Check for duplicate tag
	if _, exists := m.tagIndex[tag]; exists {
		return ErrDuplicateTag
	}

	// Validate tag format (basic validation - alphanumeric, dots, hyphens, underscores)
	for _, r := range tag {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			 (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_') {
			return ErrInvalidTag
		}
	}

	return nil
}

// saveSnapshot writes a snapshot to disk
func (m *SnapshotManager) saveSnapshot(snapshot *SnapshotData) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize snapshot: %w", err)
	}

	filePath := filepath.Join(m.SnapshotsDir, snapshot.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}

// GetSnapshot retrieves a snapshot by ID
func (m *SnapshotManager) GetSnapshot(id string) (*SnapshotData, error) {
	snapshot, exists := m.snapshots[id]
	if !exists {
		return nil, ErrSnapshotNotFound
	}
	return snapshot, nil
}

// ListSnapshots returns all snapshots sorted by timestamp (newest first)
func (m *SnapshotManager) ListSnapshots() []*SnapshotData {
	snapshots := make([]*SnapshotData, 0, len(m.snapshots))
	for _, snapshot := range m.snapshots {
		snapshots = append(snapshots, snapshot)
	}

	// Sort by timestamp (newest first)
	sortSnapshotsByTimestamp(snapshots)
	return snapshots
}

// sortSnapshotsByTimestamp sorts snapshots by timestamp in descending order
func sortSnapshotsByTimestamp(snapshots []*SnapshotData) {
	for i := 0; i < len(snapshots)-1; i++ {
		for j := i + 1; j < len(snapshots); j++ {
			if snapshots[i].Timestamp.Before(snapshots[j].Timestamp) {
				snapshots[i], snapshots[j] = snapshots[j], snapshots[i]
			}
		}
	}
}

// DeleteSnapshot removes a snapshot
func (m *SnapshotManager) DeleteSnapshot(id string) error {
	_, exists := m.snapshots[id]
	if !exists {
		return ErrSnapshotNotFound
	}

	// Remove from snapshots map
	delete(m.snapshots, id)

	// Delete file
	filePath := filepath.Join(m.SnapshotsDir, id+".json")
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete snapshot file: %w", err)
	}

	return nil
}

// NewSnapshotData creates a new SnapshotData instance
func NewSnapshotData(id, message string, author string, tree []byte) *SnapshotData {
	return &SnapshotData{
		ID:        id,
		Message:   message,
		Timestamp: time.Now(),
		Author:    author,
		Tree:      tree,
	}
}

// Save writes the snapshot data to disk
func (s *SnapshotData) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	snapshotFile := filepath.Join(path, s.ID+".json")
	if err := os.WriteFile(snapshotFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}

// Load reads a snapshot from disk
func LoadSnapshot(path string) (*SnapshotData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot file: %w", err)
	}

	var snapshot SnapshotData
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot data: %w", err)
	}

	return &snapshot, nil
}
