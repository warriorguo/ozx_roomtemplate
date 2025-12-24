package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"tile-backend/internal/model"
	"tile-backend/internal/store"
	"tile-backend/internal/validate"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TemplateHandler handles HTTP requests for templates
type TemplateHandler struct {
	store  store.TemplateStore
	logger *zap.Logger
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(store store.TemplateStore, logger *zap.Logger) *TemplateHandler {
	return &TemplateHandler{
		store:  store,
		logger: logger,
	}
}

// CreateTemplate handles POST /api/v1/templates
func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTemplateRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Extract metadata from payload if name is not provided in request body
	name := req.Name
	if name == "" {
		name = req.Payload.Meta.Name
	}

	// Validate template payload
	validationResult := validate.ValidateTemplate(&req.Payload, false)
	if !validationResult.Valid {
		h.respondValidationError(w, validationResult)
		return
	}

	// Create template model
	template := model.Template{
		ID:        uuid.New(),
		Name:      name,
		Version:   req.Payload.Meta.Version,
		Width:     req.Payload.Meta.Width,
		Height:    req.Payload.Meta.Height,
		Payload:   req.Payload,
		Thumbnail: req.Thumbnail,
	}

	// Save to database
	savedTemplate, err := h.store.Create(r.Context(), template)
	if err != nil {
		h.logger.Error("Failed to create template", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to create template", err.Error())
		return
	}

	// Respond with success
	response := model.CreateTemplateResponse{
		ID:        savedTemplate.ID,
		Name:      savedTemplate.Name,
		CreatedAt: savedTemplate.CreatedAt,
		UpdatedAt: savedTemplate.UpdatedAt,
	}

	h.respondJSON(w, http.StatusCreated, response)
}

// ListTemplates handles GET /api/v1/templates
func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	nameLike := r.URL.Query().Get("name_like")

	// Set default values
	limit := 20
	offset := 0

	// Parse limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Query database
	templates, total, err := h.store.List(r.Context(), limit, offset, nameLike)
	if err != nil {
		h.logger.Error("Failed to list templates", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to list templates", err.Error())
		return
	}

	// Respond with results
	response := model.ListTemplatesResponse{
		Total: total,
		Items: templates,
	}

	h.respondJSON(w, http.StatusOK, response)
}

// GetTemplate handles GET /api/v1/templates/{id}
func (h *TemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid UUID format", err.Error())
		return
	}

	// Query database
	template, err := h.store.Get(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "Template not found", "")
			return
		}
		h.logger.Error("Failed to get template", zap.String("id", id), zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to get template", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, template)
}

// DeleteTemplate handles DELETE /api/v1/templates/{id}
func (h *TemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid UUID format", err.Error())
		return
	}

	// Delete from database
	err := h.store.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, http.StatusNotFound, "Template not found", "")
			return
		}
		h.logger.Error("Failed to delete template", zap.String("id", id), zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to delete template", err.Error())
		return
	}

	// Return success with no content
	w.WriteHeader(http.StatusNoContent)
}

// ValidateTemplate handles POST /api/v1/templates/validate
func (h *TemplateHandler) ValidateTemplate(w http.ResponseWriter, r *http.Request) {
	var payload model.TemplatePayload

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Check if strict validation is requested
	strict := r.URL.Query().Get("strict") == "true"

	// Validate template
	validationResult := validate.ValidateTemplate(&payload, strict)

	h.respondJSON(w, http.StatusOK, validationResult)
}

// HealthCheck handles GET /health
func (h *TemplateHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := h.store.HealthCheck(r.Context()); err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		h.respondError(w, http.StatusServiceUnavailable, "Database connection failed", err.Error())
		return
	}

	response := map[string]string{
		"status": "healthy",
	}
	h.respondJSON(w, http.StatusOK, response)
}

// respondJSON sends a JSON response
func (h *TemplateHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// respondError sends an error response
func (h *TemplateHandler) respondError(w http.ResponseWriter, status int, message, details string) {
	response := model.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}

	if details != "" {
		response.Details = map[string]string{"details": details}
	}

	h.respondJSON(w, status, response)
}

// respondValidationError sends a validation error response
func (h *TemplateHandler) respondValidationError(w http.ResponseWriter, result *model.ValidationResult) {
	response := model.ErrorResponse{
		Error:   "Validation Failed",
		Message: "Template validation failed",
		Details: make(map[string]string),
	}

	// Add first few validation errors to details
	for i, err := range result.Errors {
		if i >= 5 { // Limit to first 5 errors
			response.Details["note"] = "... and more errors"
			break
		}
		key := err.Layer + "_" + strconv.Itoa(err.X) + "_" + strconv.Itoa(err.Y)
		response.Details[key] = err.Reason
	}

	h.respondJSON(w, http.StatusBadRequest, response)
}