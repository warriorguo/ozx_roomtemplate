package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	httpHandler "tile-backend/internal/http"
	"tile-backend/internal/store"
	"tile-backend/internal/store/fsstore"

	"go.uber.org/zap"
)

type Config struct {
	Port               int
	LogLevel           string
	CORSAllowedOrigins []string
	TemplatesDir       string
}

func main() {
	// Load configuration
	config := loadConfig()

	// Initialize logger
	logger := initLogger(config.LogLevel)
	defer logger.Sync()

	// Wire the filesystem-backed store. The proper config-driven path arrives
	// in ORT-67; for now the path falls back to a per-user default if
	// TEMPLATES_DIR is not set.
	var templateStore store.Store
	if fs, err := fsstore.New(config.TemplatesDir); err != nil {
		logger.Warn("Falling back to stub store",
			zap.String("templates_dir", config.TemplatesDir),
			zap.Error(err))
		templateStore = store.NewStubStore()
	} else {
		logger.Info("Filesystem store ready", zap.String("templates_dir", fs.RootDir()))
		templateStore = fs
	}

	// Setup router
	router := httpHandler.SetupRouter(templateStore, logger, config.CORSAllowedOrigins)

	// Setup HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.Int("port", config.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	config := &Config{
		Port:         getEnvInt("PORT", 8090),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		TemplatesDir: getEnv("TEMPLATES_DIR", defaultTemplatesDir()),
	}

	// Parse CORS origins
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "")
	if corsOrigins != "" {
		config.CORSAllowedOrigins = strings.Split(corsOrigins, ",")
		for i, origin := range config.CORSAllowedOrigins {
			config.CORSAllowedOrigins[i] = strings.TrimSpace(origin)
		}
	}

	return config
}

// defaultTemplatesDir picks a sensible per-user location until ORT-67 lands a
// proper config file that points at an OZX project folder.
func defaultTemplatesDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "share", "ozx-roomeditor", "templates")
	}
	return "./templates"
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

	config := zap.NewProductionConfig()
	config.Level = zapLevel
	config.DisableStacktrace = true

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	return logger
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
