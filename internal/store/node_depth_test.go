package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/id"
	"github.com/rgehrsitz/archon/internal/types"
)

// TestCalculateDepth verifies depth computation by constructing a small hierarchy on disk
// and invoking the unexported calculateDepth via a NodeStore instance.
func TestCalculateDepth(t *testing.T) {
	t.TempDir() // ensure parallel safety
	tmp := t.TempDir()
	ldr := NewLoader(tmp)

	// Create nodes: root -> A -> B
	rootID := id.NewV7()
	now := time.Now()
	root := &types.Node{ID: rootID, Name: "Root", Properties: map[string]types.Property{}, Children: []string{}, CreatedAt: now, UpdatedAt: now}
	if err := ldr.SaveNode(root); err != nil {
		t.Fatalf("save root: %v", err)
	}

	aID := id.NewV7()
	a := &types.Node{ID: aID, Name: "A", Properties: map[string]types.Property{}, Children: []string{}, CreatedAt: now, UpdatedAt: now}
	if err := ldr.SaveNode(a); err != nil {
		t.Fatalf("save A: %v", err)
	}
	// Link A under root
	root.Children = append(root.Children, aID)
	if err := ldr.SaveNode(root); err != nil {
		t.Fatalf("update root children: %v", err)
	}

	bID := id.NewV7()
	b := &types.Node{ID: bID, Name: "B", Properties: map[string]types.Property{}, Children: []string{}, CreatedAt: now, UpdatedAt: now}
	if err := ldr.SaveNode(b); err != nil {
		t.Fatalf("save B: %v", err)
	}
	// Link B under A
	a.Children = append(a.Children, bID)
	if err := ldr.SaveNode(a); err != nil {
		t.Fatalf("update A children: %v", err)
	}

	// Create NodeStore; index manager not needed for depth calculation
	ns := NewNodeStore(tmp, nil)

	// Depth for root's children should be 1 when using rootID as parent
	if d := ns.calculateDepth(rootID); d != 1 {
		t.Fatalf("depth(rootID) = %d, want 1", d)
	}
	// Depth for A's children should be 2
	if d := ns.calculateDepth(aID); d != 2 {
		t.Fatalf("depth(aID) = %d, want 2", d)
	}
	// Root depth when no parentID is provided should be 0
	if d := ns.calculateDepth(""); d != 0 {
		t.Fatalf("depth(\"\") = %d, want 0", d)
	}
	// Missing parentID should conservatively return 1 (non-root best-effort)
	missing := id.NewV7()
	if d := ns.calculateDepth(missing); d != 1 {
		t.Fatalf("depth(missing) = %d, want 1", d)
	}

	// Sanity: ensure node files exist
	if !ldr.NodeExists(rootID) || !ldr.NodeExists(aID) || !ldr.NodeExists(bID) {
		t.Fatalf("expected node files to exist under %s", filepath.Join(tmp, "nodes"))
	}
}
