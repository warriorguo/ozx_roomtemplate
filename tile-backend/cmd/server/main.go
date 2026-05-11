package main

import (
	"context"
	"flag"
	"fmt"
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

// runtimeOpts collects everything not derived from the on-disk config file.
type runtimeOpts struct {
	LogLevel           string
	CORSAllowedOrigins []string
}

func main() {
	// --- CLI flags ------------------------------------------------------
	configFlag := flag.String("config", "", "path to config.json (default: ~/.config/ozx-roomeditor/config.json)")
	flag.Parse()

	opts := loadRuntimeOpts()
	logger := initLogger(opts.LogLevel)
	defer logger.Sync()

	// --- Load on-disk config -------------------------------------------
	cfg, configPath, wroteDefault, err := config.Load(*configFlag)
	if err != nil {
		logger.Fatal("Failed to load config", zap.String("path", *configFlag), zap.Error(err))
	}
	if wroteDefault {
		logger.Info("Wrote default config — edit it to point at your OZX project",
			zap.String("path", configPath))
	} else {
		logger.Info("Loaded config", zap.String("path", configPath))
	}

	templatesDir, err := cfg.TemplatesDir()
	if err != nil {
		logger.Fatal("Invalid templates directory in config", zap.Error(err))
	}
	if cfg.UsesFallback() {
		logger.Warn("No OZX project configured — using per-user fallback templates directory",
			zap.String("templates_dir", templatesDir),
			zap.String("hint", "edit config.json and set project_root"))
	}

	// --- Wire stores ---------------------------------------------------
	var templateStore store.Store
	if fs, err := fsstore.New(templatesDir); err != nil {
		logger.Warn("Falling back to stub store",
			zap.String("templates_dir", templatesDir),
			zap.Error(err))
		templateStore = store.NewStubStore()
	} else {
		logger.Info("Filesystem store ready", zap.String("templates_dir", fs.RootDir()))
		templateStore = fs
	}

	// --- HTTP ----------------------------------------------------------
	cfgHandler := httpHandler.NewConfigHandler(cfg, configPath, templatesDir, logger)
	router := httpHandler.SetupRouter(templateStore, cfgHandler, logger, opts.CORSAllowedOrigins)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Starting server", zap.Int("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// --- Graceful shutdown ---------------------------------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// loadRuntimeOpts pulls process-runtime options (logging, CORS) from env vars.
// These deliberately stay separate from the on-disk config file because they
// are deployment-environment knobs, not editor settings the user would tweak.
func loadRuntimeOpts() *runtimeOpts {
	opts := &runtimeOpts{
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "")
	if corsOrigins != "" {
		opts.CORSAllowedOrigins = strings.Split(corsOrigins, ",")
		for i, origin := range opts.CORSAllowedOrigins {
			opts.CORSAllowedOrigins[i] = strings.TrimSpace(origin)
		}
	}
	return opts
}

// initLogger initializes the zap logger
func initLogger(level string) *zap.Logger {
	var zapLevel zap.AtomicLevel
	switch level {
	case "debug":
		zapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zapLevel
	cfg.DisableStacktrace = true

	logger, err := cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	return logger
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
