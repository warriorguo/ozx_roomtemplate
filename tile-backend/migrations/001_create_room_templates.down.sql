-- Drop trigger
DROP TRIGGER IF EXISTS update_room_templates_updated_at ON room_templates;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_room_templates_payload_gin;
DROP INDEX IF EXISTS idx_room_templates_name;
DROP INDEX IF EXISTS idx_room_templates_created_at;

-- Drop table
DROP TABLE IF EXISTS room_templates;