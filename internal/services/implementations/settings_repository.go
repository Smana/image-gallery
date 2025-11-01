package implementations

import (
	"context"
	"database/sql"
	"fmt"

	"image-gallery/internal/domain/settings"
)

// SettingsRepositoryImpl implements the settings.Repository interface
type SettingsRepositoryImpl struct {
	db *sql.DB
}

// NewSettingsRepository creates a new settings repository implementation
func NewSettingsRepository(db *sql.DB) settings.Repository {
	return &SettingsRepositoryImpl{
		db: db,
	}
}

// GetByUserID retrieves settings for a specific user
func (r *SettingsRepositoryImpl) GetByUserID(ctx context.Context, userID *string) (*settings.UserSettings, error) {
	query := `
		SELECT id, user_id, background_image_id, background_image_url,
		       background_style, background_opacity, font_family, text_theme,
		       show_tags, show_dimensions, show_content_type, grid_columns,
		       created_at, updated_at
		FROM user_settings
		WHERE user_id = $1
	`

	var s settings.UserSettings
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&s.ID,
		&s.UserID,
		&s.BackgroundImageID,
		&s.BackgroundImageURL,
		&s.BackgroundStyle,
		&s.BackgroundOpacity,
		&s.FontFamily,
		&s.TextTheme,
		&s.ShowTags,
		&s.ShowDimensions,
		&s.ShowContentType,
		&s.GridColumns,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Settings not found, return nil (not an error)
		}
		return nil, fmt.Errorf("failed to query settings by user ID: %w", err)
	}

	return &s, nil
}

// GetDefault retrieves the default settings record (user_id = NULL)
func (r *SettingsRepositoryImpl) GetDefault(ctx context.Context) (*settings.UserSettings, error) {
	query := `
		SELECT id, user_id, background_image_id, background_image_url,
		       background_style, background_opacity, font_family, text_theme,
		       show_tags, show_dimensions, show_content_type, grid_columns,
		       created_at, updated_at
		FROM user_settings
		WHERE user_id IS NULL
		LIMIT 1
	`

	var s settings.UserSettings
	err := r.db.QueryRowContext(ctx, query).Scan(
		&s.ID,
		&s.UserID,
		&s.BackgroundImageID,
		&s.BackgroundImageURL,
		&s.BackgroundStyle,
		&s.BackgroundOpacity,
		&s.FontFamily,
		&s.TextTheme,
		&s.ShowTags,
		&s.ShowDimensions,
		&s.ShowContentType,
		&s.GridColumns,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// If no default settings exist, return programmatic defaults
			return settings.DefaultSettings(), nil
		}
		return nil, fmt.Errorf("failed to query default settings: %w", err)
	}

	return &s, nil
}

// Create creates new user settings
func (r *SettingsRepositoryImpl) Create(ctx context.Context, s *settings.UserSettings) (*settings.UserSettings, error) {
	if s == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	query := `
		INSERT INTO user_settings (
			user_id, background_image_id, background_image_url,
			background_style, background_opacity, font_family, text_theme,
			show_tags, show_dimensions, show_content_type, grid_columns,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		s.UserID,
		s.BackgroundImageID,
		s.BackgroundImageURL,
		s.BackgroundStyle,
		s.BackgroundOpacity,
		s.FontFamily,
		s.TextTheme,
		s.ShowTags,
		s.ShowDimensions,
		s.ShowContentType,
		s.GridColumns,
		s.CreatedAt,
		s.UpdatedAt,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create settings: %w", err)
	}

	return s, nil
}

// Update updates existing user settings
func (r *SettingsRepositoryImpl) Update(ctx context.Context, s *settings.UserSettings) error {
	if s == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	query := `
		UPDATE user_settings
		SET background_image_id = $2,
		    background_image_url = $3,
		    background_style = $4,
		    background_opacity = $5,
		    font_family = $6,
		    text_theme = $7,
		    show_tags = $8,
		    show_dimensions = $9,
		    show_content_type = $10,
		    grid_columns = $11,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		s.ID,
		s.BackgroundImageID,
		s.BackgroundImageURL,
		s.BackgroundStyle,
		s.BackgroundOpacity,
		s.FontFamily,
		s.TextTheme,
		s.ShowTags,
		s.ShowDimensions,
		s.ShowContentType,
		s.GridColumns,
	).Scan(&s.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}

// Upsert creates or updates user settings (insert on conflict update)
func (r *SettingsRepositoryImpl) Upsert(ctx context.Context, s *settings.UserSettings) (*settings.UserSettings, error) {
	if s == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	query := `
		INSERT INTO user_settings (
			user_id, background_image_id, background_image_url,
			background_style, background_opacity, font_family, text_theme,
			show_tags, show_dimensions, show_content_type, grid_columns,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (user_id)
		DO UPDATE SET
			background_image_id = EXCLUDED.background_image_id,
			background_image_url = EXCLUDED.background_image_url,
			background_style = EXCLUDED.background_style,
			background_opacity = EXCLUDED.background_opacity,
			font_family = EXCLUDED.font_family,
			text_theme = EXCLUDED.text_theme,
			show_tags = EXCLUDED.show_tags,
			show_dimensions = EXCLUDED.show_dimensions,
			show_content_type = EXCLUDED.show_content_type,
			grid_columns = EXCLUDED.grid_columns,
			updated_at = NOW()
		RETURNING id, user_id, background_image_id, background_image_url,
		          background_style, background_opacity, font_family, text_theme,
		          show_tags, show_dimensions, show_content_type, grid_columns,
		          created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		s.UserID,
		s.BackgroundImageID,
		s.BackgroundImageURL,
		s.BackgroundStyle,
		s.BackgroundOpacity,
		s.FontFamily,
		s.TextTheme,
		s.ShowTags,
		s.ShowDimensions,
		s.ShowContentType,
		s.GridColumns,
		s.CreatedAt,
		s.UpdatedAt,
	).Scan(
		&s.ID,
		&s.UserID,
		&s.BackgroundImageID,
		&s.BackgroundImageURL,
		&s.BackgroundStyle,
		&s.BackgroundOpacity,
		&s.FontFamily,
		&s.TextTheme,
		&s.ShowTags,
		&s.ShowDimensions,
		&s.ShowContentType,
		&s.GridColumns,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to upsert settings: %w", err)
	}

	return s, nil
}
