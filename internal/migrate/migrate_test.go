package migrate

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/store"
	"github.com/rgehrsitz/archon/internal/types"
)

func createTestProject(t *testing.T, base string, schema int) {
	t.Helper()
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ldr := store.NewLoader(base)
	p := &types.Project{
		RootID:        "00000000-0000-0000-0000-000000000000",
		SchemaVersion: schema,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := ldr.SaveProject(p); err != nil {
		t.Fatalf("save project: %v", err)
	}
}

func TestRun_MissingStepReturnsError(t *testing.T) {
	dir := t.TempDir()
	createTestProject(t, dir, 0)
	// Target a far-future version to guarantee a missing step
	err := Run(dir, 0, 1000)
	if err == nil || !strings.Contains(err.Error(), "no registered migration step") {
		t.Fatalf("expected missing step error, got: %v", err)
	}
}

func TestRun_V1_MigratesTo1(t *testing.T) {
	dir := t.TempDir()
	createTestProject(t, dir, 0)
	if err := Run(dir, 0, 1); err != nil {
		t.Fatalf("run migrate 0->1: %v", err)
	}
	ldr := store.NewLoader(dir)
	p, err := ldr.LoadProject()
	if err != nil {
		t.Fatalf("reload project: %v", err)
	}
	if p.SchemaVersion != 1 {
		t.Fatalf("expected schema 1, got %d", p.SchemaVersion)
	}
}

func TestRun_Idempotent(t *testing.T) {
	dir := t.TempDir()
	createTestProject(t, dir, 0)
	if err := Run(dir, 0, 1); err != nil {
		t.Fatalf("first run: %v", err)
	}
	// Second run should be a no-op
	if err := Run(dir, 0, 1); err != nil {
		t.Fatalf("second run (idempotent) failed: %v", err)
	}
}

type badStep struct{}

func (b *badStep) Version() int                { return 2 }
func (b *badStep) Name() string                { return "Bad step (no bump)" }
func (b *badStep) IsApplied(*Context) (bool, error) { return false, nil }
func (b *badStep) Apply(ctx *Context) error    { return nil /* intentionally forget to bump */ }

func TestRun_InvalidBumpError(t *testing.T) {
	// Register a bad step for version 2
	Register(&badStep{})
	dir := t.TempDir()
	createTestProject(t, dir, 1)
	err := Run(dir, 1, 2)
	if err == nil || !strings.Contains(err.Error(), "did not set schemaVersion") {
		t.Fatalf("expected bump validation error, got: %v", err)
	}
}
