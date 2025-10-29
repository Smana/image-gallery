package observability

import (
	"fmt"
	"os"
)

// Config holds configuration for OpenTelemetry instrumentation
type Config struct {
	// Service identification
	ServiceName    string
	ServiceVersion string
	Environment    string

	// OTLP Trace Exporter configuration
	TracesEndpoint string
	TracesEnabled  bool

	// OTLP Metrics Exporter configuration
	MetricsEndpoint string
	MetricsEnabled  bool

	// Logging configuration
	LogLevel  string
	LogFormat string // json or console
}

// LoadConfig loads observability configuration from environment variables
func LoadConfig() Config {
	return Config{
		// Service identification
		ServiceName:    getEnv("OTEL_SERVICE_NAME", "image-gallery"),
		ServiceVersion: getEnv("OTEL_SERVICE_VERSION", "1.3.0"),
		Environment:    getEnv("OTEL_DEPLOYMENT_ENVIRONMENT", "development"),

		// Traces configuration
		TracesEndpoint: getEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://localhost:4318/v1/traces"),
		TracesEnabled:  getEnvBool("OTEL_TRACES_ENABLED", true),

		// Metrics configuration
		MetricsEndpoint: getEnv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://localhost:4318/v1/metrics"),
		MetricsEnabled:  getEnvBool("OTEL_METRICS_ENABLED", true),

		// Logging configuration
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
	}
}

// Validate checks if the configuration is valid
func (c Config) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	if c.TracesEnabled && c.TracesEndpoint == "" {
		return fmt.Errorf("traces endpoint is required when traces are enabled")
	}

	if c.MetricsEnabled && c.MetricsEndpoint == "" {
		return fmt.Errorf("metrics endpoint is required when metrics are enabled")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}
