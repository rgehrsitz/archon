package semantic

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/git"
)

// helper to write a node file
func writeJSON(t *testing.T, dir, path string, content string) {
	t.Helper()
	full := filepath.Join(dir, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSemanticDiffBasic(t *testing.T) {
	td := t.TempDir()

	// init repo
	repo, err := git.NewRepository(git.RepositoryConfig{Path: td})
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("init: %s", env.Message)
	}

	// project + one node A
	writeJSON(t, td, "project.json", `{"rootId":"root","schemaVersion":1}`)
	writeJSON(t, td, "nodes/root.json", `{"id":"root","name":"Root","children":["a"]}`)
	writeJSON(t, td, "nodes/a.json", `{"id":"a","name":"Alpha","children":[]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "initial", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit1: %s", env.Message)
	}

	// modify: rename a, add b under root, reorder
	writeJSON(t, td, "nodes/a.json", `{"id":"a","name":"Alpha2","children":[]}`)
	writeJSON(t, td, "nodes/b.json", `{"id":"b","name":"Beta","children":[]}`)
	writeJSON(t, td, "nodes/root.json", `{"id":"root","name":"Root","children":["b","a"]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "second", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit2: %s", env.Message)
	}

	// run semantic diff HEAD~1..HEAD
	res, env := Diff(td, "HEAD~1", "HEAD")
	if env.Code != "" {
		t.Fatalf("diff: %s", env.Message)
	}

	if res.Summary.Added != 1 {
		t.Fatalf("expected 1 added, got %d", res.Summary.Added)
	}
	if res.Summary.Renamed != 1 {
		t.Fatalf("expected 1 renamed, got %d", res.Summary.Renamed)
	}
	if res.Summary.OrderChanged != 1 {
		t.Fatalf("expected 1 order change, got %d", res.Summary.OrderChanged)
	}
}

func TestSemanticDiffMoveOnly(t *testing.T) {
	td := t.TempDir()
	repo, err := git.NewRepository(git.RepositoryConfig{Path: td})
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("init: %s", env.Message)
	}

	// base: root -> [a]
	writeJSON(t, td, "project.json", `{"rootId":"root","schemaVersion":1}`)
	writeJSON(t, td, "nodes/root.json", `{"id":"root","name":"Root","children":["a"]}`)
	writeJSON(t, td, "nodes/a.json", `{"id":"a","name":"A","children":[]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "initial", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit1: %s", env.Message)
	}

	// move a under b (create b)
	writeJSON(t, td, "nodes/b.json", `{"id":"b","name":"B","children":["a"]}`)
	writeJSON(t, td, "nodes/root.json", `{"id":"root","name":"Root","children":["b"]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "second", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit2: %s", env.Message)
	}

	res, env := Diff(td, "HEAD~1", "HEAD")
	if env.Code != "" {
		t.Fatalf("diff: %s", env.Message)
	}
	if res.Summary.Moved != 1 {
		t.Fatalf("expected moved=1, got %d", res.Summary.Moved)
	}
}

func TestSemanticDiffRenameAndProperty(t *testing.T) {
	td := t.TempDir()
	repo, err := git.NewRepository(git.RepositoryConfig{Path: td})
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("init: %s", env.Message)
	}

	writeJSON(t, td, "project.json", `{"rootId":"root","schemaVersion":1}`)
	writeJSON(t, td, "nodes/root.json", `{"id":"root","name":"Root","children":["a"]}`)
	writeJSON(t, td, "nodes/a.json", `{"id":"a","name":"Alpha","description":"d1","properties":{"x":{"typeHint":"string","value":"1"}},"children":[]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "initial", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit1: %s", env.Message)
	}

	// rename + change description and property
	writeJSON(t, td, "nodes/a.json", `{"id":"a","name":"AlphaRenamed","description":"d2","properties":{"x":{"typeHint":"string","value":"2"}},"children":[]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "second", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit2: %s", env.Message)
	}

	res, env := Diff(td, "HEAD~1", "HEAD")
	if env.Code != "" {
		t.Fatalf("diff: %s", env.Message)
	}
	if res.Summary.Renamed != 1 {
		t.Fatalf("expected renamed=1, got %d", res.Summary.Renamed)
	}
	if res.Summary.PropertyChanged != 1 {
		t.Fatalf("expected propertyChanged=1, got %d", res.Summary.PropertyChanged)
	}
}

func TestSemanticDiffReorderOnlySameSet(t *testing.T) {
	td := t.TempDir()
	repo, err := git.NewRepository(git.RepositoryConfig{Path: td})
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("init: %s", env.Message)
	}

	writeJSON(t, td, "project.json", `{"rootId":"root","schemaVersion":1}`)
	writeJSON(t, td, "nodes/root.json", `{"id":"root","name":"Root","children":["a","b","c"]}`)
	writeJSON(t, td, "nodes/a.json", `{"id":"a","name":"A","children":[]}`)
	writeJSON(t, td, "nodes/b.json", `{"id":"b","name":"B","children":[]}`)
	writeJSON(t, td, "nodes/c.json", `{"id":"c","name":"C","children":[]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "initial", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit1: %s", env.Message)
	}

	// reorder only
	writeJSON(t, td, "nodes/root.json", `{"id":"root","name":"Root","children":["c","a","b"]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "second", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit2: %s", env.Message)
	}

	res, env := Diff(td, "HEAD~1", "HEAD")
	if env.Code != "" {
		t.Fatalf("diff: %s", env.Message)
	}
	if res.Summary.OrderChanged != 1 {
		t.Fatalf("expected orderChanged=1, got %d", res.Summary.OrderChanged)
	}
	if res.Summary.Moved != 0 {
		t.Fatalf("expected moved=0, got %d", res.Summary.Moved)
	}
	if res.Summary.Renamed != 0 {
		t.Fatalf("expected renamed=0, got %d", res.Summary.Renamed)
	}
}

func TestSemanticDiffPropertyTypeChange(t *testing.T) {
	td := t.TempDir()
	repo, err := git.NewRepository(git.RepositoryConfig{Path: td})
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("init: %s", env.Message)
	}

	writeJSON(t, td, "project.json", `{"rootId":"root","schemaVersion":1}`)
	writeJSON(t, td, "nodes/root.json", `{"id":"root","name":"Root","children":["a"]}`)
	writeJSON(t, td, "nodes/a.json", `{"id":"a","name":"A","properties":{"x":{"typeHint":"string","value":"1"}},"children":[]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "initial", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit1: %s", env.Message)
	}

	// change type from string to number
	writeJSON(t, td, "nodes/a.json", `{"id":"a","name":"A","properties":{"x":{"typeHint":"number","value":2}},"children":[]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "second", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit2: %s", env.Message)
	}

	res, env := Diff(td, "HEAD~1", "HEAD")
	if env.Code != "" {
		t.Fatalf("diff: %s", env.Message)
	}
	if res.Summary.PropertyChanged != 1 {
		t.Fatalf("expected propertyChanged=1, got %d", res.Summary.PropertyChanged)
	}
}

func TestSemanticDiffToleratesMalformedNode(t *testing.T) {
	td := t.TempDir()
	repo, err := git.NewRepository(git.RepositoryConfig{Path: td})
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("init: %s", env.Message)
	}

	writeJSON(t, td, "project.json", `{"rootId":"root","schemaVersion":1}`)
	writeJSON(t, td, "nodes/root.json", `{"id":"root","name":"Root","children":["a"]}`)
	// malformed a.json
	writeJSON(t, td, "nodes/a.json", `{"id":"a","name":"A","children":[]`) // missing brace
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "initial", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit1: %s", env.Message)
	}

	// fix a.json
	writeJSON(t, td, "nodes/a.json", `{"id":"a","name":"A","children":[]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "second", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit2: %s", env.Message)
	}

	if _, env := Diff(td, "HEAD~1", "HEAD"); env.Code != "" {
		t.Fatalf("diff failed: %s", env.Message)
	}
}
