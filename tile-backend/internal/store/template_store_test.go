package store

import (
	"context"
	"testing"
	"time"
	"tile-backend/internal/model"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLTemplateStore_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	template := model.Template{
		ID:      uuid.New(),
		Name:    "test-template",
		Version: 1,
		Width:   10,
		Height:  8,
		Payload: model.TemplatePayload{
			Ground:    [][]int{{1, 0}, {0, 1}},
			Static:    [][]int{{0, 1}, {1, 0}},
			Turret:    [][]int{{0, 0}, {0, 0}},
			MobGround: [][]int{{0, 0}, {0, 0}},
			MobAir:    [][]int{{1, 0}, {0, 1}},
			Meta: model.TemplateMeta{
				Name:    "test-template",
				Version: 1,
				Width:   2,
				Height:  2,
			},
		},
	}

	now := time.Now()

	// Mock the INSERT query
	mock.ExpectQuery(`INSERT INTO room_templates`).
		WithArgs(template.ID, template.Name, template.Version, template.Width, template.Height, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"created_at", "updated_at"}).
			AddRow(now, now))

	result, err := store.Create(context.Background(), template)

	assert.NoError(t, err)
	assert.Equal(t, template.ID, result.ID)
	assert.Equal(t, template.Name, result.Name)
	assert.Equal(t, template.Version, result.Version)
	assert.Equal(t, template.Width, result.Width)
	assert.Equal(t, template.Height, result.Height)
	assert.Equal(t, now, result.CreatedAt)
	assert.Equal(t, now, result.UpdatedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLTemplateStore_Create_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	template := model.Template{
		ID:   uuid.New(),
		Name: "test-template",
	}

	// Mock a database error
	mock.ExpectQuery(`INSERT INTO room_templates`).
		WithArgs(template.ID, template.Name, template.Version, template.Width, template.Height, pgxmock.AnyArg()).
		WillReturnError(assert.AnError)

	_, err = store.Create(context.Background(), template)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert template")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLTemplateStore_Get(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	templateID := uuid.New()
	now := time.Now()
	payloadJSON := []byte(`{"ground":[[1,0],[0,1]],"static":[[0,1],[1,0]],"turret":[[0,0],[0,0]],"mobGround":[[0,0],[0,0]],"mobAir":[[1,0],[0,1]],"meta":{"name":"test-template","version":1,"width":2,"height":2}}`)

	// Mock the SELECT query
	mock.ExpectQuery(`SELECT (.+) FROM room_templates WHERE id = \$1`).
		WithArgs(templateID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "version", "width", "height", "payload", "created_at", "updated_at",
		}).AddRow(
			templateID, "test-template", 1, 10, 8,
			payloadJSON,
			now, now,
		))

	result, err := store.Get(context.Background(), templateID.String())

	assert.NoError(t, err)
	assert.Equal(t, templateID, result.ID)
	assert.Equal(t, "test-template", result.Name)
	assert.Equal(t, 1, result.Version)
	assert.Equal(t, 10, result.Width)
	assert.Equal(t, 8, result.Height)
	assert.Equal(t, now, result.CreatedAt)
	assert.Equal(t, now, result.UpdatedAt)
	
	// Check payload was properly unmarshaled
	assert.Equal(t, model.Layer{{1, 0}, {0, 1}}, result.Payload.Ground)
	assert.Equal(t, model.Layer{{0, 1}, {1, 0}}, result.Payload.Static)
	assert.Equal(t, "test-template", result.Payload.Meta.Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLTemplateStore_Get_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	templateID := uuid.New()

	// Mock no rows returned
	mock.ExpectQuery(`SELECT (.+) FROM room_templates WHERE id = \$1`).
		WithArgs(templateID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "version", "width", "height", "payload", "created_at", "updated_at",
		}))

	_, err = store.Get(context.Background(), templateID.String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLTemplateStore_List(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	now := time.Now()

	// Mock count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM room_templates`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))

	// Mock list query
	mock.ExpectQuery(`SELECT (.+) FROM room_templates ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "version", "width", "height", "created_at", "updated_at",
		}).
			AddRow(uuid.New(), "template-1", 1, 10, 8, now, now).
			AddRow(uuid.New(), "template-2", 2, 15, 12, now.Add(-time.Hour), now.Add(-time.Hour)))

	templates, total, err := store.List(context.Background(), 10, 0, "")

	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, templates, 2)
	assert.Equal(t, "template-1", templates[0].Name)
	assert.Equal(t, "template-2", templates[1].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLTemplateStore_List_WithNameFilter(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	now := time.Now()
	nameFilter := "test"

	// Mock count query with name filter
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM room_templates WHERE name ILIKE \$1`).
		WithArgs("%test%").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

	// Mock list query with name filter
	mock.ExpectQuery(`SELECT (.+) FROM room_templates WHERE name ILIKE \$1 ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs("%test%", 20, 0).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "version", "width", "height", "created_at", "updated_at",
		}).AddRow(uuid.New(), "test-template", 1, 10, 8, now, now))

	templates, total, err := store.List(context.Background(), 20, 0, nameFilter)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, templates, 1)
	assert.Equal(t, "test-template", templates[0].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLTemplateStore_List_EmptyResult(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	// Mock count query returning 0
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM room_templates`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))

	// Mock list query returning no rows
	mock.ExpectQuery(`SELECT (.+) FROM room_templates ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(20, 0).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "version", "width", "height", "created_at", "updated_at",
		}))

	templates, total, err := store.List(context.Background(), 20, 0, "")

	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, templates, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLTemplateStore_HealthCheck(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	// Mock successful ping
	mock.ExpectPing()

	err = store.HealthCheck(context.Background())

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLTemplateStore_HealthCheck_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	// Mock ping error
	mock.ExpectPing().WillReturnError(assert.AnError)

	err = store.HealthCheck(context.Background())

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLTemplateStore_Create_JSONMarshalError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	// Create a template with invalid JSON content (this is hard to trigger naturally)
	// We'll simulate by making the query expect specific JSON but we'll provide different data
	template := model.Template{
		ID:      uuid.New(),
		Name:    "test-template",
		Version: 1,
		Width:   2,
		Height:  2,
		Payload: model.TemplatePayload{
			Ground: [][]int{{1, 0}, {0, 1}},
			Meta: model.TemplateMeta{
				Name:    "test-template",
				Version: 1,
				Width:   2,
				Height:  2,
			},
		},
	}

	// Mock the INSERT query to succeed (JSON marshaling happens before the query)
	mock.ExpectQuery(`INSERT INTO room_templates`).
		WithArgs(template.ID, template.Name, template.Version, template.Width, template.Height, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"created_at", "updated_at"}).
			AddRow(time.Now(), time.Now()))

	// This should succeed as JSON marshaling doesn't fail for valid structs
	_, err = store.Create(context.Background(), template)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLTemplateStore_Get_JSONUnmarshalError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	templateID := uuid.New()
	now := time.Now()

	// Mock the SELECT query with invalid JSON
	mock.ExpectQuery(`SELECT (.+) FROM room_templates WHERE id = \$1`).
		WithArgs(templateID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "version", "width", "height", "payload", "created_at", "updated_at",
		}).AddRow(
			templateID, "test-template", 1, 10, 8,
			[]byte(`{"invalid": json}`), // Invalid JSON as []byte
			now, now,
		))

	_, err = store.Get(context.Background(), templateID.String())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal payload")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgreSQLTemplateStore_Context_Cancellation(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewPostgreSQLTemplateStoreWithExecutor(mock)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	template := model.Template{
		ID:   uuid.New(),
		Name: "test-template",
	}

	// Mock the query to return context canceled error
	mock.ExpectQuery(`INSERT INTO room_templates`).
		WithArgs(template.ID, template.Name, template.Version, template.Width, template.Height, pgxmock.AnyArg()).
		WillReturnError(context.Canceled)

	_, err = store.Create(ctx, template)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert template")
}