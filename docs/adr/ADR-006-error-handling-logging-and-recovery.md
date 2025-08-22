# ADR-006: Error Handling, Logging, and Recovery

Status: Accepted
Date: 2025-08-21
Owners: Archon Core

## Context

We need predictable errors across the Go↔UI bridge, useful logs for support, and resilience to crashes during long operations (clone, import, merge).

## Decision

- **Error envelope:** `{ code: string, message: string, details?: any }` across all boundaries.
  - Namespaces: `E_GIT_*`, `E_IO_*`, `E_SCHEMA_*`, `E_PLUGIN_*`, `E_INDEX_*`, `E_AUTH_*`.
- **User-facing messages** avoid stack traces; “More details” reveals technical info.
- **Logging:** `zerolog` with rotating files (10MB × 5) at `logs/`.
- **Crash safety:** Periodic autosave of dirty nodes; on restart, offer recovery of last working state. Merge/clone/import use temp dirs and atomically swap on success.
- **Telemetry:** Opt-in crash reports only; no usage analytics (MVP).

## Rationale

A small, consistent error surface reduces guesswork. Rotating logs and autosave are standard reliability guardrails.

## Alternatives Considered

- **Ad-hoc error types:** Leads to inconsistent handling and poor UX.
- **Always-on telemetry:** Privacy concerns; not needed to launch.

## Consequences

- Positive: Debuggable failures; recoverable sessions; consistent UX.
- Negative: Some errors need careful mapping to user-friendly messages.
- Follow-ups: Error catalog in docs; UI affordances for retry.

## Implementation Notes

- Go helpers: `internal/errors/errors.go` (codes & wrapping), `internal/logging/logging.go`.
- Frontend mapper: `src/lib/errors/index.ts` for humanized strings + actionable hints.
- Autosave: debounced (≈2s) and on app blur/close; files written under a `.archon/tmp/` session folder with journal.
- Long ops: context cancellation; progress events; rollback on failure.

## Review / Revisit

- Revisit when enabling background services or auto-update (more failure modes).
