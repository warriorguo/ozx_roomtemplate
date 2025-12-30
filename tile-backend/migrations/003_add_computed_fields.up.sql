-- Add computed fields for better querying
ALTER TABLE room_templates
ADD COLUMN walkable_ratio numeric(5,4),
ADD COLUMN room_type text,
ADD COLUMN room_attributes jsonb,
ADD COLUMN doors_connected jsonb,
ADD COLUMN static_count int,
ADD COLUMN turret_count int,
ADD COLUMN mobground_count int,
ADD COLUMN mobair_count int;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_room_templates_walkable_ratio
    ON room_templates (walkable_ratio);

CREATE INDEX IF NOT EXISTS idx_room_templates_room_type
    ON room_templates (room_type);

CREATE INDEX IF NOT EXISTS idx_room_templates_static_count
    ON room_templates (static_count);

CREATE INDEX IF NOT EXISTS idx_room_templates_turret_count
    ON room_templates (turret_count);

CREATE INDEX IF NOT EXISTS idx_room_templates_mobground_count
    ON room_templates (mobground_count);

CREATE INDEX IF NOT EXISTS idx_room_templates_mobair_count
    ON room_templates (mobair_count);

-- GIN indexes for JSONB queries
CREATE INDEX IF NOT EXISTS idx_room_templates_room_attributes_gin
    ON room_templates USING gin (room_attributes);

CREATE INDEX IF NOT EXISTS idx_room_templates_doors_connected_gin
    ON room_templates USING gin (doors_connected);

-- Add comment explaining the fields
COMMENT ON COLUMN room_templates.walkable_ratio IS 'Ratio of walkable tiles (ground=1) to total tiles (0.0000-1.0000)';
COMMENT ON COLUMN room_templates.room_type IS 'Type of room: full, bridge, or platform';
COMMENT ON COLUMN room_templates.room_attributes IS 'Room attributes: {"boss": bool, "elite": bool, "mob": bool, "treasure": bool, "teleport": bool, "story": bool}';
COMMENT ON COLUMN room_templates.doors_connected IS 'Door connectivity: {"top": bool, "right": bool, "bottom": bool, "left": bool}';
COMMENT ON COLUMN room_templates.static_count IS 'Number of static tiles (static layer cells with value 1)';
COMMENT ON COLUMN room_templates.turret_count IS 'Number of turret tiles (turret layer cells with value 1)';
COMMENT ON COLUMN room_templates.mobground_count IS 'Number of mob ground tiles (mobGround layer cells with value 1)';
COMMENT ON COLUMN room_templates.mobair_count IS 'Number of mob air tiles (mobAir layer cells with value 1)';
