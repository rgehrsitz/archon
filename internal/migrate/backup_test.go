package migrate

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestCreateBackup_CopiesCoreData(t *testing.T) {
	base := t.TempDir()
	// Seed files
	writeFile(t, filepath.Join(base, "project.json"), "{\n  \"schemaVersion\": 1\n}\n")
	writeFile(t, filepath.Join(base, "nodes", "n1.json"), "{}\n")
	writeFile(t, filepath.Join(base, "attachments", "a1.bin"), "abc")

	backupDir, err := CreateBackup(base)
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	// Verify backup dir under base/backups/
	if filepath.Dir(backupDir) != filepath.Join(base, "backups") {
		t.Fatalf("backup dir parent mismatch: %s", backupDir)
	}
	// Verify files exist
	checks := []string{
		filepath.Join(backupDir, "project.json"),
		filepath.Join(backupDir, "nodes", "n1.json"),
		filepath.Join(backupDir, "attachments", "a1.bin"),
	}
	for _, p := range checks {
		if _, err := os.Stat(p); err != nil {
			if os.IsNotExist(err) {
				t.Fatalf("backup missing: %s", p)
			}
			// unexpected error
			t.Fatalf("stat backup file: %v", err)
		}
	}
	// Ensure permissions are directories where expected
	info, err := os.Stat(filepath.Join(backupDir, "nodes"))
	if err != nil || !info.IsDir() {
		t.Fatalf("nodes dir missing or not dir")
	}
	info, err = os.Stat(filepath.Join(backupDir, "attachments"))
	if err != nil || !info.IsDir() {
		t.Fatalf("attachments dir missing or not dir")
	}
	// Ensure walk doesn't error
	err = filepath.WalkDir(backupDir, func(path string, d fs.DirEntry, err error) error { return err })
	if err != nil {
		t.Fatalf("walk backup: %v", err)
	}
}
