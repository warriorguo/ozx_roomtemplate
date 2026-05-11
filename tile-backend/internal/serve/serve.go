// Package serve wires together config, store, HTTP router, and graceful
// shutdown — the bits the cmd/server and cmd/ozx-roomeditor entrypoints
// would otherwise duplicate.
package serve

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"tile-backend/internal/config"
	httpHandler "tile-backend/internal/http"
	"tile-backend/internal/store"
	"tile-backend/internal/store/fsstore"

	"go.uber.org/zap"
)

// Options bundles everything an entrypoint can vary about the running server.
type Options struct {
	// ConfigPath, if empty, falls back to config.DefaultPath(). Pass through
	// from a CLI --config flag.
	ConfigPath string

	// FrontendFS, if non-nil, is registered as a catch-all at "/" so the
	// embedded SPA can be served alongside /api/v1. cmd/server (API-only)
	// passes nil; cmd/ozx-roomeditor passes the go:embed dist/*.
	FrontendFS fs.FS

	// OnReady runs once the HTTP server is bound and accepting requests; it
	// receives the resolved URL ("http://localhost:<port>/") and the loaded
	// config so callers can trigger a browser launch from the standalone
	// binary based on cfg.AutoOpenBrowser. May be nil.
	OnReady func(url string, cfg config.Config)

	// CORSAllowedOrigins is parsed from CORS_ALLOWED_ORIGINS by the caller
	// (which knows whether to read env vars).
	CORSAllowedOrigins []string

	// Logger is required.
	Logger *zap.Logger
}

// Run loads config, wires the store, starts the HTTP server, optionally
// invokes OnReady, and blocks until SIGINT/SIGTERM.
func Run(opts Options) error {
	if opts.Logger == nil {
		return errors.New("serve.Run: Logger is required")
	}
	logger := opts.Logger

	cfg, configPath, wroteDefault, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if wroteDefault {
		logger.Info("Wrote default config — edit it to point at your OZX project",
			zap.String("path", configPath))
	} else {
		logger.Info("Loaded config", zap.String("path", configPath))
	}

	templatesDir, err := cfg.TemplatesDir()
	if err != nil {
		return fmt.Errorf("resolve templates dir: %w", err)
	}
	if cfg.UsesFallback() {
		logger.Warn("No OZX project configured — using per-user fallback templates directory",
			zap.String("templates_dir", templatesDir),
			zap.String("hint", "edit config.json and set project_root"))
	}

	var initialStore store.Store
	if fs, err := fsstore.New(templatesDir); err != nil {
		logger.Warn("Falling back to stub store",
			zap.String("templates_dir", templatesDir),
			zap.Error(err))
		initialStore = store.NewStubStore()
	} else {
		logger.Info("Filesystem store ready", zap.String("templates_dir", fs.RootDir()))
		initialStore = fs
	}
	swappable := store.NewSwappableStore(initialStore)

	cfgHandler := httpHandler.NewConfigHandler(cfg, configPath, templatesDir, swappable, logger)
	router := httpHandler.SetupRouter(swappable, cfgHandler, logger, opts.CORSAllowedOrigins)

	if opts.FrontendFS != nil {
		if err := httpHandler.MountFrontend(router, opts.FrontendFS, logger); err != nil {
			return fmt.Errorf("mount frontend: %w", err)
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("Starting server", zap.Int("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	if opts.OnReady != nil {
		url := fmt.Sprintf("http://localhost:%d/", cfg.Port)
		// Brief sleep gives ListenAndServe a moment to bind before we hand
		// the URL off to a browser opener; if the bind fails we surface that
		// error from the select below shortly after.
		go func() {
			time.Sleep(150 * time.Millisecond)
			opts.OnReady(url, cfg)
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server: %w", err)
	case sig := <-quit:
		logger.Info("Shutting down server", zap.String("signal", sig.String()))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}
	logger.Info("Server exited")
	return nil
}

// CORSOriginsFromEnv parses a comma-separated env var into a clean slice.
// Used by both entrypoints — kept here so the parsing rule is in one place.
func CORSOriginsFromEnv(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
