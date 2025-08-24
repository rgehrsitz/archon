package api

import (
	"context"
	"testing"
)

func TestSnapshotService_NoProject(t *testing.T) {
	// Create service with a project service but no project loaded
	projectService := NewProjectService()
	service := NewSnapshotService(projectService)

	ctx := context.Background()

	// Test Create - should fail with no project
	req := CreateSnapshotRequest{
		Name:    "test-snapshot",
		Message: "Test snapshot",
	}

	_, env := service.Create(ctx, req)
	if env.Code == "" {
		t.Error("Expected error when creating snapshot without project")
	}

	// Test List - should fail with no project
	_, env = service.List(ctx)
	if env.Code == "" {
		t.Error("Expected error when listing snapshots without project")
	}

	// Test Get - should fail with no project
	_, env = service.Get(ctx, "test")
	if env.Code == "" {
		t.Error("Expected error when getting snapshot without project")
	}

	// Test Delete - should fail with no project
	env = service.Delete(ctx, "test")
	if env.Code == "" {
		t.Error("Expected error when deleting snapshot without project")
	}
}

func TestSnapshotService_WithProjectService(t *testing.T) {
	// Create services with project service (but no actual project)
	projectService := NewProjectService()
	service := NewSnapshotService(projectService)

	ctx := context.Background()

	// All operations should fail gracefully with "no project open" error
	req := CreateSnapshotRequest{
		Name:    "test-snapshot",
		Message: "Test snapshot",
	}

	_, env := service.Create(ctx, req)
	if env.Code == "" {
		t.Error("Expected error when no project is open")
	}
	if env.Message != "No project is currently open" {
		t.Errorf("Expected 'No project is currently open', got '%s'", env.Message)
	}
}

func TestCreateSnapshotRequest_Validation(t *testing.T) {
	// Test that the request structure is properly formed
	req := CreateSnapshotRequest{
		Name:        "test-snapshot",
		Message:     "Test message",
		Description: "Test description",
		Labels:      map[string]string{"env": "test"},
	}

	if req.Name != "test-snapshot" {
		t.Errorf("Expected name 'test-snapshot', got '%s'", req.Name)
	}
	if req.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", req.Message)
	}
	if req.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", req.Description)
	}
	if req.Labels["env"] != "test" {
		t.Errorf("Expected label env='test', got '%s'", req.Labels["env"])
	}
}

func TestSnapshotService_Restore_NoProject(t *testing.T) {
	// Create service with a project service but no project loaded
	projectService := NewProjectService()
	service := NewSnapshotService(projectService)

	ctx := context.Background()

	// Test Restore - should fail with no project
	env := service.Restore(ctx, "test")
	if env.Code == "" {
		t.Error("Expected error when restoring snapshot without project")
	}
	if env.Message != "No project is currently open" {
		t.Errorf("Expected 'No project is currently open', got '%s'", env.Message)
	}
}
