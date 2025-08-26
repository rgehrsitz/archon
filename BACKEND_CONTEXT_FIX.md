# Backend Context.Context Fix Required

## Issue
The Archon backend Go services are incorrectly exposing `context.Context` as frontend parameters, which causes Wails v2 binding errors:
```
error parsing arguments: json: cannot unmarshal object into Go value of type context.Context
```

## Root Cause
Based on Wails v2 documentation and GitHub issues (#3948, #1766), the backend service methods should NOT include `context.Context` as parameters exposed to the frontend.

## Current Problematic Pattern
```go
// ❌ INCORRECT - This exposes context to frontend
func (s *ProjectService) IsProjectOpen(ctx context.Context) bool {
    // implementation
}

func (s *ProjectService) CreateProject(ctx context.Context, path string, settings map[string]any) (*Project, error) {
    // implementation  
}
```

## Correct Wails v2 Pattern
```go
type ProjectService struct {
    ctx context.Context  // Store context internally
}

// Set context during startup
func (s *ProjectService) SetContext(ctx context.Context) {
    s.ctx = ctx
}

// ✅ CORRECT - Frontend methods without context
func (s *ProjectService) IsProjectOpen() bool {
    return s.isProjectOpenWithContext(s.ctx)
}

func (s *ProjectService) CreateProject(path string, settings map[string]any) (*Project, error) {
    return s.createProjectWithContext(s.ctx, path, settings)
}

// Private methods that use context internally
func (s *ProjectService) isProjectOpenWithContext(ctx context.Context) bool {
    // implementation with context
}

func (s *ProjectService) createProjectWithContext(ctx context.Context, path string, settings map[string]any) (*Project, error) {
    // implementation with context
}
```

## Files That Need Updates
1. `internal/api/project.go` (or wherever ProjectService is defined)
2. `internal/api/node.go` (NodeService) 
3. `internal/api/logging.go` (LoggingService)
4. Any other service structs exposed to frontend

## How to Fix
1. **Add context field** to service structs
2. **Add SetContext method** to each service
3. **Remove context.Context parameters** from public methods
4. **Create private methods** that use context internally
5. **Update main.go OnStartup** to call SetContext on all services

## After Fix
- Regenerate Wails bindings with `wails build` or `wails dev`
- Frontend API calls will work without context errors
- All TODO comments in frontend can be removed
- Real backend integration will be complete

## References
- Wails v2 Application Development: https://wails.io/docs/guides/application-development
- GitHub Issue #3948: Generated bindings include context.Context
- GitHub Issue #1766: Missing context .d.ts type generation