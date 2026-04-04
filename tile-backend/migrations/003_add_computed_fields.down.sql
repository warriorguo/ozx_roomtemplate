-- Drop indexes
DROP INDEX IF EXISTS idx_room_templates_walkable_ratio;
DROP INDEX IF EXISTS idx_room_templates_room_type;
DROP INDEX IF EXISTS idx_room_templates_room_category;
DROP INDEX IF EXISTS idx_room_templates_static_count;
DROP INDEX IF EXISTS idx_room_templates_chaser_count;
DROP INDEX IF EXISTS idx_room_templates_zoner_count;
DROP INDEX IF EXISTS idx_room_templates_dps_count;
DROP INDEX IF EXISTS idx_room_templates_mobair_count;
DROP INDEX IF EXISTS idx_room_templates_stage_type;
DROP INDEX IF EXISTS idx_room_templates_room_attributes_gin;
DROP INDEX IF EXISTS idx_room_templates_doors_connected_gin;

-- Drop columns
ALTER TABLE room_templates
DROP COLUMN IF EXISTS walkable_ratio,
DROP COLUMN IF EXISTS room_type,
DROP COLUMN IF EXISTS room_category,
DROP COLUMN IF EXISTS room_attributes,
DROP COLUMN IF EXISTS doors_connected,
DROP COLUMN IF EXISTS static_count,
DROP COLUMN IF EXISTS chaser_count,
DROP COLUMN IF EXISTS zoner_count,
DROP COLUMN IF EXISTS dps_count,
DROP COLUMN IF EXISTS mobair_count,
DROP COLUMN IF EXISTS stage_type;
