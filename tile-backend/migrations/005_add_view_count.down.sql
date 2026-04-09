DROP INDEX IF EXISTS idx_room_templates_view_count;
ALTER TABLE room_templates DROP COLUMN IF EXISTS view_count;
