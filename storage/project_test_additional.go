package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProject_verifyStructure(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a project instance
	project, _ := New(tempDir)

	// Test case 1: Missing components.json
	err = project.verifyStructure()
	if err == nil || err != ErrInvalidStructure {
		t.Errorf("Expected ErrInvalidStructure, got %v", err)
	}

	// Create components.json
	componentsPath := filepath.Join(tempDir, ComponentsFile)
	if err := os.WriteFile(componentsPath, []byte("[]"), 0644); err != nil {
		t.Fatalf("Failed to create components.json: %v", err)
	}

	// Test case 2: Missing attachments directory
	err = project.verifyStructure()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify attachments directory was created
	attachmentsPath := filepath.Join(tempDir, AttachmentsDir)
	if _, err := os.Stat(attachmentsPath); os.IsNotExist(err) {
		t.Error("Attachments directory was not created")
	}

	// Test case 3: Attachments is a file, not a directory
	os.RemoveAll(attachmentsPath)
	if err := os.WriteFile(attachmentsPath, []byte("not a directory"), 0644); err != nil {
		t.Fatalf("Failed to create attachments file: %v", err)
	}

	err = project.verifyStructure()
	if err == nil || err != ErrInvalidStructure {
		t.Errorf("Expected ErrInvalidStructure, got %v", err)
	}

	// Test case 4: Cannot access components file
	if err := os.Chmod(componentsPath, 0000); err == nil {
		// Only run this test if we can change permissions
		err = project.verifyStructure()
		if err == nil {
			t.Error("Expected error for inaccessible components file, got nil")
		}
		// Restore permissions for cleanup
		os.Chmod(componentsPath, 0644)
	}
}

// TestProject_GetPaths moved to the main test file

func TestProject_Initialize_EdgeCases(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case: Initialize with already initialized project
	project, _ := New(tempDir)
	if err := project.Initialize("Test Project"); err != nil {
		t.Errorf("First Initialize() error = %v", err)
	}

	// Initialize again - should not return an error
	if err := project.Initialize("Test Project"); err != nil {
		t.Errorf("Second Initialize() error = %v, expected nil", err)
	}

	// Test case: Cannot create attachments directory
	// Create a file where the attachments directory should be
	attachmentsPath := filepath.Join(tempDir, "another_project", AttachmentsDir)
	if err := os.MkdirAll(filepath.Dir(attachmentsPath), 0755); err != nil {
		t.Fatalf("Failed to create parent directory: %v", err)
	}
	if err := os.WriteFile(attachmentsPath, []byte("not a directory"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	anotherProject, _ := New(filepath.Join(tempDir, "another_project"))
	err = anotherProject.Initialize("Another Project")
	if err == nil {
		t.Error("Expected error when attachments path is a file, got nil")
	}
}

func TestOpen_EdgeCases(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "archon_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case: Config file exists but is invalid JSON
	projectPath := filepath.Join(tempDir, "invalid_config")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create components.json
	componentsPath := filepath.Join(projectPath, ComponentsFile)
	if err := os.WriteFile(componentsPath, []byte("[]"), 0644); err != nil {
		t.Fatalf("Failed to create components.json: %v", err)
	}

	// Create invalid archon.json
	configPath := filepath.Join(projectPath, ConfigFile)
	if err := os.WriteFile(configPath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create archon.json: %v", err)
	}

	// Try to open the project
	_, err = Open(projectPath)
	if err == nil {
		t.Error("Expected error when opening project with invalid config, got nil")
	}

	// Test case: Cannot access config file
	if err := os.Chmod(configPath, 0000); err == nil {
		// Only run this test if we can change permissions
		_, err = Open(projectPath)
		if err == nil {
			t.Error("Expected error for inaccessible config file, got nil")
		}
		// Restore permissions for cleanup
		os.Chmod(configPath, 0644)
	}
}
