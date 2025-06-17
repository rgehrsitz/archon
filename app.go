package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rgehrsitz/archon/model"
	"github.com/rgehrsitz/archon/plugin"
	"github.com/rgehrsitz/archon/snapshot"
	"github.com/rgehrsitz/archon/storage"
)

// App struct holds the application state and dependencies
type App struct {
	ctx context.Context
	mu  sync.RWMutex

	// Core services
	configVault *storage.ConfigVault
	snapshotMgr *snapshot.Manager
	pluginMgr   *plugin.PluginManager

	// State
	currentProject string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	// Initialize with empty path for in-memory operations initially
	a.configVault, _ = storage.NewConfigVault("")
	a.snapshotMgr = snapshot.NewManager(a.configVault)
	a.pluginMgr = plugin.NewPluginManager()

	// Initialize with sample data for demonstration
	if err := a.InitializeSampleProject(); err != nil {
		// Log error but don't fail startup
		fmt.Printf("Warning: Failed to initialize sample project: %v\n", err)
	}
}

// LoadProject loads a project from the given path
func (a *App) LoadProject(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.configVault.Load(path); err != nil {
		return err
	}
	a.currentProject = path
	return nil
}

// GetComponentTree returns the current component tree
func (a *App) GetComponentTree() (*model.ComponentTree, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.configVault.GetComponentTree()
}

// CreateSnapshot creates a new snapshot of the current state
func (a *App) CreateSnapshot(message string) (*snapshot.Snapshot, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.snapshotMgr.Create(message)
}

// GetSnapshots returns all snapshots for the current project
func (a *App) GetSnapshots() ([]*snapshot.Snapshot, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.snapshotMgr.List()
}

// LoadPlugin loads a WASM plugin
func (a *App) LoadPlugin(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.pluginMgr.LoadPlugin(path)
}

// ExecutePlugin runs a loaded plugin with the given parameters
func (a *App) ExecutePlugin(pluginID string, params map[string]interface{}) (interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.pluginMgr.Execute(pluginID, params)
}

// CreateComponent creates a new component in the current project
func (a *App) CreateComponent(component *model.Component) (*model.Component, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.configVault == nil {
		return nil, errors.New("no project loaded")
	}

	// Validate the component
	if err := component.Validate(); err != nil {
		return nil, fmt.Errorf("invalid component: %w", err)
	}

	// Add to vault
	if err := a.configVault.AddComponent(component); err != nil {
		return nil, fmt.Errorf("failed to add component: %w", err)
	}

	return component, nil
}

// CreateComponentSimple creates a new component with basic parameters
func (a *App) CreateComponentSimple(name, componentType, parentID string) (*model.Component, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.configVault == nil {
		return nil, errors.New("no project loaded")
	}

	// Generate a new UUID for the component
	id := model.GenerateID()

	// Create the component
	component := model.NewComponent(id, name, componentType)
	if parentID != "" {
		component.ParentID = parentID
	}

	// Validate the component
	if err := component.Validate(); err != nil {
		return nil, fmt.Errorf("invalid component: %w", err)
	}

	// Add to vault
	if err := a.configVault.AddComponent(component); err != nil {
		return nil, fmt.Errorf("failed to add component: %w", err)
	}

	return component, nil
}

// UpdateComponent updates an existing component
func (a *App) UpdateComponent(id string, updates *model.Component) (*model.Component, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.configVault == nil {
		return nil, errors.New("no project loaded")
	}

	component, err := a.configVault.GetComponent(id)
	if err != nil {
		return nil, fmt.Errorf("component not found: %w", err)
	}

	// Apply updates
	if updates.Name != "" {
		component.Name = updates.Name
	}
	if updates.Type != "" {
		component.Type = updates.Type
	}
	if updates.Description != "" {
		component.Description = updates.Description
	}
	if updates.ParentID != "" {
		component.ParentID = updates.ParentID
	}
	if updates.Properties != nil {
		component.Properties = updates.Properties
	}
	if updates.Metadata != nil {
		component.Metadata = updates.Metadata
	}

	// Validate and save
	if err := component.Validate(); err != nil {
		return nil, fmt.Errorf("invalid component after update: %w", err)
	}

	if err := a.configVault.UpdateComponent(component); err != nil {
		return nil, fmt.Errorf("failed to update component: %w", err)
	}

	return component, nil
}

// DeleteComponent removes a component from the current project
func (a *App) DeleteComponent(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.configVault == nil {
		return errors.New("no project loaded")
	}

	return a.configVault.DeleteComponent(id)
}

// GetComponent returns a specific component by ID
func (a *App) GetComponent(id string) (*model.Component, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.configVault == nil {
		return nil, errors.New("no project loaded")
	}

	return a.configVault.GetComponent(id)
}

// CreateProject creates a new project at the specified path
func (a *App) CreateProject(path, name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Create project directory if it doesn't exist
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}
	// Initialize the ConfigVault with the new path
	vault, err := storage.NewConfigVault(path)
	if err != nil {
		return fmt.Errorf("failed to create config vault: %w", err)
	}
	a.configVault = vault

	// Create a root component
	rootComponent := model.NewComponent("root", "Root", "system")
	rootComponent.Description = "Root component for " + name

	// Create project file with initial root component
	components := []*model.Component{rootComponent}
	data, err := json.MarshalIndent(components, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal initial project data: %w", err)
	}

	projectFile := filepath.Join(path, "project.json")
	if err := os.WriteFile(projectFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write project file: %w", err)
	}

	// Load the newly created project
	if err := a.configVault.Load(path); err != nil {
		return fmt.Errorf("failed to load newly created project: %w", err)
	}

	a.currentProject = path
	return nil
}

// InitializeSampleProject creates a sample project for demonstration
func (a *App) InitializeSampleProject() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Create sample components
	root := model.NewComponent("root", "Sample Lab Equipment", "system")
	root.Description = "Sample lab equipment hierarchy"

	microscope := model.NewComponent("microscope-001", "Zeiss Microscope", "microscope")
	microscope.ParentID = "root"
	microscope.Properties = map[string]interface{}{
		"model":         "Zeiss Axio Observer",
		"magnification": "1000x",
		"status":        "operational",
	}

	camera := model.NewComponent("camera-001", "Digital Camera", "camera")
	camera.ParentID = "microscope-001"
	camera.Properties = map[string]interface{}{
		"resolution": "4K",
		"sensor":     "CMOS",
	}
	// Create ConfigVault and initialize with sample components
	vault, err := storage.NewConfigVault("")
	if err != nil {
		return fmt.Errorf("failed to create config vault: %w", err)
	}
	a.configVault = vault
	components := []*model.Component{root, microscope, camera}

	if err := a.configVault.InitializeInMemory(components); err != nil {
		return fmt.Errorf("failed to initialize sample project: %w", err)
	}

	return nil
}
