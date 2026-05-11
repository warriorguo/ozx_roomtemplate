package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tile-backend/internal/config"
	"tile-backend/internal/store"

	"go.uber.org/zap"
)

func TestConfigEndpoint_ReturnsResolvedConfig(t *testing.T) {
	cfg := config.Config{
		ProjectRoot:     "/Users/andrew/Codes/ozx_base",
		TemplateSubdir:  "Assets/Resources/TilemapData",
		Port:            8090,
		AutoOpenBrowser: true,
	}
	configPath := "/Users/andrew/.config/ozx-roomeditor/config.json"
	templatesDir := "/Users/andrew/Codes/ozx_base/Assets/Resources/TilemapData"

	logger := zap.NewNop()
	cfgHandler := NewConfigHandler(cfg, configPath, templatesDir, logger)
	router := SetupRouter(store.NewStubStore(), cfgHandler, logger, nil)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/v1/config")
	if err != nil {
		t.Fatalf("GET /config: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}

	var got ConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ProjectRoot != cfg.ProjectRoot {
		t.Errorf("ProjectRoot: want %q, got %q", cfg.ProjectRoot, got.ProjectRoot)
	}
	if got.TemplateSubdir != cfg.TemplateSubdir {
		t.Errorf("TemplateSubdir: want %q, got %q", cfg.TemplateSubdir, got.TemplateSubdir)
	}
	if got.Port != cfg.Port {
		t.Errorf("Port: want %d, got %d", cfg.Port, got.Port)
	}
	if got.AutoOpenBrowser != cfg.AutoOpenBrowser {
		t.Errorf("AutoOpenBrowser: want %v, got %v", cfg.AutoOpenBrowser, got.AutoOpenBrowser)
	}
	if got.TemplatesDir != templatesDir {
		t.Errorf("TemplatesDir: want %q, got %q", templatesDir, got.TemplatesDir)
	}
	if got.ConfigPath != configPath {
		t.Errorf("ConfigPath: want %q, got %q", configPath, got.ConfigPath)
	}
	if got.UsesFallback {
		t.Errorf("UsesFallback: want false")
	}
}

func TestConfigEndpoint_FallbackFlag(t *testing.T) {
	cfg := config.Config{ProjectRoot: "", TemplateSubdir: "x", Port: 8090}
	cfgHandler := NewConfigHandler(cfg, "/tmp/config.json", "/tmp/templates", zap.NewNop())
	router := SetupRouter(store.NewStubStore(), cfgHandler, zap.NewNop(), nil)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/v1/config")
	if err != nil {
		t.Fatalf("GET /config: %v", err)
	}
	defer resp.Body.Close()
	var got ConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !got.UsesFallback {
		t.Error("UsesFallback should be true when ProjectRoot is empty")
	}
}
