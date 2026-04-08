package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"tile-backend/internal/generate"
	"tile-backend/internal/model"
	"tile-backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ProjectHandler handles HTTP requests for projects
type ProjectHandler struct {
	store         store.ProjectStore
	templateStore store.TemplateStore
	logger        *zap.Logger
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(store store.ProjectStore, templateStore store.TemplateStore, logger *zap.Logger) *ProjectHandler {
	return &ProjectHandler{
		store:         store,
		templateStore: templateStore,
		logger:        logger,
	}
}

// CreateProject handles POST /api/v1/projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Validate request
	if errs := model.ValidateProjectRequest(&req); len(errs) > 0 {
		resp := model.ErrorResponse{
			Error:   "Validation Failed",
			Message: "Project validation failed",
			Details: errs,
		}
		respondJSON(w, h.logger, http.StatusBadRequest, resp)
		return
	}

	project := model.Project{
		ID:               uuid.New(),
		Name:             req.Name,
		TotalRooms:       req.TotalRooms,
		ShapePctFull:     req.ShapePctFull,
		ShapePctBridge:   req.ShapePctBridge,
		ShapePctPlatform: req.ShapePctPlatform,
		DoorDistribution: req.DoorDistribution,
		StagePctStart:    req.StagePctStart,
		StagePctTeaching: req.StagePctTeaching,
		StagePctBuilding: req.StagePctBuilding,
		StagePctPressure: req.StagePctPressure,
		StagePctPeak:     req.StagePctPeak,
		StagePctRelease:  req.StagePctRelease,
		StagePctBoss:     req.StagePctBoss,
	}

	saved, err := h.store.Create(r.Context(), project)
	if err != nil {
		h.logger.Error("Failed to create project", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "Failed to create project", err.Error())
		return
	}

	respondJSON(w, h.logger, http.StatusCreated, saved)
}

// ListProjects handles GET /api/v1/projects
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	params := model.ListProjectsQueryParams{
		Limit:    20,
		Offset:   0,
		NameLike: query.Get("name_like"),
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			params.Limit = l
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			params.Offset = o
		}
	}

	projects, total, err := h.store.List(r.Context(), params)
	if err != nil {
		h.logger.Error("Failed to list projects", zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "Failed to list projects", err.Error())
		return
	}

	response := model.ListProjectsResponse{
		Total: total,
		Items: projects,
	}

	respondJSON(w, h.logger, http.StatusOK, response)
}

// GetProject handles GET /api/v1/projects/{id}
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := uuid.Parse(id); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "Invalid UUID format", err.Error())
		return
	}

	project, err := h.store.Get(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, h.logger, http.StatusNotFound, "Project not found", "")
			return
		}
		h.logger.Error("Failed to get project", zap.String("id", id), zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "Failed to get project", err.Error())
		return
	}

	respondJSON(w, h.logger, http.StatusOK, project)
}

// UpdateProject handles PUT /api/v1/projects/{id}
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := uuid.Parse(id); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "Invalid UUID format", err.Error())
		return
	}

	var req model.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	if errs := model.ValidateProjectRequest(&req); len(errs) > 0 {
		resp := model.ErrorResponse{
			Error:   "Validation Failed",
			Message: "Project validation failed",
			Details: errs,
		}
		respondJSON(w, h.logger, http.StatusBadRequest, resp)
		return
	}

	project := model.Project{
		Name:             req.Name,
		TotalRooms:       req.TotalRooms,
		ShapePctFull:     req.ShapePctFull,
		ShapePctBridge:   req.ShapePctBridge,
		ShapePctPlatform: req.ShapePctPlatform,
		DoorDistribution: req.DoorDistribution,
		StagePctStart:    req.StagePctStart,
		StagePctTeaching: req.StagePctTeaching,
		StagePctBuilding: req.StagePctBuilding,
		StagePctPressure: req.StagePctPressure,
		StagePctPeak:     req.StagePctPeak,
		StagePctRelease:  req.StagePctRelease,
		StagePctBoss:     req.StagePctBoss,
	}

	updated, err := h.store.Update(r.Context(), id, project)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, h.logger, http.StatusNotFound, "Project not found", "")
			return
		}
		h.logger.Error("Failed to update project", zap.String("id", id), zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "Failed to update project", err.Error())
		return
	}

	respondJSON(w, h.logger, http.StatusOK, updated)
}

// DeleteProject handles DELETE /api/v1/projects/{id}
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := uuid.Parse(id); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "Invalid UUID format", err.Error())
		return
	}

	err := h.store.Delete(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, h.logger, http.StatusNotFound, "Project not found", "")
			return
		}
		h.logger.Error("Failed to delete project", zap.String("id", id), zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "Failed to delete project", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetProjectStats handles GET /api/v1/projects/{id}/stats
func (h *ProjectHandler) GetProjectStats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := uuid.Parse(id); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "Invalid UUID format", err.Error())
		return
	}

	stats, err := h.store.Stats(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, h.logger, http.StatusNotFound, "Project not found", "")
			return
		}
		h.logger.Error("Failed to get project stats", zap.String("id", id), zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "Failed to get project stats", err.Error())
		return
	}

	respondJSON(w, h.logger, http.StatusOK, stats)
}

// AutoFillProject handles POST /api/v1/projects/{id}/autofill
func (h *ProjectHandler) AutoFillProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := uuid.Parse(id); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "Invalid UUID format", err.Error())
		return
	}

	// Get project
	project, err := h.store.Get(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, h.logger, http.StatusNotFound, "Project not found", "")
			return
		}
		h.logger.Error("Failed to get project", zap.String("id", id), zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "Failed to get project", err.Error())
		return
	}

	// Get current stats
	stats, err := h.store.Stats(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get project stats", zap.String("id", id), zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "Failed to get project stats", err.Error())
		return
	}

	// Run auto-fill
	result, err := generate.AutoFill(r.Context(), project, stats, h.templateStore)
	if err != nil {
		h.logger.Error("Auto-fill failed", zap.String("id", id), zap.Error(err))
		respondError(w, h.logger, http.StatusInternalServerError, "Auto-fill failed", err.Error())
		return
	}

	respondJSON(w, h.logger, http.StatusOK, result)
}
