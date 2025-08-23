package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rgehrsitz/archon/internal/git"
	"github.com/rgehrsitz/archon/internal/logging"
)

// Manager handles snapshot operations for Archon projects
// Following ADR-003: Snapshots are commit + immutable tag pairs
type Manager struct {
	projectPath string
	repo        git.Repository
}

// NewManager creates a new snapshot manager for a project
func NewManager(projectPath string) (*Manager, error) {
	config := git.RepositoryConfig{
		Path: projectPath,
	}
	
	repo, err := git.NewRepository(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return &Manager{
		projectPath: projectPath,
		repo:        repo,
	}, nil
}

// Close cleans up resources
func (m *Manager) Close() error {
	if m.repo != nil {
		return m.repo.Close()
	}
	return nil
}

// Snapshot represents a user-created checkpoint
type Snapshot struct {
	Name        string                 `json:"name"`        // Unique snapshot name
	Hash        string                 `json:"hash"`        // Git commit hash
	Message     string                 `json:"message"`     // Commit message
	Description string                 `json:"description"` // User description
	Labels      map[string]string      `json:"labels"`      // User-defined labels
	Metadata    map[string]interface{} `json:"metadata"`    // Additional metadata
	CreatedAt   time.Time              `json:"createdAt"`   // Creation timestamp
	Author      git.Author             `json:"author"`      // Author information
}

// CreateRequest represents a snapshot creation request
type CreateRequest struct {
	Name        string                 `json:"name"`        // Required: unique name
	Message     string                 `json:"message"`     // Required: commit message
	Description string                 `json:"description"` // Optional: user description
	Labels      map[string]string      `json:"labels"`      // Optional: labels
	Metadata    map[string]interface{} `json:"metadata"`    // Optional: extra data
	Author      *git.Author            `json:"author"`      // Optional: author override
}

// Create creates a new snapshot with commit + tag pair
func (m *Manager) Create(ctx context.Context, req CreateRequest) (*Snapshot, error) {
	// Validate input
	if err := m.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Check if snapshot name is unique
	if exists, err := m.snapshotExists(ctx, req.Name); err != nil {
		return nil, fmt.Errorf("failed to check snapshot existence: %w", err)
	} else if exists {
		return nil, fmt.Errorf("snapshot name '%s' already exists", req.Name)
	}

	// Add all tracked files to staging
	if env := m.repo.Add(ctx, []string{"."}); env.Code != "" {
		return nil, fmt.Errorf("failed to stage files: %s - %s", env.Code, env.Message)
	}

	// Create commit
	commit, env := m.repo.Commit(ctx, req.Message, req.Author)
	if env.Code != "" {
		return nil, fmt.Errorf("failed to create commit: %s - %s", env.Code, env.Message)
	}

	// Create immutable tag
	tagName := m.formatTagName(req.Name)
	tagMessage := fmt.Sprintf("Snapshot: %s", req.Name)
	if req.Description != "" {
		tagMessage += "\n\n" + req.Description
	}

	if env := m.repo.CreateTag(ctx, tagName, tagMessage); env.Code != "" {
		return nil, fmt.Errorf("failed to create tag: %s - %s", env.Code, env.Message)
	}

	// Create metadata file
	snapshot := &Snapshot{
		Name:        req.Name,
		Hash:        commit.Hash,
		Message:     req.Message,
		Description: req.Description,
		Labels:      req.Labels,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		Author:      commit.Author,
	}

	if err := m.saveSnapshotMetadata(snapshot); err != nil {
		logging.Log().Warn().
			Err(err).
			Str("snapshot_name", req.Name).
			Msg("Failed to save snapshot metadata file")
	}

	logging.Log().Info().
		Str("snapshot_name", req.Name).
		Str("commit_hash", commit.Hash).
		Str("tag_name", tagName).
		Msg("Snapshot created successfully")

	return snapshot, nil
}

// List returns all snapshots in chronological order
func (m *Manager) List(ctx context.Context) ([]Snapshot, error) {
	// Get all tags that match snapshot pattern
	tags, env := m.repo.ListTags(ctx)
	if env.Code != "" {
		return nil, fmt.Errorf("failed to list tags: %s - %s", env.Code, env.Message)
	}

	var snapshots []Snapshot
	for _, tag := range tags {
		if tag.IsSnapshot || strings.HasPrefix(tag.Name, "snapshot-") {
			snapshot, err := m.loadSnapshot(ctx, tag)
			if err != nil {
				logging.Log().Warn().
					Err(err).
					Str("tag_name", tag.Name).
					Msg("Failed to load snapshot from tag")
				continue
			}
			snapshots = append(snapshots, *snapshot)
		}
	}

	// Sort by creation time (most recent first)
	for i := 0; i < len(snapshots); i++ {
		for j := i + 1; j < len(snapshots); j++ {
			if snapshots[i].CreatedAt.Before(snapshots[j].CreatedAt) {
				snapshots[i], snapshots[j] = snapshots[j], snapshots[i]
			}
		}
	}

	return snapshots, nil
}

// Get retrieves a specific snapshot by name
func (m *Manager) Get(ctx context.Context, name string) (*Snapshot, error) {
	tagName := m.formatTagName(name)
	
	tags, env := m.repo.ListTags(ctx)
	if env.Code != "" {
		return nil, fmt.Errorf("failed to list tags: %s - %s", env.Code, env.Message)
	}

	for _, tag := range tags {
		if tag.Name == tagName {
			return m.loadSnapshot(ctx, tag)
		}
	}

	return nil, fmt.Errorf("snapshot '%s' not found", name)
}

// Restore restores the project to a snapshot state
func (m *Manager) Restore(ctx context.Context, name string) error {
	// Validate snapshot exists
	snapshot, err := m.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Check if working directory is clean
	status, env := m.repo.Status(ctx)
	if env.Code != "" {
		return fmt.Errorf("failed to get repository status: %s - %s", env.Code, env.Message)
	}

	if !status.IsClean {
		return fmt.Errorf("working directory has uncommitted changes; commit or stash changes before restoring")
	}

	// Restore to the snapshot commit hash
	if env := m.repo.Checkout(ctx, snapshot.Hash); env.Code != "" {
		return fmt.Errorf("failed to checkout snapshot commit: %s - %s", env.Code, env.Message)
	}

	logging.Log().Info().
		Str("snapshot_name", name).
		Str("commit_hash", snapshot.Hash).
		Msg("Successfully restored project to snapshot")

	return nil
}

// Delete removes a snapshot (tag only, not commit)
func (m *Manager) Delete(ctx context.Context, name string) error {
	if exists, err := m.snapshotExists(ctx, name); err != nil {
		return fmt.Errorf("failed to check snapshot existence: %w", err)
	} else if !exists {
		return fmt.Errorf("snapshot '%s' does not exist", name)
	}

	// Remove metadata file
	metadataPath := m.getMetadataPath(name)
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		logging.Log().Warn().
			Err(err).
			Str("metadata_path", metadataPath).
			Msg("Failed to remove snapshot metadata file")
	}

	// Note: We don't delete the Git tag here as it should remain immutable
	// This is a design decision to preserve history even when snapshots are "deleted"
	
	logging.Log().Info().
		Str("snapshot_name", name).
		Msg("Snapshot deleted (tag preserved for history)")

	return nil
}

// Helper methods

func (m *Manager) validateCreateRequest(req CreateRequest) error {
	if req.Name == "" {
		return fmt.Errorf("snapshot name is required")
	}
	if req.Message == "" {
		return fmt.Errorf("commit message is required")
	}
	
	// Validate name format (alphanumeric, dash, underscore)
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(req.Name) {
		return fmt.Errorf("snapshot name must contain only alphanumeric characters, dashes, and underscores")
	}
	
	if len(req.Name) > 50 {
		return fmt.Errorf("snapshot name must be 50 characters or less")
	}
	
	return nil
}

func (m *Manager) snapshotExists(ctx context.Context, name string) (bool, error) {
	tagName := m.formatTagName(name)
	
	tags, env := m.repo.ListTags(ctx)
	if env.Code != "" {
		return false, fmt.Errorf("failed to list tags: %s - %s", env.Code, env.Message)
	}
	
	for _, tag := range tags {
		if tag.Name == tagName {
			return true, nil
		}
	}
	
	return false, nil
}

func (m *Manager) formatTagName(name string) string {
	return "snapshot-" + name
}

func (m *Manager) parseTagName(tagName string) string {
	return strings.TrimPrefix(tagName, "snapshot-")
}

func (m *Manager) getMetadataPath(name string) string {
	snapshotsDir := filepath.Join(m.projectPath, ".archon", "snapshots")
	return filepath.Join(snapshotsDir, name+".json")
}

func (m *Manager) saveSnapshotMetadata(snapshot *Snapshot) error {
	metadataPath := m.getMetadataPath(snapshot.Name)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(metadataPath), 0o755); err != nil {
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}
	
	// Save metadata as JSON
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot metadata: %w", err)
	}
	
	if err := os.WriteFile(metadataPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write snapshot metadata: %w", err)
	}
	
	return nil
}

func (m *Manager) loadSnapshot(ctx context.Context, tag git.Tag) (*Snapshot, error) {
	name := m.parseTagName(tag.Name)
	metadataPath := m.getMetadataPath(name)
	
	// Try to load from metadata file first
	if data, err := os.ReadFile(metadataPath); err == nil {
		var snapshot Snapshot
		if err := json.Unmarshal(data, &snapshot); err == nil {
			return &snapshot, nil
		}
	}
	
	// Fall back to tag information only
	snapshot := &Snapshot{
		Name:      name,
		Hash:      tag.Hash,
		Message:   tag.Message,
		CreatedAt: tag.Date,
	}
	
	return snapshot, nil
}