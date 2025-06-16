// Package model provides the core data models for Archon.
package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// Component errors
var (
	ErrInvalidComponent = errors.New("invalid component")
	ErrInvalidID        = errors.New("invalid component ID")
	ErrDuplicateID      = errors.New("duplicate component ID")
	ErrComponentNotFound = errors.New("component not found")
)

// Component represents a physical asset in the Archon system
type Component struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	ParentID    string                 `json:"parentId,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Attachments []string               `json:"attachments,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// ComponentTree represents a hierarchical structure of components
type ComponentTree struct {
	Components map[string]*Component `json:"-"`
	RootIDs    []string              `json:"-"`
}

// ComponentSchema is the JSON schema for validating components
const ComponentSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["id", "name", "type"],
  "properties": {
    "id": {
      "type": "string",
      "pattern": "^[a-zA-Z0-9_-]+$",
      "minLength": 1,
      "maxLength": 64
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 256
    },
    "type": {
      "type": "string",
      "minLength": 1,
      "maxLength": 64
    },
    "description": {
      "type": "string"
    },
    "parentId": {
      "type": "string"
    },
    "properties": {
      "type": "object"
    },
    "attachments": {
      "type": "array",
      "items": {
        "type": "string"
      }
    },
    "metadata": {
      "type": "object",
      "additionalProperties": {
        "type": "string"
      }
    }
  },
  "additionalProperties": false
}`

// NewComponent creates a new component with the given ID, name, and type
func NewComponent(id, name, componentType string) *Component {
	return &Component{
		ID:         id,
		Name:       name,
		Type:       componentType,
		Properties: make(map[string]interface{}),
		Metadata:   make(map[string]string),
	}
}

// Validate checks if a component is valid according to the schema
func (c *Component) Validate() error {
	if c.ID == "" || c.Name == "" || c.Type == "" {
		return ErrInvalidComponent
	}

	// Validate ID format (alphanumeric, underscore, hyphen)
	if !isValidID(c.ID) {
		return ErrInvalidID
	}

	// Validate against JSON schema
	componentJSON, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal component: %w", err)
	}

	schemaLoader := gojsonschema.NewStringLoader(ComponentSchema)
	documentLoader := gojsonschema.NewBytesLoader(componentJSON)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		var errMsgs []string
		for _, desc := range result.Errors() {
			errMsgs = append(errMsgs, desc.String())
		}
		return fmt.Errorf("%w: %s", ErrInvalidComponent, strings.Join(errMsgs, "; "))
	}

	return nil
}

// isValidID checks if an ID contains only allowed characters
func isValidID(id string) bool {
	if len(id) == 0 || len(id) > 64 {
		return false
	}

	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

// NewComponentTree creates a new component tree from a list of components
func NewComponentTree(components []*Component) (*ComponentTree, error) {
	tree := &ComponentTree{
		Components: make(map[string]*Component),
	}

	// First pass: add all components to the map and check for duplicates
	for _, c := range components {
		if _, exists := tree.Components[c.ID]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateID, c.ID)
		}
		tree.Components[c.ID] = c
	}

	// Second pass: validate parent references and build root list
	for _, c := range components {
		if c.ParentID != "" {
			if _, exists := tree.Components[c.ParentID]; !exists {
				return nil, fmt.Errorf("parent component not found: %s", c.ParentID)
			}
		} else {
			tree.RootIDs = append(tree.RootIDs, c.ID)
		}
	}

	return tree, nil
}

// GetComponent returns a component by ID
func (t *ComponentTree) GetComponent(id string) (*Component, error) {
	component, exists := t.Components[id]
	if !exists {
		return nil, ErrComponentNotFound
	}
	return component, nil
}

// GetChildren returns the direct children of a component
func (t *ComponentTree) GetChildren(parentID string) []*Component {
	var children []*Component
	for _, c := range t.Components {
		if c.ParentID == parentID {
			children = append(children, c)
		}
	}
	return children
}

// GetRoots returns the root components (those without a parent)
func (t *ComponentTree) GetRoots() []*Component {
	var roots []*Component
	for _, id := range t.RootIDs {
		if component, exists := t.Components[id]; exists {
			roots = append(roots, component)
		}
	}
	return roots
}

// AddComponent adds a component to the tree
func (t *ComponentTree) AddComponent(c *Component) error {
	if _, exists := t.Components[c.ID]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateID, c.ID)
	}

	// Validate parent reference if specified
	if c.ParentID != "" {
		if _, exists := t.Components[c.ParentID]; !exists {
			return fmt.Errorf("parent component not found: %s", c.ParentID)
		}
	} else {
		t.RootIDs = append(t.RootIDs, c.ID)
	}

	t.Components[c.ID] = c
	return nil
}

// RemoveComponent removes a component and its children from the tree
func (t *ComponentTree) RemoveComponent(id string) error {
	if _, exists := t.Components[id]; !exists {
		return ErrComponentNotFound
	}

	// Remove from root IDs if it's a root
	for i, rootID := range t.RootIDs {
		if rootID == id {
			t.RootIDs = append(t.RootIDs[:i], t.RootIDs[i+1:]...)
			break
		}
	}

	// Find and remove all children recursively
	var childrenToRemove []string
	for _, c := range t.Components {
		if c.ParentID == id {
			childrenToRemove = append(childrenToRemove, c.ID)
		}
	}

	// Delete the component
	delete(t.Components, id)

	// Recursively remove children
	for _, childID := range childrenToRemove {
		_ = t.RemoveComponent(childID) // Ignore errors as we're already removing
	}

	return nil
}

// UpdateParent changes a component's parent
func (t *ComponentTree) UpdateParent(id, newParentID string) error {
	component, exists := t.Components[id]
	if !exists {
		return ErrComponentNotFound
	}

	// Check if new parent exists (if not empty)
	if newParentID != "" {
		if _, exists := t.Components[newParentID]; !exists {
			return fmt.Errorf("new parent component not found: %s", newParentID)
		}

		// Check for circular reference
		parentID := newParentID
		for parentID != "" {
			if parentID == id {
				return errors.New("circular parent reference detected")
			}
			parent, exists := t.Components[parentID]
			if !exists {
				break
			}
			parentID = parent.ParentID
		}
	}

	// If component was a root and now has a parent, remove from roots
	if component.ParentID == "" && newParentID != "" {
		for i, rootID := range t.RootIDs {
			if rootID == id {
				t.RootIDs = append(t.RootIDs[:i], t.RootIDs[i+1:]...)
				break
			}
		}
	}

	// If component now becomes a root, add to roots
	if component.ParentID != "" && newParentID == "" {
		t.RootIDs = append(t.RootIDs, id)
	}

	// Update the parent ID
	component.ParentID = newParentID
	return nil
}
