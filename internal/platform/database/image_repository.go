package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// imageRepository implements ImageRepository interface
type imageRepository struct {
	db *sql.DB
}

// NewImageRepository creates a new ImageRepository
func NewImageRepository(db *sql.DB) ImageRepository {
	return &imageRepository{db: db}
}

// Create inserts a new image record
func (r *imageRepository) Create(ctx context.Context, image *Image) error {
	query := `
		INSERT INTO images (
			filename, original_filename, content_type, file_size, 
			storage_path, thumbnail_path, width, height, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, uploaded_at, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		image.Filename,
		image.OriginalFilename,
		image.ContentType,
		image.FileSize,
		image.StoragePath,
		image.ThumbnailPath,
		image.Width,
		image.Height,
		image.Metadata,
	).Scan(
		&image.ID,
		&image.UploadedAt,
		&image.CreatedAt,
		&image.UpdatedAt,
	)

	return err
}

// GetByID retrieves an image by its ID
func (r *imageRepository) GetByID(ctx context.Context, id int) (*Image, error) {
	query := `
		SELECT id, filename, original_filename, content_type, file_size,
			   storage_path, thumbnail_path, width, height, uploaded_at,
			   metadata, created_at, updated_at
		FROM images WHERE id = $1
	`

	image := &Image{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&image.ID,
		&image.Filename,
		&image.OriginalFilename,
		&image.ContentType,
		&image.FileSize,
		&image.StoragePath,
		&image.ThumbnailPath,
		&image.Width,
		&image.Height,
		&image.UploadedAt,
		&image.Metadata,
		&image.CreatedAt,
		&image.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("image with ID %d not found", id)
	}

	return image, err
}

// GetByFilename retrieves an image by its filename
func (r *imageRepository) GetByFilename(ctx context.Context, filename string) (*Image, error) {
	query := `
		SELECT id, filename, original_filename, content_type, file_size,
			   storage_path, thumbnail_path, width, height, uploaded_at,
			   metadata, created_at, updated_at
		FROM images WHERE filename = $1
	`

	image := &Image{}
	err := r.db.QueryRowContext(ctx, query, filename).Scan(
		&image.ID,
		&image.Filename,
		&image.OriginalFilename,
		&image.ContentType,
		&image.FileSize,
		&image.StoragePath,
		&image.ThumbnailPath,
		&image.Width,
		&image.Height,
		&image.UploadedAt,
		&image.Metadata,
		&image.CreatedAt,
		&image.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("image with filename %s not found", filename)
	}

	return image, err
}

// GetByStoragePath retrieves an image by its storage path
func (r *imageRepository) GetByStoragePath(ctx context.Context, path string) (*Image, error) {
	query := `
		SELECT id, filename, original_filename, content_type, file_size,
			   storage_path, thumbnail_path, width, height, uploaded_at,
			   metadata, created_at, updated_at
		FROM images WHERE storage_path = $1
	`

	image := &Image{}
	err := r.db.QueryRowContext(ctx, query, path).Scan(
		&image.ID,
		&image.Filename,
		&image.OriginalFilename,
		&image.ContentType,
		&image.FileSize,
		&image.StoragePath,
		&image.ThumbnailPath,
		&image.Width,
		&image.Height,
		&image.UploadedAt,
		&image.Metadata,
		&image.CreatedAt,
		&image.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("image with storage path %s not found", path)
	}

	return image, err
}

// Update updates an existing image record
func (r *imageRepository) Update(ctx context.Context, image *Image) error {
	query := `
		UPDATE images SET
			filename = $2,
			original_filename = $3,
			content_type = $4,
			file_size = $5,
			storage_path = $6,
			thumbnail_path = $7,
			width = $8,
			height = $9,
			metadata = $10,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		image.ID,
		image.Filename,
		image.OriginalFilename,
		image.ContentType,
		image.FileSize,
		image.StoragePath,
		image.ThumbnailPath,
		image.Width,
		image.Height,
		image.Metadata,
	).Scan(&image.UpdatedAt)

	return err
}

// UpdateThumbnail updates just the thumbnail path for an image
func (r *imageRepository) UpdateThumbnail(ctx context.Context, id int, thumbnailPath string) error {
	query := `UPDATE images SET thumbnail_path = $2, updated_at = NOW() WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id, thumbnailPath)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image with ID %d not found", id)
	}

	return nil
}

// Delete removes an image record by ID
func (r *imageRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM images WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image with ID %d not found", id)
	}

	return nil
}

// DeleteByStoragePath removes an image record by storage path
func (r *imageRepository) DeleteByStoragePath(ctx context.Context, path string) error {
	query := `DELETE FROM images WHERE storage_path = $1`
	
	result, err := r.db.ExecContext(ctx, query, path)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image with storage path %s not found", path)
	}

	return nil
}

// List retrieves a paginated list of images
func (r *imageRepository) List(ctx context.Context, pagination PaginationParams, sort SortParams) ([]*Image, error) {
	pagination.Validate()
	
	query := fmt.Sprintf(`
		SELECT id, filename, original_filename, content_type, file_size,
			   storage_path, thumbnail_path, width, height, uploaded_at,
			   metadata, created_at, updated_at
		FROM images
		ORDER BY %s %s
		LIMIT $1 OFFSET $2
	`, sort.Field, sort.Order)

	return r.scanImages(ctx, query, pagination.Limit, pagination.Offset)
}

// ListByContentType retrieves images filtered by content type
func (r *imageRepository) ListByContentType(ctx context.Context, contentType string, pagination PaginationParams) ([]*Image, error) {
	pagination.Validate()
	
	query := `
		SELECT id, filename, original_filename, content_type, file_size,
			   storage_path, thumbnail_path, width, height, uploaded_at,
			   metadata, created_at, updated_at
		FROM images
		WHERE content_type = $1
		ORDER BY uploaded_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.scanImages(ctx, query, contentType, pagination.Limit, pagination.Offset)
}

// Search performs complex filtering and searching
func (r *imageRepository) Search(ctx context.Context, filters SearchFilters, pagination PaginationParams, sort SortParams) ([]*Image, error) {
	pagination.Validate()
	
	var conditions []string
	var args []interface{}
	argIndex := 1

	if len(filters.ContentTypes) > 0 {
		placeholders := make([]string, len(filters.ContentTypes))
		for i, contentType := range filters.ContentTypes {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, contentType)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("content_type IN (%s)", strings.Join(placeholders, ",")))
	}

	if filters.MinSize != nil {
		conditions = append(conditions, fmt.Sprintf("file_size >= $%d", argIndex))
		args = append(args, *filters.MinSize)
		argIndex++
	}

	if filters.MaxSize != nil {
		conditions = append(conditions, fmt.Sprintf("file_size <= $%d", argIndex))
		args = append(args, *filters.MaxSize)
		argIndex++
	}

	if filters.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("uploaded_at >= $%d", argIndex))
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("uploaded_at <= $%d", argIndex))
		args = append(args, *filters.EndDate)
		argIndex++
	}

	if filters.Filename != "" {
		conditions = append(conditions, fmt.Sprintf("(filename ILIKE $%d OR original_filename ILIKE $%d)", argIndex, argIndex+1))
		searchTerm := "%" + filters.Filename + "%"
		args = append(args, searchTerm, searchTerm)
		argIndex += 2
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, filename, original_filename, content_type, file_size,
			   storage_path, thumbnail_path, width, height, uploaded_at,
			   metadata, created_at, updated_at
		FROM images
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sort.Field, sort.Order, argIndex, argIndex+1)

	args = append(args, pagination.Limit, pagination.Offset)

	return r.scanImages(ctx, query, args...)
}

// GetByDateRange retrieves images within a date range
func (r *imageRepository) GetByDateRange(ctx context.Context, start, end time.Time, pagination PaginationParams) ([]*Image, error) {
	pagination.Validate()
	
	query := `
		SELECT id, filename, original_filename, content_type, file_size,
			   storage_path, thumbnail_path, width, height, uploaded_at,
			   metadata, created_at, updated_at
		FROM images
		WHERE uploaded_at >= $1 AND uploaded_at <= $2
		ORDER BY uploaded_at DESC
		LIMIT $3 OFFSET $4
	`

	return r.scanImages(ctx, query, start, end, pagination.Limit, pagination.Offset)
}

// GetRecent retrieves recently uploaded images
func (r *imageRepository) GetRecent(ctx context.Context, since time.Time, limit int) ([]*Image, error) {
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	
	query := `
		SELECT id, filename, original_filename, content_type, file_size,
			   storage_path, thumbnail_path, width, height, uploaded_at,
			   metadata, created_at, updated_at
		FROM images
		WHERE uploaded_at >= $1
		ORDER BY uploaded_at DESC
		LIMIT $2
	`

	return r.scanImages(ctx, query, since, limit)
}

// GetLargest retrieves images ordered by file size
func (r *imageRepository) GetLargest(ctx context.Context, pagination PaginationParams) ([]*Image, error) {
	pagination.Validate()
	
	query := `
		SELECT id, filename, original_filename, content_type, file_size,
			   storage_path, thumbnail_path, width, height, uploaded_at,
			   metadata, created_at, updated_at
		FROM images
		ORDER BY file_size DESC
		LIMIT $1 OFFSET $2
	`

	return r.scanImages(ctx, query, pagination.Limit, pagination.Offset)
}

// Count returns the total number of images
func (r *imageRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM images").Scan(&count)
	return count, err
}

// CountByContentType returns the count of images by content type
func (r *imageRepository) CountByContentType(ctx context.Context, contentType string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM images WHERE content_type = $1", contentType).Scan(&count)
	return count, err
}

// GetStats returns aggregate statistics about images
func (r *imageRepository) GetStats(ctx context.Context) (*ImageStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_images,
			COUNT(DISTINCT content_type) as content_types,
			COALESCE(SUM(file_size), 0) as total_size,
			COALESCE(AVG(file_size), 0) as avg_size,
			COALESCE(MAX(file_size), 0) as max_size,
			COALESCE(MIN(file_size), 0) as min_size
		FROM images
	`

	stats := &ImageStats{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalImages,
		&stats.ContentTypes,
		&stats.TotalSize,
		&stats.AverageSize,
		&stats.MaxSize,
		&stats.MinSize,
	)

	return stats, err
}

// GetWithTags retrieves images with their associated tags
func (r *imageRepository) GetWithTags(ctx context.Context, pagination PaginationParams, sort SortParams) ([]*Image, error) {
	pagination.Validate()
	
	query := fmt.Sprintf(`
		SELECT 
			i.id, i.filename, i.original_filename, i.content_type, i.file_size,
			i.storage_path, i.thumbnail_path, i.width, i.height, i.uploaded_at,
			i.metadata, i.created_at, i.updated_at,
			COALESCE(
				json_agg(
					json_build_object(
						'id', t.id, 
						'name', t.name, 
						'description', t.description,
						'color', t.color,
						'created_at', t.created_at
					) ORDER BY t.name
				) FILTER (WHERE t.id IS NOT NULL), 
				'[]'
			) as tags
		FROM images i
		LEFT JOIN image_tags it ON i.id = it.image_id
		LEFT JOIN tags t ON it.tag_id = t.id
		GROUP BY i.id, i.filename, i.original_filename, i.content_type, i.file_size, 
				 i.storage_path, i.thumbnail_path, i.width, i.height, i.uploaded_at, 
				 i.metadata, i.created_at, i.updated_at
		ORDER BY i.%s %s
		LIMIT $1 OFFSET $2
	`, sort.Field, sort.Order)

	rows, err := r.db.QueryContext(ctx, query, pagination.Limit, pagination.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*Image
	for rows.Next() {
		image := &Image{}
		var tagsJSON []byte

		err := rows.Scan(
			&image.ID,
			&image.Filename,
			&image.OriginalFilename,
			&image.ContentType,
			&image.FileSize,
			&image.StoragePath,
			&image.ThumbnailPath,
			&image.Width,
			&image.Height,
			&image.UploadedAt,
			&image.Metadata,
			&image.CreatedAt,
			&image.UpdatedAt,
			&tagsJSON,
		)
		if err != nil {
			return nil, err
		}

		// Parse tags JSON
		if len(tagsJSON) > 0 && string(tagsJSON) != "[]" {
			var tags []Tag
			if err := json.Unmarshal(tagsJSON, &tags); err == nil {
				image.Tags = tags
			}
		}

		images = append(images, image)
	}

	return images, rows.Err()
}

// GetByTags retrieves images that have specific tags
func (r *imageRepository) GetByTags(ctx context.Context, tags []string, matchAll bool, pagination PaginationParams) ([]*Image, error) {
	if len(tags) == 0 {
		return []*Image{}, nil
	}
	
	pagination.Validate()

	var query string
	args := []interface{}{tags, pagination.Limit, pagination.Offset}

	if matchAll {
		// Images must have ALL specified tags
		query = `
			SELECT DISTINCT i.id, i.filename, i.original_filename, i.content_type, i.file_size,
				   i.storage_path, i.thumbnail_path, i.width, i.height, i.uploaded_at,
				   i.metadata, i.created_at, i.updated_at
			FROM images i
			WHERE EXISTS (
				SELECT 1 FROM image_tags it 
				INNER JOIN tags t ON it.tag_id = t.id 
				WHERE it.image_id = i.id AND t.name = ANY($1::text[])
				GROUP BY it.image_id 
				HAVING COUNT(DISTINCT t.name) = $4
			)
			ORDER BY i.uploaded_at DESC
			LIMIT $2 OFFSET $3
		`
		args = append(args, len(tags))
	} else {
		// Images must have ANY of the specified tags
		query = `
			SELECT DISTINCT i.id, i.filename, i.original_filename, i.content_type, i.file_size,
				   i.storage_path, i.thumbnail_path, i.width, i.height, i.uploaded_at,
				   i.metadata, i.created_at, i.updated_at
			FROM images i
			INNER JOIN image_tags it ON i.id = it.image_id
			INNER JOIN tags t ON it.tag_id = t.id
			WHERE t.name = ANY($1::text[])
			ORDER BY i.uploaded_at DESC
			LIMIT $2 OFFSET $3
		`
	}

	return r.scanImages(ctx, query, args...)
}

// scanImages is a helper method to scan multiple image records
func (r *imageRepository) scanImages(ctx context.Context, query string, args ...interface{}) ([]*Image, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*Image
	for rows.Next() {
		image := &Image{}
		err := rows.Scan(
			&image.ID,
			&image.Filename,
			&image.OriginalFilename,
			&image.ContentType,
			&image.FileSize,
			&image.StoragePath,
			&image.ThumbnailPath,
			&image.Width,
			&image.Height,
			&image.UploadedAt,
			&image.Metadata,
			&image.CreatedAt,
			&image.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, rows.Err()
}