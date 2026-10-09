package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpDelivery "github.com/Audomsub/Tiktok-Shop/backend/internal/delivery/http"
	"github.com/Audomsub/Tiktok-Shop/backend/internal/delivery/http/handlers"
	customMiddleware "github.com/Audomsub/Tiktok-Shop/backend/internal/delivery/http/middleware"
	"github.com/Audomsub/Tiktok-Shop/backend/internal/infrastructure/config"
	"github.com/Audomsub/Tiktok-Shop/backend/internal/infrastructure/database"
	"github.com/Audomsub/Tiktok-Shop/backend/internal/usecase"
)

func main() {
	log.Println("Initializing TikTok Affiliate Analytics Backend Service...")

	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Fatal configuration error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Initialize Database Connection Pool
	dbPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Fatal database connection error: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection pool established successfully.")

	// 3. Dependency Injection (Clean Architecture)
	healthRepo := database.NewPostgresHealthRepo(dbPool)
	healthUsecase := usecase.NewHealthUsecase(healthRepo)
	healthHandler := handlers.NewHealthHandler(healthUsecase)

	snapshotRepo := database.NewPostgresSnapshotRepo(dbPool)
	crawlLogRepo := database.NewPostgresCrawlLogRepo(dbPool)
	velocityUsecase := usecase.NewVelocityUsecase()
	scoringUsecase := usecase.NewScoringUsecase()
	analyticsUsecase := usecase.NewAnalyticsUsecase(snapshotRepo, crawlLogRepo, velocityUsecase, scoringUsecase)
	jobHandler := handlers.NewJobHandler(analyticsUsecase, cfg.InternalAPIKey)

	favoriteRepo := database.NewPostgresFavoriteRepo(dbPool)
	favoriteUsecase := usecase.NewFavoriteUsecase(favoriteRepo)
	favoriteHandler := handlers.NewFavoriteHandler(favoriteUsecase)

	leaderboardRepo := database.NewPostgresLeaderboardRepo(dbPool)
	leaderboardUsecase := usecase.NewLeaderboardUsecase(leaderboardRepo, favoriteRepo)
	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardUsecase)

	categoryRepo := database.NewPostgresCategoryRepo(dbPool)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryUsecase)

	productRepo := database.NewPostgresProductRepo(dbPool)
	productUsecase := usecase.NewProductUsecase(productRepo, favoriteRepo)
	productHandler := handlers.NewProductHandler(productUsecase)

	trendRepo := database.NewPostgresTrendRepo(dbPool)
	trendUsecase := usecase.NewTrendUsecase(trendRepo)
	trendHandler := handlers.NewTrendHandler(trendUsecase)

	authMiddleware := customMiddleware.NewAuthMiddleware(cfg.SupabaseJWTSecret)

	// 4. Router Construction
	router := httpDelivery.NewRouter(httpDelivery.RouterConfig{
		AllowedOrigins:     cfg.CORSAllowedOrigins,
		HealthHandler:      healthHandler,
		JobHandler:         jobHandler,
		LeaderboardHandler: leaderboardHandler,
		CategoryHandler:    categoryHandler,
		ProductHandler:     productHandler,
		TrendHandler:       trendHandler,
		FavoriteHandler:    favoriteHandler,
		AuthMiddleware:     authMiddleware,
	})

	// 5. Server Configuration
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6. Start HTTP Server asynchronously
	go func() {
		log.Printf("Server listening on http://localhost:%s (ENV: %s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 7. Graceful Shutdown listener
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Graceful shutdown signal received. Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exited cleanly.")
}
