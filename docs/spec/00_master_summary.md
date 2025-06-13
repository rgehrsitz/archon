# Archon – Master Summary (v1.0)

> **Last updated:** <!-- Fill when editing -->

## 1  Purpose & Elevator Pitch

**Archon** is a guardian‑grade configuration management platform for physical or logical lab equipment, museum exhibits, IoT deployments, field sites, test benches—where drift causes real‑world downtime and compliance headaches.  Think “Git for hardware/software configs,” bundling a tamper‑evident history of every component, property, firmware blob, and wiring change into a single, queryable timeline.

## 2  Core Value Proposition

- **Authoritative history** – every change is snapshot‑versioned, diffable, and traceable to a signed actor.
- **Schema‑flexible JSON model** – easily extends to new asset types without DB migrations.
- **Local‑first, cloud‑optional** – works offline in secure labs, syncs later if desired.
- **Plugin runtime** – WASM sandbox lets domain experts script importers, visualisations, or compliance checks without touching core code.
- **Audit‑ready exports** – one command emits a cryptographically signed PDF or SBOM for regulators.

## 3  MVP Scope Snapshot

| Phase | Deliverable (Definition of Done) |
|-------|-----------------------------------|
| **0 Bootstrap** | Repo skeleton, CI, Windsurf Planning Mode, rules & memories seeded. |
| **1 Core Storage** | Canonical JSON schemas; single‑file load/save via *ConfigVault*; ≥90 % test coverage. |
| **2 Snapshot + Git** | Git‑backed snapshot service, semantic diff engine, CLI `archon snapshot`, golden‑file tests. |
| **3 Plugin Runtime** | WASM sandbox, `PluginContext` API, plugin manager UI, hello‑world plugin. |
| **4 UI MVP** | Svelte tree viewer, property editor, basic a11y/i18n; end‑to‑end demo. |
| **5 Enhancements** | Cloud sync, mobile PWA, LLM helpers (stretch goals). |

## 4  High‑Level Architecture

```
┌───────────────┐
│   Svelte UI   │  ← Tailwind‑styled SPA served by Wails
└──────┬────────┘
       │ REST / WebSocket
┌──────▼────────┐
│  API Gateway  │  ← Go (Wails backend)
└──────┬────────┘
       │ Go interfaces
┌──────▼────────┐      ┌─────────────────┐
│  Core Engine  ├──────►  Plugin Sandbox │ (WASM)
└──────┬────────┘      └─────────────────┘
       │ JSON files
┌──────▼────────┐
│ ConfigVault   │  ← canonical storage + Git repo
└───────────────┘
```

## 5  Key Tech Stack

| Layer | Tech | Notes |
|-------|------|-------|
| UI | **Svelte 5.34.1**, **Tailwind 4.1.8**, **Vite 6.3.5** | Compact, component‑first SPA. |
| Desktop shell | **Wails 2.10** | Cross‑platform native host. |
| Backend | **Go 1.22** | Deterministic logic, AOT‑friendly. |
| Storage | Local filesystem + **Git** | Each project = repo; future cloud sync. |
| Tests | Go `testing`, Vitest, Playwright | ≥90 % coverage rule. |

## 6  Document Map

| Doc | One‑liner |
|-----|-----------|
| **Archon_ Comprehensive Product Specification.pdf** | Canonical, full‑length spec (40 pp). |
| `archon_mvp_scope.md` | MVP boundaries & phase priorities. |
| `archon_storage_file_layout.md` | Canonical JSON layout & key ordering. |
| `archon_security_trust_model.md` | Threat model, sandboxing, signing. |
| `archon_testing_debugging_infrastructure.md` | CI, coverage, golden‑file approach. |
| `archon_tooling_ci_cd_packaging.md` | Build matrix, packaging targets. |
| `archon_collaboration_multiuser.md` | Multi‑user workflow concepts. |
| `archon_optional_visionary.md` | Stretch goals: cloud, mobile, AI assistant. |
| `archon_i_18_n_a_11_y.md` | i18n & accessibility guidelines. |

## 7  Key Terminology

| Term | Meaning |
|------|---------|
| **Asset** | A top‑level physical thing: microscope, server rack, drone. |
| **Component** | Sub‑part of an asset (board, sensor, cable). |
| **Snapshot** | Immutable capture of full hierarchy at a point in time. |
| **Plugin** | WASM module adding import, analysis or UI. |
| **ConfigVault** | Storage layer providing canonical JSON & Git history. |

## 8  Non‑Goals for MVP

- No multi‑tenant SaaS portal.
- No live device polling; Archon is *pre‑deployment* / *audit* focused.
- Mobile apps, cloud sync, on‑prem federation = Phase 5+.
- No binary diffing beyond simple checksum validation (future work).

## 9  References & Next Steps

1. Read `archon_mvp_scope.md` for detailed phase definitions.
2. Phase 0 bootstrap tasks live in `.windsurf/plan.md` once drafted.
3. All pull requests must cite relevant spec anchors.

> *“Archon guards the configuration so your hardware can guard the mission.”*

