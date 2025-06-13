## Archon: Tooling, CI/CD & Packaging Strategy

To support team collaboration, maintain high release velocity, and deliver a seamless user experience, Archon requires a robust tooling, continuous integration/continuous deployment (CI/CD), and packaging strategy across platforms. This document outlines:
1. Developer Tooling & Local Workflows
2. CI System for Building & Testing (Go + Svelte)
3. Cross-Platform Packaging & Release Automation
4. Auto-Update Mechanism & Notification Strategy
5. Git Versioning Strategy for Application and Plugins
6. Security & Integrity in CI/CD
7. Monitoring & Metrics (Post-Release)

---

### 1. Developer Tooling & Local Workflows

#### 1.1 Standardized Project Layout
- **Mono-repo Structure**: Keep frontend (Svelte) and backend (Go) in a single repository with clear directories:
  - `/cmd/archon` – main Go application entrypoint.
  - `/internal/` or `/pkg/` – Go packages (Git abstraction, storage, plugin loader, etc.).
  - `/frontend/` – Svelte project (Vite-based). Contains source, assets, i18n files.
  - `/plugins/hello-world-template` – example plugin scaffold.
  - `/scripts/` – helper scripts (e.g., for extraction, signing CLI).
  - `/ci/` – CI configuration helpers or shared workflows.
- **Consistent CLI**: Develop a CLI wrapper (e.g., `archon dev`, `archon build`, `archon test`) to simplify common tasks for contributors.

#### 1.2 Local Build & Test Commands
- **Backend Build**: Provide `go build` integration:
  - `go mod tidy` to ensure dependencies.
  - Build flags for embedding version info (e.g., via ldflags: version, commit hash, build date).
- **Frontend Build**: Provide `npm install` and `npm run build` under `/frontend`:
  - Vite build produces production assets consumed by Wails.
  - Development mode: `npm run dev` with live-reload, proxying to Go backend if needed.
- **Combined Build**: A top-level script or Makefile invoking frontend build then Wails build to bundle Go and assets.
- **Local Testing**:
  - **Go Unit Tests**: `go test ./...` including coverage reports.
  - **Frontend Unit Tests**: e.g., Jest or Vitest for Svelte components where applicable.
  - **Integration Tests**: Scripts launching the app in test mode for basic flows; can be part of CI.
  - **Plugin Tests**: Use test harness to run plugin code against mocked PluginContext. Include commands to run plugin unit tests.
- **Linting & Formatting**:
  - **Go**: `go fmt`, `golangci-lint` with rules for code quality.
  - **Svelte/TS**: ESLint, Prettier, stylelint if CSS linting.
  - Pre-commit hooks (e.g., Husky) to enforce formatting before commit.

#### 1.3 Developer Experience Enhancements
- **Hot-Reload for Frontend**: In dev mode, auto-reload Svelte changes; integrate with Wails dev mode if possible.
- **Mock Services**: Provide mock Git repos or test fixtures to quickly test snapshot/diff features locally.
- **Debug Tools**: Enable Debug Panel in dev builds; allow verbose logging.
- **Documentation Generation**: Scripts to generate API docs for Go packages and plugin API reference; host locally (e.g., via `godoc` or markdown generation).

---

### 2. CI System for Building & Testing

#### 2.1 CI Provider Choice & Configuration
- **Provider**: Use GitHub Actions (or similar: GitLab CI, CircleCI) for cross-platform support and integration with GitHub repository.
- **Matrix Builds**:
  - **OS Matrix**: macOS-latest, ubuntu-latest, windows-latest. Ensure builds on all three OSes.
  - **Go Versions**: Test against supported Go versions (e.g., Go 1.20+, depending on chosen version).
  - **Node Versions**: For frontend, test with maintained Node LTS versions.
- **Workflow Triggers**:
  - **Push to main/develop branches**: Run full build/test matrix.
  - **Pull Request**: Run linting, unit tests, frontend tests, and build checks.
  - **Tag/Release**: Trigger packaging jobs for release artifacts.

#### 2.2 CI Steps
- **Checkout Code**: `actions/checkout` with submodules if used.
- **Setup Go**: `actions/setup-go` with specified version; run `go mod download`.
- **Setup Node**: `actions/setup-node` with specified version; `npm ci` in `/frontend`.
- **Lint & Format Checks**:
  - Run `golangci-lint run`; run ESLint/Prettier checks.
- **Unit Tests**:
  - Run `go test ./... -coverprofile=coverage.out`.
  - Run frontend tests: `npm run test` or Vitest tests.
- **Build Frontend**:
  - `npm run build`; ensure no errors.
- **Build Go & Bundle**:
  - Invoke Wails build commands: embed frontend assets and compile Go binary for host OS.
  - Capture build artifacts (binaries, HTML/JS assets).
- **Integration/E2E Tests (Optional)**:
  - Launch application in headless/test mode and run basic smoke tests (e.g., open a temp project, perform operations via CLI or UI scripts).
- **Artifact Upload**:
  - On success, upload build artifacts (binaries, installers) to CI storage for later packaging.
- **Code Coverage & Reports**:
  - Publish coverage reports; optionally fail if coverage below threshold.
- **Security Scans**:
  - Static analysis tools for dependencies (e.g., `go audit`, npm audit) to flag vulnerabilities.

#### 2.3 Secrets & Credentials in CI
- **Protected Secrets**:
  - Store signing certificates/keys (for code signing on Windows/macOS) securely in CI secrets vault.
  - Git hosting tokens for publishing releases or interacting with plugin registry (future).
- **Least Privilege**:
  - Limit access of secrets to specific jobs (e.g., only on release tags).
- **Environment Isolation**:
  - Use ephemeral runners or containers to avoid leftover state.

---

### 3. Cross-Platform Packaging & Release Automation

#### 3.1 Packaging Tools & Strategies
- **Wails Packaging**:
  - Use Wails’ built-in packaging support to produce native bundles:
    - **Windows**: Generate `.exe` and installer (e.g., NSIS or WiX) to include binary, dependencies, and assets. Optionally create portable ZIP.
    - **macOS**: Generate `.app` bundle, code-signed and notarized for Gatekeeper compliance. Use Apple Developer certificates stored in CI secrets.
    - **Linux**: Produce AppImage for wide compatibility; optionally DEB/RPM for specific distributions. Consider Flatpak manifest for sandboxed distribution later.
- **Installer Generation**:
  - **Windows Installer**: Use NSIS or WiX; configure shortcuts, uninstaller entry.
  - **macOS Notarization**: In CI, after building `.app`, code sign and submit for notarization via Apple notarization API. Use `xcrun altool` in macOS runner.
  - **Linux AppImage**: Use tools like `appimagetool`; ensure bundled runtime dependencies (Go static binary plus frontend assets) are included.
- **Version Embedding**:
  - Embed application version (SemVer) and commit hash at build time via Go ldflags; include in UI About screen.
- **Checksum & Artifact Verification**:
  - Generate SHA256 checksums for installers/artifacts; publish checksums alongside downloads.

#### 3.2 Release Automation
- **Release Trigger**:
  - On Git tag push (matching release SemVer pattern, e.g., `v1.2.3`), trigger packaging workflow.
- **Artifact Publishing**:
  - Upload installers to GitHub Releases (or alternative distribution channel).
  - Update download pages or Homebrew/Cask formulas (if supported).
- **Release Notes Generation**:
  - Automate changelog generation from commit messages (e.g., using conventional commits) to create release notes.
- **Sentry/Crash Reporting Integration** (Optional MVP+):
  - Integrate with crash-reporting service; include DSN in build to capture crash telemetry for post-release diagnostics.

#### 3.3 Signing & Security
- **Code Signing**:
  - **Windows**: Sign executables and installer with Authenticode certificate stored securely in CI.
  - **macOS**: Sign `.app` and notarize.
  - **Linux**: GPG-sign AppImage or provide signature files alongside.
- **Plugin Signing Verification**:
  - In packaging, ensure that bundled example plugins are signed and/or include public keys for trust store.

---

### 4. Auto-Update Mechanism & Notification Strategy

#### 4.1 Update Check Strategy
- **Periodic Update Checks**:
  - On app startup (and optionally periodically in background), check a trusted endpoint for latest version metadata (e.g., JSON manifest on CDN).
  - Compare current version to latest SemVer.
- **User Notification**:
  - If newer stable version available, show non-blocking notification/prompt: “New version X.Y.Z available. Download now?”
  - Provide link to download or trigger in-app update if implemented.
- **Privacy & Network**:
  - Respect offline mode; provide setting to disable update checks.
  - Perform checks over HTTPS; validate certificate.

#### 4.2 In-App Update Mechanism (Optional for MVP)
- **Simple Approach**:
  - Direct user to download/install manually from releases page or website.
  - For MVP, notification + link may suffice; avoid complex auto-update implementation initially.
- **Future Auto-Update**:
  - Evaluate using established frameworks:
    - **Windows**: Squirrel.Windows or WinGet integration for auto-update.
    - **macOS**: Sparkle framework adapted for Wails apps; or use Homebrew Cask auto-update flows.
    - **Linux**: AppImage update tools (e.g., AppImageUpdate) or integrate with distro package managers (Snap, Flatpak) for auto-updates.
  - Secure update pipeline: signed update packages; verify signature before installing.

#### 4.3 Update Rollout & Telemetry
- **Staged Rollout**:
  - Optionally roll out updates gradually (e.g., beta channel vs stable) based on version channels.
- **Telemetry for Update Adoption**:
  - If privacy policy allows and with user consent, track update check success and adoption rates (generic metrics, no personal data) to inform release cadence.

---

### 5. Git Versioning Strategy for Application and Plugins

#### 5.1 Application Versioning
- **SemVer Discipline**:
  - Use Semantic Versioning (MAJOR.MINOR.PATCH) for Archon releases.
  - Enforce version bump policy: breaking changes increment MAJOR; new features increment MINOR; bug fixes increment PATCH.
- **Tagging**:
  - Create annotated Git tags for each release (e.g., `v1.0.0`). Include release notes in tag annotation or linked changelog.
- **Branching Model**:
  - Use GitFlow-like or trunk-based workflow:
    - Main/trunk branch holds stable code.
    - Feature branches merge via PRs into main.
    - Release branches or tags for preparing releases; hotfix branches for urgent fixes.
  - Document chosen model in CONTRIBUTING.md.
- **Commit Message Conventions**:
  - Adopt Conventional Commits or similar to facilitate changelog automation.

#### 5.2 Plugin Versioning & Compatibility
- **Plugin SemVer**:
  - Plugins follow their own SemVer independent of Archon core, but declare compatibility range with Archon API version.
  - Manifest includes `archonApiVersion: "^1.0.0"` to specify compatible core versions.
- **Plugin Releases**:
  - Plugins maintained in separate repos; use annotated tags for plugin versions.
  - Publish plugin releases (e.g., GitHub Releases) where users can download signed plugin packages.
- **Compatibility Checks**:
  - At plugin load time, Archon verifies plugin’s declared compatibility against its own version; warn or disable if incompatible.

#### 5.3 Changelog & Release Notes
- **Automated Generation**:
  - Use commit messages and PR titles to auto-generate changelog entries between tags.
- **Publishing**:
  - Include changelog in GitHub Releases, documentation site, and within Archon’s “About” or “Release Notes” view.

---

### 6. Security & Integrity in CI/CD

#### 6.1 Dependency Auditing
- **Go Dependencies**:
  - Use `go mod verify` and vulnerability scanning tools (e.g., `govulncheck`).
- **Frontend Dependencies**:
  - Run `npm audit`; use Dependabot or similar to keep dependencies up to date.

#### 6.2 Code Signing & Secrets Management
- See Section 3 for signing installers; securely manage signing keys in CI secrets.
- **CI Permissions**:
  - Limit write access; use least-privilege tokens for publishing releases or interacting with plugin registry.

#### 6.3 Artifact Verification
- **Checksums & Signatures**:
  - Publish checksums and signatures for release artifacts; users or update mechanism verify before installation.

#### 6.4 Secure Update Channels
- **HTTPS Endpoints**:
  - Host update metadata and binaries on secure HTTPS endpoints with valid certificates.
- **Signature Verification**:
  - If implementing auto-update, verify update package signatures before applying.

---

### 7. Monitoring & Metrics (Post-Release)

#### 7.1 Telemetry (Opt-In)
- **Crash Reporting**:
  - Integrate opt-in crash reporting (Sentry or similar) to capture unhandled errors and stack traces.
- **Usage Metrics**:
  - Collect anonymized usage metrics (with user consent) to understand feature adoption (e.g., how often snapshots taken, plugin usage frequency).
- **Update Adoption**:
  - Track how many users update to new versions; measure time to adoption.

#### 7.2 Logs & Diagnostics
- **User Logs**:
  - Provide easy way for users to collect logs for troubleshooting; include in docs.
- **Remote Error Alerts**:
  - Aggregate crash reports or error logs in dashboards for engineering to monitor stability.

---

*End of Tooling, CI/CD & Packaging Strategy Deep Dive*

