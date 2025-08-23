package api

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rgehrsitz/archon/internal/logging"
)

type logLine struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func writeLinesJSON(path string, messages []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, m := range messages {
		if err := enc.Encode(logLine{Level: "info", Message: m}); err != nil {
			return err
		}
	}
	return nil
}

func writeLinesJSONGzip(path string, messages []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	enc := json.NewEncoder(gz)
	for _, m := range messages {
		if err := enc.Encode(logLine{Level: "info", Message: m}); err != nil {
			return err
		}
	}
	return nil
}

func TestGetRecentLogs_RotationAware(t *testing.T) {
	dir := t.TempDir()

	// Ensure logger config points at our temp dir without touching outputs
	logging.GetLogger().GetConfig().LogDirectory = dir

	// Create files: current, rotated, rotated gz
	current := filepath.Join(dir, "archon.log")
	older1 := filepath.Join(dir, "archon.log.1")
	older2gz := filepath.Join(dir, "archon.log.2.gz")

	if err := writeLinesJSON(older1, []string{"O1-A", "O1-B"}); err != nil {
		t.Fatalf("write older1: %v", err)
	}
	if err := writeLinesJSONGzip(older2gz, []string{"O2-A", "O2-B"}); err != nil {
		t.Fatalf("write older2.gz: %v", err)
	}
	if err := writeLinesJSON(current, []string{"C-A", "C-B", "C-C"}); err != nil {
		t.Fatalf("write current: %v", err)
	}

	// Ensure deterministic mod times: current newest, older1 next, older2.gz oldest
	now := time.Now()
	if err := os.Chtimes(older2gz, now.Add(-2*time.Hour), now.Add(-2*time.Hour)); err != nil {
		t.Fatalf("chtimes older2.gz: %v", err)
	}
	if err := os.Chtimes(older1, now.Add(-1*time.Hour), now.Add(-1*time.Hour)); err != nil {
		t.Fatalf("chtimes older1: %v", err)
	}
	if err := os.Chtimes(current, now, now); err != nil {
		t.Fatalf("chtimes current: %v", err)
	}

	svc := &LoggingService{}
	logs, env := svc.GetRecentLogs(context.Background(), 4)
	if env.Code != "" {
		t.Fatalf("GetRecentLogs error: %s - %s", env.Code, env.Message)
	}

	if len(logs) != 4 {
		t.Fatalf("got %d logs, want 4", len(logs))
	}
	// Expect last 4 chronologically across files: from older to newer overall
	// Given processing order (current, older1, older2.gz) collecting from end, then reversed
	// The last 4 overall should be: O1-B, C-A, C-B, C-C
	expect := []string{"O1-B", "C-A", "C-B", "C-C"}
	for i, e := range expect {
		m := logs[i]
		msg, _ := m["message"].(string)
		if msg != e {
			t.Fatalf("log[%d]=%q, want %q", i, msg, e)
		}
	}
}
