package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qlfactory/sovereign-firm/pkg/auth"
	"github.com/qlfactory/sovereign-firm/pkg/cache"
	"github.com/qlfactory/sovereign-firm/pkg/firm/workflows"
	"github.com/qlfactory/sovereign-firm/pkg/storage"
	"github.com/qlfactory/sovereign-firm/pkg/streaming"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

// Handler implements the API handlers
type Handler struct {
	db             *pgxpool.Pool
	temporalClient client.Client
	streamHub      *streaming.Hub
	projectCache   *cache.ProjectCache
	artifactStore  *storage.ArtifactStore
}

// NewHandler creates a new API handler
func NewHandler(
	db interface{},
	temporalClient client.Client,
	streamHub *streaming.Hub,
	projectCache *cache.ProjectCache,
	artifactStore *storage.ArtifactStore,
) *Handler {
	var pool *pgxpool.Pool
	if db != nil {
		pool = db.(*pgxpool.Pool)
	}
	return &Handler{
		db:             pool,
		temporalClient: temporalClient,
		streamHub:      streamHub,
		projectCache:   projectCache,
		artifactStore:  artifactStore,
	}
}

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
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

// ProjectResponse represents a project in API responses
type ProjectResponse struct {
	ID          string    `json:"id"`
	WorkflowID  string    `json:"workflow_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Phase       string    `json:"phase"`
	Status      string    `json:"status"`
	CreatedBy   string    `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProjectListResponse represents a list of projects
type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// MessageRequest represents a message to send to a workflow
type MessageRequest struct {
	Message string `json:"message"`
}

// ListProjects lists all projects for the current tenant
func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := auth.GetTenantID(ctx)
	if tenantID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if DB is available
	if h.db == nil {
		jsonError(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	// Parse pagination params
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Set tenant context for RLS
	tx, err := h.db.Begin(ctx)
	if err != nil {
		log.Printf("ListProjects: begin tx error: %v", err)
		jsonError(w, "database error: begin", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	// SET LOCAL doesn't support parameterized values, so we use fmt.Sprintf
	// The tenantID is a validated UUID from the JWT token, so this is safe
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.tenant_id = '%s'", tenantID))
	if err != nil {
		log.Printf("ListProjects: set tenant error: %v", err)
		jsonError(w, "database error: set tenant", http.StatusInternalServerError)
		return
	}

	// Get total count
	var total int
	err = tx.QueryRow(ctx,
		"SELECT COUNT(*) FROM projects WHERE tenant_id = $1 AND status != 'archived'",
		tenantID,
	).Scan(&total)
	if err != nil {
		log.Printf("ListProjects: count error: %v", err)
		jsonError(w, "database error: count", http.StatusInternalServerError)
		return
	}

	// Get projects
	rows, err := tx.Query(ctx,
		`SELECT id, workflow_id, name, COALESCE(description, ''), phase, status,
		        COALESCE(created_by::text, ''), created_at, updated_at
		 FROM projects
		 WHERE tenant_id = $1 AND status != 'archived'
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		tenantID, limit, offset,
	)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	projects := make([]ProjectResponse, 0)
	for rows.Next() {
		var p ProjectResponse
		if err := rows.Scan(&p.ID, &p.WorkflowID, &p.Name, &p.Description, &p.Phase, &p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		projects = append(projects, p)
	}

	tx.Commit(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProjectListResponse{
		Projects: projects,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	})
}

// CreateProject creates a new project and starts a Temporal workflow
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := auth.GetTenantID(ctx)
	userID := auth.GetUserID(ctx)
	if tenantID == "" || userID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	// Generate IDs
	workflowID := fmt.Sprintf("consultancy_%s_%d", strings.ReplaceAll(req.Name, " ", "_"), time.Now().Unix())
	projectID := uuid.New().String()

	// Create workflow config
	config := workflows.ConsultancyConfig{
		ProjectID:         projectID,
		ProjectName:       req.Name,
		ClientID:          tenantID,
		InitialMessage:    req.InitialMessage,
		EnableFullStack:   req.EnableFullStack,
		EnableDeployment:  req.EnableDeployment,
		EnableSRE:         req.EnableSRE,
		PreferredFrontend: req.PreferredFrontend,
		PreferredBackend:  req.PreferredBackend,
		PreferredDatabase: req.PreferredDatabase,
		PreferredCloud:    req.PreferredCloud,
		MaxCodeAttempts:   3,
		MaxTestAttempts:   3,
	}

	// Serialize config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		jsonError(w, "failed to serialize config", http.StatusInternalServerError)
		return
	}

	// Start transaction
	tx, err := h.db.Begin(ctx)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	// Insert project into database
	var createdAt, updatedAt time.Time
	err = tx.QueryRow(ctx,
		`INSERT INTO projects (id, tenant_id, workflow_id, name, description, config, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING created_at, updated_at`,
		projectID, tenantID, workflowID, req.Name, req.Description, configJSON, userID,
	).Scan(&createdAt, &updatedAt)
	if err != nil {
		log.Printf("CreateProject: insert error: %v", err)
		jsonError(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	// Start Temporal workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "sovereign-firm-tasks",
	}

	we, err := h.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.ConsultancyWorkflow, config)
	if err != nil {
		jsonError(w, fmt.Sprintf("failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		jsonError(w, "failed to commit", http.StatusInternalServerError)
		return
	}

	// Broadcast phase change
	h.streamHub.BroadcastPhaseChange(workflowID, "", "INTAKE", "Consultancy workflow started")

	log.Printf("Created project %s with workflow %s (run: %s)", projectID, workflowID, we.GetRunID())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ProjectResponse{
		ID:          projectID,
		WorkflowID:  workflowID,
		Name:        req.Name,
		Description: req.Description,
		Phase:       "INTAKE",
		Status:      "active",
		CreatedBy:   userID,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	})
}

// GetProject gets a single project
func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := auth.GetTenantID(ctx)
	idOrWorkflowID := chi.URLParam(r, "projectId")

	if tenantID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tx, err := h.db.Begin(ctx)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	// SET LOCAL doesn't support parameterized values
	tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.tenant_id = '%s'", tenantID))

	query := `SELECT id, workflow_id, name, COALESCE(description, ''), phase, status,
		        COALESCE(created_by::text, ''), created_at, updated_at
		 FROM projects WHERE (id::text = $1 OR workflow_id = $1) AND tenant_id = $2`

	var p ProjectResponse
	err = tx.QueryRow(ctx, query, idOrWorkflowID, tenantID).Scan(
		&p.ID, &p.WorkflowID, &p.Name, &p.Description, &p.Phase, &p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		log.Printf("GetProject: query error for %s: %v", idOrWorkflowID, err)
		jsonError(w, "project not found", http.StatusNotFound)
		return
	}

	tx.Commit(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// GetProjectState gets the full Temporal workflow state
func (h *Handler) GetProjectState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := auth.GetTenantID(ctx)
	idOrWorkflowID := chi.URLParam(r, "projectId")

	if tenantID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get workflow ID and project ID from database
	var workflowID, projectID string
	err := h.db.QueryRow(ctx,
		"SELECT workflow_id, id FROM projects WHERE (id::text = $1 OR workflow_id = $1) AND tenant_id = $2",
		idOrWorkflowID, tenantID,
	).Scan(&workflowID, &projectID)
	if err != nil {
		jsonError(w, "project not found", http.StatusNotFound)
		return
	}

	// Try cache first (if available)
	if h.projectCache != nil {
		cachedState, _ := h.projectCache.GetState(ctx, projectID)
		if cachedState != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			json.NewEncoder(w).Encode(cachedState)
			return
		}
	}

	// Query Temporal
	resp, err := h.temporalClient.QueryWorkflow(ctx, workflowID, "", "get_state")
	if err != nil {
		log.Printf("GetProjectState: Temporal query failed for %s: %v", workflowID, err)
		// Return minimal state from database instead of 500
		var phase, status string
		h.db.QueryRow(ctx, "SELECT phase, status FROM projects WHERE workflow_id = $1", workflowID).Scan(&phase, &status)
		minimalState := workflows.ConsultancyState{
			ProjectID: projectID,
			Phase:     workflows.ConsultancyPhase(phase),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(minimalState)
		return
	}

	var state workflows.ConsultancyState
	if err := resp.Get(&state); err != nil {
		log.Printf("GetProjectState: Failed to decode state for %s: %v", workflowID, err)
		// Return minimal state instead of 500
		minimalState := workflows.ConsultancyState{
			ProjectID: projectID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(minimalState)
		return
	}

	// Cache the result (if cache available)
	if h.projectCache != nil {
		cacheState := &cache.ProjectState{
			Phase:     string(state.Phase),
			Status:    "active",
			Progress:  calculateProgress(string(state.Phase)),
			FileCount: len(state.AllCodeFiles),
		}
		h.projectCache.SetState(ctx, projectID, cacheState)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(state)
}

// SendMessage sends a message to the workflow
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := auth.GetTenantID(ctx)
	projectID := chi.URLParam(r, "projectId")

	if tenantID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get workflow ID
	var workflowID string
	err := h.db.QueryRow(ctx,
		"SELECT workflow_id FROM projects WHERE id = $1 AND tenant_id = $2",
		projectID, tenantID,
	).Scan(&workflowID)
	if err != nil {
		jsonError(w, "project not found", http.StatusNotFound)
		return
	}

	// Broadcast user message
	h.streamHub.BroadcastChatMessage(workflowID, "user", "", req.Message)

	// Signal Temporal
	signal := workflows.UserMessageSignal{Message: req.Message}
	err = h.temporalClient.SignalWorkflow(ctx, workflowID, "", "USER_MESSAGE", signal)
	if err != nil {
		h.streamHub.BroadcastError(workflowID, "SIGNAL_FAILED", "Failed to send message", err.Error())
		jsonError(w, fmt.Sprintf("failed to signal workflow: %v", err), http.StatusInternalServerError)
		return
	}

	// Invalidate cache (if available)
	if h.projectCache != nil {
		h.projectCache.InvalidateState(ctx, projectID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"sent"}`))
}

// StreamProject handles WebSocket connections
func (h *Handler) StreamProject(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	tenantID := auth.GetTenantID(r.Context())

	// For WebSocket, auth can come from query param
	if tenantID == "" {
		token := r.URL.Query().Get("token")
		if token != "" {
			// Validate token and extract tenant
			// This is simplified - in production, validate JWT
			tenantID = "default"
		}
	}

	// Get workflow ID
	var workflowID string
	err := h.db.QueryRow(r.Context(),
		"SELECT workflow_id FROM projects WHERE id = $1",
		projectID,
	).Scan(&workflowID)
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	log.Printf("WebSocket connection for project %s (workflow: %s)", projectID, workflowID)
	h.streamHub.ServeWs(w, r, workflowID)
}

// ArchiveProject archives a project
func (h *Handler) ArchiveProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := auth.GetTenantID(ctx)
	projectID := chi.URLParam(r, "projectId")

	if tenantID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := h.db.Exec(ctx,
		"UPDATE projects SET status = 'archived' WHERE id = $1 AND tenant_id = $2",
		projectID, tenantID,
	)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected() == 0 {
		jsonError(w, "project not found", http.StatusNotFound)
		return
	}

	// Invalidate cache (if available)
	if h.projectCache != nil {
		h.projectCache.InvalidateAll(ctx, projectID)
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListArtifacts lists project artifacts
func (h *Handler) ListArtifacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := auth.GetTenantID(ctx)
	projectID := chi.URLParam(r, "projectId")

	if tenantID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify project belongs to tenant
	var exists bool
	h.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND tenant_id = $2)",
		projectID, tenantID,
	).Scan(&exists)
	if !exists {
		jsonError(w, "project not found", http.StatusNotFound)
		return
	}

	var artifactType *storage.ArtifactType
	if t := r.URL.Query().Get("type"); t != "" {
		at := storage.ArtifactType(t)
		artifactType = &at
	}

	artifacts, err := h.artifactStore.List(ctx, tenantID, projectID, artifactType)
	if err != nil {
		jsonError(w, "failed to list artifacts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artifacts)
}

// GetArtifact returns an artifact (redirect to presigned URL)
func (h *Handler) GetArtifact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := auth.GetTenantID(ctx)
	projectID := chi.URLParam(r, "projectId")
	artifactID := chi.URLParam(r, "artifactId")

	if tenantID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get artifact path from database
	var path string
	err := h.db.QueryRow(ctx,
		`SELECT a.path FROM artifacts a
		 JOIN projects p ON a.project_id = p.id
		 WHERE a.id = $1 AND p.id = $2 AND p.tenant_id = $3`,
		artifactID, projectID, tenantID,
	).Scan(&path)
	if err != nil {
		jsonError(w, "artifact not found", http.StatusNotFound)
		return
	}

	// Generate presigned URL
	url, err := h.artifactStore.GetPresignedURL(ctx, path)
	if err != nil {
		jsonError(w, "failed to generate download URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

// HealthCheck returns basic health status
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

// ReadinessCheck checks all dependencies
func (h *Handler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := map[string]bool{
		"database": false,
		"temporal": false,
	}

	// Check database
	if err := h.db.Ping(ctx); err == nil {
		checks["database"] = true
	}

	// Check Temporal (simple connection check)
	if _, err := h.temporalClient.CheckHealth(ctx, nil); err == nil {
		checks["temporal"] = true
	}

	allReady := true
	for _, ready := range checks {
		if !ready {
			allReady = false
			break
		}
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allReady {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": status,
		"checks": checks,
	})
}

// Helper functions

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func calculateProgress(phase string) int {
	phases := map[string]int{
		"INTAKE":       5,
		"PLANNING":     15,
		"ARCHITECTURE": 25,
		"DATABASE":     35,
		"BACKEND":      50,
		"FRONTEND":     65,
		"INTEGRATION":  75,
		"TESTING":      85,
		"DEPLOYMENT":   95,
		"HANDOFF":      100,
		"COMPLETED":    100,
	}
	if progress, ok := phases[phase]; ok {
		return progress
	}
	return 0
}

// =========================================================================
// Brownfield Import Handlers
// =========================================================================

// ImportBrownfieldRequest represents a request to import an existing project
type ImportBrownfieldRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	RepoURL     string `json:"repo_url,omitempty"`
	LocalPath   string `json:"local_path,omitempty"`
}

// ImportBrownfieldResponse represents the response from starting a brownfield import
type ImportBrownfieldResponse struct {
	WorkflowID string `json:"workflow_id"`
	ProjectID  string `json:"project_id"`
	Status     string `json:"status"`
}

// ImportStatusResponse represents the status of a brownfield import
type ImportStatusResponse struct {
	Status        string   `json:"status"` // "analyzing", "complete", "error"
	FilesIndexed  int      `json:"files_indexed,omitempty"`
	ChunksCreated int      `json:"chunks_created,omitempty"`
	SymbolsFound  int      `json:"symbols_found,omitempty"`
	PrimaryLang   string   `json:"primary_lang,omitempty"`
	DetectedStack []string `json:"detected_stack,omitempty"`
	Error         string   `json:"error,omitempty"`
	Phase         string   `json:"phase,omitempty"`
}

// ImportBrownfield starts a brownfield import workflow
func (h *Handler) ImportBrownfield(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := auth.GetTenantID(ctx)
	userID := auth.GetUserID(ctx)
	if tenantID == "" || userID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req ImportBrownfieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.RepoURL == "" && req.LocalPath == "" {
		jsonError(w, "either repo_url or local_path is required", http.StatusBadRequest)
		return
	}

	// Validate Git URL format if provided
	if req.RepoURL != "" {
		if !strings.HasPrefix(req.RepoURL, "http://") &&
			!strings.HasPrefix(req.RepoURL, "https://") &&
			!strings.HasPrefix(req.RepoURL, "git@") {
			jsonError(w, "invalid repository URL format", http.StatusBadRequest)
			return
		}
	}

	// Generate workflow ID
	workflowID := fmt.Sprintf("brownfield_%s_%d", strings.ReplaceAll(req.Name, " ", "_"), time.Now().Unix())

	// Create workflow config with brownfield settings
	config := workflows.ConsultancyConfig{
		ProjectName:     req.Name,
		ClientID:        tenantID,
		InitialMessage:  fmt.Sprintf("Importing existing project: %s", req.Description),
		MaxCodeAttempts: 3,
		MaxTestAttempts: 3,
		// Brownfield-specific config
		BrownfieldConfig: &workflows.BrownfieldConfig{
			RepoURL:   req.RepoURL,
			LocalPath: req.LocalPath,
		},
	}

	// Serialize config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		jsonError(w, "failed to serialize config", http.StatusInternalServerError)
		return
	}

	// Check if DB is available
	if h.db == nil {
		jsonError(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	// Start transaction
	tx, err := h.db.Begin(ctx)
	if err != nil {
		jsonError(w, "database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	// Insert project into database with brownfield type
	var projectID string
	err = tx.QueryRow(ctx,
		`INSERT INTO projects (tenant_id, workflow_id, name, description, config, created_by, phase)
		 VALUES ($1, $2, $3, $4, $5, $6, 'ANALYZING')
		 RETURNING id`,
		tenantID, workflowID, req.Name, req.Description, configJSON, userID,
	).Scan(&projectID)
	if err != nil {
		log.Printf("ImportBrownfield: insert error: %v", err)
		jsonError(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	// Start Temporal workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "sovereign-firm-tasks",
	}

	_, err = h.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.ConsultancyWorkflow, config)
	if err != nil {
		jsonError(w, fmt.Sprintf("failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		jsonError(w, "failed to commit", http.StatusInternalServerError)
		return
	}

	// Broadcast phase change
	h.streamHub.BroadcastPhaseChange(workflowID, "", "ANALYZING", "Brownfield import started")

	log.Printf("Started brownfield import for project %s with workflow %s", projectID, workflowID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(ImportBrownfieldResponse{
		WorkflowID: workflowID,
		ProjectID:  projectID,
		Status:     "analyzing",
	})
}

// GetImportStatus returns the status of a brownfield import
func (h *Handler) GetImportStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := auth.GetTenantID(ctx)
	projectID := chi.URLParam(r, "projectId")

	if tenantID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if DB is available
	if h.db == nil {
		jsonError(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	// Get workflow ID from database
	var workflowID, phase string
	err := h.db.QueryRow(ctx,
		"SELECT workflow_id, phase FROM projects WHERE id = $1 AND tenant_id = $2",
		projectID, tenantID,
	).Scan(&workflowID, &phase)
	if err != nil {
		jsonError(w, "project not found", http.StatusNotFound)
		return
	}

	// Query Temporal for brownfield status
	resp, err := h.temporalClient.QueryWorkflow(ctx, workflowID, "", "get_brownfield_status")
	if err != nil {
		log.Printf("GetImportStatus: Temporal query failed for %s: %v", workflowID, err)
		// If query fails, return basic status from database
		status := "analyzing"
		if phase == "PLANNING" || phase == "ARCHITECTURE" || phase == "SIZING" || phase == "DEVELOPMENT" {
			status = "complete"
		} else if phase == "FAILED" {
			status = "error"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ImportStatusResponse{
			Status: status,
			Phase:  phase,
		})
		return
	}

	var bfStatus workflows.BrownfieldStatus
	if err := resp.Get(&bfStatus); err != nil {
		log.Printf("GetImportStatus: Failed to decode brownfield status for %s: %v", workflowID, err)
		// Return basic status if can't decode
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ImportStatusResponse{
			Status: "analyzing",
			Phase:  phase,
		})
		return
	}

	// If still analyzing, try to get granular progress from pending activity heartbeats
	if bfStatus.Status == "analyzing" {
		desc, err := h.temporalClient.DescribeWorkflowExecution(ctx, workflowID, "")
		if err == nil {
			for _, pending := range desc.PendingActivities {
				if pending.ActivityType.GetName() == "BrownfieldAnalyzeProject" && pending.HeartbeatDetails != nil {
					// The activity heartbeats with a subset of BrownfieldResult
					var heartbeat struct {
						FilesIndexed  int `json:"files_indexed"`
						ChunksCreated int `json:"chunks_created"`
						SymbolsFound  int `json:"symbols_found"`
					}
					// Use the data converter to decode heartbeat details
					if err := converter.GetDefaultDataConverter().FromPayloads(pending.HeartbeatDetails, &heartbeat); err == nil {
						bfStatus.FilesIndexed = heartbeat.FilesIndexed
						bfStatus.ChunksCreated = heartbeat.ChunksCreated
						bfStatus.SymbolsFound = heartbeat.SymbolsFound
					}
				}
			}
		}
	}

	log.Printf("GetImportStatus: %s status=%s files=%d chunks=%d phase=%s",
		workflowID, bfStatus.Status, bfStatus.FilesIndexed, bfStatus.ChunksCreated, bfStatus.Phase)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ImportStatusResponse{
		Status:        bfStatus.Status,
		Phase:         bfStatus.Phase,
		FilesIndexed:  bfStatus.FilesIndexed,
		ChunksCreated: bfStatus.ChunksCreated,
		SymbolsFound:  bfStatus.SymbolsFound,
		PrimaryLang:   bfStatus.PrimaryLang,
		DetectedStack: bfStatus.DetectedStack,
		Error:         bfStatus.Error,
	})
}
