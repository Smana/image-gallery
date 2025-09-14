package implementations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"image-gallery/internal/domain/image"
)

// ImageServiceImpl implements the image.ImageService interface
type ImageServiceImpl struct {
	imageRepo image.Repository
	tagRepo   image.TagRepository
	storage   image.StorageService
	processor image.ImageProcessor
	validator image.ValidationService
	eventPub  image.EventPublisher // can be nil
	cache     image.CacheService   // can be nil
}

// NewImageService creates a new image service implementation
func NewImageService(
	imageRepo image.Repository,
	tagRepo image.TagRepository,
	storage image.StorageService,
	processor image.ImageProcessor,
	validator image.ValidationService,
	eventPub image.EventPublisher,
	cache image.CacheService,
) image.ImageService {
	return &ImageServiceImpl{
		imageRepo: imageRepo,
		tagRepo:   tagRepo,
		storage:   storage,
		processor: processor,
		validator: validator,
		eventPub:  eventPub,
		cache:     cache,
	}
}

// CreateImage handles the complete image creation process
func (s *ImageServiceImpl) CreateImage(ctx context.Context, req *image.CreateImageRequest, data io.Reader) (*image.Image, error) {
	if req == nil {
		return nil, fmt.Errorf("create request cannot be nil")
	}

	if err := s.validateCreateRequest(ctx, req, data); err != nil {
		return nil, err
	}

	storageResp, err := s.storeImageFile(ctx, req, data)
	if err != nil {
		return nil, err
	}

	tags, err := s.processTags(ctx, req.Tags)
	if err != nil {
		return nil, err
	}

	img := s.buildImageObject(req, storageResp, tags)
	if err := img.Validate(); err != nil {
		return nil, fmt.Errorf("final image validation failed: %w", err)
	}

	if err := s.saveImageToDatabase(ctx, img, storageResp); err != nil {
		return nil, err
	}

	s.handlePostCreation(ctx, img)

	return img, nil
}

func (s *ImageServiceImpl) validateCreateRequest(ctx context.Context, req *image.CreateImageRequest, data io.Reader) error {
	if err := s.validator.ValidateImageUpload(ctx, req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	if err := req.Validate(); err != nil {
		return fmt.Errorf("domain validation failed: %w", err)
	}
	return s.processor.ValidateImage(ctx, data, req.ContentType)
}

func (s *ImageServiceImpl) storeImageFile(ctx context.Context, req *image.CreateImageRequest, data io.Reader) (string, error) {
	filename := s.generateUniqueFilename(req.OriginalFilename)
	storageResp, err := s.storage.Store(ctx, filename, req.ContentType, data)
	if err != nil {
		return "", fmt.Errorf("failed to store image: %w", err)
	}
	return storageResp, nil
}

func (s *ImageServiceImpl) processTags(ctx context.Context, tagNames []string) ([]image.Tag, error) {
	tags := make([]image.Tag, 0, len(tagNames))
	for _, tagName := range tagNames {
		tag, err := s.tagRepo.GetOrCreate(ctx, tagName)
		if err != nil {
			return nil, fmt.Errorf("failed to create/get tag %s: %w", tagName, err)
		}
		tags = append(tags, *tag)
	}
	return tags, nil
}

func (s *ImageServiceImpl) buildImageObject(req *image.CreateImageRequest, storageResp string, tags []image.Tag) *image.Image {
	now := time.Now()
	return &image.Image{
		Filename:         s.generateUniqueFilename(req.OriginalFilename),
		OriginalFilename: req.OriginalFilename,
		ContentType:      req.ContentType,
		FileSize:         req.FileSize,
		StoragePath:      storageResp,
		Width:            req.Width,
		Height:           req.Height,
		UploadedAt:       now,
		Metadata:         req.Metadata,
		CreatedAt:        now,
		UpdatedAt:        now,
		Tags:             tags,
	}
}

func (s *ImageServiceImpl) saveImageToDatabase(ctx context.Context, img *image.Image, storageResp string) error {
	if err := s.imageRepo.Create(ctx, img); err != nil {
		if deleteErr := s.storage.Delete(ctx, storageResp); deleteErr != nil {
			_ = deleteErr // explicitly ignore cleanup errors
		}
		return fmt.Errorf("failed to save image to database: %w", err)
	}
	return nil
}

func (s *ImageServiceImpl) handlePostCreation(ctx context.Context, img *image.Image) {
	if s.cache != nil {
		if err := s.cache.InvalidateImageLists(ctx); err != nil {
			_ = err
		}
	}
	if s.eventPub != nil {
		if err := s.eventPub.PublishImageCreated(ctx, img); err != nil {
			_ = err
		}
	}
}

// generateUniqueFilename creates a unique filename using timestamp and hash
func (s *ImageServiceImpl) generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	base := filepath.Base(originalFilename)
	base = base[:len(base)-len(ext)] // Remove extension

	// Create hash of original filename and current time for uniqueness
	hasher := sha256.New()
	hasher.Write([]byte(base + strconv.FormatInt(time.Now().UnixNano(), 10)))
	hash := hex.EncodeToString(hasher.Sum(nil))[:8] // Use first 8 chars

	return fmt.Sprintf("%s_%s%s", base, hash, ext)
}

// GetImage retrieves an image by ID with all related data
func (s *ImageServiceImpl) GetImage(ctx context.Context, id int) (*image.Image, error) {
	// Try to get from cache first
	if s.cache != nil {
		if cachedImage, err := s.cache.GetImage(ctx, id); err == nil {
			return cachedImage, nil
		}
		// If cache miss or error, continue to database
	}

	// Get from database
	img, err := s.imageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result for future requests
	if s.cache != nil {
		// Cache for 1 hour (3600 seconds)
		if cacheErr := s.cache.SetImage(ctx, img, 3600); cacheErr != nil {
			// Cache errors are intentionally ignored to avoid impacting user requests
			// In production, this would be logged for monitoring purposes
			_ = cacheErr // explicitly ignore
		}
	}

	return img, nil
}

// ListImages retrieves images based on criteria
func (s *ImageServiceImpl) ListImages(ctx context.Context, req *image.ListImagesRequest) (*image.ListImagesResponse, error) {
	// Generate cache key based on request parameters
	cacheKey := generateListCacheKey(req)

	// Try to get from cache first
	if s.cache != nil {
		if cachedResponse, err := s.cache.GetImageList(ctx, cacheKey); err == nil {
			return cachedResponse, nil
		}
		// If cache miss or error, continue to database
	}

	// Get from database
	response, err := s.imageRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	// Cache the result for future requests
	if s.cache != nil {
		// Cache for 30 minutes (1800 seconds) - lists change more frequently
		if cacheErr := s.cache.SetImageList(ctx, cacheKey, response, 1800); cacheErr != nil {
			// Cache errors are intentionally ignored to avoid impacting user requests
			// In production, this would be logged for monitoring purposes
			_ = cacheErr // explicitly ignore
		}
	}

	return response, nil
}

// generateListCacheKey creates a cache key based on list request parameters
func generateListCacheKey(req *image.ListImagesRequest) string {
	if req == nil {
		return "list_default"
	}

	return fmt.Sprintf("list_page_%d_pagesize_%d_tag_%s",
		req.Page, req.PageSize, req.Tag)
}

// UpdateImage modifies an existing image
func (s *ImageServiceImpl) UpdateImage(ctx context.Context, id int, req *image.UpdateImageRequest) (*image.Image, error) {
	if req == nil {
		return nil, fmt.Errorf("update request cannot be nil")
	}

	if err := s.validator.ValidateImageUpdate(ctx, id, req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	existing, err := s.imageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing image: %w", err)
	}

	tags, err := s.processTags(ctx, req.Tags)
	if err != nil {
		return nil, err
	}

	existing.Tags = tags
	existing.UpdatedAt = time.Now()

	if err := existing.Validate(); err != nil {
		return nil, fmt.Errorf("updated image validation failed: %w", err)
	}

	if err := s.imageRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update image in database: %w", err)
	}

	s.handlePostUpdate(ctx, id, existing)

	return existing, nil
}

func (s *ImageServiceImpl) handlePostUpdate(ctx context.Context, id int, img *image.Image) {
	if s.cache != nil {
		if err := s.cache.DeleteImage(ctx, id); err != nil {
			_ = err
		}
		if err := s.cache.InvalidateImageLists(ctx); err != nil {
			_ = err
		}
	}
	if s.eventPub != nil {
		if err := s.eventPub.PublishImageUpdated(ctx, img); err != nil {
			_ = err
		}
	}
}

// DeleteImage removes an image and its associated files
func (s *ImageServiceImpl) DeleteImage(ctx context.Context, id int) error {
	if err := s.validator.ValidateImageDeletion(ctx, id); err != nil {
		return fmt.Errorf("deletion validation failed: %w", err)
	}

	img, err := s.imageRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get image for deletion: %w", err)
	}

	if err := s.imageRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete image from database: %w", err)
	}

	s.cleanupStorage(ctx, img.StoragePath)
	s.handlePostDeletion(ctx, id)

	return nil
}

func (s *ImageServiceImpl) cleanupStorage(ctx context.Context, storagePath string) {
	if storagePath != "" {
		if err := s.storage.Delete(ctx, storagePath); err != nil {
			_ = err
		}
	}
}

func (s *ImageServiceImpl) handlePostDeletion(ctx context.Context, id int) {
	if s.cache != nil {
		if err := s.cache.DeleteImage(ctx, id); err != nil {
			_ = err
		}
		if err := s.cache.InvalidateImageLists(ctx); err != nil {
			_ = err
		}
	}
	if s.eventPub != nil {
		if err := s.eventPub.PublishImageDeleted(ctx, id); err != nil {
			_ = err
		}
	}
}

// DownloadImage provides access to the original image file
func (s *ImageServiceImpl) DownloadImage(ctx context.Context, id int) (io.ReadCloser, string, error) {
	// TODO: Implement image download workflow
	return nil, "", nil
}

// GenerateImageURL creates a URL for accessing an image
func (s *ImageServiceImpl) GenerateImageURL(ctx context.Context, id int, expiry int64) (string, error) {
	// TODO: Implement URL generation workflow
	return "", nil
}

// GetImageStats returns statistics about images
func (s *ImageServiceImpl) GetImageStats(ctx context.Context) (*image.ImageStats, error) {
	// TODO: Implement stats calculation
	return &image.ImageStats{
		TotalImages: 0,
		TotalSize:   0,
	}, nil
}
