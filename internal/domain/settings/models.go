package settings

import (
	"fmt"
	"time"
)

// UserSettings represents user customization preferences for the gallery
type UserSettings struct {
	ID                 int       `json:"id" db:"id"`
	UserID             *string   `json:"user_id,omitempty" db:"user_id"`
	BackgroundImageID  *int      `json:"background_image_id,omitempty" db:"background_image_id"`
	BackgroundImageURL *string   `json:"background_image_url,omitempty" db:"background_image_url"`
	BackgroundStyle    string    `json:"background_style" db:"background_style"`
	BackgroundOpacity  float32   `json:"background_opacity" db:"background_opacity"`
	FontFamily         string    `json:"font_family" db:"font_family"`
	TextTheme          string    `json:"text_theme" db:"text_theme"` // "light" or "dark"
	ShowTags           bool      `json:"show_tags" db:"show_tags"`
	ShowDimensions     bool      `json:"show_dimensions" db:"show_dimensions"`
	ShowContentType    bool      `json:"show_content_type" db:"show_content_type"`
	GridColumns        int       `json:"grid_columns" db:"grid_columns"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// BackgroundStyle represents valid background style options
type BackgroundStyle string

const (
	BackgroundStyleCover   BackgroundStyle = "cover"
	BackgroundStyleContain BackgroundStyle = "contain"
	BackgroundStyleRepeat  BackgroundStyle = "repeat"
)

// TextTheme represents valid text theme options
type TextTheme string

const (
	TextThemeLight TextTheme = "light"
	TextThemeDark  TextTheme = "dark"
)

// FontFamily represents valid font family options
type FontFamily string

const (
	FontFamilySystemUI   FontFamily = "system-ui"
	FontFamilyArial      FontFamily = "Arial"
	FontFamilyRoboto     FontFamily = "Roboto"
	FontFamilyOpenSans   FontFamily = "Open Sans"
	FontFamilyLato       FontFamily = "Lato"
	FontFamilyMontserrat FontFamily = "Montserrat"
)

// GetSettingsRequest represents a request to get user settings
type GetSettingsRequest struct {
	UserID *string `json:"user_id,omitempty"`
}

// UpdateSettingsRequest represents a request to update user settings
type UpdateSettingsRequest struct {
	UserID             *string  `json:"user_id,omitempty"`
	BackgroundImageID  *int     `json:"background_image_id,omitempty"`
	BackgroundImageURL *string  `json:"background_image_url,omitempty"`
	BackgroundStyle    *string  `json:"background_style,omitempty"`
	BackgroundOpacity  *float32 `json:"background_opacity,omitempty"`
	FontFamily         *string  `json:"font_family,omitempty"`
	TextTheme          *string  `json:"text_theme,omitempty"`
	ShowTags           *bool    `json:"show_tags,omitempty"`
	ShowDimensions     *bool    `json:"show_dimensions,omitempty"`
	ShowContentType    *bool    `json:"show_content_type,omitempty"`
	GridColumns        *int     `json:"grid_columns,omitempty"`
}

// Validate validates the update settings request
//
//nolint:gocyclo // Comprehensive field validation with multiple checks
func (r *UpdateSettingsRequest) Validate() error {
	if r.BackgroundStyle != nil {
		style := BackgroundStyle(*r.BackgroundStyle)
		if style != BackgroundStyleCover && style != BackgroundStyleContain && style != BackgroundStyleRepeat {
			return fmt.Errorf("invalid background style: %s (must be cover, contain, or repeat)", *r.BackgroundStyle)
		}
	}

	if r.BackgroundOpacity != nil {
		if *r.BackgroundOpacity < 0.0 || *r.BackgroundOpacity > 1.0 {
			return fmt.Errorf("invalid background opacity: %f (must be between 0.0 and 1.0)", *r.BackgroundOpacity)
		}
	}

	if r.TextTheme != nil {
		theme := TextTheme(*r.TextTheme)
		if theme != TextThemeLight && theme != TextThemeDark {
			return fmt.Errorf("invalid text theme: %s (must be light or dark)", *r.TextTheme)
		}
	}

	if r.GridColumns != nil {
		if *r.GridColumns < 2 || *r.GridColumns > 6 {
			return fmt.Errorf("invalid grid columns: %d (must be between 2 and 6)", *r.GridColumns)
		}
	}

	return nil
}

// DefaultSettings returns the default user settings
func DefaultSettings() *UserSettings {
	defaultUserID := "default"
	return &UserSettings{
		UserID:            &defaultUserID,
		BackgroundStyle:   string(BackgroundStyleCover),
		BackgroundOpacity: 0.3,
		FontFamily:        string(FontFamilySystemUI),
		TextTheme:         string(TextThemeLight),
		ShowTags:          true,
		ShowDimensions:    true,
		ShowContentType:   true,
		GridColumns:       5,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}
