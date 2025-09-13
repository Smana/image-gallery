package implementations

import (
	"context"
	"fmt"
	"io"

	"image-gallery/internal/domain/image"
)

// ImageServiceImpl implements the image.ImageService interface
type ImageServiceImpl struct {
	imageRepo    image.Repository
	tagRepo      image.TagRepository
	storage      image.StorageService
	processor    image.ImageProcessor
	validator    image.ValidationService
	eventPub     image.EventPublisher // can be nil
	cache        image.CacheService   // can be nil
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
	// TODO: Implement full image creation workflow
	// 1. Validate the request
	// 2. Process and validate the image data
	// 3. Store the image file
	// 4. Create database record
	// 5. Publish events
	return nil, nil
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
			// Log cache error but don't fail the request
			// In a real app, you might want to use a proper logger
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
			// Log cache error but don't fail the request
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
	// TODO: Implement full image update workflow
	return nil, nil
}

// DeleteImage removes an image and its associated files
func (s *ImageServiceImpl) DeleteImage(ctx context.Context, id int) error {
	// TODO: Implement full image deletion workflow
	// 1. Validate deletion is allowed
	// 2. Remove from storage
	// 3. Remove from database
	// 4. Invalidate cache
	// 5. Publish events
	
	// For now, just invalidate the cache
	if s.cache != nil {
		// Remove the specific image from cache
		if err := s.cache.DeleteImage(ctx, id); err != nil {
			// Log error but don't fail deletion
		}
		
		// Invalidate all image lists since they might contain this image
		if err := s.cache.InvalidateImageLists(ctx); err != nil {
			// Log error but don't fail deletion
		}
	}
	
	return nil
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