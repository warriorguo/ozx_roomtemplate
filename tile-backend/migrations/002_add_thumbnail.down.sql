-- Remove thumbnail column and index
DROP INDEX IF EXISTS idx_room_templates_thumbnail_exists;
ALTER TABLE room_templates DROP COLUMN IF EXISTS thumbnail;