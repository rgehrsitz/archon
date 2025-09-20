package store

import (
	"testing"

	"github.com/rgehrsitz/archon/internal/types"
)

// TestReorderChildren_NoIndex ensures ReorderChildren returns successfully when index is disabled.
func TestReorderChildren_NoIndex(t *testing.T) {

	tmp := t.TempDir()
	ps, err := NewProjectStore(tmp)
	if err != nil {
		t.Fatalf("NewProjectStore: %v", err)
	}
	defer ps.Close()

	// Create project + root node already exists via CreateProject
	proj, errEnv := ps.CreateProject(map[string]any{})
	if errEnv != nil {
		t.Fatalf("CreateProject: %v", errEnv)
	}

	ns := NewNodeStore(tmp, ps.IndexManager)

	// Create two children under root
	a, _ := ns.CreateNode(&types.CreateNodeRequest{ParentID: proj.RootID, Name: "A", Properties: map[string]types.Property{}, Description: ""})
	b, _ := ns.CreateNode(&types.CreateNodeRequest{ParentID: proj.RootID, Name: "B", Properties: map[string]types.Property{}, Description: ""})
	if a == nil || b == nil {
		t.Fatalf("expected children to be created")
	}

	// Reorder children to [B, A]
	req := &types.ReorderChildrenRequest{ParentID: proj.RootID, OrderedChildIDs: []string{b.ID, a.ID}}
	if err := ns.ReorderChildren(req); err != nil {
		t.Fatalf("ReorderChildren (no index): %v", err)
	}
}

// TestReorderChildren_IndexEnabled ensures ReorderChildren succeeds and parent is present in index
// when index manager is enabled (skips if FTS5/SQLite not available on platform).
func TestReorderChildren_IndexEnabled(t *testing.T) {
	// Ensure index is not disabled by env

	tmp := t.TempDir()
	ps, err := NewProjectStore(tmp)
	if err != nil {
		t.Fatalf("NewProjectStore: %v", err)
	}
	defer ps.Close()

	// If index manager is disabled or unhealthy, skip
	if ps.IndexManager == nil || ps.IndexManager.Health() != nil {
		t.Skipf("index not available on this platform; skipping")
	}
	if ver, _ := ps.IndexManager.GetSchemaVersion(); ver == "0" {
		t.Skipf("index disabled (schema version=0); skipping")
	}

	// Create project structure
	proj, errEnv := ps.CreateProject(map[string]any{})
	if errEnv != nil {
		t.Fatalf("CreateProject: %v", errEnv)
	}

	ns := NewNodeStore(tmp, ps.IndexManager)
	a, _ := ns.CreateNode(&types.CreateNodeRequest{ParentID: proj.RootID, Name: "A", Properties: map[string]types.Property{}})
	b, _ := ns.CreateNode(&types.CreateNodeRequest{ParentID: proj.RootID, Name: "B", Properties: map[string]types.Property{}})
	if a == nil || b == nil {
		t.Fatalf("expected children to be created")
	}

	// Reorder children
	req := &types.ReorderChildrenRequest{ParentID: proj.RootID, OrderedChildIDs: []string{b.ID, a.ID}}
	if err := ns.ReorderChildren(req); err != nil {
		t.Fatalf("ReorderChildren (index enabled): %v", err)
	}

	// Validate parent exists in index and child_count is consistent (2)
	results, err := ps.IndexManager.GetNodesByDepth(0, 10)
	if err != nil {
		t.Fatalf("GetNodesByDepth: %v", err)
	}
	found := false
	for _, r := range results {
		if r.NodeID == proj.RootID {
			found = true
			if r.ChildCount != 2 {
				t.Fatalf("root child_count in index = %d, want 2", r.ChildCount)
			}
			break
		}
	}
	if !found {
		t.Fatalf("root not found in index results at depth 0")
	}
}
