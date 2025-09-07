package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			expected: &Config{
				Port:        "8080",
				DatabaseURL: "",
				Storage: StorageConfig{
					Endpoint:        "localhost:9000",
					AccessKeyID:     "minioadmin",
					SecretAccessKey: "minioadmin",
					BucketName:      "images",
					UseSSL:          false,
					Region:          "us-east-1",
				},
			},
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"PORT":                "3000",
				"DATABASE_URL":        "postgres://test:test@localhost/testdb",
				"STORAGE_ENDPOINT":    "s3.amazonaws.com",
				"STORAGE_ACCESS_KEY":  "access123",
				"STORAGE_SECRET_KEY":  "secret123",
				"STORAGE_BUCKET":      "mybucket",
				"STORAGE_USE_SSL":     "true",
				"STORAGE_REGION":      "us-west-2",
			},
			expected: &Config{
				Port:        "3000",
				DatabaseURL: "postgres://test:test@localhost/testdb",
				Storage: StorageConfig{
					Endpoint:        "s3.amazonaws.com",
					AccessKeyID:     "access123",
					SecretAccessKey: "secret123",
					BucketName:      "mybucket",
					UseSSL:          true,
					Region:          "us-west-2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			config, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if config.Port != tt.expected.Port {
				t.Errorf("Port = %v, expected %v", config.Port, tt.expected.Port)
			}

			if config.DatabaseURL != tt.expected.DatabaseURL {
				t.Errorf("DatabaseURL = %v, expected %v", config.DatabaseURL, tt.expected.DatabaseURL)
			}

			if config.Storage.Endpoint != tt.expected.Storage.Endpoint {
				t.Errorf("Storage.Endpoint = %v, expected %v", config.Storage.Endpoint, tt.expected.Storage.Endpoint)
			}

			if config.Storage.UseSSL != tt.expected.Storage.UseSSL {
				t.Errorf("Storage.UseSSL = %v, expected %v", config.Storage.UseSSL, tt.expected.Storage.UseSSL)
			}
		})
	}
}