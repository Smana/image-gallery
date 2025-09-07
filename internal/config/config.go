// Package config provides application configuration management
// with validation, environment parsing, and Go 1.25 best practices
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents the application configuration
type Config struct {
	Environment string
	Port        string
	Host        string
	DatabaseURL string
	Storage     StorageConfig
	Logging     *LoggingConfig
	Server      *ServerConfig
}

// StorageConfig holds object storage configuration
type StorageConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
	Region          string
	MaxUploadSize   int64
	AllowedTypes    []string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
	Output string
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Load creates a new configuration from environment variables with validation
func Load() (*Config, error) {
	useSSL, _ := strconv.ParseBool(getEnv("STORAGE_USE_SSL", "false"))
	maxUploadSize := parseSize(getEnv("MAX_UPLOAD_SIZE", "10MB"))
	allowedTypes := parseList(getEnv("ALLOWED_FILE_TYPES", "image/jpeg,image/png,image/gif,image/webp"))

	readTimeout, _ := time.ParseDuration(getEnv("READ_TIMEOUT", "10s"))
	writeTimeout, _ := time.ParseDuration(getEnv("WRITE_TIMEOUT", "10s"))
	idleTimeout, _ := time.ParseDuration(getEnv("SERVER_TIMEOUT", "30s"))

	config := &Config{
		Environment: getEnv("GO_ENV", "development"),
		Port:        getEnv("PORT", "8080"),
		Host:        getEnv("HOST", "localhost"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		Storage: StorageConfig{
			Endpoint:        getEnv("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("STORAGE_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("STORAGE_SECRET_KEY", "minioadmin"),
			BucketName:      getEnv("STORAGE_BUCKET", "images"),
			UseSSL:          useSSL,
			Region:          getEnv("STORAGE_REGION", "us-east-1"),
			MaxUploadSize:   maxUploadSize,
			AllowedTypes:    allowedTypes,
		},
		Logging: &LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		Server: &ServerConfig{
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}

	// Validate configuration before returning
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseSize parses size strings like "10MB", "512KB" into bytes
func parseSize(sizeStr string) int64 {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	if strings.HasSuffix(sizeStr, "MB") {
		numStr := strings.TrimSuffix(sizeStr, "MB")
		if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
			return num * 1024 * 1024
		}
	}

	if strings.HasSuffix(sizeStr, "KB") {
		numStr := strings.TrimSuffix(sizeStr, "KB")
		if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
			return num * 1024
		}
	}

	// Default to 10MB if parsing fails
	return 10 * 1024 * 1024
}

// parseList parses comma-separated strings into slices
func parseList(listStr string) []string {
	if listStr == "" {
		return []string{}
	}

	items := strings.Split(listStr, ",")
	result := make([]string, 0, len(items))

	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// MustLoad loads configuration and panics on error
// Useful for startup scenarios where invalid config should crash the application
func MustLoad() *Config {
	config, err := Load()
	if err != nil {
		panic(err)
	}
	return config
}
