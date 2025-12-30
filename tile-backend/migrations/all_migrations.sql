-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create room_templates table
CREATE TABLE IF NOT EXISTS room_templates (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text NOT NULL DEFAULT '',
    version int NOT NULL,
    width int NOT NULL,
    height int NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_room_templates_created_at
    ON room_templates (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_room_templates_name
    ON room_templates (name);

-- Optional: GIN index for JSONB queries (for future metadata filtering)
CREATE INDEX IF NOT EXISTS idx_room_templates_payload_gin
    ON room_templates USING gin (payload);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_room_templates_updated_at
    BEFORE UPDATE ON room_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();-- Add thumbnail column to room_templates table
ALTER TABLE room_templates 
ADD COLUMN thumbnail TEXT;

-- Add index on thumbnail column for potential future queries
CREATE INDEX IF NOT EXISTS idx_room_templates_thumbnail_exists 
ON room_templates (id) 
WHERE thumbnail IS NOT NULL;-- Add computed fields for better querying
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
