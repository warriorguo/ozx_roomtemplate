package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"tile-backend/internal/model"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockTemplateStore is a mock implementation of TemplateStore
type MockTemplateStore struct {
	mock.Mock
}

func (m *MockTemplateStore) Create(ctx context.Context, template model.Template) (*model.Template, error) {
	args := m.Called(ctx, template)
	return args.Get(0).(*model.Template), args.Error(1)
}

func (m *MockTemplateStore) Get(ctx context.Context, id string) (*model.Template, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Template), args.Error(1)
}

func (m *MockTemplateStore) List(ctx context.Context, limit, offset int, nameLike string) ([]model.TemplateSummary, int, error) {
	args := m.Called(ctx, limit, offset, nameLike)
	return args.Get(0).([]model.TemplateSummary), args.Get(1).(int), args.Error(2)
}

func (m *MockTemplateStore) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func createTestHandler() *TemplateHandler {
	logger := zap.NewNop() // No-op logger for testing
	mockStore := &MockTemplateStore{}
	return NewTemplateHandler(mockStore, logger)
}

func TestTemplateHandler_CreateTemplate_Success(t *testing.T) {
	handler := createTestHandler()
	mockStore := handler.store.(*MockTemplateStore)

	// Create test request
	req := model.CreateTemplateRequest{
		Name: "test-template",
		Payload: model.TemplatePayload{
			Ground: [][]int{
				{1, 1, 1, 1},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
			},
			Static: [][]int{
				{0, 1, 1, 0},
				{1, 0, 0, 1},
				{1, 0, 0, 1},
				{0, 1, 1, 0},
			},
			Turret: [][]int{
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			MobGround: [][]int{
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
				{0, 0, 0, 0},
			},
			MobAir: [][]int{
				{0, 1, 1, 0},
				{1, 0, 0, 1},
				{1, 0, 0, 1},
				{0, 1, 1, 0},
			},
			Meta: model.TemplateMeta{
				Name:    "test-template",
				Version: 1,
				Width:   4,
				Height:  4,
			},
		},
	}

	// Mock store response
	now := time.Now()
	expectedTemplate := model.Template{
		ID:        uuid.New(),
		Name:      "test-template",
		Version:   1,
		Width:     4,
		Height:    4,
		Payload:   req.Payload,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockStore.On("Create", mock.Anything, mock.MatchedBy(func(t model.Template) bool {
		return t.Name == "test-template" && t.Version == 1
	})).Return(&expectedTemplate, nil)

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.CreateTemplate(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.CreateTemplateResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTemplate.ID, response.ID)
	assert.Equal(t, expectedTemplate.Name, response.Name)

	mockStore.AssertExpectations(t)
}

func TestTemplateHandler_CreateTemplate_InvalidJSON(t *testing.T) {
	handler := createTestHandler()

	// Create HTTP request with invalid JSON
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.CreateTemplate(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response model.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Bad Request", response.Error)
	assert.Equal(t, "Invalid JSON", response.Message)
}

func TestTemplateHandler_CreateTemplate_ValidationFailed(t *testing.T) {
	handler := createTestHandler()

	// Create test request with invalid template (width too small)
	req := model.CreateTemplateRequest{
		Name: "test-template",
		Payload: model.TemplatePayload{
			Ground:    [][]int{{1}}, // Too small
			Static:    [][]int{{0}},
			Turret:    [][]int{{0}},
			MobGround: [][]int{{0}},
			MobAir:    [][]int{{0}},
			Meta: model.TemplateMeta{
				Name:    "test-template",
				Version: 1,
				Width:   1, // Too small
				Height:  1, // Too small
			},
		},
	}

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.CreateTemplate(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response model.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Validation Failed", response.Error)
	assert.Equal(t, "Template validation failed", response.Message)
}

func TestTemplateHandler_CreateTemplate_StoreError(t *testing.T) {
	handler := createTestHandler()
	mockStore := handler.store.(*MockTemplateStore)

	// Create valid test request
	req := model.CreateTemplateRequest{
		Name: "test-template",
		Payload: model.TemplatePayload{
			Ground:    [][]int{{1, 1}, {1, 1}},
			Static:    [][]int{{0, 1}, {1, 0}},
			Turret:    [][]int{{0, 0}, {0, 0}},
			MobGround: [][]int{{0, 0}, {0, 0}},
			MobAir:    [][]int{{0, 1}, {1, 0}},
			Meta: model.TemplateMeta{
				Name:    "test-template",
				Version: 1,
				Width:   2,
				Height:  2,
			},
		},
	}

	// Mock store error
	mockStore.On("Create", mock.Anything, mock.Anything).Return((*model.Template)(nil), fmt.Errorf("database error"))

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.CreateTemplate(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response model.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Internal Server Error", response.Error)
	assert.Equal(t, "Failed to create template", response.Message)

	mockStore.AssertExpectations(t)
}

func TestTemplateHandler_GetTemplate_Success(t *testing.T) {
	handler := createTestHandler()
	mockStore := handler.store.(*MockTemplateStore)

	templateID := uuid.New()
	now := time.Now()

	expectedTemplate := model.Template{
		ID:      templateID,
		Name:    "test-template",
		Version: 1,
		Width:   2,
		Height:  2,
		Payload: model.TemplatePayload{
			Ground: [][]int{{1, 1}, {1, 1}},
			Meta: model.TemplateMeta{
				Name:    "test-template",
				Version: 1,
				Width:   2,
				Height:  2,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Mock store response
	mockStore.On("Get", mock.Anything, templateID.String()).Return(&expectedTemplate, nil)

	// Create HTTP request
	httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+templateID.String(), nil)
	w := httptest.NewRecorder()

	// Add URL params using chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", templateID.String())
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	// Execute
	handler.GetTemplate(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Template
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTemplate.ID, response.ID)
	assert.Equal(t, expectedTemplate.Name, response.Name)

	mockStore.AssertExpectations(t)
}

func TestTemplateHandler_GetTemplate_InvalidUUID(t *testing.T) {
	handler := createTestHandler()

	// Create HTTP request with invalid UUID
	httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/templates/invalid-uuid", nil)
	w := httptest.NewRecorder()

	// Add URL params using chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	// Execute
	handler.GetTemplate(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response model.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Bad Request", response.Error)
	assert.Equal(t, "Invalid UUID format", response.Message)
}

func TestTemplateHandler_GetTemplate_NotFound(t *testing.T) {
	handler := createTestHandler()
	mockStore := handler.store.(*MockTemplateStore)

	templateID := uuid.New()

	// Mock store not found error
	mockStore.On("Get", mock.Anything, templateID.String()).Return((*model.Template)(nil), fmt.Errorf("template not found"))

	// Create HTTP request
	httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+templateID.String(), nil)
	w := httptest.NewRecorder()

	// Add URL params using chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", templateID.String())
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	// Execute
	handler.GetTemplate(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response model.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Not Found", response.Error)
	assert.Equal(t, "Template not found", response.Message)

	mockStore.AssertExpectations(t)
}

func TestTemplateHandler_ListTemplates_Success(t *testing.T) {
	handler := createTestHandler()
	mockStore := handler.store.(*MockTemplateStore)

	now := time.Now()
	expectedItems := []model.TemplateSummary{
		{
			ID:        uuid.New(),
			Name:      "template-1",
			Version:   1,
			Width:     10,
			Height:    8,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New(),
			Name:      "template-2",
			Version:   2,
			Width:     15,
			Height:    12,
			CreatedAt: now.Add(-time.Hour),
			UpdatedAt: now.Add(-time.Hour),
		},
	}

	// Mock store response
	mockStore.On("List", mock.Anything, 20, 0, "").Return(expectedItems, 2, nil)

	// Create HTTP request
	httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.ListTemplates(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	var response model.ListTemplatesResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Items, 2)
	assert.Equal(t, "template-1", response.Items[0].Name)
	assert.Equal(t, "template-2", response.Items[1].Name)

	mockStore.AssertExpectations(t)
}

func TestTemplateHandler_ListTemplates_WithQueryParams(t *testing.T) {
	handler := createTestHandler()
	mockStore := handler.store.(*MockTemplateStore)

	// Mock store response
	mockStore.On("List", mock.Anything, 10, 5, "test").Return([]model.TemplateSummary{}, 0, nil)

	// Create HTTP request with query parameters
	httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/templates?limit=10&offset=5&name_like=test", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.ListTemplates(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	mockStore.AssertExpectations(t)
}

func TestTemplateHandler_ListTemplates_InvalidParams(t *testing.T) {
	handler := createTestHandler()
	mockStore := handler.store.(*MockTemplateStore)

	// Mock store response with default values
	mockStore.On("List", mock.Anything, 20, 0, "").Return([]model.TemplateSummary{}, 0, nil)

	// Create HTTP request with invalid query parameters
	httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/templates?limit=invalid&offset=negative", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.ListTemplates(w, httpReq)

	// Verify - should use default values and succeed
	assert.Equal(t, http.StatusOK, w.Code)

	mockStore.AssertExpectations(t)
}

func TestTemplateHandler_ValidateTemplate_Success(t *testing.T) {
	handler := createTestHandler()

	// Create valid payload
	payload := model.TemplatePayload{
		Ground:    [][]int{{1, 1}, {1, 1}},
		Static:    [][]int{{0, 1}, {1, 0}},
		Turret:    [][]int{{0, 0}, {0, 0}},
		MobGround: [][]int{{0, 0}, {0, 0}},
		MobAir:    [][]int{{0, 1}, {1, 0}},
		Meta: model.TemplateMeta{
			Name:    "test-template",
			Version: 1,
			Width:   2,
			Height:  2,
		},
	}

	// Create HTTP request
	reqBody, _ := json.Marshal(payload)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/templates/validate", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.ValidateTemplate(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	var response model.ValidationResult
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response.Valid)
	assert.Len(t, response.Errors, 0)
}

func TestTemplateHandler_ValidateTemplate_WithStrictMode(t *testing.T) {
	handler := createTestHandler()

	// Create payload with logical errors (static without ground)
	payload := model.TemplatePayload{
		Ground:    [][]int{{0, 1}, {1, 1}},
		Static:    [][]int{{1, 1}, {1, 0}}, // Static at (0,0) where ground=0
		Turret:    [][]int{{0, 0}, {0, 0}},
		MobGround: [][]int{{0, 0}, {0, 0}},
		MobAir:    [][]int{{0, 1}, {1, 0}},
		Meta: model.TemplateMeta{
			Name:    "test-template",
			Version: 1,
			Width:   2,
			Height:  2,
		},
	}

	// Create HTTP request with strict=true
	reqBody, _ := json.Marshal(payload)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/templates/validate?strict=true", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.ValidateTemplate(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	var response model.ValidationResult
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.False(t, response.Valid)
	assert.Greater(t, len(response.Errors), 0)
}

func TestTemplateHandler_ValidateTemplate_InvalidJSON(t *testing.T) {
	handler := createTestHandler()

	// Create HTTP request with invalid JSON
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/templates/validate", bytes.NewReader([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.ValidateTemplate(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response model.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Bad Request", response.Error)
	assert.Equal(t, "Invalid JSON", response.Message)
}

func TestTemplateHandler_HealthCheck_Success(t *testing.T) {
	handler := createTestHandler()
	mockStore := handler.store.(*MockTemplateStore)

	// Mock successful health check
	mockStore.On("HealthCheck", mock.Anything).Return(nil)

	// Create HTTP request
	httpReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.HealthCheck(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])

	mockStore.AssertExpectations(t)
}

func TestTemplateHandler_HealthCheck_DatabaseError(t *testing.T) {
	handler := createTestHandler()
	mockStore := handler.store.(*MockTemplateStore)

	// Mock health check error
	mockStore.On("HealthCheck", mock.Anything).Return(fmt.Errorf("database connection failed"))

	// Create HTTP request
	httpReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.HealthCheck(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response model.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Service Unavailable", response.Error)
	assert.Equal(t, "Database connection failed", response.Message)

	mockStore.AssertExpectations(t)
}

func TestTemplateHandler_CreateTemplate_NameFromMeta(t *testing.T) {
	handler := createTestHandler()
	mockStore := handler.store.(*MockTemplateStore)

	// Create test request without name in root, but with name in meta
	req := model.CreateTemplateRequest{
		// Name: "", // Empty name at root level
		Payload: model.TemplatePayload{
			Ground:    [][]int{{1, 1}, {1, 1}},
			Static:    [][]int{{0, 1}, {1, 0}},
			Turret:    [][]int{{0, 0}, {0, 0}},
			MobGround: [][]int{{0, 0}, {0, 0}},
			MobAir:    [][]int{{0, 1}, {1, 0}},
			Meta: model.TemplateMeta{
				Name:    "meta-template-name", // Name should come from here
				Version: 1,
				Width:   2,
				Height:  2,
			},
		},
	}

	// Mock store response
	now := time.Now()
	expectedTemplate := model.Template{
		ID:        uuid.New(),
		Name:      "meta-template-name", // Should use name from meta
		Version:   1,
		Width:     2,
		Height:    2,
		Payload:   req.Payload,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockStore.On("Create", mock.Anything, mock.MatchedBy(func(t model.Template) bool {
		return t.Name == "meta-template-name"
	})).Return(&expectedTemplate, nil)

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	handler.CreateTemplate(w, httpReq)

	// Verify
	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.CreateTemplateResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "meta-template-name", response.Name)

	mockStore.AssertExpectations(t)
}