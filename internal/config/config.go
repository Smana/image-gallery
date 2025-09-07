package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DatabaseURL string
	Storage     StorageConfig
}

type StorageConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
	Region          string
}

func Load() (*Config, error) {
	useSSL, _ := strconv.ParseBool(getEnv("STORAGE_USE_SSL", "false"))

	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		Storage: StorageConfig{
			Endpoint:        getEnv("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("STORAGE_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("STORAGE_SECRET_KEY", "minioadmin"),
			BucketName:      getEnv("STORAGE_BUCKET", "images"),
			UseSSL:          useSSL,
			Region:          getEnv("STORAGE_REGION", "us-east-1"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}