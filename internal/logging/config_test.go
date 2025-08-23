package logging

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type envOnDisk struct {
	Environment string `json:"environment"`
	LogLevel    string `json:"log_level"`
	Console     bool   `json:"console"`
	File        bool   `json:"file"`
	MaxSize     int    `json:"max_size_mb"`
	MaxBackups  int    `json:"max_backups"`
	MaxAge      int    `json:"max_age_days"`
	Compress    bool   `json:"compress"`
	Directory   string `json:"directory"`
}

func TestUpdateEnvironmentConfig_NumericAndAliases(t *testing.T) {
	root := t.TempDir()
	env := "development"

	updates := map[string]interface{}{
		"max_size_mb": 12.0,              // float64 should be accepted
		"max_backups": 7.0,               // float64 should be rounded
		"max_age_days": 14.2,             // float64 rounded to 14
		"directory":    "logs-custom",    // relative path resolved
		"console":      true,
		"file":         false,
		"log_level":    "error",
		"compress":     true,
	}

	if err := UpdateEnvironmentConfig(root, env, updates); err != nil {
		t.Fatalf("UpdateEnvironmentConfig error: %v", err)
	}

	// Verify file contents on disk
	cfgPath := filepath.Join(root, ".archon", "logging.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var got envOnDisk
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.LogLevel != "error" {
		t.Errorf("LogLevel = %s, want error", got.LogLevel)
	}
	if !got.Console {
		t.Errorf("Console = false, want true")
	}
	if got.File {
		t.Errorf("File = true, want false")
	}
	if got.MaxSize != 12 {
		t.Errorf("MaxSize = %d, want 12", got.MaxSize)
	}
	if got.MaxBackups != 7 {
		t.Errorf("MaxBackups = %d, want 7", got.MaxBackups)
	}
	if got.MaxAge != 14 {
		t.Errorf("MaxAge = %d, want 14", got.MaxAge)
	}

	expectedDir := filepath.Join(root, "logs-custom")
	if got.Directory != expectedDir {
		t.Errorf("Directory = %s, want %s", got.Directory, expectedDir)
	}
}
