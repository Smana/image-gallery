package implementations

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"image-gallery/internal/domain/image"
)

// ValidationServiceImpl implements the image.ValidationService interface
type ValidationServiceImpl struct{}

// NewValidationService creates a new validation service implementation
func NewValidationService() image.ValidationService {
	return &ValidationServiceImpl{}
}

// ValidateImageUpload validates an image upload request
func (v *ValidationServiceImpl) ValidateImageUpload(ctx context.Context, req *image.CreateImageRequest) error {
	if req == nil {
		return errors.New("create request cannot be nil")
	}

	if err := v.validateBasicFields(req); err != nil {
		return err
	}

	if err := v.validateContentType(req.ContentType); err != nil {
		return err
	}

	if err := v.validateFileSize(req.FileSize); err != nil {
		return err
	}

	if err := v.validateDimensions(req.Width, req.Height); err != nil {
		return err
	}

	if err := v.validateTags(req.Tags); err != nil {
		return err
	}

	return nil
}

func (v *ValidationServiceImpl) validateBasicFields(req *image.CreateImageRequest) error {
	if strings.TrimSpace(req.OriginalFilename) == "" {
		return errors.New("filename cannot be empty")
	}
	if strings.TrimSpace(req.ContentType) == "" {
		return errors.New("content type cannot be empty")
	}
	return nil
}

func (v *ValidationServiceImpl) validateContentType(contentType string) error {
	supportedTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif",
		"image/webp", "image/bmp", "image/tiff",
	}
	for _, supportedType := range supportedTypes {
		if contentType == supportedType {
			return nil
		}
	}
	return fmt.Errorf("unsupported content type: %s", contentType)
}

func (v *ValidationServiceImpl) validateFileSize(fileSize int64) error {
	const maxFileSize = 50 * 1024 * 1024 // 50MB
	if fileSize <= 0 {
		return errors.New("file size must be greater than 0")
	}
	if fileSize > maxFileSize {
		return fmt.Errorf("file size too large: %d bytes, maximum allowed: %d bytes", fileSize, maxFileSize)
	}
	return nil
}

func (v *ValidationServiceImpl) validateDimensions(width, height *int) error {
	if width != nil && *width <= 0 {
		return errors.New("image width must be greater than 0")
	}
	if height != nil && *height <= 0 {
		return errors.New("image height must be greater than 0")
	}
	return nil
}

func (v *ValidationServiceImpl) validateTags(tags []string) error {
	if len(tags) > 20 {
		return errors.New("too many tags: maximum 20 tags allowed")
	}
	for _, tag := range tags {
		if strings.TrimSpace(tag) == "" {
			return errors.New("tag names cannot be empty")
		}
		if len(tag) > 50 {
			return fmt.Errorf("tag name too long: '%s', maximum 50 characters allowed", tag)
		}
	}
	return nil
}

// ValidateImageUpdate validates an image update request
func (v *ValidationServiceImpl) ValidateImageUpdate(ctx context.Context, id int, req *image.UpdateImageRequest) error {
	// TODO: Implement update validation rules
	return nil
}

// ValidateImageDeletion validates if an image can be deleted
func (v *ValidationServiceImpl) ValidateImageDeletion(ctx context.Context, id int) error {
	// TODO: Implement deletion validation rules
	// - Check if image is referenced elsewhere
	// - Validate permissions
	return nil
}

// ValidateTagOperation validates tag operations
func (v *ValidationServiceImpl) ValidateTagOperation(ctx context.Context, operation string, tagID int) error {
	// TODO: Implement tag operation validation
	return nil
}
