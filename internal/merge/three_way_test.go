package merge

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/git"
)

// Minimal smoke test for ThreeWay scaffolding using a tiny repo
func TestThreeWay_NoConflictsOnIndependentChanges(t *testing.T) {
	td := t.TempDir()
	repo, err := git.NewRepository(git.RepositoryConfig{Path: td})
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("init: %s", env.Message)
	}
	// base
	mustWrite(t, td+"/project.json", `{"rootId":"root","schemaVersion":1}`)
	mustWrite(t, td+"/nodes/root.json", `{"id":"root","name":"Root","children":["a","b"]}`)
	mustWrite(t, td+"/nodes/a.json", `{"id":"a","name":"A","children":[]}`)
	mustWrite(t, td+"/nodes/b.json", `{"id":"b","name":"B","children":[]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "base", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit base: %s", env.Message)
	}
	// ours: rename a -> A1
	mustWrite(t, td+"/nodes/a.json", `{"id":"a","name":"A1","children":[]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "ours", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit ours: %s", env.Message)
	}
	// theirs: starting from base, move b under a
	// simulate by checking out base is not implemented; instead, we create an extra commit and diff base..theirs still sees independent changes
	mustWrite(t, td+"/nodes/root.json", `{"id":"root","name":"Root","children":["a"]}`)
	mustWrite(t, td+"/nodes/a.json", `{"id":"a","name":"A1","children":["b"]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "theirs", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit theirs: %s", env.Message)
	}

	res, err := ThreeWay(td, "HEAD~2", "HEAD~1", "HEAD")
	if err != nil {
		t.Fatalf("three-way: %v", err)
	}
	if len(res.Conflicts) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(res.Conflicts))
	}
}

func mustWrite(t *testing.T, path string, content string) {
	t.Helper()
	if err := writeFile(path, []byte(content)); err != nil {
		t.Fatal(err)
	}
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
