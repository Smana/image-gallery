package settings

import "context"

// Repository defines the interface for settings data access
type Repository interface {
	// GetByUserID retrieves settings for a specific user
	// Returns nil if no settings exist for the user
	GetByUserID(ctx context.Context, userID *string) (*UserSettings, error)

	// GetDefault retrieves the default settings record
	GetDefault(ctx context.Context) (*UserSettings, error)

	// Create creates new user settings
	Create(ctx context.Context, settings *UserSettings) (*UserSettings, error)

	// Update updates existing user settings
	Update(ctx context.Context, settings *UserSettings) error

	// Upsert creates or updates user settings
	Upsert(ctx context.Context, settings *UserSettings) (*UserSettings, error)
}

// SettingsService defines the interface for settings business logic
type SettingsService interface {
	// GetSettings retrieves settings for a user (or default if not found)
	GetSettings(ctx context.Context, req *GetSettingsRequest) (*UserSettings, error)

	// UpdateSettings updates user settings
	UpdateSettings(ctx context.Context, req *UpdateSettingsRequest) (*UserSettings, error)

	// ResetSettings resets user settings to defaults
	ResetSettings(ctx context.Context, userID *string) (*UserSettings, error)
}
