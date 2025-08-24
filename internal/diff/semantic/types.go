package semantic

import "encoding/json"

// ChangeType enumerates semantic change categories
//
// NodeAdded: a node exists in B but not in A
// NodeRemoved: a node exists in A but not in B
// NodeRenamed: name changed (ID stable, parent unchanged)
// NodeMoved: parent changed (ID stable)
// PropertyChanged: any non-name metadata changed (description/properties)
// OrderChanged: a parent's children ordering changed
// AttachmentChanged: reserved for future use (ADR-002)
//
// Additional types can be added without breaking consumers if they ignore unknown types.
type ChangeType string

const (
	ChangeNodeAdded         ChangeType = "NodeAdded"
	ChangeNodeRemoved       ChangeType = "NodeRemoved"
	ChangeNodeRenamed       ChangeType = "NodeRenamed"
	ChangeNodeMoved         ChangeType = "NodeMoved"
	ChangePropertyChanged   ChangeType = "PropertyChanged"
	ChangeOrderChanged      ChangeType = "OrderChanged"
	ChangeAttachmentChanged ChangeType = "AttachmentChanged"
)

// Change represents a single semantic change between two refs
// Fields are optional depending on Type
// Kept simple for v1; can extend later.
type Change struct {
	Type       ChangeType `json:"type"`
	NodeID     string     `json:"nodeId,omitempty"`
	ParentID   string     `json:"parentId,omitempty"` // for OrderChanged (the parent whose order changed)
	NameFrom   string     `json:"nameFrom,omitempty"`
	NameTo     string     `json:"nameTo,omitempty"`
	ParentFrom string     `json:"parentFrom,omitempty"`
	ParentTo   string     `json:"parentTo,omitempty"`
	// Property diffs: changed keys (added/updated/removed)
	ChangedProperties []PropertyDelta `json:"changedProperties,omitempty"`
	// Order diffs: previous and next ordering (full lists for simplicity)
	OrderFrom []string `json:"orderFrom,omitempty"`
	OrderTo   []string `json:"orderTo,omitempty"`
}

// PropertyDelta describes a single property change
// Kind: added|removed|updated
// Old/New are omitted when not applicable

type PropertyDelta struct {
	Key  string          `json:"key"`
	Kind string          `json:"kind"`
	Old  json.RawMessage `json:"old,omitempty"`
	New  json.RawMessage `json:"new,omitempty"`
}

// Summary aggregates counts by change type
// Useful for CLI summaries and UI badges

type Summary struct {
	Total             int `json:"total"`
	Added             int `json:"added"`
	Removed           int `json:"removed"`
	Renamed           int `json:"renamed"`
	Moved             int `json:"moved"`
	PropertyChanged   int `json:"propertyChanged"`
	OrderChanged      int `json:"orderChanged"`
	AttachmentChanged int `json:"attachmentChanged"`
}

// Result is the top-level semantic diff payload

type Result struct {
	From    string   `json:"from"`
	To      string   `json:"to"`
	Changes []Change `json:"changes"`
	Summary Summary  `json:"summary"`
}
