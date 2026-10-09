package domain

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	// UserIDContextKey is the context key for authenticated Supabase user UUID
	UserIDContextKey contextKey = "user_id"
	// UserEmailContextKey is the context key for authenticated user email
	UserEmailContextKey contextKey = "user_email"
	// UserRoleContextKey is the context key for authenticated user role
	UserRoleContextKey contextKey = "user_role"
)

// SupabaseClaims represents standard JWT claims emitted by Supabase Auth (GoTrue)
type SupabaseClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// GetUserIDFromContext retrieves the authenticated user ID from context if available
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	val := ctx.Value(UserIDContextKey)
	if val == nil {
		return "", false
	}
	userID, ok := val.(string)
	return userID, ok && userID != ""
}

// GetUserEmailFromContext retrieves the authenticated user's email from context if available
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	val := ctx.Value(UserEmailContextKey)
	if val == nil {
		return "", false
	}
	email, ok := val.(string)
	return email, ok && email != ""
}
