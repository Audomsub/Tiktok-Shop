package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all backend environment configuration parameters
type Config struct {
	Port                string
	Env                 string
	DatabaseURL         string
	SupabaseURL         string
	SupabaseServiceKey  string
	SupabaseJWTSecret   string
	InternalAPIKey      string
	CORSAllowedOrigins  string
}

// Load reads configuration from .env and environment variables
func Load() (*Config, error) {
	// Attempt loading .env file (ignore error if file does not exist, e.g. in container environments)
	_ = godotenv.Load()

	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		Env:                getEnv("ENV", "development"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseServiceKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseJWTSecret:  getEnv("SUPABASE_JWT_SECRET", ""),
		InternalAPIKey:     getEnv("INTERNAL_API_KEY", "default-dev-secret-key"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required but not set")
	}

	return cfg, nil
}

// getEnv retrieves environment variable with a fallback default value
func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
