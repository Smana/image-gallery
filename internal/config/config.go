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
	Cache       CacheConfig
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

// CacheConfig holds Redis cache configuration
type CacheConfig struct {
	Enabled         bool
	Address         string
	Password        string
	Database        int
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolSize        int
	MinIdleConns    int
	MaxIdleConns    int
	MaxConnAge      time.Duration
	PoolTimeout     time.Duration
	IdleTimeout     time.Duration
	DefaultTTL      time.Duration
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

	// Cache configuration
	cacheEnabled, _ := strconv.ParseBool(getEnv("CACHE_ENABLED", "true"))
	cacheDB, _ := strconv.Atoi(getEnv("CACHE_DATABASE", "0"))
	cacheMaxRetries, _ := strconv.Atoi(getEnv("CACHE_MAX_RETRIES", "3"))
	cachePoolSize, _ := strconv.Atoi(getEnv("CACHE_POOL_SIZE", "10"))
	cacheMinIdleConns, _ := strconv.Atoi(getEnv("CACHE_MIN_IDLE_CONNS", "5"))
	cacheMaxIdleConns, _ := strconv.Atoi(getEnv("CACHE_MAX_IDLE_CONNS", "10"))

	cacheMinRetryBackoff, _ := time.ParseDuration(getEnv("CACHE_MIN_RETRY_BACKOFF", "8ms"))
	cacheMaxRetryBackoff, _ := time.ParseDuration(getEnv("CACHE_MAX_RETRY_BACKOFF", "512ms"))
	cacheDialTimeout, _ := time.ParseDuration(getEnv("CACHE_DIAL_TIMEOUT", "5s"))
	cacheReadTimeout, _ := time.ParseDuration(getEnv("CACHE_READ_TIMEOUT", "3s"))
	cacheWriteTimeout, _ := time.ParseDuration(getEnv("CACHE_WRITE_TIMEOUT", "3s"))
	cacheMaxConnAge, _ := time.ParseDuration(getEnv("CACHE_MAX_CONN_AGE", "30m"))
	cachePoolTimeout, _ := time.ParseDuration(getEnv("CACHE_POOL_TIMEOUT", "4s"))
	cacheIdleTimeout, _ := time.ParseDuration(getEnv("CACHE_IDLE_TIMEOUT", "5m"))
	cacheDefaultTTL, _ := time.ParseDuration(getEnv("CACHE_DEFAULT_TTL", "1h"))

	config := &Config{
		Environment: getEnv("GO_ENV", "development"),
		Port:        getEnv("PORT", "8080"),
		Host:        getEnv("HOST", "localhost"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		Storage: StorageConfig{
			Endpoint:        getEnv("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("STORAGE_ACCESS_KEY", ""),
			SecretAccessKey: getEnv("STORAGE_SECRET_KEY", ""),
			BucketName:      getEnv("STORAGE_BUCKET", "images"),
			UseSSL:          useSSL,
			Region:          getEnv("STORAGE_REGION", "us-east-1"),
			MaxUploadSize:   maxUploadSize,
			AllowedTypes:    allowedTypes,
		},
		Cache: CacheConfig{
			Enabled:         cacheEnabled,
			Address:         getEnv("CACHE_ADDRESS", "localhost:6379"),
			Password:        getEnv("CACHE_PASSWORD", ""),
			Database:        cacheDB,
			MaxRetries:      cacheMaxRetries,
			MinRetryBackoff: cacheMinRetryBackoff,
			MaxRetryBackoff: cacheMaxRetryBackoff,
			DialTimeout:     cacheDialTimeout,
			ReadTimeout:     cacheReadTimeout,
			WriteTimeout:    cacheWriteTimeout,
			PoolSize:        cachePoolSize,
			MinIdleConns:    cacheMinIdleConns,
			MaxIdleConns:    cacheMaxIdleConns,
			MaxConnAge:      cacheMaxConnAge,
			PoolTimeout:     cachePoolTimeout,
			IdleTimeout:     cacheIdleTimeout,
			DefaultTTL:      cacheDefaultTTL,
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
