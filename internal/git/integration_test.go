package git

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestGitIntegration tests basic Git repository functionality
func TestGitIntegration(t *testing.T) {
	// Skip if no git available
	if os.Getenv("CI") != "" {
		t.Skip("Skipping Git integration test in CI")
	}

	tempDir := t.TempDir()

	// Create repository config
	config := RepositoryConfig{
		Path: tempDir,
	}

	// Create repository instance
	repo, err := NewRepository(config)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test 1: Repository should not exist initially
	if repo.IsRepository() {
		t.Error("Expected repository to not exist initially")
	}

	// Test 2: Initialize repository
	initErr := repo.Init(ctx)
	if initErr.Code != "" {
		t.Fatalf("Failed to initialize repository: %s - %s", initErr.Code, initErr.Message)
	}

	// Test 3: Repository should exist after init
	if !repo.IsRepository() {
		t.Error("Expected repository to exist after init")
	}

	// Test 4: Get current branch (might be empty for new repo)
	branch, branchErr := repo.GetCurrentBranch(ctx)
	if branchErr.Code != "" {
		t.Logf("No current branch (expected for new repo): %s", branchErr.Message)
	} else if branch != "" {
		t.Logf("Current branch: %s", branch)
	}

	// Test 5: Check repository status
	status, statusErr := repo.Status(ctx)
	if statusErr.Code != "" {
		t.Fatalf("Failed to get repository status: %s - %s", statusErr.Code, statusErr.Message)
	}

	if status == nil {
		t.Fatal("Expected status to not be nil")
	}

	// New repository should be clean
	if !status.IsClean {
		t.Error("Expected new repository to be clean")
	}

	t.Logf("Repository status: branch=%s, clean=%t", status.Branch, status.IsClean)
}

// TestIsValidRepository tests the IsValidRepository function
func TestIsValidRepository(t *testing.T) {
	// Test with non-existent path
	if IsValidRepository("/nonexistent/path") {
		t.Error("Expected non-existent path to not be valid repository")
	}

	// Test with empty path
	if IsValidRepository("") {
		t.Error("Expected empty path to not be valid repository")
	}

	// Test with current directory (which is not a git repo)
	if IsValidRepository(".") {
		t.Error("Expected current directory to not be a git repository")
	}
}

// TestRepositoryConfig tests repository configuration
func TestRepositoryConfig(t *testing.T) {
	tempDir := t.TempDir()

	config := RepositoryConfig{
		Path: tempDir,
		PreferCLI: []string{"clone", "push", "pull"},
		PreferGoGit: []string{"status", "history"},
	}

	repo, err := NewRepository(config)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Cast to router to test preferences (this is internal testing)
	if router, ok := repo.(*repositoryRouter); ok {
		if !router.shouldUseCLI("clone") {
			t.Error("Expected to prefer CLI for clone operation")
		}
		if router.shouldUseCLI("status") {
			t.Error("Expected to prefer go-git for status operation")
		}
	} else {
		t.Skip("Cannot test preferences without router implementation")
	}
}