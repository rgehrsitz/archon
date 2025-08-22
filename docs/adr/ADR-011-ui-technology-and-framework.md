# ADR-011: UI Technology for Flexibility, Compatibility & UX

Status: Accepted
Date: 2025-08-22
Owners: Archon Core

## Context

We need a desktop UI with excellent cross-platform support, strong Go integration, low runtime overhead, and a modern component ecosystem. Candidates: **Wails (Go + system WebView)**, **Tauri (Rust + system WebView)**, **Electron (Chromium + Node)**.

## Decision

- **Framework:** **Wails v2** for production today (v3 is in alpha; watch it for multi-window & API improvements).  [oai_citation:30‡v3alpha.wails.io](https://v3alpha.wails.io/?utm_source=chatgpt.com)
- **Frontend stack:** **Svelte + TypeScript** in the Wails frontend (but we keep the UI framework pluggable—React/Vue also work).
- **Why Wails?**
  - Built for **Go**: native bindings and event bridge reduce glue code vs Rust (Tauri).  [oai_citation:31‡wails.io](https://wails.io/?utm_source=chatgpt.com)
  - Uses **system WebView** (WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux), yielding small bundles and low memory compared with Chrome-bundled stacks.  [oai_citation:32‡wails.io](https://wails.io/blog/the-road-to-wails-v3/?utm_source=chatgpt.com) [oai_citation:33‡Tauri](https://v2.tauri.app/reference/webview-versions/?utm_source=chatgpt.com)
  - Actively maintained; v2 stable for production; v3 adds multi-window/tray when we’re ready.  [oai_citation:34‡wails.io](https://wails.io/changelog/?utm_source=chatgpt.com) [oai_citation:35‡v3alpha.wails.io](https://v3alpha.wails.io/whats-new/?utm_source=chatgpt.com)
- **Why not Tauri/Electron for v1?**
  - **Tauri** is excellent on security & footprint, but it centers on Rust; our backend is Go, so we’d add a Rust layer or IPC hop. (Good benchmark; revisit only if Wails lacks required Web APIs on a target OS.)  [oai_citation:36‡Tauri](https://v2.tauri.app/security/?utm_source=chatgpt.com)
  - **Electron** offers maximum Web API compatibility (Chromium) and mature auto-update tooling, but it **bundles Chromium**, increasing app size and memory usage—counter to our lightweight goal.  [oai_citation:37‡electronjs.org](https://electronjs.org/docs/latest/?utm_source=chatgpt.com)

## Rationale

Wails lets us build a **native-feeling**, lightweight desktop app with a Go core and modern web UI, matching our team’s language and the spec’s portability goals. Using the OS WebView minimizes footprint while keeping the door open to React/Vue if desired.  [oai_citation:38‡wails.io](https://wails.io/?utm_source=chatgpt.com)

## Alternatives Considered

- **Tauri**: Great security model & small bundles; Rust adds toolchain complexity for a Go-first team.  [oai_citation:39‡Tauri](https://v2.tauri.app/security/?utm_source=chatgpt.com)
- **Electron**: Max compatibility and ecosystem, but heavier runtime due to embedded Chromium.  [oai_citation:40‡electronjs.org](https://electronjs.org/docs/latest/?utm_source=chatgpt.com)

## Consequences

- Positive: Tight Go integration; small installers; native menus/dialogs; modern UI via Svelte.
- Negative: OS WebView **feature variance** (e.g., some advanced media/WebRTC APIs differ across platforms); if a critical Web API is missing on Linux WebKit, we may need a targeted Electron or Chromium-backed fallback window for that feature only.  [oai_citation:41‡GitHub](https://github.com/wailsapp/wails/discussions/2697?utm_source=chatgpt.com)
- Follow-ups: Monitor Wails v3 maturity; add browser-feature probes and graceful degradation.

## Implementation Notes

- Templates: `wails init -t svelte` (or React). Supported platforms documented in Wails install docs.  [oai_citation:42‡wails.io](https://wails.io/docs/gettingstarted/installation/?utm_source=chatgpt.com)
- Multi-window/tray planned via Wails v3 when it exits alpha.  [oai_citation:43‡v3alpha.wails.io](https://v3alpha.wails.io/whats-new/?utm_source=chatgpt.com)

## Review / Revisit

- Revisit if OS WebView gaps block key workflows, or when Wails v3 stabilizes (multi-window, tray, improved bindings).
