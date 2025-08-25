package plugins

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"runtime"
	"sync"
	"testing"
)

func writeSecretsJSON(t *testing.T, dir string, data map[string]any) string {
	 t.Helper()
	 secretsDir := filepath.Join(dir, ".archon")
	 if err := os.MkdirAll(secretsDir, 0o755); err != nil {
	 	 t.Fatalf("mkdir .archon: %v", err)
	 }
	 path := filepath.Join(secretsDir, "secrets.json")
	 b, err := json.Marshal(data)
	 if err != nil { t.Fatalf("marshal: %v", err) }
	 if err := os.WriteFile(path, b, 0o644); err != nil {
	 	 t.Fatalf("write secrets.json: %v", err)
	 }
	 return path
}

func TestFileSecretsStore_LoadMissingFile(t *testing.T) {
	 base := t.TempDir()
	 // Do not create secrets.json
	 path := filepath.Join(base, ".archon", "secrets.json")

	 s, err := NewFileSecretsStore(path)
	 if err != nil {
	 	 t.Fatalf("NewFileSecretsStore error: %v", err)
	 }

	 ctx := context.Background()
	 if v, ok, err := s.Get(ctx, "nope"); err != nil || ok || v != nil {
	 	 t.Fatalf("expected not found nil,nil; got v=%v ok=%v err=%v", v, ok, err)
	 }

	 keys, err := s.List(ctx, "")
	 if err != nil {
	 	 t.Fatalf("List error: %v", err)
	 }
	 if len(keys) != 0 {
	 	 t.Fatalf("expected 0 keys, got %d", len(keys))
	 }
}

func TestFileSecretsStore_LoadJSONAndGetList(t *testing.T) {
	 base := t.TempDir()
	 writeSecretsJSON(t, base, map[string]any{
	 	 "jira.token": map[string]any{"value": "abc123", "redacted": false, "metadata": map[string]any{"svc": "jira"}},
	 	 "github.token": map[string]any{"value": "zzz"},
	 })

	 s, err := NewFileSecretsStore(filepath.Join(base, ".archon", "secrets.json"))
	 if err != nil { t.Fatalf("NewFileSecretsStore: %v", err) }

	 ctx := context.Background()
	 v, ok, err := s.Get(ctx, "jira.token")
	 if err != nil || !ok || v == nil { t.Fatalf("Get jira.token failed: v=%v ok=%v err=%v", v, ok, err) }
	 if v.Name != "jira.token" || v.Value != "abc123" || v.Redacted {
	 	 t.Fatalf("unexpected secret: %#v", v)
	 }
	 if v.Metadata == nil || v.Metadata["svc"] != "jira" {
	 	 t.Fatalf("expected metadata svc=jira, got %#v", v.Metadata)
	 }

	 // Default Name from key when empty (github.token has no explicit name)
	 v2, ok, err := s.Get(ctx, "github.token")
	 if err != nil || !ok || v2 == nil { t.Fatalf("Get github.token failed: %v %v %v", v2, ok, err) }
	 if v2.Name != "github.token" || v2.Value != "zzz" { t.Fatalf("unexpected v2: %#v", v2) }

	 // List by prefix
	 keys, err := s.List(ctx, "jira")
	 if err != nil { t.Fatalf("List: %v", err) }
	 found := map[string]bool{}
	 for _, k := range keys { found[k] = true }
	 if !found["jira.token"] || found["github.token"] {
	 	 t.Fatalf("prefix filter failed, keys=%v", keys)
	 }
}

func TestFileSecretsStore_InvalidJSON(t *testing.T) {
	 base := t.TempDir()
	 secretsDir := filepath.Join(base, ".archon")
	 if err := os.MkdirAll(secretsDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
	 path := filepath.Join(secretsDir, "secrets.json")
	 if err := os.WriteFile(path, []byte("not-json"), 0o644); err != nil { t.Fatalf("write: %v", err) }

	 if _, err := NewFileSecretsStore(path); err == nil {
	 	 t.Fatalf("expected error for invalid JSON")
	 }
}

func TestFileSecretsStore_ConcurrencySafe(t *testing.T) {
	 base := t.TempDir()
	 data := map[string]any{}
	 for i := 0; i < 200; i++ {
	 	 key := "k" + strconv.Itoa(i)
	 	 data[key] = map[string]any{"value": key}
	 }
	 writeSecretsJSON(t, base, data)

	 s, err := NewFileSecretsStore(filepath.Join(base, ".archon", "secrets.json"))
	 if err != nil { t.Fatalf("NewFileSecretsStore: %v", err) }

	 ctx := context.Background()
	 var wg sync.WaitGroup
	 workers := runtime.NumCPU() * 2
	 for w := 0; w < workers; w++ {
	 	 wg.Add(1)
	 	 go func(id int) {
	 	 	 defer wg.Done()
	 	 	 for i := 0; i < 1000; i++ {
	 	 	 	 // alternate gets and lists
	 	 	 	 if i%2 == 0 {
	 	 	 	 	 _, _, _ = s.Get(ctx, "k"+strconv.Itoa(i%200))
	 	 	 	 } else {
	 	 	 	 	 _, _ = s.List(ctx, "k1")
	 	 	 	 }
	 	 	 }
	 	 }(w)
	 }
	 wg.Wait()
}
