package merge

type ChangeType string

const (
	ChangeRename   ChangeType = "rename"
	ChangeMove     ChangeType = "move"
	ChangeProperty ChangeType = "property"
	ChangeStructure ChangeType = "structure"
)

type Change struct {
	Type      ChangeType
	NodeID    string
	Field     string
	OldValue  any
	NewValue  any
}

type Conflict struct {
	NodeID   string
	Field    string
	Ours     any
	Theirs   any
	Rule     string
}
