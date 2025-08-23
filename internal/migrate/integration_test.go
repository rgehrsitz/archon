package migrate

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/store"
	"github.com/rgehrsitz/archon/internal/types"
)

// TestCompleteBackupAndMigrationFlow tests the full end-to-end migration workflow
func TestCompleteBackupAndMigrationFlow(t *testing.T) {
	dir := t.TempDir()
	
	// Create a project with schema 0 and some data
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "nodes"), 0o755); err != nil {
		t.Fatalf("mkdir nodes: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "attachments"), 0o755); err != nil {
		t.Fatalf("mkdir attachments: %v", err)
	}
	
	// Create project.json with schema 0
	ldr := store.NewLoader(dir)
	project := &types.Project{
		RootID:        "00000000-0000-0000-0000-000000000000",
		SchemaVersion: 0,
		Settings:      map[string]any{"testKey": "testValue"},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := ldr.SaveProject(project); err != nil {
		t.Fatalf("save project: %v", err)
	}
	
	// Create some nodes and attachments
	nodeFile := filepath.Join(dir, "nodes", "node1.json")
	if err := os.WriteFile(nodeFile, []byte(`{"id":"node1","name":"Test Node"}`), 0o644); err != nil {
		t.Fatalf("write node: %v", err)
	}
	
	attachmentFile := filepath.Join(dir, "attachments", "file1.txt")
	if err := os.WriteFile(attachmentFile, []byte("test attachment content"), 0o644); err != nil {
		t.Fatalf("write attachment: %v", err)
	}
	
	// Record original modification time
	origStat, err := os.Stat(filepath.Join(dir, "project.json"))
	if err != nil {
		t.Fatalf("stat original project: %v", err)
	}
	
	// Create backup first (as API does)
	if _, err := CreateBackup(dir); err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	
	// Execute migration
	if err := Run(dir, 0, types.CurrentSchemaVersion); err != nil {
		t.Fatalf("Run migration: %v", err)
	}
	
	// Verify project was migrated
	migratedProject, err := ldr.LoadProject()
	if err != nil {
		t.Fatalf("load migrated project: %v", err)
	}
	if migratedProject.SchemaVersion != types.CurrentSchemaVersion {
		t.Errorf("expected schema %d, got %d", types.CurrentSchemaVersion, migratedProject.SchemaVersion)
	}
	
	// Verify settings were preserved
	if migratedProject.Settings["testKey"] != "testValue" {
		t.Errorf("project settings not preserved during migration")
	}
	
	// Verify backup was created
	backupsDir := filepath.Join(dir, "backups")
	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		t.Fatalf("read backups dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 backup, got %d", len(entries))
	}
	
	backupPath := filepath.Join(backupsDir, entries[0].Name())
	
	// Verify backup contains all expected files
	expectedBackupFiles := []string{
		"project.json",
		"nodes/node1.json", 
		"attachments/file1.txt",
	}
	
	for _, expectedFile := range expectedBackupFiles {
		backupFile := filepath.Join(backupPath, expectedFile)
		if _, err := os.Stat(backupFile); err != nil {
			if os.IsNotExist(err) {
				t.Errorf("backup missing file: %s", expectedFile)
			} else {
				t.Errorf("stat backup file %s: %v", expectedFile, err)
			}
		}
	}
	
	// Verify backup project.json has original schema
	backupLoader := store.NewLoader(backupPath)
	backupProject, err := backupLoader.LoadProject()
	if err != nil {
		t.Fatalf("load backup project: %v", err)
	}
	if backupProject.SchemaVersion != 0 {
		t.Errorf("backup should have original schema 0, got %d", backupProject.SchemaVersion)
	}
	
	// Verify backup was created before migration (timestamp check)
	backupStat, err := os.Stat(filepath.Join(backupPath, "project.json"))
	if err != nil {
		t.Fatalf("stat backup project: %v", err)
	}
	
	currentStat, err := os.Stat(filepath.Join(dir, "project.json"))
	if err != nil {
		t.Fatalf("stat current project: %v", err)
	}
	
	// Backup should be older than current (backup was made before migration)
	if !backupStat.ModTime().Before(currentStat.ModTime()) {
		t.Errorf("backup timestamp should be before current project timestamp")
	}
	
	// Original should be different from current (migration updated it)
	if !origStat.ModTime().Before(currentStat.ModTime()) {
		t.Errorf("current project should be newer than original")
	}
}

// TestBackupTimestampFormat verifies backup directory naming follows expected format
func TestBackupTimestampFormat(t *testing.T) {
	dir := t.TempDir()
	
	// Create minimal project
	if err := os.WriteFile(filepath.Join(dir, "project.json"), []byte(`{"schemaVersion":0}`), 0o644); err != nil {
		t.Fatalf("write project: %v", err)
	}
	
	// Create backup
	backupDir, err := CreateBackup(dir)
	if err != nil {
		t.Fatalf("CreateBackup: %v", err)
	}
	
	// Verify timestamp format: YYYYMMDDTHHMMSSZ
	backupName := filepath.Base(backupDir)
	if len(backupName) != 16 {
		t.Errorf("expected timestamp length 16, got %d: %s", len(backupName), backupName)
	}
	
	// Verify it's a valid timestamp
	_, err = time.Parse("20060102T150405Z", backupName)
	if err != nil {
		t.Errorf("invalid timestamp format: %s, error: %v", backupName, err)
	}
	
	// Verify backup is in backups subdirectory
	expected := filepath.Join(dir, "backups", backupName)
	if backupDir != expected {
		t.Errorf("expected backup path %s, got %s", expected, backupDir)
	}
}

// TestMultipleBackupsPreserveOrder verifies multiple backups maintain chronological order
func TestMultipleBackupsPreserveOrder(t *testing.T) {
	dir := t.TempDir()
	
	// Create minimal project
	if err := os.WriteFile(filepath.Join(dir, "project.json"), []byte(`{"schemaVersion":0}`), 0o644); err != nil {
		t.Fatalf("write project: %v", err)
	}
	
	// Create first backup
	backup1, err := CreateBackup(dir)
	if err != nil {
		t.Fatalf("CreateBackup 1: %v", err)
	}
	
	// Small delay to ensure different timestamps
	time.Sleep(1 * time.Second)
	
	// Create second backup
	backup2, err := CreateBackup(dir)
	if err != nil {
		t.Fatalf("CreateBackup 2: %v", err)
	}
	
	// Verify backup1 timestamp < backup2 timestamp
	name1 := filepath.Base(backup1)
	name2 := filepath.Base(backup2)
	
	time1, err := time.Parse("20060102T150405Z", name1)
	if err != nil {
		t.Fatalf("parse backup1 time: %v", err)
	}
	
	time2, err := time.Parse("20060102T150405Z", name2)
	if err != nil {
		t.Fatalf("parse backup2 time: %v", err)
	}
	
	if !time1.Before(time2) {
		t.Errorf("backup1 (%s) should be before backup2 (%s)", name1, name2)
	}
}