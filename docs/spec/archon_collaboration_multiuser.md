## Archon: Tooling, CI/CD & Packaging Strategy

To support team collaboration, maintain high release velocity, and deliver a seamless user experience, Archon requires a robust tooling, continuous integration/continuous deployment (CI/CD), and packaging strategy across platforms. This document outlines:
1. Developer Tooling & Local Workflows
2. CI System for Building & Testing (Go + Svelte)
3. Cross-Platform Packaging & Release Automation
4. Auto-Update Mechanism & Notification Strategy
5. Git Versioning Strategy for Application and Plugins
6. Security & Integrity in CI/CD
7. Monitoring & Metrics (Post-Release)
8. Optional/Future Considerations

---

### 1. Developer Tooling & Local Workflows

#### 1.1 Standardized Project Layout
- **Mono-repo Structure**: Keep frontend (Svelte) and backend (Go) in a single repository with clear directories:
  - `/cmd/archon` – main Go application entrypoint.
  - `/internal/` or `/pkg/` – Go packages (Git abstraction, storage, plugin loader, etc.).
  - `/frontend/` – Svelte project (Vite-based). Contains source, assets, i18n files.
  - `/plugins/hello-world-template` – example plugin scaffold with its own CI in its own repo.
  - `/scripts/` – helper scripts (e.g., for extraction, signing CLI).
  - `/ci/` – CI configuration helpers or shared workflows.
- **Scalability Note**: For larger codebases, consider tools like Nx or turborepo in future; for MVP, a simple structured mono-repo is sufficient.
- **Consistent CLI**: Provide a CLI wrapper (e.g., `archon dev`, `archon build`, `archon test`) to simplify common tasks for contributors.

#### 1.2 Local Build & Test Commands
- **Backend Build**: `go build` integration:
  - `go mod tidy` to manage dependencies.
  - Embed version info (version, commit hash, build date) via ldflags.
- **Frontend Build**: `npm ci` and `npm run build` in `/frontend`:
  - Vite build produces production assets for Wails.
  - Development mode: `npm run dev` with hot-reload, leveraging Wails’ dev integration for live-reload of frontend changes.
- **Combined Build**: Top-level script or Makefile invoking frontend build then `wails build` to bundle Go and assets.
- **Local Testing**:
  - **Go Unit Tests**: `go test ./...`, coverage reports.
  - **Frontend Unit Tests**: Use Vitest or Jest for Svelte components.
  - **Integration Tests**: Scripts launching the app in test mode for basic flows; include plugin harness tests.
  - **Plugin Tests**: Each plugin repo has its own CI pipeline; in main Archon repo, use a mocked PluginContext test harness for core plugin API tests.
- **Linting & Formatting**:
  - **Go**: `go fmt`, `golangci-lint` rules.
  - **Svelte/TS**: ESLint, Prettier, stylelint.
  - Pre-commit hooks (e.g., Husky) enforce formatting before commit.

#### 1.3 Developer Experience Enhancements
- **Hot-Reload for Frontend**: Wails v3 integrates with Vite dev server by default; changes in Svelte reflect immediately in the app.
- **Mock Services**: Provide mock Git repositories or test fixtures to quickly test snapshot/diff features locally.
- **Debug Tools**: Enable Debug Panel in dev builds; verbose logging via environment flags.
- **Documentation Generation & Maintenance**: Scripts to generate API docs for Go packages and plugin API reference; document process to keep docs updated with API changes.
- **SBOM Generation (Future)**: Plan to generate a Software Bill of Materials for releases to support regulated environments.

---

### 2. CI System for Building & Testing

#### 2.1 CI Provider & Configuration
- **Provider**: Use GitHub Actions (or equivalent) for cross-platform builds and integration.
- **Matrix Builds**:
  - **OS Matrix**: macOS-latest, ubuntu-latest, windows-latest.
  - **Go Versions**: Test against supported Go versions.
  - **Node Versions**: Test with maintained Node LTS versions.
- **Workflow Triggers**:
  - **Pull Requests**: Run linting, unit tests, frontend tests, build verification.
  - **Push to main/develop**: Full build/test matrix, integration tests.
  - **Tag/Release**: Trigger packaging and release workflows.

#### 2.2 CI Steps
- **Checkout Code**: `actions/checkout`.
- **Setup Environments**:
  - `actions/setup-go`, `go mod download`.
  - `actions/setup-node`, `npm ci` in `/frontend`.
- **Lint & Format Checks**:
  - Run `golangci-lint`; ESLint/Prettier.
- **Unit Tests**:
  - Go tests: `go test ./... -coverprofile=coverage.out`.
  - Frontend tests: `npm run test`.
- **Build Frontend**:
  - `npm run build`; ensure no errors.
- **Build & Bundle**:
  - `wails build` or equivalent: embed frontend assets, compile Go binary for host OS.
  - Capture build artifacts.
- **Integration/E2E Tests**:
  - Launch app in headless/test mode; automate basic workflows.
- **Plugin CI Integration**:
  - For official plugins maintained in separate repos, ensure their CI pipelines run tests and build signed releases. In Archon’s CI, include tests against Hello-World plugin using mocked context.
- **Artifact Upload**:
  - Upload build artifacts for packaging.
- **Coverage & Reports**:
  - Publish coverage; enforce thresholds.
- **Security Scans**:
  - Run `go vet`, `govulncheck`, `npm audit`.

#### 2.3 Secrets & Credentials
- **Protected Secrets**:
  - Signing certificates/keys for Windows/macOS in CI secrets vault.
  - GitHub tokens for publishing releases.
- **Least Privilege**:
  - Limit secrets access to release jobs.
- **Environment Isolation**:
  - Use ephemeral runners/containers.

---

### 3. Cross-Platform Packaging & Release Automation

#### 3.1 Packaging Tools & Strategies
- **Wails Packaging**:
  - **Windows**: Generate `.exe` and installer (NSIS/WiX), portable ZIP.
  - **macOS**: Generate `.app`, code-sign and notarize via Apple Developer certs.
  - **Linux**: Produce AppImage; optionally DEB/RPM and Flatpak manifest for future.
- **Installer Generation**:
  - **Windows Installer**: NSIS or WiX; configure shortcuts/uninstaller.
  - **macOS Notarization**: CI macOS runner signs and notarizes `.app`.
  - **Linux AppImage**: Use `appimagetool`; bundle dependencies.
- **Version Embedding**:
  - Embed SemVer and commit hash via ldflags; display in About screen.
- **Checksum & Verification**:
  - Generate SHA256 checksums; publish alongside artifacts.

#### 3.2 Release Automation
- **Release Trigger**:
  - On annotated Git tag push (e.g., `v1.0.0`), trigger packaging workflow.
- **Artifact Publishing**:
  - Upload installers to GitHub Releases; update Homebrew/Cask formulas if applicable.
- **Release Notes Generation**:
  - Automate changelog from Conventional Commits for release notes.
- **Documentation Updates**:
  - Publish updated docs alongside release.

#### 3.3 Signing & Security
- **Code Signing**:
  - **Windows**: Authenticode sign executables/installers.
  - **macOS**: Sign and notarize `.app`.
  - **Linux**: GPG-sign AppImage or provide signature files.
- **Plugin Signing Verification**:
  - Ensure bundled example plugins are signed; include public keys in trust store.

---

### 4. Auto-Update Mechanism & Notification Strategy

#### 4.1 Update Check Strategy
- **Periodic Checks**:
  - On startup (and optionally periodically), fetch version metadata from secure endpoint.
  - Compare current SemVer; notify user if newer.
- **User Notification**:
  - Non-blocking prompt: “New version X.Y.Z available. Download?” Links to release.
- **Privacy & Network**:
  - Respect offline mode; setting to disable checks; use HTTPS.

#### 4.2 In-App Update Mechanism (MVP vs Future)
- **MVP Approach**:
  - Notification + link to download manually; avoid complex auto-update in v1.
- **Future Auto-Update**:
  - Investigate frameworks: Squirrel.Windows/WinGet, Sparkle for macOS, AppImageUpdate for Linux, or package manager integration.
  - Ensure secure pipeline: signed updates, signature verification before install.

#### 4.3 Update Rollout & Telemetry
- **Staged Rollout**:
  - Support channels (beta vs stable) in future.
- **Telemetry (Opt-In)**:
  - Track update check success/adoption rates anonymously with consent.

---

### 5. Git Versioning Strategy for Application and Plugins

#### 5.1 Application Versioning
- **SemVer Discipline**:
  - MAJOR.MINOR.PATCH; document bump policy.
- **Tagging**:
  - Annotated Git tags (`v1.0.0`) with release notes.
- **Branching Model**:
  - Trunk-based or GitFlow; document in CONTRIBUTING.md.
- **Commit Conventions**:
  - Adopt Conventional Commits to automate changelog.

#### 5.2 Plugin Versioning & Compatibility
- **Plugin SemVer**:
  - Each plugin repo uses SemVer; manifest declares `archonApiVersion` compatibility (e.g., `^1.0.0`).
- **Plugin CI/CD**:
  - Encourage each plugin repo to have its own CI pipeline for tests and signed releases.
- **Compatibility Checks**:
  - At load time, Archon verifies plugin’s compatibility range; warns or disables incompatible plugins.

#### 5.3 Changelog & Release Notes
- **Automated Generation**:
  - Use commit messages and PR titles to generate changelog entries.
- **Publishing**:
  - Include changelog in GitHub Releases, website docs, and Archon’s “About” view.

---

### 6. Security & Integrity in CI/CD

#### 6.1 Dependency Auditing
- **Go Dependencies**:
  - `go mod verify`, `govulncheck` scans.
- **Frontend Dependencies**:
  - `npm audit`; Dependabot for updates.

#### 6.2 Code Signing & Secrets Management
- **CI Secrets**:
  - Store signing keys in CI secrets vault; restrict access.
- **Least Privilege**:
  - Use minimal-permission tokens for publishing.

#### 6.3 Artifact Verification
- **Checksums & Signatures**:
  - Publish SHA256 and GPG signatures for artifacts.

#### 6.4 Secure Update Channels
- **HTTPS Endpoints**:
  - Host update metadata and binaries on secure servers.
- **Signature Verification**:
  - Verify update packages before applying in future auto-update.

---

### 7. Monitoring & Metrics (Post-Release)

#### 7.1 Telemetry (Opt-In)
- **Crash Reporting**:
  - Integrate opt-in crash reporting (Sentry) to capture unhandled errors.
- **Usage Metrics**:
  - Anonymized metrics (with consent) on feature adoption (snapshot frequency, plugin usage).
- **Update Adoption**:
  - Track update check success/adoption rates.

#### 7.2 Logs & Diagnostics
- **User Logs**:
  - Provide easy log collection for troubleshooting; document in help.
- **Remote Error Alerts**:
  - Aggregate crash reports in dashboards for engineering monitoring.
- **SBOM Publication**:
  - Include SBOM alongside releases for transparency in regulated contexts.

---

### 8. Optional / Future Considerations
- **Monorepo Tooling**: Evaluate Nx, turborepo, or Bazel if project grows significantly.
- **Plugin Ecosystem CI**: Establish centralized CI templates or GitHub Actions workflows for plugin repos to standardize testing and release.
- **Auto-Update Frameworks**: Implement in-app update mechanisms using appropriate platform tools.
- **Flatpak/Snap Distributions**: For Linux, provide sandboxed distributions via Flatpak or Snap.
- **Advanced Security**: Harden build pipeline with reproducible builds, SBOM verification, supply chain security.
- **Analytics Dashboard**: Internal dashboards showing telemetry trends to guide product decisions.

---

*End of Tooling, CI/CD & Packaging Strategy Deep Dive (Updated with plugin CI/CD pipelines, Wails dev mode clarification, SBOM, and documentation maintenance emphasis)*

