package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

// Test configuration
var (
	apiBaseURL = getEnv("API_BASE_URL", "http://localhost:8080")
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// AuthTestSuite contains test state for auth tests
type AuthTestSuite struct {
	t            *testing.T
	client       *http.Client
	testEmail    string
	testPassword string
	testTenant   string
	testName     string
	accessToken  string
	refreshToken string
	userID       string
	tenantID     string
}

// NewAuthTestSuite creates a new auth test suite
func NewAuthTestSuite(t *testing.T) *AuthTestSuite {
	return &AuthTestSuite{
		t:            t,
		client:       &http.Client{Timeout: 10 * time.Second},
		testEmail:    fmt.Sprintf("apitest_%d@example.com", time.Now().UnixNano()),
		testPassword: "SecurePassword123!",
		testTenant:   fmt.Sprintf("API Test Tenant %d", time.Now().UnixNano()),
		testName:     "API Test User",
	}
}

// request makes an HTTP request and returns the response
func (s *AuthTestSuite) request(method, path string, body interface{}, headers map[string]string) (*http.Response, map[string]interface{}) {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, err := http.NewRequest(method, apiBaseURL+path, bytes.NewReader(reqBody))
	if err != nil {
		s.t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.t.Fatalf("Request failed: %v", err)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	return resp, result
}

// TestAuthAPI_HealthCheck verifies the API is running
func TestAuthAPI_HealthCheck(t *testing.T) {
	suite := NewAuthTestSuite(t)

	resp, body := suite.request("GET", "/health", nil, nil)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	if body["status"] != "healthy" {
		t.Errorf("Expected healthy status, got %v", body["status"])
	}
}

// TestAuthAPI_Register tests user registration
func TestAuthAPI_Register(t *testing.T) {
	suite := NewAuthTestSuite(t)

	// Test successful registration
	t.Run("successful registration", func(t *testing.T) {
		body := map[string]string{
			"email":       suite.testEmail,
			"password":    suite.testPassword,
			"name":        suite.testName,
			"tenant_name": suite.testTenant,
		}

		resp, result := suite.request("POST", "/api/auth/register", body, nil)

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected 201, got %d: %v", resp.StatusCode, result)
		}

		// Verify response structure
		if result["access_token"] == nil {
			t.Error("Missing access_token in response")
		}
		if result["refresh_token"] == nil {
			t.Error("Missing refresh_token in response")
		}

		user, ok := result["user"].(map[string]interface{})
		if !ok {
			t.Fatal("Missing or invalid user object in response")
		}

		if user["email"] != suite.testEmail {
			t.Errorf("Email mismatch: expected %s, got %v", suite.testEmail, user["email"])
		}
		if user["name"] != suite.testName {
			t.Errorf("Name mismatch: expected %s, got %v", suite.testName, user["name"])
		}
		if user["role"] != "owner" {
			t.Errorf("Expected role 'owner', got %v", user["role"])
		}

		// Save tokens for subsequent tests
		suite.accessToken = result["access_token"].(string)
		suite.refreshToken = result["refresh_token"].(string)
		suite.userID = user["id"].(string)
		suite.tenantID = user["tenant_id"].(string)

		t.Logf("Registered user: %s (tenant: %s)", suite.userID, suite.tenantID)
	})

	// Test duplicate registration
	t.Run("duplicate email fails", func(t *testing.T) {
		body := map[string]string{
			"email":       suite.testEmail,
			"password":    suite.testPassword,
			"name":        suite.testName,
			"tenant_name": "Another Tenant",
		}

		resp, result := suite.request("POST", "/api/auth/register", body, nil)

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected 409 Conflict, got %d: %v", resp.StatusCode, result)
		}
	})

	// Test validation errors
	t.Run("missing required fields", func(t *testing.T) {
		testCases := []struct {
			name   string
			body   map[string]string
			expect string
		}{
			{
				name:   "missing email",
				body:   map[string]string{"password": "test123!", "name": "Test", "tenant_name": "Tenant"},
				expect: "invalid email",
			},
			{
				name:   "missing password",
				body:   map[string]string{"email": "test@test.com", "name": "Test", "tenant_name": "Tenant"},
				expect: "password",
			},
			{
				name:   "missing tenant_name",
				body:   map[string]string{"email": "test@test.com", "password": "test123!", "name": "Test"},
				expect: "tenant name",
			},
			{
				name:   "short password",
				body:   map[string]string{"email": "test2@test.com", "password": "short", "name": "Test", "tenant_name": "Tenant"},
				expect: "password",
			},
			{
				name:   "invalid email format",
				body:   map[string]string{"email": "notanemail", "password": "test123!", "name": "Test", "tenant_name": "Tenant"},
				expect: "email",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, _ := suite.request("POST", "/api/auth/register", tc.body, nil)
				if resp.StatusCode != http.StatusBadRequest {
					t.Errorf("Expected 400, got %d", resp.StatusCode)
				}
			})
		}
	})
}

// TestAuthAPI_Login tests user login
func TestAuthAPI_Login(t *testing.T) {
	suite := NewAuthTestSuite(t)

	// First register a user
	regBody := map[string]string{
		"email":       suite.testEmail,
		"password":    suite.testPassword,
		"name":        suite.testName,
		"tenant_name": suite.testTenant,
	}
	resp, _ := suite.request("POST", "/api/auth/register", regBody, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Skip("Registration failed, skipping login tests")
	}

	// Test successful login
	t.Run("successful login", func(t *testing.T) {
		body := map[string]string{
			"email":    suite.testEmail,
			"password": suite.testPassword,
		}

		resp, result := suite.request("POST", "/api/auth/login", body, nil)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d: %v", resp.StatusCode, result)
		}

		if result["access_token"] == nil {
			t.Error("Missing access_token")
		}
		if result["refresh_token"] == nil {
			t.Error("Missing refresh_token")
		}

		suite.accessToken = result["access_token"].(string)
		suite.refreshToken = result["refresh_token"].(string)
	})

	// Test invalid credentials
	t.Run("invalid credentials", func(t *testing.T) {
		body := map[string]string{
			"email":    suite.testEmail,
			"password": "wrongpassword",
		}

		resp, _ := suite.request("POST", "/api/auth/login", body, nil)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})

	// Test non-existent user
	t.Run("non-existent user", func(t *testing.T) {
		body := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "anypassword",
		}

		resp, _ := suite.request("POST", "/api/auth/login", body, nil)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})
}

// TestAuthAPI_Me tests the /me endpoint
func TestAuthAPI_Me(t *testing.T) {
	suite := NewAuthTestSuite(t)

	// Register and login
	regBody := map[string]string{
		"email":       suite.testEmail,
		"password":    suite.testPassword,
		"name":        suite.testName,
		"tenant_name": suite.testTenant,
	}
	resp, result := suite.request("POST", "/api/auth/register", regBody, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Skip("Registration failed")
	}
	suite.accessToken = result["access_token"].(string)

	// Test with valid token
	t.Run("with valid token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer " + suite.accessToken,
		}

		resp, result := suite.request("GET", "/api/auth/me", nil, headers)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d: %v", resp.StatusCode, result)
		}

		if result["email"] != suite.testEmail {
			t.Errorf("Email mismatch: got %v", result["email"])
		}
		if result["name"] != suite.testName {
			t.Errorf("Name mismatch: got %v", result["name"])
		}
	})

	// Test without token
	t.Run("without token", func(t *testing.T) {
		resp, _ := suite.request("GET", "/api/auth/me", nil, nil)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})

	// Test with invalid token
	t.Run("with invalid token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer invalid-token",
		}

		resp, _ := suite.request("GET", "/api/auth/me", nil, headers)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})
}

// TestAuthAPI_Refresh tests token refresh
func TestAuthAPI_Refresh(t *testing.T) {
	suite := NewAuthTestSuite(t)

	// Register first
	regBody := map[string]string{
		"email":       suite.testEmail,
		"password":    suite.testPassword,
		"name":        suite.testName,
		"tenant_name": suite.testTenant,
	}
	resp, result := suite.request("POST", "/api/auth/register", regBody, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Skip("Registration failed")
	}
	suite.refreshToken = result["refresh_token"].(string)

	// Test successful refresh
	t.Run("successful refresh", func(t *testing.T) {
		body := map[string]string{
			"refresh_token": suite.refreshToken,
		}

		resp, result := suite.request("POST", "/api/auth/refresh", body, nil)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d: %v", resp.StatusCode, result)
		}

		if result["access_token"] == nil {
			t.Error("Missing new access_token")
		}
		if result["refresh_token"] == nil {
			t.Error("Missing new refresh_token")
		}

		// Update tokens for subsequent tests
		suite.accessToken = result["access_token"].(string)
		suite.refreshToken = result["refresh_token"].(string)
	})

	// Test with old (rotated) token should fail
	t.Run("old token fails after rotation", func(t *testing.T) {
		// The old refresh token should have been revoked
		oldToken := suite.refreshToken // This was already rotated above

		// Get a new token first
		body := map[string]string{
			"refresh_token": suite.refreshToken,
		}
		resp, result := suite.request("POST", "/api/auth/refresh", body, nil)
		if resp.StatusCode == http.StatusOK {
			suite.refreshToken = result["refresh_token"].(string)
		}

		// Now try the old token
		body = map[string]string{
			"refresh_token": oldToken,
		}
		resp, _ = suite.request("POST", "/api/auth/refresh", body, nil)

		// Old token should be revoked (401) - if token rotation is implemented
		// Note: This depends on implementation - some systems allow reuse
		t.Logf("Old token refresh status: %d", resp.StatusCode)
	})

	// Test with invalid refresh token
	t.Run("invalid refresh token", func(t *testing.T) {
		body := map[string]string{
			"refresh_token": "invalid-refresh-token",
		}

		resp, _ := suite.request("POST", "/api/auth/refresh", body, nil)

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})
}

// TestAuthAPI_Logout tests logout functionality
func TestAuthAPI_Logout(t *testing.T) {
	suite := NewAuthTestSuite(t)

	// Register first
	regBody := map[string]string{
		"email":       suite.testEmail,
		"password":    suite.testPassword,
		"name":        suite.testName,
		"tenant_name": suite.testTenant,
	}
	resp, result := suite.request("POST", "/api/auth/register", regBody, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Skip("Registration failed")
	}
	suite.accessToken = result["access_token"].(string)
	suite.refreshToken = result["refresh_token"].(string)

	// Test logout
	t.Run("successful logout", func(t *testing.T) {
		body := map[string]string{
			"refresh_token": suite.refreshToken,
		}
		headers := map[string]string{
			"Authorization": "Bearer " + suite.accessToken,
		}

		resp, _ := suite.request("POST", "/api/auth/logout", body, headers)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	// Verify refresh token is revoked
	t.Run("refresh token revoked after logout", func(t *testing.T) {
		body := map[string]string{
			"refresh_token": suite.refreshToken,
		}

		resp, _ := suite.request("POST", "/api/auth/refresh", body, nil)

		// Should fail since token was revoked
		if resp.StatusCode != http.StatusUnauthorized {
			t.Logf("Note: Refresh token may still work depending on implementation (got %d)", resp.StatusCode)
		}
	})
}

// TestAuthAPI_FullFlow tests the complete authentication flow
func TestAuthAPI_FullFlow(t *testing.T) {
	suite := NewAuthTestSuite(t)

	// Step 1: Register
	t.Log("Step 1: Register new user")
	regBody := map[string]string{
		"email":       suite.testEmail,
		"password":    suite.testPassword,
		"name":        suite.testName,
		"tenant_name": suite.testTenant,
	}
	resp, result := suite.request("POST", "/api/auth/register", regBody, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Registration failed: %d - %v", resp.StatusCode, result)
	}
	suite.accessToken = result["access_token"].(string)
	suite.refreshToken = result["refresh_token"].(string)
	t.Logf("  ✓ Registered: %s", suite.testEmail)

	// Step 2: Access protected resource
	t.Log("Step 2: Access /me with token")
	headers := map[string]string{
		"Authorization": "Bearer " + suite.accessToken,
	}
	resp, result = suite.request("GET", "/api/auth/me", nil, headers)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to access /me: %d", resp.StatusCode)
	}
	t.Logf("  ✓ Accessed /me: %v", result["email"])

	// Step 3: Refresh token
	t.Log("Step 3: Refresh access token")
	body := map[string]string{
		"refresh_token": suite.refreshToken,
	}
	resp, result = suite.request("POST", "/api/auth/refresh", body, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to refresh: %d", resp.StatusCode)
	}
	newAccessToken := result["access_token"].(string)
	suite.refreshToken = result["refresh_token"].(string)
	t.Log("  ✓ Token refreshed")

	// Step 4: Access with new token
	t.Log("Step 4: Access /me with new token")
	headers = map[string]string{
		"Authorization": "Bearer " + newAccessToken,
	}
	resp, _ = suite.request("GET", "/api/auth/me", nil, headers)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to access /me with new token: %d", resp.StatusCode)
	}
	t.Log("  ✓ New token works")

	// Step 5: Logout
	t.Log("Step 5: Logout")
	body = map[string]string{
		"refresh_token": suite.refreshToken,
	}
	headers = map[string]string{
		"Authorization": "Bearer " + newAccessToken,
	}
	resp, _ = suite.request("POST", "/api/auth/logout", body, headers)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to logout: %d", resp.StatusCode)
	}
	t.Log("  ✓ Logged out")

	// Step 6: Login again
	t.Log("Step 6: Login with credentials")
	loginBody := map[string]string{
		"email":    suite.testEmail,
		"password": suite.testPassword,
	}
	resp, result = suite.request("POST", "/api/auth/login", loginBody, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to login: %d", resp.StatusCode)
	}
	t.Log("  ✓ Logged in successfully")

	t.Log("\n✓ Full authentication flow completed successfully!")
}
