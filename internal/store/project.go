package store

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/id"
	"github.com/rgehrsitz/archon/internal/index"
	"github.com/rgehrsitz/archon/internal/types"
)

// ProjectStore handles project-level operations: create/open project structure
type ProjectStore struct {
	basePath     string
	loader       *Loader
	IndexManager *index.Manager
}

// NewProjectStore creates a new project store
func NewProjectStore(basePath string) (*ProjectStore, error) {
	indexManager, err := index.NewManager(basePath)
	if err != nil {
		return nil, err
	}
	
	return &ProjectStore{
		basePath:     basePath,
		loader:       NewLoader(basePath),
		IndexManager: indexManager,
	}, nil
}

// Close cleanly shuts down the project store
func (ps *ProjectStore) Close() error {
	if ps.IndexManager != nil {
		return ps.IndexManager.Close()
	}
	return nil
}

// CreateProject creates a new Archon project at the specified path
func (ps *ProjectStore) CreateProject(settings map[string]any) (*types.Project, error) {
	// Check if project already exists
	if ps.loader.ProjectExists() {
		return nil, errors.New(errors.ErrProjectExists, "Project already exists at this location")
	}
	
	// Create base directory
	if err := os.MkdirAll(ps.basePath, 0755); err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create project directory", err)
	}
	
	// Create .archon directory structure
	archonDir := filepath.Join(ps.basePath, ".archon")
	indexDir := filepath.Join(archonDir, "index")
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create .archon directory", err)
	}
	
	// Create nodes directory
	nodesDir := filepath.Join(ps.basePath, "nodes")
	if err := os.MkdirAll(nodesDir, 0755); err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create nodes directory", err)
	}
	
	// Create attachments directory
	attachmentsDir := filepath.Join(ps.basePath, "attachments")
	if err := os.MkdirAll(attachmentsDir, 0755); err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to create attachments directory", err)
	}
	
	// Generate root node ID
	rootID := id.NewV7()
	now := time.Now()
	
	// Create project structure
	project := &types.Project{
		RootID:        rootID,
		SchemaVersion: types.CurrentSchemaVersion,
		Settings:      settings,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	
	// Validate project
	if validationErrors := ValidateProject(project); len(validationErrors) > 0 {
		return nil, errors.FromValidationErrors(validationErrors)
	}
	
	// Create root node
	rootNode := &types.Node{
		ID:          rootID,
		Name:        "Root",
		Description: "Project root node",
		Properties:  make(map[string]types.Property),
		Children:    []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	// Save project file
	if err := ps.loader.SaveProject(project); err != nil {
		return nil, err
	}
	
	// Save root node
	if err := ps.loader.SaveNode(rootNode); err != nil {
		return nil, err
	}
	
	// TODO: Initialize Git repository and LFS (future implementation)
	
	return project, nil
}

// OpenProject opens an existing Archon project
func (ps *ProjectStore) OpenProject() (*types.Project, error) {
	// Check if project exists
	if !ps.loader.ProjectExists() {
		return nil, errors.New(errors.ErrProjectNotFound, "Project not found at this location")
	}
	
	// Load project
	project, err := ps.loader.LoadProject()
	if err != nil {
		return nil, err
	}
	
	// Validate project
	if validationErrors := ValidateProject(project); len(validationErrors) > 0 {
		return nil, errors.FromValidationErrors(validationErrors)
	}
	
	// TODO: Check schema version and migrate if needed (future implementation)
	// TODO: Validate Git repository state (future implementation)
	
	return project, nil
}

// GetProjectInfo returns basic project information without full load
func (ps *ProjectStore) GetProjectInfo() (map[string]any, error) {
	project, err := ps.loader.LoadProject()
	if err != nil {
		return nil, err
	}
	
	return map[string]any{
		"rootId":        project.RootID,
		"schemaVersion": project.SchemaVersion,
		"createdAt":     project.CreatedAt,
		"updatedAt":     project.UpdatedAt,
		"settings":      project.Settings,
	}, nil
}

// UpdateProjectSettings updates project settings
func (ps *ProjectStore) UpdateProjectSettings(settings map[string]any) error {
	project, err := ps.loader.LoadProject()
	if err != nil {
		return err
	}
	
	// Merge settings
	if project.Settings == nil {
		project.Settings = make(map[string]any)
	}
	for key, value := range settings {
		project.Settings[key] = value
	}
	
	return ps.loader.SaveProject(project)
}

// ProjectExists checks if a valid Archon project exists at the path
func (ps *ProjectStore) ProjectExists() bool {
	return ps.loader.ProjectExists()
}

// InitializeDirectories ensures all required project directories exist
func (ps *ProjectStore) InitializeDirectories() error {
	directories := []string{
		ps.basePath,
		filepath.Join(ps.basePath, ".archon"),
		filepath.Join(ps.basePath, ".archon", "index"),
		filepath.Join(ps.basePath, "nodes"),
		filepath.Join(ps.basePath, "attachments"),
	}
	
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.WrapError(errors.ErrStorageFailure, "Failed to create directory: "+dir, err)
		}
	}
	
	return nil
}
