package implementations

import (
	"context"
	"encoding/json"
	"fmt"

	"image-gallery/internal/domain/image"
	"image-gallery/internal/platform/database"
)

// ImageRepositoryAdapter adapts database ImageRepository to domain Repository interface
type ImageRepositoryAdapter struct {
	dbRepo database.ImageRepository
}

// NewImageRepositoryAdapter creates a domain repository adapter
func NewImageRepositoryAdapter(dbRepo database.ImageRepository) image.Repository {
	return &ImageRepositoryAdapter{dbRepo: dbRepo}
}

func (a *ImageRepositoryAdapter) Create(ctx context.Context, img *image.Image) error {
	dbImage := &database.Image{
		Filename:         img.Filename,
		OriginalFilename: img.OriginalFilename,
		ContentType:      img.ContentType,
		FileSize:         img.FileSize,
		StoragePath:      img.StoragePath,
		ThumbnailPath:    img.ThumbnailPath,
		Width:            img.Width,
		Height:           img.Height,
		UploadedAt:       img.UploadedAt,
		Metadata:         database.Metadata{},
	}

	// Convert domain metadata to database metadata if present
	if img.Metadata != nil {
		if err := json.Unmarshal(img.Metadata, &dbImage.Metadata); err != nil {
			return fmt.Errorf("failed to convert metadata: %w", err)
		}
	}

	if err := a.dbRepo.Create(ctx, dbImage); err != nil {
		return err
	}

	// Update the domain image with generated values
	img.ID = dbImage.ID
	img.CreatedAt = dbImage.CreatedAt
	img.UpdatedAt = dbImage.UpdatedAt
	img.UploadedAt = dbImage.UploadedAt

	return nil
}

func (a *ImageRepositoryAdapter) GetByID(ctx context.Context, id int) (*image.Image, error) {
	dbImage, err := a.dbRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return a.convertToBaseImage(dbImage), nil
}

func (a *ImageRepositoryAdapter) List(ctx context.Context, req *image.ListImagesRequest) (*image.ListImagesResponse, error) {
	if req.Tag != "" {
		// Use tag-specific search
		pagination := database.PaginationParams{
			Limit:  req.PageSize,
			Offset: req.GetOffset(),
		}
		
		tags := []string{req.Tag}
		dbImages, err := a.dbRepo.GetByTags(ctx, tags, false, pagination)
		if err != nil {
			return nil, err
		}

		// Get total count for this tag
		count, err := a.CountByTag(ctx, req.Tag)
		if err != nil {
			return nil, err
		}

		images := make([]image.Image, len(dbImages))
		for i, dbImg := range dbImages {
			images[i] = *a.convertToBaseImage(dbImg)
		}

		response := &image.ListImagesResponse{
			Images:     images,
			TotalCount: count,
			Page:       req.Page,
			PageSize:   req.PageSize,
		}
		response.CalculateTotalPages()

		return response, nil
	}

	// Regular list without tag filter
	pagination := database.PaginationParams{
		Limit:  req.PageSize,
		Offset: req.GetOffset(),
	}
	
	sort := database.SortParams{
		Field: "created_at",
		Order: "DESC",
	}

	dbImages, err := a.dbRepo.List(ctx, pagination, sort)
	if err != nil {
		return nil, err
	}

	// Get total count
	totalCount, err := a.dbRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	images := make([]image.Image, len(dbImages))
	for i, dbImg := range dbImages {
		images[i] = *a.convertToBaseImage(dbImg)
	}

	response := &image.ListImagesResponse{
		Images:     images,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	response.CalculateTotalPages()

	return response, nil
}

func (a *ImageRepositoryAdapter) Update(ctx context.Context, img *image.Image) error {
	dbImage := &database.Image{
		ID:               img.ID,
		Filename:         img.Filename,
		OriginalFilename: img.OriginalFilename,
		ContentType:      img.ContentType,
		FileSize:         img.FileSize,
		StoragePath:      img.StoragePath,
		ThumbnailPath:    img.ThumbnailPath,
		Width:            img.Width,
		Height:           img.Height,
		Metadata:         database.Metadata{},
	}

	// Convert domain metadata to database metadata if present
	if img.Metadata != nil {
		if err := json.Unmarshal(img.Metadata, &dbImage.Metadata); err != nil {
			return fmt.Errorf("failed to convert metadata: %w", err)
		}
	}

	if err := a.dbRepo.Update(ctx, dbImage); err != nil {
		return err
	}

	img.UpdatedAt = dbImage.UpdatedAt
	return nil
}

func (a *ImageRepositoryAdapter) Delete(ctx context.Context, id int) error {
	return a.dbRepo.Delete(ctx, id)
}

func (a *ImageRepositoryAdapter) GetByFilename(ctx context.Context, filename string) (*image.Image, error) {
	dbImage, err := a.dbRepo.GetByFilename(ctx, filename)
	if err != nil {
		return nil, err
	}

	return a.convertToBaseImage(dbImage), nil
}

func (a *ImageRepositoryAdapter) ExistsByFilename(ctx context.Context, filename string) (bool, error) {
	_, err := a.dbRepo.GetByFilename(ctx, filename)
	if err != nil {
		if err.Error() == fmt.Sprintf("image with filename %s not found", filename) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (a *ImageRepositoryAdapter) CountByTag(ctx context.Context, tagName string) (int, error) {
	// This is a simplified implementation. In a real system, you'd want a dedicated count query
	// For now, we'll use a reasonable limit and count the results
	pagination := database.PaginationParams{
		Limit:  1000, // Reasonable limit for counting
		Offset: 0,
	}
	
	tags := []string{tagName}
	dbImages, err := a.dbRepo.GetByTags(ctx, tags, false, pagination)
	if err != nil {
		return 0, err
	}
	
	return len(dbImages), nil
}

// Helper method to convert database Image to domain Image
func (a *ImageRepositoryAdapter) convertToBaseImage(dbImg *database.Image) *image.Image {
	img := &image.Image{
		ID:               dbImg.ID,
		Filename:         dbImg.Filename,
		OriginalFilename: dbImg.OriginalFilename,
		ContentType:      dbImg.ContentType,
		FileSize:         dbImg.FileSize,
		StoragePath:      dbImg.StoragePath,
		ThumbnailPath:    dbImg.ThumbnailPath,
		Width:            dbImg.Width,
		Height:           dbImg.Height,
		UploadedAt:       dbImg.UploadedAt,
		CreatedAt:        dbImg.CreatedAt,
		UpdatedAt:        dbImg.UpdatedAt,
	}

	// Convert database metadata to domain metadata
	if len(dbImg.Metadata) > 0 {
		metadataBytes, err := json.Marshal(dbImg.Metadata)
		if err == nil {
			img.Metadata = metadataBytes
		}
	}

	return img
}