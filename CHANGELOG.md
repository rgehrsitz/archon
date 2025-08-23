# Changelog

All notable changes to this project will be documented in this file.

## 2025-08-23

- Snapshot system: Implemented creation (commit + immutable tag), metadata files in `.archon/snapshots/`, listing and restore by name. Tag creation uses Git CLI; tag listing uses go-git for speed.
- Git: Hybrid router now routes tag listing to go-git; CLI remains for porcelain ops.
- Index: Added automatic fallback to disable SQLite FTS index when FTS5 isn’t available (logs a warning). Tests/packages can also opt-out with `ARCHON_DISABLE_INDEX=1`.
