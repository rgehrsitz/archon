package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel represents the logging level
type LogLevel string

const (
	// Log levels in order of severity
	LevelTrace LogLevel = "trace"
	LevelDebug LogLevel = "debug" 
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

// Config represents the logging configuration
type Config struct {
	// Level is the minimum log level to output
	Level LogLevel `json:"level"`
	
	// OutputConsole enables console output
	OutputConsole bool `json:"outputConsole"`
	
	// OutputFile enables file output
	OutputFile bool `json:"outputFile"`
	
	// LogDirectory is where log files are stored
	LogDirectory string `json:"logDirectory"`
	
	// MaxFileSize is the maximum size of each log file in megabytes
	MaxFileSize int `json:"maxFileSize"`
	
	// MaxBackups is the maximum number of old log files to keep
	MaxBackups int `json:"maxBackups"`
	
	// MaxAge is the maximum age of log files in days
	MaxAge int `json:"maxAge"`
	
	// CompressBackups determines if old log files should be compressed
	CompressBackups bool `json:"compressBackups"`
}

// DefaultConfig returns a sensible default configuration per ADR-006
func DefaultConfig() *Config {
	return &Config{
		Level:           LevelInfo,
		OutputConsole:   true,
		OutputFile:      true,
		LogDirectory:    "logs",
		MaxFileSize:     10,  // 10MB per ADR-006
		MaxBackups:      5,   // 5 files per ADR-006
		MaxAge:          30,  // 30 days
		CompressBackups: true,
	}
}

// Logger wraps zerolog with additional functionality
type Logger struct {
	logger zerolog.Logger
	config *Config
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config *Config) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	var writers []io.Writer
	
	// Console output with pretty formatting for development
	if config.OutputConsole {
		console := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
		writers = append(writers, console)
	}
	
	// File output with rotation
	if config.OutputFile {
		if err := os.MkdirAll(config.LogDirectory, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		
		fileWriter := &lumberjack.Logger{
			Filename:   filepath.Join(config.LogDirectory, "archon.log"),
			MaxSize:    config.MaxFileSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.CompressBackups,
		}
		writers = append(writers, fileWriter)
	}
	
	// Create multi-writer if we have multiple outputs
	var output io.Writer
	if len(writers) == 1 {
		output = writers[0]
	} else if len(writers) > 1 {
		output = io.MultiWriter(writers...)
	} else {
		// Fallback to console if no writers configured
		output = os.Stderr
	}
	
	// Configure zerolog
	logger := zerolog.New(output).With().
		Timestamp().
		Caller().
		Logger()
	
	// Set log level
	switch config.Level {
	case LevelTrace:
		logger = logger.Level(zerolog.TraceLevel)
	case LevelDebug:
		logger = logger.Level(zerolog.DebugLevel)
	case LevelInfo:
		logger = logger.Level(zerolog.InfoLevel)
	case LevelWarn:
		logger = logger.Level(zerolog.WarnLevel)
	case LevelError:
		logger = logger.Level(zerolog.ErrorLevel)
	case LevelFatal:
		logger = logger.Level(zerolog.FatalLevel)
	default:
		logger = logger.Level(zerolog.InfoLevel)
	}
	
	return &Logger{
		logger: logger,
		config: config,
	}, nil
}

// WithContext returns a new logger with additional context fields
func (l *Logger) WithContext(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	
	return &Logger{
		logger: ctx.Logger(),
		config: l.config,
	}
}

// WithRequestID adds a request ID to the logger context
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("request_id", requestID).Logger(),
		config: l.config,
	}
}

// WithOperation adds an operation name to the logger context
func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("operation", operation).Logger(),
		config: l.config,
	}
}

// WithError adds an error to the logger context
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		logger: l.logger.With().Err(err).Logger(),
		config: l.config,
	}
}

// Trace logs a message at trace level
func (l *Logger) Trace() *zerolog.Event {
	return l.logger.Trace()
}

// Debug logs a message at debug level
func (l *Logger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Info logs a message at info level
func (l *Logger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Warn logs a message at warn level
func (l *Logger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Error logs a message at error level
func (l *Logger) Error() *zerolog.Event {
	return l.logger.Error()
}

// Fatal logs a message at fatal level and exits
func (l *Logger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

// GetConfig returns the current logger configuration
func (l *Logger) GetConfig() *Config {
	return l.config
}

// UpdateLevel changes the logging level at runtime
func (l *Logger) UpdateLevel(level LogLevel) {
	switch level {
	case LevelTrace:
		l.logger = l.logger.Level(zerolog.TraceLevel)
	case LevelDebug:
		l.logger = l.logger.Level(zerolog.DebugLevel)
	case LevelInfo:
		l.logger = l.logger.Level(zerolog.InfoLevel)
	case LevelWarn:
		l.logger = l.logger.Level(zerolog.WarnLevel)
	case LevelError:
		l.logger = l.logger.Level(zerolog.ErrorLevel)
	case LevelFatal:
		l.logger = l.logger.Level(zerolog.FatalLevel)
	}
	l.config.Level = level
}