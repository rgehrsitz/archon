package migrate

import (
    "fmt"
    "sort"

    "github.com/rgehrsitz/archon/internal/store"
    "github.com/rgehrsitz/archon/internal/types"
)

// Context carries project state for a migration run.
type Context struct {
    BasePath string
    Loader   *store.Loader
    Project  *types.Project
}

// Step defines a forward-only, idempotent migration for a specific target version.
// Contract: After Apply succeeds, project.SchemaVersion must equal Version().
type Step interface {
    Version() int
    Name() string
    IsApplied(*Context) (bool, error)
    Apply(*Context) error
}

var registry = map[int]Step{}

// Register adds a step to the registry, keyed by its Version().
func Register(step Step) {
    registry[step.Version()] = step
}

// StepDescriptor is a serializable view of a registered migration step.
type StepDescriptor struct {
    Version int    `json:"version"`
    Name    string `json:"name"`
}

// RegisteredSteps returns a snapshot of known steps sorted by version.
func RegisteredSteps() []StepDescriptor {
    versions := make([]int, 0, len(registry))
    for v := range registry {
        versions = append(versions, v)
    }
    sort.Ints(versions)
    out := make([]StepDescriptor, 0, len(versions))
    for _, v := range versions {
        out = append(out, StepDescriptor{Version: v, Name: registry[v].Name()})
    }
    return out
}

// StepForVersion returns the step and a bool indicating if it exists.
func StepForVersion(v int) (Step, bool) {
    st, ok := registry[v]
    return st, ok
}

// Run executes all steps from (current, target], in ascending version order.
// It loads the project before each step, skips already-applied steps, and
// requires each applied step to bump SchemaVersion accordingly. If any
// version in the range is missing a registered step, an error is returned.
func Run(basePath string, current, target int) error {
    if target <= current {
        return nil
    }
    loader := store.NewLoader(basePath)
    // Execute versions in order and ensure coverage
    versions := make([]int, 0, target-current)
    for v := current + 1; v <= target; v++ {
        versions = append(versions, v)
    }
    sort.Ints(versions)

    for _, v := range versions {
        step, ok := registry[v]
        if !ok {
            return fmt.Errorf("no registered migration step for schema version %d", v)
        }
        // Reload project each iteration for freshness
        proj, err := loader.LoadProject()
        if err != nil {
            return err
        }
        ctx := &Context{BasePath: basePath, Loader: loader, Project: proj}

        applied, err := step.IsApplied(ctx)
        if err != nil {
            return err
        }
        if applied {
            continue
        }
        if err := step.Apply(ctx); err != nil {
            return err
        }
        // Validate version bump
        proj, err = loader.LoadProject()
        if err != nil {
            return err
        }
        if proj.SchemaVersion != v {
            return fmt.Errorf("migration %d (%s) did not set schemaVersion to %d", v, step.Name(), v)
        }
    }
    return nil
}
