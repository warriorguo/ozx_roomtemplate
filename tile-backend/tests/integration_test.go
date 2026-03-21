package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	httpHandler "tile-backend/internal/http"
	"tile-backend/internal/model"
	"tile-backend/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type IntegrationTestSuite struct {
	suite.Suite
	db     *pgxpool.Pool
	server *httptest.Server
	logger *zap.Logger
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Skip integration tests if TEST_INTEGRATION environment variable is not set
	if os.Getenv("TEST_INTEGRATION") == "" {
		suite.T().Skip("Skipping integration tests. Set TEST_INTEGRATION=1 to run them.")
	}

	// Initialize logger
	suite.logger = zap.NewNop()

	// Initialize test database
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://liuli@localhost:5432/postgres?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(suite.T(), err)

	suite.db, err = pgxpool.NewWithConfig(context.Background(), config)
	require.NoError(suite.T(), err)

	// Test database connection
	err = suite.db.Ping(context.Background())
	require.NoError(suite.T(), err)

	// Initialize stores
	templateStore := store.NewPostgreSQLTemplateStore(suite.db)

	// Setup HTTP server
	router := httpHandler.SetupRouter(templateStore, suite.logger, []string{})
	suite.server = httptest.NewServer(router)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *IntegrationTestSuite) SetupTest() {
	// Clean up test data before each test
	_, err := suite.db.Exec(context.Background(), "TRUNCATE TABLE room_templates")
	require.NoError(suite.T(), err)
}

func (suite *IntegrationTestSuite) TestCreateAndGetTemplate() {
	// Create a template
	createReq := model.CreateTemplateRequest{
		Name: "integration-test-template",
		Payload: model.TemplatePayload{
			Ground: [][]int{{1, 1, 0}, {0, 1, 1}, {1, 0, 1}},
			Static: [][]int{{0, 1, 0}, {0, 0, 1}, {1, 0, 0}},
			Chaser: [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			Zoner:  [][]int{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
			DPS:    [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}},
			MobAir: [][]int{{0, 0, 1}, {1, 0, 0}, {0, 1, 0}},
			Meta: model.TemplateMeta{
				Name:    "integration-test-template",
				Version: 1,
				Width:   3,
				Height:  3,
			},
		},
	}

	// POST /api/v1/templates
	createBody, _ := json.Marshal(createReq)
	resp, err := http.Post(suite.server.URL+"/api/v1/templates", "application/json", bytes.NewReader(createBody))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var createResp model.CreateTemplateResponse
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	require.NoError(suite.T(), err)

	assert.NotEmpty(suite.T(), createResp.ID)
	assert.Equal(suite.T(), "integration-test-template", createResp.Name)

	// GET /api/v1/templates/{id}
	getResp, err := http.Get(suite.server.URL + "/api/v1/templates/" + createResp.ID.String())
	require.NoError(suite.T(), err)
	defer getResp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, getResp.StatusCode)

	var template model.Template
	err = json.NewDecoder(getResp.Body).Decode(&template)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), createResp.ID, template.ID)
	assert.Equal(suite.T(), "integration-test-template", template.Name)
	assert.Equal(suite.T(), 1, template.Version)
	assert.Equal(suite.T(), 3, template.Width)
	assert.Equal(suite.T(), 3, template.Height)

	// Verify payload
	assert.Equal(suite.T(), [][]int{{1, 1, 0}, {0, 1, 1}, {1, 0, 1}}, template.Payload.Ground)
	assert.Equal(suite.T(), [][]int{{0, 1, 0}, {0, 0, 1}, {1, 0, 0}}, template.Payload.Static)
	assert.Equal(suite.T(), "integration-test-template", template.Payload.Meta.Name)
}

func (suite *IntegrationTestSuite) TestListTemplates() {
	// Create multiple templates
	templates := []model.CreateTemplateRequest{
		{
			Name: "template-alpha",
			Payload: model.TemplatePayload{
				Ground: [][]int{{1, 1}, {1, 1}},
				Static: [][]int{{0, 1}, {1, 0}},
				Chaser: [][]int{{0, 0}, {0, 0}},
				Zoner:  [][]int{{0, 0}, {0, 0}},
				DPS:    [][]int{{0, 0}, {0, 0}},
				MobAir: [][]int{{0, 1}, {1, 0}},
				Meta: model.TemplateMeta{
					Name:    "template-alpha",
					Version: 1,
					Width:   2,
					Height:  2,
				},
			},
		},
		{
			Name: "template-beta",
			Payload: model.TemplatePayload{
				Ground: [][]int{{1, 0}, {0, 1}},
				Static: [][]int{{0, 0}, {0, 0}},
				Chaser: [][]int{{1, 0}, {0, 1}},
				Zoner:  [][]int{{0, 0}, {0, 0}},
				DPS:    [][]int{{0, 0}, {0, 0}},
				MobAir: [][]int{{0, 0}, {0, 0}},
				Meta: model.TemplateMeta{
					Name:    "template-beta",
					Version: 2,
					Width:   2,
					Height:  2,
				},
			},
		},
		{
			Name: "another-template",
			Payload: model.TemplatePayload{
				Ground: [][]int{{1, 1}, {1, 1}},
				Static: [][]int{{1, 0}, {0, 1}},
				Chaser: [][]int{{0, 0}, {0, 0}},
				Zoner:  [][]int{{0, 0}, {0, 0}},
				DPS:    [][]int{{0, 0}, {0, 0}},
				MobAir: [][]int{{0, 0}, {0, 0}},
				Meta: model.TemplateMeta{
					Name:    "another-template",
					Version: 1,
					Width:   2,
					Height:  2,
				},
			},
		},
	}

	// Create templates
	for _, tmpl := range templates {
		createBody, _ := json.Marshal(tmpl)
		resp, err := http.Post(suite.server.URL+"/api/v1/templates", "application/json", bytes.NewReader(createBody))
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	}

	// Test list all templates
	resp, err := http.Get(suite.server.URL + "/api/v1/templates")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var listResp model.ListTemplatesResponse
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), 3, listResp.Total)
	assert.Len(suite.T(), listResp.Items, 3)

	// Test list with name filter
	resp, err = http.Get(suite.server.URL + "/api/v1/templates?name_like=template")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), 2, listResp.Total) // Should match "template-alpha" and "template-beta"
	assert.Len(suite.T(), listResp.Items, 2)

	// Test list with pagination
	resp, err = http.Get(suite.server.URL + "/api/v1/templates?limit=2&offset=1")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), 3, listResp.Total) // Total count should still be 3
	assert.Len(suite.T(), listResp.Items, 2)   // But we should get only 2 items due to limit
}

func (suite *IntegrationTestSuite) TestValidateTemplate() {
	// Test valid template
	validPayload := model.TemplatePayload{
		Ground: [][]int{{1, 1}, {1, 1}},
		Static: [][]int{{0, 1}, {1, 0}},
		Chaser: [][]int{{0, 0}, {0, 0}},
		Zoner:  [][]int{{0, 0}, {0, 0}},
		DPS:    [][]int{{0, 0}, {0, 0}},
		MobAir: [][]int{{0, 1}, {1, 0}},
		Meta: model.TemplateMeta{
			Name:    "valid-template",
			Version: 1,
			Width:   2,
			Height:  2,
		},
	}

	validateBody, _ := json.Marshal(validPayload)
	resp, err := http.Post(suite.server.URL+"/api/v1/templates/validate", "application/json", bytes.NewReader(validateBody))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var validationResp model.ValidationResult
	err = json.NewDecoder(resp.Body).Decode(&validationResp)
	require.NoError(suite.T(), err)

	assert.True(suite.T(), validationResp.Valid)
	assert.Len(suite.T(), validationResp.Errors, 0)

	// Test invalid template with strict validation
	invalidPayload := model.TemplatePayload{
		Ground: [][]int{{0, 1}, {1, 1}},
		Static: [][]int{{1, 1}, {1, 0}}, // Static at (0,0) where ground=0
		Chaser: [][]int{{0, 0}, {0, 0}},
		Zoner:  [][]int{{0, 0}, {0, 0}},
		DPS:    [][]int{{0, 0}, {0, 0}},
		MobAir: [][]int{{0, 1}, {1, 0}},
		Meta: model.TemplateMeta{
			Name:    "invalid-template",
			Version: 1,
			Width:   2,
			Height:  2,
		},
	}

	validateBody, _ = json.Marshal(invalidPayload)
	resp, err = http.Post(suite.server.URL+"/api/v1/templates/validate?strict=true", "application/json", bytes.NewReader(validateBody))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&validationResp)
	require.NoError(suite.T(), err)

	assert.False(suite.T(), validationResp.Valid)
	assert.Greater(suite.T(), len(validationResp.Errors), 0)
}

func (suite *IntegrationTestSuite) TestHealthCheck() {
	resp, err := http.Get(suite.server.URL + "/health")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var healthResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "healthy", healthResp["status"])
}

func (suite *IntegrationTestSuite) TestErrorCases() {
	// Test creating template with invalid JSON
	resp, err := http.Post(suite.server.URL+"/api/v1/templates", "application/json", bytes.NewReader([]byte("invalid json")))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	// Test getting non-existent template
	resp, err = http.Get(suite.server.URL + "/api/v1/templates/123e4567-e89b-12d3-a456-426614174000")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	// Test getting template with invalid UUID
	resp, err = http.Get(suite.server.URL + "/api/v1/templates/invalid-uuid")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	// Test creating template with validation errors
	invalidPayload := model.CreateTemplateRequest{
		Name: "invalid-template",
		Payload: model.TemplatePayload{
			Ground: [][]int{{1}}, // Too small
			Static: [][]int{{0}},
			Chaser: [][]int{{0}},
			Zoner:  [][]int{{0}},
			DPS:    [][]int{{0}},
			MobAir: [][]int{{0}},
			Meta: model.TemplateMeta{
				Name:    "invalid-template",
				Version: 1,
				Width:   1, // Too small
				Height:  1, // Too small
			},
		},
	}

	createBody, _ := json.Marshal(invalidPayload)
	resp, err = http.Post(suite.server.URL+"/api/v1/templates", "application/json", bytes.NewReader(createBody))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	var errorResp model.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Validation Failed", errorResp.Error)
	assert.Equal(suite.T(), "Template validation failed", errorResp.Message)
	assert.NotEmpty(suite.T(), errorResp.Details)
}

func (suite *IntegrationTestSuite) TestConcurrentOperations() {
	// Test concurrent template creation
	numGoroutines := 10
	resultChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			createReq := model.CreateTemplateRequest{
				Name: fmt.Sprintf("concurrent-template-%d", index),
				Payload: model.TemplatePayload{
					Ground: [][]int{{1, 1}, {1, 1}},
					Static: [][]int{{0, 1}, {1, 0}},
					Chaser: [][]int{{0, 0}, {0, 0}},
					Zoner:  [][]int{{0, 0}, {0, 0}},
					DPS:    [][]int{{0, 0}, {0, 0}},
					MobAir: [][]int{{0, 1}, {1, 0}},
					Meta: model.TemplateMeta{
						Name:    fmt.Sprintf("concurrent-template-%d", index),
						Version: 1,
						Width:   2,
						Height:  2,
					},
				},
			}

			createBody, _ := json.Marshal(createReq)
			resp, err := http.Post(suite.server.URL+"/api/v1/templates", "application/json", bytes.NewReader(createBody))
			if err != nil {
				resultChan <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				resultChan <- fmt.Errorf("expected status 201, got %d", resp.StatusCode)
				return
			}

			resultChan <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-resultChan
		assert.NoError(suite.T(), err)
	}

	// Verify all templates were created
	resp, err := http.Get(suite.server.URL + "/api/v1/templates")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	var listResp model.ListTemplatesResponse
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), numGoroutines, listResp.Total)
}

func (suite *IntegrationTestSuite) TestGenerateBridge_Success() {
	// Test generating a bridge with two doors
	generateReq := map[string]interface{}{
		"width":  15,
		"height": 12,
		"doors":  []string{"top", "bottom"},
	}

	generateBody, _ := json.Marshal(generateReq)
	resp, err := http.Post(suite.server.URL+"/api/v1/generate/bridge", "application/json", bytes.NewReader(generateBody))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var generateResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&generateResp)
	require.NoError(suite.T(), err)

	// Verify payload exists
	payload, ok := generateResp["payload"].(map[string]interface{})
	require.True(suite.T(), ok, "payload should exist")

	// Verify ground layer exists and has correct dimensions
	ground, ok := payload["ground"].([]interface{})
	require.True(suite.T(), ok, "ground layer should exist")
	assert.Equal(suite.T(), 12, len(ground), "ground should have 12 rows")

	firstRow, ok := ground[0].([]interface{})
	require.True(suite.T(), ok)
	assert.Equal(suite.T(), 15, len(firstRow), "ground should have 15 columns")

	// Verify roomType is bridge
	roomType, ok := payload["roomType"].(string)
	require.True(suite.T(), ok)
	assert.Equal(suite.T(), "bridge", roomType)

	// Verify doors
	doors, ok := payload["doors"].(map[string]interface{})
	require.True(suite.T(), ok)
	assert.Equal(suite.T(), float64(1), doors["top"])
	assert.Equal(suite.T(), float64(0), doors["right"])
	assert.Equal(suite.T(), float64(1), doors["bottom"])
	assert.Equal(suite.T(), float64(0), doors["left"])

	// Verify other layers exist and are empty
	for _, layerName := range []string{"softEdge", "bridge", "static", "turret", "mobGround", "mobAir"} {
		layer, ok := payload[layerName].([]interface{})
		require.True(suite.T(), ok, "%s layer should exist", layerName)
		assert.Equal(suite.T(), 12, len(layer), "%s should have 12 rows", layerName)
	}
}

func (suite *IntegrationTestSuite) TestGenerateBridge_FourDoors() {
	generateReq := map[string]interface{}{
		"width":  20,
		"height": 20,
		"doors":  []string{"top", "right", "bottom", "left"},
	}

	generateBody, _ := json.Marshal(generateReq)
	resp, err := http.Post(suite.server.URL+"/api/v1/generate/bridge", "application/json", bytes.NewReader(generateBody))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var generateResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&generateResp)
	require.NoError(suite.T(), err)

	payload, ok := generateResp["payload"].(map[string]interface{})
	require.True(suite.T(), ok)

	// Verify all doors are set
	doors, ok := payload["doors"].(map[string]interface{})
	require.True(suite.T(), ok)
	assert.Equal(suite.T(), float64(1), doors["top"])
	assert.Equal(suite.T(), float64(1), doors["right"])
	assert.Equal(suite.T(), float64(1), doors["bottom"])
	assert.Equal(suite.T(), float64(1), doors["left"])
}

func (suite *IntegrationTestSuite) TestGenerateBridge_NotEnoughDoors() {
	// Test with only one door - should fail
	generateReq := map[string]interface{}{
		"width":  10,
		"height": 10,
		"doors":  []string{"top"},
	}

	generateBody, _ := json.Marshal(generateReq)
	resp, err := http.Post(suite.server.URL+"/api/v1/generate/bridge", "application/json", bytes.NewReader(generateBody))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(suite.T(), err)

	assert.Contains(suite.T(), errorResp["message"], "2 doors")
}

func (suite *IntegrationTestSuite) TestGenerateBridge_NoDoors() {
	generateReq := map[string]interface{}{
		"width":  10,
		"height": 10,
		"doors":  []string{},
	}

	generateBody, _ := json.Marshal(generateReq)
	resp, err := http.Post(suite.server.URL+"/api/v1/generate/bridge", "application/json", bytes.NewReader(generateBody))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
}

func (suite *IntegrationTestSuite) TestGenerateBridge_InvalidDimensions() {
	// Test with dimensions too small
	generateReq := map[string]interface{}{
		"width":  2,
		"height": 2,
		"doors":  []string{"top", "bottom"},
	}

	generateBody, _ := json.Marshal(generateReq)
	resp, err := http.Post(suite.server.URL+"/api/v1/generate/bridge", "application/json", bytes.NewReader(generateBody))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(suite.T(), err)

	assert.Contains(suite.T(), errorResp["message"], "dimension")
}

func (suite *IntegrationTestSuite) TestGenerateBridge_InvalidJSON() {
	resp, err := http.Post(suite.server.URL+"/api/v1/generate/bridge", "application/json", bytes.NewReader([]byte("invalid json")))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)
}

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
