package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tempConfigPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "subdir", FileName)
}

func TestLoad_MissingWritesDefault(t *testing.T) {
	path := tempConfigPath(t)

	cfg, resolved, wrote, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !wrote {
		t.Errorf("expected wroteDefault=true on first run")
	}
	if resolved != path {
		t.Errorf("resolved: want %s, got %s", path, resolved)
	}
	if cfg != Default() {
		t.Errorf("default config mismatch: %+v", cfg)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("default config not written: %v", err)
	}
}

func TestLoad_ExistingFile(t *testing.T) {
	path := tempConfigPath(t)
	want := Config{
		ProjectRoot:     "/tmp/ozx",
		TemplateSubdir:  "Assets/Tilemaps",
		Port:            9000,
		AutoOpenBrowser: false,
	}
	if err := Save(path, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, _, wrote, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if wrote {
		t.Errorf("expected wroteDefault=false for existing file")
	}
	if got != want {
		t.Errorf("config mismatch:\n got %+v\nwant %+v", got, want)
	}
}

func TestLoad_FillsMissingFields(t *testing.T) {
	path := tempConfigPath(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Partial config: only ProjectRoot set; TemplateSubdir and Port empty.
	partial := map[string]interface{}{"project_root": "/tmp/ozx"}
	data, _ := json.Marshal(partial)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, _, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.TemplateSubdir != Default().TemplateSubdir {
		t.Errorf("TemplateSubdir not defaulted: %q", cfg.TemplateSubdir)
	}
	if cfg.Port != Default().Port {
		t.Errorf("Port not defaulted: %d", cfg.Port)
	}
	if cfg.ProjectRoot != "/tmp/ozx" {
		t.Errorf("ProjectRoot lost: %q", cfg.ProjectRoot)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	path := tempConfigPath(t)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if _, _, _, err := Load(path); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoad_DefaultPathWhenEmpty(t *testing.T) {
	// Redirect HOME so we don't touch the real user config dir.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// Windows uses APPDATA; harmless on other OSes.
	t.Setenv("APPDATA", tmp)

	_, resolved, wrote, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\"): %v", err)
	}
	if !wrote {
		t.Errorf("expected default file to be written under HOME")
	}
	if !strings.HasPrefix(resolved, tmp) {
		t.Errorf("resolved path %q not under tmp %q", resolved, tmp)
	}
	if !strings.HasSuffix(resolved, FileName) {
		t.Errorf("resolved path doesn't end in %s: %s", FileName, resolved)
	}
}

func TestTemplatesDir_WithProjectRoot(t *testing.T) {
	projectRoot := t.TempDir()
	cfg := Config{
		ProjectRoot:    projectRoot,
		TemplateSubdir: "Assets/Resources/TilemapData",
	}
	got, err := cfg.TemplatesDir()
	if err != nil {
		t.Fatalf("TemplatesDir: %v", err)
	}
	want := filepath.Join(projectRoot, "Assets/Resources/TilemapData")
	if got != want {
		t.Errorf("TemplatesDir: want %s, got %s", want, got)
	}
	if cfg.UsesFallback() {
		t.Errorf("UsesFallback: want false")
	}
}

func TestTemplatesDir_MissingProjectRootErrors(t *testing.T) {
	cfg := Config{ProjectRoot: "/does/not/exist/anywhere", TemplateSubdir: "x"}
	if _, err := cfg.TemplatesDir(); err == nil {
		t.Fatal("expected error for missing project_root")
	}
}

func TestTemplatesDir_FallbackWhenEmpty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := Config{}
	got, err := cfg.TemplatesDir()
	if err != nil {
		t.Fatalf("TemplatesDir: %v", err)
	}
	if !cfg.UsesFallback() {
		t.Errorf("UsesFallback: want true")
	}
	if !strings.HasPrefix(got, tmp) {
		t.Errorf("fallback %q not under tmp HOME %q", got, tmp)
	}
}

func TestSave_AtomicAndReadable(t *testing.T) {
	path := tempConfigPath(t)
	cfg := Config{ProjectRoot: "/tmp/ozx", TemplateSubdir: "a/b", Port: 1234, AutoOpenBrowser: true}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// Confirm no .tmp file left behind.
	entries, _ := os.ReadDir(filepath.Dir(path))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("stale temp file: %s", e.Name())
		}
	}
	// Confirm file is human-readable JSON.
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"project_root"`) {
		t.Errorf("saved file not human-readable JSON: %s", data)
	}
}
