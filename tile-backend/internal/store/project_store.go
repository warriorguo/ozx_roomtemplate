package store

import (
	"context"
	"encoding/json"
	"fmt"
	"tile-backend/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProjectStore defines the interface for project storage operations
type ProjectStore interface {
	Create(ctx context.Context, project model.Project) (*model.Project, error)
	List(ctx context.Context, params model.ListProjectsQueryParams) ([]model.ProjectSummary, int, error)
	Get(ctx context.Context, id string) (*model.Project, error)
	Update(ctx context.Context, id string, project model.Project) (*model.Project, error)
	Delete(ctx context.Context, id string) error
}

// PostgreSQLProjectStore implements ProjectStore using PostgreSQL
type PostgreSQLProjectStore struct {
	db DBExecutor
}

// NewPostgreSQLProjectStore creates a new PostgreSQL project store
func NewPostgreSQLProjectStore(db *pgxpool.Pool) *PostgreSQLProjectStore {
	return &PostgreSQLProjectStore{db: db}
}

// NewPostgreSQLProjectStoreWithExecutor creates a new PostgreSQL project store with custom executor
func NewPostgreSQLProjectStoreWithExecutor(db DBExecutor) *PostgreSQLProjectStore {
	return &PostgreSQLProjectStore{db: db}
}

// Create saves a new project to the database
func (s *PostgreSQLProjectStore) Create(ctx context.Context, project model.Project) (*model.Project, error) {
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}

	doorJSON, err := json.Marshal(project.DoorDistribution)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal door_distribution: %w", err)
	}

	query := `
		INSERT INTO room_projects (
			id, name, total_rooms,
			shape_pct_full, shape_pct_bridge, shape_pct_platform,
			door_distribution,
			stage_pct_start, stage_pct_teaching, stage_pct_building,
			stage_pct_pressure, stage_pct_peak, stage_pct_release, stage_pct_boss
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING created_at, updated_at`

	err = s.db.QueryRow(ctx, query,
		project.ID,
		project.Name,
		project.TotalRooms,
		project.ShapePctFull,
		project.ShapePctBridge,
		project.ShapePctPlatform,
		doorJSON,
		project.StagePctStart,
		project.StagePctTeaching,
		project.StagePctBuilding,
		project.StagePctPressure,
		project.StagePctPeak,
		project.StagePctRelease,
		project.StagePctBoss,
	).Scan(&project.CreatedAt, &project.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert project: %w", err)
	}

	return &project, nil
}

// List retrieves projects with pagination and filtering
func (s *PostgreSQLProjectStore) List(ctx context.Context, params model.ListProjectsQueryParams) ([]model.ProjectSummary, int, error) {
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	if params.NameLike != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("p.name ILIKE $%d", argIndex))
		args = append(args, "%"+params.NameLike+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			whereClause += " AND " + whereClauses[i]
		}
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM room_projects p " + whereClause
	var total int
	err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get paginated results with template count
	listQuery := fmt.Sprintf(`
		SELECT
			p.id, p.name, p.total_rooms,
			p.shape_pct_full, p.shape_pct_bridge, p.shape_pct_platform,
			p.door_distribution,
			p.stage_pct_start, p.stage_pct_teaching, p.stage_pct_building,
			p.stage_pct_pressure, p.stage_pct_peak, p.stage_pct_release, p.stage_pct_boss,
			COALESCE(tc.cnt, 0) AS template_count,
			p.created_at, p.updated_at
		FROM room_projects p
		LEFT JOIN (
			SELECT project_id, COUNT(*) AS cnt
			FROM room_templates
			WHERE project_id IS NOT NULL
			GROUP BY project_id
		) tc ON p.id = tc.project_id
		%s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argIndex, argIndex+1)

	args = append(args, params.Limit, params.Offset)

	rows, err := s.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []model.ProjectSummary
	for rows.Next() {
		var p model.ProjectSummary
		var doorJSON []byte

		err := rows.Scan(
			&p.ID, &p.Name, &p.TotalRooms,
			&p.ShapePctFull, &p.ShapePctBridge, &p.ShapePctPlatform,
			&doorJSON,
			&p.StagePctStart, &p.StagePctTeaching, &p.StagePctBuilding,
			&p.StagePctPressure, &p.StagePctPeak, &p.StagePctRelease, &p.StagePctBoss,
			&p.TemplateCount,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan project: %w", err)
		}

		if doorJSON != nil {
			if err := json.Unmarshal(doorJSON, &p.DoorDistribution); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal door_distribution: %w", err)
			}
		}

		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating rows: %w", err)
	}

	return projects, total, nil
}

// Get retrieves a project by ID
func (s *PostgreSQLProjectStore) Get(ctx context.Context, id string) (*model.Project, error) {
	projectID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	query := `
		SELECT
			id, name, total_rooms,
			shape_pct_full, shape_pct_bridge, shape_pct_platform,
			door_distribution,
			stage_pct_start, stage_pct_teaching, stage_pct_building,
			stage_pct_pressure, stage_pct_peak, stage_pct_release, stage_pct_boss,
			created_at, updated_at
		FROM room_projects
		WHERE id = $1`

	var p model.Project
	var doorJSON []byte

	err = s.db.QueryRow(ctx, query, projectID).Scan(
		&p.ID, &p.Name, &p.TotalRooms,
		&p.ShapePctFull, &p.ShapePctBridge, &p.ShapePctPlatform,
		&doorJSON,
		&p.StagePctStart, &p.StagePctTeaching, &p.StagePctBuilding,
		&p.StagePctPressure, &p.StagePctPeak, &p.StagePctRelease, &p.StagePctBoss,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to query project: %w", err)
	}

	if doorJSON != nil {
		if err := json.Unmarshal(doorJSON, &p.DoorDistribution); err != nil {
			return nil, fmt.Errorf("failed to unmarshal door_distribution: %w", err)
		}
	}

	return &p, nil
}

// Update updates a project by ID
func (s *PostgreSQLProjectStore) Update(ctx context.Context, id string, project model.Project) (*model.Project, error) {
	projectID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %w", err)
	}

	doorJSON, err := json.Marshal(project.DoorDistribution)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal door_distribution: %w", err)
	}

	query := `
		UPDATE room_projects SET
			name = $2,
			total_rooms = $3,
			shape_pct_full = $4,
			shape_pct_bridge = $5,
			shape_pct_platform = $6,
			door_distribution = $7,
			stage_pct_start = $8,
			stage_pct_teaching = $9,
			stage_pct_building = $10,
			stage_pct_pressure = $11,
			stage_pct_peak = $12,
			stage_pct_release = $13,
			stage_pct_boss = $14
		WHERE id = $1
		RETURNING created_at, updated_at`

	project.ID = projectID
	err = s.db.QueryRow(ctx, query,
		projectID,
		project.Name,
		project.TotalRooms,
		project.ShapePctFull,
		project.ShapePctBridge,
		project.ShapePctPlatform,
		doorJSON,
		project.StagePctStart,
		project.StagePctTeaching,
		project.StagePctBuilding,
		project.StagePctPressure,
		project.StagePctPeak,
		project.StagePctRelease,
		project.StagePctBoss,
	).Scan(&project.CreatedAt, &project.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &project, nil
}

// Delete removes a project by ID
func (s *PostgreSQLProjectStore) Delete(ctx context.Context, id string) error {
	projectID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %w", err)
	}

	result, err := s.db.Exec(ctx, "DELETE FROM room_projects WHERE id = $1", projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("project not found")
	}

	return nil
}
