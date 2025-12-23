package http

import (
	"tile-backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

// SetupRouter creates and configures the HTTP router
func SetupRouter(templateStore store.TemplateStore, logger *zap.Logger, corsOrigins []string) *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(RecoveryMiddleware(logger))
	r.Use(LoggerMiddleware(logger))
	r.Use(RequestSizeLimitMiddleware(2 * 1024 * 1024)) // 2MB limit

	// CORS configuration
	if len(corsOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   corsOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}))
	}

	// Create handlers
	templateHandler := NewTemplateHandler(templateStore, logger)

	// Health check endpoint
	r.Get("/health", templateHandler.HealthCheck)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/templates", func(r chi.Router) {
			r.Post("/", templateHandler.CreateTemplate)
			r.Get("/", templateHandler.ListTemplates)
			r.Get("/{id}", templateHandler.GetTemplate)
			r.Post("/validate", templateHandler.ValidateTemplate)
		})
	})

	return r
}