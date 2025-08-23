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

- [ ] **Snapshot System** - Commit + immutable tag pairs
  - Create snapshots (commit + tag)
  - Snapshot metadata and notes
  - Linear history presentation in UI
  - Snapshot comparison and restoration

- [ ] **Semantic Diff Engine** - Rename/move/property change detection (ADR-003)
  - Structured change detection (not text-based)
  - Move detection via parent ID changes
  - Property and structural change identification
  - Diff serialization for storage and UI

- [ ] **3-Way Merge System** - Conflict resolution with sibling ordering (ADR-003)
  - Field-level merge resolution
  - Sibling order conflict handling
  - Interactive conflict resolution
  - Merge strategy configuration

### Phase 3: Content & Plugin System

- [ ] **Content-Addressed Attachments** - LFS integration for ≥1MB files (ADR-002)
  - Content hashing and deduplication
  - Git LFS integration for large files
  - Attachment reference validation
  - Garbage collection for unused attachments

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

- [ ] **CLI Interface** - Automation commands for power users
  - Project operations (create, open, validate)
  - Node manipulation commands
  - Import/export automation
  - Scripting and batch operations

## 🎯 Current Focus

**Recommended Next Steps:** Phase 2 (Git & Versioning Layer)

1. Git Integration - Hybrid CLI/go-git implementation for version control
2. Snapshot System - Commit + immutable tag pairs for user-friendly versioning

## 📊 Progress Summary

- **Foundation Layer**: ✅ Complete (6/6 major components)
- **Data Extensions**: ✅ Complete (3/3 components)
- **Git & Versioning**: ⏳ 1/4 components (Git Integration complete)
- **Content & Plugins**: ⏳ 0/3 components
- **User Interface**: ⏳ 0/6 components
- **Distribution**: ⏳ 0/2 components

**Overall Progress**: ~50% complete (foundation + data + git integration complete)

---

*Last Updated: 2025-08-23*
*Next Review: After Phase 1 completion*
