package contract

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/api"
)

// =============================================================================
// Frontend Types (mirrors frontend/app/lib/api/types.ts)
// =============================================================================

// FrontendCreateProjectRequest mirrors the frontend TypeScript CreateProjectRequest
type FrontendCreateProjectRequest struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	InitialMessage    string `json:"initial_message,omitempty"`
	EnableFullStack   bool   `json:"enable_full_stack,omitempty"`
	EnableDeployment  bool   `json:"enable_deployment,omitempty"`
	EnableSRE         bool   `json:"enable_sre,omitempty"`
	PreferredFrontend string `json:"preferred_frontend,omitempty"`
	PreferredBackend  string `json:"preferred_backend,omitempty"`
	PreferredDatabase string `json:"preferred_database,omitempty"`
	PreferredCloud    string `json:"preferred_cloud,omitempty"`
}

// FrontendProject mirrors the frontend TypeScript Project interface
type FrontendProject struct {
	ID          string `json:"id"`
	WorkflowID  string `json:"workflow_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Phase       string `json:"phase"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// FrontendProjectListResponse mirrors the frontend expected list response
type FrontendProjectListResponse struct {
	Projects []FrontendProject `json:"projects"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// FrontendSendMessageRequest mirrors the frontend TypeScript SendMessageRequest
type FrontendSendMessageRequest struct {
	Message string `json:"message"`
}

// FrontendSendMessageResponse mirrors the frontend TypeScript SendMessageResponse
type FrontendSendMessageResponse struct {
	Status  string `json:"status,omitempty"`
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
}

// =============================================================================
// Contract Tests
// =============================================================================

// TestCreateProjectRequestContract validates frontend/backend request compatibility
func TestCreateProjectRequestContract(t *testing.T) {
	// This is what the frontend sends
	frontendPayload := FrontendCreateProjectRequest{
		Name:              "Test Project",
		Description:       "A test project description",
		InitialMessage:    "Build a task management app",
		EnableFullStack:   true,
		EnableDeployment:  false,
		EnableSRE:         false,
		PreferredFrontend: "React",
		PreferredBackend:  "Go",
		PreferredDatabase: "PostgreSQL",
		PreferredCloud:    "AWS",
	}

	// Serialize as frontend would
	jsonBytes, err := json.Marshal(frontendPayload)
	if err != nil {
		t.Fatalf("Failed to marshal frontend payload: %v", err)
	}

	// Deserialize into backend struct
	var backendReq api.CreateProjectRequest
	if err := json.Unmarshal(jsonBytes, &backendReq); err != nil {
		t.Fatalf("Backend cannot deserialize frontend CreateProjectRequest: %v", err)
	}

	// Validate all fields transferred correctly
	if backendReq.Name != frontendPayload.Name {
		t.Errorf("Name mismatch: got %q, want %q", backendReq.Name, frontendPayload.Name)
	}
	if backendReq.Description != frontendPayload.Description {
		t.Errorf("Description mismatch: got %q, want %q", backendReq.Description, frontendPayload.Description)
	}
	if backendReq.InitialMessage != frontendPayload.InitialMessage {
		t.Errorf("InitialMessage mismatch: got %q, want %q", backendReq.InitialMessage, frontendPayload.InitialMessage)
	}
	if backendReq.EnableFullStack != frontendPayload.EnableFullStack {
		t.Errorf("EnableFullStack mismatch: got %v, want %v", backendReq.EnableFullStack, frontendPayload.EnableFullStack)
	}
	if backendReq.PreferredFrontend != frontendPayload.PreferredFrontend {
		t.Errorf("PreferredFrontend mismatch: got %q, want %q", backendReq.PreferredFrontend, frontendPayload.PreferredFrontend)
	}
	if backendReq.PreferredBackend != frontendPayload.PreferredBackend {
		t.Errorf("PreferredBackend mismatch: got %q, want %q", backendReq.PreferredBackend, frontendPayload.PreferredBackend)
	}
	if backendReq.PreferredDatabase != frontendPayload.PreferredDatabase {
		t.Errorf("PreferredDatabase mismatch: got %q, want %q", backendReq.PreferredDatabase, frontendPayload.PreferredDatabase)
	}
}

// TestProjectResponseContract validates backend response can be consumed by frontend
func TestProjectResponseContract(t *testing.T) {
	// This is what the backend sends
	backendResponse := api.ProjectResponse{
		ID:          "project-123",
		WorkflowID:  "consultancy_Test_1234567890",
		Name:        "Test Project",
		Description: "A test project",
		Phase:       "INTAKE",
		Status:      "active",
		CreatedBy:   "user-456",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	jsonBytes, err := json.Marshal(backendResponse)
	if err != nil {
		t.Fatalf("Failed to marshal backend response: %v", err)
	}

	// Deserialize into frontend struct
	var frontendResp FrontendProject
	if err := json.Unmarshal(jsonBytes, &frontendResp); err != nil {
		t.Fatalf("Frontend cannot deserialize backend ProjectResponse: %v", err)
	}

	// Validate key fields
	if frontendResp.ID != backendResponse.ID {
		t.Errorf("ID mismatch: got %q, want %q", frontendResp.ID, backendResponse.ID)
	}
	if frontendResp.WorkflowID != backendResponse.WorkflowID {
		t.Errorf("WorkflowID mismatch: got %q, want %q", frontendResp.WorkflowID, backendResponse.WorkflowID)
	}
	if frontendResp.Name != backendResponse.Name {
		t.Errorf("Name mismatch: got %q, want %q", frontendResp.Name, backendResponse.Name)
	}
	if frontendResp.Phase != backendResponse.Phase {
		t.Errorf("Phase mismatch: got %q, want %q", frontendResp.Phase, backendResponse.Phase)
	}
	if frontendResp.Status != backendResponse.Status {
		t.Errorf("Status mismatch: got %q, want %q", frontendResp.Status, backendResponse.Status)
	}
}

// TestProjectListResponseContract validates list response compatibility
func TestProjectListResponseContract(t *testing.T) {
	// This is what the backend sends
	backendResponse := api.ProjectListResponse{
		Projects: []api.ProjectResponse{
			{
				ID:         "project-1",
				WorkflowID: "wf-1",
				Name:       "Project 1",
				Phase:      "INTAKE",
				Status:     "active",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			{
				ID:         "project-2",
				WorkflowID: "wf-2",
				Name:       "Project 2",
				Phase:      "DEVELOPMENT",
				Status:     "active",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		Total:  2,
		Limit:  20,
		Offset: 0,
	}

	jsonBytes, err := json.Marshal(backendResponse)
	if err != nil {
		t.Fatalf("Failed to marshal backend response: %v", err)
	}

	// Deserialize into frontend struct
	var frontendResp FrontendProjectListResponse
	if err := json.Unmarshal(jsonBytes, &frontendResp); err != nil {
		t.Fatalf("Frontend cannot deserialize backend ProjectListResponse: %v", err)
	}

	// Validate fields
	if len(frontendResp.Projects) != len(backendResponse.Projects) {
		t.Errorf("Projects count mismatch: got %d, want %d", len(frontendResp.Projects), len(backendResponse.Projects))
	}
	if frontendResp.Total != backendResponse.Total {
		t.Errorf("Total mismatch: got %d, want %d", frontendResp.Total, backendResponse.Total)
	}
	if frontendResp.Limit != backendResponse.Limit {
		t.Errorf("Limit mismatch: got %d, want %d", frontendResp.Limit, backendResponse.Limit)
	}
	if frontendResp.Offset != backendResponse.Offset {
		t.Errorf("Offset mismatch: got %d, want %d", frontendResp.Offset, backendResponse.Offset)
	}

	// Validate first project
	if len(frontendResp.Projects) > 0 {
		if frontendResp.Projects[0].ID != backendResponse.Projects[0].ID {
			t.Errorf("First project ID mismatch")
		}
	}
}

// TestMessageRequestContract validates message request compatibility
func TestMessageRequestContract(t *testing.T) {
	// This is what the frontend sends
	frontendPayload := FrontendSendMessageRequest{
		Message: "/approve",
	}

	jsonBytes, err := json.Marshal(frontendPayload)
	if err != nil {
		t.Fatalf("Failed to marshal frontend payload: %v", err)
	}

	// Deserialize into backend struct
	var backendReq api.MessageRequest
	if err := json.Unmarshal(jsonBytes, &backendReq); err != nil {
		t.Fatalf("Backend cannot deserialize frontend MessageRequest: %v", err)
	}

	if backendReq.Message != frontendPayload.Message {
		t.Errorf("Message mismatch: got %q, want %q", backendReq.Message, frontendPayload.Message)
	}
}

// TestMessageResponseContract validates message response compatibility
func TestMessageResponseContract(t *testing.T) {
	// Backend sends status: "sent"
	backendResponse := `{"status":"sent"}`

	var frontendResp FrontendSendMessageResponse
	if err := json.Unmarshal([]byte(backendResponse), &frontendResp); err != nil {
		t.Fatalf("Frontend cannot deserialize backend message response: %v", err)
	}

	if frontendResp.Status != "sent" {
		t.Errorf("Status mismatch: got %q, want %q", frontendResp.Status, "sent")
	}
}

// TestProjectPhaseValues documents valid phase values
func TestProjectPhaseValues(t *testing.T) {
	validPhases := []string{
		"INTAKE",
		"SIZING",
		"PLANNING",
		"ARCHITECTURE",
		"DEVELOPMENT",
		"TESTING",
		"DEPLOYMENT",
		"OPERATIONS",
		"HANDOFF",
		"COMPLETE",
		"REVIEW",
		"FAILED",
	}

	t.Logf("Valid project phases:")
	for _, phase := range validPhases {
		t.Logf("  - %s", phase)
	}

	// Verify phases can be serialized/deserialized
	for _, phase := range validPhases {
		project := api.ProjectResponse{
			ID:        "test",
			Phase:     phase,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		jsonBytes, _ := json.Marshal(project)
		var decoded FrontendProject
		if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
			t.Errorf("Failed to deserialize phase %s: %v", phase, err)
		}
		if decoded.Phase != phase {
			t.Errorf("Phase mismatch for %s: got %s", phase, decoded.Phase)
		}
	}
}

// TestProjectStatusValues documents valid status values
func TestProjectStatusValues(t *testing.T) {
	validStatuses := []string{
		"active",
		"archived",
		"completed",
		"failed",
	}

	t.Logf("Valid project statuses:")
	for _, status := range validStatuses {
		t.Logf("  - %s", status)
	}
}

// TestMinimalCreateProjectRequest tests minimum required fields
func TestMinimalCreateProjectRequest(t *testing.T) {
	// Minimum valid request (only name is required)
	minimalRequest := `{"name":"Test Project"}`

	var backendReq api.CreateProjectRequest
	if err := json.Unmarshal([]byte(minimalRequest), &backendReq); err != nil {
		t.Fatalf("Failed to parse minimal request: %v", err)
	}

	if backendReq.Name == "" {
		t.Error("Name should be set")
	}

	// Optional fields should be zero values
	if backendReq.Description != "" {
		t.Error("Description should be empty")
	}
	if backendReq.EnableFullStack != false {
		t.Error("EnableFullStack should be false by default")
	}
}

// TestProjectConfigFields documents all project configuration fields
func TestProjectConfigFields(t *testing.T) {
	t.Log("CreateProjectRequest fields:")
	t.Log("  Required:")
	t.Log("    - name: string")
	t.Log("")
	t.Log("  Optional:")
	t.Log("    - description: string")
	t.Log("    - initial_message: string")
	t.Log("    - enable_full_stack: boolean")
	t.Log("    - enable_deployment: boolean")
	t.Log("    - enable_sre: boolean")
	t.Log("    - preferred_frontend: string (React, Vue, Angular, etc.)")
	t.Log("    - preferred_backend: string (Go, Python, Node.js, etc.)")
	t.Log("    - preferred_database: string (PostgreSQL, MySQL, MongoDB, etc.)")
	t.Log("    - preferred_cloud: string (AWS, GCP, Azure, etc.)")
}

// TestJSONFieldNamingConvention verifies snake_case JSON naming
func TestJSONFieldNamingConvention(t *testing.T) {
	// All JSON fields should use snake_case
	request := api.CreateProjectRequest{
		Name:              "Test",
		InitialMessage:    "message",
		EnableFullStack:   true,
		PreferredFrontend: "React",
	}

	jsonBytes, _ := json.Marshal(request)
	var rawMap map[string]interface{}
	json.Unmarshal(jsonBytes, &rawMap)

	// Check for snake_case keys
	expectedKeys := []string{
		"name",
		"initial_message",
		"enable_full_stack",
		"preferred_frontend",
	}

	for _, key := range expectedKeys {
		if _, exists := rawMap[key]; !exists {
			t.Errorf("Expected snake_case key %q not found", key)
		}
	}

	// Verify no camelCase keys
	camelCaseKeys := []string{
		"initialMessage",
		"enableFullStack",
		"preferredFrontend",
	}

	for _, key := range camelCaseKeys {
		if _, exists := rawMap[key]; exists {
			t.Errorf("Unexpected camelCase key %q found (should be snake_case)", key)
		}
	}
}

// BenchmarkProjectSerialization measures serialization performance
func BenchmarkProjectSerialization(b *testing.B) {
	project := api.ProjectResponse{
		ID:          "project-123",
		WorkflowID:  "consultancy_Test_1234567890",
		Name:        "Test Project",
		Description: "A test project description",
		Phase:       "DEVELOPMENT",
		Status:      "active",
		CreatedBy:   "user-456",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonBytes, _ := json.Marshal(project)
		var frontendProject FrontendProject
		json.Unmarshal(jsonBytes, &frontendProject)
	}
}
