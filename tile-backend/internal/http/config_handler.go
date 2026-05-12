package http

import (
	"encoding/json"
	"net/http"
	"sync"

	"tile-backend/internal/config"
	"tile-backend/internal/store"
	"tile-backend/internal/store/fsstore"

	"go.uber.org/zap"
)

// ConfigHandler exposes read and write access to the active user
// configuration so the frontend can display the current OZX project folder
// and (per ORT-69) switch to a different one without restarting the binary.
//
// All mutable state — the loaded config, its on-disk path, and the resolved
// templates directory — is guarded by a single RWMutex. The handler also
// owns the SwappableStore so a successful PUT can repoint the backing
// fsstore in-process.
type ConfigHandler struct {
	mu           sync.RWMutex
	cfg          config.Config
	configPath   string
	templatesDir string

	store  *store.SwappableStore
	logger *zap.Logger
}

// NewConfigHandler builds a ConfigHandler around an already-loaded Config.
// store is the SwappableStore that PUT /config will repoint when the user
// changes the project root.
func NewConfigHandler(
	cfg config.Config,
	configPath, templatesDir string,
	swappable *store.SwappableStore,
	logger *zap.Logger,
) *ConfigHandler {
	return &ConfigHandler{
		cfg:          cfg,
		configPath:   configPath,
		templatesDir: templatesDir,
		store:        swappable,
		logger:       logger,
	}
}

// ConfigResponse is the shape returned by GET /api/v1/config. It is a
// superset of the on-disk Config — adds resolved fields and the source path.
type ConfigResponse struct {
	ProjectRoot     string `json:"project_root"`
	TemplateSubdir  string `json:"template_subdir"`
	Port            int    `json:"port"`
	AutoOpenBrowser bool   `json:"auto_open_browser"`
	OzxRoomViewPath string `json:"ozx_room_view_path"`

	// Computed: where the fsstore is actually reading and writing.
	TemplatesDir string `json:"templates_dir"`
	// ConfigPath is the absolute path of the config file on disk so users can
	// edit it from the UI.
	ConfigPath string `json:"config_path"`
	// UsesFallback is true when no OZX project_root was configured and we
	// fell back to the per-user templates directory.
	UsesFallback bool `json:"uses_fallback"`
}

// UpdateConfigRequest is the shape accepted by PUT /api/v1/config. Each field
// is optional; omitted fields keep their existing value.
type UpdateConfigRequest struct {
	ProjectRoot     *string `json:"project_root,omitempty"`
	TemplateSubdir  *string `json:"template_subdir,omitempty"`
	Port            *int    `json:"port,omitempty"`
	AutoOpenBrowser *bool   `json:"auto_open_browser,omitempty"`
	OzxRoomViewPath *string `json:"ozx_room_view_path,omitempty"`
}

// GetConfig handles GET /api/v1/config.
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	resp := h.snapshot()
	h.mu.RUnlock()
	respondJSON(w, h.logger, http.StatusOK, resp)
}

// UpdateConfig handles PUT /api/v1/config. On success the new config is
// persisted to disk and the in-process fsstore is swapped to the new
// templates directory so subsequent template requests hit the new project.
func (h *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	next := h.cfg
	if req.ProjectRoot != nil {
		next.ProjectRoot = *req.ProjectRoot
	}
	if req.TemplateSubdir != nil {
		next.TemplateSubdir = *req.TemplateSubdir
	}
	if req.Port != nil {
		next.Port = *req.Port
	}
	if req.AutoOpenBrowser != nil {
		next.AutoOpenBrowser = *req.AutoOpenBrowser
	}
	if req.OzxRoomViewPath != nil {
		next.OzxRoomViewPath = *req.OzxRoomViewPath
	}

	nextDir, err := next.TemplatesDir()
	if err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "Invalid configuration", err.Error())
		return
	}

	// Eagerly create the new fsstore so we can fail fast if it's unwritable
	// rather than persisting a config that points at a broken location.
	newFs, err := fsstore.New(nextDir)
	if err != nil {
		respondError(w, h.logger, http.StatusBadRequest, "Could not open templates directory", err.Error())
		return
	}

	if err := config.Save(h.configPath, next); err != nil {
		respondError(w, h.logger, http.StatusInternalServerError, "Could not save config", err.Error())
		return
	}

	if h.store != nil {
		h.store.Swap(newFs)
	}
	h.cfg = next
	h.templatesDir = nextDir
	h.logger.Info("Config updated",
		zap.String("project_root", next.ProjectRoot),
		zap.String("templates_dir", nextDir))

	respondJSON(w, h.logger, http.StatusOK, h.snapshot())
}

// snapshot builds the response payload from the current state. The caller
// must hold the mutex (read or write).
func (h *ConfigHandler) snapshot() ConfigResponse {
	return ConfigResponse{
		ProjectRoot:     h.cfg.ProjectRoot,
		TemplateSubdir:  h.cfg.TemplateSubdir,
		Port:            h.cfg.Port,
		AutoOpenBrowser: h.cfg.AutoOpenBrowser,
		OzxRoomViewPath: h.cfg.OzxRoomViewPath,
		TemplatesDir:    h.templatesDir,
		ConfigPath:      h.configPath,
		UsesFallback:    h.cfg.UsesFallback(),
	}
}

