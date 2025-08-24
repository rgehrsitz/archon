package plugins

import (
	"context"
	"os"
	"testing"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/git"
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
    pm.DeclarePermissions(pluginID, []Permission{PermissionReadRepo, PermissionWriteRepo, PermissionIndexWrite})

    logger := logging.NewTestLogger()
    // gitRepo: nil (not used by read paths)
    hs = NewHostService(logger, ns, nil, idx, pm, nil, nil)

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

// --- Write-path tests ---

// gitRepoStub implements git.Repository for tests
type gitRepoStub struct{ commitHash string }

func (g *gitRepoStub) IsRepository() bool { return true }
func (g *gitRepoStub) Init(ctx context.Context) errors.Envelope { return errors.Envelope{} }
func (g *gitRepoStub) GetRemoteURL(remote string) (string, errors.Envelope) { return "", errors.Envelope{} }
func (g *gitRepoStub) SetRemoteURL(remote, url string) errors.Envelope { return errors.Envelope{} }
func (g *gitRepoStub) Status(ctx context.Context) (*git.Status, errors.Envelope) { return &git.Status{}, errors.Envelope{} }
func (g *gitRepoStub) GetCurrentBranch(ctx context.Context) (string, errors.Envelope) { return "main", errors.Envelope{} }
func (g *gitRepoStub) GetCommitHistory(ctx context.Context, limit int) ([]git.Commit, errors.Envelope) {
    return nil, errors.Envelope{}
}
func (g *gitRepoStub) Clone(ctx context.Context, url, path string) errors.Envelope { return errors.Envelope{} }
func (g *gitRepoStub) Fetch(ctx context.Context, remote string) errors.Envelope { return errors.Envelope{} }
func (g *gitRepoStub) Pull(ctx context.Context, remote, branch string) errors.Envelope { return errors.Envelope{} }
func (g *gitRepoStub) Push(ctx context.Context, remote, branch string) errors.Envelope { return errors.Envelope{} }
func (g *gitRepoStub) Add(ctx context.Context, paths []string) errors.Envelope { return errors.Envelope{} }
func (g *gitRepoStub) Commit(ctx context.Context, message string, author *git.Author) (*git.Commit, errors.Envelope) {
    return &git.Commit{Hash: g.commitHash}, errors.Envelope{}
}
func (g *gitRepoStub) CreateTag(ctx context.Context, name, message string) errors.Envelope { return errors.Envelope{} }
func (g *gitRepoStub) ListTags(ctx context.Context) ([]git.Tag, errors.Envelope) { return nil, errors.Envelope{} }
func (g *gitRepoStub) Checkout(ctx context.Context, ref string) errors.Envelope { return errors.Envelope{} }
func (g *gitRepoStub) InitLFS(ctx context.Context) errors.Envelope { return errors.Envelope{} }
func (g *gitRepoStub) IsLFSEnabled(ctx context.Context) (bool, errors.Envelope) { return false, errors.Envelope{} }
func (g *gitRepoStub) TrackLFSPattern(ctx context.Context, pattern string) errors.Envelope { return errors.Envelope{} }
func (g *gitRepoStub) GetDiff(ctx context.Context, from, to string) (*git.Diff, errors.Envelope) { return &git.Diff{}, errors.Envelope{} }
func (g *gitRepoStub) Close() error { return nil }

func TestHostService_ApplyMutations_WritePermissionEnforced(t *testing.T) {
    basePath, rootID, _, ns, idx, pm, hs, pluginID := setupHostTest(t)
    _ = basePath

    ctx := context.Background()

    // Create mutation under root
    muts := []*Mutation{
        {
            Type:     MutationCreate,
            ParentID: rootID,
            Data: &NodeData{
                Name:        "New Child",
                Description: "created by test",
                Properties:  map[string]any{"x": 1},
            },
        },
    }

    // Without permission -> unauthorized
    env := hs.ApplyMutations(ctx, pluginID, muts)
    if env.Code == "" || env.Code != "UNAUTHORIZED" {
        t.Fatalf("expected UNAUTHORIZED, got: %+v", env)
    }

    // Grant write permission and apply
    env = pm.GrantPermission(pluginID, PermissionWriteRepo, false, 0)
    if env.Code != "" {
        t.Fatalf("failed to grant write permission: %+v", env)
    }
    env = hs.ApplyMutations(ctx, pluginID, muts)
    if env.Code != "" {
        t.Fatalf("unexpected error applying mutations: %+v", env)
    }

    // Verify child count increased via HostService (grant read)
    env = pm.GrantPermission(pluginID, PermissionReadRepo, false, 0)
    if env.Code != "" {
        t.Fatalf("failed to grant read permission: %+v", env)
    }
    kids, env2 := hs.ListChildren(ctx, pluginID, rootID)
    if env2.Code != "" {
        t.Fatalf("unexpected error listing children: %+v", env2)
    }
    if len(kids) != 2 {
        // One child from setup + one created now
        t.Fatalf("expected 2 children, got %d", len(kids))
    }

    _ = ns
    _ = idx
}

func TestHostService_Commit_WritePermissionEnforced(t *testing.T) {
    basePath, _, _, ns, idx, pm, hs, pluginID := setupHostTest(t)
    _ = basePath
    _ = ns
    _ = idx

    ctx := context.Background()

    // Without permission -> unauthorized
    hash, env := hs.Commit(ctx, pluginID, "test commit")
    if env.Code == "" || env.Code != "UNAUTHORIZED" {
        t.Fatalf("expected UNAUTHORIZED, got: %+v", env)
    }
    if hash != "" {
        t.Fatalf("expected empty hash on unauthorized commit")
    }

    // With permission -> success using stub git repo
    env = pm.GrantPermission(pluginID, PermissionWriteRepo, false, 0)
    if env.Code != "" {
        t.Fatalf("failed to grant write permission: %+v", env)
    }
    stub := &gitRepoStub{commitHash: "abc123"}
    logger := logging.NewTestLogger()
    hs2 := NewHostService(logger, ns, stub, idx, pm, nil, nil)

    hash, env = hs2.Commit(ctx, pluginID, "") // empty message allowed
    if env.Code != "" {
        t.Fatalf("unexpected error committing: %+v", env)
    }
    if hash != "abc123" {
        t.Fatalf("expected hash 'abc123', got %q", hash)
    }
}

func TestHostService_Snapshot_WritePermissionEnforced(t *testing.T) {
    basePath, _, _, ns, idx, pm, hs, pluginID := setupHostTest(t)
    _ = basePath

    ctx := context.Background()

    // Without permission -> unauthorized
    hash, env := hs.Snapshot(ctx, pluginID, "test snapshot")
    if env.Code == "" || env.Code != "UNAUTHORIZED" {
        t.Fatalf("expected UNAUTHORIZED, got: %+v", env)
    }
    if hash != "" {
        t.Fatalf("expected empty hash on unauthorized snapshot")
    }

    // With permission -> success using stub git repo
    env = pm.GrantPermission(pluginID, PermissionWriteRepo, false, 0)
    if env.Code != "" {
        t.Fatalf("failed to grant write permission: %+v", env)
    }
    stub := &gitRepoStub{commitHash: "snap001"}
    logger := logging.NewTestLogger()
    hs2 := NewHostService(logger, ns, stub, idx, pm, nil, nil)

    hash, env = hs2.Snapshot(ctx, pluginID, "")
    if env.Code != "" {
        t.Fatalf("unexpected error snapshotting: %+v", env)
    }
    if hash != "snap001" {
        t.Fatalf("expected hash 'snap001', got %q", hash)
    }
}

func TestHostService_IndexPut_IndexWritePermissionEnforced(t *testing.T) {
    basePath, _, childID, _, idx, pm, hs, pluginID := setupHostTest(t)
    _ = basePath
    _ = idx

    ctx := context.Background()

    // Without permission -> unauthorized
    env := hs.IndexPut(ctx, pluginID, childID, "content")
    if env.Code == "" || env.Code != "UNAUTHORIZED" {
        t.Fatalf("expected UNAUTHORIZED, got: %+v", env)
    }

    // Grant permission -> success (index disabled -> no-op)
    env = pm.GrantPermission(pluginID, PermissionIndexWrite, false, 0)
    if env.Code != "" {
        t.Fatalf("failed to grant indexWrite permission: %+v", env)
    }
    env = hs.IndexPut(ctx, pluginID, childID, "content")
    if env.Code != "" {
        t.Fatalf("unexpected error: %+v", env)
    }
}
