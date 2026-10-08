package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware provides JWT authentication guards for HTTP routes
type AuthMiddleware struct {
	jwtSecret []byte
}

// NewAuthMiddleware constructs a new AuthMiddleware instance with the Supabase JWT secret
func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(jwtSecret),
	}
}

// RequireAuth enforces a valid Supabase JWT Bearer token on protected endpoints
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeUnauthorized(w, "missing Authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeUnauthorized(w, "malformed Authorization header, expected 'Bearer <token>'")
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		claims, err := m.parseAndValidateToken(tokenStr)
		if err != nil {
			writeUnauthorized(w, fmt.Sprintf("invalid or expired token: %v", err))
			return
		}

		userID := claims.Subject
		if userID == "" {
			writeUnauthorized(w, "missing user id in token claims")
			return
		}

		ctx := context.WithValue(r.Context(), domain.UserIDContextKey, userID)
		if claims.Email != "" {
			ctx = context.WithValue(ctx, domain.UserEmailContextKey, claims.Email)
		}
		if claims.Role != "" {
			ctx = context.WithValue(ctx, domain.UserRoleContextKey, claims.Role)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth parses a Supabase JWT Bearer token if present, but does not block unauthenticated requests
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			next.ServeHTTP(w, r)
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		claims, err := m.parseAndValidateToken(tokenStr)
		if err != nil || claims.Subject == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), domain.UserIDContextKey, claims.Subject)
		if claims.Email != "" {
			ctx = context.WithValue(ctx, domain.UserEmailContextKey, claims.Email)
		}
		if claims.Role != "" {
			ctx = context.WithValue(ctx, domain.UserRoleContextKey, claims.Role)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// parseAndValidateToken decodes and checks the JWT signature and standard claims
func (m *AuthMiddleware) parseAndValidateToken(tokenStr string) (*domain.SupabaseClaims, error) {
	claims := &domain.SupabaseClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing algorithm: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token signature is invalid")
	}

	return claims, nil
}

// writeUnauthorized returns a standard JSON error response with HTTP 401
func writeUnauthorized(w http.ResponseWriter, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": fmt.Sprintf("unauthorized: %s", reason),
	})
}
