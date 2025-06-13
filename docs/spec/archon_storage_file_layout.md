## Archon: Storage & File Layout Design

Clear, consistent file and directory conventions are critical for portability, maintainability, and user trust. This section specifies:

1. Project directory structure and file naming.
2. Single-file vs. multi-file JSON organization, with canonical serialization emphasis.
3. Attachments storage and Git LFS strategy, including onboarding checks.
4. Handling of temporary states, undo, and autosave, with clear recovery UI wording.
5. Cross-platform path considerations and migration strategy (MVP-focused).
6. Implementation guidance and testing.

---

### 1. Project Directory Structure

Each Archon project corresponds to a Git repository directory. A clear, opinionated layout helps users understand where data lives and facilitates tooling (diff, backup, import/export).

#### 1.1 Root Layout

```
<project-root>/
├── archon.json                   # Project metadata and settings (e.g., schema_version, storage_mode)
├── components/                   # (Optional) Multi-file component definitions
│   ├── <component-id>.json       # Individual component files
│   └── ...
├── components.json               # (Alternative) Single-file components hierarchy (flat map + root reference)
├── attachments/                  # Storage for user attachments
│   ├── <attachment-id>_<filename> # Actual files tracked via Git LFS
│   └── ...
├── .archon/                      # Internal data (ignored from user-facing views)
│   ├── autosave/                 # Autosave snapshots or draft states
│   ├── temp/                     # Temporary working files during operations
│   └── plugin_data/              # Plugin-specific storage if needed
├── .gitattributes                # Ensure attachments tracked via LFS when enabled
├── .git/                         # Git repository folder
└── README.md                     # Optional project-specific README or instructions (e.g., LFS onboarding)
```

- **archon.json**: Central project metadata (e.g., `schema_version`, `storage_mode`, auto-snapshot config, plugin settings). Versioned in Git.
- **components.json**: Single-file representation of full hierarchy using flat map + root reference. Preferred for small projects.
- **components/**: Directory containing one JSON file per component, keyed by `component.id`. Used in multi-file mode for scalability.
- **attachments/**: Stores attachments under `<attachment-id>_<original-filename>`. Files tracked via Git LFS when available.
- **.archon/**: Internal state not for direct user editing:
  - **autosave/**: Temporary snapshots/drafts.
  - **temp/**: Working files during operations (import, merge resolution).
  - **plugin\_data/**: Plugin-specific caches or credentials, cleaned responsibly.
- **.gitattributes**: Configures LFS tracking for attachments; generated or updated when LFS is enabled.
- **README.md**: Should include LFS onboarding instructions if attachments are used.

Key Recommendation: Default to single-file mode for new/small projects. Offer a one-way migration tool (MVP) to multi-file mode when project grows. Deferring multi→single migration reduces complexity.

---

### 2. JSON Organization: Single vs Multi-File

#### 2.1 Canonical Serialization

- **Critical**: Enforce canonical JSON formatting before commits to minimize diff noise and merge conflicts:
  - Pretty-print with sorted keys at all levels.
  - In single-file mode, a stable key ordering: e.g., `root_id`, then sorted `components` keys; within each component: fixed field order (`id`, `name`, `type`, `template`, `properties`, `children`, `attachments`, `tags`, `notes`, audit fields).
  - In multi-file mode, each component file serialized similarly with sorted keys.
  - Implement formatting step before staging changes for snapshot or merge operations.

#### 2.2 Single-File Mode

- **Structure**: Flat map + root reference:
  ```jsonc
  {
    "root_id": "root-uuid",
    "components": {
      "root-uuid": { "id":"root-uuid", "name":"Root", "children":["child1","child2"], /* ... */ },
      "child1": { /* ... */ },
      /* ... */
    }
  }
  ```
- **Pros**: Simple load/save; easy indexing; suitable for small projects.
- **Cons**: Larger diffs and potential merge conflicts in large teams.
- **Usage**: Default on new projects. Monitor component count; when above threshold (e.g., 500), suggest migrating to multi-file.

#### 2.3 Multi-File Mode

- **Structure**:
  - `archon.json` contains `root_component_id`.
  - `components/<id>.json` for each component:
    ```jsonc
    {
      "id": "<id>",
      "name": "...",
      "type": "...",
      "template": "...",
      "properties": { /* ... */ },
      "children": ["child-id-1", ...],
      "attachments": [ /* metadata items */ ],
      "tags": [...],
      "notes": "...",
      "created_at": "...",
      "created_by": "...",
      "updated_at": "...",
      "updated_by": "..."
    }
    ```
- **Pros**: Granular edits reduce merge conflicts; scalable for large hierarchies; easier for plugins to target specific files.
- **Cons**: Many files; slightly more complex loading logic.
- **Load/Save**: On update, write only affected files; on load, read recursively from `root_component_id`.

#### 2.4 Mode Migration (MVP Focus)

- **Single→Multi**: Provide a migration tool that:
  1. Reads `components.json` flat map.
  2. Creates `components/` directory and writes each component to `<id>.json` with canonical serialization.
  3. Updates `archon.json` with `storage_mode: "multi-file"` and `root_component_id`.
  4. Stage all new files and delete `components.json` in one commit to avoid inconsistent states.
- **Multi→Single**: Defer to future release; not required for MVP.

---

### 3. Attachments Storage & Git LFS Strategy

Attachments can be large and must not bloat repository. Use Git LFS with guided onboarding.

#### 3.1 Directory & Naming

- **attachments/**: Store files as `<attachment-id>_<sanitized-original-filename>`.
- **Metadata in Component**:
  ```jsonc
  "attachments": [
    {
      "id": "<uuid>",
      "filename": "original-name.pdf",
      "mime_type": "application/pdf",
      "size_bytes": 123456,
      "uploaded_at": "2025-06-04T12:35:10Z",
      "source": "user_upload",
      "description": "Warranty document"
    }
  ]
  ```

#### 3.2 Git LFS Onboarding & Configuration

- **On Project Open/Create**:
  - Check if Git LFS is installed.
  - If not found and attachments feature is invoked: show guided dialog explaining why LFS is needed, with a link to Git LFS download/install instructions, and an option to defer but warn about large-file issues.
- **.gitattributes**: Automatically configure when first attachment is added or user enables LFS:
  ```gitattributes
  attachments/** filter=lfs diff=lfs merge=lfs -text
  ```
- **Size Threshold**: Optionally only track files >1MB via automatic detection; small files can remain in Git normal storage to reduce LFS overhead.
- **Operations**:
  - On add: copy file into `attachments/`, LFS-track if needed, commit with metadata update.
  - On remove: remove metadata, prompt user whether to delete physical file; unreferenced files can be detected and offered for cleanup.
- **Clone Behavior**: When opening a cloned repo, detect missing LFS objects; prompt user to install/initialize LFS or skip but warn that attachments may not load.

#### 3.3 Attachment Lifecycle and Garbage Collection

- **Add/Update**: For rename, retain same `id`, update `filename` metadata, and rename file accordingly.
- **Remove**: Remove metadata reference; optionally delete file. Provide a “Clean Unreferenced Attachments” UI that scans `attachments/` for files not referenced in any component and offers deletion.
- **Versioning**: Git/LFS retains history; users can retrieve older versions if needed.

---

### 4. Temporary States, Undo, and Autosave

To prevent data loss and support user mistakes, Archon uses multi-layered temporary storage under `.archon/`.

#### 4.1 Autosave Mechanism & Recovery UI

- **Trigger Points**: Periodic intervals or after significant edits if auto-save enabled.
- **Storage Location**: `.archon/autosave/<timestamp>/` containing:
  - Single-file mode: copy of `components.json`.
  - Multi-file mode: snapshots of modified component files.
- **Retention Policy**: Controlled via `archon.json` settings (e.g., keep last N or time-based).
- **Recovery on Open**: On startup, compare last commit timestamp vs latest autosave. If an autosave is newer, show clear dialog:
  > “Archon detected an unsaved draft from [TIME]. Your last snapshot was from [DATE\_TIME]. Would you like to: [Load the Draft] or [Discard the Draft]?”
- **User Action**:
  - **Load the Draft**: Restore files from autosave to working state (not committed) for user review.
  - **Discard the Draft**: Delete autosave files and continue with last committed state.
- **UI Clarity**: Include timestamps and context (e.g., “You added components after last snapshot; recover those changes?”).

#### 4.2 Temporary Working Files

- **Operation Isolation**: During imports, merges, or plugin operations, write intermediary data to `.archon/temp/`.
- **Atomic Writes**: On success, replace main files atomically (e.g., write to temp then rename), preventing partial writes on failures.
- **Cleanup**: After operation completion or on startup detecting stale temp data, prompt user to clean or recover if appropriate.

#### 4.3 Undo Before Snapshot

- **UI-Level Undo Stack**: In-memory undo for property edits and simple operations, reset on snapshot or app restart. Provides quick reversal of mistakes before committing.

#### 4.4 Save vs Snapshot

- **Save**: Write in-memory changes to disk (components.json or individual files). Does not commit to Git.
- **Snapshot**: Save + Git commit + tag. Distinct action for durable history.
- **Dirty Indicators**: UI flags unsaved changes; prompt save or snapshot before closing or major operations.

---

### 5. Cross-Platform Path & Permissions Considerations

#### 5.1 Path Handling

- Use Go’s `filepath` utilities for OS-independent paths.
- **Config Directories**: Use OS conventions for global settings (`%APPDATA%`, `~/.config/Archon`, etc.).
- **Filename Sanitization**: When importing attachments with arbitrary names, sanitize invalid characters per OS but preserve original name in metadata.

#### 5.2 Git LFS on Platforms

- **LFS Detection**: Detect presence of Git LFS; if absent, guide user through installation.
- **Commands**: Use appropriate Git CLI or Go library with LFS support; handle platform-specific invocation.
- **Permissions**: Handle file permission issues (e.g., read-only files) gracefully with clear error messages.

---

### 6. Implementation Guidance & Testing

#### 6.1 Storage Abstraction

- **Storage Interface**:
  ```go
  type Storage interface {
    LoadAllComponents() ([]Component, error)
    SaveComponent(comp Component) error
    DeleteComponent(id string) error
    SaveRootReference(rootID string) error
    // Autosave and temp operations
    SaveAutosave(timestamp string, data AutosaveData) error
    ListAutosaves() ([]AutosaveInfo, error)
    LoadAutosave(timestamp string) (AutosaveData, error)
    CleanupTemp() error
  }
  ```
- **Implementations**:
  - **SingleFileStorage**: Reads/writes `components.json` with canonical serialization.
  - **MultiFileStorage**: Manages `components/<id>.json` files and `archon.json` root reference.
- **Migration Tool**: One-way Single→Multi as MVP: ensure atomic commit of file split.

#### 6.2 Attachment Manager

- **Add/Remove/Rename**: Handle file operations, metadata updates, and Git/LFS tracking.
- **LFS Onboarding**: On first attachment, detect LFS; if missing, show guided installation dialog with link and instructions.
- **Cleanup UI**: Scan for unreferenced files; present list for deletion.

#### 6.3 Autosave & Temp Files

- **Scheduler**: In-memory scheduler triggers autosave; write to `.archon/autosave/<timestamp>/`.
- **Recovery Flow**: On startup, detect newer autosaves; show recovery dialog with clear timestamps/context.
- **Atomic Operations**: Use temp directories and atomic rename on successful operations.

#### 6.4 Validation & Sanitization

- **Filename & Path Sanitization**: Prevent directory traversal; restrict operations within project root or permitted paths.
- **JSON Schema Validation**: Validate component structure before writing; if invalid, store in autosave until user resolves.
- **Canonical Serialization Enforcement**: Always format JSON before saving or committing.

#### 6.5 Testing

- **Unit Tests**:
  - Storage implementations: test read/write in both modes; test migration logic and canonical formatting.
  - Attachment Manager: simulate adding/removing files; LFS detection scenarios.
  - Autosave: simulate triggers, recovery, and retention policy.
- **Integration Tests**:
  - Large project: simulate many components; test Single→Multi migration; test concurrent edits in multi-file mode.
  - LFS absent/present: test guided onboarding flows.
- **Cross-Platform Tests**:
  - Verify path handling, filename sanitization, and LFS commands on Windows, macOS, Linux.
- **UI Tests**:
  - Verify recovery dialog, attachment dialogs, storage mode switch, autosave notifications.

---

### 7. Documentation & User Guidance

- **README Template**: Include instructions on storage modes, LFS onboarding, autosave behaviour, and migration.
- **LFS Onboarding Dialog**: Clear messaging on why LFS is needed, link to official download, simple steps.
- **Migration Guide**: Describe Single→Multi migration benefits and process; note Multi→Single deferred.
- **Autosave Recovery**: Document how recovery works and how to interpret dialog messages.
- **Best Practices**: Advice on storage mode choice based on project size, attachment naming, backup strategies.
- **Error Messages**: Provide actionable errors when file operations or LFS issues occur.

---

*End of Storage & File Layout Design Deep Dive (Refined with Onboarding & MVP Focus)*

