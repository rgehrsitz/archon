package store

import (
	"os"
	"testing"

	"github.com/rgehrsitz/archon/internal/index/sqlite"
	"github.com/rgehrsitz/archon/internal/types"
)

// TestMoveNode_IndexEnabled_ReindexesSubtree verifies that moving a node triggers
// reindexing of the moved node and its descendants, reflected in updated depths.
// Skips when the index is disabled/unavailable on the platform.
func TestMoveNode_IndexEnabled_ReindexesSubtree(t *testing.T) {
	// Ensure index not disabled by env
	os.Unsetenv("ARCHON_DISABLE_INDEX")

	tmp := t.TempDir()
	ps, err := NewProjectStore(tmp)
	if err != nil {
		t.Fatalf("NewProjectStore: %v", err)
	}
	defer ps.Close()

	// Skip if index manager is disabled/unhealthy
	if ps.IndexManager == nil || ps.IndexManager.Health() != nil {
		t.Skipf("index not available on this platform; skipping")
	}
	if ver, _ := ps.IndexManager.GetSchemaVersion(); ver == "0" {
		t.Skipf("index disabled (schema version=0); skipping")
	}

	// Create project
	proj, envErr := ps.CreateProject(map[string]any{})
	if envErr != nil {
		t.Fatalf("CreateProject: %v", envErr)
	}

	ns := NewNodeStore(tmp, ps.IndexManager)

	// Build hierarchy: root -> A -> B -> C
	a, _ := createChild(t, ns, proj.RootID, "A")
	b, _ := createChild(t, ns, a.ID, "B")
	c, _ := createChild(t, ns, b.ID, "C")

	// Sanity: before move, B should be at depth 2 and C at depth 3
	// We'll proceed to move B under root and then verify depths change to 1 and 2

	// Move B from A to root
	moveReq := &types.MoveNodeRequest{NodeID: b.ID, NewParentID: proj.RootID, Position: -1}
	if err := ns.MoveNode(moveReq); err != nil {
		t.Fatalf("MoveNode: %v", err)
	}

	// Query index for depths
	depth1, err := ps.IndexManager.GetNodesByDepth(1, 100)
	if err != nil {
		t.Fatalf("GetNodesByDepth(1): %v", err)
	}
	depth2, err := ps.IndexManager.GetNodesByDepth(2, 100)
	if err != nil {
		t.Fatalf("GetNodesByDepth(2): %v", err)
	}

	if !containsNode(depth1, b.ID) {
		t.Fatalf("expected B to be at depth 1 after move")
	}
	if !containsNode(depth2, c.ID) {
		t.Fatalf("expected C to be at depth 2 after move")
	}
}

// createChild helper to create a child under parent and return node
func createChild(t *testing.T, ns *NodeStore, parentID, name string) (*types.Node, error) {
	t.Helper()
	n, err := ns.CreateNode(&types.CreateNodeRequest{ParentID: parentID, Name: name, Properties: map[string]types.Property{}})
	if err != nil {
		t.Fatalf("CreateNode %s: %v", name, err)
	}
	return n, nil
}

// containsNode checks if a node with id exists in search results
func containsNode(results []sqlite.SearchResult, id string) bool {
	for _, r := range results {
		if r.NodeID == id {
			return true
		}
	}
	return false
}
