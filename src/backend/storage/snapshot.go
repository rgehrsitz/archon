// Package storage provides snapshot functionality for Archon projects.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Snapshot represents a point-in-time capture of the project state.
type Snapshot struct {
	ID          string        `json:"id"`          // Unique snapshot ID (e.g., UUID)
	Timestamp   time.Time     `json:"timestamp"`   // When the snapshot was created
	Tag         string        `json:"tag,omitempty"` // Optional user label
	Project     ProjectConfig `json:"project"`     // Project config at snapshot
	Components  []Component   `json:"components"`  // All components at snapshot
	Attachments []Attachment  `json:"attachments"` // Attachment metadata only
	Author      string        `json:"author,omitempty"` // For future multi-user
	Message     string        `json:"message,omitempty"` // Commit message
}

// snapshotsDir is the directory for storing snapshots.
const snapshotsDir = "snapshots"

// tagsIndexFile is the file for mapping tags to snapshot IDs.
const tagsIndexFile = "tags.json"

// tagsIndex maps snapshot tags to IDs.
type tagsIndex map[string]string

// loadTagsIndex loads the tag index from the snapshots directory.
func (p *Project) loadTagsIndex() (tagsIndex, error) {
	path := filepath.Join(p.Path, snapshotsDir, tagsIndexFile)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return tagsIndex{}, nil
		}
		return nil, fmt.Errorf("failed to open tags index: %w", err)
	}
	defer f.Close()
	var idx tagsIndex
	if err := json.NewDecoder(f).Decode(&idx); err != nil {
		return nil, fmt.Errorf("failed to decode tags index: %w", err)
	}
	return idx, nil
}

// saveTagsIndex saves the tag index to the snapshots directory.
func (p *Project) saveTagsIndex(idx tagsIndex) error {
	path := filepath.Join(p.Path, snapshotsDir, tagsIndexFile)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create tags index: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(idx)
}

// CreateSnapshot creates a manual snapshot of the current project state. Tag must be unique if provided.
// Returns the new snapshot ID or error.
func (p *Project) CreateSnapshot(tag, message, author string) (string, error) {
	// Load components
	components, err := LoadComponents(p.GetComponentsPath())
	if err != nil {
		return "", fmt.Errorf("failed to load components: %w", err)
	}
	// Load attachments
	attachments, err := p.LoadAttachments()
	if err != nil {
		return "", fmt.Errorf("failed to load attachments: %w", err)
	}
	// Tag uniqueness check
	if tag != "" {
		idx, err := p.loadTagsIndex()
		if err != nil {
			return "", err
		}
		if _, exists := idx[tag]; exists {
			return "", fmt.Errorf("tag '%s' already exists", tag)
		}
	}
	// Generate ID (timestamp-based for MVP)
	snapID := time.Now().UTC().Format("20060102T150405.000000000Z07")
	snap := Snapshot{
		ID:          snapID,
		Timestamp:   time.Now().UTC(),
		Tag:         tag,
		Project:     p.Config,
		Components:  components,
		Attachments: attachments,
		Author:      author,
		Message:     message,
	}
	if err := p.SaveSnapshot(snap); err != nil {
		return "", err
	}
	// Update tag index if tag provided
	if tag != "" {
		idx, err := p.loadTagsIndex()
		if err != nil {
			return "", err
		}
		idx[tag] = snapID
		if err := p.saveTagsIndex(idx); err != nil {
			return "", err
		}
	}
	return snapID, nil
}

// SaveSnapshot serializes and saves a snapshot to the project's snapshots directory.
func (p *Project) SaveSnapshot(snap Snapshot) error {
	dir := filepath.Join(p.Path, snapshotsDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshots dir: %w", err)
	}
	file := filepath.Join(dir, snap.ID+".json")
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to create snapshot file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(snap)
}

// LoadSnapshot loads a snapshot by ID from the project's snapshots directory.
func (p *Project) LoadSnapshot(id string) (*Snapshot, error) {
	file := filepath.Join(p.Path, snapshotsDir, id+".json")
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open snapshot file: %w", err)
	}
	defer f.Close()
	var snap Snapshot
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		return nil, fmt.Errorf("failed to decode snapshot: %w", err)
	}
	return &snap, nil
}

// ListSnapshots lists all snapshots in the project's snapshots directory.
func (p *Project) ListSnapshots() ([]Snapshot, error) {
	dir := filepath.Join(p.Path, snapshotsDir)
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Snapshot{}, nil
		}
		return nil, fmt.Errorf("failed to read snapshots dir: %w", err)
	}
	var snaps []Snapshot
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}
		id := f.Name()[:len(f.Name())-5] // Remove .json
		snap, err := p.LoadSnapshot(id)
		if err == nil {
			// Ignore corrupt snapshots for now
			snaps = append(snaps, *snap)
		}
	}
	return snaps, nil
}
