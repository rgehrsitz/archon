# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Archon is a desktop knowledge workbench for hierarchical projects with first-class snapshots, semantic diff/merge, and Git-backed sync. It's built with Wails v2 (Go backend) and Svelte 5 frontend with Tailwind CSS.

## Development Commands

### Primary Development
- `wails dev` - Start development server with hot reload (Go backend + Svelte frontend)
- `wails build` - Build production distributable package

### Frontend Only
- `cd frontend && npm run dev` - Frontend development server only
- `cd frontend && npm run build` - Build frontend
- `cd frontend && npm run check` - Run Svelte type checking

## Architecture Overview

### Core Data Model
- **Identity**: Each node has an immutable UUIDv7 `id` with sibling-unique names
- **Storage**: Sharded JSON files (`/nodes/<id>.json`) + SQLite index (`/.archon/index/archon.db`)
- **History**: Git-backed snapshots with semantic diff/merge as primary UX
- **Hierarchy**: Strict tree structure (no DAG in v1), meaningful child ordering

### Backend Structure (`internal/`)
- `store/` - Node storage and project management
- `index/sqlite/` - SQLite indexing for fast search
- `git/` - Hybrid Git implementation (system git + go-git)
- `merge/` - Semantic merge operations
- `diff/` - Semantic diff engine
- `plugins/` - Plugin system backend with host services
- `snapshot/` - Git-backed snapshot management
- `api/` - Wails service layer
- `types/` - Core data models

### Frontend Structure (`frontend/src/`)
- `lib/api/` - Go service wrappers
- `lib/components/ui/` - bits-ui based components (50+ component categories, 250+ components)
- `lib/plugins/` - Plugin system runtime with sandboxed execution
- Built with Svelte 5 runes, Tailwind 4 CSS-first config

### Key Technical Details
- Uses UUIDv7 for time-sortable identifiers (`internal/id/uuid.go`)
- Content-addressed attachments via Git LFS
- Rebuildable SQLite index for performance
- Plugin system with sandboxed JS/TS workers
- Error envelopes and rotating logs for reliability

## Prerequisites
- Go 1.23+
- Node.js 18+
- Wails CLI

## Important Notes
- The project uses a hybrid Git approach: system git for porcelain/LFS/credentials, go-git for fast reads
- Schema versioning with forward migration (see `docs/adr/ADR-007-data-migration-and-schema-versioning.md`)
- Child order is meaningful and preserved in storage
- Names must be unique among siblings only (not globally)