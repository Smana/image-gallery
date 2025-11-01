package implementations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"image-gallery/internal/domain/image"
	"image-gallery/internal/platform/database"
)

// TagRepositoryImpl implements the image.TagRepository interface
type TagRepositoryImpl struct {
	db        *sql.DB
	dbTagRepo database.TagRepository
}

// NewTagRepository creates a new tag repository implementation
func NewTagRepository(db *sql.DB) image.TagRepository {
	return &TagRepositoryImpl{
		db:        db,
		dbTagRepo: database.NewTagRepository(db),
	}
}

// Create stores a new tag in the repository
func (r *TagRepositoryImpl) Create(ctx context.Context, tag *image.Tag) error {
	if tag == nil {
		return fmt.Errorf("tag cannot be nil")
	}

	query := `
		INSERT INTO tags (name, created_at)
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, tag.Name, tag.CreatedAt).Scan(&tag.ID)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// GetByID retrieves a tag by its ID
func (r *TagRepositoryImpl) GetByID(ctx context.Context, id int) (*image.Tag, error) {
	// TODO: Implement actual database query
	return nil, nil
}

// GetByName retrieves a tag by its name
func (r *TagRepositoryImpl) GetByName(ctx context.Context, name string) (*image.Tag, error) {
	if name == "" {
		return nil, fmt.Errorf("tag name cannot be empty")
	}

	query := `SELECT id, name, created_at FROM tags WHERE name = $1`
	var tag image.Tag
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&tag.ID,
		&tag.Name,
		&tag.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Tag not found, return nil (not an error for this method)
		}
		return nil, fmt.Errorf("failed to query tag by name: %w", err)
	}

	return &tag, nil
}

// List retrieves all tags or tags matching criteria
func (r *TagRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*image.Tag, error) {
	// TODO: Implement actual database query
	return []*image.Tag{}, nil
}

// Delete removes a tag from the repository
func (r *TagRepositoryImpl) Delete(ctx context.Context, id int) error {
	// TODO: Implement actual database deletion
	return nil
}

// GetOrCreate gets an existing tag or creates a new one
func (r *TagRepositoryImpl) GetOrCreate(ctx context.Context, name string) (*image.Tag, error) {
	if name == "" {
		return nil, fmt.Errorf("tag name cannot be empty")
	}

	// First try to get existing tag
	existing, err := r.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing tag: %w", err)
	}

	// If tag exists, return it
	if existing != nil {
		return existing, nil
	}

	// Tag doesn't exist, create it
	newTag := &image.Tag{
		Name:      name,
		CreatedAt: time.Now(),
	}

	// Validate the tag before creating
	if err := newTag.Validate(); err != nil {
		return nil, fmt.Errorf("tag validation failed: %w", err)
	}

	// Create the tag in database
	if err := r.Create(ctx, newTag); err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return newTag, nil
}

// GetPopularTags returns the most frequently used tags
func (r *TagRepositoryImpl) GetPopularTags(ctx context.Context, limit int) ([]*image.Tag, error) {
	// TODO: Implement popular tags query
	return []*image.Tag{}, nil
}

// ExistsByName checks if a tag with the given name exists
func (r *TagRepositoryImpl) ExistsByName(ctx context.Context, name string) (bool, error) {
	// TODO: Implement existence check
	return false, nil
}

// GetPredefinedTags returns all predefined tags
func (r *TagRepositoryImpl) GetPredefinedTags(ctx context.Context) ([]*image.Tag, error) {
	// Get predefined tags from database layer
	dbTags, err := r.dbTagRepo.GetPredefined(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get predefined tags: %w", err)
	}

	// Convert database tags to domain tags
	domainTags := make([]*image.Tag, len(dbTags))
	for i, dbTag := range dbTags {
		domainTags[i] = &image.Tag{
			ID:        dbTag.ID,
			Name:      dbTag.Name,
			CreatedAt: dbTag.CreatedAt,
		}
	}

	return domainTags, nil
}
