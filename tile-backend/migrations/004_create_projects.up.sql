-- Create room_projects table (named to avoid conflict with other apps sharing the database)
CREATE TABLE IF NOT EXISTS room_projects (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text NOT NULL DEFAULT '',
    total_rooms int NOT NULL CHECK (total_rooms > 0),

    -- Room shape distribution (percentages, must sum to 100)
    shape_pct_full int NOT NULL DEFAULT 0 CHECK (shape_pct_full >= 0 AND shape_pct_full <= 100),
    shape_pct_bridge int NOT NULL DEFAULT 0 CHECK (shape_pct_bridge >= 0 AND shape_pct_bridge <= 100),
    shape_pct_platform int NOT NULL DEFAULT 0 CHECK (shape_pct_platform >= 0 AND shape_pct_platform <= 100),

    -- Door open distribution (bitmask string key "0"-"15" -> room count)
    door_distribution jsonb NOT NULL DEFAULT '{}',

    -- Stage type distribution (percentages, must sum to 100)
    stage_pct_start int NOT NULL DEFAULT 0,
    stage_pct_teaching int NOT NULL DEFAULT 0,
    stage_pct_building int NOT NULL DEFAULT 0,
    stage_pct_pressure int NOT NULL DEFAULT 0,
    stage_pct_peak int NOT NULL DEFAULT 0,
    stage_pct_release int NOT NULL DEFAULT 0,
    stage_pct_boss int NOT NULL DEFAULT 0,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT chk_shape_pct_sum CHECK (shape_pct_full + shape_pct_bridge + shape_pct_platform = 100),
    CONSTRAINT chk_stage_pct_sum CHECK (stage_pct_start + stage_pct_teaching + stage_pct_building + stage_pct_pressure + stage_pct_peak + stage_pct_release + stage_pct_boss = 100)
);

-- Add project_id FK to room_templates
ALTER TABLE room_templates ADD COLUMN IF NOT EXISTS project_id uuid REFERENCES room_projects(id) ON DELETE SET NULL;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_room_projects_created_at ON room_projects (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_room_projects_name ON room_projects (name);
CREATE INDEX IF NOT EXISTS idx_room_templates_project_id ON room_templates (project_id);

-- Reuse existing updated_at trigger function
CREATE TRIGGER update_room_projects_updated_at
    BEFORE UPDATE ON room_projects
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
