package sqlite

// Incremental writer placeholder.

type Writer struct{}

func NewWriter(db *DB) *Writer { return &Writer{} }
