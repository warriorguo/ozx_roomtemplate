package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// Check command
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/migrate/main.go [up|down]")
	}

	command := os.Args[1]
	var migrationFile string

	switch command {
	case "up":
		migrationFile = "migrations/002_add_thumbnail.up.sql"
	case "down":
		migrationFile = "migrations/002_add_thumbnail.down.sql"
	default:
		log.Fatal("Command must be 'up' or 'down'")
	}

	// Read migration file
	sql, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("Unable to read migration file %s: %v\n", migrationFile, err)
	}

	// Execute migration
	_, err = pool.Exec(ctx, string(sql))
	if err != nil {
		log.Fatalf("Unable to execute migration: %v\n", err)
	}

	fmt.Printf("Migration %s executed successfully\n", command)
}