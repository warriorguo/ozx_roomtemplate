package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"tile-backend/internal/config"
	"tile-backend/internal/model"
	"tile-backend/internal/store"
	"tile-backend/internal/store/fsstore"

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
	swap := store.NewSwappableStore(store.NewStubStore())
	cfgHandler := NewConfigHandler(cfg, configPath, templatesDir, swap, logger)
	router := SetupRouter(swap, cfgHandler, logger, nil)
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
	swap := store.NewSwappableStore(store.NewStubStore())
	cfgHandler := NewConfigHandler(cfg, "/tmp/config.json", "/tmp/templates", swap, zap.NewNop())
	router := SetupRouter(swap, cfgHandler, zap.NewNop(), nil)
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

// TestPutConfig_SwapsStoreAndPersistsToDisk verifies the full "Switch
// project" flow: PUT /api/v1/config with a new project_root, confirm the
// response, that the file on disk is updated, and that POST /templates
// followed by GET /templates uses the new directory.
func TestPutConfig_SwapsStoreAndPersistsToDisk(t *testing.T) {
	// Set up an initial project pointing at projectA, plus a future projectB.
	projectA := t.TempDir()
	projectB := t.TempDir()
	configFile := filepath.Join(t.TempDir(), "config.json")

	initial := config.Config{
		ProjectRoot:     projectA,
		TemplateSubdir:  "templates",
		Port:            8090,
		AutoOpenBrowser: false,
	}
	if err := config.Save(configFile, initial); err != nil {
		t.Fatalf("Save initial: %v", err)
	}
	initialDir, _ := initial.TemplatesDir()
	fs, err := fsstore.New(initialDir)
	if err != nil {
		t.Fatalf("fsstore.New: %v", err)
	}
	swap := store.NewSwappableStore(fs)
	cfgHandler := NewConfigHandler(initial, configFile, initialDir, swap, zap.NewNop())
	router := SetupRouter(swap, cfgHandler, zap.NewNop(), nil)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	// Drop a template into projectA so we can prove it does NOT show up after
	// we switch to projectB.
	if _, err := fs.Create(context.Background(), fixture("only-in-A")); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// PUT /api/v1/config → switch to projectB.
	body := []byte(`{"project_root":` + jsonString(projectB) + `}`)
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /config: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT status: want 200, got %d", resp.StatusCode)
	}
	var got ConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp.Body.Close()
	if got.ProjectRoot != projectB {
		t.Errorf("ProjectRoot after PUT: want %q, got %q", projectB, got.ProjectRoot)
	}
	wantTemplatesDir := filepath.Join(projectB, "templates")
	if got.TemplatesDir != wantTemplatesDir {
		t.Errorf("TemplatesDir after PUT: want %q, got %q", wantTemplatesDir, got.TemplatesDir)
	}

	// Config file on disk should reflect the change.
	persisted, _, _, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	if persisted.ProjectRoot != projectB {
		t.Errorf("persisted ProjectRoot: want %q, got %q", projectB, persisted.ProjectRoot)
	}

	// GET /api/v1/templates should now hit projectB and see no templates.
	tmplResp, err := http.Get(srv.URL + "/api/v1/templates")
	if err != nil {
		t.Fatalf("GET /templates: %v", err)
	}
	defer tmplResp.Body.Close()
	var list model.ListTemplatesResponse
	if err := json.NewDecoder(tmplResp.Body).Decode(&list); err != nil {
		t.Fatalf("decode templates: %v", err)
	}
	if list.Total != 0 {
		t.Errorf("templates after switch: want 0, got %d", list.Total)
	}
}

func TestPutConfig_RejectsMissingProjectRoot(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "config.json")
	initial := config.Default()
	initial.ProjectRoot = t.TempDir()
	if err := config.Save(configFile, initial); err != nil {
		t.Fatalf("Save: %v", err)
	}
	initialDir, _ := initial.TemplatesDir()
	swap := store.NewSwappableStore(store.NewStubStore())
	cfgHandler := NewConfigHandler(initial, configFile, initialDir, swap, zap.NewNop())
	router := SetupRouter(swap, cfgHandler, zap.NewNop(), nil)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	body := []byte(`{"project_root":"/this/does/not/exist/anywhere/at/all"}`)
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /config: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: want 400 for missing dir, got %d", resp.StatusCode)
	}
}

// fixture returns a valid 4x4 template ready to write to a store. Mirrors the
// helper in fsstore_test.go but is local here to avoid importing test code.
func fixture(name string) model.Template {
	return model.Template{
		Name:    name,
		Version: 1,
		Width:   4,
		Height:  4,
		Payload: model.TemplatePayload{
			Ground: [][]int{{1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}, {1, 1, 1, 1}},
			Static: [][]int{{0, 1, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 1, 0, 0}},
			Chaser: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			Zoner:  [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			DPS:    [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			MobAir: [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
			Meta:   model.TemplateMeta{Name: name, Version: 1, Width: 4, Height: 4},
		},
	}
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
