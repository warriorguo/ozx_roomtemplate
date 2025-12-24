-- Add thumbnail column to room_templates table
ALTER TABLE room_templates 
ADD COLUMN thumbnail TEXT;

-- Add index on thumbnail column for potential future queries
CREATE INDEX IF NOT EXISTS idx_room_templates_thumbnail_exists 
ON room_templates (id) 
WHERE thumbnail IS NOT NULL;