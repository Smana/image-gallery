-- Create images table for storing image metadata
CREATE TABLE IF NOT EXISTS images (
    id SERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL CHECK (file_size >= 0),
    storage_path VARCHAR(500) NOT NULL UNIQUE,
    thumbnail_path VARCHAR(500),
    width INTEGER CHECK (width > 0),
    height INTEGER CHECK (height > 0),
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for efficient queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_images_filename ON images(filename);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_images_original_filename ON images(original_filename);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_images_content_type ON images(content_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_images_uploaded_at ON images(uploaded_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_images_file_size ON images(file_size DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_images_storage_path ON images(storage_path);

-- Create tags table for image categorization
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    color VARCHAR(7), -- Hex color code like #FF5733
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for tag searches
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tags_name_lower ON tags(LOWER(name));

-- Create many-to-many relationship between images and tags
CREATE TABLE IF NOT EXISTS image_tags (
    image_id INTEGER NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (image_id, tag_id)
);

-- Create indexes for efficient tag queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_image_tags_image_id ON image_tags(image_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_image_tags_tag_id ON image_tags(tag_id);

-- Create albums/collections table (optional enhancement)
CREATE TABLE IF NOT EXISTS albums (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    thumbnail_image_id INTEGER REFERENCES images(id) ON DELETE SET NULL,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create image_albums junction table
CREATE TABLE IF NOT EXISTS image_albums (
    image_id INTEGER NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    album_id INTEGER NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    position INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (image_id, album_id)
);

-- Create indexes for album queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_albums_name ON albums(name);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_albums_is_public ON albums(is_public);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_image_albums_album_id ON image_albums(album_id, position);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_image_albums_image_id ON image_albums(image_id);

-- Add updated_at trigger for images table
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_images_updated_at 
    BEFORE UPDATE ON images 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_albums_updated_at 
    BEFORE UPDATE ON albums 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();