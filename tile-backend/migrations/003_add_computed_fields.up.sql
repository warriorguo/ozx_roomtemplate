-- Add computed fields for better querying
ALTER TABLE room_templates
ADD COLUMN IF NOT EXISTS walkable_ratio numeric(5,4),
ADD COLUMN IF NOT EXISTS room_type text,
ADD COLUMN IF NOT EXISTS room_category text,
ADD COLUMN IF NOT EXISTS room_attributes jsonb,
ADD COLUMN IF NOT EXISTS doors_connected jsonb,
ADD COLUMN IF NOT EXISTS static_count int,
ADD COLUMN IF NOT EXISTS chaser_count int,
ADD COLUMN IF NOT EXISTS zoner_count int,
ADD COLUMN IF NOT EXISTS dps_count int,
ADD COLUMN IF NOT EXISTS mobair_count int,
ADD COLUMN IF NOT EXISTS stage_type text;

-- Drop legacy columns if they exist (renamed: turret→chaser, mobground→zoner)
ALTER TABLE room_templates
DROP COLUMN IF EXISTS turret_count,
DROP COLUMN IF EXISTS mobground_count;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_room_templates_walkable_ratio
    ON room_templates (walkable_ratio);

CREATE INDEX IF NOT EXISTS idx_room_templates_room_type
    ON room_templates (room_type);

CREATE INDEX IF NOT EXISTS idx_room_templates_room_category
    ON room_templates (room_category);

CREATE INDEX IF NOT EXISTS idx_room_templates_static_count
    ON room_templates (static_count);

CREATE INDEX IF NOT EXISTS idx_room_templates_chaser_count
    ON room_templates (chaser_count);

CREATE INDEX IF NOT EXISTS idx_room_templates_zoner_count
    ON room_templates (zoner_count);

CREATE INDEX IF NOT EXISTS idx_room_templates_dps_count
    ON room_templates (dps_count);

CREATE INDEX IF NOT EXISTS idx_room_templates_mobair_count
    ON room_templates (mobair_count);

CREATE INDEX IF NOT EXISTS idx_room_templates_stage_type
    ON room_templates (stage_type);

-- Drop legacy indexes
DROP INDEX IF EXISTS idx_room_templates_turret_count;
DROP INDEX IF EXISTS idx_room_templates_mobground_count;

-- GIN indexes for JSONB queries
CREATE INDEX IF NOT EXISTS idx_room_templates_room_attributes_gin
    ON room_templates USING gin (room_attributes);

CREATE INDEX IF NOT EXISTS idx_room_templates_doors_connected_gin
    ON room_templates USING gin (doors_connected);

-- Column comments
COMMENT ON COLUMN room_templates.walkable_ratio IS 'Ratio of walkable tiles (ground=1) to total tiles (0.0000-1.0000)';
COMMENT ON COLUMN room_templates.room_type IS 'Room shape: full, bridge, or platform (internal name kept for DB compat)';
COMMENT ON COLUMN room_templates.room_category IS 'Room category: normal, basement, test, cave';
COMMENT ON COLUMN room_templates.room_attributes IS 'Room attributes: {"boss": bool, "elite": bool, "mob": bool, "treasure": bool, "teleport": bool, "story": bool}';
COMMENT ON COLUMN room_templates.doors_connected IS 'Door connectivity: {"top": bool, "right": bool, "bottom": bool, "left": bool}';
COMMENT ON COLUMN room_templates.static_count IS 'Number of static tiles (static layer cells with value 1)';
COMMENT ON COLUMN room_templates.chaser_count IS 'Number of chaser tiles (chaser layer cells with value 1)';
COMMENT ON COLUMN room_templates.zoner_count IS 'Number of zoner tiles (zoner layer cells with value 1)';
COMMENT ON COLUMN room_templates.dps_count IS 'Number of DPS tiles (dps layer cells with value 1)';
COMMENT ON COLUMN room_templates.mobair_count IS 'Number of mob air tiles (mobAir layer cells with value 1)';
COMMENT ON COLUMN room_templates.stage_type IS 'Stage type: teaching, building, pressure, peak, release, boss';
