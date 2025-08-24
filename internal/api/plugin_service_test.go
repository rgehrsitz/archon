package api

import (
	"context"
	"testing"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/logging"
)

func TestPluginServiceReadOnlyGuards(t *testing.T) {
	// Quiet logger for tests
	cfg := logging.DefaultConfig()
	cfg.OutputConsole = false
	cfg.OutputFile = false
	logger, _ := logging.NewLogger(cfg)

	ps := &ProjectService{}
	ps.readOnly = true

	svc := NewPluginService(*logger, ps)

	// ApplyMutations should be blocked in read-only mode
	if err := svc.PluginApplyMutations(context.Background(), "test.plugin", nil); err.Code != errors.ErrSchemaVersion {
		t.Fatalf("expected ErrSchemaVersion for ApplyMutations, got: %+v", err)
	}

	// IndexPut should be blocked in read-only mode
	if err := svc.PluginIndexPut(context.Background(), "test.plugin", "node-1", "content"); err.Code != errors.ErrSchemaVersion {
		t.Fatalf("expected ErrSchemaVersion for IndexPut, got: %+v", err)
	}
}
