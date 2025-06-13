## Archon: Testing & Debugging Infrastructure

Robust testing and debugging infrastructure are vital for ensuring long-term stability, maintainability, and observability of Archon. This section covers:
1. Automated test harnesses (plugin execution, diff accuracy, storage behaviors).
2. Snapshot integrity checks (schema validation, optional checksum/tamper detection).
3. Developer tools (debug panel for plugin logs, tree diffs, Git state inspection, IPC monitor).
4. Error reporting strategy (logging levels, UI notifications, crash telemetry, sanitized issue bundles).
5. CI/CD integration and quality gates.
6. Implementation guidance and best practices.

---

### 1. Automated Test Harnesses

Automated tests guard against regressions and verify core functionality. Key areas:

#### 1.1 Plugin Execution Testing
- **Goal**: Validate that plugins load correctly, their lifecycle hooks run as expected, and they handle edge cases.
- **Mocked PluginContext**: A critical enabler. Provide a test harness supplying a mock `PluginContext` with in-memory or temporary filesystem-backed fake project data, permissions simulation, and IPC stubs. This isolation accelerates plugin development and ensures ecosystem quality.
- **Sample Plugins Suite**: Maintain representative test plugins for each category: import, export, sync, validation, visualization, automation.
- **Test Scenarios**:
  - **Initialization**: Ensure `init` hook runs without errors; context APIs return expected values.
  - **Normal Operation**: For import plugins, feed sample inputs (e.g., CSV content) and verify integration into fake project state.
  - **Error Conditions**: Simulate invalid input, permission denials, network timeouts for sync plugins; ensure plugins handle errors gracefully and report via context APIs.
  - **Async/Long-Running**: Simulate delays; verify cancellation tokens work and plugin aborts correctly.
- **Permission Enforcement**: In tests, mock denied operations (e.g., file or network) and verify plugins respect declared permissions.
- **Isolation**: Run plugin tests in isolated environments, using temporary directories to avoid side effects.
- **Implementation**:
  - **Backend Plugins (Go)**: Use Go’s `testing` package; inject fake contexts.
  - **Frontend Plugins (JS/TS)**: Use Jest/Mocha; mock Wails IPC and context methods.
- **Golden Files for Behavior**: For complex plugin outputs (e.g., generated component trees), compare against golden fixtures to catch regressions.
- **CI Integration**: Run plugin test suite on each commit; alert on failures.

#### 1.2 Diff Accuracy Testing
- **Goal**: Verify structural diff engine produces correct summaries and detailed changes across scenarios.
- **Golden Diff Reports**: Use predefined pairs of JSON hierarchies with known changes; store expected diff report objects as golden files. On changes to diff logic, compare outputs to these golden fixtures to detect regressions.
- **Synthetic Trees**: Create test cases covering:
  - Additions, removals, moves, property modifications, attachment and metadata updates.
  - Deep nesting and large-scale hierarchies.
  - Ordering differences: ensure canonical serialization avoids false positives.
- **Lazy-Loading Behavior**: Test summary-only diff returns correct counts; requesting detailed diff for a component yields correct sub-diff consistent with golden data.
- **Performance Benchmarks**: In tests, measure diff computation on large trees; verify progress events and acceptable timing.
- **Implementation**:
  - Write unit tests in Go for DiffService using in-memory structures or temporary files.
  - Use table-driven tests mapping input pairs to expected diff results loaded from golden JSON fixtures.

#### 1.3 Storage & File Layout Testing
- **Goal**: Ensure single-file and multi-file storage modes behave correctly, including mode migration, read/write, and autosave.
- **Storage Interface Tests**:
  - **SingleFileStorage**: Load initial small hierarchy, modify, save, reload; verify structure matches.
  - **MultiFileStorage**: Read/write individual component files, update children arrays, reload hierarchy correctly.
- **Mode Migration (Single→Multi)**: Simulate project with `components.json`, run migration tool, verify creation of correct component files, update `archon.json`, and that hierarchy remains identical.
- **Autosave Tests**: Simulate triggers; ensure `.archon/autosave/<timestamp>/` entries created; retention policy prunes old entries correctly.
- **Attachment Manager Tests**: Use temporary directories; simulate adding/removing/renaming attachments; mock absence of Git LFS to verify onboarding prompts logic; validate metadata updates and file operations.
- **Implementation**:
  - Go tests using temporary filesystem structures.
  - Golden fixtures for small sample hierarchies and expected file layouts.

#### 1.4 Snapshot & Git Operation Tests
- **Goal**: Validate snapshot creation, tagging, diff integration, pull/merge conflict detection, and resolution flows.
- **Temporary Git Repos**: Initialize ephemeral Git repos in tests, configure canonical serialization behavior.
- **Snapshot Lifecycle**:
  - Create snapshots with names/descriptions; verify tag creation, overwrite handling simulated via test inputs.
  - Auto-snapshot triggers: simulate operations leading to auto-snapshots; verify tags follow timestamp patterns and grouping metadata recorded.
- **Diff Integration**: After creating snapshots, perform diff calls; compare semantic diff output to golden expectations.
- **Merge & Conflict**: Simulate divergent edits on component JSON (single and multi-file modes); run `git pull` in test, detect conflict markers; test parsing logic yields correct conflict data structures for UI.
- **Implementation**:
  - Go tests, using os/exec or go-git for Git operations; simulate user decisions via injected callbacks/mocks.
  - Golden fixtures for expected conflict data structures.

#### 1.5 UI Component Tests
- **Goal**: Verify Svelte components for snapshot history, diff viewer, storage mode switch, autosave recovery dialog, plugin log panel, debug panel functions.
- **Unit Testing**: Svelte component tests using @testing-library/svelte; mock IPC responses to verify component behavior (e.g., grouping auto-snapshots, showing recovery prompts).
- **E2E Testing**: Playwright automation for core user journeys:
  - New project creation, manual snapshot, view history, diff expansions, resolve simulated conflicts via UI.
  - Autosave recovery: simulate stale autosave, verify recovery dialog.
  - Attachment workflow: simulate missing LFS, verify onboarding dialog appears.
  - Debug Panel: open plugin log viewer and IPC monitor, verify logs appear correctly.
- **Mock IPC**: For unit tests, mock backend endpoints; for E2E, use a test backend with controllable scenarios.
- **Implementation**:
  - CI runs headless browsers for E2E; include fixtures and test projects.

---

### 2. Snapshot Integrity Checks

Ensuring snapshot integrity avoids silent corruption or data drift. Two-tier strategy:

#### 2.1 JSON Schema Validation (Must-Have)
- **Goal**: Prevent invalid data entering history.
- **Pre-Snapshot**: Always run JSON Schema validation on in-memory or on-disk JSON before committing; block snapshot on violations with clear UI dialog explaining errors; allow override only after explicit user confirmation.
- **Post-Load**: When loading a snapshot state for viewing or rollback, validate JSON schema; if invalid, surface detailed errors to user.
- **Implementation**:
  - Use a Go JSON Schema validation library; embed schema versions in `archon.json`; maintain schema definitions under version control.

#### 2.2 Checksum & Tamper Detection (Optional Advanced)
- **Goal**: Provide optional high-assurance integrity for critical environments.
- **Approach**:
  - After snapshot commit, compute checksums (e.g., SHA256) of key files and store in `.archon/checksums/<tag>.json` or embed in annotated tag message.
  - On integrity check command or load, recompute and compare; alert user on mismatch indicating possible corruption or external modification.
- **Optional Feature**: Label as advanced; implement post-MVP based on user feedback.
- **Implementation**:
  - Modular checksum service; allow enabling/disabling via `archon.json` settings.

#### 2.3 Automated Integrity Tests in CI
- On CI, run schema validation against sample snapshots; verify no drift across versions; ensure formatting and schema updates are backwards compatible.

---

### 3. Developer Tools & Observability

Provide visibility into internal state aiding debugging for core team and plugin authors.

#### 3.1 Debug Panel / Developer Console
- **Plugin Log Viewer** (Priority 1): Central feature showing plugin logs tagged by plugin_id, timestamp, and level; filter by plugin or severity. Essential for plugin developers and immediate insight.
- **IPC Monitor** (Priority 1): Display IPC calls between frontend and backend: method names, arguments, return values, timestamps. Crucial for diagnosing boundary bugs in Wails-based architecture.
- **Git State Inspector**: Show current Git branch, uncommitted changes, last commit info, tags; offer buttons to open external Git tools for advanced operations.
- **Tree Diff Explorer**: Interactive view to load two JSON states and display semantic diff; useful for developers to validate diff engine behavior on arbitrary data.
- **Error Log Viewer**: Aggregate backend/frontend error logs with stack traces, timestamps, and context; allow searching/filtering.
- **Performance Metrics Panel**: Display metrics like diff computation time, plugin execution durations, autosave durations, memory usage spikes during heavy operations; assist performance tuning.
- **Access Patterns**:
  - Developer Mode Toggle: Hidden by default; enable via settings or environment variable.
  - Keyboard Shortcuts/Menu: Quick access to debug panel when in dev mode.

#### 3.2 Logging Infrastructure
- **Log Levels**: DEBUG, INFO, WARN, ERROR.
- **Structured Logging**: Use JSON-lines format for logs, capturing fields: timestamp, level, component (e.g., DiffService), correlation ID, message, context data.
- **Log Sinks**:
  - **Console**: In dev mode, logs to terminal.
  - **File**: Rotating logs under `.archon/logs/archon.log`, segmented by date or size.
  - **In-App Viewer**: Display recent logs in Debug Panel.
- **Correlation IDs**: Assign unique ID per operation (e.g., snapshot.create, diff.compute) logged across steps to trace flow in logs.

---

### 4. Error Reporting Strategy

Coherent error reporting improves user experience and facilitates issue diagnosis.

#### 4.1 UI Notifications
- **Transient Toasts**: Non-critical events (e.g., snapshot success, autosave done) with INFO level.
- **Modal Dialogs**: Critical errors blocking operations (e.g., schema validation failure before snapshot, merge conflicts). Provide clear explanation, remediation steps, and link to relevant docs.
- **Inline Validation**: For form inputs (e.g., invalid IP format), show inline errors.

#### 4.2 Logging and Persistence
- **Error Logs**: Capture errors with stack traces in `.archon/logs/`; include context (operation name, component IDs).
- **Crash Reporting**:
  - Catch uncaught exceptions/panics in frontend/backend.
  - Write local crash report with timestamp, error details, minimal context.
  - **Sanitized Issue Bundle**: In “Report Issue” dialog, automatically package:
    - Recent log files (`archon.log`).
    - Sanitized `archon.json` (removing sensitive values).
    - Application metadata (Archon version, OS, Plugin API version).
  - Prompt user to send anonymized bundle to support or GitHub issue, with clear explanation and privacy assurances.
- **Telemetry (Optional & Consent-Based)**:
  - With user opt-in, collect anonymized usage metrics and error incidence; send securely; respect privacy.

#### 4.3 Developer Feedback Loop
- **In-App “Report Issue”**: Trigger packaging of sanitized bundle; allow user to add comments and send to support endpoint or copy link to GitHub issue template.
- **Documentation Links**: On known error conditions, provide direct links to relevant docs or FAQ entries in dialogs.

---

### 5. CI/CD Integration & Quality Gates

Automate tests and enforce standards to ensure stability on every change.

#### 5.1 CI Pipelines
- **Unit Tests**: Run backend (Go) and frontend (JS/TS) unit tests on every commit/PR.
- **Integration Tests**: Use ephemeral Git repos for snapshot/diff/merge tests; plugin test suite; storage migration tests.
- **E2E Tests**: Playwright automation for UI flows in headless mode; use fixtures and test projects covering core journeys.
- **Linting & Formatting**: Enforce canonical JSON formatting, Go fmt, ESLint, and commit message conventions.
- **Security Scans**: Dependency vulnerability checks.

#### 5.2 Quality Gates
- **Test Coverage**: Minimum coverage thresholds for critical modules: DiffService, PluginLoader, SnapshotService.
- **Regression Tests with Golden Files**: For diff engine and plugin outputs, compare against golden fixtures to catch unintended changes.
- **Performance Baseline**: Monitor diff computation time on representative large hierarchies; fail CI if regression exceeds threshold.

#### 5.3 Release Automation
- On CI success, build cross-platform binaries via Wails; run smoke tests on artifacts.
- Generate changelog from commit history; package and publish releases.

---

### 6. Implementation Guidance & Best Practices

- **Modular Design & Dependency Injection**: Write testable modules (Storage, DiffService, PluginLoader) accepting interfaces for filesystem, Git, network.
- **Mock Data Repositories**: Provide fixtures for plugin tests and diff tests; include representative sample hierarchies in test repo.
- **Isolated Test Environments**: Use temporary directories and Git repos in tests to avoid side effects on user data.
- **Logging in Tests**: Capture and inspect logs to assert expected behavior during error conditions.
- **Incremental Debug Panel Development**: Start with Plugin Log Viewer and IPC Monitor, then add Git State Inspector and Tree Diff Explorer.
- **Documentation for Contributors**: Add guidelines on writing plugin tests, adding diff test cases, extending CI pipelines.
- **Progressive Enhancement**: Begin with JSON Schema validation and core unit tests; add optional checksum features post-MVP based on user need.

---

*End of Testing & Debugging Infrastructure Deep Dive (Refined with Golden Files, PluginContext Emphasis, Sanitized Issue Bundles)*

