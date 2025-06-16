package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rgehrsitz/archon/storage"
)

// Snapshot represents a versioned state of the configuration
type Snapshot struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author"`
	Tree      []byte    `json:"tree"` // Serialized component tree
}

// Manager handles creation and management of snapshots
type Manager struct {
	mu sync.RWMutex

	configVault *storage.ConfigVault
	snapshots   []*Snapshot
}

// NewManager creates a new snapshot manager
func NewManager(vault *storage.ConfigVault) *Manager {
	return &Manager{
		configVault: vault,
		snapshots:   make([]*Snapshot, 0),
	}
}

// Create creates a new snapshot of the current state
func (m *Manager) Create(message string) (*Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get current tree
	tree, err := m.configVault.GetComponentTree()
	if err != nil {
		return nil, fmt.Errorf("no project loaded")
	}

	// Serialize tree
	treeData, err := json.Marshal(tree)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tree: %w", err)
	}

	// Create snapshot
	snapshot := &Snapshot{
		ID:        fmt.Sprintf("snapshot-%d", len(m.snapshots)+1),
		Message:   message,
		Timestamp: time.Now(),
		Author:    "system", // TODO: Get from user context
		Tree:      treeData,
	}

	// Add to list
	m.snapshots = append(m.snapshots, snapshot)

	// Save to disk
	if err := m.save(); err != nil {
		return nil, fmt.Errorf("failed to save snapshot: %w", err)
	}

	return snapshot, nil
}

// List returns all snapshots
func (m *Manager) List() ([]*Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if a project is loaded
	_, err := m.configVault.GetComponentTree()
	if err != nil {
		return nil, fmt.Errorf("no project loaded")
	}

	// Make a copy to prevent modification
	snapshots := make([]*Snapshot, len(m.snapshots))
	copy(snapshots, m.snapshots)
	return snapshots, nil
}

// Get returns a specific snapshot by ID
func (m *Manager) Get(id string) (*Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, snapshot := range m.snapshots {
		if snapshot.ID == id {
			return snapshot, nil
		}
	}
	return nil, fmt.Errorf("snapshot %s not found", id)
}

// Restore restores the state to a specific snapshot
func (m *Manager) Restore(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find snapshot
	var target *Snapshot
	for _, snapshot := range m.snapshots {
		if snapshot.ID == id {
			target = snapshot
			break
		}
	}
	if target == nil {
		return fmt.Errorf("snapshot %s not found", id)
	}

	// Deserialize tree
	var tree interface{}
	if err := json.Unmarshal(target.Tree, &tree); err != nil {
		return fmt.Errorf("failed to deserialize tree: %w", err)
	}

	// TODO: Implement tree restoration
	// This will require changes to the ConfigVault to support restoring state

	return nil
}

// save writes snapshots to disk
func (m *Manager) save() error {
	// Get project path
	// TODO: Add method to ConfigVault to get project path
	projectPath := "" // m.configVault.GetProjectPath()

	// Create snapshots directory
	snapshotsDir := filepath.Join(projectPath, ".snapshots")
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	// Write snapshots file
	data, err := json.MarshalIndent(m.snapshots, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshots: %w", err)
	}

	snapshotsFile := filepath.Join(snapshotsDir, "snapshots.json")
	if err := os.WriteFile(snapshotsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshots file: %w", err)
	}

	return nil
}

// load reads snapshots from disk
func (m *Manager) load() error {
	// Get project path
	// TODO: Add method to ConfigVault to get project path
	projectPath := "" // m.configVault.GetProjectPath()

	// Read snapshots file
	snapshotsFile := filepath.Join(projectPath, ".snapshots", "snapshots.json")
	data, err := os.ReadFile(snapshotsFile)
	if err != nil {
		if os.IsNotExist(err) {
			// No snapshots yet, that's OK
			return nil
		}
		return fmt.Errorf("failed to read snapshots file: %w", err)
	}

	// Parse snapshots
	if err := json.Unmarshal(data, &m.snapshots); err != nil {
		return fmt.Errorf("failed to parse snapshots: %w", err)
	}

	return nil
}
