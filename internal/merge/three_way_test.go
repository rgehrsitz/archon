package merge

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/git"
	"github.com/rgehrsitz/archon/internal/store"
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

// Test applying non-conflicting changes using the existing working scenario  
func TestThreeWay_ApplyChanges(t *testing.T) {
	td := t.TempDir()
	repo, err := git.NewRepository(git.RepositoryConfig{Path: td})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = repo.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("init: %s", env.Message)
	}
	
	// Use proper UUIDs for node IDs
	rootID := "01234567-0123-7abc-0123-0123456789ab"
	nodeAID := "01234567-0123-7abc-0123-0123456789cd"
	nodeBID := "01234567-0123-7abc-0123-0123456789ef"
	
	// base
	mustWrite(t, td+"/project.json", `{"rootId":"`+rootID+`","schemaVersion":1}`)
	mustWrite(t, td+"/nodes/"+rootID+".json", `{"id":"`+rootID+`","name":"Root","children":["`+nodeAID+`","`+nodeBID+`"]}`)
	mustWrite(t, td+"/nodes/"+nodeAID+".json", `{"id":"`+nodeAID+`","name":"A","children":[],"properties":{"tag":{"value":"original"}}}`)
	mustWrite(t, td+"/nodes/"+nodeBID+".json", `{"id":"`+nodeBID+`","name":"B","children":[]}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "base", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit base: %s", env.Message)
	}
	
	// ours: rename a -> A_Renamed
	mustWrite(t, td+"/nodes/"+nodeAID+".json", `{"id":"`+nodeAID+`","name":"A_Renamed","children":[],"properties":{"tag":{"value":"original"}}}`)
	_ = repo.Add(ctx, []string{"-A"})
	if _, env := repo.Commit(ctx, "ours", &git.Author{Name: "T", Email: "t@e"}); env.Code != "" {
		t.Fatalf("commit ours: %s", env.Message)
	}
	
	// theirs: starting from base, change property value (simulated by extra commit) 
	mustWrite(t, td+"/nodes/"+nodeAID+".json", `{"id":"`+nodeAID+`","name":"A_Renamed","children":[],"properties":{"tag":{"value":"updated"}}}`)
	mustWrite(t, td+"/nodes/"+nodeBID+".json", `{"id":"`+nodeBID+`","name":"B_Moved","children":[]}`)
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

	// Reset to base to test applying changes
	if env := repo.Checkout(ctx, "HEAD~2"); env.Code != "" {
		t.Fatalf("reset to base: %s", env.Message)
	}

	// Verify we're in base state
	loader := store.NewLoader(td)
	node, err := loader.LoadNode(nodeAID)
	if err != nil {
		t.Fatalf("failed to load node a: %v", err)
	}
	if node.Name != "A" {
		t.Errorf("expected base state name 'A', got '%s'", node.Name)
	}

	// Apply the merge
	if err := res.Apply(td); err != nil {
		t.Fatalf("failed to apply changes: %v", err)
	}

	// Verify changes were applied
	node, err = loader.LoadNode(nodeAID)
	if err != nil {
		t.Fatalf("failed to load node a after apply: %v", err)
	}
	
	// Should have the rename from ours
	if node.Name != "A_Renamed" {
		t.Errorf("expected applied name 'A_Renamed', got '%s'", node.Name)
	}
	
	// Should have the property update from theirs  
	tagProp, exists := node.Properties["tag"]
	if !exists {
		t.Errorf("expected property 'tag' to exist")
	} else if tagProp.Value != "updated" {
		t.Errorf("expected property tag='updated', got '%v'", tagProp.Value)
	}

	// Verify b was also renamed
	nodeB, err := loader.LoadNode(nodeBID)
	if err != nil {
		t.Fatalf("failed to load node b: %v", err)
	}
	if nodeB.Name != "B_Moved" {
		t.Errorf("expected node B name 'B_Moved', got '%s'", nodeB.Name)
	}

	if len(res.Applied) == 0 {
		t.Errorf("expected some applied changes, got %d", len(res.Applied))
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

// Test conflict detection when both sides rename the same node differently  
func TestThreeWay_ConflictOnSameNodeRename(t *testing.T) {
	td := t.TempDir()
	repo := setupTestRepo(t, td)
	defer func() { _ = repo.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Base commit with node A
	mustWrite(t, td+"/project.json", `{"rootId":"root","schemaVersion":1}`)
	mustWrite(t, td+"/nodes/root.json", `{"id":"root","name":"Root","children":["a"]}`)
	mustWrite(t, td+"/nodes/a.json", `{"id":"a","name":"A","children":[]}`)
	baseCommit, _ := commitAndGetHash(t, repo, ctx, "base")

	// Ours: rename A -> Alpha
	mustWrite(t, td+"/nodes/a.json", `{"id":"a","name":"Alpha","children":[]}`)
	oursCommit, _ := commitAndGetHash(t, repo, ctx, "ours")

	// Reset to base and create theirs: rename A -> Aaron
	checkoutCommit(t, repo, ctx, baseCommit)
	mustWrite(t, td+"/nodes/a.json", `{"id":"a","name":"Aaron","children":[]}`)
	theirsCommit, _ := commitAndGetHash(t, repo, ctx, "theirs")

	res, err := ThreeWay(td, baseCommit, oursCommit, theirsCommit)
	if err != nil {
		t.Fatalf("three-way: %v", err)
	}

	if len(res.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(res.Conflicts))
	}

	if res.Conflicts[0].Field != "a|name" {
		t.Errorf("expected conflict on 'a|name', got '%s'", res.Conflicts[0].Field)
	}
}

// Test conflict detection when both sides move the same node to different parents
func TestThreeWay_ConflictOnSameNodeMove(t *testing.T) {
	td := t.TempDir()
	repo := setupTestRepo(t, td)
	defer func() { _ = repo.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Use proper UUIDs for node IDs
	rootID := "01234567-0123-7abc-0123-0123456789ab"
	nodeAID := "01234567-0123-7abc-0123-0123456789cd"
	nodeBID := "01234567-0123-7abc-0123-0123456789ef"
	nodeCID := "01234567-0123-7abc-0123-0123456789gh"

	// Base commit with nodes A, B, C under root
	mustWrite(t, td+"/project.json", `{"rootId":"`+rootID+`","schemaVersion":1}`)
	mustWrite(t, td+"/nodes/"+rootID+".json", `{"id":"`+rootID+`","name":"Root","children":["`+nodeAID+`","`+nodeBID+`","`+nodeCID+`"]}`)
	mustWrite(t, td+"/nodes/"+nodeAID+".json", `{"id":"`+nodeAID+`","name":"A","children":[]}`)
	mustWrite(t, td+"/nodes/"+nodeBID+".json", `{"id":"`+nodeBID+`","name":"B","children":[]}`)
	mustWrite(t, td+"/nodes/"+nodeCID+".json", `{"id":"`+nodeCID+`","name":"C","children":[]}`)
	baseCommit, _ := commitAndGetHash(t, repo, ctx, "base")

	// Ours: move C under A
	mustWrite(t, td+"/nodes/"+rootID+".json", `{"id":"`+rootID+`","name":"Root","children":["`+nodeAID+`","`+nodeBID+`"]}`)
	mustWrite(t, td+"/nodes/"+nodeAID+".json", `{"id":"`+nodeAID+`","name":"A","children":["`+nodeCID+`"]}`)
	oursCommit, _ := commitAndGetHash(t, repo, ctx, "ours")

	// Reset to base and create theirs: move C under B
	checkoutCommit(t, repo, ctx, baseCommit)
	mustWrite(t, td+"/nodes/"+rootID+".json", `{"id":"`+rootID+`","name":"Root","children":["`+nodeAID+`","`+nodeBID+`"]}`)
	mustWrite(t, td+"/nodes/"+nodeBID+".json", `{"id":"`+nodeBID+`","name":"B","children":["`+nodeCID+`"]}`)
	theirsCommit, _ := commitAndGetHash(t, repo, ctx, "theirs")

	res, err := ThreeWay(td, baseCommit, oursCommit, theirsCommit)
	if err != nil {
		t.Fatalf("three-way: %v", err)
	}

	if len(res.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(res.Conflicts))
	}

	if res.Conflicts[0].Field != nodeCID+"|parent" {
		t.Errorf("expected conflict on '%s|parent', got '%s'", nodeCID, res.Conflicts[0].Field)
	}
}

// Test successful application of non-conflicting changes
func TestThreeWay_ApplyNonConflictingChanges(t *testing.T) {
	td := t.TempDir()
	repo := setupTestRepo(t, td)
	defer func() { _ = repo.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Use proper UUIDs for node IDs
	rootID := "01234567-0123-7abc-0123-0123456789ab"
	nodeAID := "01234567-0123-7abc-0123-0123456789cd"
	nodeBID := "01234567-0123-7abc-0123-0123456789ef"

	// Base commit with nodes A, B
	mustWrite(t, td+"/project.json", `{"rootId":"`+rootID+`","schemaVersion":1}`)
	mustWrite(t, td+"/nodes/"+rootID+".json", `{"id":"`+rootID+`","name":"Root","children":["`+nodeAID+`","`+nodeBID+`"]}`)
	mustWrite(t, td+"/nodes/"+nodeAID+".json", `{"id":"`+nodeAID+`","name":"A","children":[],"properties":{"tag":{"value":"old"}}}`)
	mustWrite(t, td+"/nodes/"+nodeBID+".json", `{"id":"`+nodeBID+`","name":"B","children":[]}`)
	baseCommit, _ := commitAndGetHash(t, repo, ctx, "base")

	// Ours: rename A -> Alpha 
	mustWrite(t, td+"/nodes/"+nodeAID+".json", `{"id":"`+nodeAID+`","name":"Alpha","children":[],"properties":{"tag":{"value":"old"}}}`)
	oursCommit, _ := commitAndGetHash(t, repo, ctx, "ours")

	// Reset to base and create theirs: update A's property
	checkoutCommit(t, repo, ctx, baseCommit)
	mustWrite(t, td+"/nodes/"+nodeAID+".json", `{"id":"`+nodeAID+`","name":"A","children":[],"properties":{"tag":{"value":"new"}}}`)
	theirsCommit, _ := commitAndGetHash(t, repo, ctx, "theirs")

	res, err := ThreeWay(td, baseCommit, oursCommit, theirsCommit)
	if err != nil {
		t.Fatalf("three-way: %v", err)
	}

	if len(res.Conflicts) != 0 {
		t.Fatalf("expected 0 conflicts, got %d", len(res.Conflicts))
	}

	if len(res.OursOnly) != 1 {
		t.Fatalf("expected 1 ours-only change, got %d", len(res.OursOnly))
	}

	if len(res.TheirsOnly) != 1 {
		t.Fatalf("expected 1 theirs-only change, got %d", len(res.TheirsOnly))
	}

	// Reset to base state to test applying changes
	checkoutCommit(t, repo, ctx, baseCommit)

	// Apply the non-conflicting changes
	if err := res.Apply(td); err != nil {
		t.Fatalf("failed to apply changes: %v", err)
	}

	// Verify the applied changes
	loader := store.NewLoader(td)
	
	// Should have both rename (Alpha) and property update (new)
	node, err := loader.LoadNode(nodeAID)
	if err != nil {
		t.Fatalf("failed to load node a: %v", err)
	}

	if node.Name != "Alpha" {
		t.Errorf("expected node name 'Alpha', got '%s'", node.Name)
	}

	tagProp, exists := node.Properties["tag"]
	if !exists {
		t.Errorf("expected property 'tag' to exist")
	} else if tagProp.Value != "new" {
		t.Errorf("expected property tag='new', got '%v'", tagProp.Value)
	}

	if len(res.Applied) != 2 {
		t.Errorf("expected 2 applied changes, got %d", len(res.Applied))
	}
}

// Test property conflict detection
func TestThreeWay_PropertyConflict(t *testing.T) {
	td := t.TempDir()
	repo := setupTestRepo(t, td)
	defer func() { _ = repo.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Use proper UUIDs for node IDs
	rootID := "01234567-0123-7abc-0123-0123456789ab"
	nodeAID := "01234567-0123-7abc-0123-0123456789cd"

	// Base commit with node A having a property
	mustWrite(t, td+"/project.json", `{"rootId":"`+rootID+`","schemaVersion":1}`)
	mustWrite(t, td+"/nodes/"+rootID+".json", `{"id":"`+rootID+`","name":"Root","children":["`+nodeAID+`"]}`)
	mustWrite(t, td+"/nodes/"+nodeAID+".json", `{"id":"`+nodeAID+`","name":"A","children":[],"properties":{"tag":{"value":"original"}}}`)
	baseCommit, _ := commitAndGetHash(t, repo, ctx, "base")

	// Ours: update property to "ours"
	mustWrite(t, td+"/nodes/"+nodeAID+".json", `{"id":"`+nodeAID+`","name":"A","children":[],"properties":{"tag":{"value":"ours"}}}`)
	oursCommit, _ := commitAndGetHash(t, repo, ctx, "ours")

	// Reset to base and create theirs: update property to "theirs"
	checkoutCommit(t, repo, ctx, baseCommit)
	mustWrite(t, td+"/nodes/"+nodeAID+".json", `{"id":"`+nodeAID+`","name":"A","children":[],"properties":{"tag":{"value":"theirs"}}}`)
	theirsCommit, _ := commitAndGetHash(t, repo, ctx, "theirs")

	res, err := ThreeWay(td, baseCommit, oursCommit, theirsCommit)
	if err != nil {
		t.Fatalf("three-way: %v", err)
	}

	if len(res.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(res.Conflicts))
	}

	expectedField := nodeAID + "|property:tag"
	if res.Conflicts[0].Field != expectedField {
		t.Errorf("expected conflict on '%s', got '%s'", expectedField, res.Conflicts[0].Field)
	}
}

// Helper functions for testing
func setupTestRepo(t *testing.T, td string) git.Repository {
	t.Helper()
	repo, err := git.NewRepository(git.RepositoryConfig{Path: td})
	if err != nil {
		t.Fatal(err)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	if env := repo.Init(ctx); env.Code != "" {
		t.Fatalf("init: %s", env.Message)
	}
	return repo
}

func commitAndGetHash(t *testing.T, repo git.Repository, ctx context.Context, message string) (string, *git.Commit) {
	t.Helper()
	_ = repo.Add(ctx, []string{"-A"})
	commit, env := repo.Commit(ctx, message, &git.Author{Name: "T", Email: "t@e"})
	if env.Code != "" {
		t.Fatalf("commit %s: %s", message, env.Message)
	}
	return commit.Hash, commit
}

func checkoutCommit(t *testing.T, repo git.Repository, ctx context.Context, commitHash string) {
	t.Helper()
	if env := repo.Checkout(ctx, commitHash); env.Code != "" {
		t.Fatalf("checkout %s: %s", commitHash, env.Message)
	}
}
