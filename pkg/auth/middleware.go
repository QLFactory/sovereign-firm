package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	ClaimsContextKey contextKey = "claims"
	UserIDContextKey contextKey = "user_id"
	TenantContextKey contextKey = "tenant_id"
)

var roleLevels = map[string]int{
	"owner":  4,
	"admin":  3,
	"member": 2,
	"viewer": 1,
}

// Middleware creates an authentication middleware
type Middleware struct {
	jwtManager *JWTManager
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware(jwtManager *JWTManager) *Middleware {
	return &Middleware{jwtManager: jwtManager}
}

// Authenticate is a middleware that validates JWT tokens
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		// Validate token
		claims, err := m.jwtManager.ValidateAccessToken(tokenStr)
		if err != nil {
			switch err {
			case ErrExpiredToken:
				http.Error(w, `{"error":"token expired"}`, http.StatusUnauthorized)
			default:
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			}
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, TenantContextKey, claims.TenantID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is a middleware that validates JWT tokens but doesn't require them
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
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

		claims, err := m.jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, TenantContextKey, claims.TenantID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireMinRole is a middleware that checks if the user has at least the specified role level.
func (m *Middleware) RequireMinRole(minRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			userLevel := roleLevels[claims.Role]
			requiredLevel := roleLevels[minRole]

			if userLevel >= requiredLevel && requiredLevel > 0 {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		})
	}
}

// GetClaims extracts claims from context
func GetClaims(ctx context.Context) *Claims {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// GetTenantID extracts tenant ID from context
func GetTenantID(ctx context.Context) string {
	tenantID, ok := ctx.Value(TenantContextKey).(string)
	if !ok {
		return ""
	}
	return tenantID
}
