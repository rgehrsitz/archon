# Backend Host Services Implementation Plan

## Overview
Implement the Go backend services that support the frontend plugin system, bridging TypeScript plugin interfaces to actual Archon functionality.

## Status (2025-08-24)

- __Just completed__: Fixed backend compilation for plugin system
  - Standardized zerolog usage across plugin backend (`logger.Info().Msg(...)`)
  - Corrected `ProjectService.GetCurrentProject()` usage and plugin dir path in `internal/api/plugin_service.go`
  - Replaced nonexistent `logging.Global()` with `*logging.GetLogger()` in `app.go`
  - Aligned index manager API usage in host/manager with `internal/index/index.go`
- __In progress__: ADR-013 alignment for host services
  - Wails bindings in `internal/api/plugin_service.go`
  - Permission enforcement in `internal/plugins/permissions.go`
  - Secrets and network proxy scaffolding

## Phase 2.1: Core Backend Services

### 1. Plugin Host Service (Go)
**Location**: `internal/api/plugin_service.go`
- New Wails service for plugin operations
- Bridge between frontend plugin runtime and Go backend
- Permission enforcement at the Go level

### 2. Repository Operations Bridge
**Location**: `internal/api/plugin_service.go` methods
- `GetNode(id string) (*types.Node, error)`
- `ListChildren(parentId string) ([]string, error)` 
- `ApplyMutations(mutations []PluginMutation) error`
- `CreateCommit(message string) (string, error)`
- Uses existing `store` package for actual operations

### 3. Plugin Permission Enforcement
**Location**: `internal/plugins/permissions.go`
- Go-side permission validation
- Plugin manifest loading and verification
- Permission grant storage and checking

### 4. Plugin Manager Backend
**Location**: `internal/plugins/manager.go`
- Plugin discovery and loading
- Installation metadata persistence  
- Plugin lifecycle management (enable/disable)
- Integration with existing project structure

### 5. Network Proxy Service
**Location**: `internal/plugins/network.go`
- Scoped HTTP client for plugin network requests
- Rate limiting and security controls
- URL allowlist enforcement

### 6. Secrets Management
**Location**: `internal/plugins/secrets.go`
- Encrypted storage for plugin authentication tokens
- Scoped access based on plugin permissions
- Integration with OS keychain where available

## Integration Points

### Existing Systems
- **Node Store** (`internal/store`): Repository operations
- **Git Service** (`internal/git`): Commit and snapshot operations  
- **Index** (`internal/index`): Search operations for plugins
- **Project** (`internal/project`): Plugin storage within .archon/

### New Wails Bindings
```go
//go:build wails
type PluginService struct {
    store    *store.Store
    git      *git.Service  
    index    *index.Index
    manager  *plugins.Manager
}

func (ps *PluginService) GetNode(ctx context.Context, id string) (*Node, error)
func (ps *PluginService) ApplyMutations(ctx context.Context, mutations []Mutation) error
// ... other methods
```

## File Structure
```
internal/plugins/
├── manager.go          # Plugin lifecycle management
├── permissions.go      # Permission enforcement  
├── host.go            # Host service implementations
├── network.go         # HTTP proxy for plugins
├── secrets.go         # Secrets management
└── types.go           # Go types matching TypeScript interfaces

internal/api/
├── plugin_service.go   # Wails service bindings
```

## Implementation Order
1. **Plugin Types & Permissions** - Go structs matching TypeScript interfaces
2. **Basic Host Service** - Repository operations (GetNode, ApplyMutations)
3. **Plugin Manager** - Discovery and loading
4. **Network & Secrets** - External service support
5. **Wails Integration** - Service bindings and frontend bridge

## Benefits
- Plugins become actually functional with real data
- Validates our TypeScript interfaces against implementation
- Enables building real plugins (not just examples)
- Sets up foundation for advanced plugin types

## Testing Strategy
- Unit tests for each service component
- Integration tests with actual plugin execution
- End-to-end test with CSV importer against real repositories