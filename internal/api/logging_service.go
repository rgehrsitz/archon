package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"compress/gzip"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/logging"
)

// LoggingService provides Wails-bound logging operations
type LoggingService struct {
	projectService *ProjectService
}

// NewLoggingService creates a new logging service
func NewLoggingService(projectService *ProjectService) *LoggingService {
	return &LoggingService{
		projectService: projectService,
	}
}

// LogEntry represents a frontend log entry
type LogEntry struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
}

// LoggingConfig represents the current logging configuration
type LoggingConfig struct {
	Level           string `json:"level"`
	OutputConsole   bool   `json:"outputConsole"`
	OutputFile      bool   `json:"outputFile"`
	LogDirectory    string `json:"logDirectory"`
	MaxFileSize     int    `json:"maxFileSize"`
	MaxBackups      int    `json:"maxBackups"`
	MaxAge          int    `json:"maxAge"`
	CompressBackups bool   `json:"compressBackups"`
}

// LogMessage logs a message from the frontend
func (s *LoggingService) LogMessage(ctx context.Context, entry LogEntry) errors.Envelope {
	logger := logging.GetLogger()
	
	// Add frontend context
	if entry.Context != nil {
		logger = logger.WithContext(entry.Context)
	}
	
	// Add frontend marker
	logger = logger.WithContext(map[string]interface{}{
		"source": "frontend",
	})
	
	// Log based on level
	switch entry.Level {
	case "trace":
		logger.Trace().Msg(entry.Message)
	case "debug":
		logger.Debug().Msg(entry.Message)
	case "info":
		logger.Info().Msg(entry.Message)
	case "warn", "warning":
		logger.Warn().Msg(entry.Message)
	case "error":
		logger.Error().Msg(entry.Message)
	case "fatal":
		logger.Fatal().Msg(entry.Message)
	default:
		logger.Info().Msg(entry.Message)
	}
	
	return errors.Envelope{}
}

// GetLoggingConfig returns the current logging configuration
func (s *LoggingService) GetLoggingConfig(ctx context.Context) (*LoggingConfig, errors.Envelope) {
	logger := logging.GetLogger()
	config := logger.GetConfig()
	
	return &LoggingConfig{
		Level:           string(config.Level),
		OutputConsole:   config.OutputConsole,
		OutputFile:      config.OutputFile,
		LogDirectory:    config.LogDirectory,
		MaxFileSize:     config.MaxFileSize,
		MaxBackups:      config.MaxBackups,
		MaxAge:          config.MaxAge,
		CompressBackups: config.CompressBackups,
	}, errors.Envelope{}
}

// SetLogLevel updates the current logging level
func (s *LoggingService) SetLogLevel(ctx context.Context, level string) errors.Envelope {
	var logLevel logging.LogLevel
	switch level {
	case "trace":
		logLevel = logging.LevelTrace
	case "debug":
		logLevel = logging.LevelDebug
	case "info":
		logLevel = logging.LevelInfo
	case "warn", "warning":
		logLevel = logging.LevelWarn
	case "error":
		logLevel = logging.LevelError
	case "fatal":
		logLevel = logging.LevelFatal
	default:
		return errors.New(errors.ErrInvalidInput, fmt.Sprintf("Invalid log level: %s", level))
	}
	
	// Capture old level, then update level
	old := logging.GetLevel()
	logging.SetLevel(logLevel)
	
	// Log the change
	logging.Log().Info().
		Str("old_level", string(old)).
		Str("new_level", level).
		Msg("Log level changed from frontend")
	
	return errors.Envelope{}
}

// GetLogHealth returns health information about the logging system
func (s *LoggingService) GetLogHealth(ctx context.Context) (map[string]interface{}, errors.Envelope) {
	health := logging.Health()
	return health, errors.Envelope{}
}

// GetRecentLogs returns recent log entries (if available)
func (s *LoggingService) GetRecentLogs(ctx context.Context, limit int) ([]map[string]interface{}, errors.Envelope) {
	if limit <= 0 {
		limit = 100
	}

	cfg := logging.GetLogger().GetConfig()
	dir := cfg.LogDirectory

	// Collect candidate files: archon.log and rotated variants (including compressed)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]interface{}{}, errors.Envelope{}
		}
		return nil, errors.WrapError(errors.ErrStorageFailure, "Failed to list log directory", err)
	}

	type fileInfo struct {
		path string
		mod  int64
		isCurrent bool
	}

	candidates := make([]fileInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "archon.log") {
			continue
		}
		full := filepath.Join(dir, name)
		info, statErr := e.Info()
		if statErr != nil {
			continue
		}
		candidates = append(candidates, fileInfo{path: full, mod: info.ModTime().UnixNano(), isCurrent: name == "archon.log"})
	}

	if len(candidates) == 0 {
		return []map[string]interface{}{}, errors.Envelope{}
	}

	// Sort by current first, then by modtime desc
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].isCurrent != candidates[j].isCurrent {
			return candidates[i].isCurrent
		}
		return candidates[i].mod > candidates[j].mod
	})

	// Helper to read file content (supports .gz)
	readAll := func(p string) ([]byte, error) {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		var r io.Reader = f
		if strings.HasSuffix(p, ".gz") {
			gz, gzErr := gzip.NewReader(f)
			if gzErr != nil {
				return nil, gzErr
			}
			defer gz.Close()
			r = gz
		}
		return io.ReadAll(r)
	}

	// Collect last N lines across files, starting from newest
	rev := make([]map[string]interface{}, 0, limit)
	total := 0
	for _, fi := range candidates {
		if total >= limit {
			break
		}
		data, rerr := readAll(fi.path)
		if rerr != nil {
			// Skip unreadable files
			continue
		}
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		// Walk from end to start to get most recent first from this file
		for i := len(lines) - 1; i >= 0 && total < limit; i-- {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(line), &m); err == nil {
				rev = append(rev, m)
			} else {
				rev = append(rev, map[string]interface{}{"raw": lines[i]})
			}
			total++
		}
	}

	// Reverse to chronological order
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}

	return rev, errors.Envelope{}
}

// UpdateLoggingConfig updates the logging configuration
func (s *LoggingService) UpdateLoggingConfig(ctx context.Context, updates map[string]interface{}) errors.Envelope {
	if s.projectService.currentProject == nil {
		return errors.New(errors.ErrNoProject, "No project is currently open")
	}
	
	_, currentPath := s.projectService.GetCurrentProject()
	if currentPath == "" {
		return errors.New(errors.ErrNoProject, "No current project path available")
	}
	
	// Get current environment (could be enhanced to detect from project settings)
	environment := logging.GetEnvironmentFromOS()
	
	err := logging.UpdateEnvironmentConfig(currentPath, environment, updates)
	if err != nil {
		return errors.WrapError(errors.ErrStorageFailure, "Failed to update logging configuration", err)
	}
	
	// Log the configuration change
	logging.Log().Info().
		Interface("updates", updates).
		Str("environment", environment).
		Msg("Logging configuration updated from frontend")
	
	return errors.Envelope{}
}

// GetLogLevels returns available log levels
func (s *LoggingService) GetLogLevels(ctx context.Context) ([]string, errors.Envelope) {
	levels := []string{
		"trace",
		"debug", 
		"info",
		"warn",
		"error",
		"fatal",
	}
	
	return levels, errors.Envelope{}
}

// InitializeProjectLogging initializes logging for a project
func (s *LoggingService) InitializeProjectLogging(ctx context.Context, projectPath string) errors.Envelope {
	err := logging.InitializeFromEnvironment(projectPath)
	if err != nil {
		return errors.WrapError(errors.ErrStorageFailure, "Failed to initialize project logging", err)
	}
	
	logging.Log().Info().
		Str("project_path", projectPath).
		Str("environment", logging.GetEnvironmentFromOS()).
		Msg("Project logging initialized")
	
	return errors.Envelope{}
}