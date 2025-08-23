package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	config := &Config{
		Level:           LevelInfo,
		OutputConsole:   false,
		OutputFile:      false,
		LogDirectory:    "",
		MaxFileSize:     10,
		MaxBackups:      5,
		MaxAge:          30,
		CompressBackups: true,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	if logger.config.Level != LevelInfo {
		t.Errorf("Expected log level %s, got %s", LevelInfo, logger.config.Level)
	}
}

func TestLoggerWithFileOutput(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon-logging-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &Config{
		Level:           LevelDebug,
		OutputConsole:   false,
		OutputFile:      true,
		LogDirectory:    tempDir,
		MaxFileSize:     1, // 1MB for testing
		MaxBackups:      3,
		MaxAge:          1,
		CompressBackups: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test logging at different levels
	logger.Debug().Msg("Debug message")
	logger.Info().Msg("Info message")
	logger.Warn().Msg("Warning message")
	logger.Error().Msg("Error message")

	// Give logger time to write
	time.Sleep(100 * time.Millisecond)

	// Check if log file was created
	logFile := filepath.Join(tempDir, "archon.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Read log file and verify content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	expectedMessages := []string{"Debug message", "Info message", "Warning message", "Error message"}
	
	for _, msg := range expectedMessages {
		if !strings.Contains(logContent, msg) {
			t.Errorf("Expected log file to contain '%s', but it didn't", msg)
		}
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		configLevel LogLevel
		logLevel    string
		shouldLog   bool
	}{
		{LevelError, "debug", false},
		{LevelError, "info", false},
		{LevelError, "warn", false},
		{LevelError, "error", true},
		{LevelInfo, "debug", false},
		{LevelInfo, "info", true},
		{LevelInfo, "warn", true},
		{LevelInfo, "error", true},
		{LevelDebug, "debug", true},
		{LevelDebug, "info", true},
		{LevelDebug, "warn", true},
		{LevelDebug, "error", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.configLevel)+"_"+tt.logLevel, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "archon-level-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			config := &Config{
				Level:           tt.configLevel,
				OutputConsole:   false,
				OutputFile:      true,
				LogDirectory:    tempDir,
				MaxFileSize:     10,
				MaxBackups:      3,
				MaxAge:          1,
				CompressBackups: false,
			}

			logger, err := NewLogger(config)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			testMessage := "Test message for " + tt.logLevel
			
			// Log message at the specified level
			switch tt.logLevel {
			case "debug":
				logger.Debug().Msg(testMessage)
			case "info":
				logger.Info().Msg(testMessage)
			case "warn":
				logger.Warn().Msg(testMessage)
			case "error":
				logger.Error().Msg(testMessage)
			}

			time.Sleep(100 * time.Millisecond)

			// Check if message was logged
			logFile := filepath.Join(tempDir, "archon.log")
			content, err := os.ReadFile(logFile)
			
			if tt.shouldLog {
				if err != nil {
					t.Fatalf("Expected log file to exist but got error: %v", err)
				}
				if !strings.Contains(string(content), testMessage) {
					t.Errorf("Expected log file to contain '%s'", testMessage)
				}
			} else {
				if err == nil && strings.Contains(string(content), testMessage) {
					t.Errorf("Expected log file NOT to contain '%s'", testMessage)
				}
			}
		})
	}
}

func TestLoggerWithContext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archon-context-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &Config{
		Level:        LevelInfo,
		OutputConsole: false,
		OutputFile:   true,
		LogDirectory: tempDir,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test context fields
	contextLogger := logger.WithContext(map[string]interface{}{
		"user_id": "12345",
		"action":  "test_operation",
	})

	contextLogger.Info().Msg("Context test message")

	time.Sleep(100 * time.Millisecond)

	// Verify context fields appear in log
	logFile := filepath.Join(tempDir, "archon.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "user_id") || !strings.Contains(logContent, "12345") {
		t.Error("Expected log to contain user_id context field")
	}
	if !strings.Contains(logContent, "action") || !strings.Contains(logContent, "test_operation") {
		t.Error("Expected log to contain action context field")
	}
}

func TestUpdateLevel(t *testing.T) {
	config := &Config{
		Level:        LevelInfo,
		OutputConsole: false,
		OutputFile:   false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Initial level should be Info
	if logger.config.Level != LevelInfo {
		t.Errorf("Expected initial level %s, got %s", LevelInfo, logger.config.Level)
	}

	// Update to Debug
	logger.UpdateLevel(LevelDebug)
	if logger.config.Level != LevelDebug {
		t.Errorf("Expected updated level %s, got %s", LevelDebug, logger.config.Level)
	}

	// Update to Error
	logger.UpdateLevel(LevelError)
	if logger.config.Level != LevelError {
		t.Errorf("Expected updated level %s, got %s", LevelError, logger.config.Level)
	}
}