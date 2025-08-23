package api

import (
    "context"
    "fmt"
    "path/filepath"

    "github.com/rgehrsitz/archon/internal/errors"
    "github.com/rgehrsitz/archon/internal/migrate"
    "github.com/rgehrsitz/archon/internal/store"
    "github.com/rgehrsitz/archon/internal/types"
)

// MigrationService provides listing, planning, and execution for schema migrations.
type MigrationService struct{}

func NewMigrationService() *MigrationService { return &MigrationService{} }

// ListRegisteredSteps returns all known migration steps in ascending order.
func (s *MigrationService) ListRegisteredSteps(ctx context.Context) ([]migrate.StepDescriptor, errors.Envelope) {
    return migrate.RegisteredSteps(), errors.Envelope{}
}

// Plan returns the steps that would run for the project at path from its current
// schema to the application's CurrentSchemaVersion. No side-effects.
func (s *MigrationService) Plan(ctx context.Context, path string) ([]migrate.StepDescriptor, errors.Envelope) {
    cleanPath, err := filepath.Abs(path)
    if err != nil {
        return nil, errors.WrapError(errors.ErrInvalidPath, "Invalid project path", err)
    }
    ldr := store.NewLoader(cleanPath)
    project, err := ldr.LoadProject()
    if err != nil {
        if env, ok := err.(errors.Envelope); ok {
            return nil, env
        }
        return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to load project", err)
    }

    current := project.SchemaVersion
    target := types.CurrentSchemaVersion

    // Newer project than app: no plan; report read-only scenario.
    if current > target {
        return []migrate.StepDescriptor{}, errors.New(errors.ErrSchemaVersion, "Project schema is newer than this application")
    }
    if current == target {
        return []migrate.StepDescriptor{}, errors.Envelope{}
    }

    plan := make([]migrate.StepDescriptor, 0, target-current)
    for v := current + 1; v <= target; v++ {
        if st, ok := migrate.StepForVersion(v); ok {
            plan = append(plan, migrate.StepDescriptor{Version: v, Name: st.Name()})
        } else {
            // Surface gap as a migration failure in the envelope
            return nil, errors.New(errors.ErrMigrationFailure, fmt.Sprintf("No registered migration step for schema version %d", v))
        }
    }
    return plan, errors.Envelope{}
}

// Execute creates a pre-migration backup and runs all pending steps for the project at path.
func (s *MigrationService) Execute(ctx context.Context, path string) errors.Envelope {
    cleanPath, err := filepath.Abs(path)
    if err != nil {
        return errors.WrapError(errors.ErrInvalidPath, "Invalid project path", err)
    }
    ldr := store.NewLoader(cleanPath)
    project, err := ldr.LoadProject()
    if err != nil {
        if env, ok := err.(errors.Envelope); ok {
            return env
        }
        return errors.WrapError(errors.ErrStorageFailure, "Failed to load project", err)
    }

    current := project.SchemaVersion
    target := types.CurrentSchemaVersion

    if current > target {
        return errors.New(errors.ErrSchemaVersion, "Project schema is newer than this application; cannot migrate")
    }
    if current == target {
        return errors.Envelope{}
    }

    if _, err := migrate.CreateBackup(cleanPath); err != nil {
        return errors.WrapError(errors.ErrMigrationFailure, "Failed to create pre-migration backup", err)
    }
    if err := migrate.Run(cleanPath, current, target); err != nil {
        return errors.WrapError(errors.ErrMigrationFailure, "Migration failed", err)
    }
    // Post-verify
    project, err = ldr.LoadProject()
    if err != nil {
        return errors.WrapError(errors.ErrStorageFailure, "Failed to reload project post-migration", err)
    }
    if project.SchemaVersion != target {
        return errors.New(errors.ErrMigrationFailure, "Post-migration schema version mismatch")
    }

    return errors.Envelope{}
}
