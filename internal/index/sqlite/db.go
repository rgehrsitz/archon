package sqlite

// db.go: Placeholder for SQLite index management per ADR-005.
// Real implementation will open .archon/index/archon.db and manage lifecycle.

type DB struct{}

func Open(path string) (*DB, error) {
	// TODO: implement using SQLite driver
	return &DB{}, nil
}

func (d *DB) Close() error { return nil }
