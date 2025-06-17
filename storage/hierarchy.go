// Package storage provides functionality for managing Archon project storage.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Hierarchy errors
var (
	ErrComponentNotFound = errors.New("component not found")
	ErrInvalidParent     = errors.New("invalid parent component")
	ErrCircularReference = errors.New("circular reference detected")
	ErrDuplicateID       = errors.New("duplicate component ID")
)

// Hierarchy represents the component hierarchy structure
type Hierarchy struct {
	RootID     string               `json:"rootId"`
	Components map[string]Component `json:"components"`
}

// NewHierarchy creates a new empty hierarchy
func NewHierarchy() *Hierarchy {
	return &Hierarchy{
		Components: make(map[string]Component),
	}
}

// LoadHierarchy loads the component hierarchy from the project
func (v *ConfigVault) LoadHierarchy() (*Hierarchy, error) {
	data, err := os.ReadFile(v.GetComponentsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return NewHierarchy(), nil
		}
		return nil, fmt.Errorf("failed to read components file: %w", err)
	}

	var hierarchy Hierarchy
	if err := json.Unmarshal(data, &hierarchy); err != nil {
		return nil, fmt.Errorf("failed to parse components: %w", err)
	}

	return &hierarchy, nil
}

// SaveHierarchy saves the component hierarchy to the project
func (v *ConfigVault) SaveHierarchy(h *Hierarchy) error {
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize hierarchy: %w", err)
	}

	if err := os.WriteFile(v.GetComponentsPath(), data, 0o644); err != nil {
		return fmt.Errorf("failed to write components file: %w", err)
	}
	v.changeCount++
	v.autoSnapshot("")
	return nil
}

// AddComponent adds a new component to the hierarchy
func (h *Hierarchy) AddComponent(comp Component) error {
	// Validate component
	if err := ValidateComponent(&comp); err != nil {
		return err
	}

	// Check for duplicate ID
	if _, exists := h.Components[comp.ID]; exists {
		return ErrDuplicateID
	}

	// If this is the first component, make it the root
	if len(h.Components) == 0 {
		h.RootID = comp.ID
		comp.ParentID = "" // Root has no parent
	} else if comp.ParentID != "" {
		// Validate parent exists
		parent, exists := h.Components[comp.ParentID]
		if !exists {
			return ErrInvalidParent
		}

		// Add to parent's children
		parent.Children = append(parent.Children, comp.ID)
		h.Components[parent.ID] = parent
	}

	// Add component
	h.Components[comp.ID] = comp
	return nil
}

// UpdateComponent updates an existing component
func (h *Hierarchy) UpdateComponent(comp Component) error {
	// Validate component
	if err := ValidateComponent(&comp); err != nil {
		return err
	}

	// Check if component exists
	oldComp, exists := h.Components[comp.ID]
	if !exists {
		return ErrComponentNotFound
	}

	// If parent changed, update parent references
	if oldComp.ParentID != comp.ParentID {
		// Remove from old parent's children
		if oldComp.ParentID != "" {
			oldParent := h.Components[oldComp.ParentID]
			oldParent.Children = removeFromSlice(oldParent.Children, comp.ID)
			h.Components[oldParent.ID] = oldParent
		}

		// Add to new parent's children
		if comp.ParentID != "" {
			newParent, exists := h.Components[comp.ParentID]
			if !exists {
				return ErrInvalidParent
			}
			newParent.Children = append(newParent.Children, comp.ID)
			h.Components[newParent.ID] = newParent
		}
	}

	// Update component
	h.Components[comp.ID] = comp
	return nil
}

// DeleteComponent removes a component from the hierarchy
func (h *Hierarchy) DeleteComponent(id string) error {
	comp, exists := h.Components[id]
	if !exists {
		return ErrComponentNotFound
	}

	// Don't allow deleting the root component
	if id == h.RootID {
		return errors.New("cannot delete root component")
	}

	// Remove from parent's children
	if comp.ParentID != "" {
		parent := h.Components[comp.ParentID]
		parent.Children = removeFromSlice(parent.Children, id)
		h.Components[parent.ID] = parent
	}

	// Delete component
	delete(h.Components, id)
	return nil
}

// GetComponent retrieves a component by ID
func (h *Hierarchy) GetComponent(id string) (Component, error) {
	comp, exists := h.Components[id]
	if !exists {
		return Component{}, ErrComponentNotFound
	}
	return comp, nil
}

// GetChildren returns all child components of a given component
func (h *Hierarchy) GetChildren(id string) ([]Component, error) {
	comp, exists := h.Components[id]
	if !exists {
		return nil, ErrComponentNotFound
	}

	children := make([]Component, 0, len(comp.Children))
	for _, childID := range comp.Children {
		child, exists := h.Components[childID]
		if !exists {
			continue // Skip missing children
		}
		children = append(children, child)
	}

	return children, nil
}

// GetAncestors returns all ancestor components of a given component
func (h *Hierarchy) GetAncestors(id string) ([]Component, error) {
	comp, exists := h.Components[id]
	if !exists {
		return nil, ErrComponentNotFound
	}

	ancestors := make([]Component, 0)
	current := comp
	for current.ParentID != "" {
		parent, exists := h.Components[current.ParentID]
		if !exists {
			break // Stop if parent is missing
		}
		ancestors = append(ancestors, parent)
		current = parent
	}

	return ancestors, nil
}

// Helper function to remove an item from a string slice
func removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
