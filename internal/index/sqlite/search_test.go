package sqlite

import (
	"os"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/types"
)

func setupTestDB(t *testing.T) (*DB, *IndexWriter, *SearchEngine, func()) {
	tempDir, err := os.MkdirTemp("", "archon-search-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	db, err := Open(tempDir)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	writer := NewIndexWriter(db)
	search := NewSearchEngine(db)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return db, writer, search, cleanup
}

func createTestNode(id, name, description string, properties map[string]types.Property) *types.Node {
	return &types.Node{
		ID:          id,
		Name:        name,
		Description: description,
		Properties:  properties,
		Children:    []string{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestSearchEngine_SearchNodes(t *testing.T) {
	_, writer, search, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test nodes
	nodes := []*types.Node{
		createTestNode("node1", "Manufacturing Plant", "Primary production facility", map[string]types.Property{
			"location": {TypeHint: "string", Value: "Detroit"},
		}),
		createTestNode("node2", "Assembly Line A", "Main assembly line for products", map[string]types.Property{
			"capacity": {TypeHint: "number", Value: 100},
		}),
		createTestNode("node3", "Quality Control", "Testing and quality assurance", map[string]types.Property{
			"status": {TypeHint: "string", Value: "operational"},
		}),
	}

	// Index all nodes
	for i, node := range nodes {
		err := writer.IndexNode(node, "", i)
		if err != nil {
			t.Fatalf("failed to index node %s: %v", node.ID, err)
		}
	}

	// Test search by name
	results, err := search.SearchNodes("Manufacturing", 10)
	if err != nil {
		t.Fatalf("failed to search nodes: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'Manufacturing', got %d", len(results))
	}
	if len(results) > 0 && results[0].NodeID != "node1" {
		t.Errorf("expected node1, got %s", results[0].NodeID)
	}

	// Test search by description
	results, err = search.SearchNodes("assembly", 10)
	if err != nil {
		t.Fatalf("failed to search nodes by description: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'assembly', got %d", len(results))
	}

	// Test wildcard search (wildcards are added automatically)
	results, err = search.SearchNodes("line", 10)
	if err != nil {
		t.Fatalf("failed to search with wildcard: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for wildcard search, got %d", len(results))
	}

	// Test search with no results
	results, err = search.SearchNodes("nonexistent", 10)
	if err != nil {
		t.Fatalf("failed to search for nonexistent term: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for nonexistent term, got %d", len(results))
	}
}

func TestSearchEngine_SearchProperties(t *testing.T) {
	_, writer, search, cleanup := setupTestDB(t)
	defer cleanup()

	// Create node with properties
	node := createTestNode("node1", "Test Node", "A test node", map[string]types.Property{
		"location": {TypeHint: "string", Value: "Detroit Michigan"},
		"capacity": {TypeHint: "number", Value: 100},
		"status":   {TypeHint: "string", Value: "operational"},
	})

	err := writer.IndexNode(node, "", 0)
	if err != nil {
		t.Fatalf("failed to index node: %v", err)
	}

	// Test search by property value
	results, err := search.SearchProperties("Detroit", 10)
	if err != nil {
		t.Fatalf("failed to search properties: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for property search, got %d", len(results))
	}
	if len(results) > 0 {
		if results[0].NodeID != "node1" {
			t.Errorf("expected node1, got %s", results[0].NodeID)
		}
		if results[0].Key != "location" {
			t.Errorf("expected key 'location', got %s", results[0].Key)
		}
	}

	// Test search by property key
	results, err = search.SearchProperties("capacity", 10)
	if err != nil {
		t.Fatalf("failed to search properties by key: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for key search, got %d", len(results))
	}
}

func TestSearchEngine_SearchByPath(t *testing.T) {
	_, writer, search, cleanup := setupTestDB(t)
	defer cleanup()

	// Create nodes with hierarchical paths
	nodes := []NodeWithContext{
		{createTestNode("root", "Factory", "Root factory node", nil), "", 0},
		{createTestNode("line1", "Assembly Line 1", "First assembly line", nil), "root", 1},
		{createTestNode("station1", "Station A", "First station", nil), "line1", 2},
	}

	err := writer.RebuildIndex(nodes)
	if err != nil {
		t.Fatalf("failed to rebuild index: %v", err)
	}

	// Test path search
	results, err := search.SearchByPath("Factory", 10)
	if err != nil {
		t.Fatalf("failed to search by path: %v", err)
	}

	// Should find all nodes that have "Factory" in their path
	if len(results) < 1 {
		t.Errorf("expected at least 1 result for path search, got %d", len(results))
	}
}

func TestSearchEngine_GetNodesByDepth(t *testing.T) {
	_, writer, search, cleanup := setupTestDB(t)
	defer cleanup()

	// Create nodes at different depths
	nodes := []NodeWithContext{
		{createTestNode("root1", "Root 1", "First root", nil), "", 0},
		{createTestNode("root2", "Root 2", "Second root", nil), "", 0},
		{createTestNode("child1", "Child 1", "First child", nil), "root1", 1},
		{createTestNode("child2", "Child 2", "Second child", nil), "root2", 1},
		{createTestNode("grandchild1", "Grandchild 1", "First grandchild", nil), "child1", 2},
	}

	err := writer.RebuildIndex(nodes)
	if err != nil {
		t.Fatalf("failed to rebuild index: %v", err)
	}

	// Test depth 0 (root nodes)
	results, err := search.GetNodesByDepth(0, 10)
	if err != nil {
		t.Fatalf("failed to get nodes by depth 0: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 nodes at depth 0, got %d", len(results))
	}

	// Test depth 1 (children)
	results, err = search.GetNodesByDepth(1, 10)
	if err != nil {
		t.Fatalf("failed to get nodes by depth 1: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 nodes at depth 1, got %d", len(results))
	}

	// Test depth 2 (grandchildren)
	results, err = search.GetNodesByDepth(2, 10)
	if err != nil {
		t.Fatalf("failed to get nodes by depth 2: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 node at depth 2, got %d", len(results))
	}
}

func TestSearchEngine_BuildFTSQuery(t *testing.T) {
	_, _, search, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple*"},
		{"two words", "two* words*"},
		{`"exact phrase"`, `"exact phrase"`},
		{"mixed \"exact\" and loose", `mixed* "exact" and* loose*`},
		{"", ""},
		{"  whitespace  ", "whitespace*"},
	}

	for _, test := range tests {
		result := search.buildFTSQuery(test.input)
		if result != test.expected {
			t.Errorf("buildFTSQuery(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}