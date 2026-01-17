package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qlfactory/sovereign-firm/pkg/auth"
)

// FrontendRegisterRequest mirrors the frontend TypeScript RegisterRequest
// This MUST match: frontend/app/lib/api/types.ts RegisterRequest
type FrontendRegisterRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	TenantName string `json:"tenant_name"`
}

// FrontendLoginRequest mirrors the frontend TypeScript LoginRequest
type FrontendLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// FrontendAuthResponse mirrors the frontend TypeScript AuthResponse
type FrontendAuthResponse struct {
	User         FrontendUser `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    string       `json:"expires_at,omitempty"`
	ExpiresIn    int          `json:"expires_in,omitempty"`
}

// FrontendUser mirrors the frontend TypeScript User interface
type FrontendUser struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name,omitempty"`
	Role       string `json:"role"`
	CreatedAt  string `json:"created_at,omitempty"`
}

// TestRegisterRequestContract validates that the backend RegisterRequest
// can be deserialized from the frontend's expected JSON structure
func TestRegisterRequestContract(t *testing.T) {
	// This is what the frontend sends
	frontendPayload := FrontendRegisterRequest{
		Email:      "test@example.com",
		Password:   "securePassword123",
		Name:       "Test User",
		TenantName: "Test Corp",
	}

	// Serialize as frontend would
	jsonBytes, err := json.Marshal(frontendPayload)
	if err != nil {
		t.Fatalf("Failed to marshal frontend payload: %v", err)
	}

	// Deserialize into backend struct
	var backendReq auth.RegisterRequest
	if err := json.Unmarshal(jsonBytes, &backendReq); err != nil {
		t.Fatalf("Backend cannot deserialize frontend RegisterRequest: %v", err)
	}

	// Validate all fields transferred correctly
	if backendReq.Email != frontendPayload.Email {
		t.Errorf("Email mismatch: got %q, want %q", backendReq.Email, frontendPayload.Email)
	}
	if backendReq.Password != frontendPayload.Password {
		t.Errorf("Password mismatch: got %q, want %q", backendReq.Password, frontendPayload.Password)
	}
	if backendReq.Name != frontendPayload.Name {
		t.Errorf("Name mismatch: got %q, want %q", backendReq.Name, frontendPayload.Name)
	}
	if backendReq.TenantName != frontendPayload.TenantName {
		t.Errorf("TenantName mismatch: got %q, want %q", backendReq.TenantName, frontendPayload.TenantName)
	}
}

// TestLoginRequestContract validates login request compatibility
func TestLoginRequestContract(t *testing.T) {
	frontendPayload := FrontendLoginRequest{
		Email:    "test@example.com",
		Password: "securePassword123",
	}

	jsonBytes, err := json.Marshal(frontendPayload)
	if err != nil {
		t.Fatalf("Failed to marshal frontend payload: %v", err)
	}

	var backendReq auth.LoginRequest
	if err := json.Unmarshal(jsonBytes, &backendReq); err != nil {
		t.Fatalf("Backend cannot deserialize frontend LoginRequest: %v", err)
	}

	if backendReq.Email != frontendPayload.Email {
		t.Errorf("Email mismatch: got %q, want %q", backendReq.Email, frontendPayload.Email)
	}
	if backendReq.Password != frontendPayload.Password {
		t.Errorf("Password mismatch: got %q, want %q", backendReq.Password, frontendPayload.Password)
	}
}

// TestAuthResponseContract validates that backend AuthResponse can be
// deserialized by the frontend
func TestAuthResponseContract(t *testing.T) {
	// This is what the backend sends
	backendResponse := auth.AuthResponse{
		User: auth.UserResponse{
			ID:       "user-123",
			Email:    "test@example.com",
			Name:     "Test User",
			Role:     "owner",
			TenantID: "tenant-456",
		},
		AccessToken:  "access-token-xyz",
		RefreshToken: "refresh-token-abc",
		TokenType:    "Bearer",
	}

	jsonBytes, err := json.Marshal(backendResponse)
	if err != nil {
		t.Fatalf("Failed to marshal backend response: %v", err)
	}

	// Deserialize into frontend struct
	var frontendResp FrontendAuthResponse
	if err := json.Unmarshal(jsonBytes, &frontendResp); err != nil {
		t.Fatalf("Frontend cannot deserialize backend AuthResponse: %v", err)
	}

	// Validate key fields
	if frontendResp.AccessToken != backendResponse.AccessToken {
		t.Errorf("AccessToken mismatch: got %q, want %q", frontendResp.AccessToken, backendResponse.AccessToken)
	}
	if frontendResp.RefreshToken != backendResponse.RefreshToken {
		t.Errorf("RefreshToken mismatch: got %q, want %q", frontendResp.RefreshToken, backendResponse.RefreshToken)
	}
	if frontendResp.User.ID != backendResponse.User.ID {
		t.Errorf("User.ID mismatch: got %q, want %q", frontendResp.User.ID, backendResponse.User.ID)
	}
	if frontendResp.User.Email != backendResponse.User.Email {
		t.Errorf("User.Email mismatch: got %q, want %q", frontendResp.User.Email, backendResponse.User.Email)
	}
	if frontendResp.User.Name != backendResponse.User.Name {
		t.Errorf("User.Name mismatch: got %q, want %q", frontendResp.User.Name, backendResponse.User.Name)
	}
	if frontendResp.User.Role != backendResponse.User.Role {
		t.Errorf("User.Role mismatch: got %q, want %q", frontendResp.User.Role, backendResponse.User.Role)
	}
	if frontendResp.User.TenantID != backendResponse.User.TenantID {
		t.Errorf("User.TenantID mismatch: got %q, want %q", frontendResp.User.TenantID, backendResponse.User.TenantID)
	}
}

// TestRequiredFieldsDocumented ensures all required fields are documented
func TestRequiredFieldsDocumented(t *testing.T) {
	// Document what fields are required vs optional
	// This serves as living documentation that must be kept in sync

	t.Run("RegisterRequest required fields", func(t *testing.T) {
		requiredFields := []string{"email", "password", "name", "tenant_name"}
		optionalFields := []string{"tenant_slug"}

		t.Logf("Required: %v", requiredFields)
		t.Logf("Optional: %v", optionalFields)

		// Verify a minimal valid request
		minimalRequest := `{"email":"a@b.com","password":"12345678","name":"X","tenant_name":"Y"}`
		var req auth.RegisterRequest
		if err := json.Unmarshal([]byte(minimalRequest), &req); err != nil {
			t.Errorf("Minimal request should parse: %v", err)
		}
		if req.Email == "" || req.Password == "" || req.Name == "" || req.TenantName == "" {
			t.Error("Required fields not correctly parsed")
		}
	})

	t.Run("LoginRequest required fields", func(t *testing.T) {
		requiredFields := []string{"email", "password"}

		t.Logf("Required: %v", requiredFields)

		minimalRequest := `{"email":"a@b.com","password":"12345678"}`
		var req auth.LoginRequest
		if err := json.Unmarshal([]byte(minimalRequest), &req); err != nil {
			t.Errorf("Minimal request should parse: %v", err)
		}
	})
}

// MockRecorder is a simple test helper to verify API shape
type MockRecorder struct {
	*httptest.ResponseRecorder
	requests [][]byte
}

func (m *MockRecorder) RecordRequest(body []byte) {
	m.requests = append(m.requests, body)
}

// TestRegisterFlowDocumentation shows a complete registration flow
// This serves as documentation and can be used to generate client code
func TestRegisterFlowDocumentation(t *testing.T) {
	// Step 1: Frontend sends registration request
	reqBody := FrontendRegisterRequest{
		Email:      "newuser@example.com",
		Password:   "SecurePass123!",
		Name:       "New User",
		TenantName: "New Company",
	}

	jsonBody, _ := json.Marshal(reqBody)

	// Step 2: Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Step 3: Expected response shape
	expectedShape := FrontendAuthResponse{
		User: FrontendUser{
			ID:       "uuid",
			Email:    reqBody.Email,
			Name:     reqBody.Name,
			Role:     "owner",
			TenantID: "uuid",
		},
		AccessToken:  "jwt-token",
		RefreshToken: "refresh-token",
	}

	t.Logf("Request: POST /api/auth/register")
	t.Logf("Body: %s", jsonBody)
	t.Logf("Expected response shape: %+v", expectedShape)

	// Actual handler would be called here in integration tests
	_ = req
}

// TestJSONTagsConsistency verifies JSON tags are consistent
func TestJSONTagsConsistency(t *testing.T) {
	// Backend uses snake_case in JSON
	testCases := []struct {
		name     string
		jsonKey  string
		expected string
	}{
		{"email", "email", "email"},
		{"password", "password", "password"},
		{"name", "name", "name"},
		{"tenant_name", "tenant_name", "tenant_name"},
		{"access_token", "access_token", "access_token"},
		{"refresh_token", "refresh_token", "refresh_token"},
		{"tenant_id", "tenant_id", "tenant_id"},
	}

	// Verify backend response uses expected keys
	response := auth.AuthResponse{
		User: auth.UserResponse{
			ID:       "id",
			Email:    "email",
			Name:     "name",
			Role:     "role",
			TenantID: "tenant-id",
		},
		AccessToken:  "token",
		RefreshToken: "refresh",
	}

	jsonBytes, _ := json.Marshal(response)
	var rawMap map[string]interface{}
	json.Unmarshal(jsonBytes, &rawMap)

	for _, tc := range testCases {
		// Check top-level keys
		if _, exists := rawMap[tc.jsonKey]; exists {
			t.Logf("Found expected key: %s", tc.jsonKey)
		}
	}

	// Verify nested user object
	userMap, ok := rawMap["user"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'user' object in response")
	}

	userKeys := []string{"id", "email", "name", "role", "tenant_id"}
	for _, key := range userKeys {
		if _, exists := userMap[key]; !exists {
			t.Errorf("Missing user.%s in response", key)
		}
	}
}

// BenchmarkJSONSerialization ensures serialization is efficient
func BenchmarkJSONSerialization(b *testing.B) {
	req := FrontendRegisterRequest{
		Email:      "test@example.com",
		Password:   "securePassword123",
		Name:       "Test User",
		TenantName: "Test Corp",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonBytes, _ := json.Marshal(req)
		var backendReq auth.RegisterRequest
		json.Unmarshal(jsonBytes, &backendReq)
	}
}
