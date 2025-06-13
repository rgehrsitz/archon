1  Project identity

Project Name: Archon

Primary Purpose: Authoritative configuration‑management system for physical assets ("Config‐as‑code for the real world").

Spec Version (authoritative docs): v1.0

2  Core tech stack (locked for Phase 1‑4)

Backend: Go 1.22 + Wails 2.10 (embedded in /src/backend)

Frontend: Svelte 5.34.1, TailwindCSS 4.1.8, Vite 6.3.5 (rooted in /src/ui)

Build commands:

wails dev → fullstack dev server

wails build -platform windows/amd64 → production bundle

3  Repository conventions

Directory split: /src/backend (Go) ⇄ /src/ui (Svelte)

Config files: wails.json at repo root; Tailwind config in /src/ui/tailwind.config.js.

Task IDs: monotonic decimal (0001_…).

No file outside the active task may be modified.

4  Quality & CI policy

Per‑file test coverage target: ≥ 90 % lines for every file touched.

Fast‑lane CI: lint + unit tests + rules/plan checks.

Slow‑lane CI: integration + govulncheck + npm audit + docker packaging.

5  Key spec anchors (docs/spec/)

component-schema

serializer-order

snapshot-flow

6  Security & tooling

GitHub MCP plugin: PAT scopes repo, pull_requests only.

No secrets committed to git; signing keys live in manual release workflow.

7  Planning & rules reminders

Planning Mode: always enabled; .windsurf/plan.md is the canonical roadmap.

Rules files: .windsurf/rules/*.md enforce style, coverage, and lint requirements.

Glue‑task policy: allowed for non‑API refactors; must carry glue-task label.