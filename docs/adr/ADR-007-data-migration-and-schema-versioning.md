# ADR-007: Data Migration & Schema Versioning

Status: Accepted
Date: 2025-08-21
Owners: Archon Core

## Context

The on-disk format will evolve. We need a safe process to open older projects, migrate forward, and protect users from accidental downgrades.

## Decision

- `project.json` contains `"schemaVersion": "<semver>"`.
- On open:
  1) If project version == app version ⇒ proceed.
  2) If project version < app version ⇒ run **forward migrations** in order.
  3) If project version > app version ⇒ open read-only with a clear warning; allow **export**, but no writes.
- **Backups:** Before migration, write a timestamped backup under `/backups/<ISO8601>/`.
- **Migrations:** Pure functions registered in `internal/migrate/migrate.go` with idempotent checks and unit tests. Each migration updates `schemaVersion`.
- **Compatibility window:** Guarantee read of N−1 minor versions (e.g., `1.(x-1).y`) at minimum.
- **Index:** `.archon/index/archon.db` is rebuild-only; never migrated.
- **Docs:** Migration notes summarized in `docs/migrations.md`.

## Rationale

Semver signaling + forward-only, reversible migrations are predictable and safe. Read-only for newer projects protects data integrity.

## Alternatives Considered

- **Arbitrary read/write across all versions:** Hard to guarantee correctness; overburdens tests.
- **No backups:** Risky; migration is inherently changeful.

## Consequences

- Positive: Safe upgrades, clear rules for compatibility.
- Negative: Some older projects may need an intermediate upgrade step.
- Follow-ups: CLI `archon migrate --from <path>`; migration report artifacts.

## Implementation Notes

- Migration registry: `internal/migrate/migrate.go` with `type Step func(*Project) error`.
- Backup writer: `internal/migrate/backup.go`.
- Tests: Golden files for representative projects across versions.

## Review / Revisit

- Revisit policy when introducing DAG/cross-links or breaking property typing.
