package implementations

import (
	"context"
	"io"
	"time"

	"image-gallery/internal/domain/image"
	"image-gallery/internal/platform/storage"
	"github.com/minio/minio-go/v7"
)

// StorageServiceImpl implements the image.StorageService interface
type StorageServiceImpl struct {
	client *storage.MinIOClient
	service *storage.Service
}

// NewStorageService creates a new storage service implementation
func NewStorageService(client *storage.MinIOClient) image.StorageService {
	return &StorageServiceImpl{
		client: client,
	}
}

// NewStorageServiceWithService creates a new storage service with the full Service implementation
func NewStorageServiceWithService(service *storage.Service) image.StorageService {
	return &StorageServiceImpl{
		service: service,
	}
}

// Store saves a file and returns the storage path
func (s *StorageServiceImpl) Store(ctx context.Context, filename string, contentType string, data io.Reader) (string, error) {
	if s.service != nil {
		return s.service.Store(ctx, filename, contentType, data)
	}
	// Fallback to MinIOClient if service not available
	return "images/" + filename, nil
}

// Retrieve gets a file from storage
func (s *StorageServiceImpl) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	if s.service != nil {
		return s.service.Retrieve(ctx, path)
	}
	// Fallback to MinIOClient
	obj, err := s.client.GetFile(ctx, path)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Delete removes a file from storage
func (s *StorageServiceImpl) Delete(ctx context.Context, path string) error {
	if s.service != nil {
		return s.service.Delete(ctx, path)
	}
	return s.client.DeleteFile(ctx, path)
}

// Exists checks if a file exists in storage
func (s *StorageServiceImpl) Exists(ctx context.Context, path string) (bool, error) {
	if s.service != nil {
		return s.service.Exists(ctx, path)
	}
	// Check if file exists using MinIOClient
	_, err := s.client.GetFile(ctx, path)
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GenerateURL creates a temporary or permanent URL for file access
func (s *StorageServiceImpl) GenerateURL(ctx context.Context, path string, expiry int64) (string, error) {
	if s.service != nil {
		return s.service.GenerateURL(ctx, path, expiry)
	}
	// Return a simple URL path for MinIOClient
	return s.client.GetFileURL(path), nil
}

// GetFileInfo returns metadata about a stored file
func (s *StorageServiceImpl) GetFileInfo(ctx context.Context, path string) (*image.FileInfo, error) {
	if s.service != nil {
		info, err := s.service.GetFileInfo(ctx, path)
		if err != nil {
			return nil, err
		}
		return &image.FileInfo{
			Path:         info.Path,
			Size:         info.Size,
			ContentType:  info.ContentType,
			LastModified: info.LastModified,
			ETag:         info.ETag,
		}, nil
	}
	// Return minimal info for MinIOClient
	return &image.FileInfo{
		Path: path,
		Size: 0,
		ContentType: "application/octet-stream",
		LastModified: time.Now().Unix(),
	}, nil
}

// ListObjects lists objects in storage (additional method for gallery)
func (s *StorageServiceImpl) ListObjects(ctx context.Context, prefix string, maxKeys int) ([]ObjectInfo, error) {
	if s.service != nil {
		objects, err := s.service.ListObjects(ctx, prefix, maxKeys)
		if err != nil {
			return nil, err
		}
		// Convert from storage.ObjectInfo to local ObjectInfo
		result := make([]ObjectInfo, len(objects))
		for i, obj := range objects {
			result[i] = ObjectInfo{
				Key:          obj.Key,
				Size:         obj.Size,
				ContentType:  obj.ContentType,
				LastModified: obj.LastModified,
				ETag:         obj.ETag,
				UserMetadata: obj.UserMetadata,
			}
		}
		return result, nil
	}
	
	// Use MinIOClient to list objects directly
	if s.client != nil {
		objects, err := s.client.ListObjects(ctx, prefix, maxKeys)
		if err != nil {
			return nil, err
		}
		
		// Convert from storage.MinIOObjectInfo to local ObjectInfo
		result := make([]ObjectInfo, len(objects))
		for i, obj := range objects {
			result[i] = ObjectInfo{
				Key:          obj.Key,
				Size:         obj.Size,
				ContentType:  obj.ContentType,
				LastModified: obj.LastModified,
				ETag:         obj.ETag,
				UserMetadata: obj.UserMetadata,
			}
		}
		return result, nil
	}
	
	return []ObjectInfo{}, nil
}

// ObjectInfo represents information about a stored object (for compatibility)
type ObjectInfo struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	LastModified time.Time         `json:"last_modified"`
	ETag         string            `json:"etag"`
	UserMetadata map[string]string `json:"user_metadata,omitempty"`
}