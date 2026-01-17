package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// ProjectsTestSuite contains test state for projects API tests
type ProjectsTestSuite struct {
	t            *testing.T
	client       *http.Client
	accessToken  string
	refreshToken string
	tenantID     string
	userID       string
	projectID    string
	workflowID   string
}

// NewProjectsTestSuite creates a new projects test suite
func NewProjectsTestSuite(t *testing.T) *ProjectsTestSuite {
	return &ProjectsTestSuite{
		t:      t,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// setup registers a user and gets auth tokens
func (s *ProjectsTestSuite) setup() {
	testEmail := fmt.Sprintf("projecttest_%d@example.com", time.Now().UnixNano())
	testPassword := "SecurePassword123!"
	testTenant := fmt.Sprintf("Project Test Tenant %d", time.Now().UnixNano())

	// Register user
	regBody := map[string]string{
		"email":       testEmail,
		"password":    testPassword,
		"name":        "Project Test User",
		"tenant_name": testTenant,
	}

	body, _ := json.Marshal(regBody)
	resp, err := s.client.Post(apiBaseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		s.t.Fatalf("Setup failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		s.t.Fatalf("Setup: registration failed with status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	s.accessToken = result["access_token"].(string)
	s.refreshToken = result["refresh_token"].(string)

	user := result["user"].(map[string]interface{})
	s.userID = user["id"].(string)
	s.tenantID = user["tenant_id"].(string)

	s.t.Logf("Setup complete: user=%s, tenant=%s", s.userID, s.tenantID)
}

// request makes an authenticated HTTP request
func (s *ProjectsTestSuite) request(method, path string, body interface{}) (*http.Response, map[string]interface{}) {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, err := http.NewRequest(method, apiBaseURL+path, bytes.NewReader(reqBody))
	if err != nil {
		s.t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.accessToken)
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

// requestNoAuth makes an unauthenticated request
func (s *ProjectsTestSuite) requestNoAuth(method, path string, body interface{}) (*http.Response, map[string]interface{}) {
	oldToken := s.accessToken
	s.accessToken = ""
	defer func() { s.accessToken = oldToken }()
	return s.request(method, path, body)
}

// TestProjectsAPI_RequiresAuth verifies that project endpoints require authentication
func TestProjectsAPI_RequiresAuth(t *testing.T) {
	suite := NewProjectsTestSuite(t)
	// Don't setup - test without auth

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/projects"},
		{"POST", "/api/projects"},
		{"GET", "/api/projects/some-id"},
		{"DELETE", "/api/projects/some-id"},
		{"GET", "/api/projects/some-id/state"},
		{"POST", "/api/projects/some-id/message"},
	}

	for _, ep := range endpoints {
		t.Run(fmt.Sprintf("%s %s", ep.method, ep.path), func(t *testing.T) {
			resp, _ := suite.requestNoAuth(ep.method, ep.path, nil)
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected 401, got %d", resp.StatusCode)
			}
		})
	}
}

// TestProjectsAPI_ListProjects tests the list projects endpoint
func TestProjectsAPI_ListProjects(t *testing.T) {
	suite := NewProjectsTestSuite(t)
	suite.setup()

	// Test empty list
	t.Run("empty list", func(t *testing.T) {
		resp, result := suite.request("GET", "/api/projects", nil)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d: %v", resp.StatusCode, result)
		}

		projects, ok := result["projects"].([]interface{})
		if !ok {
			t.Fatal("Missing projects array in response")
		}

		// Should be empty for new user
		t.Logf("Found %d projects", len(projects))
	})

	// Test pagination params
	t.Run("with pagination", func(t *testing.T) {
		resp, result := suite.request("GET", "/api/projects?limit=5&offset=0", nil)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		if result["limit"] == nil || result["offset"] == nil || result["total"] == nil {
			t.Error("Missing pagination fields in response")
		}

		t.Logf("Pagination: limit=%v, offset=%v, total=%v",
			result["limit"], result["offset"], result["total"])
	})
}

// TestProjectsAPI_CreateProject tests project creation
func TestProjectsAPI_CreateProject(t *testing.T) {
	suite := NewProjectsTestSuite(t)
	suite.setup()

	// Test successful creation
	t.Run("successful creation", func(t *testing.T) {
		body := map[string]interface{}{
			"name":               "Test Project",
			"description":        "A test project created by API tests",
			"initial_message":    "Build a simple todo app",
			"enable_full_stack":  true,
			"preferred_frontend": "React",
			"preferred_backend":  "Go",
			"preferred_database": "PostgreSQL",
		}

		resp, result := suite.request("POST", "/api/projects", body)

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected 201, got %d: %v", resp.StatusCode, result)
			return
		}

		// Verify response fields
		requiredFields := []string{"id", "workflow_id", "name", "phase", "status", "created_at"}
		for _, field := range requiredFields {
			if result[field] == nil {
				t.Errorf("Missing field: %s", field)
			}
		}

		if result["name"] != "Test Project" {
			t.Errorf("Name mismatch: got %v", result["name"])
		}
		if result["phase"] != "INTAKE" {
			t.Errorf("Expected initial phase INTAKE, got %v", result["phase"])
		}
		if result["status"] != "active" {
			t.Errorf("Expected status active, got %v", result["status"])
		}

		suite.projectID = result["id"].(string)
		suite.workflowID = result["workflow_id"].(string)

		t.Logf("Created project: id=%s, workflow=%s", suite.projectID, suite.workflowID)
	})

	// Test validation
	t.Run("missing name fails", func(t *testing.T) {
		body := map[string]interface{}{
			"description": "No name provided",
		}

		resp, _ := suite.request("POST", "/api/projects", body)

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", resp.StatusCode)
		}
	})

	// Test empty body
	t.Run("empty body fails", func(t *testing.T) {
		resp, _ := suite.request("POST", "/api/projects", map[string]interface{}{})

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", resp.StatusCode)
		}
	})
}

// TestProjectsAPI_GetProject tests getting a single project
func TestProjectsAPI_GetProject(t *testing.T) {
	suite := NewProjectsTestSuite(t)
	suite.setup()

	// Create a project first
	createBody := map[string]interface{}{
		"name":        "Get Test Project",
		"description": "Project for get test",
	}
	resp, result := suite.request("POST", "/api/projects", createBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create project: %d", resp.StatusCode)
	}
	suite.projectID = result["id"].(string)
	suite.workflowID = result["workflow_id"].(string)

	// Test get by ID
	t.Run("get by id", func(t *testing.T) {
		resp, result := suite.request("GET", "/api/projects/"+suite.projectID, nil)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d: %v", resp.StatusCode, result)
			return
		}

		if result["id"] != suite.projectID {
			t.Errorf("ID mismatch: expected %s, got %v", suite.projectID, result["id"])
		}
		if result["workflow_id"] != suite.workflowID {
			t.Errorf("Workflow ID mismatch: expected %s, got %v", suite.workflowID, result["workflow_id"])
		}
	})

	// Test get non-existent project
	t.Run("not found", func(t *testing.T) {
		resp, _ := suite.request("GET", "/api/projects/non-existent-id", nil)

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", resp.StatusCode)
		}
	})
}

// TestProjectsAPI_SendMessage tests sending messages to workflows
func TestProjectsAPI_SendMessage(t *testing.T) {
	suite := NewProjectsTestSuite(t)
	suite.setup()

	// Create a project first
	createBody := map[string]interface{}{
		"name":        "Message Test Project",
		"description": "Project for message test",
	}
	resp, result := suite.request("POST", "/api/projects", createBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create project: %d", resp.StatusCode)
	}
	suite.projectID = result["id"].(string)

	// Wait a bit for workflow to start
	time.Sleep(2 * time.Second)

	// Test sending a message
	t.Run("send message", func(t *testing.T) {
		body := map[string]string{
			"message": "This is a test message",
		}

		resp, result := suite.request("POST", "/api/projects/"+suite.projectID+"/message", body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d: %v", resp.StatusCode, result)
		}

		if result["status"] != "sent" {
			t.Errorf("Expected status 'sent', got %v", result["status"])
		}
	})

	// Test sending approval
	t.Run("send approval", func(t *testing.T) {
		body := map[string]string{
			"message": "/approve",
		}

		resp, result := suite.request("POST", "/api/projects/"+suite.projectID+"/message", body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d: %v", resp.StatusCode, result)
		}
	})

	// Test message to non-existent project
	t.Run("message to non-existent project", func(t *testing.T) {
		body := map[string]string{
			"message": "test",
		}

		resp, _ := suite.request("POST", "/api/projects/non-existent-id/message", body)

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", resp.StatusCode)
		}
	})
}

// TestProjectsAPI_ArchiveProject tests project archival
func TestProjectsAPI_ArchiveProject(t *testing.T) {
	suite := NewProjectsTestSuite(t)
	suite.setup()

	// Create a project first
	createBody := map[string]interface{}{
		"name":        "Archive Test Project",
		"description": "Project for archive test",
	}
	resp, result := suite.request("POST", "/api/projects", createBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create project: %d", resp.StatusCode)
	}
	suite.projectID = result["id"].(string)

	// Test archive
	t.Run("archive project", func(t *testing.T) {
		resp, _ := suite.request("DELETE", "/api/projects/"+suite.projectID, nil)

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected 204, got %d", resp.StatusCode)
		}
	})

	// Verify archived project doesn't appear in list
	t.Run("archived not in list", func(t *testing.T) {
		resp, result := suite.request("GET", "/api/projects", nil)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
			return
		}

		projects := result["projects"].([]interface{})
		for _, p := range projects {
			proj := p.(map[string]interface{})
			if proj["id"] == suite.projectID {
				t.Error("Archived project should not appear in list")
			}
		}
	})

	// Test archive non-existent project (using valid UUID format)
	t.Run("archive non-existent", func(t *testing.T) {
		// Use a valid UUID format that doesn't exist
		resp, _ := suite.request("DELETE", "/api/projects/00000000-0000-0000-0000-000000000000", nil)

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", resp.StatusCode)
		}
	})
}

// TestProjectsAPI_TenantIsolation verifies projects are isolated between tenants
func TestProjectsAPI_TenantIsolation(t *testing.T) {
	// Create two users in different tenants
	suite1 := NewProjectsTestSuite(t)
	suite1.setup()

	suite2 := NewProjectsTestSuite(t)
	suite2.setup()

	// User 1 creates a project
	createBody := map[string]interface{}{
		"name":        "Tenant 1 Project",
		"description": "This belongs to tenant 1",
	}
	resp, result := suite1.request("POST", "/api/projects", createBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create project: %d", resp.StatusCode)
	}
	tenant1ProjectID := result["id"].(string)

	// User 2 should NOT see User 1's project
	t.Run("tenant isolation in list", func(t *testing.T) {
		resp, result := suite2.request("GET", "/api/projects", nil)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
			return
		}

		projects := result["projects"].([]interface{})
		for _, p := range projects {
			proj := p.(map[string]interface{})
			if proj["id"] == tenant1ProjectID {
				t.Error("User 2 should not see User 1's project")
			}
		}
	})

	// User 2 should NOT be able to access User 1's project directly
	t.Run("tenant isolation in get", func(t *testing.T) {
		resp, _ := suite2.request("GET", "/api/projects/"+tenant1ProjectID, nil)

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 (not found due to tenant isolation), got %d", resp.StatusCode)
		}
	})

	// User 2 should NOT be able to delete User 1's project
	t.Run("tenant isolation in delete", func(t *testing.T) {
		resp, _ := suite2.request("DELETE", "/api/projects/"+tenant1ProjectID, nil)

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected 404 (not found due to tenant isolation), got %d", resp.StatusCode)
		}
	})
}

// TestProjectsAPI_FullLifecycle tests the complete project lifecycle
func TestProjectsAPI_FullLifecycle(t *testing.T) {
	suite := NewProjectsTestSuite(t)
	suite.setup()

	t.Log("Step 1: Create project")
	createBody := map[string]interface{}{
		"name":               "Lifecycle Test Project",
		"description":        "Testing full project lifecycle",
		"initial_message":    "Build a simple REST API",
		"enable_full_stack":  true,
		"preferred_frontend": "React",
		"preferred_backend":  "Go",
		"preferred_database": "PostgreSQL",
	}
	resp, result := suite.request("POST", "/api/projects", createBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create project: %d - %v", resp.StatusCode, result)
	}
	suite.projectID = result["id"].(string)
	suite.workflowID = result["workflow_id"].(string)
	t.Logf("  ✓ Created: %s (workflow: %s)", suite.projectID, suite.workflowID)

	t.Log("Step 2: Verify in list")
	resp, result = suite.request("GET", "/api/projects", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to list projects: %d", resp.StatusCode)
	}
	found := false
	projects := result["projects"].([]interface{})
	for _, p := range projects {
		if p.(map[string]interface{})["id"] == suite.projectID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Project not found in list")
	}
	t.Log("  ✓ Found in list")

	t.Log("Step 3: Get project details")
	resp, result = suite.request("GET", "/api/projects/"+suite.projectID, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to get project: %d", resp.StatusCode)
	}
	t.Logf("  ✓ Got details: phase=%v, status=%v", result["phase"], result["status"])

	t.Log("Step 4: Send approval message")
	time.Sleep(2 * time.Second) // Wait for workflow
	msgBody := map[string]string{"message": "/approve"}
	resp, _ = suite.request("POST", "/api/projects/"+suite.projectID+"/message", msgBody)
	if resp.StatusCode != http.StatusOK {
		t.Logf("  ⚠ Message send returned %d (workflow may have advanced)", resp.StatusCode)
	} else {
		t.Log("  ✓ Message sent")
	}

	t.Log("Step 5: Archive project")
	resp, _ = suite.request("DELETE", "/api/projects/"+suite.projectID, nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Failed to archive: %d", resp.StatusCode)
	}
	t.Log("  ✓ Archived")

	t.Log("Step 6: Verify not in list")
	resp, result = suite.request("GET", "/api/projects", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to list projects: %d", resp.StatusCode)
	}
	projects = result["projects"].([]interface{})
	for _, p := range projects {
		if p.(map[string]interface{})["id"] == suite.projectID {
			t.Fatal("Archived project should not be in list")
		}
	}
	t.Log("  ✓ Not in list after archive")

	t.Log("\n✓ Full project lifecycle completed successfully!")
}
