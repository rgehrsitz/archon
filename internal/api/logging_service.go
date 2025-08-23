package api

import (
	"context"
	"fmt"

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
	
	logging.SetLevel(logLevel)
	
	// Log the change
	logging.Info().Info().
		Str("old_level", string(logging.GetLevel())).
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
// Note: This would require implementing log reading functionality
func (s *LoggingService) GetRecentLogs(ctx context.Context, limit int) ([]map[string]interface{}, errors.Envelope) {
	// TODO: Implement log reading from files
	// For now, return empty array with a note
	return []map[string]interface{}{
		{
			"level":     "info",
			"message":   "Log reading not yet implemented",
			"timestamp": "2024-08-23T00:00:00Z",
			"source":    "logging_service",
		},
	}, errors.Envelope{}
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
	logging.Info().Info().
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
	
	logging.Info().Info().
		Str("project_path", projectPath).
		Str("environment", logging.GetEnvironmentFromOS()).
		Msg("Project logging initialized")
	
	return errors.Envelope{}
}