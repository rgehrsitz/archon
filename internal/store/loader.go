package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/id"
	"github.com/rgehrsitz/archon/internal/types"
)

// Loader handles reading and writing project.json and nodes/<id>.json files
type Loader struct {
	basePath string
}

// NewLoader creates a new loader for the given project path
func NewLoader(basePath string) *Loader {
	return &Loader{basePath: basePath}
}

// LoadProject reads and parses project.json
func (l *Loader) LoadProject() (*types.Project, error) {
	projectPath := filepath.Join(l.basePath, "project.json")
	
	data, err := os.ReadFile(projectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrProjectNotFound, "Project file not found")
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to read project file", err)
	}
	
	var project types.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to parse project file", err)
	}
	
	return &project, nil
}

// SaveProject writes project.json
func (l *Loader) SaveProject(project *types.Project) error {
	project.UpdatedAt = time.Now()
	
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return errors.WrapError(errors.ErrStorageFailure, "Failed to serialize project", err)
	}
	
	projectPath := filepath.Join(l.basePath, "project.json")
	if err := os.WriteFile(projectPath, data, 0644); err != nil {
		return errors.WrapError(errors.ErrStorageFailure, "Failed to write project file", err)
	}
	
	return nil
}

// LoadNode reads and parses nodes/<id>.json
func (l *Loader) LoadNode(nodeID string) (*types.Node, error) {
	if !id.IsValid(nodeID) {
		return nil, errors.New(errors.ErrInvalidUUID, "Invalid node ID format")
	}
	
	nodePath := filepath.Join(l.basePath, "nodes", fmt.Sprintf("%s.json", nodeID))
	
	data, err := os.ReadFile(nodePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrNodeNotFound, "Node file not found")
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to read node file", err)
	}
	
	var node types.Node
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to parse node file", err)
	}
	
	return &node, nil
}

// SaveNode writes nodes/<id>.json
func (l *Loader) SaveNode(node *types.Node) error {
	if !id.IsValid(node.ID) {
		return errors.New(errors.ErrInvalidUUID, "Invalid node ID format")
	}
	
	node.UpdatedAt = time.Now()
	
	data, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return errors.WrapError(errors.ErrStorageFailure, "Failed to serialize node", err)
	}
	
	// Ensure nodes directory exists
	nodesDir := filepath.Join(l.basePath, "nodes")
	if err := os.MkdirAll(nodesDir, 0755); err != nil {
		return errors.WrapError(errors.ErrStorageFailure, "Failed to create nodes directory", err)
	}
	
	nodePath := filepath.Join(nodesDir, fmt.Sprintf("%s.json", node.ID))
	if err := os.WriteFile(nodePath, data, 0644); err != nil {
		return errors.WrapError(errors.ErrStorageFailure, "Failed to write node file", err)
	}
	
	return nil
}

// DeleteNode removes nodes/<id>.json
func (l *Loader) DeleteNode(nodeID string) error {
	if !id.IsValid(nodeID) {
		// If ID is invalid, treat as non-existent (no error)
		return nil
	}
	
	nodePath := filepath.Join(l.basePath, "nodes", fmt.Sprintf("%s.json", nodeID))
	if err := os.Remove(nodePath); err != nil && !os.IsNotExist(err) {
		return errors.WrapError(errors.ErrStorageFailure, "Failed to delete node file", err)
	}
	
	return nil
}

// ListNodeFiles returns all node IDs that have files in the nodes directory
func (l *Loader) ListNodeFiles() ([]string, error) {
	nodesDir := filepath.Join(l.basePath, "nodes")
	
	entries, err := os.ReadDir(nodesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to read nodes directory", err)
	}
	
	var nodeIDs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			nodeID := strings.TrimSuffix(name, ".json")
			if id.IsValid(nodeID) {
				nodeIDs = append(nodeIDs, nodeID)
			}
		}
	}
	
	return nodeIDs, nil
}

// NodeExists checks if a node file exists
func (l *Loader) NodeExists(nodeID string) bool {
	if !id.IsValid(nodeID) {
		return false
	}
	
	nodePath := filepath.Join(l.basePath, "nodes", fmt.Sprintf("%s.json", nodeID))
	_, err := os.Stat(nodePath)
	return err == nil
}

// ProjectExists checks if project.json exists
func (l *Loader) ProjectExists() bool {
	projectPath := filepath.Join(l.basePath, "project.json")
	_, err := os.Stat(projectPath)
	return err == nil
}
