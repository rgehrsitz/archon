package snapshot

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/git"
)

// TestSnapshotCreateAndList tests basic snapshot creation and listing
func TestSnapshotCreateAndList(t *testing.T) {
	// Skip if no git available or in CI
	if os.Getenv("CI") != "" {
		t.Skip("Skipping snapshot test in CI")
	}

	tempDir := t.TempDir()

	// Initialize Git repository first
	config := git.RepositoryConfig{Path: tempDir}
	repo, err := git.NewRepository(config)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize repository
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("Failed to init repository: %s - %s", env.Code, env.Message)
	}

	// Configure git user for commits
	if err := exec.Command("git", "-C", tempDir, "config", "user.name", "Test User").Run(); err != nil {
		t.Fatalf("Failed to set git user.name: %v", err)
	}
	if err := exec.Command("git", "-C", tempDir, "config", "user.email", "test@test.com").Run(); err != nil {
		t.Fatalf("Failed to set git user.email: %v", err)
	}

	// Create some initial content
	testFile := tempDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("Initial content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create snapshot manager
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create snapshot manager: %v", err)
	}
	defer manager.Close()

	// Test 1: List should be empty initially
	snapshots, err := manager.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list snapshots: %v", err)
	}
	if len(snapshots) != 0 {
		t.Errorf("Expected 0 snapshots, got %d", len(snapshots))
	}

	// Test 2: Create first snapshot
	req1 := CreateRequest{
		Name:        "initial-state",
		Message:     "Initial project state",
		Description: "First snapshot of the project",
		Labels:      map[string]string{"type": "milestone"},
	}

	snap1, err := manager.Create(ctx, req1)
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	if snap1.Name != req1.Name {
		t.Errorf("Expected snapshot name '%s', got '%s'", req1.Name, snap1.Name)
	}
	if snap1.Message != req1.Message {
		t.Errorf("Expected message '%s', got '%s'", req1.Message, snap1.Message)
	}
	if snap1.Hash == "" {
		t.Error("Expected non-empty commit hash")
	}

	// Debug: Check if tag was created manually
	gitTags, err := exec.Command("git", "-C", tempDir, "tag", "-l").Output()
	if err != nil {
		t.Logf("Failed to check git tags manually: %v", err)
	} else {
		t.Logf("Manual git tag check output: %s", string(gitTags))
	}

	// Debug: Try calling the repository's ListTags directly 
	repoTags, repoEnv := repo.ListTags(ctx)
	if repoEnv.Code != "" {
		t.Logf("Repository ListTags error: %s - %s", repoEnv.Code, repoEnv.Message)
	} else {
		t.Logf("Repository found %d tags", len(repoTags))
		for i, tag := range repoTags {
			t.Logf("Repo Tag %d: %s (snapshot: %v)", i, tag.Name, tag.IsSnapshot)
		}
	}

	// Test 3: List should now have one snapshot
	snapshots, err = manager.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list snapshots: %v", err)
	}
	t.Logf("Found %d snapshots", len(snapshots))
	for i, snap := range snapshots {
		t.Logf("Snapshot %d: %s (hash: %s)", i, snap.Name, snap.Hash)
	}
	if len(snapshots) != 1 {
		t.Errorf("Expected 1 snapshot, got %d", len(snapshots))
	}

	// Test 4: Get specific snapshot
	retrieved, err := manager.Get(ctx, "initial-state")
	if err != nil {
		t.Fatalf("Failed to get snapshot: %v", err)
	}
	if retrieved.Name != snap1.Name {
		t.Errorf("Retrieved snapshot name mismatch: expected '%s', got '%s'", snap1.Name, retrieved.Name)
	}

	// Test 5: Create second snapshot with different content
	if err := os.WriteFile(testFile, []byte("Updated content"), 0o644); err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	req2 := CreateRequest{
		Name:    "updated-state",
		Message: "Updated project state",
	}

	snap2, err := manager.Create(ctx, req2)
	if err != nil {
		t.Fatalf("Failed to create second snapshot: %v", err)
	}

	// Test 6: List should now have two snapshots (most recent first)
	snapshots, err = manager.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list snapshots: %v", err)
	}
	if len(snapshots) != 2 {
		t.Errorf("Expected 2 snapshots, got %d", len(snapshots))
	}

	// Verify ordering (most recent first)
	if snapshots[0].Name != snap2.Name {
		t.Errorf("Expected most recent snapshot '%s' first, got '%s'", snap2.Name, snapshots[0].Name)
	}
	if snapshots[1].Name != snap1.Name {
		t.Errorf("Expected older snapshot '%s' second, got '%s'", snap1.Name, snapshots[1].Name)
	}
}

// TestSnapshotValidation tests input validation
func TestSnapshotValidation(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create snapshot manager: %v", err)
	}
	defer manager.Close()

	// ctx := context.Background() // Not needed for validation tests

	testCases := []struct {
		name      string
		req       CreateRequest
		expectErr bool
	}{
		{
			name: "valid request",
			req: CreateRequest{
				Name:    "valid-name",
				Message: "Valid message",
			},
			expectErr: false,
		},
		{
			name: "empty name",
			req: CreateRequest{
				Name:    "",
				Message: "Valid message",
			},
			expectErr: true,
		},
		{
			name: "empty message",
			req: CreateRequest{
				Name:    "valid-name",
				Message: "",
			},
			expectErr: true,
		},
		{
			name: "invalid characters in name",
			req: CreateRequest{
				Name:    "invalid@name",
				Message: "Valid message",
			},
			expectErr: true,
		},
		{
			name: "name too long",
			req: CreateRequest{
				Name:    "this-name-is-way-too-long-and-exceeds-the-fifty-character-limit",
				Message: "Valid message",
			},
			expectErr: true,
		},
		{
			name: "valid name with dashes and underscores",
			req: CreateRequest{
				Name:    "valid_name-123",
				Message: "Valid message",
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.validateCreateRequest(tc.req)
			if tc.expectErr && err == nil {
				t.Errorf("Expected validation error, got none")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

// TestSnapshotTagFormat tests tag name formatting
func TestSnapshotTagFormat(t *testing.T) {
	tempDir := t.TempDir()
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create snapshot manager: %v", err)
	}
	defer manager.Close()

	testCases := []struct {
		input    string
		expected string
	}{
		{"my-snapshot", "snapshot-my-snapshot"},
		{"test123", "snapshot-test123"},
		{"feature_branch", "snapshot-feature_branch"},
	}

	for _, tc := range testCases {
		result := manager.formatTagName(tc.input)
		if result != tc.expected {
			t.Errorf("formatTagName(%s) = %s, expected %s", tc.input, result, tc.expected)
		}

		// Test reverse operation
		parsed := manager.parseTagName(result)
		if parsed != tc.input {
			t.Errorf("parseTagName(%s) = %s, expected %s", result, parsed, tc.input)
		}
	}
}

// TestSnapshotRestore tests snapshot restoration functionality
func TestSnapshotRestore(t *testing.T) {
	// Skip if no git available or in CI
	if os.Getenv("CI") != "" {
		t.Skip("Skipping snapshot test in CI")
	}

	tempDir := t.TempDir()

	// Initialize Git repository first
	config := git.RepositoryConfig{Path: tempDir}
	repo, err := git.NewRepository(config)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize repository
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("Failed to init repository: %s - %s", env.Code, env.Message)
	}

	// Configure git user for commits
	if err := exec.Command("git", "-C", tempDir, "config", "user.name", "Test User").Run(); err != nil {
		t.Fatalf("Failed to set git user.name: %v", err)
	}
	if err := exec.Command("git", "-C", tempDir, "config", "user.email", "test@test.com").Run(); err != nil {
		t.Fatalf("Failed to set git user.email: %v", err)
	}

	// Create initial content
	testFile := tempDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("Initial content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create snapshot manager
	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create snapshot manager: %v", err)
	}
	defer manager.Close()

	// Create first snapshot
	req1 := CreateRequest{
		Name:    "initial-state",
		Message: "Initial project state",
	}

	snap1, err := manager.Create(ctx, req1)
	if err != nil {
		t.Fatalf("Failed to create first snapshot: %v", err)
	}

	// Modify content
	if err := os.WriteFile(testFile, []byte("Modified content"), 0o644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Create second snapshot
	req2 := CreateRequest{
		Name:    "modified-state",
		Message: "Modified project state",
	}

	snap2, err := manager.Create(ctx, req2)
	if err != nil {
		t.Fatalf("Failed to create second snapshot: %v", err)
	}

	// Verify current content is modified
	currentContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read current content: %v", err)
	}
	if string(currentContent) != "Modified content" {
		t.Errorf("Expected 'Modified content', got '%s'", string(currentContent))
	}

	// Restore to first snapshot
	if err := manager.Restore(ctx, snap1.Name); err != nil {
		t.Fatalf("Failed to restore snapshot: %v", err)
	}

	// Verify content was restored
	restoredContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read restored content: %v", err)
	}
	if string(restoredContent) != "Initial content" {
		t.Errorf("Expected 'Initial content', got '%s'", string(restoredContent))
	}

	// Test restore with uncommitted changes (should fail)
	if err := os.WriteFile(testFile, []byte("Uncommitted changes"), 0o644); err != nil {
		t.Fatalf("Failed to create uncommitted changes: %v", err)
	}

	if err := manager.Restore(ctx, snap2.Name); err == nil {
		t.Error("Expected restore to fail with uncommitted changes, but it succeeded")
	}
}