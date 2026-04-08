package http

import (
	"tile-backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

// SetupRouter creates and configures the HTTP router
func SetupRouter(templateStore store.TemplateStore, projectStore store.ProjectStore, logger *zap.Logger, corsOrigins []string) *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(RecoveryMiddleware(logger))
	r.Use(LoggerMiddleware(logger))
	r.Use(RequestSizeLimitMiddleware(2 * 1024 * 1024)) // 2MB limit

	// CORS configuration - allow all origins for development
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link", "Content-Length"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Create handlers
	templateHandler := NewTemplateHandler(templateStore, logger)
	projectHandler := NewProjectHandler(projectStore, logger)

	// Health check endpoint
	r.Get("/health", templateHandler.HealthCheck)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/templates", func(r chi.Router) {
			r.Post("/", templateHandler.CreateTemplate)
			r.Get("/", templateHandler.ListTemplates)
			r.Get("/{id}", templateHandler.GetTemplate)
			r.Delete("/{id}", templateHandler.DeleteTemplate)
			r.Post("/validate", templateHandler.ValidateTemplate)
		})

		// Project endpoints
		r.Route("/projects", func(r chi.Router) {
			r.Post("/", projectHandler.CreateProject)
			r.Get("/", projectHandler.ListProjects)
			r.Get("/{id}", projectHandler.GetProject)
			r.Get("/{id}/stats", projectHandler.GetProjectStats)
			r.Put("/{id}", projectHandler.UpdateProject)
			r.Delete("/{id}", projectHandler.DeleteProject)
		})

		// Generation endpoints
		r.Route("/generate", func(r chi.Router) {
			r.Post("/bridge", templateHandler.GenerateBridge)
			r.Post("/platform", templateHandler.GeneratePlatform)
			r.Post("/fullroom", templateHandler.GenerateFullRoom)
		})

		// Stage config endpoint
		r.Get("/stage-configs", templateHandler.GetStageConfigs)
	})

	return r
}
