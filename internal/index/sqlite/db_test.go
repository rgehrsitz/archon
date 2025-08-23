package sqlite

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/types"
)

func TestDB_Open(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon-db-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := Open(tempDir)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Check that database file was created
	dbPath := filepath.Join(tempDir, ".archon", "index", "archon.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file was not created at %s", dbPath)
	}

	// Check that we can ping the database
	if err := db.Ping(); err != nil {
		t.Errorf("failed to ping database: %v", err)
	}
}

func TestDB_GetSchemaVersion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon-db-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := Open(tempDir)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	version, err := db.GetSchemaVersion()
	if err != nil {
		t.Fatalf("failed to get schema version: %v", err)
	}

	if version != "1" {
		t.Errorf("expected schema version '1', got '%s'", version)
	}
}

func TestDB_TablesExist(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon-db-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := Open(tempDir)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Check that all expected tables exist
	expectedTables := []string{
		"index_meta",
		"nodes",
		"properties", 
		"nodes_fts",
		"properties_fts",
	}

	for _, tableName := range expectedTables {
		var count int
		err := db.conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
		if err != nil {
			t.Errorf("failed to check for table %s: %v", tableName, err)
			continue
		}
		if count == 0 {
			t.Errorf("table %s does not exist", tableName)
		}
	}
}

func TestIndexWriter_Integration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon-writer-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	db, err := Open(tempDir)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	writer := NewIndexWriter(db)
	
	// Create test node
	node := &types.Node{
		ID:          "01234567-89ab-cdef-0123-456789abcdef",
		Name:        "Test Node",
		Description: "A test node for indexing",
		Properties: map[string]types.Property{
			"status": {TypeHint: "string", Value: "active"},
			"count":  {TypeHint: "number", Value: 42},
		},
		Children:  []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Index the node
	err = writer.IndexNode(node, "", 0)
	if err != nil {
		t.Fatalf("failed to index node: %v", err)
	}

	// Verify node was indexed
	var name, description string
	var depth, childCount int
	err = db.conn.QueryRow("SELECT name, description, depth, child_count FROM nodes WHERE id = ?", node.ID).
		Scan(&name, &description, &depth, &childCount)
	if err != nil {
		t.Fatalf("failed to query indexed node: %v", err)
	}

	if name != node.Name {
		t.Errorf("expected name '%s', got '%s'", node.Name, name)
	}
	if description != node.Description {
		t.Errorf("expected description '%s', got '%s'", node.Description, description)
	}

	// Verify properties were indexed
	var propCount int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM properties WHERE node_id = ?", node.ID).Scan(&propCount)
	if err != nil {
		t.Fatalf("failed to count properties: %v", err)
	}
	if propCount != 2 {
		t.Errorf("expected 2 properties, got %d", propCount)
	}

	// Test removal
	err = writer.RemoveNode(node.ID)
	if err != nil {
		t.Fatalf("failed to remove node: %v", err)
	}

	// Verify node was removed
	err = db.conn.QueryRow("SELECT name FROM nodes WHERE id = ?", node.ID).Scan(&name)
	if err == nil {
		t.Errorf("node was not removed from index")
	}

	// Verify properties were removed
	err = db.conn.QueryRow("SELECT COUNT(*) FROM properties WHERE node_id = ?", node.ID).Scan(&propCount)
	if err != nil {
		t.Fatalf("failed to count properties after removal: %v", err)
	}
	if propCount != 0 {
		t.Errorf("expected 0 properties after removal, got %d", propCount)
	}
}