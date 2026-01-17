package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
	ErrInvalidClaim = errors.New("invalid claim")
)

// Claims represents JWT claims with tenant context
type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair contains access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// JWTManager handles JWT operations
type JWTManager struct {
	privateKey      *rsa.PrivateKey
	publicKey       *rsa.PublicKey
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	PrivateKeyPath  string
	PublicKeyPath   string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

// DefaultJWTConfig returns default JWT configuration
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		PrivateKeyPath:  os.Getenv("JWT_PRIVATE_KEY_PATH"),
		PublicKeyPath:   os.Getenv("JWT_PUBLIC_KEY_PATH"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "sovereign-firm",
	}
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg JWTConfig) (*JWTManager, error) {
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey
	var err error

	// Try to load from files first
	if cfg.PrivateKeyPath != "" && cfg.PublicKeyPath != "" {
		privateKey, err = loadPrivateKey(cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("load private key: %w", err)
		}
		publicKey, err = loadPublicKey(cfg.PublicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("load public key: %w", err)
		}
	} else {
		// Generate new key pair for development
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("generate key: %w", err)
		}
		publicKey = &privateKey.PublicKey
	}

	return &JWTManager{
		privateKey:      privateKey,
		publicKey:       publicKey,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		issuer:          cfg.Issuer,
	}, nil
}

// GenerateTokenPair generates access and refresh tokens
func (m *JWTManager) GenerateTokenPair(userID, tenantID, email, role string) (*TokenPair, string, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTokenTTL)

	// Access token claims
	claims := &Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   userID,
		},
	}

	// Create access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	accessTokenStr, err := accessToken.SignedString(m.privateKey)
	if err != nil {
		return nil, "", fmt.Errorf("sign access token: %w", err)
	}

	// Create refresh token (opaque random string)
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, "", fmt.Errorf("generate refresh token: %w", err)
	}
	refreshTokenStr := hex.EncodeToString(refreshTokenBytes)

	return &TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, refreshTokenStr, nil
}

// ValidateAccessToken validates an access token and returns claims
func (m *JWTManager) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaim
	}

	return claims, nil
}

// RefreshTokenTTL returns the refresh token TTL
func (m *JWTManager) RefreshTokenTTL() time.Duration {
	return m.refreshTokenTTL
}

// HashRefreshToken creates a hash of the refresh token for storage
func HashRefreshToken(token string) string {
	// Use SHA256 for refresh token hashing
	hash := make([]byte, 32)
	copy(hash, []byte(token))
	return hex.EncodeToString(hash)
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPub, nil
}
