-- Add view_count column to room_templates for gallery review tracking
ALTER TABLE room_templates ADD COLUMN IF NOT EXISTS view_count int NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_room_templates_view_count ON room_templates (view_count ASC);
