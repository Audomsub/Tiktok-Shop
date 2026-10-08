package http

import (
	"net/http"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/delivery/http/handlers"
	customMiddleware "github.com/Audomsub/Tiktok-Shop/backend/internal/delivery/http/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// RouterConfig holds dependencies required to construct the HTTP router
type RouterConfig struct {
	AllowedOrigins string
	HealthHandler  *handlers.HealthHandler
	JobHandler         *handlers.JobHandler
	LeaderboardHandler *handlers.LeaderboardHandler
	CategoryHandler    *handlers.CategoryHandler
	ProductHandler     *handlers.ProductHandler
}

// NewRouter constructs and configures the application HTTP router
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Standard resilient middlewares
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(customMiddleware.EnableCORS(cfg.AllowedOrigins))

	// Healthcheck endpoint
	r.Get("/health", cfg.HealthHandler.Check)

	// API v1 Sub-router
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message":"pong"}`))
		})

		// Public Winning Leaderboard
		if cfg.LeaderboardHandler != nil {
			r.Get("/leaderboard", cfg.LeaderboardHandler.GetLeaderboard)
		}

		// Public Product Taxonomy & Catalog
		if cfg.CategoryHandler != nil {
			r.Get("/categories", cfg.CategoryHandler.GetCategories)
		}
		if cfg.ProductHandler != nil {
			r.Get("/products", cfg.ProductHandler.GetCatalog)
		}

		// Ingestion & calculation jobs
		if cfg.JobHandler != nil {
			r.Post("/jobs/compute-scores", cfg.JobHandler.ComputeScores)
		}
	})

	return r
}
