package api

import (
    "os"
    "path/filepath"
    "testing"
    "time"

    "github.com/rgehrsitz/archon/internal/errors"
    "github.com/rgehrsitz/archon/internal/id"
    "github.com/rgehrsitz/archon/internal/store"
    "github.com/rgehrsitz/archon/internal/types"
)

func seedProject(t *testing.T, base string, schema int) {
    t.Helper()
    if err := os.MkdirAll(base, 0o755); err != nil {
        t.Fatalf("mkdir base: %v", err)
    }
    ldr := store.NewLoader(base)
    p := &types.Project{
        RootID:        id.NewV7(),
        SchemaVersion: schema,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }
    if err := ldr.SaveProject(p); err != nil {
        t.Fatalf("save project.json: %v", err)
    }
}

// TestMain disables the SQLite index for this package's tests to avoid requiring FTS5.
func TestMain(m *testing.M) {
    os.Setenv("ARCHON_DISABLE_INDEX", "1")
    os.Exit(m.Run())
}

func TestOpenProject_NewerSchema_ReadOnlyAndNoBackup(t *testing.T) {
    base := t.TempDir()
    // Create a project with a newer schema than supported
    seedProject(t, base, types.CurrentSchemaVersion+1)

    svc := NewProjectService()
    proj := svc.OpenProject(base)
    if proj == nil {
        t.Fatalf("OpenProject returned nil project")
    }
    if proj.SchemaVersion != types.CurrentSchemaVersion+1 {
        t.Fatalf("unexpected project schema: got %d", proj.SchemaVersion)
    }

    // Verify read-only flag is true via GetProjectInfo
    info, env := svc.GetProjectInfo()
    if env.Code != "" {
        t.Fatalf("GetProjectInfo error: %s - %s", env.Code, env.Message)
    }
    if ro, _ := info["readOnly"].(bool); !ro {
        t.Fatalf("expected readOnly=true for newer schema")
    }

    // Writes should be rejected
    env = svc.UpdateProjectSettings(map[string]any{"k": "v"})
    if env.Code != errors.ErrSchemaVersion {
        t.Fatalf("expected ErrSchemaVersion on write in read-only mode, got: %s", env.Code)
    }

    // Ensure no backup directory was created
    if _, err := os.Stat(filepath.Join(base, "backups")); err == nil {
        t.Fatalf("unexpected backups/ directory created for newer schema")
    } else if !os.IsNotExist(err) {
        t.Fatalf("stat backups/: %v", err)
    }
}

func TestOpenProject_EqualSchema_NoMigrationNoBackup(t *testing.T) {
    base := t.TempDir()
    seedProject(t, base, types.CurrentSchemaVersion)

    svc := NewProjectService()
    proj := svc.OpenProject(base)
    if proj == nil {
        t.Fatalf("OpenProject returned nil project")
    }
    if proj.SchemaVersion != types.CurrentSchemaVersion {
        t.Fatalf("unexpected project schema: got %d", proj.SchemaVersion)
    }

    info, env := svc.GetProjectInfo()
    if env.Code != "" {
        t.Fatalf("GetProjectInfo error: %s - %s", env.Code, env.Message)
    }
    if ro, _ := info["readOnly"].(bool); ro {
        t.Fatalf("did not expect readOnly for equal schema")
    }

    // Ensure no backup directory was created
    if _, err := os.Stat(filepath.Join(base, "backups")); err == nil {
        t.Fatalf("unexpected backups/ directory created for equal schema")
    } else if !os.IsNotExist(err) {
        t.Fatalf("stat backups/: %v", err)
    }
}

func TestOpenProject_OlderSchema_TriggersBackupAndMigration(t *testing.T) {
    base := t.TempDir()
    // Seed legacy project with schema 0
    seedProject(t, base, 0)

    svc := NewProjectService()
    proj := svc.OpenProject(base)
    if proj == nil {
        t.Fatalf("OpenProject returned nil project")
    }
    // After migration, schema should be current and project not read-only
    if proj.SchemaVersion != types.CurrentSchemaVersion {
        t.Fatalf("expected schema migrated to %d, got %d", types.CurrentSchemaVersion, proj.SchemaVersion)
    }
    info, env := svc.GetProjectInfo()
    if env.Code != "" {
        t.Fatalf("GetProjectInfo error: %s - %s", env.Code, env.Message)
    }
    if ro, _ := info["readOnly"].(bool); ro {
        t.Fatalf("did not expect readOnly after successful migration")
    }
    // Backup directory should exist with at least project.json copied
    backupsDir := filepath.Join(base, "backups")
    entries, err := os.ReadDir(backupsDir)
    if err != nil {
        t.Fatalf("expected backups directory to exist: %v", err)
    }
    if len(entries) == 0 {
        t.Fatalf("expected at least one timestamped backup directory")
    }
    // Check the latest backup dir contains project.json
    backupPath := filepath.Join(backupsDir, entries[0].Name())
    if _, err := os.Stat(filepath.Join(backupPath, "project.json")); err != nil {
        t.Fatalf("expected project.json in backup: %v", err)
    }
}

// NOTE: A full integration test for the older-schema path (backup + forward migrations)
// requires either allowing schemaVersion==0 through validation in store.ValidateProject
// or introducing a newer target version (> CurrentSchemaVersion) with a registered step.
// As of now, CurrentSchemaVersion == 1 and validation rejects schemaVersion==0, so such
// a test cannot pass without adjusting the model/steps.
