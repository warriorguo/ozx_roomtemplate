// server is the API-only entrypoint, intended for local development where
// the frontend is served separately by `npm run dev`. The bundled
// standalone variant lives at cmd/ozx-roomeditor.
package main

import (
	"flag"
	"fmt"
	"os"

	"tile-backend/internal/serve"

	"go.uber.org/zap"
)

func main() {
	configFlag := flag.String("config", "", "path to config.json (default: ~/.config/ozx-roomeditor/config.json)")
	portFlag := flag.Int("port", 0, "override the port from config.json (0 = use config value)")
	flag.Parse()

	logger := newLogger(getEnv("LOG_LEVEL", "info"))
	defer logger.Sync()

	opts := serve.Options{
		ConfigPath:         *configFlag,
		PortOverride:       *portFlag,
		CORSAllowedOrigins: serve.CORSOriginsFromEnv(os.Getenv("CORS_ALLOWED_ORIGINS")),
		Logger:             logger,
	}
	if err := serve.Run(opts); err != nil {
		logger.Fatal("server failed", zap.Error(err))
	}
}

func newLogger(level string) *zap.Logger {
	var zapLevel zap.AtomicLevel
	switch level {
	case "debug":
		zapLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
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
		panic(fmt.Sprintf("logger: %v", err))
	}
	return logger
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
