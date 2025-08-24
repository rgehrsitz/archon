package plugins

import (
	"context"
	"os"
	"testing"

	"github.com/rgehrsitz/archon/internal/index"
	"github.com/rgehrsitz/archon/internal/logging"
	"github.com/rgehrsitz/archon/internal/store"
	"github.com/rgehrsitz/archon/internal/types"
)

// setupHostTest creates a temp project with root + one child, and returns helpers
func setupHostTest(t *testing.T) (basePath string, rootID string, childID string, ns *store.NodeStore, idx *index.Manager, pm *PermissionManager, hs *HostService, pluginID string) {
	t.Helper()

	// Disable index (FTS5) for CI/portable tests
	_ = os.Setenv("ARCHON_DISABLE_INDEX", "1")

	basePath = t.TempDir()
	ps, err := store.NewProjectStore(basePath)
	if err != nil {
		t.Fatalf("failed to create project store: %v", err)
	}
	t.Cleanup(func() { _ = ps.Close() })

	proj, err := ps.CreateProject(map[string]any{"name": "test"})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	rootID = proj.RootID

	idx = ps.IndexManager
	ns = store.NewNodeStore(basePath, idx)

	// Create a child node under root
	child, err := ns.CreateNode(&types.CreateNodeRequest{
		ParentID:    rootID,
		Name:        "Child",
		Description: "child node",
		Properties:  map[string]types.Property{"k": {Value: "v"}},
	})
	if err != nil {
		t.Fatalf("failed to create child node: %v", err)
	}
	childID = child.ID

	pm = NewPermissionManager()
	pluginID = "com.test.plugin"
	pm.DeclarePermissions(pluginID, []Permission{PermissionReadRepo})

	logger := logging.NewTestLogger()
	// gitRepo: nil (not used by read paths)
	hs = NewHostService(logger, ns, nil, idx, pm)

	return
}

func TestHostService_GetNode_ReadPermissionEnforced(t *testing.T) {
	basePath, rootID, childID, _, _, pm, hs, pluginID := setupHostTest(t)
	_ = basePath
	_ = rootID

	ctx := context.Background()

	// Without permission -> unauthorized
	n, env := hs.GetNode(ctx, pluginID, childID)
	if env.Code == "" || env.Code != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got: %+v", env)
	}
	if n != nil {
		t.Fatalf("expected nil node on unauthorized access")
	}

	// Grant read permission
	env = pm.GrantPermission(pluginID, PermissionReadRepo, false, 0)
	if env.Code != "" {
		t.Fatalf("failed to grant permission: %+v", env)
	}

	// With permission -> success
	n, env = hs.GetNode(ctx, pluginID, childID)
	if env.Code != "" {
		t.Fatalf("unexpected error: %+v", env)
	}
	if n == nil || n.ID != childID {
		t.Fatalf("expected node %s, got %#v", childID, n)
	}
}

func TestHostService_ListChildren_ReadPermissionEnforced(t *testing.T) {
	basePath, rootID, _, _, _, pm, hs, pluginID := setupHostTest(t)
	_ = basePath

	ctx := context.Background()

	// Without permission -> unauthorized
	kids, env := hs.ListChildren(ctx, pluginID, rootID)
	if env.Code == "" || env.Code != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got: %+v", env)
	}
	if kids != nil {
		t.Fatalf("expected nil children on unauthorized access")
	}

	// Grant read permission
	env = pm.GrantPermission(pluginID, PermissionReadRepo, false, 0)
	if env.Code != "" {
		t.Fatalf("failed to grant permission: %+v", env)
	}

	// With permission -> success
	kids, env = hs.ListChildren(ctx, pluginID, rootID)
	if env.Code != "" {
		t.Fatalf("unexpected error: %+v", env)
	}
	if len(kids) != 1 {
		t.Fatalf("expected 1 child, got %d", len(kids))
	}
}

func TestHostService_Query_ReadPermissionEnforced(t *testing.T) {
	basePath, _, _, _, _, pm, hs, pluginID := setupHostTest(t)
	_ = basePath

	ctx := context.Background()

	// Without permission -> unauthorized
	results, env := hs.Query(ctx, pluginID, "name:Child", 10)
	if env.Code == "" || env.Code != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got: %+v", env)
	}
	if results != nil {
		t.Fatalf("expected nil results on unauthorized access")
	}

	// Grant read permission
	env = pm.GrantPermission(pluginID, PermissionReadRepo, false, 0)
	if env.Code != "" {
		t.Fatalf("failed to grant permission: %+v", env)
	}

	// With permission -> success; index is disabled so expect no results and no error (nil or empty slice OK)
	results, env = hs.Query(ctx, pluginID, "name:Child", 10)
	if env.Code != "" {
		t.Fatalf("unexpected error: %+v", env)
	}
	if len(results) != 0 {
		// With disabled index, search returns empty; this assertion ensures call succeeded
		t.Fatalf("expected 0 results with disabled index, got %d", len(results))
	}
}
