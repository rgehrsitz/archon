package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rgehrsitz/archon/model"
)

// ConfigVault errors
var (
	ErrInvalidPath      = errors.New("invalid project path")
	ErrProjectNotFound  = errors.New("project not found")
	ErrProjectExists    = errors.New("project already exists")
	ErrInvalidStructure = errors.New("invalid project structure")
)

// ProjectConfig represents the configuration stored in archon.json
type ProjectConfig struct {
	Version     string            `json:"version"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ConfigVault layout defines the standard file structure
const (
	ComponentsFile = "components.json"
	ConfigFile     = "archon.json"
	AttachmentsDir = "attachments"
)

// ConfigVault handles storage and retrieval of configuration data
// This is the primary storage layer for Archon projects
type ConfigVault struct {
	mu sync.RWMutex

	// Current state
	rootPath    string
	tree        *model.ComponentTree
	config      ProjectConfig
	initialized bool
	changeCount int // for auto-snapshot trigger (MVP)
}

// NewConfigVault creates a new ConfigVault instance for the given path
func NewConfigVault(path string) (*ConfigVault, error) {
	if path == "" {
		// Allow empty path for in-memory operations
		tree, _ := model.NewComponentTree([]*model.Component{})
		return &ConfigVault{
			tree: tree,
			config: ProjectConfig{
				Version:  "1.0",
				Metadata: make(map[string]string),
			},
		}, nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	tree, _ := model.NewComponentTree([]*model.Component{})
	return &ConfigVault{
		rootPath: absPath,
		tree:     tree,
		config: ProjectConfig{
			Version:  "1.0",
			Metadata: make(map[string]string),
		},
	}, nil
}

// Initialize creates the project structure at the specified path
func (v *ConfigVault) Initialize(name string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.initialized {
		return nil
	}

	if v.rootPath == "" {
		return ErrInvalidPath
	}

	// Check if project already exists
	if _, err := os.Stat(v.rootPath); err == nil {
		// Path exists, check if it's already an Archon project
		if _, err := os.Stat(filepath.Join(v.rootPath, ConfigFile)); err == nil {
			return ErrProjectExists
		}
	}

	// Create project directory if it doesn't exist
	if err := os.MkdirAll(v.rootPath, 0o755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create attachments directory
	if err := os.MkdirAll(filepath.Join(v.rootPath, AttachmentsDir), 0o755); err != nil {
		return fmt.Errorf("failed to create attachments directory: %w", err)
	}

	// Set project config
	v.config.Name = name
	v.config.Version = "1.0"

	// Create empty components.json
	if err := os.WriteFile(filepath.Join(v.rootPath, ComponentsFile), []byte("[]"), 0o644); err != nil {
		return fmt.Errorf("failed to create components file: %w", err)
	}

	// Write project config
	if err := v.SaveConfig(); err != nil {
		return err
	}

	v.initialized = true
	return nil
}

// Load loads a project from the configured path
func (v *ConfigVault) Load(path string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// If path is provided, update our path
	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		v.rootPath = absPath
	}

	if v.rootPath == "" {
		return ErrInvalidPath
	}

	// Check if project exists
	configPath := filepath.Join(v.rootPath, ConfigFile)
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("failed to access project: %w", err)
	}

	// Verify project structure
	if err := v.verifyStructure(); err != nil {
		return err
	}

	// Load project config
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read project config: %w", err)
	}

	if err := json.Unmarshal(configData, &v.config); err != nil {
		return fmt.Errorf("failed to parse project config: %w", err)
	}

	// Load components
	components, err := v.LoadComponents()
	if err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}

	// Create a new component tree from the loaded components
	tree, err := model.NewComponentTree(components)
	if err != nil {
		return fmt.Errorf("failed to create component tree: %w", err)
	}

	v.tree = tree
	v.initialized = true
	return nil
}

// LoadComponents loads all components from the project's components.json file
func (v *ConfigVault) LoadComponents() ([]*model.Component, error) {
	if v.rootPath == "" {
		return nil, ErrInvalidPath
	}

	data, err := os.ReadFile(v.GetComponentsPath())
	if err != nil {
		return nil, fmt.Errorf("failed to read components file: %w", err)
	}

	var components []*model.Component
	if err := json.Unmarshal(data, &components); err != nil {
		return nil, fmt.Errorf("failed to parse components: %w", err)
	}

	// Validate all components
	for _, c := range components {
		if err := c.Validate(); err != nil {
			return nil, fmt.Errorf("invalid component %s: %w", c.ID, err)
		}
	}

	return components, nil
}

// Save saves the current state to disk
func (v *ConfigVault) Save() error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.rootPath == "" || v.rootPath == "<in-memory>" {
		return fmt.Errorf("no valid project path for saving")
	}

	// Convert tree to components array
	var components []*model.Component
	for _, component := range v.tree.Components {
		components = append(components, component)
	}

	return v.SaveComponents(components)
}

// SaveComponents saves the given components to the project's components.json file
func (v *ConfigVault) SaveComponents(components []*model.Component) error {
	if v.rootPath == "" || v.rootPath == "<in-memory>" {
		return fmt.Errorf("no valid project path for saving")
	}

	// Validate all components before saving
	for _, c := range components {
		if err := c.Validate(); err != nil {
			return fmt.Errorf("invalid component %s: %w", c.ID, err)
		}
	}

	data, err := json.MarshalIndent(components, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize components: %w", err)
	}

	if err := os.WriteFile(v.GetComponentsPath(), data, 0o644); err != nil {
		return fmt.Errorf("failed to write components file: %w", err)
	}

	// Increment change count for auto-snapshot
	v.changeCount++
	v.autoSnapshot("")

	return nil
}

// SaveConfig writes the project configuration to archon.json
func (v *ConfigVault) SaveConfig() error {
	if v.rootPath == "" || v.rootPath == "<in-memory>" {
		return fmt.Errorf("no valid project path for saving config")
	}

	configData, err := json.MarshalIndent(v.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize project config: %w", err)
	}

	if err := os.WriteFile(filepath.Join(v.rootPath, ConfigFile), configData, 0o644); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	// Auto-snapshot on config change
	v.autoSnapshot("Config change")

	return nil
}

// GetComponentTree returns the current component tree
func (v *ConfigVault) GetComponentTree() (*model.ComponentTree, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.tree == nil {
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

	// Check if component exists
	existing, exists := v.tree.Components[component.ID]
	if !exists {
		return model.ErrComponentNotFound
	}

	// Validate component
	if err := component.Validate(); err != nil {
		return fmt.Errorf("invalid component: %w", err)
	}

	// Update the component in place
	v.tree.Components[component.ID] = component

	// If we have a valid path, save to disk
	if v.rootPath != "" && v.rootPath != "<in-memory>" {
		// Convert tree to components array and save
		var components []*model.Component
		for _, c := range v.tree.Components {
			components = append(components, c)
		}
		if err := v.SaveComponents(components); err != nil {
			// Rollback the change
			v.tree.Components[component.ID] = existing
			return fmt.Errorf("failed to persist component update: %w", err)
		}
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

	// If we have a valid path, save to disk
	if v.rootPath != "" && v.rootPath != "<in-memory>" {
		// Convert tree to components array and save
		var components []*model.Component
		for _, c := range v.tree.Components {
			components = append(components, c)
		}
		if err := v.SaveComponents(components); err != nil {
			return fmt.Errorf("failed to persist component addition: %w", err)
		}
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
		return nil, model.ErrComponentNotFound
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

	// If we have a valid path, save to disk
	if v.rootPath != "" && v.rootPath != "<in-memory>" {
		// Convert tree to components array and save
		var components []*model.Component
		for _, c := range v.tree.Components {
			components = append(components, c)
		}
		if err := v.SaveComponents(components); err != nil {
			return fmt.Errorf("failed to persist component deletion: %w", err)
		}
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
	v.initialized = true
	return nil
}

// GetComponentsPath returns the absolute path to the components.json file
func (v *ConfigVault) GetComponentsPath() string {
	if v.rootPath == "" || v.rootPath == "<in-memory>" {
		return ""
	}
	return filepath.Join(v.rootPath, ComponentsFile)
}

// GetAttachmentsPath returns the absolute path to the attachments directory
func (v *ConfigVault) GetAttachmentsPath() string {
	if v.rootPath == "" || v.rootPath == "<in-memory>" {
		return ""
	}
	return filepath.Join(v.rootPath, AttachmentsDir)
}

// GetConfig returns the current project configuration
func (v *ConfigVault) GetConfig() ProjectConfig {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.config
}

// UpdateLastSnapshot updates the last snapshot reference in the project config
func (v *ConfigVault) UpdateLastSnapshot(snapshotID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.config.Metadata == nil {
		v.config.Metadata = make(map[string]string)
	}
	v.config.Metadata["lastSnapshot"] = snapshotID
	return v.SaveConfig()
}

// GetLastSnapshot returns the ID of the last snapshot
func (v *ConfigVault) GetLastSnapshot() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.config.Metadata["lastSnapshot"]
}

// HasUnsavedChanges returns true if there are unsaved changes in the project
func (v *ConfigVault) HasUnsavedChanges() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.changeCount > 0
}

// ResetChangeCount resets the change counter
func (v *ConfigVault) ResetChangeCount() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.changeCount = 0
}

// verifyStructure checks if the project has the required file structure
func (v *ConfigVault) verifyStructure() error {
	if v.rootPath == "" || v.rootPath == "<in-memory>" {
		return nil // In-memory projects don't need file structure
	}

	// Check for components.json
	if _, err := os.Stat(filepath.Join(v.rootPath, ComponentsFile)); err != nil {
		if os.IsNotExist(err) {
			return ErrInvalidStructure
		}
		return fmt.Errorf("failed to access components file: %w", err)
	}

	// Check for attachments directory
	attachmentsPath := filepath.Join(v.rootPath, AttachmentsDir)
	info, err := os.Stat(attachmentsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create attachments directory if it doesn't exist
			if err := os.MkdirAll(attachmentsPath, 0o755); err != nil {
				return fmt.Errorf("failed to create attachments directory: %w", err)
			}
		} else {
			return fmt.Errorf("failed to access attachments directory: %w", err)
		}
	} else if !info.IsDir() {
		return ErrInvalidStructure
	}

	return nil
}

// autoSnapshot triggers an auto-snapshot if changeCount >= 5 or always if reason is given.
func (v *ConfigVault) autoSnapshot(reason string) {
	if reason == "" && v.changeCount < 5 {
		return
	}
	tag := "auto-" + time.Now().UTC().Format("20060102-150405")
	msg := reason
	// TODO: Implement actual snapshot creation once snapshot manager is updated
	// For now, just log what would happen
	if reason != "" || v.changeCount >= 5 {
		fmt.Printf("[Archon] Auto-snapshot triggered: %s (%s)\n", tag, msg)
		v.changeCount = 0
	}
}
