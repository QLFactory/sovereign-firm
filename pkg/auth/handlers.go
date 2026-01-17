package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler handles authentication endpoints
type Handler struct {
	db         *pgxpool.Pool
	jwtManager *JWTManager
}

// NewHandler creates a new auth handler
func NewHandler(db interface{}, jwtManager *JWTManager) *Handler {
	var pool *pgxpool.Pool
	if db != nil {
		pool = db.(*pgxpool.Pool)
	}
	return &Handler{
		db:         pool,
		jwtManager: jwtManager,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	TenantName  string `json:"tenant_name"`
	TenantSlug  string `json:"tenant_slug,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	TokenType    string       `json:"token_type"`
}

// UserResponse represents user info in responses
type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if !emailRegex.MatchString(req.Email) {
		jsonError(w, "invalid email format", http.StatusBadRequest)
		return
	}

	// Validate password
	if err := ValidatePassword(req.Password); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate tenant name
	req.TenantName = strings.TrimSpace(req.TenantName)
	if req.TenantName == "" {
		jsonError(w, "tenant name is required", http.StatusBadRequest)
		return
	}

	// Generate slug if not provided
	if req.TenantSlug == "" {
		req.TenantSlug = slugify(req.TenantName)
	}

	// Hash password
	passwordHash, err := HashPassword(req.Password, nil)
	if err != nil {
		jsonError(w, "failed to process password", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	// Start transaction
	tx, err := h.db.Begin(ctx)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	// Check if email already exists
	var existingID string
	err = tx.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingID)
	if err == nil {
		jsonError(w, "email already registered", http.StatusConflict)
		return
	}
	if err != pgx.ErrNoRows {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}

	// Create tenant
	var tenantID string
	err = tx.QueryRow(ctx,
		`INSERT INTO tenants (name, slug) VALUES ($1, $2) RETURNING id`,
		req.TenantName, req.TenantSlug,
	).Scan(&tenantID)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			jsonError(w, "tenant slug already exists", http.StatusConflict)
			return
		}
		jsonError(w, "failed to create tenant", http.StatusInternalServerError)
		return
	}

	// Create user as owner
	var userID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users (tenant_id, email, password_hash, name, role)
		 VALUES ($1, $2, $3, $4, 'owner') RETURNING id`,
		tenantID, req.Email, passwordHash, req.Name,
	).Scan(&userID)
	if err != nil {
		jsonError(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate tokens
	tokenPair, refreshToken, err := h.jwtManager.GenerateTokenPair(userID, tenantID, req.Email, "owner")
	if err != nil {
		jsonError(w, "failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Store refresh token hash
	tokenHash := hashToken(refreshToken)
	_, err = tx.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, time.Now().Add(h.jwtManager.RefreshTokenTTL()),
	)
	if err != nil {
		jsonError(w, "failed to store refresh token", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		jsonError(w, "failed to complete registration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		User: UserResponse{
			ID:       userID,
			Email:    req.Email,
			Name:     req.Name,
			Role:     "owner",
			TenantID: tenantID,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
	})
}

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	ctx := r.Context()

	// Get user by email
	var userID, tenantID, passwordHash, name, role string
	err := h.db.QueryRow(ctx,
		`SELECT id, tenant_id, password_hash, COALESCE(name, ''), role FROM users WHERE email = $1`,
		req.Email,
	).Scan(&userID, &tenantID, &passwordHash, &name, &role)
	if err != nil {
		if err == pgx.ErrNoRows {
			jsonError(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}

	// Verify password
	valid, err := VerifyPassword(req.Password, passwordHash)
	if err != nil || !valid {
		jsonError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate tokens
	tokenPair, refreshToken, err := h.jwtManager.GenerateTokenPair(userID, tenantID, req.Email, role)
	if err != nil {
		jsonError(w, "failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Store refresh token hash
	tokenHash := hashToken(refreshToken)
	_, err = h.db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, time.Now().Add(h.jwtManager.RefreshTokenTTL()),
	)
	if err != nil {
		jsonError(w, "failed to store refresh token", http.StatusInternalServerError)
		return
	}

	// Update last login
	h.db.Exec(ctx, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		User: UserResponse{
			ID:       userID,
			Email:    req.Email,
			Name:     name,
			Role:     role,
			TenantID: tenantID,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
	})
}

// Refresh handles token refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tokenHash := hashToken(req.RefreshToken)

	// Get and validate refresh token
	var userID string
	err := h.db.QueryRow(ctx,
		`SELECT user_id FROM refresh_tokens
		 WHERE token_hash = $1 AND NOT revoked AND expires_at > NOW()`,
		tokenHash,
	).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			jsonError(w, "invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}

	// Revoke old refresh token (rotation)
	h.db.Exec(ctx, `UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`, tokenHash)

	// Get user info
	var tenantID, email, name, role string
	err = h.db.QueryRow(ctx,
		`SELECT tenant_id, email, COALESCE(name, ''), role FROM users WHERE id = $1`,
		userID,
	).Scan(&tenantID, &email, &name, &role)
	if err != nil {
		jsonError(w, "user not found", http.StatusUnauthorized)
		return
	}

	// Generate new tokens
	tokenPair, newRefreshToken, err := h.jwtManager.GenerateTokenPair(userID, tenantID, email, role)
	if err != nil {
		jsonError(w, "failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Store new refresh token
	newTokenHash := hashToken(newRefreshToken)
	_, err = h.db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, newTokenHash, time.Now().Add(h.jwtManager.RefreshTokenTTL()),
	)
	if err != nil {
		jsonError(w, "failed to store refresh token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		User: UserResponse{
			ID:       userID,
			Email:    email,
			Name:     name,
			Role:     role,
			TenantID: tenantID,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
	})
}

// Logout handles user logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body, try to revoke all tokens for the user
		claims := GetClaims(r.Context())
		if claims != nil {
			h.db.Exec(r.Context(),
				`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`,
				claims.UserID,
			)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"logged out"}`))
		return
	}

	// Revoke specific refresh token
	tokenHash := hashToken(req.RefreshToken)
	h.db.Exec(r.Context(),
		`UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`,
		tokenHash,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"logged out"}`))
}

// Me returns current user info
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	var name string
	err := h.db.QueryRow(ctx,
		`SELECT COALESCE(name, '') FROM users WHERE id = $1`,
		claims.UserID,
	).Scan(&name)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserResponse{
		ID:       claims.UserID,
		Email:    claims.Email,
		Name:     name,
		Role:     claims.Role,
		TenantID: claims.TenantID,
	})
}

// helper functions

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	s = reg.ReplaceAllString(s, "")
	return s
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CleanupExpiredTokens removes expired refresh tokens
func (h *Handler) CleanupExpiredTokens(ctx context.Context) error {
	_, err := h.db.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = TRUE`,
	)
	return err
}
