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
    EXECUTE FUNCTION update_updated_at_column();