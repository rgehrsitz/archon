package merge

import semdiff "github.com/rgehrsitz/archon/internal/diff/semantic"

// Conflict captures a conflicting change on the same logical field
type Conflict struct {
	NodeID string
	Field  string // name|parent|property:<key>|order
	Ours   any
	Theirs any
	Rule   string // reason/category
}

// Resolution summarizes a 3-way merge attempt 
type Resolution struct {
	Base      string
	Ours      string
	Theirs    string
	Conflicts []Conflict
	// Changes that were non-conflicting 
	OursOnly   []semdiff.Change
	TheirsOnly []semdiff.Change
	// Applied changes (when Apply is called)
	Applied []semdiff.Change
}
