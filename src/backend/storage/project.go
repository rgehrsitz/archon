// Package storage provides functionality for managing Archon project storage.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Project errors
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

// Project represents an Archon project with its storage structure
type Project struct {
	Path        string        `json:"-"`
	Config      ProjectConfig `json:"config"`
	initialized bool

	changeCount int // for auto-snapshot trigger (MVP)
}

// ProjectLayout defines the standard file structure for an Archon project
const (
	ComponentsFile = "components.json"
	ConfigFile     = "archon.json"
	AttachmentsDir = "attachments"
)

// New creates a new Project instance for the given path
func New(path string) (*Project, error) {
	if path == "" {
		return nil, ErrInvalidPath
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return &Project{
		Path: absPath,
		Config: ProjectConfig{
			Version:  "1.0",
			Metadata: make(map[string]string),
		},
	}, nil
}

// Initialize creates the project structure at the specified path
func (p *Project) Initialize(name string) error {
	if p.initialized {
		return nil
	}

	// Check if project already exists
	if _, err := os.Stat(p.Path); err == nil {
		// Path exists, check if it's already an Archon project
		if _, err := os.Stat(filepath.Join(p.Path, ConfigFile)); err == nil {
			return ErrProjectExists
		}
	}

	// Create project directory if it doesn't exist
	if err := os.MkdirAll(p.Path, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create attachments directory
	if err := os.MkdirAll(filepath.Join(p.Path, AttachmentsDir), 0755); err != nil {
		return fmt.Errorf("failed to create attachments directory: %w", err)
	}

	// Set project config
	p.Config.Name = name
	p.Config.Version = "1.0"

	// Create empty components.json
	if err := os.WriteFile(filepath.Join(p.Path, ComponentsFile), []byte("[]"), 0644); err != nil {
		return fmt.Errorf("failed to create components file: %w", err)
	}

	// Write project config
	if err := p.SaveConfig(); err != nil {
		return err
	}

	p.initialized = true
	return nil
}

// Open loads an existing project from the specified path
func Open(path string) (*Project, error) {
	project, err := New(path)
	if err != nil {
		return nil, err
	}

	// Check if project exists
	configPath := filepath.Join(project.Path, ConfigFile)
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to access project: %w", err)
	}

	// Verify project structure
	if err := project.verifyStructure(); err != nil {
		return nil, err
	}

	// Load project config
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	if err := json.Unmarshal(configData, &project.Config); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	project.initialized = true
	return project, nil
}

// SaveConfig writes the project configuration to archon.json
func (p *Project) SaveConfig() error {
	configData, err := json.MarshalIndent(p.Config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize project config: %w", err)
	}

	if err := os.WriteFile(filepath.Join(p.Path, ConfigFile), configData, 0644); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	// Auto-snapshot on config change
	p.autoSnapshot("Config change")

	return nil
}

// ImportConfigWithAutoSnapshot is a stub for MVP: triggers auto-snapshot before import.
func (p *Project) ImportConfigWithAutoSnapshot(importPath string) error {
	p.autoSnapshot("Pre-import")
	// TODO: implement import logic
	return nil
}

// autoSnapshot triggers an auto-snapshot if changeCount >= 5 or always if reason is given.
func (p *Project) autoSnapshot(reason string) {
	if reason == "" && p.changeCount < 5 {
		return
	}
	tag := "auto-" + time.Now().UTC().Format("20060102-150405")
	msg := reason
	_, err := p.CreateSnapshot(tag, msg, "auto")
	if err == nil {
		fmt.Printf("[Archon] Auto-snapshot created: %s (%s)\n", tag, msg)
		p.changeCount = 0
	} else {
		fmt.Printf("[Archon] Auto-snapshot failed: %v\n", err)
	}
}


// verifyStructure checks if the project has the required file structure
func (p *Project) verifyStructure() error {
	// Check for components.json
	if _, err := os.Stat(filepath.Join(p.Path, ComponentsFile)); err != nil {
		if os.IsNotExist(err) {
			return ErrInvalidStructure
		}
		return fmt.Errorf("failed to access components file: %w", err)
	}

	// Check for attachments directory
	attachmentsPath := filepath.Join(p.Path, AttachmentsDir)
	info, err := os.Stat(attachmentsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create attachments directory if it doesn't exist
			if err := os.MkdirAll(attachmentsPath, 0755); err != nil {
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

// GetComponentsPath returns the absolute path to the components.json file
func (p *Project) GetComponentsPath() string {
	return filepath.Join(p.Path, ComponentsFile)
}

// GetAttachmentsPath returns the absolute path to the attachments directory
func (p *Project) GetAttachmentsPath() string {
	return filepath.Join(p.Path, AttachmentsDir)
}
