## Archon: Security & Trust Model

Security and trust are foundational for Archon, especially given its extensible plugin architecture and Git integration involving remote repositories and credentials. This specification covers:
1. Sandboxing plugins (frontend focus, backend deferred).
2. Signing and verifying plugins or plugin manifests with an MVP CLI.
3. Authentication and identity: attributing changes; protecting Git remote credentials via OS keyrings.
4. Secure storage of user configuration and credentials (platform keyrings, optional encryption as advanced feature).
5. Secure IPC and data flows with strict validation.
6. Permissions model, least-privilege principles, and optional user roles.
7. Audit and logging for security events, with integrity protections.
8. Implementation guidance and adversarial testing.

---

### 1. Plugin Sandboxing

#### 1.1 Execution Contexts
- **Frontend (JS/TS) Plugins Only (MVP Focus)**:
  - Run entirely within a sandboxed environment in the WebView context. No direct filesystem or network access.
  - All sensitive operations (filesystem, network, Git) must go through Archon’s vetted PluginContext API which enforces permissions.
  - Defer support for backend (Go) plugins to later versions, due to higher complexity and attack surface.

#### 1.2 Permission Model
- **Declarative Permissions**:
  - Plugin manifests (`archon-plugin.json`) must declare required permissions (e.g., `file_system:read`, `file_system:write`, `network:access`, `git:execute`, `plugin_data_access`).
  - Permissions scoped and restricted: e.g., `file_system:read:{paths}`, `network:access:{hosts}`. Default deny.
- **User Consent & Onboarding**:
  - On installation or first run, present clear dialog listing permissions with plain-language explanations (e.g., “This plugin will read files in your project to import configurations”).
  - User approves or denies; denial disables plugin or certain features.
- **Runtime Enforcement**:
  - PluginContext API enforces declared permissions at each call.
  - Filesystem operations constrained to project directory or `.archon/plugin_data` subdirectory; network calls restricted to whitelisted hosts.
  - Git operations limited to repository within project root.
  - UI actions limited to plugin’s designated UI area; no arbitrary DOM manipulation outside.

#### 1.3 Isolation Techniques
- **JS Sandbox**:
  - Execute plugins in isolated JS contexts (e.g., sandboxes or iframes with restricted globals) so they cannot access global window or Node APIs directly.
  - Prevent dynamic code execution via `eval` or similar in plugin code where possible; if needed, review rigorously.
- **Resource Limits & Cancellation**:
  - Enforce CPU/timeouts and memory limits on plugin operations. Use cancellation tokens for long-running tasks to allow aborting if exceeding limits.
  - Monitor performance to detect runaway behavior.

---

### 2. Plugin Signing and Verification

A pragmatic MVP approach provides strong security without requiring centralized infrastructure initially.

#### 2.1 Plugin Manifest Signing (MVP)
- **Manifest Structure**:
  - `archon-plugin.json` includes metadata: plugin ID, version, author, permissions, compatible Archon API versions, and a signature field (e.g., base64-encoded).
- **CLI Signing Tool**:
  - Provide a simple CLI that plugin authors use to sign their manifest (and optionally code checksums) with their private key (e.g., RSA/ECDSA).
  - Document usage in dev guides: generating key pairs, storing private keys securely, and producing signatures.
- **Local Trust Store & UI**:
  - Archon maintains a local trust store of public keys. Initially empty; users add trusted plugin author keys via UI.
  - On loading a plugin, Archon verifies manifest signature against trust store. If valid, plugin loads; if missing or invalid, warn user and request explicit override to enable.
- **Trust Levels**:
  - **Official Plugins**: Signed by Archon maintainers; their public key included by default, auto-trusted.
  - **Third-Party Plugins**: Signed by author; user adds author’s public key to trust store to allow.
  - **Unsigned Plugins**: Allowed only in development mode or after explicit user override, with a prominent warning about risks.

#### 2.2 Code Integrity and Compatibility
- **Checksum Verification**:
  - Plugin package includes checksums (e.g., SHA256) of code files. After installation, Archon verifies code matches expected checksums before execution.
- **API Version Constraints**:
  - Manifest declares compatible Archon API SemVer range. On load, Archon verifies compatibility; incompatible plugins are disabled or prompt update.

#### 2.3 Plugin Distribution & Future Registry
- **Local Install**:
  - Users install plugins from filesystem; Archon checks signatures or warns on unsigned.
- **Official Registry (Future)**:
  - Plan for a centralized plugin registry where plugins are vetted and signed by Archon team. For MVP, defer but design manifest and signing to be compatible later.
- **Updates**:
  - Allow plugin update workflow: check new version signatures, verify compatibility, and prompt user to trust updated author keys if changed.

---

### 3. Authentication & Identity Model

#### 3.1 Change Attribution
- **Git-Based Attribution**:
  - Use Git author and committer fields (user.name and user.email) for snapshots and commits. Configure via Archon settings on first run.
- **User Profile Setup**:
  - On initial launch, prompt user to configure name/email; store in Archon settings (not containing secrets).
- **Audit Logs**:
  - Maintain fine-grained audit logs for critical actions (e.g., plugin installs, permission grants, credential access) under `.archon/logs/audit.log`, with timestamp, user identity, and operation details.
  - Protect audit log integrity via Git tracking or checksums; consider signing entries to detect tampering.

#### 3.2 Protecting Git Remote Credentials
- **OS Keyring Integration** (Highlight as major selling point):
  - Store Git credentials (HTTPS tokens, SSH passphrases) securely in platform-native keyrings:
    - Windows Credential Manager
    - macOS Keychain
    - Linux Secret Service (Libsecret) or KDE Wallet; fallback to encrypted local store with user passphrase, with clear warning.
  - When performing Git operations, retrieve credentials from keyring; never store plaintext in config files.
- **SSH Key Usage**:
  - Support user’s existing SSH keys; if passphrase-protected, prompt and cache in memory or keyring temporarily.
- **OAuth Tokens**:
  - For GitHub/GitLab, allow user to input personal access tokens with minimal scopes; store securely.
  - If implementing OAuth flow, handle tokens securely and store in keyring.
- **Credential Lifecycle**:
  - Provide UI to manage stored credentials: add, remove, rotate tokens. Notify user when credentials expire or become invalid.
- **Network Security**:
  - Enforce TLS for HTTPS; rely on system SSH for SSH.

---

### 4. Secure Storage of User Config & Optional Encryption

#### 4.1 Minimal Sensitive Data in Config Files
- Store only non-sensitive settings in versioned files (`archon.json`). Sensitive references (e.g., keyring identifiers) stored in keyring, not in plaintext.
- Ensure config files have restrictive file permissions (user-only read/write).

#### 4.2 OS Keyring & Encrypted Local Store
- **Primary**: Use OS keyring for credentials and plugin secrets (e.g., API keys). Namespace entries per plugin or Git host.
- **Fallback**: If no keyring, prompt user to set a passphrase to encrypt local store (e.g., encrypted file in `.archon/`); warn about risks and require passphrase on startup or before operations needing secrets.

#### 4.3 Optional Advanced Encryption
- **Full Project Encryption** (Enterprise Feature): Offer to encrypt project data (components, attachments) with user passphrase for high-assurance scenarios. Use strong encryption algorithms; manage keys via keyring or passphrase-derived keys.
  - Defer until after core release; design storage abstraction to accommodate encryption layer later.

---

### 5. Secure IPC & Data Flows

#### 5.1 IPC between Frontend and Backend
- **Input Validation**:
  - Every IPC endpoint validates and sanitizes inputs: file paths constrained to project directories, component IDs validated against expected patterns, size limits on payloads to prevent resource exhaustion.
- **Permission Checks for Plugin-Initiated Calls**:
  - When plugin triggers IPC, verify plugin’s declared permissions before executing sensitive operations.
- **Output Sanitization**:
  - Backend responses must not leak internal sensitive data; errors return sanitized messages while detailed stack traces go to logs.
- **Error Handling**:
  - Return structured error objects; log full details internally.

#### 5.2 Network Communications
- **TLS Verification**:
  - Enforce certificate validation for all network calls, including plugin sync operations.
- **Proxy Support**:
  - Honor system proxy settings; allow user configuration if needed.

#### 5.3 File Operations
- **Path Sanitization & Atomic Writes**:
  - Prevent directory traversal; constrain writes to project root or approved plugin_data directories; use atomic temp-and-rename patterns.
- **Temporary Files**:
  - Store only in `.archon/temp/`; ensure these are not world-readable.

---

### 6. Permissions Model & Least-Privilege Principles

- **Principle of Least Privilege**:
  - Plugin permissions default to minimal; require explicit manifest declaration and user consent for additional scopes.
  - Core modules operate with minimal necessary privileges (e.g., only access project files when needed).
- **User Roles (Optional Advanced)**:
  - For multi-user local setups, optionally define roles (admin vs read-only) controlling plugin installation or project setting changes. Use local Archon user accounts separate from OS user accounts.
- **Runtime Checks**:
  - Every sensitive operation (filesystem, network, Git) enforces permission context from PluginContext or core services.

---

### 7. Audit & Logging for Security Events

- **Audit Logs**:
  - Log security-relevant events: plugin installations/updates, permission grants/denials, failed permission attempts, credential access, encryption operations.
  - Entries include timestamp, user identity, event details. Append-only under `.archon/logs/audit.log`.
- **Log Integrity**:
  - Protect audit logs via Git tracking or checksums; consider signing entries to detect tampering.
- **User Notifications**:
  - Notify user on key events: e.g., “Plugin X requested permission Y; you granted.”
- **Review Tools**:
  - Expose audit log viewer in Debug Panel (developer mode) for inspection and troubleshooting.

---

### 8. Implementation Guidance & Adversarial Testing

#### 8.1 Sandboxing & Permission Enforcement Tests
- **Adversarial Plugin Tests**:
  - Actively develop malicious test plugins attempting to bypass sandbox: read arbitrary filesystem paths, perform unauthorized network calls, exceed resource limits. Verify sandbox blocks these attempts.
- **Permission Enforcement**:
  - Mock PluginContext with denied permissions; test that API calls are rejected cleanly without side effects.
- **Resource Limits**:
  - Simulate CPU/memory exhaustion within plugin; ensure timeouts or memory caps terminate execution safely.

#### 8.2 Signing & Verification Tests
- **Signature Validation**:
  - Test valid and invalid manifest signatures; simulate tampered manifests and verify Archon rejects or warns.
- **Compatibility Checks**:
  - Load plugins with mismatched API versions; verify proper disablement and user notification.

#### 8.3 Credential Storage Tests
- **Keyring Integration**:
  - On Windows, macOS, Linux: test storing and retrieving credentials; simulate missing keyring and fallback encrypted store flows.
- **Passphrase Handling**:
  - For fallback encryption, test correct handling of passphrase prompts, incorrect passphrase rejection, and secure in-memory handling.

#### 8.4 IPC & Input Validation Tests
- **Malformed IPC Requests**:
  - Send invalid or malicious IPC payloads; backend validation should reject safely without crashes.
- **Boundary Conditions**:
  - Test large payloads against size limits to prevent DoS.

#### 8.5 Audit & Log Integrity Tests
- **Audit Entry Generation**:
  - Simulate security events; verify correct logging.
- **Tamper Detection**:
  - Modify audit log entries manually; run integrity checks to detect alterations.

#### 8.6 End-to-End Security Testing
- **Threat Modeling**:
  - Identify and document potential threats (malicious plugin, compromised credentials, Git repo tampering).
- **Penetration Tests**:
  - Conduct tests: plugin attempts to escape sandbox; unauthorized file access; intercepting IPC.
- **User Flows**:
  - Validate secure onboarding: plugin install and permission dialogs; credential storage and retrieval; encryption enabling/disabling.

---

*End of Security & Trust Model Deep Dive (Refined with MVP Signing CLI, Frontend-Only Plugins, Keyring Emphasis, Adversarial Testing)*

