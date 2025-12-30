#!/bin/bash

# Apply all database migrations for tile-backend
# Usage: ./apply_migrations.sh [DATABASE_URL]

set -e

# Database connection
DATABASE_URL="${1:-postgres://postgres:pass@localhost:5432/appdb?sslmode=disable}"

# Parse connection string
DB_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
DB_PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
DB_NAME=$(echo $DATABASE_URL | sed -n 's/.*\/\([^?]*\).*/\1/p')
DB_USER=$(echo $DATABASE_URL | sed -n 's/.*\/\/\([^:]*\):.*/\1/p')
DB_PASS=$(echo $DATABASE_URL | sed -n 's/.*:\/\/[^:]*:\([^@]*\)@.*/\1/p')

echo "=========================================="
echo "Applying Database Migrations"
echo "=========================================="
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "User: $DB_USER"
echo ""

export PGPASSWORD="$DB_PASS"

# Function to run SQL file
run_migration() {
    local file=$1
    local description=$2
    echo "Running: $file"
    echo "  $description"
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file" > /dev/null 2>&1; then
        echo "  ✓ Success"
    else
        echo "  ✗ Failed"
        echo "  Note: If table already exists, this is expected."
    fi
    echo ""
}

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "Error: psql command not found!"
    echo "Please install PostgreSQL client tools."
    echo ""
    echo "On macOS: brew install postgresql"
    echo "On Ubuntu/Debian: sudo apt-get install postgresql-client"
    echo "On RHEL/CentOS: sudo yum install postgresql"
    exit 1
fi

# Check database connection
echo "Testing database connection..."
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
    echo "✗ Failed to connect to database"
    echo "Please check your database connection settings."
    exit 1
fi
echo "✓ Database connection successful"
echo ""

# Change to script directory
cd "$(dirname "$0")"

# Apply migrations in order
echo "Applying migrations..."
echo ""

run_migration "migrations/001_create_room_templates.up.sql" \
    "Create room_templates table and basic structure"

run_migration "migrations/002_add_thumbnail.up.sql" \
    "Add thumbnail column for template previews"

run_migration "migrations/003_add_computed_fields.up.sql" \
    "Add computed fields for advanced querying"

echo "=========================================="
echo "Migration completed!"
echo "=========================================="
echo ""

# Show current table structure
echo "Current table structure:"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "\d room_templates"
echo ""

# Show statistics
echo "Database statistics:"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT
    COUNT(*) as total_templates,
    COUNT(walkable_ratio) as templates_with_stats
FROM room_templates;
"

unset PGPASSWORD
