package types

// Project represents the root project metadata stored in project.json
// Schema is forward-migrated (ADR-007).
type Project struct {
	RootID        string         `json:"rootId"`
	SchemaVersion int            `json:"schemaVersion"`
	Settings      map[string]any `json:"settings,omitempty"`
}

type Property struct {
	Key      string `json:"key"`
	TypeHint string `json:"typeHint,omitempty"`
	Value    any    `json:"value"`
}

// Node is the authoritative node record stored in nodes/<id>.json
// Child order is meaningful; sibling names must be unique.
type Node struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Properties  map[string]Property `json:"properties,omitempty"`
	Children    []string            `json:"children"`
}
