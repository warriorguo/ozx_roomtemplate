// ozx-roomeditor is the standalone, self-contained binary for the
// local-client variant of the editor. It bundles the SPA via go:embed, loads
// the user config, starts the HTTP server, and (optionally) launches the
// default browser pointed at the local URL.
//
// Run with `--config <path>` to use a non-default config location, or with
// no flags to use ~/.config/ozx-roomeditor/config.json.
package main

import (
	"flag"
	"fmt"
	"os"

	"tile-backend/internal/browser"
	"tile-backend/internal/config"
	"tile-backend/internal/serve"
	"tile-backend/internal/web"

	"go.uber.org/zap"
)

func main() {
	configFlag := flag.String("config", "", "path to config.json (default: ~/.config/ozx-roomeditor/config.json)")
	portFlag := flag.Int("port", 0, "override the port from config.json (0 = use config value)")
	noBrowser := flag.Bool("no-browser", false, "do not launch the default browser even if auto_open_browser is true")
	flag.Parse()

	logger := newLogger(getEnv("LOG_LEVEL", "info"))
	defer logger.Sync()

	assets, err := web.Assets()
	if err != nil {
		logger.Fatal("Embedded frontend missing", zap.Error(err))
	}

	opts := serve.Options{
		ConfigPath:         *configFlag,
		PortOverride:       *portFlag,
		FrontendFS:         assets,
		CORSAllowedOrigins: serve.CORSOriginsFromEnv(os.Getenv("CORS_ALLOWED_ORIGINS")),
		Logger:             logger,
		OnReady: func(url string, cfg config.Config) {
			fmt.Println()
			fmt.Println("  ┌──────────────────────────────────────────────────────┐")
			fmt.Printf("  │  OZX Room Editor running at  %-23s │\n", url)
			fmt.Println("  │  Press Ctrl-C to quit.                               │")
			fmt.Println("  └──────────────────────────────────────────────────────┘")
			fmt.Println()

			if *noBrowser || !cfg.AutoOpenBrowser {
				return
			}
			if err := browser.Open(url); err != nil {
				logger.Warn("Could not auto-open browser; copy the URL above into a browser",
					zap.String("url", url),
					zap.Error(err))
			}
		},
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
