package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestGetDiffBasic verifies that GetDiff returns file-level changes and summary between two commits
func TestGetDiffBasic(t *testing.T) {
	// Skip in CI to avoid environment-specific git issues
	if os.Getenv("CI") != "" {
		t.Skip("Skipping diff integration test in CI")
	}

	tempDir := t.TempDir()

	repo, err := NewRepository(RepositoryConfig{Path: tempDir})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize repo
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("Failed to init repo: %s - %s", env.Code, env.Message)
	}

	// Create initial file and commit
	filePath := filepath.Join(tempDir, "a.txt")
	if err := os.WriteFile(filePath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if env := repo.Add(ctx, []string{"a.txt"}); env.Code != "" {
		t.Fatalf("Failed to add file: %s", env.Message)
	}
	if _, env := repo.Commit(ctx, "initial", &Author{Name: "Test", Email: "test@example.com"}); env.Code != "" {
		t.Fatalf("Failed to commit: %s", env.Message)
	}

	// Modify file and commit again
	if err := os.WriteFile(filePath, []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}
	if env := repo.Add(ctx, []string{"a.txt"}); env.Code != "" {
		t.Fatalf("Failed to add modified file: %s", env.Message)
	}
	if _, env := repo.Commit(ctx, "update", &Author{Name: "Test", Email: "test@example.com"}); env.Code != "" {
		t.Fatalf("Failed to commit update: %s", env.Message)
	}

	// Diff between the two commits
	diff, env := repo.GetDiff(ctx, "HEAD~1", "HEAD")
	if env.Code != "" {
		t.Fatalf("GetDiff failed: %s - %s", env.Code, env.Message)
	}
	if diff == nil {
		t.Fatal("Expected diff to be non-nil")
	}

	if diff.Summary.FilesChanged != 1 {
		t.Fatalf("Expected 1 file changed, got %d", diff.Summary.FilesChanged)
	}
	if diff.Summary.Additions < 1 {
		t.Fatalf("Expected at least 1 addition, got %d", diff.Summary.Additions)
	}
	if len(diff.Files) != 1 {
		t.Fatalf("Expected 1 file diff, got %d", len(diff.Files))
	}
	f := diff.Files[0]
	if f.Path != "a.txt" {
		t.Fatalf("Expected file path 'a.txt', got '%s'", f.Path)
	}
	if f.Status != FileStatusModified {
		t.Fatalf("Expected status 'modified', got '%s'", f.Status)
	}
}

// TestGetDiffWithTags verifies diff works when refs are snapshot-like tags
func TestGetDiffWithTags(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping diff integration test in CI")
	}
	tempDir := t.TempDir()
	repo, err := NewRepository(RepositoryConfig{Path: tempDir})
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	defer repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("Init: %s", env.Message)
	}

	// Commit 1
	if err := os.WriteFile(filepath.Join(tempDir, "f.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if env := repo.Add(ctx, []string{"f.txt"}); env.Code != "" {
		t.Fatal(env.Message)
	}
	if _, env := repo.Commit(ctx, "c1", &Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatal(env.Message)
	}
	if env := repo.CreateTag(ctx, "snapshot-c1", ""); env.Code != "" {
		t.Fatal(env.Message)
	}

	// Commit 2
	if err := os.WriteFile(filepath.Join(tempDir, "f.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if env := repo.Add(ctx, []string{"f.txt"}); env.Code != "" {
		t.Fatal(env.Message)
	}
	if _, env := repo.Commit(ctx, "c2", &Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatal(env.Message)
	}
	if env := repo.CreateTag(ctx, "snapshot-c2", ""); env.Code != "" {
		t.Fatal(env.Message)
	}

	diff, env := repo.GetDiff(ctx, "snapshot-c1", "snapshot-c2")
	if env.Code != "" {
		t.Fatalf("GetDiff: %s", env.Message)
	}
	if diff == nil || diff.Summary.FilesChanged != 1 {
		t.Fatalf("unexpected summary: %+v", diff)
	}
	if diff.Summary.Additions < 1 {
		t.Fatalf("expected additions >= 1, got %d", diff.Summary.Additions)
	}
	if diff.Summary.Deletions != 0 {
		t.Fatalf("expected deletions == 0, got %d", diff.Summary.Deletions)
	}
}

// TestGetDiffAddedDeleted verifies added and deleted files are reported
func TestGetDiffAddedDeleted(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI")
	}
	tempDir := t.TempDir()
	repo, err := NewRepository(RepositoryConfig{Path: tempDir})
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	defer repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("Init: %s", env.Message)
	}

	// Commit 1: create a base file so repo has an initial commit
	if err := os.WriteFile(filepath.Join(tempDir, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if env := repo.Add(ctx, []string{"base.txt"}); env.Code != "" {
		t.Fatal(env.Message)
	}
	if _, env := repo.Commit(ctx, "base", &Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatal(env.Message)
	}

	// Commit 2: add file
	if err := os.WriteFile(filepath.Join(tempDir, "new.txt"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if env := repo.Add(ctx, []string{"new.txt"}); env.Code != "" {
		t.Fatal(env.Message)
	}
	if _, env := repo.Commit(ctx, "add", &Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatal(env.Message)
	}

	// Commit 3: delete the file
	if err := os.Remove(filepath.Join(tempDir, "new.txt")); err != nil {
		t.Fatal(err)
	}
	if env := repo.Add(ctx, []string{"-A"}); env.Code != "" {
		t.Fatal(env.Message)
	}
	if _, env := repo.Commit(ctx, "del", &Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatal(env.Message)
	}

	d1, env := repo.GetDiff(ctx, "HEAD~2", "HEAD~1") // base -> add new file
	if env.Code != "" {
		t.Fatalf("GetDiff1: %s", env.Message)
	}
	if len(d1.Files) != 1 || d1.Files[0].Status != FileStatusAdded {
		t.Fatalf("expected added, got %+v", d1.Files)
	}
	if d1.Summary.FilesChanged != 1 {
		t.Fatalf("expected filesChanged=1, got %d", d1.Summary.FilesChanged)
	}
	if d1.Summary.Additions < 1 {
		t.Fatalf("expected additions >= 1, got %d", d1.Summary.Additions)
	}
	if d1.Summary.Deletions != 0 {
		t.Fatalf("expected deletions == 0, got %d", d1.Summary.Deletions)
	}

	d2, env := repo.GetDiff(ctx, "HEAD~1", "HEAD") // add -> delete new file
	if env.Code != "" {
		t.Fatalf("GetDiff2: %s", env.Message)
	}
	if len(d2.Files) != 1 || d2.Files[0].Status != FileStatusDeleted {
		t.Fatalf("expected deleted, got %+v", d2.Files)
	}
	if d2.Summary.FilesChanged != 1 {
		t.Fatalf("expected filesChanged=1, got %d", d2.Summary.FilesChanged)
	}
	if d2.Summary.Additions != 0 {
		t.Fatalf("expected additions == 0, got %d", d2.Summary.Additions)
	}
	if d2.Summary.Deletions < 1 {
		t.Fatalf("expected deletions >= 1, got %d", d2.Summary.Deletions)
	}
}

// TestGetDiffRenameHeuristic verifies we flag path changes as renamed
func TestGetDiffRenameHeuristic(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI")
	}
	tempDir := t.TempDir()
	repo, err := NewRepository(RepositoryConfig{Path: tempDir})
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	defer repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("Init: %s", env.Message)
	}

	// Commit 1
	if err := os.WriteFile(filepath.Join(tempDir, "old.txt"), []byte("data\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if env := repo.Add(ctx, []string{"old.txt"}); env.Code != "" {
		t.Fatal(env.Message)
	}
	if _, env := repo.Commit(ctx, "c1", &Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatal(env.Message)
	}

	// Commit 2: rename
	if err := os.Rename(filepath.Join(tempDir, "old.txt"), filepath.Join(tempDir, "new.txt")); err != nil {
		t.Fatal(err)
	}
	if env := repo.Add(ctx, []string{"-A"}); env.Code != "" {
		t.Fatal(env.Message)
	}
	if _, env := repo.Commit(ctx, "rename", &Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatal(env.Message)
	}

	diff, env := repo.GetDiff(ctx, "HEAD~1", "HEAD")
	if env.Code != "" {
		t.Fatalf("GetDiff: %s", env.Message)
	}
	if len(diff.Files) != 1 {
		t.Fatalf("expected 1 file diff, got %d", len(diff.Files))
	}
	if diff.Files[0].Status != FileStatusRenamed {
		t.Fatalf("expected renamed, got %s", diff.Files[0].Status)
	}
	if diff.Files[0].OldPath == "" || diff.Files[0].Path == "" {
		t.Fatalf("expected old and new path set: %+v", diff.Files[0])
	}
	if diff.Summary.FilesChanged != 1 {
		t.Fatalf("expected filesChanged=1, got %d", diff.Summary.FilesChanged)
	}
	if diff.Summary.Additions != 0 || diff.Summary.Deletions != 0 {
		t.Fatalf("expected no content changes for pure rename, got +%d -%d", diff.Summary.Additions, diff.Summary.Deletions)
	}
}
