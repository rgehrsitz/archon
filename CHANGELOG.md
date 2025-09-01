# Changelog

All notable changes to this project will be documented in this file.

## 2025-09-01

- **Frontend (UI)**: Implemented core workbench interface with Miller columns and TreeView navigation
  - Created `MillerColumns.svelte` with adaptive column layout, sliding behavior, and virtualized rows
  - Built `TreeView.svelte` with expandable hierarchy navigation and lazy loading
  - Added `CommandBar.svelte` with view mode toggle, breadcrumbs, and search functionality
  - Implemented `InspectorPanel.svelte` for node property editing
  - Added seamless view switching with persistent selection state and navigation reconstruction
- **Backend (Services)**: Enhanced project and dialog services for desktop integration
  - Implemented `DialogService` with native OS file dialog support for project selection
  - Fixed Wails serialization issues by simplifying backend method return types
  - Updated project service methods to remove error envelopes and improve frontend integration
- **Development**: Resolved all diagnostic issues and improved code quality
  - Fixed Svelte 5 compatibility issues and deprecated event dispatcher patterns
  - Updated module resolution configuration for proper UI component imports
  - Fixed Go test files to match updated API signatures
  - Achieved zero TypeScript/compilation errors across entire codebase

## 2025-08-24

- Backend (Plugins): Fixed compilation by standardizing zerolog chaining (`logger.Info().Msg(...)`), aligning index manager API calls with `internal/index/index.go`, and correcting `ProjectService.GetCurrentProject()` usage in `internal/api/plugin_service.go`.
- App wiring: Replaced nonexistent `logging.Global()` with `*logging.GetLogger()` in `app.go`.
- Docs: Updated `ROADMAP.md` to reflect current in-progress work (ADR-013 Backend Host Services alignment) and added status updates in plugin system implementation docs.

## 2025-08-23

- Snapshot system: Implemented creation (commit + immutable tag), metadata files in `.archon/snapshots/`, listing and restore by name. Tag creation uses Git CLI; tag listing uses go-git for speed.
- Git: Hybrid router now routes tag listing to go-git; CLI remains for porcelain ops.
- Index: Added automatic fallback to disable SQLite FTS index when FTS5 isn’t available (logs a warning). Tests/packages can also opt-out with `ARCHON_DISABLE_INDEX=1`.
- CLI: Added `archon diff` command with `--summary-only` and `--json` flags for machine-readable output.
