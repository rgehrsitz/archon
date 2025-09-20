package sqlite

import (
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// skipIfNoFTS5 skips the test if the SQLite driver does not have FTS5 enabled.
func skipIfNoFTS5(t *testing.T) {
	t.Helper()
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		// If we cannot even open, skip to avoid cascading failures in environments without sqlite
		t.Skipf("sqlite not available: %v", err)
		return
	}
	defer conn.Close()
	if _, err := conn.Exec("CREATE VIRTUAL TABLE t USING fts5(content)"); err != nil {
		if strings.Contains(err.Error(), "no such module: fts5") {
			t.Skip("SQLite FTS5 not available; skipping FTS-dependent tests")
			return
		}
		// Other errors indicate something else wrong; fail fast
		t.Fatalf("unexpected SQLite error while probing FTS5: %v", err)
	}
}
