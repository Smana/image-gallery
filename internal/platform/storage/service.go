package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"path/filepath"
	"strings"
	"time"

	"image-gallery/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Service implements the domain StorageService interface
type Service struct {
	client     *minio.Client
	bucketName string
	config     *config.StorageConfig
}

// NewService creates a new storage service
func NewService(cfg *config.StorageConfig) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("storage config cannot be nil")
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	service := &Service{
		client:     client,
		bucketName: cfg.BucketName,
		config:     cfg,
	}

	if err := service.ensureBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return service, nil
}

// Store implements StorageService.Store
func (s *Service) Store(ctx context.Context, filename string, contentType string, data io.Reader) (string, error) {
	if filename == "" {
		return "", errors.New("filename cannot be empty")
	}
	
	if contentType == "" {
		return "", errors.New("content type cannot be empty")
	}

	if data == nil {
		return "", errors.New("data cannot be nil")
	}

	// Security validations
	if err := s.validateFileSecurity(filename, contentType); err != nil {
		return "", fmt.Errorf("security validation failed: %w", err)
	}

	// Validate content type
	if !s.isValidContentType(contentType) {
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Create a buffered reader for magic number validation
	bufferedData, err := s.validateFileContent(data, contentType)
	if err != nil {
		return "", fmt.Errorf("file content validation failed: %w", err)
	}

	// Generate unique storage path
	storagePath := s.generateStoragePath(filename)

	// Calculate file size and hash with limits
	sizeReader := &sizeCountingReader{reader: bufferedData, maxSize: s.getMaxFileSize()}
	hashReader := &hashingReader{reader: sizeReader, hasher: sha256.New()}

	// Upload to MinIO
	info, err := s.client.PutObject(ctx, s.bucketName, storagePath, hashReader, -1, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"original-filename": filename,
			"upload-time":       time.Now().UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Validate upload
	if info.Size == 0 {
		// Clean up failed upload
		s.client.RemoveObject(ctx, s.bucketName, storagePath, minio.RemoveObjectOptions{})
		return "", errors.New("uploaded file has zero size")
	}

	// Note: MaxUploadSize validation moved to application layer
	// This can be added to config if needed

	return storagePath, nil
}

// Retrieve implements StorageService.Retrieve
func (s *Service) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	if path == "" {
		return nil, errors.New("path cannot be empty")
	}

	obj, err := s.client.GetObject(ctx, s.bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	// Verify object exists by attempting to read stat
	_, err = obj.Stat()
	if err != nil {
		obj.Close()
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	return obj, nil
}

// Delete implements StorageService.Delete
func (s *Service) Delete(ctx context.Context, path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	// Check if object exists before deletion
	exists, err := s.Exists(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to check if object exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("file not found: %s", path)
	}

	err = s.client.RemoveObject(ctx, s.bucketName, path, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// Exists implements StorageService.Exists
func (s *Service) Exists(ctx context.Context, path string) (bool, error) {
	if path == "" {
		return false, errors.New("path cannot be empty")
	}

	_, err := s.client.StatObject(ctx, s.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat object: %w", err)
	}

	return true, nil
}

// GenerateURL implements StorageService.GenerateURL
func (s *Service) GenerateURL(ctx context.Context, path string, expiry int64) (string, error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}

	if expiry <= 0 {
		expiry = 3600 // Default 1 hour
	}

	duration := time.Duration(expiry) * time.Second
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, path, duration, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// FileInfo represents metadata about a stored file
type FileInfo struct {
	Path         string
	Size         int64
	ContentType  string
	LastModified int64
	ETag         string
}

// GetFileInfo implements StorageService.GetFileInfo
func (s *Service) GetFileInfo(ctx context.Context, path string) (*FileInfo, error) {
	if path == "" {
		return nil, errors.New("path cannot be empty")
	}

	stat, err := s.client.StatObject(ctx, s.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	return &FileInfo{
		Path:         path,
		Size:         stat.Size,
		ContentType:  stat.ContentType,
		LastModified: stat.LastModified.Unix(),
		ETag:         stat.ETag,
	}, nil
}

// Health checks the health of the storage service
func (s *Service) Health(ctx context.Context) error {
	// Check if we can list objects (basic connectivity test)
	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		MaxKeys: 1,
	})

	// Consume one object from the channel or timeout
	select {
	case <-objectCh:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("storage health check timeout")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ListObjects lists objects in the bucket with pagination
func (s *Service) ListObjects(ctx context.Context, prefix string, maxKeys int) ([]ObjectInfo, error) {
	if maxKeys <= 0 {
		maxKeys = 1000 // Default limit
	}

	options := minio.ListObjectsOptions{
		Prefix:     prefix,
		MaxKeys:    maxKeys,
		Recursive:  true,
		WithMetadata: true,
	}

	objectCh := s.client.ListObjects(ctx, s.bucketName, options)
	
	var objects []ObjectInfo
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}

		objects = append(objects, ObjectInfo{
			Key:          object.Key,
			Size:         object.Size,
			ContentType:  object.ContentType,
			LastModified: object.LastModified,
			ETag:         object.ETag,
			UserMetadata: object.UserMetadata,
		})
	}

	return objects, nil
}

// Copy copies an object from one path to another
func (s *Service) Copy(ctx context.Context, srcPath, dstPath string) error {
	if srcPath == "" || dstPath == "" {
		return errors.New("source and destination paths cannot be empty")
	}

	srcOptions := minio.CopySrcOptions{
		Bucket: s.bucketName,
		Object: srcPath,
	}

	dstOptions := minio.CopyDestOptions{
		Bucket: s.bucketName,
		Object: dstPath,
	}

	_, err := s.client.CopyObject(ctx, dstOptions, srcOptions)
	if err != nil {
		return fmt.Errorf("failed to copy object from %s to %s: %w", srcPath, dstPath, err)
	}

	return nil
}

// GetObjectURL returns a public URL for an object (if bucket policy allows)
func (s *Service) GetObjectURL(path string) string {
	if s.config.UseSSL {
		return fmt.Sprintf("https://%s/%s/%s", s.config.Endpoint, s.bucketName, path)
	}
	return fmt.Sprintf("http://%s/%s/%s", s.config.Endpoint, s.bucketName, path)
}

// Private helper methods

func (s *Service) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{
			Region: s.config.Region,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) generateStoragePath(filename string) string {
	// Create a hash-based directory structure for better distribution
	hash := sha256.Sum256([]byte(filename + time.Now().String()))
	hashStr := fmt.Sprintf("%x", hash)
	
	// Use first 2 characters for directory structure
	dir1 := hashStr[:2]
	dir2 := hashStr[2:4]
	
	// Add timestamp to ensure uniqueness
	timestamp := time.Now().Unix()
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filepath.Base(filename), ext)
	
	// Clean filename for storage
	cleanBase := s.sanitizeFilename(base)
	
	return fmt.Sprintf("%s/%s/%s_%d%s", dir1, dir2, cleanBase, timestamp, ext)
}

func (s *Service) sanitizeFilename(filename string) string {
	// Remove or replace problematic characters
	sanitized := strings.ReplaceAll(filename, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	sanitized = strings.ReplaceAll(sanitized, "..", "_")
	
	// Limit length
	if len(sanitized) > 100 {
		sanitized = sanitized[:100]
	}
	
	return sanitized
}

func (s *Service) isValidContentType(contentType string) bool {
	// Basic supported content types - this could be moved to config
	supportedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return supportedTypes[contentType]
}

// ObjectInfo represents information about a stored object
type ObjectInfo struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	LastModified time.Time         `json:"last_modified"`
	ETag         string            `json:"etag"`
	UserMetadata map[string]string `json:"user_metadata,omitempty"`
}

// Helper readers

type sizeCountingReader struct {
	reader  io.Reader
	size    int64
	maxSize int64
}

func (r *sizeCountingReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.size += int64(n)
	
	// Check if we've exceeded the maximum file size
	if r.maxSize > 0 && r.size > r.maxSize {
		return 0, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", r.maxSize)
	}
	
	return
}

type hashingReader struct {
	reader io.Reader
	hasher hash.Hash
}

func (r *hashingReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if n > 0 {
		r.hasher.Write(p[:n])
	}
	return
}

// Legacy support - keep the original MinIOClient from minio.go working

// Legacy methods for backward compatibility
func (s *Service) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := s.Store(ctx, objectName, contentType, reader)
	return err
}

func (s *Service) GetFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return s.Retrieve(ctx, objectName)
}

func (s *Service) DeleteFile(ctx context.Context, objectName string) error {
	return s.Delete(ctx, objectName)
}

func (s *Service) GetFileURL(objectName string) string {
	return s.GetObjectURL(objectName)
}

// Security validation methods

// validateFileSecurity performs comprehensive file security validations
func (s *Service) validateFileSecurity(filename, contentType string) error {
	// Check for path traversal attempts
	if err := s.validateFilename(filename); err != nil {
		return err
	}
	
	// Validate file extension matches content type
	if err := s.validateExtensionContentType(filename, contentType); err != nil {
		return err
	}
	
	// Additional security checks can be added here
	// - Blacklisted filename patterns
	// - Suspicious file extensions
	// - Rate limiting per IP/user (would need context)
	
	return nil
}

// validateFilename checks for security issues in filenames
func (s *Service) validateFilename(filename string) error {
	// Check for path traversal attempts
	if strings.Contains(filename, "..") {
		return errors.New("path traversal attempt detected")
	}
	
	// Check for absolute paths
	if strings.HasPrefix(filename, "/") || strings.HasPrefix(filename, "\\") {
		return errors.New("absolute paths not allowed")
	}
	
	// Check for hidden files (starting with .)
	if strings.HasPrefix(filepath.Base(filename), ".") {
		return errors.New("hidden files not allowed")
	}
	
	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"data:",
		"vbscript:",
		"onload=",
		"onerror=",
	}
	
	lowerFilename := strings.ToLower(filename)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerFilename, pattern) {
			return fmt.Errorf("suspicious pattern detected: %s", pattern)
		}
	}
	
	// Validate filename length
	if len(filename) > 255 {
		return errors.New("filename too long")
	}
	
	// Check for null bytes
	if strings.Contains(filename, "\x00") {
		return errors.New("null bytes not allowed in filename")
	}
	
	return nil
}

// validateExtensionContentType ensures file extension matches declared content type
func (s *Service) validateExtensionContentType(filename, contentType string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	
	// Map of extensions to expected content types
	expectedTypes := map[string][]string{
		".jpg":  {"image/jpeg", "image/jpg"},
		".jpeg": {"image/jpeg", "image/jpg"},
		".png":  {"image/png"},
		".gif":  {"image/gif"},
		".webp": {"image/webp"},
	}
	
	if expected, exists := expectedTypes[ext]; exists {
		for _, expectedType := range expected {
			if contentType == expectedType {
				return nil
			}
		}
		return fmt.Errorf("content type %s does not match file extension %s", contentType, ext)
	}
	
	return fmt.Errorf("unsupported file extension: %s", ext)
}

// validateFileContent validates the actual file content against declared content type
func (s *Service) validateFileContent(data io.Reader, contentType string) (io.Reader, error) {
	// Read first 512 bytes for magic number validation
	buffer := make([]byte, 512)
	n, err := io.ReadFull(data, buffer)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}
	
	// Validate magic numbers (file signatures)
	if err := s.validateMagicNumber(buffer[:n], contentType); err != nil {
		return nil, err
	}
	
	// Create new reader that includes the header we already read
	return io.MultiReader(bytes.NewReader(buffer[:n]), data), nil
}

// validateMagicNumber checks file magic numbers against content type
func (s *Service) validateMagicNumber(header []byte, contentType string) error {
	if len(header) < 4 {
		return errors.New("file too small to validate")
	}
	
	// Define magic number patterns
	magicNumbers := map[string][][]byte{
		"image/jpeg": {
			{0xFF, 0xD8, 0xFF}, // JPEG
		},
		"image/jpg": {
			{0xFF, 0xD8, 0xFF}, // JPEG (alternative content type)
		},
		"image/png": {
			{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG
		},
		"image/gif": {
			{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}, // GIF87a
			{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}, // GIF89a
		},
		"image/webp": {
			{0x52, 0x49, 0x46, 0x46}, // RIFF (WebP container)
		},
	}
	
	patterns, exists := magicNumbers[contentType]
	if !exists {
		return fmt.Errorf("magic number validation not supported for content type: %s", contentType)
	}
	
	for _, pattern := range patterns {
		if len(header) >= len(pattern) {
			match := true
			for i, b := range pattern {
				if header[i] != b {
					match = false
					break
				}
			}
			if match {
				// For WebP, need additional validation
				if contentType == "image/webp" && len(header) >= 12 {
					// Check for WEBP signature at offset 8
					if !(header[8] == 0x57 && header[9] == 0x45 && header[10] == 0x42 && header[11] == 0x50) {
						continue
					}
				}
				return nil
			}
		}
	}
	
	return fmt.Errorf("file content does not match declared content type %s", contentType)
}

// getMaxFileSize returns the maximum allowed file size in bytes
func (s *Service) getMaxFileSize() int64 {
	// Default: 10MB, could be made configurable
	return 10 * 1024 * 1024
}