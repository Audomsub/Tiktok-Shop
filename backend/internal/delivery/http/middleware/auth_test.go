package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Audomsub/Tiktok-Shop/backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "super-secret-jwt-key-for-unit-tests-12345"

// generateTestToken creates a signed test JWT with configurable claims and secret
func generateTestToken(secret string, sub string, email string, role string, exp time.Time) (string, error) {
	claims := &domain.SupabaseClaims{
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "supabase-gotrue",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	authMiddleware := NewAuthMiddleware(testSecret)

	// Protected handler that verifies user context
	protectedHandler := authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := domain.GetUserIDFromContext(r.Context())
		if !ok || userID == "" {
			t.Errorf("expected user_id in context, but got empty")
		}
		email, _ := domain.GetUserEmailFromContext(r.Context())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"user_id": userID,
			"email":   email,
		})
	}))

	t.Run("Passes with valid signed token and injects user context", func(t *testing.T) {
		validToken, err := generateTestToken(testSecret, "usr-uuid-123", "test@example.com", "authenticated", time.Now().Add(1*time.Hour))
		if err != nil {
			t.Fatalf("failed generating test token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed decoding json response: %v", err)
		}

		if resp["user_id"] != "usr-uuid-123" {
			t.Errorf("expected user_id 'usr-uuid-123', got '%s'", resp["user_id"])
		}
		if resp["email"] != "test@example.com" {
			t.Errorf("expected email 'test@example.com', got '%s'", resp["email"])
		}
	})

	t.Run("Rejects missing Authorization header with 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("Rejects malformed Authorization header with 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		rec := httptest.NewRecorder()

		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("Rejects expired JWT token with 401", func(t *testing.T) {
		expiredToken, err := generateTestToken(testSecret, "usr-uuid-123", "test@example.com", "authenticated", time.Now().Add(-1*time.Hour))
		if err != nil {
			t.Fatalf("failed generating test token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		rec := httptest.NewRecorder()

		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("Rejects forged token signed with wrong secret with 401", func(t *testing.T) {
		forgedToken, err := generateTestToken("wrong-fake-secret", "usr-uuid-123", "test@example.com", "authenticated", time.Now().Add(1*time.Hour))
		if err != nil {
			t.Fatalf("failed generating test token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+forgedToken)
		rec := httptest.NewRecorder()

		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("Rejects token with missing subject (user_id) with 401", func(t *testing.T) {
		tokenWithoutSub, err := generateTestToken(testSecret, "", "test@example.com", "authenticated", time.Now().Add(1*time.Hour))
		if err != nil {
			t.Fatalf("failed generating test token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenWithoutSub)
		rec := httptest.NewRecorder()

		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})
}

func TestAuthMiddleware_OptionalAuth(t *testing.T) {
	authMiddleware := NewAuthMiddleware(testSecret)

	handler := authMiddleware.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := domain.GetUserIDFromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if ok {
			_ = json.NewEncoder(w).Encode(map[string]string{"user_id": userID})
		} else {
			_ = json.NewEncoder(w).Encode(map[string]string{"user_id": "anonymous"})
		}
	}))

	t.Run("Extracts user_id if valid token is provided", func(t *testing.T) {
		token, err := generateTestToken(testSecret, "usr-optional-456", "opt@example.com", "authenticated", time.Now().Add(1*time.Hour))
		if err != nil {
			t.Fatalf("failed generating test token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/public-optional", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp["user_id"] != "usr-optional-456" {
			t.Errorf("expected user_id 'usr-optional-456', got '%s'", resp["user_id"])
		}
	})

	t.Run("Proceeds as anonymous without error when Authorization header is absent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public-optional", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp["user_id"] != "anonymous" {
			t.Errorf("expected user_id 'anonymous', got '%s'", resp["user_id"])
		}
	})

	t.Run("Proceeds as anonymous without error when token is invalid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/public-optional", nil)
		req.Header.Set("Authorization", "Bearer invalid-garbage-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}

		var resp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp["user_id"] != "anonymous" {
			t.Errorf("expected user_id 'anonymous', got '%s'", resp["user_id"])
		}
	})
}
