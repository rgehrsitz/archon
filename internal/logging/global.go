package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	globalLogger *Logger
	loggerMutex  sync.RWMutex
)

// Initialize sets up the global logger with configuration
func Initialize(config *Config) error {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	
	logger, err := NewLogger(config)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	
	globalLogger = logger
	return nil
}

// InitializeDefault initializes the global logger with default configuration
func InitializeDefault(projectRoot string) error {
	config := DefaultConfig()
	config.LogDirectory = filepath.Join(projectRoot, "logs")
	return Initialize(config)
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	
	if globalLogger == nil {
		// Fallback: create a basic logger if not initialized
		config := DefaultConfig()
		config.LogDirectory = "logs"
		logger, err := NewLogger(config)
		if err != nil {
			// Last resort: return the basic logger we tried to create
			return logger
		}
		globalLogger = logger
	}
	
	return globalLogger
}

// SetLevel updates the global logger level
func SetLevel(level LogLevel) {
	logger := GetLogger()
	logger.UpdateLevel(level)
}

// GetLevel returns the current global logger level
func GetLevel() LogLevel {
	logger := GetLogger()
	return logger.GetConfig().Level
}

// WithContext creates a logger with additional context
func WithContext(fields map[string]interface{}) *Logger {
	return GetLogger().WithContext(fields)
}

// WithRequestID creates a logger with a request ID
func WithRequestID(requestID string) *Logger {
	return GetLogger().WithRequestID(requestID)
}

// WithOperation creates a logger with an operation name
func WithOperation(operation string) *Logger {
	return GetLogger().WithOperation(operation)
}

// WithError creates a logger with an error context
func WithError(err error) *Logger {
	return GetLogger().WithError(err)
}

// Global logging convenience functions

// Trace logs at trace level
func Trace() *Logger {
	return GetLogger()
}

// Debug logs at debug level  
func Debug() *Logger {
	return GetLogger()
}

// Info logs at info level
func Info() *Logger {
    return GetLogger()
}

// Warn logs at warn level
func Warn() *Logger {
	return GetLogger()
}

// Error logs at error level
func Error() *Logger {
	return GetLogger()
}

// Fatal logs at fatal level
func Fatal() *Logger {
    return GetLogger()
}

// Log returns the global logger for fluent calls, e.g., Log().Info().Msg("...")
func Log() *Logger {
    return GetLogger()
}

// Simple message helpers for common use without chaining
func TraceMsg(msg string) { GetLogger().Trace().Msg(msg) }
func DebugMsg(msg string) { GetLogger().Debug().Msg(msg) }
func InfoMsg(msg string)  { GetLogger().Info().Msg(msg) }
func WarnMsg(msg string)  { GetLogger().Warn().Msg(msg) }
func ErrorMsg(msg string) { GetLogger().Error().Msg(msg) }
func FatalMsg(msg string) { GetLogger().Fatal().Msg(msg) }

// LogError logs a structured error with correlation
func LogError(ctx context.Context, err error, message string) {
	GetLogger().LogError(ctx, err, message)
}

// LogOperation logs the execution of an operation
func LogOperation(ctx context.Context, operation string, fn func() error) error {
	return GetLogger().LogOperation(ctx, operation, fn)
}

// LogStorageOperation logs storage operations
func LogStorageOperation(ctx context.Context, operation, nodeID, path string, fn func() error) error {
	return GetLogger().LogStorageOperation(ctx, operation, nodeID, path, fn)
}

// LogIndexOperation logs index operations
func LogIndexOperation(ctx context.Context, operation string, count int, fn func() error) error {
	return GetLogger().LogIndexOperation(ctx, operation, count, fn)
}

// NewTrace creates a new trace context for operation tracking
func NewTrace(operation string) *TraceContext {
	return NewTraceContext(operation)
}

// ContextWithTrace creates a context with trace information
func ContextWithTrace(ctx context.Context, trace *TraceContext) context.Context {
	return NewContextWithTrace(ctx, trace)
}

// TraceFromCtx retrieves trace from context
func TraceFromCtx(ctx context.Context) *TraceContext {
	return TraceFromContext(ctx)
}

// ConfigureForEnvironment configures logging based on environment
func ConfigureForEnvironment(projectRoot, environment string) error {
	config := DefaultConfig()
	config.LogDirectory = filepath.Join(projectRoot, "logs")
	
	switch environment {
	case "development", "dev":
		config.Level = LevelDebug
		config.OutputConsole = true
		config.OutputFile = true
	case "production", "prod":
		config.Level = LevelInfo
		config.OutputConsole = false
		config.OutputFile = true
		config.MaxFileSize = 20 // Larger files in production
		config.MaxBackups = 10  // More backups in production
	case "test":
		config.Level = LevelWarn
		config.OutputConsole = false
		config.OutputFile = false
	default:
		config.Level = LevelInfo
		config.OutputConsole = true
		config.OutputFile = true
	}
	
	return Initialize(config)
}

// Shutdown gracefully closes the logger
func Shutdown() {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	
	if globalLogger != nil {
		// If using file output, we should flush any pending writes
		// Note: lumberjack doesn't have an explicit Close method,
		// but zerolog will handle cleanup automatically
		globalLogger = nil
	}
}

// Health checks the logger health and returns status information
func Health() map[string]interface{} {
	logger := GetLogger()
	config := logger.GetConfig()
	
	health := map[string]interface{}{
		"initialized":     globalLogger != nil,
		"level":           string(config.Level),
		"console_output":  config.OutputConsole,
		"file_output":     config.OutputFile,
		"log_directory":   config.LogDirectory,
	}
	
	// Check if log directory exists and is writable
	if config.OutputFile {
		if info, err := os.Stat(config.LogDirectory); err != nil {
			health["directory_status"] = "missing"
			health["directory_error"] = err.Error()
		} else if !info.IsDir() {
			health["directory_status"] = "not_directory"
		} else {
			health["directory_status"] = "ok"
			
			// Check if we can write to the directory
			testFile := filepath.Join(config.LogDirectory, ".write_test")
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				health["write_status"] = "failed"
				health["write_error"] = err.Error()
			} else {
				os.Remove(testFile)
				health["write_status"] = "ok"
			}
		}
	}
	
	return health
}

