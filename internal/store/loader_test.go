package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/types"
)

func createTempDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "archon-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return tmpDir
}

func TestLoaderProjectOperations(t *testing.T) {
	tmpDir := createTempDir(t)
	loader := NewLoader(tmpDir)
	
	// Test project doesn't exist initially
	if loader.ProjectExists() {
		t.Error("Project should not exist initially")
	}
	
	// Test loading non-existent project
	_, err := loader.LoadProject()
	if err == nil {
		t.Error("Should fail to load non-existent project")
	}
	
	envelope, ok := err.(errors.Envelope)
	if !ok || envelope.Code != errors.ErrProjectNotFound {
		t.Errorf("Expected PROJECT_NOT_FOUND error, got %v", err)
	}
	
	// Create a project
	project := &types.Project{
		RootID:        "01234567-89ab-cdef-0123-456789abcdef",
		SchemaVersion: types.CurrentSchemaVersion,
		Settings:      map[string]any{"test": "value"},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	// Save project
	err = loader.SaveProject(project)
	if err != nil {
		t.Fatalf("Failed to save project: %v", err)
	}
	
	// Test project exists now
	if !loader.ProjectExists() {
		t.Error("Project should exist after saving")
	}
	
	// Load project
	loadedProject, err := loader.LoadProject()
	if err != nil {
		t.Fatalf("Failed to load project: %v", err)
	}
	
	// Verify project data
	if loadedProject.RootID != project.RootID {
		t.Errorf("Expected root ID %s, got %s", project.RootID, loadedProject.RootID)
	}
	
	if loadedProject.SchemaVersion != project.SchemaVersion {
		t.Errorf("Expected schema version %d, got %d", project.SchemaVersion, loadedProject.SchemaVersion)
	}
	
	if loadedProject.Settings["test"] != "value" {
		t.Errorf("Expected settings test=value, got %v", loadedProject.Settings["test"])
	}
	
	// Test update timestamp is updated on save
	time.Sleep(1 * time.Millisecond)
	originalUpdateTime := loadedProject.UpdatedAt
	
	err = loader.SaveProject(loadedProject)
	if err != nil {
		t.Fatalf("Failed to save project again: %v", err)
	}
	
	reloadedProject, err := loader.LoadProject()
	if err != nil {
		t.Fatalf("Failed to reload project: %v", err)
	}
	
	if !reloadedProject.UpdatedAt.After(originalUpdateTime) {
		t.Error("UpdatedAt should be updated on save")
	}
}

func TestLoaderNodeOperations(t *testing.T) {
	tmpDir := createTempDir(t)
	loader := NewLoader(tmpDir)
	
	nodeID := "01234567-89ab-cdef-0123-456789abcdef"
	
	// Test node doesn't exist initially
	if loader.NodeExists(nodeID) {
		t.Error("Node should not exist initially")
	}
	
	// Test loading non-existent node
	_, err := loader.LoadNode(nodeID)
	if err == nil {
		t.Error("Should fail to load non-existent node")
	}
	
	envelope, ok := err.(errors.Envelope)
	if !ok || envelope.Code != errors.ErrNodeNotFound {
		t.Errorf("Expected NODE_NOT_FOUND error, got %v", err)
	}
	
	// Create a node
	node := &types.Node{
		ID:          nodeID,
		Name:        "Test Node",
		Description: "A test node",
		Properties: map[string]types.Property{
			"test": {TypeHint: "string", Value: "value"},
		},
		Children:  []string{"11111111-2222-3333-4444-555555555555"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Save node
	err = loader.SaveNode(node)
	if err != nil {
		t.Fatalf("Failed to save node: %v", err)
	}
	
	// Test node exists now
	if !loader.NodeExists(nodeID) {
		t.Error("Node should exist after saving")
	}
	
	// Load node
	loadedNode, err := loader.LoadNode(nodeID)
	if err != nil {
		t.Fatalf("Failed to load node: %v", err)
	}
	
	// Verify node data
	if loadedNode.ID != node.ID {
		t.Errorf("Expected ID %s, got %s", node.ID, loadedNode.ID)
	}
	
	if loadedNode.Name != node.Name {
		t.Errorf("Expected name %s, got %s", node.Name, loadedNode.Name)
	}
	
	if loadedNode.Description != node.Description {
		t.Errorf("Expected description %s, got %s", node.Description, loadedNode.Description)
	}
	
	if len(loadedNode.Children) != 1 || loadedNode.Children[0] != "11111111-2222-3333-4444-555555555555" {
		t.Errorf("Expected children [11111111-2222-3333-4444-555555555555], got %v", loadedNode.Children)
	}
	
	if loadedNode.Properties["test"].Value != "value" {
		t.Errorf("Expected property test=value, got %v", loadedNode.Properties["test"].Value)
	}
	
	// Test update timestamp is updated on save
	time.Sleep(1 * time.Millisecond)
	originalUpdateTime := loadedNode.UpdatedAt
	
	err = loader.SaveNode(loadedNode)
	if err != nil {
		t.Fatalf("Failed to save node again: %v", err)
	}
	
	reloadedNode, err := loader.LoadNode(nodeID)
	if err != nil {
		t.Fatalf("Failed to reload node: %v", err)
	}
	
	if !reloadedNode.UpdatedAt.After(originalUpdateTime) {
		t.Error("UpdatedAt should be updated on save")
	}
}

func TestLoaderNodeDeletion(t *testing.T) {
	tmpDir := createTempDir(t)
	loader := NewLoader(tmpDir)
	
	nodeID := "01234567-89ab-cdef-0123-456789abcdef"
	
	// Create and save a node
	node := &types.Node{
		ID:        nodeID,
		Name:      "Test Node",
		Children:  []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	err := loader.SaveNode(node)
	if err != nil {
		t.Fatalf("Failed to save node: %v", err)
	}
	
	// Verify node exists
	if !loader.NodeExists(nodeID) {
		t.Error("Node should exist after saving")
	}
	
	// Delete node
	err = loader.DeleteNode(nodeID)
	if err != nil {
		t.Fatalf("Failed to delete node: %v", err)
	}
	
	// Verify node no longer exists
	if loader.NodeExists(nodeID) {
		t.Error("Node should not exist after deletion")
	}
	
	// Deleting non-existent node should not error
	err = loader.DeleteNode("non-existent-id")
	if err != nil {
		t.Errorf("Deleting non-existent node should not error: %v", err)
	}
}

func TestLoaderListNodeFiles(t *testing.T) {
	tmpDir := createTempDir(t)
	loader := NewLoader(tmpDir)
	
	// Initially no nodes
	nodeIDs, err := loader.ListNodeFiles()
	if err != nil {
		t.Fatalf("Failed to list node files: %v", err)
	}
	
	if len(nodeIDs) != 0 {
		t.Errorf("Expected 0 nodes initially, got %d", len(nodeIDs))
	}
	
	// Create some nodes
	testNodeIDs := []string{
		"01234567-89ab-cdef-0123-456789abcdef",
		"11111111-2222-3333-4444-555555555555",
		"22222222-3333-4444-5555-666666666666",
	}
	
	for _, nodeID := range testNodeIDs {
		node := &types.Node{
			ID:        nodeID,
			Name:      "Test Node " + nodeID[:8],
			Children:  []string{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		err := loader.SaveNode(node)
		if err != nil {
			t.Fatalf("Failed to save node %s: %v", nodeID, err)
		}
	}
	
	// List nodes again
	nodeIDs, err = loader.ListNodeFiles()
	if err != nil {
		t.Fatalf("Failed to list node files: %v", err)
	}
	
	if len(nodeIDs) != len(testNodeIDs) {
		t.Errorf("Expected %d nodes, got %d", len(testNodeIDs), len(nodeIDs))
	}
	
	// Verify all expected node IDs are present
	nodeIDSet := make(map[string]bool)
	for _, id := range nodeIDs {
		nodeIDSet[id] = true
	}
	
	for _, expectedID := range testNodeIDs {
		if !nodeIDSet[expectedID] {
			t.Errorf("Expected node ID %s not found in list", expectedID)
		}
	}
}

func TestLoaderInvalidUUIDs(t *testing.T) {
	tmpDir := createTempDir(t)
	loader := NewLoader(tmpDir)
	
	invalidUUIDs := []string{
		"",
		"invalid-uuid",
		"01234567-89ab-cdef-0123-456789abcde", // too short
		"01234567-89ab-ghij-0123-456789abcdef", // invalid hex
	}
	
	for _, invalidID := range invalidUUIDs {
		t.Run("Invalid UUID: "+invalidID, func(t *testing.T) {
			// Test loading
			_, err := loader.LoadNode(invalidID)
			if err == nil {
				t.Error("Should fail to load node with invalid UUID")
			}
			
			envelope, ok := err.(errors.Envelope)
			if !ok || envelope.Code != errors.ErrInvalidUUID {
				t.Errorf("Expected INVALID_UUID error, got %v", err)
			}
			
			// Test saving
			node := &types.Node{
				ID:        invalidID,
				Name:      "Test",
				Children:  []string{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			
			err = loader.SaveNode(node)
			if err == nil {
				t.Error("Should fail to save node with invalid UUID")
			}
			
			envelope, ok = err.(errors.Envelope)
			if !ok || envelope.Code != errors.ErrInvalidUUID {
				t.Errorf("Expected INVALID_UUID error, got %v", err)
			}
			
			// Test deletion (should not error for invalid UUIDs, treated as non-existent)
			err = loader.DeleteNode(invalidID)
			if err != nil {
				t.Errorf("Should not error when deleting node with invalid UUID: %v", err)
			}
			
			// Test exists
			exists := loader.NodeExists(invalidID)
			if exists {
				t.Error("Invalid UUID should not exist")
			}
		})
	}
}

func TestLoaderCorruptedJSON(t *testing.T) {
	tmpDir := createTempDir(t)
	loader := NewLoader(tmpDir)
	
	// Create corrupted project.json
	projectPath := filepath.Join(tmpDir, "project.json")
	err := os.WriteFile(projectPath, []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupted project.json: %v", err)
	}
	
	// Try to load corrupted project
	_, err = loader.LoadProject()
	if err == nil {
		t.Error("Should fail to load corrupted project.json")
	}
	
	envelope, ok := err.(errors.Envelope)
	if !ok || envelope.Code != errors.ErrStorageFailure {
		t.Errorf("Expected STORAGE_FAILURE error, got %v", err)
	}
	
	// Create corrupted node file
	nodesDir := filepath.Join(tmpDir, "nodes")
	err = os.MkdirAll(nodesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nodes directory: %v", err)
	}
	
	nodeID := "01234567-89ab-cdef-0123-456789abcdef"
	nodePath := filepath.Join(nodesDir, nodeID+".json")
	err = os.WriteFile(nodePath, []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupted node file: %v", err)
	}
	
	// Try to load corrupted node
	_, err = loader.LoadNode(nodeID)
	if err == nil {
		t.Error("Should fail to load corrupted node file")
	}
	
	envelope, ok = err.(errors.Envelope)
	if !ok || envelope.Code != errors.ErrStorageFailure {
		t.Errorf("Expected STORAGE_FAILURE error, got %v", err)
	}
}

func TestLoaderDirectoryCreation(t *testing.T) {
	tmpDir := createTempDir(t)
	
	// Use a non-existent subdirectory
	projectDir := filepath.Join(tmpDir, "subdir", "project")
	loader := NewLoader(projectDir)
	
	node := &types.Node{
		ID:        "01234567-89ab-cdef-0123-456789abcdef",
		Name:      "Test Node",
		Children:  []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Should create nodes directory automatically
	err := loader.SaveNode(node)
	if err != nil {
		t.Fatalf("Failed to save node with automatic directory creation: %v", err)
	}
	
	// Verify directory was created
	nodesDir := filepath.Join(projectDir, "nodes")
	if _, err := os.Stat(nodesDir); os.IsNotExist(err) {
		t.Error("Nodes directory should have been created automatically")
	}
	
	// Verify node was saved
	if !loader.NodeExists(node.ID) {
		t.Error("Node should exist after saving")
	}
}

// Benchmark tests
func BenchmarkLoaderSaveNode(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "archon-bench-*")
	defer os.RemoveAll(tmpDir)
	
	loader := NewLoader(tmpDir)
	node := &types.Node{
		ID:        "01234567-89ab-cdef-0123-456789abcdef",
		Name:      "Benchmark Node",
		Children:  []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader.SaveNode(node)
	}
}

func BenchmarkLoaderLoadNode(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "archon-bench-*")
	defer os.RemoveAll(tmpDir)
	
	loader := NewLoader(tmpDir)
	node := &types.Node{
		ID:        "01234567-89ab-cdef-0123-456789abcdef",
		Name:      "Benchmark Node",
		Children:  []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	loader.SaveNode(node)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader.LoadNode(node.ID)
	}
}