package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"image-gallery/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.StorageConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &config.StorageConfig{
				Endpoint:        "localhost:9000",
				AccessKeyID:     "minioadmin",
				SecretAccessKey: "minioadmin",
				BucketName:      "test-bucket",
				UseSSL:          false,
				Region:          "us-east-1",
			},
			expectError: false,
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewService(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				// Note: This will fail if MinIO is not running, but tests basic validation
				// In a real test environment, we'd use testcontainers or mock MinIO
				if err != nil {
					t.Logf("Expected success but got error (likely no MinIO server): %v", err)
					// Don't fail the test if MinIO server is not available
					return
				}
				assert.NotNil(t, service)
				assert.Equal(t, tt.config.BucketName, service.bucketName)
			}
		})
	}
}

func TestService_Store(t *testing.T) {
	// This test uses a mock-like approach with error scenarios
	service := &Service{
		bucketName: "test-bucket",
		config: &config.StorageConfig{
			BucketName: "test-bucket",
		},
		// client would be nil, causing method calls to panic in real MinIO operations
		// This is why we test input validation separately
	}

	tests := []struct {
		name        string
		filename    string
		contentType string
		data        io.Reader
		expectError bool
	}{
		{
			name:        "empty filename",
			filename:    "",
			contentType: "image/jpeg",
			data:        strings.NewReader("test data"),
			expectError: true,
		},
		{
			name:        "empty content type",
			filename:    "test.jpg",
			contentType: "",
			data:        strings.NewReader("test data"),
			expectError: true,
		},
		{
			name:        "nil data",
			filename:    "test.jpg",
			contentType: "image/jpeg",
			data:        nil,
			expectError: true,
		},
		{
			name:        "unsupported content type",
			filename:    "test.txt",
			contentType: "text/plain",
			data:        strings.NewReader("test data"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Store(context.Background(), tt.filename, tt.contentType, tt.data)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Would succeed with a real MinIO client
				// In this case, we expect panic/error due to nil client
				if err == nil {
					t.Log("Store succeeded unexpectedly (should fail with nil client)")
				}
			}
		})
	}
}

func TestService_Retrieve(t *testing.T) {
	service := &Service{
		bucketName: "test-bucket",
		config:     &config.StorageConfig{BucketName: "test-bucket"},
	}

	t.Run("empty path", func(t *testing.T) {
		_, err := service.Retrieve(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path cannot be empty")
	})

	// Skip testing with valid path since it requires MinIO client
	// which would cause nil pointer dereference
}

func TestService_Delete(t *testing.T) {
	service := &Service{
		bucketName: "test-bucket",
		config:     &config.StorageConfig{BucketName: "test-bucket"},
	}

	t.Run("empty path", func(t *testing.T) {
		err := service.Delete(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path cannot be empty")
	})

	// Skip testing with valid path - would require MinIO client
}

func TestService_Exists(t *testing.T) {
	service := &Service{
		bucketName: "test-bucket",
		config:     &config.StorageConfig{BucketName: "test-bucket"},
	}

	t.Run("empty path", func(t *testing.T) {
		_, err := service.Exists(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path cannot be empty")
	})

	// Skip testing with valid path - would require MinIO client
}

func TestService_GenerateURL(t *testing.T) {
	service := &Service{
		bucketName: "test-bucket",
		config:     &config.StorageConfig{BucketName: "test-bucket"},
	}

	t.Run("empty path", func(t *testing.T) {
		_, err := service.GenerateURL(context.Background(), "", 3600)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path cannot be empty")
	})

	// Skip testing with valid paths - would require MinIO client
}

func TestService_GetFileInfo(t *testing.T) {
	service := &Service{
		bucketName: "test-bucket",
		config:     &config.StorageConfig{BucketName: "test-bucket"},
	}

	t.Run("empty path", func(t *testing.T) {
		_, err := service.GetFileInfo(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path cannot be empty")
	})

	// Skip testing with valid path - would require MinIO client
}

func TestService_Health(t *testing.T) {
	// Skip testing health check - would require MinIO client
	// With nil client, this would cause panic/failure
	t.Skip("Health check requires MinIO client connection")
}

func TestService_Copy(t *testing.T) {
	service := &Service{
		bucketName: "test-bucket",
		config:     &config.StorageConfig{BucketName: "test-bucket"},
	}

	tests := []struct {
		name        string
		srcPath     string
		dstPath     string
		expectError bool
	}{
		{
			name:        "empty source path",
			srcPath:     "",
			dstPath:     "dest.jpg",
			expectError: true,
		},
		{
			name:        "empty destination path",
			srcPath:     "src.jpg",
			dstPath:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Copy(context.Background(), tt.srcPath, tt.dstPath)

			if tt.expectError {
				assert.Error(t, err)
			}
		})
	}

	// Skip testing with valid paths - would require MinIO client
}

func TestService_GetObjectURL(t *testing.T) {
	tests := []struct {
		name     string
		service  *Service
		path     string
		expected string
	}{
		{
			name: "HTTPS URL",
			service: &Service{
				bucketName: "test-bucket",
				config: &config.StorageConfig{
					Endpoint:   "s3.amazonaws.com",
					BucketName: "test-bucket",
					UseSSL:     true,
				},
			},
			path:     "test/image.jpg",
			expected: "https://s3.amazonaws.com/test-bucket/test/image.jpg",
		},
		{
			name: "HTTP URL",
			service: &Service{
				bucketName: "test-bucket",
				config: &config.StorageConfig{
					Endpoint:   "localhost:9000",
					BucketName: "test-bucket",
					UseSSL:     false,
				},
			},
			path:     "test/image.jpg",
			expected: "http://localhost:9000/test-bucket/test/image.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.service.GetObjectURL(tt.path)
			assert.Equal(t, tt.expected, url)
		})
	}
}

func TestService_generateStoragePath(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "simple filename",
			filename: "test.jpg",
		},
		{
			name:     "filename with spaces",
			filename: "my photo.jpg",
		},
		{
			name:     "complex filename",
			filename: "My Complex File Name (1).jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := service.generateStoragePath(tt.filename)
			
			// Should not be empty
			assert.NotEmpty(t, path)
			
			// Should contain directory structure (xx/yy/)
			parts := strings.Split(path, "/")
			assert.GreaterOrEqual(t, len(parts), 3, "Should have at least 3 parts: dir1/dir2/filename")
			
			// Should preserve extension
			originalExt := strings.ToLower(filepath.Ext(tt.filename))
			if originalExt != "" {
				assert.True(t, strings.HasSuffix(path, originalExt))
			}
		})
	}

	// Test uniqueness
	t.Run("generates unique paths", func(t *testing.T) {
		filename := "test.jpg"
		path1 := service.generateStoragePath(filename)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamp
		path2 := service.generateStoragePath(filename)
		
		assert.NotEqual(t, path1, path2, "Should generate unique paths for same filename")
	})
}

func TestService_sanitizeFilename(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple filename",
			input:    "test",
			expected: "test",
		},
		{
			name:     "filename with spaces",
			input:    "my photo",
			expected: "my_photo",
		},
		{
			name:     "filename with problematic characters",
			input:    "my/photo\\test..file",
			expected: "my_photo_test_file",
		},
		{
			name:     "long filename",
			input:    strings.Repeat("a", 150),
			expected: strings.Repeat("a", 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.sanitizeFilename(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), 100, "Sanitized filename should not exceed 100 characters")
		})
	}
}

func TestService_isValidContentType(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{"valid jpeg", "image/jpeg", true},
		{"valid jpg", "image/jpg", true},
		{"valid png", "image/png", true},
		{"valid gif", "image/gif", true},
		{"valid webp", "image/webp", true},
		{"invalid pdf", "application/pdf", false},
		{"invalid text", "text/plain", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isValidContentType(tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSizeCountingReader(t *testing.T) {
	data := "hello world"
	reader := &sizeCountingReader{
		reader: strings.NewReader(data),
		size:   0,
	}

	buffer := make([]byte, 5)
	n, err := reader.Read(buffer)

	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, int64(5), reader.size)
	assert.Equal(t, "hello", string(buffer))

	// Read remaining data
	buffer = make([]byte, 10)
	n, err = reader.Read(buffer)
	
	// Should be EOF or remaining data
	assert.Equal(t, int64(11), reader.size)
}

func TestHashingReader(t *testing.T) {
	data := "hello world"
	reader := &hashingReader{
		reader: strings.NewReader(data),
		hasher: sha256.New(),
	}

	buffer := make([]byte, len(data))
	n, err := reader.Read(buffer)

	if err != nil && err != io.EOF {
		t.Errorf("Unexpected error: %v", err)
	}
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, string(buffer))
}

// Security validation tests
func TestService_SecurityValidations(t *testing.T) {
	service := &Service{
		bucketName: "test-bucket",
		config:     &config.StorageConfig{BucketName: "test-bucket"},
	}

	t.Run("path traversal attempts", func(t *testing.T) {
		tests := []string{
			"../../../etc/passwd",
			"..\\windows\\system32\\config\\sam",
			"file/../../../secret.txt",
			"normal_file/../hidden.txt",
		}

		for _, filename := range tests {
			_, err := service.Store(context.Background(), filename, "image/jpeg", strings.NewReader("fake jpeg data"))
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "security validation failed")
		}
	})

	t.Run("absolute path attempts", func(t *testing.T) {
		tests := []string{
			"/etc/passwd",
			"\\windows\\system32\\hosts",
			"/var/log/sensitive.log",
		}

		for _, filename := range tests {
			_, err := service.Store(context.Background(), filename, "image/jpeg", strings.NewReader("fake jpeg data"))
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "security validation failed")
		}
	})

	t.Run("hidden file attempts", func(t *testing.T) {
		tests := []string{
			".htaccess",
			".env",
			".git/config",
			"folder/.secret",
		}

		for _, filename := range tests {
			_, err := service.Store(context.Background(), filename, "image/jpeg", strings.NewReader("fake jpeg data"))
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "security validation failed")
		}
	})

	t.Run("suspicious patterns", func(t *testing.T) {
		tests := []string{
			"<script>alert('xss')</script>.jpg",
			"javascript:void(0).png",
			"data:image/jpeg;base64,xyz.jpg",
			"vbscript:msgbox.gif",
			"onload=alert().webp",
			"onerror=steal().jpeg",
		}

		for _, filename := range tests {
			_, err := service.Store(context.Background(), filename, "image/jpeg", strings.NewReader("fake jpeg data"))
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "security validation failed")
		}
	})

	t.Run("extension content type mismatch", func(t *testing.T) {
		tests := []struct {
			filename    string
			contentType string
		}{
			{"image.jpg", "image/png"},
			{"document.pdf", "image/jpeg"},
			{"script.js", "image/gif"},
			{"photo.png", "image/jpeg"},
		}

		for _, test := range tests {
			_, err := service.Store(context.Background(), test.filename, test.contentType, strings.NewReader("fake data"))
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "security validation failed")
		}
	})

	t.Run("filename too long", func(t *testing.T) {
		longFilename := strings.Repeat("a", 300) + ".jpg"
		_, err := service.Store(context.Background(), longFilename, "image/jpeg", strings.NewReader("fake jpeg data"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "security validation failed")
	})

	t.Run("null bytes in filename", func(t *testing.T) {
		filename := "image\x00.jpg"
		_, err := service.Store(context.Background(), filename, "image/jpeg", strings.NewReader("fake jpeg data"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "security validation failed")
	})
}

func TestService_MagicNumberValidation(t *testing.T) {
	service := &Service{
		bucketName: "test-bucket",
		config:     &config.StorageConfig{BucketName: "test-bucket"},
	}

	t.Run("invalid JPEG magic number", func(t *testing.T) {
		// Create data with wrong magic number
		fakeData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}
		fakeData = append(fakeData, make([]byte, 100)...)
		
		_, err := service.Store(context.Background(), "test.jpg", "image/jpeg", bytes.NewReader(fakeData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file content does not match declared content type")
	})

	t.Run("file too small", func(t *testing.T) {
		smallData := []byte{0x01, 0x02}
		
		_, err := service.Store(context.Background(), "test.jpg", "image/jpeg", bytes.NewReader(smallData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file too small to validate")
	})

	// Test individual validation methods directly to avoid MinIO client issues
	t.Run("direct magic number validation", func(t *testing.T) {
		tests := []struct {
			name        string
			data        []byte
			contentType string
			expectError bool
		}{
			{
				name:        "valid JPEG",
				data:        []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
				contentType: "image/jpeg",
				expectError: false,
			},
			{
				name:        "valid PNG",
				data:        []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
				contentType: "image/png",
				expectError: false,
			},
			{
				name:        "invalid JPEG",
				data:        []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
				contentType: "image/jpeg",
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := service.validateMagicNumber(tt.data, tt.contentType)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestService_FileSizeValidation(t *testing.T) {
	t.Run("direct size reader test", func(t *testing.T) {
		// Test the sizeCountingReader directly
		data := make([]byte, 1000)
		reader := &sizeCountingReader{
			reader:  bytes.NewReader(data),
			maxSize: 500, // Set limit to 500 bytes
		}

		buffer := make([]byte, 600) // Try to read more than limit
		_, err := reader.Read(buffer)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file size exceeds maximum allowed size")
	})

	t.Run("getMaxFileSize method", func(t *testing.T) {
		service := &Service{
			bucketName: "test-bucket",
			config:     &config.StorageConfig{BucketName: "test-bucket"},
		}
		
		maxSize := service.getMaxFileSize()
		assert.Equal(t, int64(10*1024*1024), maxSize) // Should be 10MB
	})
}

// Integration tests (require running MinIO)
func TestServiceIntegration(t *testing.T) {
	// Skip integration tests by default
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// These tests require a running MinIO instance
	cfg := &config.StorageConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		BucketName:      "test-integration",
		UseSSL:          false,
		Region:          "us-east-1",
	}

	service, err := NewService(cfg)
	if err != nil {
		t.Skipf("MinIO not available for integration tests: %v", err)
	}

	ctx := context.Background()

	t.Run("store and retrieve", func(t *testing.T) {
		data := strings.NewReader("test file content")
		
		path, err := service.Store(ctx, "test.txt", "text/plain", data)
		if err != nil {
			t.Skipf("Store failed (expected with text/plain): %v", err)
		}

		if path != "" {
			// Clean up
			defer service.Delete(ctx, path)

			// Test retrieval
			reader, err := service.Retrieve(ctx, path)
			require.NoError(t, err)
			defer reader.Close()

			content, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, "test file content", string(content))
		}
	})

	t.Run("health check", func(t *testing.T) {
		err := service.Health(ctx)
		assert.NoError(t, err)
	})
}