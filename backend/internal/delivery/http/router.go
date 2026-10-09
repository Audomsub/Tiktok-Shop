package http

import (
	"encoding/json"
	"net/http"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/delivery/http/handlers"
	customMiddleware "github.com/Audomsub/Tiktok-Shop/backend/internal/delivery/http/middleware"
	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// RouterConfig holds dependencies required to construct the HTTP router
type RouterConfig struct {
	AllowedOrigins     string
	HealthHandler      *handlers.HealthHandler
	JobHandler         *handlers.JobHandler
	LeaderboardHandler *handlers.LeaderboardHandler
	CategoryHandler    *handlers.CategoryHandler
	ProductHandler     *handlers.ProductHandler
	TrendHandler       *handlers.TrendHandler
	FavoriteHandler    *handlers.FavoriteHandler
	AuthMiddleware     *customMiddleware.AuthMiddleware
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

		// Public Product Taxonomy
		if cfg.CategoryHandler != nil {
			r.Get("/categories", cfg.CategoryHandler.GetCategories)
		}

		// Historical Trends
		if cfg.TrendHandler != nil {
			r.Get("/products/{id}/trends", cfg.TrendHandler.GetProductTrends)
		}

		// Ingestion & calculation jobs
		if cfg.JobHandler != nil {
			r.Post("/jobs/compute-scores", cfg.JobHandler.ComputeScores)
		}

		// Winning Leaderboard & Product Catalog (Optional Auth to enrich is_favorited)
		r.Group(func(optional chi.Router) {
			if cfg.AuthMiddleware != nil {
				optional.Use(cfg.AuthMiddleware.OptionalAuth)
			}
			if cfg.LeaderboardHandler != nil {
				optional.Get("/leaderboard", cfg.LeaderboardHandler.GetLeaderboard)
			}
			if cfg.ProductHandler != nil {
				optional.Get("/products/export", cfg.ProductHandler.ExportCatalog)
				optional.Get("/products", cfg.ProductHandler.GetCatalog)
			}
		})

		// Protected endpoints (Requires Supabase JWT)
		if cfg.AuthMiddleware != nil {
			r.Group(func(protected chi.Router) {
				protected.Use(cfg.AuthMiddleware.RequireAuth)

				protected.Get("/auth/me", func(w http.ResponseWriter, r *http.Request) {
					userID, _ := domain.GetUserIDFromContext(r.Context())
					email, _ := domain.GetUserEmailFromContext(r.Context())
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]string{
						"user_id": userID,
						"email":   email,
					})
				})

				if cfg.FavoriteHandler != nil {
					protected.Get("/favorites", cfg.FavoriteHandler.GetFavorites)
					protected.Post("/favorites", cfg.FavoriteHandler.AddFavorite)
					protected.Delete("/favorites/{productId}", cfg.FavoriteHandler.RemoveFavorite)
				}
			})
		}
	})

	return r
}
