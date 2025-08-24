# Changelog

All notable changes to this project will be documented in this file.

## 2025-08-24

- Backend (Plugins): Fixed compilation by standardizing zerolog chaining (`logger.Info().Msg(...)`), aligning index manager API calls with `internal/index/index.go`, and correcting `ProjectService.GetCurrentProject()` usage in `internal/api/plugin_service.go`.
- App wiring: Replaced nonexistent `logging.Global()` with `*logging.GetLogger()` in `app.go`.
- Docs: Updated `ROADMAP.md` to reflect current in-progress work (ADR-013 Backend Host Services alignment) and added status updates in plugin system implementation docs.

## 2025-08-23

- Snapshot system: Implemented creation (commit + immutable tag), metadata files in `.archon/snapshots/`, listing and restore by name. Tag creation uses Git CLI; tag listing uses go-git for speed.
- Git: Hybrid router now routes tag listing to go-git; CLI remains for porcelain ops.
- Index: Added automatic fallback to disable SQLite FTS index when FTS5 isn’t available (logs a warning). Tests/packages can also opt-out with `ARCHON_DISABLE_INDEX=1`.
- CLI: Added `archon diff` command with `--summary-only` and `--json` flags for machine-readable output.
