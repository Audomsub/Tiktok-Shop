package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type cachedClaim struct {
	claims    *domain.SupabaseClaims
	expiresAt time.Time
}

// AuthMiddleware provides JWT authentication guards for HTTP routes
type AuthMiddleware struct {
	jwtSecret   []byte
	supabaseURL string
	supabaseKey string
	tokenCache  sync.Map
	httpClient  *http.Client
}

// NewAuthMiddleware constructs a new AuthMiddleware instance with Supabase JWT secret and optional Supabase URL/Key
func NewAuthMiddleware(jwtSecret string, opts ...string) *AuthMiddleware {
	var supabaseURL, supabaseKey string
	if len(opts) > 0 {
		supabaseURL = strings.TrimRight(strings.TrimSpace(opts[0]), "/")
	}
	if len(opts) > 1 {
		supabaseKey = strings.TrimSpace(opts[1])
	}

	return &AuthMiddleware{
		jwtSecret:   []byte(jwtSecret),
		supabaseURL: supabaseURL,
		supabaseKey: supabaseKey,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
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
	// 1. Check in-memory cache
	if val, ok := m.tokenCache.Load(tokenStr); ok {
		entry := val.(*cachedClaim)
		if time.Now().Before(entry.expiresAt) {
			return entry.claims, nil
		}
		m.tokenCache.Delete(tokenStr)
	}

	claims := &domain.SupabaseClaims{}

	// 2. Try local HMAC verification if token is HS256 and secret is configured
	if len(m.jwtSecret) > 0 {
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing algorithm: %v", token.Header["alg"])
			}
			return m.jwtSecret, nil
		})
		if err == nil && token.Valid {
			exp := time.Now().Add(5 * time.Minute)
			if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(exp) {
				exp = claims.ExpiresAt.Time
			}
			m.tokenCache.Store(tokenStr, &cachedClaim{claims: claims, expiresAt: exp})
			return claims, nil
		}
	}

	// 3. Fallback: Validate via Supabase Auth API (Supports ES256 & asymmetric signing)
	if m.supabaseURL != "" {
		req, err := http.NewRequest(http.MethodGet, m.supabaseURL+"/auth/v1/user", nil)
		if err == nil {
			req.Header.Set("Authorization", "Bearer "+tokenStr)
			if m.supabaseKey != "" {
				req.Header.Set("apikey", m.supabaseKey)
			}

			resp, err := m.httpClient.Do(req)
			if err == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var su struct {
						ID    string `json:"id"`
						Email string `json:"email"`
						Role  string `json:"role"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&su); err == nil && su.ID != "" {
						verifiedClaims := &domain.SupabaseClaims{
							Email: su.Email,
							Role:  su.Role,
							RegisteredClaims: jwt.RegisteredClaims{
								Subject: su.ID,
							},
						}
						m.tokenCache.Store(tokenStr, &cachedClaim{
							claims:    verifiedClaims,
							expiresAt: time.Now().Add(5 * time.Minute),
						})
						return verifiedClaims, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("token signature is invalid or unverifiable")
}

// writeUnauthorized returns a standard JSON error response with HTTP 401
func writeUnauthorized(w http.ResponseWriter, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": fmt.Sprintf("unauthorized: %s", reason),
	})
}
