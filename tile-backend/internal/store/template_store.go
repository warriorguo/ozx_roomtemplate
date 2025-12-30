package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"tile-backend/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TemplateStore defines the interface for template storage operations
type TemplateStore interface {
	Create(ctx context.Context, template model.Template) (*model.Template, error)
	List(ctx context.Context, params model.ListTemplatesQueryParams) ([]model.TemplateSummary, int, error)
	Get(ctx context.Context, id string) (*model.Template, error)
	Delete(ctx context.Context, id string) error
	HealthCheck(ctx context.Context) error
}

// DBExecutor defines the interface for database operations we need
type DBExecutor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
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

	// Compute stats before saving
	model.ComputeTemplateStats(&template)

	// Marshal payload to JSON
	payloadJSON, err := json.Marshal(template.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Marshal room attributes
	roomAttributesJSON, err := model.SerializeRoomAttributes(template.RoomAttributes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal room attributes: %w", err)
	}

	// Marshal doors connected
	doorsConnectedJSON, err := model.SerializeDoorsConnected(template.DoorsConnected)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal doors connected: %w", err)
	}

	query := `
		INSERT INTO room_templates (
			id, name, version, width, height, payload, thumbnail,
			walkable_ratio, room_type, room_attributes, doors_connected,
			static_count, turret_count, mobground_count, mobair_count
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING created_at, updated_at`

	err = s.db.QueryRow(ctx, query,
		template.ID,
		template.Name,
		template.Version,
		template.Width,
		template.Height,
		payloadJSON,
		template.Thumbnail,
		template.WalkableRatio,
		template.RoomType,
		roomAttributesJSON,
		doorsConnectedJSON,
		template.StaticCount,
		template.TurretCount,
		template.MobGroundCount,
		template.MobAirCount,
	).Scan(&template.CreatedAt, &template.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert template: %w", err)
	}

	return &template, nil
}

// List retrieves templates with pagination and filtering
func (s *PostgreSQLTemplateStore) List(ctx context.Context, params model.ListTemplatesQueryParams) ([]model.TemplateSummary, int, error) {
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	// Build WHERE clauses
	if params.NameLike != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+params.NameLike+"%")
		argIndex++
	}

	if params.RoomType != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("room_type = $%d", argIndex))
		args = append(args, params.RoomType)
		argIndex++
	}

	if params.MinWalkableRatio != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("walkable_ratio >= $%d", argIndex))
		args = append(args, *params.MinWalkableRatio)
		argIndex++
	}

	if params.MaxWalkableRatio != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("walkable_ratio <= $%d", argIndex))
		args = append(args, *params.MaxWalkableRatio)
		argIndex++
	}

	// Static count filters
	if params.MinStaticCount != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("static_count >= $%d", argIndex))
		args = append(args, *params.MinStaticCount)
		argIndex++
	}
	if params.MaxStaticCount != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("static_count <= $%d", argIndex))
		args = append(args, *params.MaxStaticCount)
		argIndex++
	}

	// Turret count filters
	if params.MinTurretCount != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("turret_count >= $%d", argIndex))
		args = append(args, *params.MinTurretCount)
		argIndex++
	}
	if params.MaxTurretCount != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("turret_count <= $%d", argIndex))
		args = append(args, *params.MaxTurretCount)
		argIndex++
	}

	// MobGround count filters
	if params.MinMobGroundCount != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("mobground_count >= $%d", argIndex))
		args = append(args, *params.MinMobGroundCount)
		argIndex++
	}
	if params.MaxMobGroundCount != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("mobground_count <= $%d", argIndex))
		args = append(args, *params.MaxMobGroundCount)
		argIndex++
	}

	// MobAir count filters
	if params.MinMobAirCount != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("mobair_count >= $%d", argIndex))
		args = append(args, *params.MinMobAirCount)
		argIndex++
	}
	if params.MaxMobAirCount != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("mobair_count <= $%d", argIndex))
		args = append(args, *params.MaxMobAirCount)
		argIndex++
	}

	// Room attributes filters (JSONB queries)
	if params.HasBoss != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(room_attributes->>'boss')::boolean = $%d", argIndex))
		args = append(args, *params.HasBoss)
		argIndex++
	}
	if params.HasElite != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(room_attributes->>'elite')::boolean = $%d", argIndex))
		args = append(args, *params.HasElite)
		argIndex++
	}
	if params.HasMob != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(room_attributes->>'mob')::boolean = $%d", argIndex))
		args = append(args, *params.HasMob)
		argIndex++
	}
	if params.HasTreasure != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(room_attributes->>'treasure')::boolean = $%d", argIndex))
		args = append(args, *params.HasTreasure)
		argIndex++
	}
	if params.HasTeleport != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(room_attributes->>'teleport')::boolean = $%d", argIndex))
		args = append(args, *params.HasTeleport)
		argIndex++
	}
	if params.HasStory != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(room_attributes->>'story')::boolean = $%d", argIndex))
		args = append(args, *params.HasStory)
		argIndex++
	}

	// Door connectivity filters (JSONB queries)
	if params.TopDoorConnected != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(doors_connected->>'top')::boolean = $%d", argIndex))
		args = append(args, *params.TopDoorConnected)
		argIndex++
	}
	if params.RightDoorConnected != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(doors_connected->>'right')::boolean = $%d", argIndex))
		args = append(args, *params.RightDoorConnected)
		argIndex++
	}
	if params.BottomDoorConnected != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(doors_connected->>'bottom')::boolean = $%d", argIndex))
		args = append(args, *params.BottomDoorConnected)
		argIndex++
	}
	if params.LeftDoorConnected != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(doors_connected->>'left')::boolean = $%d", argIndex))
		args = append(args, *params.LeftDoorConnected)
		argIndex++
	}

	// Build WHERE clause string
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			whereClause += " AND " + whereClauses[i]
		}
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
		SELECT
			id, name, version, width, height, thumbnail,
			walkable_ratio, room_type, room_attributes, doors_connected,
			static_count, turret_count, mobground_count, mobair_count,
			created_at, updated_at
		FROM room_templates %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argIndex, argIndex+1)

	args = append(args, params.Limit, params.Offset)

	rows, err := s.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []model.TemplateSummary
	for rows.Next() {
		var template model.TemplateSummary
		var roomAttributesJSON []byte
		var doorsConnectedJSON []byte

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Version,
			&template.Width,
			&template.Height,
			&template.Thumbnail,
			&template.WalkableRatio,
			&template.RoomType,
			&roomAttributesJSON,
			&doorsConnectedJSON,
			&template.StaticCount,
			&template.TurretCount,
			&template.MobGroundCount,
			&template.MobAirCount,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan template: %w", err)
		}

		// Deserialize JSON fields
		if roomAttributesJSON != nil {
			attrs, err := model.DeserializeRoomAttributes(roomAttributesJSON)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to deserialize room attributes: %w", err)
			}
			template.RoomAttributes = attrs
		}

		if doorsConnectedJSON != nil {
			doors, err := model.DeserializeDoorsConnected(doorsConnectedJSON)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to deserialize doors connected: %w", err)
			}
			template.DoorsConnected = doors
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
		SELECT
			id, name, version, width, height, payload, thumbnail,
			walkable_ratio, room_type, room_attributes, doors_connected,
			static_count, turret_count, mobground_count, mobair_count,
			created_at, updated_at
		FROM room_templates
		WHERE id = $1`

	var template model.Template
	var payloadJSON []byte
	var roomAttributesJSON []byte
	var doorsConnectedJSON []byte

	err = s.db.QueryRow(ctx, query, templateID).Scan(
		&template.ID,
		&template.Name,
		&template.Version,
		&template.Width,
		&template.Height,
		&payloadJSON,
		&template.Thumbnail,
		&template.WalkableRatio,
		&template.RoomType,
		&roomAttributesJSON,
		&doorsConnectedJSON,
		&template.StaticCount,
		&template.TurretCount,
		&template.MobGroundCount,
		&template.MobAirCount,
		&template.CreatedAt,
		&template.UpdatedAt,
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

	// Deserialize room attributes
	if roomAttributesJSON != nil {
		attrs, err := model.DeserializeRoomAttributes(roomAttributesJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize room attributes: %w", err)
		}
		template.RoomAttributes = attrs
	}

	// Deserialize doors connected
	if doorsConnectedJSON != nil {
		doors, err := model.DeserializeDoorsConnected(doorsConnectedJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize doors connected: %w", err)
		}
		template.DoorsConnected = doors
	}

	return &template, nil
}

// Delete removes a template by ID
func (s *PostgreSQLTemplateStore) Delete(ctx context.Context, id string) error {
	// Validate UUID format
	templateID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %w", err)
	}

	query := `DELETE FROM room_templates WHERE id = $1`
	
	result, err := s.db.Exec(ctx, query, templateID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	// Check if any rows were affected
	if result.RowsAffected() == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
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