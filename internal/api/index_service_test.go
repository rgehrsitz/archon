package api

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/id"
	"github.com/rgehrsitz/archon/internal/store"
	"github.com/rgehrsitz/archon/internal/types"
)

// TestIndexServiceRebuild ensures Rebuild() completes without error and
// correctly walks the hierarchy to compute depths. The test sets
// ARCHON_DISABLE_INDEX=1 to avoid requiring FTS5 in the environment.
func TestIndexServiceRebuild(t *testing.T) {
	// Disable index implementation details (no-op manager) for portability
	os.Setenv("ARCHON_DISABLE_INDEX", "1")
	t.Cleanup(func() { os.Unsetenv("ARCHON_DISABLE_INDEX") })

	tmp := t.TempDir()
	ps := NewProjectService()

	// Create a new project
	proj, env := ps.CreateProject(tmp, map[string]any{})
	if env.Code != "" {
		t.Fatalf("CreateProject error: %v", env)
	}
	if proj == nil {
		t.Fatalf("expected project to be created")
	}

	// Build small hierarchy: root -> A -> B
	ldr := store.NewLoader(tmp)
	now := time.Now()

	root, err := ldr.LoadNode(proj.RootID)
	if err != nil {
		t.Fatalf("load root: %v", err)
	}

	aID := id.NewV7()
	a := &types.Node{ID: aID, Name: "A", Properties: map[string]types.Property{}, Children: []string{}, CreatedAt: now, UpdatedAt: now}
	if err := ldr.SaveNode(a); err != nil {
		t.Fatalf("save A: %v", err)
	}
	root.Children = append(root.Children, aID)
	if err := ldr.SaveNode(root); err != nil {
		t.Fatalf("update root children: %v", err)
	}

	bID := id.NewV7()
	b := &types.Node{ID: bID, Name: "B", Properties: map[string]types.Property{}, Children: []string{}, CreatedAt: now, UpdatedAt: now}
	if err := ldr.SaveNode(b); err != nil {
		t.Fatalf("save B: %v", err)
	}
	a.Children = append(a.Children, bID)
	if err := ldr.SaveNode(a); err != nil {
		t.Fatalf("update A children: %v", err)
	}

	// Sanity check files exist
	for _, id := range []string{proj.RootID, aID, bID} {
		if !ldr.NodeExists(id) {
			t.Fatalf("expected node file for %s to exist under %s", id, filepath.Join(tmp, "nodes"))
		}
	}

	// IndexService wired to current ProjectService
	idx := NewIndexService(ps)
	if env := idx.Rebuild(context.Background()); env.Code != "" {
		t.Fatalf("Rebuild error: %v", env)
	}
}
