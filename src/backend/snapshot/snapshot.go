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
	ErrInvalidTag      = errors.New("invalid tag")
	ErrTagExists       = errors.New("tag already exists")
)

// Snapshot represents a point-in-time capture of component configuration
type Snapshot struct {
	ID          string            `json:"id"`
	Tag         string            `json:"tag,omitempty"`
	Description string            `json:"description,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	Components  json.RawMessage   `json:"components"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SnapshotManager handles the creation, retrieval, and management of snapshots
type SnapshotManager struct {
	SnapshotsDir string
	snapshots    map[string]*Snapshot
	tagIndex     map[string]string // Maps tags to snapshot IDs
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
		snapshots:    make(map[string]*Snapshot),
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

		var snapshot Snapshot
		if err := json.Unmarshal(data, &snapshot); err != nil {
			return fmt.Errorf("failed to parse snapshot file %s: %w", file.Name(), err)
		}

		m.snapshots[snapshot.ID] = &snapshot
		if snapshot.Tag != "" {
			m.tagIndex[snapshot.Tag] = snapshot.ID
		}
	}

	return nil
}

// CreateSnapshot creates a new snapshot with the given components data
func (m *SnapshotManager) CreateSnapshot(components []byte, tag, description string) (*Snapshot, error) {
	// Validate tag if provided
	if tag != "" {
		if !isValidTag(tag) {
			return nil, ErrInvalidTag
		}

		// Check if tag already exists
		if _, exists := m.tagIndex[tag]; exists {
			return nil, ErrTagExists
		}
	}

	// Create snapshot
	id := uuid.New().String()
	snapshot := &Snapshot{
		ID:          id,
		Tag:         tag,
		Description: description,
		Timestamp:   time.Now().UTC(),
		Components:  components,
		Metadata:    make(map[string]string),
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

// saveSnapshot writes a snapshot to disk
func (m *SnapshotManager) saveSnapshot(snapshot *Snapshot) error {
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
func (m *SnapshotManager) GetSnapshot(id string) (*Snapshot, error) {
	snapshot, exists := m.snapshots[id]
	if !exists {
		return nil, ErrSnapshotNotFound
	}
	return snapshot, nil
}

// GetSnapshotByTag retrieves a snapshot by tag
func (m *SnapshotManager) GetSnapshotByTag(tag string) (*Snapshot, error) {
	id, exists := m.tagIndex[tag]
	if !exists {
		return nil, ErrSnapshotNotFound
	}
	return m.GetSnapshot(id)
}

// ListSnapshots returns all snapshots sorted by timestamp (newest first)
func (m *SnapshotManager) ListSnapshots() []*Snapshot {
	snapshots := make([]*Snapshot, 0, len(m.snapshots))
	for _, snapshot := range m.snapshots {
		snapshots = append(snapshots, snapshot)
	}

	// Sort by timestamp (newest first)
	sortSnapshotsByTimestamp(snapshots)
	return snapshots
}

// sortSnapshotsByTimestamp sorts snapshots by timestamp in descending order
func sortSnapshotsByTimestamp(snapshots []*Snapshot) {
	for i := 0; i < len(snapshots)-1; i++ {
		for j := i + 1; j < len(snapshots); j++ {
			if snapshots[i].Timestamp.Before(snapshots[j].Timestamp) {
				snapshots[i], snapshots[j] = snapshots[j], snapshots[i]
			}
		}
	}
}

// UpdateTag updates or adds a tag to a snapshot
func (m *SnapshotManager) UpdateTag(id, tag string) error {
	if tag != "" && !isValidTag(tag) {
		return ErrInvalidTag
	}

	// Check if tag already exists on a different snapshot
	if existingID, exists := m.tagIndex[tag]; exists && existingID != id {
		return ErrTagExists
	}

	snapshot, exists := m.snapshots[id]
	if !exists {
		return ErrSnapshotNotFound
	}

	// Remove old tag from index if it exists
	if snapshot.Tag != "" {
		delete(m.tagIndex, snapshot.Tag)
	}

	// Update tag
	snapshot.Tag = tag
	if tag != "" {
		m.tagIndex[tag] = id
	}

	// Save updated snapshot
	return m.saveSnapshot(snapshot)
}

// DeleteSnapshot removes a snapshot
func (m *SnapshotManager) DeleteSnapshot(id string) error {
	snapshot, exists := m.snapshots[id]
	if !exists {
		return ErrSnapshotNotFound
	}

	// Remove from tag index if tagged
	if snapshot.Tag != "" {
		delete(m.tagIndex, snapshot.Tag)
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

// isValidTag checks if a tag contains only allowed characters
func isValidTag(tag string) bool {
	if len(tag) == 0 || len(tag) > 64 {
		return false
	}

	for _, r := range tag {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}
