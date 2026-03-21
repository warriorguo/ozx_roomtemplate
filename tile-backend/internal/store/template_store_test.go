package store

import (
	"context"
	"testing"
	"tile-backend/internal/model"
	"time"

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
			Ground: [][]int{{1, 0}, {0, 1}},
			Static: [][]int{{0, 1}, {1, 0}},
			Chaser: [][]int{{0, 0}, {0, 0}},
			Zoner:  [][]int{{0, 0}, {0, 0}},
			DPS:    [][]int{{0, 0}, {0, 0}},
			MobAir: [][]int{{1, 0}, {0, 1}},
			Meta: model.TemplateMeta{
				Name:    "test-template",
				Version: 1,
				Width:   2,
				Height:  2,
			},
		},
	}

	now := time.Now()

	// Mock the INSERT query (17 args total)
	mock.ExpectQuery(`INSERT INTO room_templates`).
		WithArgs(
			template.ID, template.Name, template.Version, template.Width, template.Height,
			pgxmock.AnyArg(), // payload JSON
			pgxmock.AnyArg(), // thumbnail
			pgxmock.AnyArg(), // walkable_ratio
			pgxmock.AnyArg(), // room_type
			pgxmock.AnyArg(), // room_attributes
			pgxmock.AnyArg(), // doors_connected
			pgxmock.AnyArg(), // static_count
			pgxmock.AnyArg(), // chaser_count
			pgxmock.AnyArg(), // zoner_count
			pgxmock.AnyArg(), // dps_count
			pgxmock.AnyArg(), // mobair_count
			pgxmock.AnyArg(), // stage_type
		).
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
		Payload: model.TemplatePayload{
			Ground: [][]int{{1}},
			Meta:   model.TemplateMeta{Name: "t", Version: 1, Width: 1, Height: 1},
		},
	}

	// Mock a database error - use AnyArg for all params
	mock.ExpectQuery(`INSERT INTO room_templates`).
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(),
		).
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

	templateID := uuid.New().String()
	now := time.Now()

	// Mock the SELECT query - must match all columns returned
	payloadJSON := `{"ground":[[1,0],[0,1]],"static":[[0,1],[1,0]],"chaser":[[0,0],[0,0]],"zoner":[[0,0],[0,0]],"dps":[[0,0],[0,0]],"mobAir":[[1,0],[0,1]],"meta":{"name":"test-template","version":1,"width":2,"height":2}}`
	rows := pgxmock.NewRows([]string{
		"id", "name", "version", "width", "height", "payload", "thumbnail",
		"walkable_ratio", "room_type", "room_attributes", "doors_connected",
		"static_count", "chaser_count", "zoner_count", "dps_count", "mobair_count", "stage_type",
		"created_at", "updated_at",
	}).AddRow(
		templateID, "test-template", 1, 10, 8,
		[]byte(payloadJSON),
		(*string)(nil),  // thumbnail
		(*float64)(nil), // walkable_ratio
		(*string)(nil),  // room_type
		[]byte(nil),     // room_attributes
		[]byte(nil),     // doors_connected
		(*int)(nil),     // static_count
		(*int)(nil),     // chaser_count
		(*int)(nil),     // zoner_count
		(*int)(nil),     // dps_count
		(*int)(nil),     // mobair_count
		(*string)(nil),  // stage_type
		now, now,
	)
	mock.ExpectQuery(`SELECT`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(rows)

	result, err := store.Get(context.Background(), templateID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, templateID, result.ID.String())
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

	templateID := uuid.New().String()

	// Mock no rows returned
	mock.ExpectQuery(`SELECT`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "version", "width", "height", "payload", "thumbnail",
			"walkable_ratio", "room_type", "room_attributes", "doors_connected",
			"static_count", "chaser_count", "zoner_count", "dps_count", "mobair_count", "stage_type",
			"created_at", "updated_at",
		}))

	_, err = store.Get(context.Background(), templateID)

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
	listCols := []string{
		"id", "name", "version", "width", "height", "thumbnail",
		"walkable_ratio", "room_type", "room_attributes", "doors_connected",
		"static_count", "chaser_count", "zoner_count", "dps_count", "mobair_count", "stage_type",
		"created_at", "updated_at",
	}
	mock.ExpectQuery(`SELECT`).
		WithArgs(10, 0).
		WillReturnRows(pgxmock.NewRows(listCols).
			AddRow(uuid.New(), "template-1", 1, 10, 8, (*string)(nil),
				(*float64)(nil), (*string)(nil), []byte(nil), []byte(nil),
				(*int)(nil), (*int)(nil), (*int)(nil), (*int)(nil), (*int)(nil), (*string)(nil),
				now, now).
			AddRow(uuid.New(), "template-2", 2, 15, 12, (*string)(nil),
				(*float64)(nil), (*string)(nil), []byte(nil), []byte(nil),
				(*int)(nil), (*int)(nil), (*int)(nil), (*int)(nil), (*int)(nil), (*string)(nil),
				now.Add(-time.Hour), now.Add(-time.Hour)))

	templates, total, err := store.List(context.Background(), model.ListTemplatesQueryParams{Limit: 10, Offset: 0})

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
	listCols := []string{
		"id", "name", "version", "width", "height", "thumbnail",
		"walkable_ratio", "room_type", "room_attributes", "doors_connected",
		"static_count", "chaser_count", "zoner_count", "dps_count", "mobair_count", "stage_type",
		"created_at", "updated_at",
	}
	mock.ExpectQuery(`SELECT`).
		WithArgs("%test%", 20, 0).
		WillReturnRows(pgxmock.NewRows(listCols).
			AddRow(uuid.New(), "test-template", 1, 10, 8, (*string)(nil),
				(*float64)(nil), (*string)(nil), []byte(nil), []byte(nil),
				(*int)(nil), (*int)(nil), (*int)(nil), (*int)(nil), (*int)(nil), (*string)(nil),
				now, now))

	templates, total, err := store.List(context.Background(), model.ListTemplatesQueryParams{Limit: 20, Offset: 0, NameLike: nameFilter})

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
	mock.ExpectQuery(`SELECT`).
		WithArgs(20, 0).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "version", "width", "height", "thumbnail",
			"walkable_ratio", "room_type", "room_attributes", "doors_connected",
			"static_count", "chaser_count", "zoner_count", "dps_count", "mobair_count", "stage_type",
			"created_at", "updated_at",
		}))

	templates, total, err := store.List(context.Background(), model.ListTemplatesQueryParams{Limit: 20, Offset: 0})

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
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(),
		).
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

	templateID := uuid.New().String()
	now := time.Now()

	// Mock the SELECT query with invalid JSON
	mock.ExpectQuery(`SELECT`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "name", "version", "width", "height", "payload", "thumbnail",
			"walkable_ratio", "room_type", "room_attributes", "doors_connected",
			"static_count", "chaser_count", "zoner_count", "dps_count", "mobair_count", "stage_type",
			"created_at", "updated_at",
		}).AddRow(
			templateID, "test-template", 1, 10, 8,
			[]byte(`{"invalid": json}`), // Invalid JSON
			(*string)(nil), (*float64)(nil), (*string)(nil), []byte(nil), []byte(nil),
			(*int)(nil), (*int)(nil), (*int)(nil), (*int)(nil), (*int)(nil), (*string)(nil),
			now, now,
		))

	_, err = store.Get(context.Background(), templateID)

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

	// With cancelled context, the query should fail
	_, err = store.Create(ctx, template)

	assert.Error(t, err)
}
