# Plugin System Implementation Phases

This document outlines the phased implementation approach for the Comprehensive Plugin System as defined in ADR-013.

## Overview

The plugin system represents a major platform feature with 10 plugin types, sandboxed execution, permission management, and extensive host services. To manage complexity and deliver value incrementally, we'll implement in 3 phases.

> Status (2025-08-24): Frontend runtime, manifests, permissions, and discovery are implemented under `frontend/src/lib/plugins/`. Backend host services integration in Go is now the immediate next step (ADR-013 alignment), followed by UI integration and advanced types.

## Phase 1: Core Infrastructure (Foundation)

**Goal**: Establish the runtime, security, and basic plugin lifecycle.

### 1.1 Core Runtime
- [ ] Web Worker sandbox environment
- [ ] Plugin loader and lifecycle management
- [ ] TypeScript API definitions (`frontend/src/plugins/api.ts`)
- [ ] Plugin manifest system (manifest.json parsing and validation)
- [ ] Basic error handling and structured failures

### 1.2 Permission System
- [ ] Permission enumeration and validation
- [ ] Consent dialog UI for permission requests
- [ ] Runtime permission enforcement for host services
- [ ] Permission scoping (e.g., `secrets:jira*`)

### 1.3 Host Services (Core Set)
- [ ] `getNode()`, `listChildren()`, `query()` (readRepo permission)
- [ ] `apply()`, `commit()`, `snapshot()` (writeRepo permission)
- [ ] `readAttachment()`, `writeAttachment()` (attachments permission)
- [ ] Basic UI services: `notify()` toasts

### 1.4 Plugin Discovery
- [ ] Local plugin directory scanning (`~/.archon/plugins`)
- [ ] Plugin validation and integrity checking
- [ ] Plugin installation from local directory

### 1.5 Basic Plugin Types (Start Small)
- [ ] **Importer**: Most mature concept, builds on ADR-004
- [ ] **Validator**: Simple read-only checks to prove the concept

**Deliverables**: Working sandbox with 2 plugin types, permission system, local installation.

## Phase 2: Extension & Integration

**Goal**: Add remaining plugin types and external service integration.

### 2.1 Complete Plugin Type Coverage
- [ ] **Exporter**: Serialize nodes to external formats
- [ ] **Transformer**: Bulk edit operations
- [ ] **Provider**: External system connectors (Jira foundation)
- [ ] **AttachmentProcessor**: File analysis and metadata extraction
- [ ] **ConflictResolver**: Custom merge strategies

### 2.2 Advanced Host Services
- [ ] `host.fetch()` with network permission and rate limiting
- [ ] Secret management (`secrets.get/set/delete()`)
- [ ] Index services (`indexPut()` for SearchIndexer)

### 2.3 Event System
- [ ] Plugin event subscriptions (onBeforeCommit, onAfterCommit, etc.)
- [ ] Event bus and lifecycle hooks
- [ ] Workflow automation capabilities

### 2.4 Reference Provider Implementation
- [ ] **Jira Provider Plugin**: Demonstrates Provider + Validator + Events
  - JQL-based data pulling
  - OAuth 2.0 / API token authentication
  - Issue creation and updates
  - Webhook integration planning

**Deliverables**: Full plugin type coverage, event system, Jira integration demo.

## Phase 3: UI Integration & Advanced Features

**Goal**: Rich UI integration and advanced platform features.

### 3.1 UI-Focused Plugin Types
- [ ] **Panel**: Custom UI panels with React/Svelte integration
- [ ] **UIContrib**: Commands, menus, keybindings
- [ ] Advanced UI host services: `showPanel()`, `showModal()`, `registerCommand()`

### 3.2 Advanced Plugin Types
- [ ] **SearchIndexer**: Domain-specific tokenization for SQLite FTS
- [ ] Enhanced **ConflictResolver** with UI integration

### 3.3 Plugin Management UI
- [ ] Plugin browser and management interface
- [ ] Plugin configuration UI
- [ ] Permission review and management
- [ ] Plugin debugging and logs

### 3.4 Advanced Installation
- [ ] Plugin installation from URLs
- [ ] Plugin signing and verification
- [ ] Plugin update mechanisms

**Deliverables**: Full UI integration, plugin management interface, advanced security features.

## Implementation Strategy

### Architecture Decisions

1. **Frontend-First**: Plugin system is primarily frontend-focused (Wails frontend)
2. **TypeScript Native**: Strong typing throughout for developer experience  
3. **Modular Host Services**: Each host service is independently gated by permissions
4. **Incremental API Evolution**: Plugin API versioning for backward compatibility

### Development Approach

1. **Start with Importer**: Most mature concept from ADR-004
2. **Build Horizontally**: Complete core infrastructure before adding plugin types
3. **Reference Implementation**: Use real-world plugins (CSV, Jira) to validate APIs
4. **Security-First**: Implement permission system early and test thoroughly

### File Structure

```
frontend/src/plugins/
├── api.ts                 # Core TypeScript interfaces
├── runtime/              
│   ├── sandbox.ts         # Web Worker management
│   ├── permissions.ts     # Permission validation and enforcement
│   ├── lifecycle.ts       # Plugin lifecycle management
│   └── host-services.ts   # Host service implementations
├── types/                 # Plugin type implementations
│   ├── importer.ts
│   ├── validator.ts
│   ├── provider.ts
│   └── ...
├── ui/                    # Plugin management UI
│   ├── PluginBrowser.svelte
│   ├── PermissionDialog.svelte
│   └── ...
└── examples/              # Reference plugin implementations
    ├── csv-importer/
    ├── jira-provider/
    └── pdf-processor/
```

### Testing Strategy

1. **Sandbox Security**: Verify isolation and permission enforcement
2. **Plugin API Stability**: Contract testing for host services
3. **Reference Plugins**: Real-world plugin development and testing
4. **Permission Edge Cases**: Security testing for privilege escalation

### Success Criteria

**Phase 1**: Can install and run a CSV importer plugin with proper sandboxing
**Phase 2**: Can integrate with Jira to pull/push issues with proper authentication  
**Phase 3**: Can extend Archon UI with custom panels and commands

## Timeline Estimation

- **Phase 1**: 3-4 weeks (foundation is critical to get right)
- **Phase 2**: 2-3 weeks (building on established patterns)  
- **Phase 3**: 2-3 weeks (UI integration complexity)

**Total**: ~8-10 weeks for full plugin system implementation

## Risk Mitigation

1. **Security Risk**: Extensive permission testing and code review
2. **API Stability**: Version plugin APIs and maintain compatibility layer
3. **Performance Risk**: Monitor Web Worker overhead and implement resource limits
4. **Complexity Risk**: Start simple and expand incrementally rather than big-bang approach

## Next Steps

1. Begin with Phase 1.1: Core Runtime implementation
2. Create TypeScript interfaces in `frontend/src/plugins/api.ts`
3. Implement Web Worker sandbox environment
4. Build CSV importer as first reference implementation

This phased approach ensures we deliver value incrementally while building a robust, secure, and extensible plugin platform.