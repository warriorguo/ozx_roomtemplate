// Package config loads and persists the local-client editor's user
// configuration: which OZX project folder to open, where the room templates
// live inside it, and a few runtime knobs (port, auto-open-browser).
//
// The config file is a single JSON document under a per-user config directory.
// On first run the binary writes a default and logs the path so the user can
// edit it; subsequent runs read the same file.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// AppName is the directory under the user config root that holds our files.
const AppName = "ozx-roomeditor"

// FileName is the config file's basename.
const FileName = "config.json"

// Config is the on-disk schema. The JSON tags are the public contract.
type Config struct {
	// ProjectRoot is the absolute path to an OZX project (e.g. an ozx_base
	// checkout). May be empty on first run — the binary will still start, but
	// templates will be stored under a per-user fallback directory.
	ProjectRoot string `json:"project_root"`

	// TemplateSubdir is the path relative to ProjectRoot where room template
	// JSON files live. Defaults to "Assets/Resources/TilemapData" to match
	// OZX's Unity Resources convention.
	TemplateSubdir string `json:"template_subdir"`

	// Port is the local HTTP listen port for the editor server.
	Port int `json:"port"`

	// AutoOpenBrowser controls whether the bundled binary launches the default
	// browser on start (used by ORT-68).
	AutoOpenBrowser bool `json:"auto_open_browser"`
}

// Default returns the config written on first run.
func Default() Config {
	return Config{
		ProjectRoot:     "",
		TemplateSubdir:  "Assets/Resources/TilemapData",
		Port:            8090,
		AutoOpenBrowser: true,
	}
}

// DefaultPath returns the platform-appropriate location of the config file.
//
//   - macOS/Linux: $HOME/.config/ozx-roomeditor/config.json
//   - Windows:     %APPDATA%/ozx-roomeditor/config.json
//
// This deliberately deviates from os.UserConfigDir() on macOS (which would
// return ~/Library/Application Support) to match the spec in ORT-67 and keep
// the file in a predictable, user-editable place.
func DefaultPath() (string, error) {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", errors.New("config: %APPDATA% is not set")
		}
		return filepath.Join(appData, AppName, FileName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: resolve home: %w", err)
	}
	return filepath.Join(home, ".config", AppName, FileName), nil
}

// Load reads the config at path. If path is empty, the platform default is
// used. If the file does not exist, a default config is written to that path
// and returned, with wroteDefault=true so callers can log the new location.
func Load(path string) (cfg Config, resolvedPath string, wroteDefault bool, err error) {
	if path == "" {
		path, err = DefaultPath()
		if err != nil {
			return Config{}, "", false, err
		}
	}
	resolvedPath = path

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		// Fall through to unmarshal.
	case errors.Is(err, os.ErrNotExist):
		cfg = Default()
		if err := write(path, cfg); err != nil {
			return Config{}, resolvedPath, false, err
		}
		return cfg, resolvedPath, true, nil
	default:
		return Config{}, resolvedPath, false, fmt.Errorf("config: read %s: %w", path, err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, resolvedPath, false, fmt.Errorf("config: parse %s: %w", path, err)
	}
	// Fill in any missing fields with defaults so old/partial files still work.
	def := Default()
	if cfg.TemplateSubdir == "" {
		cfg.TemplateSubdir = def.TemplateSubdir
	}
	if cfg.Port == 0 {
		cfg.Port = def.Port
	}
	return cfg, resolvedPath, false, nil
}

// Save writes the config to path, creating parent directories as needed.
func Save(path string, cfg Config) error {
	return write(path, cfg)
}

// TemplatesDir returns the absolute directory where templates are stored,
// derived from ProjectRoot + TemplateSubdir. If ProjectRoot is empty, a
// per-user fallback directory is returned instead so the editor still works
// before the user has configured an OZX project.
//
// The returned path is NOT created — that responsibility belongs to fsstore.
func (c Config) TemplatesDir() (string, error) {
	if c.ProjectRoot == "" {
		return fallbackTemplatesDir()
	}
	abs, err := filepath.Abs(c.ProjectRoot)
	if err != nil {
		return "", fmt.Errorf("config: resolve project_root: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		return "", fmt.Errorf("config: project_root not accessible: %w", err)
	}
	return filepath.Join(abs, c.TemplateSubdir), nil
}

// UsesFallback reports whether TemplatesDir() would return the per-user
// fallback path instead of an OZX-project-relative one. Useful for logging.
func (c Config) UsesFallback() bool { return c.ProjectRoot == "" }

func write(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("config: mkdir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("config: write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("config: rename: %w", err)
	}
	return nil
}

func fallbackTemplatesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./templates", nil //nolint:nilerr // best-effort fallback
	}
	return filepath.Join(home, ".local", "share", AppName, "templates"), nil
}
