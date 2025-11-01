-- Add predefined tags system to support configurable allowed tags
-- This enables a curated tag selection UI for uploads and consistent tag management

-- Add columns to tags table for predefined tag system
ALTER TABLE tags ADD COLUMN IF NOT EXISTS is_predefined BOOLEAN DEFAULT false;
ALTER TABLE tags ADD COLUMN IF NOT EXISTS display_order INT DEFAULT 0;
ALTER TABLE tags ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
ALTER TABLE tags ADD COLUMN IF NOT EXISTS category VARCHAR(50);

-- Add index for efficient predefined tag queries
CREATE INDEX IF NOT EXISTS idx_tags_predefined ON tags(is_predefined, is_active, display_order);

-- Insert initial predefined tags organized by category
-- Colors are not stored here - they're generated algorithmically via hash function

-- Subject/Content category
INSERT INTO tags (name, description, is_predefined, is_active, display_order, category) VALUES
('landscape', 'Wide natural scenery and outdoor views', true, true, 1, 'subject'),
('portrait', 'People and individual subjects', true, true, 2, 'subject'),
('nature', 'Plants, animals, and natural elements', true, true, 3, 'subject'),
('urban', 'Cities, buildings, and urban environments', true, true, 4, 'subject'),
('wildlife', 'Animals in their natural habitat', true, true, 5, 'subject'),
('architecture', 'Buildings and structural designs', true, true, 6, 'subject'),
('food', 'Culinary subjects and meals', true, true, 7, 'subject'),
('travel', 'Travel destinations and experiences', true, true, 8, 'subject')
ON CONFLICT (name) DO UPDATE SET
    is_predefined = EXCLUDED.is_predefined,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    category = EXCLUDED.category,
    description = EXCLUDED.description;

-- Time/Lighting category
INSERT INTO tags (name, description, is_predefined, is_active, display_order, category) VALUES
('sunset', 'Images captured during sunset', true, true, 10, 'time'),
('sunrise', 'Images captured during sunrise', true, true, 11, 'time'),
('golden-hour', 'Warm lighting shortly after sunrise or before sunset', true, true, 12, 'time'),
('blue-hour', 'Twilight period with deep blue sky', true, true, 13, 'time'),
('night', 'Nighttime photography', true, true, 14, 'time')
ON CONFLICT (name) DO UPDATE SET
    is_predefined = EXCLUDED.is_predefined,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    category = EXCLUDED.category,
    description = EXCLUDED.description;

-- Style/Processing category
INSERT INTO tags (name, description, is_predefined, is_active, display_order, category) VALUES
('black-and-white', 'Monochrome imagery', true, true, 20, 'style'),
('vintage', 'Retro or aged aesthetic', true, true, 21, 'style'),
('minimalist', 'Simple, clean compositions', true, true, 22, 'style'),
('vibrant', 'Bold, saturated colors', true, true, 23, 'style'),
('moody', 'Dark, atmospheric imagery', true, true, 24, 'style'),
('abstract', 'Non-representational art', true, true, 25, 'style')
ON CONFLICT (name) DO UPDATE SET
    is_predefined = EXCLUDED.is_predefined,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    category = EXCLUDED.category,
    description = EXCLUDED.description;

-- Technique category
INSERT INTO tags (name, description, is_predefined, is_active, display_order, category) VALUES
('macro', 'Extreme close-up photography', true, true, 30, 'technique'),
('long-exposure', 'Extended shutter speed effects', true, true, 31, 'technique'),
('hdr', 'High dynamic range imaging', true, true, 32, 'technique'),
('panorama', 'Wide-angle or stitched images', true, true, 33, 'technique'),
('aerial', 'Drone or elevated perspective', true, true, 34, 'technique'),
('bokeh', 'Aesthetic out-of-focus areas', true, true, 35, 'technique')
ON CONFLICT (name) DO UPDATE SET
    is_predefined = EXCLUDED.is_predefined,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    category = EXCLUDED.category,
    description = EXCLUDED.description;
