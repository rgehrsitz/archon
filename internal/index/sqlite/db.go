package sqlite

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

type DB struct {
	conn *sql.DB
	path string
}

func Open(projectRoot string) (*DB, error) {
	indexDir := filepath.Join(projectRoot, ".archon", "index")
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %w", err)
	}

	dbPath := filepath.Join(indexDir, "archon.db")
	conn, err := sql.Open("sqlite3", dbPath+"?_fk=1&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn, path: dbPath}

	if err := db.initializeSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func (d *DB) Close() error {
	if d.conn == nil {
		return nil
	}
	return d.conn.Close()
}

func (d *DB) initializeSchema() error {
	_, err := d.conn.Exec(schemaSQL)
	return err
}

func (d *DB) GetSchemaVersion() (string, error) {
	var version string
	err := d.conn.QueryRow("SELECT value FROM index_meta WHERE key = 'schema_version'").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("failed to get schema version: %w", err)
	}
	return version, nil
}

func (d *DB) Ping() error {
	return d.conn.Ping()
}

func (d *DB) GetConnection() *sql.DB {
	return d.conn
}
