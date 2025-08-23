package types

import "time"

const CurrentSchemaVersion = 1

// Project represents the root project metadata stored in project.json
// Schema is forward-migrated (ADR-007).
type Project struct {
	RootID        string         `json:"rootId"`
	SchemaVersion int            `json:"schemaVersion"`
	Settings      map[string]any `json:"settings,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

// Property represents a typed key-value property on a node
type Property struct {
	TypeHint string `json:"typeHint,omitempty"` // "string", "number", "boolean", "date", "attachment"
	Value    any    `json:"value"`
}

// Attachment represents a content-addressed file reference
type Attachment struct {
	Type     string `json:"type"`     // Always "attachment"
	Hash     string `json:"hash"`     // Content hash
	Filename string `json:"filename"` // Original filename
	Size     int64  `json:"size"`     // File size in bytes
}

// Node is the authoritative node record stored in nodes/<id>.json
// Child order is meaningful; sibling names must be unique among siblings.
type Node struct {
	ID          string              `json:"id"`          // UUIDv7
	Name        string              `json:"name"`        // Must be unique among siblings
	Description string              `json:"description,omitempty"`
	Properties  map[string]Property `json:"properties,omitempty"`
	Children    []string            `json:"children"`    // Ordered list of child IDs
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// CreateNodeRequest represents a request to create a new node
type CreateNodeRequest struct {
	ParentID    string            `json:"parentId"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Properties  map[string]Property `json:"properties,omitempty"`
}

// UpdateNodeRequest represents a request to update an existing node
type UpdateNodeRequest struct {
	ID          string              `json:"id"`
	Name        *string             `json:"name,omitempty"`        // Pointer allows distinction between empty and unset
	Description *string             `json:"description,omitempty"`
	Properties  map[string]Property `json:"properties,omitempty"`
}

// MoveNodeRequest represents a request to move a node to a new parent
type MoveNodeRequest struct {
	NodeID      string `json:"nodeId"`
	NewParentID string `json:"newParentId"`
	Position    int    `json:"position,omitempty"` // Optional position in new parent's children
}

// ReorderChildrenRequest represents a request to reorder a parent's children
type ReorderChildrenRequest struct {
	ParentID         string   `json:"parentId"`
	OrderedChildIDs []string `json:"orderedChildIds"`
}
