-- Remove project_id from room_templates
ALTER TABLE room_templates DROP COLUMN IF EXISTS project_id;

-- Drop trigger
DROP TRIGGER IF EXISTS update_room_projects_updated_at ON room_projects;

-- Drop room_projects table
DROP TABLE IF EXISTS room_projects;
