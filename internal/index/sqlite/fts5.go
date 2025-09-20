//go:build fts5

// Package sqlite provides FTS5-enabled SQLite functionality.
// This file ensures FTS5 is always compiled in when the fts5 build tag is present.
package sqlite

// FTS5Enabled indicates that FTS5 is available in this build.
const FTS5Enabled = true
