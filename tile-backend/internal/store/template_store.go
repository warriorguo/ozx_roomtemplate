package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"tile-backend/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TemplateStore defines the interface for template storage operations
type TemplateStore interface {
	Create(ctx context.Context, template model.Template) (*model.Template, error)
	List(ctx context.Context, limit, offset int, nameLike string) ([]model.TemplateSummary, int, error)
	Get(ctx context.Context, id string) (*model.Template, error)
	HealthCheck(ctx context.Context) error
}

// DBExecutor defines the interface for database operations we need
type DBExecutor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Ping(ctx context.Context) error
}

// PostgreSQLTemplateStore implements TemplateStore using PostgreSQL
type PostgreSQLTemplateStore struct {
	db DBExecutor
}

// NewPostgreSQLTemplateStore creates a new PostgreSQL template store
func NewPostgreSQLTemplateStore(db *pgxpool.Pool) *PostgreSQLTemplateStore {
	return &PostgreSQLTemplateStore{
		db: db,
	}
}

// NewPostgreSQLTemplateStoreWithExecutor creates a new PostgreSQL template store with custom executor
func NewPostgreSQLTemplateStoreWithExecutor(db DBExecutor) *PostgreSQLTemplateStore {
	return &PostgreSQLTemplateStore{
		db: db,
	}
}

// Create saves a new template to the database
func (s *PostgreSQLTemplateStore) Create(ctx context.Context, template model.Template) (*model.Template, error) {
	// Generate UUID if not provided
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}

	// Marshal payload to JSON
	payloadJSON, err := json.Marshal(template.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO room_templates (id, name, version, width, height, payload)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	var createdAt, updatedAt string
	err = s.db.QueryRow(ctx, query,
		template.ID,
		template.Name,
		template.Version,
		template.Width,
		template.Height,
		payloadJSON,
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert template: %w", err)
	}

	// Parse timestamps
	template.CreatedAt, err = parseTimestamp(createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	template.UpdatedAt, err = parseTimestamp(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &template, nil
}

// List retrieves templates with pagination and optional name filtering
func (s *PostgreSQLTemplateStore) List(ctx context.Context, limit, offset int, nameLike string) ([]model.TemplateSummary, int, error) {
	var whereClause string
	var args []interface{}
	argIndex := 1

	// Build WHERE clause for name filtering
	if nameLike != "" {
		whereClause = "WHERE name ILIKE $" + fmt.Sprintf("%d", argIndex)
		args = append(args, "%"+nameLike+"%")
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM room_templates " + whereClause
	var total int
	err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get paginated results
	listQuery := fmt.Sprintf(`
		SELECT id, name, version, width, height, created_at, updated_at
		FROM room_templates %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []model.TemplateSummary
	for rows.Next() {
		var template model.TemplateSummary
		var createdAt, updatedAt string

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Version,
			&template.Width,
			&template.Height,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan template: %w", err)
		}

		// Parse timestamps
		template.CreatedAt, err = parseTimestamp(createdAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse created_at: %w", err)
		}

		template.UpdatedAt, err = parseTimestamp(updatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return templates, total, nil
}

// Get retrieves a template by ID
func (s *PostgreSQLTemplateStore) Get(ctx context.Context, id string) (*model.Template, error) {
	// Validate UUID format
	templateID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	query := `
		SELECT id, name, version, width, height, payload, created_at, updated_at
		FROM room_templates
		WHERE id = $1`

	var template model.Template
	var payloadJSON []byte
	var createdAt, updatedAt string

	err = s.db.QueryRow(ctx, query, templateID).Scan(
		&template.ID,
		&template.Name,
		&template.Version,
		&template.Width,
		&template.Height,
		&payloadJSON,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to query template: %w", err)
	}

	// Unmarshal payload JSON
	err = json.Unmarshal(payloadJSON, &template.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Parse timestamps
	template.CreatedAt, err = parseTimestamp(createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	template.UpdatedAt, err = parseTimestamp(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	return &template, nil
}

// HealthCheck verifies the database connection
func (s *PostgreSQLTemplateStore) HealthCheck(ctx context.Context) error {
	return s.db.Ping(ctx)
}

// parseTimestamp is a helper function to parse timestamp strings
func parseTimestamp(timestampStr string) (time.Time, error) {
	// PostgreSQL returns timestamps in RFC3339 format
	return time.Parse(time.RFC3339, timestampStr)
}