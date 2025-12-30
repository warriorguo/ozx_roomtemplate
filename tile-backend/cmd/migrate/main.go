package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Get database URL from environment or argument
	databaseURL := os.Getenv("DATABASE_URL")
	if len(os.Args) > 1 {
		databaseURL = os.Args[1]
	}
	if databaseURL == "" {
		databaseURL = "postgres://postgres:pass@localhost:5432/appdb?sslmode=disable"
	}

	log.Printf("Connecting to database...")

	// Connect to database
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Printf("✓ Connected successfully\n")

	// Get the migrations directory
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Try relative to the binary location
		migrationsDir = filepath.Join("..", "..", "migrations")
	}

	// List of migrations in order
	migrations := []struct {
		file        string
		description string
	}{
		{"001_create_room_templates.up.sql", "Create room_templates table"},
		{"002_add_thumbnail.up.sql", "Add thumbnail column"},
		{"003_add_computed_fields.up.sql", "Add computed fields for querying"},
	}

	log.Printf("\nApplying migrations...\n")
	log.Printf("==========================================\n\n")

	// Apply each migration
	for _, migration := range migrations {
		filePath := filepath.Join(migrationsDir, migration.file)

		log.Printf("Running: %s\n", migration.file)
		log.Printf("  %s\n", migration.description)

		// Read SQL file
		sqlBytes, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("  ⚠ Warning: Could not read file: %v\n", err)
			continue
		}

		// Execute SQL
		_, err = pool.Exec(context.Background(), string(sqlBytes))
		if err != nil {
			// Check if it's a "already exists" error
			if strings.Contains(err.Error(), "already exists") {
				log.Printf("  ⚠ Already exists (skipping)\n")
			} else {
				log.Printf("  ✗ Error: %v\n", err)
			}
		} else {
			log.Printf("  ✓ Success\n")
		}
		log.Printf("\n")
	}

	log.Printf("==========================================\n")
	log.Printf("Migration completed!\n\n")

	// Show table structure
	log.Printf("Current table columns:\n")
	rows, err := pool.Query(context.Background(), `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'room_templates'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Printf("Could not query table structure: %v\n", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var colName, dataType, nullable string
			if err := rows.Scan(&colName, &dataType, &nullable); err != nil {
				continue
			}
			log.Printf("  - %-25s %s\n", colName, dataType)
		}
	}

	// Show statistics
	log.Printf("\nDatabase statistics:\n")
	var totalCount, statsCount int
	err = pool.QueryRow(context.Background(), `
		SELECT
			COUNT(*) as total_templates,
			COUNT(walkable_ratio) as templates_with_stats
		FROM room_templates
	`).Scan(&totalCount, &statsCount)
	if err != nil {
		log.Printf("Could not query statistics: %v\n", err)
	} else {
		log.Printf("  Total templates: %d\n", totalCount)
		log.Printf("  Templates with computed fields: %d\n", statsCount)
	}

	log.Printf("\n✓ Done!\n")
}
