package store

import (
	"testing"
	"github.com/rgehrsitz/archon/internal/types"
)

// TestExampleProjectIntegration tests loading our example project
func TestExampleProjectIntegration(t *testing.T) {
	// Test loading the basic hierarchy example project
	examplePath := "../../examples/projects/basic-hierarchy"
	
	// Test project store
	projectStore := NewProjectStore(examplePath)
	
	if !projectStore.ProjectExists() {
		t.Skip("Example project not found - this is expected if running tests in isolation")
	}
	
	// Load project
	project, err := projectStore.OpenProject()
	if err != nil {
		t.Fatalf("Failed to load example project: %v", err)
	}
	
	// Verify project structure
	if project.RootID != "01234567-89ab-cdef-0123-456789abcdef" {
		t.Errorf("Expected root ID 01234567-89ab-cdef-0123-456789abcdef, got %s", project.RootID)
	}
	
	if project.SchemaVersion != types.CurrentSchemaVersion {
		t.Errorf("Expected schema version %d, got %d", types.CurrentSchemaVersion, project.SchemaVersion)
	}
	
	// Test node store
	nodeStore := NewNodeStore(examplePath)
	
	// Load root node
	rootNode, err := nodeStore.GetNode(project.RootID)
	if err != nil {
		t.Fatalf("Failed to load root node: %v", err)
	}
	
	if rootNode.Name != "Manufacturing Plant" {
		t.Errorf("Expected root node name 'Manufacturing Plant', got %s", rootNode.Name)
	}
	
	if len(rootNode.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(rootNode.Children))
	}
	
	// Test loading children
	for _, childID := range rootNode.Children {
		child, err := nodeStore.GetNode(childID)
		if err != nil {
			t.Fatalf("Failed to load child node %s: %v", childID, err)
		}
		
		// Verify child has valid structure
		if child.Name == "" {
			t.Errorf("Child node %s has empty name", childID)
		}
		
		if child.ID != childID {
			t.Errorf("Child node ID mismatch: expected %s, got %s", childID, child.ID)
		}
	}
	
	// Test specific nodes
	productionFloorID := "11111111-2222-3333-4444-555555555555"
	productionFloor, err := nodeStore.GetNode(productionFloorID)
	if err != nil {
		t.Fatalf("Failed to load production floor node: %v", err)
	}
	
	if productionFloor.Name != "Production Floor" {
		t.Errorf("Expected production floor name 'Production Floor', got %s", productionFloor.Name)
	}
	
	// Verify properties
	if capacity, exists := productionFloor.Properties["capacity"]; !exists {
		t.Error("Production floor should have capacity property")
	} else if capacity.Value != float64(1000) { // JSON unmarshals numbers as float64
		t.Errorf("Expected capacity 1000, got %v", capacity.Value)
	}
	
	// Test grandchildren (assembly lines)
	if len(productionFloor.Children) != 2 {
		t.Errorf("Expected production floor to have 2 children, got %d", len(productionFloor.Children))
	}
	
	// Test node path functionality
	assemblyLineID := productionFloor.Children[0]
	path, err := nodeStore.GetNodePath(assemblyLineID)
	if err != nil {
		t.Fatalf("Failed to get node path: %v", err)
	}
	
	expectedPathLength := 3 // Root -> Production Floor -> Assembly Line
	if len(path) != expectedPathLength {
		t.Errorf("Expected path length %d, got %d", expectedPathLength, len(path))
	}
	
	// Verify path order (root first)
	if path[0].ID != project.RootID {
		t.Errorf("First node in path should be root, got %s", path[0].ID)
	}
	
	if path[1].ID != productionFloorID {
		t.Errorf("Second node in path should be production floor, got %s", path[1].ID)
	}
	
	if path[2].ID != assemblyLineID {
		t.Errorf("Third node in path should be assembly line, got %s", path[2].ID)
	}
}

// TestProjectCreationAndManipulation tests creating a new project and manipulating it
func TestProjectCreationAndManipulation(t *testing.T) {
	tmpDir := createTempDir(t)
	
	// Create new project
	projectStore := NewProjectStore(tmpDir)
	
	settings := map[string]any{
		"name":        "Test Integration Project",
		"description": "A project created during integration testing",
	}
	
	project, err := projectStore.CreateProject(settings)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	
	// Verify project was created
	if !projectStore.ProjectExists() {
		t.Error("Project should exist after creation")
	}
	
	// Test node operations
	nodeStore := NewNodeStore(tmpDir)
	
	// Get root node
	rootNode, err := nodeStore.GetNode(project.RootID)
	if err != nil {
		t.Fatalf("Failed to get root node: %v", err)
	}
	
	if rootNode.Name != "Root" {
		t.Errorf("Expected root node name 'Root', got %s", rootNode.Name)
	}
	
	// Create child node
	createReq := &types.CreateNodeRequest{
		ParentID:    project.RootID,
		Name:        "Test Child",
		Description: "A test child node",
		Properties: map[string]types.Property{
			"test_prop": {TypeHint: "string", Value: "test_value"},
		},
	}
	
	child, err := nodeStore.CreateNode(createReq)
	if err != nil {
		t.Fatalf("Failed to create child node: %v", err)
	}
	
	// Verify child was created
	if child.Name != "Test Child" {
		t.Errorf("Expected child name 'Test Child', got %s", child.Name)
	}
	
	if child.Properties["test_prop"].Value != "test_value" {
		t.Errorf("Expected property value 'test_value', got %v", child.Properties["test_prop"].Value)
	}
	
	// Verify parent's children list was updated
	updatedRoot, err := nodeStore.GetNode(project.RootID)
	if err != nil {
		t.Fatalf("Failed to reload root node: %v", err)
	}
	
	if len(updatedRoot.Children) != 1 {
		t.Errorf("Expected root to have 1 child, got %d", len(updatedRoot.Children))
	}
	
	if updatedRoot.Children[0] != child.ID {
		t.Errorf("Expected root's child to be %s, got %s", child.ID, updatedRoot.Children[0])
	}
	
	// Test updating node
	updateReq := &types.UpdateNodeRequest{
		ID:          child.ID,
		Name:        stringPtr("Updated Child Name"),
		Description: stringPtr("Updated description"),
	}
	
	updatedChild, err := nodeStore.UpdateNode(updateReq)
	if err != nil {
		t.Fatalf("Failed to update child node: %v", err)
	}
	
	if updatedChild.Name != "Updated Child Name" {
		t.Errorf("Expected updated name 'Updated Child Name', got %s", updatedChild.Name)
	}
	
	if updatedChild.Description != "Updated description" {
		t.Errorf("Expected updated description 'Updated description', got %s", updatedChild.Description)
	}
	
	// Test creating grandchild and moving it
	grandchildReq := &types.CreateNodeRequest{
		ParentID: child.ID,
		Name:     "Grandchild",
	}
	
	grandchild, err := nodeStore.CreateNode(grandchildReq)
	if err != nil {
		t.Fatalf("Failed to create grandchild: %v", err)
	}
	
	// Move grandchild to root
	moveReq := &types.MoveNodeRequest{
		NodeID:      grandchild.ID,
		NewParentID: project.RootID,
		Position:    0, // Insert at beginning
	}
	
	err = nodeStore.MoveNode(moveReq)
	if err != nil {
		t.Fatalf("Failed to move grandchild: %v", err)
	}
	
	// Verify move
	finalRoot, err := nodeStore.GetNode(project.RootID)
	if err != nil {
		t.Fatalf("Failed to reload root after move: %v", err)
	}
	
	if len(finalRoot.Children) != 2 {
		t.Errorf("Expected root to have 2 children after move, got %d", len(finalRoot.Children))
	}
	
	if finalRoot.Children[0] != grandchild.ID {
		t.Errorf("Expected grandchild to be first child after move, got %s", finalRoot.Children[0])
	}
	
	// Verify old parent no longer has the grandchild
	finalChild, err := nodeStore.GetNode(child.ID)
	if err != nil {
		t.Fatalf("Failed to reload child after move: %v", err)
	}
	
	if len(finalChild.Children) != 0 {
		t.Errorf("Expected child to have 0 children after move, got %d", len(finalChild.Children))
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}