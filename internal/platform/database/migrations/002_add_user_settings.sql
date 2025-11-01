-- Create user_settings table for storing user customization preferences
CREATE TABLE IF NOT EXISTS user_settings (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) DEFAULT 'default' UNIQUE, -- 'default' for anonymous user settings
    background_image_id INTEGER REFERENCES images(id) ON DELETE SET NULL,
    background_image_url TEXT,
    background_style VARCHAR(50) DEFAULT 'cover' CHECK (background_style IN ('cover', 'contain', 'repeat')),
    background_opacity NUMERIC(3,2) DEFAULT 0.30 CHECK (background_opacity >= 0.0 AND background_opacity <= 1.0),
    font_family VARCHAR(100) DEFAULT 'system-ui',
    text_theme VARCHAR(20) DEFAULT 'light' CHECK (text_theme IN ('light', 'dark')),
    show_tags BOOLEAN DEFAULT true,
    show_dimensions BOOLEAN DEFAULT true,
    show_content_type BOOLEAN DEFAULT true,
    grid_columns INTEGER DEFAULT 5 CHECK (grid_columns >= 2 AND grid_columns <= 6),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for user_id lookups
CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);

-- Create index for background image lookups
CREATE INDEX IF NOT EXISTS idx_user_settings_background_image_id ON user_settings(background_image_id);

-- Add updated_at trigger for user_settings table
CREATE TRIGGER update_user_settings_updated_at
    BEFORE UPDATE ON user_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default settings record (user_id = 'default' for anonymous users)
INSERT INTO user_settings (
    user_id,
    background_style,
    background_opacity,
    font_family,
    text_theme,
    show_tags,
    show_dimensions,
    show_content_type,
    grid_columns
) VALUES (
    'default',
    'cover',
    0.30,
    'system-ui',
    'light',
    true,
    true,
    true,
    5
) ON CONFLICT (user_id) DO NOTHING;
