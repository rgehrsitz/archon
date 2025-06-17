package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rgehrsitz/archon/model"
)

// ConfigVault handles storage and retrieval of configuration data
type ConfigVault struct {
	mu sync.RWMutex

	// Current state
	rootPath string
	tree     *model.ComponentTree
}

// NewConfigVault creates a new ConfigVault instance
func NewConfigVault() *ConfigVault {
	tree, _ := model.NewComponentTree([]*model.Component{})
	return &ConfigVault{
		tree: tree,
	}
}

// Load loads a project from the given path
func (v *ConfigVault) Load(path string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Verify path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", path)
	}

	// Read project file
	projectFile := filepath.Join(path, "project.json")
	data, err := os.ReadFile(projectFile)
	if err != nil {
		return fmt.Errorf("failed to read project file: %w", err)
	}

	// Parse project data
	var components []*model.Component
	if err := json.Unmarshal(data, &components); err != nil {
		return fmt.Errorf("failed to parse project file: %w", err)
	}

	// Create a new component tree from the parsed components
	tree, err := model.NewComponentTree(components)
	if err != nil {
		return fmt.Errorf("failed to create component tree: %w", err)
	}

	v.rootPath = path
	v.tree = tree
	return nil
}

// Save saves the current state to disk
func (v *ConfigVault) Save() error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.rootPath == "" {
		return fmt.Errorf("no project loaded")
	}

	// Convert tree to components array
	var components []*model.Component
	for _, component := range v.tree.Components {
		components = append(components, component)
	}

	// Marshal components to JSON
	data, err := json.MarshalIndent(components, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project data: %w", err)
	}

	// Write to file
	projectFile := filepath.Join(v.rootPath, "project.json")
	if err := os.WriteFile(projectFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write project file: %w", err)
	}

	return nil
}

// GetComponentTree returns the current component tree
func (v *ConfigVault) GetComponentTree() (*model.ComponentTree, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.rootPath == "" {
		return nil, fmt.Errorf("no project loaded")
	}

	return v.tree, nil
}

// UpdateComponent updates a component in the tree
func (v *ConfigVault) UpdateComponent(component *model.Component) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.tree == nil {
		return fmt.Errorf("no project loaded")
	}

	// Validate component
	if err := component.Validate(); err != nil {
		return fmt.Errorf("invalid component: %w", err)
	}

	// Update component in tree
	if err := v.tree.AddComponent(component); err != nil {
		return fmt.Errorf("failed to update component: %w", err)
	}

	return nil
}

// AddComponent adds a new component to the tree
func (v *ConfigVault) AddComponent(component *model.Component) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.tree == nil {
		return fmt.Errorf("no project loaded")
	}

	// Validate component
	if err := component.Validate(); err != nil {
		return fmt.Errorf("invalid component: %w", err)
	}

	// Add component to tree
	if err := v.tree.AddComponent(component); err != nil {
		return fmt.Errorf("failed to add component: %w", err)
	}

	return nil
}

// GetComponent retrieves a component by ID
func (v *ConfigVault) GetComponent(id string) (*model.Component, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.tree == nil {
		return nil, fmt.Errorf("no project loaded")
	}

	component := v.tree.Components[id]
	if component == nil {
		return nil, fmt.Errorf("component not found: %s", id)
	}

	return component, nil
}

// DeleteComponent removes a component from the tree
func (v *ConfigVault) DeleteComponent(id string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.tree == nil {
		return fmt.Errorf("no project loaded")
	}

	// Delete component from tree
	if err := v.tree.RemoveComponent(id); err != nil {
		return fmt.Errorf("failed to delete component: %w", err)
	}

	return nil
}

// InitializeInMemory initializes the vault with components without requiring a file path
func (v *ConfigVault) InitializeInMemory(components []*model.Component) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Create a new component tree from the components
	tree, err := model.NewComponentTree(components)
	if err != nil {
		return fmt.Errorf("failed to create component tree: %w", err)
	}

	v.tree = tree
	v.rootPath = "<in-memory>" // Special marker for in-memory projects
	return nil
}
