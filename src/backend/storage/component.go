// Package storage defines the Component type and serialization for components.json.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

// Component represents a single node in the Archon hierarchy.
type Component struct {
	ID         string                 `json:"id"`           // Unique identifier
	Name       string                 `json:"name"`         // Human-readable name
	Type       string                 `json:"type"`         // Component type (e.g., rack, device)
	ParentID   string                 `json:"parentId,omitempty"` // Parent component, empty for root
	Properties map[string]interface{} `json:"properties,omitempty"` // Arbitrary key-value pairs
	Children   []string               `json:"children,omitempty"`   // IDs of child components
}

// LoadComponents loads components from the given file path (components.json).
func LoadComponents(path string) ([]Component, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read components file: %w", err)
	}
	var components []Component
	if err := json.Unmarshal(data, &components); err != nil {
		return nil, fmt.Errorf("failed to parse components: %w", err)
	}
	return components, nil
}

// SaveComponents saves the components slice to the given file path (components.json).
func SaveComponents(path string, components []Component) error {
	data, err := json.MarshalIndent(components, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize components: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write components file: %w", err)
	}
	return nil
}

// (Optional) ValidateComponent can be extended to enforce required fields or schema rules.
func ValidateComponent(c *Component) error {
	if c.ID == "" {
		return fmt.Errorf("component ID is required")
	}
	if c.Name == "" {
		return fmt.Errorf("component name is required")
	}
	return nil
}
