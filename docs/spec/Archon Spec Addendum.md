# Addendum: Canonical Data Model, Storage Layout, Diff/Merge, Plugins, Search, and Ops (v1)

This addendum finalizes several “how” decisions for Archon v1, aligning with the original spec’s JSON-based model, Git-backed snapshots, and import plugin architecture while tightening scope for a reliable MVP.  ￼  ￼

A. Canonical Data Model & Constraints

A1. Identity & addressing (final):

- Stable UUID v7 per node (lexicographically sortable). Optional human-readable prefix is OK (e.g., node_01J9…). IDs are immutable across moves/renames to preserve diff/merge integrity and meaningful “move” detection.
- Sibling-level unique names (case-insensitive). Global uniqueness not required. Child order is meaningful and must be preserved (optical path, pipelines, etc.).
- Strict tree for MVP (no DAG/cross-links); evaluate DAG in v2 after user feedback.
- Rationale: Matches the original spec’s hierarchical JSON model while ensuring Git merges and semantic diffs remain intelligible to non-Git users.  ￼

A2. Properties & typing (final):

- Keys are free-form UTF-8; reserve _ prefix for Archon metadata.
- Values support basic types: string | number | boolean | null, with optional "typeHint" on a property (e.g., "date", "date-time", "units:V").
- Design leaves room for future namespacing (e.g., vendor.serialNumber) and optional “schema packs” without breaking existing data.
- Rationale: The spec emphasizes flexible key–value properties and postpones rigid schemas to keep the tool domain-agnostic.  ￼

A3. Attachments (final):

- Store binaries via Git LFS or content-addressed under /attachments/.
- Nodes reference attachments in properties, e.g.:

```json
"manualPDF": {
  "type": "attachment",
  "hash": "sha256:…",
  "filename": "manual.pdf"
}
```

- Default threshold: LFS for files > 1 MB; smaller allowed inline in /attachments/ (still referenced by hash).
- Rationale: Avoids ballooning Git history while preserving referential integrity.

A4. Timestamps/actors (final):

- Rely on Git commit metadata for “who/when changed,” not per-field change logs.
- Rationale: Reinforces the snapshot mental model already articulated in the spec.  ￼

⸻

B. On-Disk Layout & Performance

B1. Sharded repo layout (final):

```text
/project.json                 # project metadata { rootId, schemaVersion, settings }
/nodes/<id>.json              # one file per node (node schema, children as ID refs)
/index/archon.db              # SQLite index (local cache, rebuildable)
/attachments/<hash>.<ext>     # binaries (Git LFS for >1MB)
```

- Children are stored as ID references (not embedded subtrees).
- This supersedes the “single JSON file” example in the spec; the spec already allows “a set of JSON files,” and we standardize on that for v1 to reduce merge contention and improve diffs.  ￼
- Rationale: Minimizes conflicts when multiple users edit different branches of the tree; supports large projects gracefully.

B2. File size/scale targets (v1):

- 10k nodes smooth; 50k nodes acceptable with virtualization and index (see Section E).
- Depth is unconstrained, but warn on >64 levels.

⸻

C. Versioning, Diff & Merge Policy

C1. Authoritative diff (final):

- Semantic diff is primary; textual diff is auxiliary for power users. The UI surfaces human-readable changes:
- Property changes: Camera → Resolution: 5MP → 10MP
- Structural: Moved "Cooling System" under "Microscope"
- Rationale: Exactly the direction called out in the spec for structured/semantic diffs.  ￼

C2. Move semantics:

- Moves detected by comparing parent IDs across snapshots for the same node ID. Present as “Move,” not delete+add.
- Rationale: Gives domain experts accurate narratives of change.  ￼

C3. Three-way merge rules (field-level):

- Auto-merge for non-overlapping edits.
- Conflict when both branches modify the same property key on the same node from a shared base.
- Delete vs Edit/Move on same node ⇒ conflict.
- Sibling order: keep stable order; conflicting reorders prompt a reorder widget.
- UI: side-by-side per-field chooser; record decisions in the merge commit.
- Rationale: Predictable Git-backed collaboration while shielding non-Git users with a clear UX.

The spec already orients users to snapshots/tags for simplicity.  ￼

C4. Snapshots:

- Create immutable Git tags with unique names (UI term “Snapshot” == commit/tag pair).
- Optional snapshot manifest for descriptions/labels; never rewrites history.
- Rationale: Aligns with the spec’s “Snapshot (commit/tag)” abstraction.  ￼

⸻

D. Plugin System Scope & Sandbox (Import MVP)

D1. Execution model (final):

- JavaScript/TypeScript plugins only for MVP, running in a sandboxed frontend worker.
- No filesystem access; host injects file bytes chosen by the user.
- No network by default. If a plugin needs network, present a per-run consent dialog.
- Rationale: Mirrors the spec’s preference for front-end scripting for portability and safety.  ￼  ￼

D2. Plugin API (final):

- File: frontend/src/plugins/types.ts

```ts
export type ImportResult = { root: ArchonNode };

export type ArchonNode = {
  id: string; name: string; description?: string;
  properties: Record<string, string | number | boolean | null>;
  children: string[]; // child IDs (not embedded)
};

export interface ImportPlugin {
  meta: { id: string; name: string; version: string; formats: string[] };
  run(input: Uint8Array | string, options?: Record<string, unknown>): Promise<ImportResult>;
}
```

- Distribution (MVP): local directory ~/.archon/plugins. Signing/registry can wait until there’s a plugin ecosystem.
- Rationale: Matches the spec’s extensibility goals while minimizing platform complexity.  ￼  ￼

D3. Import flow:

- Validate → preview tree → user selects merge target (new project or under a node) → commit to a draft state → Snapshot.
- Optional: run imports in an ephemeral branch to allow review before merge, as the spec suggests.  ￼

⸻

E. Search & Indexing (Local)

E1. SQLite index (final):

- File: .archon/index/archon.db (rebuildable cache; not required for correctness).
- Tables:
- nodes(id TEXT PRIMARY KEY, name TEXT, parent_id TEXT)
- properties(node_id TEXT, key TEXT, type TEXT, value_text TEXT, value_num REAL, value_bool INTEGER, value_date TEXT)
- fts_properties (FTS5) on name, key, value_text
- Updates: Incremental on node/property changes; full rebuild on clone/import.
- Rationale: Keeps search instant at 10k+ nodes without external deps.

E2. Virtualized UI:

- Virtualize any list/tree level > 200 rows; lazy-load children on expand; prefetch around viewport.

⸻

F. Git Workflow & Credentials

F1. UI workflow:

- Linear history in
 the app with background merges as needed. Advanced users can branch externally; Archon honors it but keeps the UI simple (“Project/History/Snapshot/Sync”), consistent with the spec.  ￼

F2. Credentials:

- Prefer system Git credential helpers; store HTTPS tokens in OS keychain.
- SSH: generate ed25519 keypair; assist user in adding keys to GitHub/GitLab.

⸻

G. Error Handling & Recovery

- Standard error envelope: { code, message, details? }.
- Logging: rotating files (e.g., 10MB × 5).
- Crash safety: autosave working set and offer to reopen on restart.
- Import/Export: progress with cancel; malformed input yields a validation report and safe abort.

⸻

H. Product & UX Scope for MVP

- Templates/schemas: defer; users can copy subtrees to roll their own templates.
- Bulk ops: basic multi-select delete; CSV round-trip later.
- Approvals/audit workflows: rely on Git history for MVP.
- Internationalization: UTF-8 now; English-first UI; make strings localizable later.
- Rationale: Focuses on core value: trustworthy hierarchical editing, snapshots, and import.  ￼

⸻

I. Distribution, Updates, Privacy

- OS support: Windows 10+, macOS 12+, Ubuntu 20.04+.
- Updates: Start with manual “Check for Updates.” Auto-update, code-signing pipelines, and notarization can follow once usage warrants.
- Telemetry: Opt-in crash reporting only; no usage analytics for MVP.

⸻

J. Implementation Phasing (pragmatic)

1. Phase 1: Core data model, sharded storage (/nodes/*.json), basic Git (init/commit/tag/history/clone), tree UI.
2. Phase 2: Semantic diff/merge (move detection by ID), property editor polish, snapshot UX.
3. Phase 3: Plugin system (JS, sandboxed), import flows, SQLite search.
4. Phase 4: Performance polish, virtualization tuning, cross-platform CI, signing/notarization.

⸻

K. Key File/Module Map (for the team)

- Schemas & project files
- project.json — rootId, schemaVersion, settings
- `/nodes/<id>.json` — single node record with children: string[]
- Search
- .archon/index/archon.db — SQLite cache (rebuildable)
- Plugins
- frontend/src/plugins/types.ts — interfaces (see D2)
- ~/.archon/plugins — local plugin directory (MVP)
- Git & storage (Go)
- internal/git/service.go — init, status, commit/tag, diff, merge, pull/push, LFS setup
- internal/store/loader.go — load/save nodes, ID allocation (UUID v7), ordering, parents
- internal/index/sqlite/ — indexer, incremental updates
- Diff/Merge
- internal/merge/semantic.go — property/structure diff (move detection by parent change)
- internal/merge/three_way.go — field-level three-way merges & conflicts
- Frontend
- src/lib/diff/DiffViewer.svelte, src/lib/diff/MergePanel.svelte — semantic diff & conflict UI
- src/lib/ui/VirtualList.svelte — virtualization
- src/routes/import/ImportWizard.svelte — validate/preview/merge flow

⸻

Compatibility Notes with the Original Spec

- JSON as canonical store with flexible key–value properties is preserved; we standardize on sharded files vs one monolith for collaboration/scale.  ￼
- Git-backed snapshots, history, and user-friendly terminology (Snapshot/History/Sync) remain the grounding model.  ￼
- Plugin imports stay JS-centric and sandboxed as the spec proposes; Go plugins can be revisited later.  ￼  ￼
- Semantic diff is elevated to the primary UX (text diff stays available), directly reflecting the spec’s intent.  ￼

⸻

If you want, I can also convert each decision block (A–K) into concise ADRs (one file per decision) so the team can PR/comment line-by-line.
