# Archon Development Roadmap

This document tracks progress against the complete Archon vision as defined in the ADRs and product specification.

## ✅ Completed (Foundation Layer)

### Core Data Model

- [x] UUIDv7 ID generation with time-sortable properties (ADR-001)
- [x] Hierarchical node structure with meaningful child ordering
- [x] Sharded JSON storage (one file per node) (ADR-002)
- [x] Property system with type hints (string, number, boolean, date, attachment)
- [x] Sibling name uniqueness validation (case-insensitive)
- [x] Schema versioning foundation (ADR-007)

### Storage Layer

- [x] Project operations (create, open, save, validate)
- [x] Node CRUD operations with validation
- [x] Parent-child relationship management
- [x] Node move and reorder operations
- [x] Property management (set, delete, update)
- [x] Atomic file operations with rollback
- [x] Comprehensive error envelope system

### Wails Integration

- [x] ProjectService API with context handling
- [x] NodeService API with full node operations
- [x] TypeScript frontend wrappers
- [x] Error propagation from Go to frontend
- [x] Service binding in main application

### Quality Assurance

- [x] UUID v7 generator tests (format, uniqueness, time ordering)
- [x] Error envelope system tests
- [x] Validation system tests (nodes, projects, requests)
- [x] Storage layer integration tests
- [x] Example project validation
- [x] End-to-end manipulation tests

### Development Infrastructure

- [x] Example "basic-hierarchy" project (manufacturing plant)
- [x] Proper module naming (github.com/rgehrsitz/archon)
- [x] CLAUDE.md documentation for AI assistance
- [x] CI workflow with GitHub Actions (vet, staticcheck, build, tests)

## 🔄 In Progress

Currently no active development tasks.

## 📋 Pending (Prioritized Roadmap)

### Phase 1: Data Layer Extensions

- [x] **SQLite Search Index** - Fast search with FTS5 (ADR-005)
  - Rebuildable index in `/.archon/index/archon.db`
  - Full-text search on names, descriptions, properties
  - Incremental updates on node changes
  - Index health monitoring and rebuild capability

- [x] **Logging System** - Rotating logs with error envelopes (ADR-006)
  - Structured logging with zerolog
  - Rotating files in `logs/` directory (10MB x 5 files)
  - Error correlation and debugging support
  - Configurable log levels
  - Recent logs retrieval with rotation awareness (includes `.gz`)

- [x] **Schema Migration System** - Forward-only migrations with backup (ADR-007)
  - Versioned migration steps
  - Automatic backup creation in `/backups/<ISO8601>/`
  - Read-only mode for newer schemas
  - Migration validation and rollback safety

### Phase 2: Git & Versioning Layer

- [x] **Git Integration** - Hybrid CLI/go-git implementation (ADR-008, ADR-010)
  - System git for porcelain operations (push, pull, credentials)
  - go-git for fast read operations (log, diff, tree walking)
  - Git repository initialization and LFS setup
  - Remote repository configuration and sync
  - Wails service integration with project-aware operations
  - Tag listing via go-git; tag creation via CLI
  - Basic CLI diff command with `--summary-only` and `--json` flags

- [ ] **Snapshot System** - Commit + immutable tag pairs
  - [x] Create snapshots (commit + tag)
  - [x] Snapshot metadata file (`.archon/snapshots/<name>.json`)
  - [x] List/get snapshots and restore by name
  - [ ] Linear history presentation in UI
  - [ ] Snapshot comparison UI and diff integration

- [x] **Semantic Diff Engine** - Rename/move/property change detection (ADR-003)
  - [x] Structured change detection (not text-based)
  - [x] Move detection via parent ID changes
  - [x] Property and structural change identification
  - [x] Diff serialization for storage and UI
  - [x] CLI filters with `--only` flag for change type filtering
  - [x] Deterministic ordering of changes and output

- [x] **3-Way Merge System** - Conflict resolution with sibling ordering (ADR-003)
  - [x] Field-level merge resolution and conflict detection
  - [x] Non-conflicting change application (rename, move, property, order)
  - [x] CLI merge command with dry-run, JSON output, and verbose reporting
  - [x] Comprehensive test coverage for conflict vs non-conflict scenarios
  - [ ] Interactive conflict resolution
  - [ ] Merge strategy configuration

### Phase 3: Content & Plugin System

- [x] **Content-Addressed Attachments** - LFS integration for ≥1MB files (ADR-002)
  - [x] Content hashing and deduplication using SHA-256
  - [x] Git LFS integration for large files (≥1MB configurable threshold)
  - [x] Attachment reference validation for node properties
  - [x] Attachment management API (store, retrieve, delete, verify)
  - [x] CLI commands for attachment operations (add, list, get, remove, verify)
  - [x] Path sharding for efficient file organization
  - [x] Comprehensive test coverage for attachment system
  - [x] Garbage collection for unused attachments with dry-run support

- [ ] **Import Plugin System** - Sandboxed JS/TS workers (ADR-004)
  - Web Worker sandbox environment
  - Plugin API definition and validation
  - No filesystem/network access without consent
  - Plugin lifecycle management

- [ ] **CSV Import Plugin** - Example plugin implementation
  - Demonstrate plugin architecture
  - CSV parsing and validation
  - Preview and mapping interface
  - Error handling and rollback

### Phase 4: User Interface

- [ ] **UI Components** - Tree navigation, property editors, diff viewers
  - Virtualized tree component for large hierarchies
  - Property editor with type-specific inputs
  - Diff viewer with semantic change display
  - Context menus and keyboard shortcuts

- [ ] **Project Dashboard** - Recent snapshots, sync status, quick actions
  - Project overview and statistics
  - Recent activity and snapshot list
  - Git sync status and controls
  - Quick import/export actions

- [ ] **Hierarchy Workbench** - Multi-pane layout with tree navigation
  - Resizable three-pane layout
  - Tree navigation with search/filter
  - Property details panel
  - History sidebar with snapshot navigation

- [ ] **Diff & Merge UI** - Visual conflict resolution interface
  - Side-by-side diff view
  - Interactive merge conflict resolution
  - Undo/redo for merge decisions
  - Preview mode before applying changes

- [ ] **Import Wizard** - Plugin validation, preview, target selection
  - Plugin selection and validation
  - Data preview and mapping
  - Target location selection (new project vs. existing node)
  - Batch import with progress tracking

- [ ] **Settings UI** - Git remote config, LFS settings, theme selection
  - Git remote repository configuration
  - LFS threshold and storage settings
  - Theme and UI preferences
  - Plugin management and configuration

### Phase 5: Distribution & CLI

- [ ] **Build & Release** - MSI/DMG/AppImage with code signing (ADR-009)
  - Multi-platform build pipeline
  - Code signing for Windows (Authenticode) and macOS (Developer ID)
  - Linux AppImage with GPG signatures
  - Automated release with checksums

- [x] **CLI Interface** - Automation commands for power users
  - [x] Three-way merge command with conflict detection and resolution
  - [x] Semantic diff command with filtering and JSON output  
  - [x] Snapshot management commands (create, list, restore)
  - [x] Attachment management commands (add, list, get, remove, verify)
  - [x] Comprehensive test coverage for CLI commands
  - [ ] Project operations (create, open, validate)
  - [ ] Node manipulation commands  
  - [ ] Import/export automation
  - [ ] Additional scripting and batch operations

## 🎯 Current Focus

**Recommended Next Steps:** Content system complete, move to Import Plugin System

1. **Import Plugin System** - Implement sandboxed JS/TS workers for extensible data import
2. Consider Snapshot System UI integration - Linear history presentation and snapshot comparison UI  
3. Build & Release pipeline for distributing the CLI tools

## 📊 Progress Summary

- **Foundation Layer**: ✅ Complete (6/6 major components)
- **Data Extensions**: ✅ Complete (3/3 components)
- **Git & Versioning**: ✅ 3/4 components (Git Integration, Semantic Diff, 3-Way Merge core functionality complete)
- **Content & Plugins**: ✅ 1/3 components (Content-Addressed Attachments complete)
- **User Interface**: ⏳ 0/6 components
- **Distribution**: ⏳ 1/2 components (CLI Interface core functionality complete)

**Overall Progress**: ~75% complete (foundation + data + git/versioning + content + CLI core features complete)

---

*Last Updated: 2025-08-24*
*Next Review: After Phase 1 completion*
