package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// EnvironmentConfig represents environment-specific logging configuration
type EnvironmentConfig struct {
	Environment string  `json:"environment"`
	LogLevel    string  `json:"log_level"`
	Console     bool    `json:"console"`
	File        bool    `json:"file"`
	MaxSize     int     `json:"max_size_mb"`
	MaxBackups  int     `json:"max_backups"`
	MaxAge      int     `json:"max_age_days"`
	Compress    bool    `json:"compress"`
	Directory   string  `json:"directory"`
}

// DefaultEnvironmentConfigs returns default configurations for different environments
func DefaultEnvironmentConfigs() map[string]*EnvironmentConfig {
	return map[string]*EnvironmentConfig{
		"development": {
			Environment: "development",
			LogLevel:    "debug",
			Console:     true,
			File:        true,
			MaxSize:     10,
			MaxBackups:  5,
			MaxAge:      7,
			Compress:    false,
			Directory:   "logs",
		},
		"production": {
			Environment: "production",
			LogLevel:    "info",
			Console:     false,
			File:        true,
			MaxSize:     20,
			MaxBackups:  10,
			MaxAge:      30,
			Compress:    true,
			Directory:   "logs",
		},
		"test": {
			Environment: "test",
			LogLevel:    "warn",
			Console:     false,
			File:        false,
			MaxSize:     10,
			MaxBackups:  3,
			MaxAge:      1,
			Compress:    false,
			Directory:   "logs",
		},
	}
}

// LoadConfigFromFile loads logging configuration from a JSON file
func LoadConfigFromFile(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var envConfig EnvironmentConfig
	if err := json.Unmarshal(data, &envConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return envConfigToConfig(&envConfig)
}

// SaveConfigToFile saves the current environment configuration to a file
func SaveConfigToFile(config *EnvironmentConfig, configPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// LoadOrCreateConfig loads configuration from file or creates default for environment
func LoadOrCreateConfig(projectRoot, environment, configPath string) (*Config, error) {
	// Try to load existing config file
	if config, err := LoadConfigFromFile(configPath); err == nil {
		return config, nil
	}
	
	// Create default configuration for the environment
	defaults := DefaultEnvironmentConfigs()
	envConfig, exists := defaults[environment]
	if !exists {
		envConfig = defaults["development"] // fallback
		envConfig.Environment = environment
	}
	
	// Set absolute directory path
	if !filepath.IsAbs(envConfig.Directory) {
		envConfig.Directory = filepath.Join(projectRoot, envConfig.Directory)
	}
	
	// Save the default configuration for future use
	if err := SaveConfigToFile(envConfig, configPath); err != nil {
		// Log warning but don't fail - we can still use the config
		fmt.Fprintf(os.Stderr, "Warning: failed to save default config: %v\n", err)
	}
	
	return envConfigToConfig(envConfig)
}

// envConfigToConfig converts EnvironmentConfig to Config
func envConfigToConfig(envConfig *EnvironmentConfig) (*Config, error) {
	// Convert log level string to LogLevel
	var level LogLevel
	switch envConfig.LogLevel {
	case "trace":
		level = LevelTrace
	case "debug":
		level = LevelDebug
	case "info":
		level = LevelInfo
	case "warn", "warning":
		level = LevelWarn
	case "error":
		level = LevelError
	case "fatal":
		level = LevelFatal
	default:
		level = LevelInfo
	}
	
	return &Config{
		Level:           level,
		OutputConsole:   envConfig.Console,
		OutputFile:      envConfig.File,
		LogDirectory:    envConfig.Directory,
		MaxFileSize:     envConfig.MaxSize,
		MaxBackups:      envConfig.MaxBackups,
		MaxAge:          envConfig.MaxAge,
		CompressBackups: envConfig.Compress,
	}, nil
}

// configToEnvConfig converts Config to EnvironmentConfig
func configToEnvConfig(config *Config, environment string) *EnvironmentConfig {
	return &EnvironmentConfig{
		Environment: environment,
		LogLevel:    string(config.Level),
		Console:     config.OutputConsole,
		File:        config.OutputFile,
		MaxSize:     config.MaxFileSize,
		MaxBackups:  config.MaxBackups,
		MaxAge:      config.MaxAge,
		Compress:    config.CompressBackups,
		Directory:   config.LogDirectory,
	}
}

// GetEnvironmentFromOS returns the environment from OS environment variables
func GetEnvironmentFromOS() string {
	env := os.Getenv("ARCHON_ENV")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = os.Getenv("NODE_ENV") // Common in web development
	}
	if env == "" {
		env = "development" // Default fallback
	}
	return env
}

// InitializeFromEnvironment initializes logging based on environment detection
func InitializeFromEnvironment(projectRoot string) error {
	environment := GetEnvironmentFromOS()
	configPath := filepath.Join(projectRoot, ".archon", "logging.json")
	
	config, err := LoadOrCreateConfig(projectRoot, environment, configPath)
	if err != nil {
		return fmt.Errorf("failed to load logging config for environment %s: %w", environment, err)
	}
	
	return Initialize(config)
}

// UpdateEnvironmentConfig updates the logging configuration for an environment
func UpdateEnvironmentConfig(projectRoot, environment string, updates map[string]interface{}) error {
	configPath := filepath.Join(projectRoot, ".archon", "logging.json")
	
	// Load existing or create default
	config, err := LoadOrCreateConfig(projectRoot, environment, configPath)
	if err != nil {
		return fmt.Errorf("failed to load existing config: %w", err)
	}
	
	// Convert to environment config for easier manipulation
	envConfig := configToEnvConfig(config, environment)
	
	// Apply updates
	for key, value := range updates {
		switch key {
		case "log_level":
			if str, ok := value.(string); ok {
				envConfig.LogLevel = str
			}
		case "console":
			if b, ok := value.(bool); ok {
				envConfig.Console = b
			}
		case "file":
			if b, ok := value.(bool); ok {
				envConfig.File = b
			}
		case "max_size":
			if i, ok := value.(int); ok {
				envConfig.MaxSize = i
			}
		case "max_backups":
			if i, ok := value.(int); ok {
				envConfig.MaxBackups = i
			}
		case "max_age":
			if i, ok := value.(int); ok {
				envConfig.MaxAge = i
			}
		case "compress":
			if b, ok := value.(bool); ok {
				envConfig.Compress = b
			}
		}
	}
	
	// Save updated configuration
	if err := SaveConfigToFile(envConfig, configPath); err != nil {
		return fmt.Errorf("failed to save updated config: %w", err)
	}
	
	// Convert back and reinitialize logger
	newConfig, err := envConfigToConfig(envConfig)
	if err != nil {
		return fmt.Errorf("failed to convert updated config: %w", err)
	}
	
	return Initialize(newConfig)
}