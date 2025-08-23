# Migrations

This document describes Archon’s forward-only schema migration system, how and when migrations run, how to author new steps, and the project safety guarantees (backup + read-only enforcement).

## Overview
* __Forward-only__: Migrations only move projects from an older schema to a newer schema. No downgrades.
* __Idempotent__: Each step must be safe to apply multiple times. Steps check `IsApplied()` before mutating.
* __Complete coverage__: Every intermediate version between current project and target must have a registered step.
* __Safety__: A timestamped on-disk backup is created before any migration. Mutating APIs are read-only if the project is newer than the app supports.

Authoritative current version: `types.CurrentSchemaVersion` (see `internal/types/model.go`).

## When migrations run
`ProjectService.OpenProject()` inspects the opened project’s `schemaVersion` and takes one of three paths (see `internal/api/project_service.go`):
1. __Newer than app__: project version > `types.CurrentSchemaVersion` → project is opened __read-only__; all mutating API calls return `errors.ErrSchemaVersion`.
2. __Older than app__: project version < `types.CurrentSchemaVersion` → create a pre-migration backup, then run forward migrations via `migrate.Run()`.
3. __Equal__: no action required.

Post-migration, the project is reloaded and verified to equal `types.CurrentSchemaVersion`. Any mismatch returns `errors.ErrMigrationFailure`.

## Backup
Before any migration, a timestamped backup folder is created with `migrate.CreateBackup(basePath)` (see `internal/migrate/backup.go`).

Backup layout:
```
<project_root>/backups/<UTC-ISO8601>/
  project.json          # if present
  nodes/                # directory copied recursively (if present)
  attachments/          # directory copied recursively (if present)
```

Backups use UTC timestamps formatted as `YYYYMMDDTHHMMSSZ` (e.g., `20250823T182012Z`).

## Migration execution
Migrations are executed by `migrate.Run(basePath, current, target)` (see `internal/migrate/migrate.go`). Rules:
* __Coverage required__: Every version in `(current, target]` must have a registered step, otherwise `Run` returns an error.
* __Ordering__: Steps are run in ascending version order. The project is reloaded before each step.
* __Idempotency__: `IsApplied(ctx)` is checked; if true, the step is skipped.
* __Version bump enforcement__: After `Apply` succeeds, the project is reloaded and must have `SchemaVersion == step.Version()`; otherwise an error is returned.

## Authoring a migration step
Implement the `Step` interface and register it for the target version:

```go
// internal/migrate/step_v2.go
package migrate

type stepV2 struct{}

func (s *stepV2) Version() int { return 2 }
func (s *stepV2) Name() string { return "Add XYZ field" }

func (s *stepV2) IsApplied(ctx *Context) (bool, error) {
    // Determine whether project already satisfies v2 invariants
    return ctx.Project.SchemaVersion >= 2, nil
}

func (s *stepV2) Apply(ctx *Context) error {
    p := ctx.Project
    if p.SchemaVersion >= 2 {
        return nil // idempotent guard
    }
    // TODO: perform deterministic, forward-only mutations here.
    //       read/write using ctx.Loader (e.g., ctx.Loader.SaveProject(p)).
    p.SchemaVersion = 2
    return ctx.Loader.SaveProject(p)
}

func init() { Register(&stepV2{}) }
```

Guidelines:
* __Deterministic__: No randomization or environment-dependent behavior.
* __Minimal__: Only change what is necessary to reach the new invariants and bump `SchemaVersion`.
* __Idempotent__: Guard for already-upgraded projects and re-runs.
* __SchemaVersion__: Ensure `Apply()` persists `SchemaVersion` to the step’s target.

## Read-only enforcement for newer projects
When a project’s `SchemaVersion` is greater than the app’s `types.CurrentSchemaVersion`, Archon opens the project in read-only mode. Mutating endpoints return `errors.ErrSchemaVersion` (see corresponding services in `internal/api/`), including but not limited to:
* `NodeService` mutators (create/update/delete/move/reorder/property set/delete)
* `ProjectService.UpdateProjectSettings`
* `LoggingService.UpdateLoggingConfig`
* `ImportService.Run`
* `SnapshotService.Create`
* `IndexService.Rebuild`

This prevents corruption by writing with an older binary.

## Example: what happens on open
1. User selects a project at `<path>`.
2. `OpenProject` loads `project.json` and inspects `schemaVersion`.
3. If older, Archon:
   - Creates backup with `migrate.CreateBackup(<path>)`.
   - Calls `migrate.Run(<path>, project.SchemaVersion, types.CurrentSchemaVersion)`.
   - Reloads project and verifies the version equals `types.CurrentSchemaVersion`.
4. If newer, Archon enables read-only mode.

## Testing
Unit tests cover core behaviors (see `internal/migrate/`):
* `migrate_test.go`: missing step error, v1 migration success, idempotency, invalid bump detection.
* `backup_test.go`: backup folder creation and file copying.

Run tests:
```bash
go test ./internal/migrate -count=1
```

## Troubleshooting
* __Migration failed__: The project remains backed up under `backups/<timestamp>/`. Inspect the error (`errors.ErrMigrationFailure`) and logs; fix the bug and re-run.
* __Project opened read-only__: The project schema is newer than this app. Upgrade Archon to a version that supports the newer schema.

## References
* ADR-007: Data Migration and Schema Versioning — `docs/adr/ADR-007-data-migration-and-schema-versioning.md`
* Code:
  - `internal/migrate/migrate.go` — `Run`, `Step`, `Context`, registry
  - `internal/migrate/backup.go` — `CreateBackup`
  - `internal/api/project_service.go` — open flow, backup + migrate, read-only
  - `internal/types/model.go` — `CurrentSchemaVersion`
