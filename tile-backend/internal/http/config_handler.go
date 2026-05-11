package http

import (
	"net/http"

	"tile-backend/internal/config"

	"go.uber.org/zap"
)

// ConfigHandler exposes read-only access to the active user configuration.
// The frontend uses GET /api/v1/config to display the current OZX project
// folder and decide whether it's running against a local-mode backend.
type ConfigHandler struct {
	cfg          config.Config
	configPath   string
	templatesDir string
	logger       *zap.Logger
}

// NewConfigHandler builds a ConfigHandler around an already-loaded Config.
// configPath is the absolute path the config was read from (shown to users so
// they know what file to edit). templatesDir is the resolved templates root
// (project_root + template_subdir, or the fallback).
func NewConfigHandler(cfg config.Config, configPath, templatesDir string, logger *zap.Logger) *ConfigHandler {
	return &ConfigHandler{
		cfg:          cfg,
		configPath:   configPath,
		templatesDir: templatesDir,
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

	// Computed: where the fsstore is actually reading and writing.
	TemplatesDir string `json:"templates_dir"`
	// ConfigPath is the absolute path of the config file on disk so users can
	// edit it from the UI.
	ConfigPath string `json:"config_path"`
	// UsesFallback is true when no OZX project_root was configured and we
	// fell back to the per-user templates directory.
	UsesFallback bool `json:"uses_fallback"`
}

// GetConfig handles GET /api/v1/config.
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	resp := ConfigResponse{
		ProjectRoot:     h.cfg.ProjectRoot,
		TemplateSubdir:  h.cfg.TemplateSubdir,
		Port:            h.cfg.Port,
		AutoOpenBrowser: h.cfg.AutoOpenBrowser,
		TemplatesDir:    h.templatesDir,
		ConfigPath:      h.configPath,
		UsesFallback:    h.cfg.UsesFallback(),
	}
	respondJSON(w, h.logger, http.StatusOK, resp)
}
