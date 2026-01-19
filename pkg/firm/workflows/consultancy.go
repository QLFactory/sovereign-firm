// Package workflows provides Temporal workflows for the Sovereign Firm.
package workflows

import (
	"fmt"
	"strings"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/firm/activities"
	"go.temporal.io/sdk/workflow"
)

// CodeBundle represents generated code files from backend activities
type CodeBundle struct {
	Stack      string            `json:"stack"`
	Files      map[string]string `json:"files"`
	EntryPoint string            `json:"entry_point"`
	StartCmd   string            `json:"start_cmd"`
	InstallCmd string            `json:"install_cmd"`
	TestCmd    string            `json:"test_cmd"`
	BuildCmd   string            `json:"build_cmd"`
}

// DatabaseBundle represents generated code files from database activities
type DatabaseBundle struct {
	DatabaseType string            `json:"database_type"`
	ORM          string            `json:"orm"`
	Files        map[string]string `json:"files"`
	MigrateCmd   string            `json:"migrate_cmd"`
	SeedCmd      string            `json:"seed_cmd"`
	ResetCmd     string            `json:"reset_cmd"`
}

// DevOpsBundle represents generated files from DevOps activities
type DevOpsBundle struct {
	Files       map[string]string `json:"files"`
	BuildCmd    string            `json:"build_cmd"`
	DeployCmd   string            `json:"deploy_cmd"`
	TestCmd     string            `json:"test_cmd"`
	Environment string            `json:"environment"`
}

// ConsultancyPhase represents the current phase of the consultancy workflow
type ConsultancyPhase string

const (
	PhaseIntake       ConsultancyPhase = "INTAKE"
	PhaseSizing       ConsultancyPhase = "SIZING"
	PhasePlanning     ConsultancyPhase = "PLANNING"
	PhaseArchitecture ConsultancyPhase = "ARCHITECTURE"
	PhaseDevelopment  ConsultancyPhase = "DEVELOPMENT"
	PhaseTesting      ConsultancyPhase = "TESTING"
	PhaseDeployment   ConsultancyPhase = "DEPLOYMENT"
	PhaseOperations   ConsultancyPhase = "OPERATIONS"
	PhaseHandoff      ConsultancyPhase = "HANDOFF"
	PhaseComplete     ConsultancyPhase = "COMPLETE"
	PhaseReview       ConsultancyPhase = "REVIEW"
	PhaseFailed       ConsultancyPhase = "FAILED"
)

// BrownfieldConfig contains configuration for brownfield import
type BrownfieldConfig struct {
	RepoURL      string   `json:"repo_url,omitempty"`      // Git repository URL to clone
	LocalPath    string   `json:"local_path,omitempty"`    // Local directory path
	MaxFiles     int      `json:"max_files,omitempty"`     // Maximum files to index (0 = unlimited)
	SkipPatterns []string `json:"skip_patterns,omitempty"` // Glob patterns to skip
	IncludeTests bool     `json:"include_tests,omitempty"` // Include test files in analysis
}

// BrownfieldStatus represents the status of a brownfield import operation
type BrownfieldStatus struct {
	Status        string   `json:"status"` // "analyzing", "complete", "error"
	FilesIndexed  int      `json:"files_indexed"`
	ChunksCreated int      `json:"chunks_created"`
	SymbolsFound  int      `json:"symbols_found"`
	PrimaryLang   string   `json:"primary_lang,omitempty"`
	DetectedStack []string `json:"detected_stack,omitempty"`
	Error         string   `json:"error,omitempty"`
	Phase         string   `json:"phase,omitempty"`
}

// ConsultancyConfig contains configuration for the consultancy workflow
type ConsultancyConfig struct {
	// Project settings
	ProjectID        string `json:"project_id"` // Database UUID
	ProjectName      string `json:"project_name"`
	ClientID         string `json:"client_id"`
	InitialMessage   string `json:"initial_message"`   // Initial project description (skips intake chat)
	EnableFullStack  bool   `json:"enable_full_stack"` // Full stack vs frontend-only
	EnableDeployment bool   `json:"enable_deployment"` // Generate deployment configs
	EnableSRE        bool   `json:"enable_sre"`        // Generate monitoring/alerting

	// Tech preferences (if empty, architect will decide)
	PreferredFrontend string `json:"preferred_frontend"` // React, Vue, Angular
	PreferredBackend  string `json:"preferred_backend"`  // Node.js, Python, Go
	PreferredDatabase string `json:"preferred_database"` // PostgreSQL, MongoDB
	PreferredCloud    string `json:"preferred_cloud"`    // AWS, Azure, GCP

	// Limits
	MaxCodeAttempts int `json:"max_code_attempts"` // Default 3
	MaxTestAttempts int `json:"max_test_attempts"` // Default 3

	// Brownfield import (optional)
	BrownfieldConfig *BrownfieldConfig `json:"brownfield_config,omitempty"`
}

// ConsultancyState holds the complete state of a consultancy project
type ConsultancyState struct {
	// Project identity
	ProjectID   string           `json:"project_id"`  // Database UUID
	WorkflowID  string           `json:"workflow_id"` // Temporal ID
	ProjectName string           `json:"project_name"`
	ClientID    string           `json:"client_id"`
	Phase       ConsultancyPhase `json:"phase"`
	StartedAt   time.Time        `json:"started_at"`
	UpdatedAt   time.Time        `json:"updated_at"`

	// Discovery/Intake
	ChatHistory  string `json:"chat_history"`
	Requirements string `json:"requirements"` // Structured requirements document

	// Planning
	ComplexityScore map[string]interface{} `json:"complexity_score"`
	EffortEstimate  map[string]interface{} `json:"effort_estimate"`
	ProjectPlan     map[string]interface{} `json:"project_plan"`
	Proposal        map[string]interface{} `json:"proposal"`

	// Architecture
	TechStack      map[string]interface{} `json:"tech_stack"`
	SystemDesign   map[string]interface{} `json:"system_design"`
	DatabaseSchema map[string]interface{} `json:"database_schema"`
	APISpec        map[string]interface{} `json:"api_spec"`

	// Development artifacts
	FrontendCode map[string]string `json:"frontend_code"`
	BackendCode  map[string]string `json:"backend_code"`
	DatabaseCode map[string]string `json:"database_code"` // Migrations, schemas
	AllCodeFiles map[string]string `json:"all_code_files"`

	// Testing artifacts
	UnitTests        map[string]string      `json:"unit_tests"`
	IntegrationTests map[string]string      `json:"integration_tests"`
	E2ETests         map[string]string      `json:"e2e_tests"`
	TestCoverage     map[string]interface{} `json:"test_coverage"`

	// Deployment artifacts
	Dockerfile    string            `json:"dockerfile"`
	DockerCompose string            `json:"docker_compose"`
	KubeManifests map[string]string `json:"kube_manifests"`
	HelmChart     map[string]string `json:"helm_chart"`
	CIPipeline    string            `json:"ci_pipeline"`
	InfraCode     map[string]string `json:"infra_code"` // Terraform, etc.

	// Operations artifacts
	MonitoringConfig map[string]interface{} `json:"monitoring_config"`
	AlertRules       map[string]interface{} `json:"alert_rules"`
	Dashboards       map[string]interface{} `json:"dashboards"`
	Runbooks         map[string]string      `json:"runbooks"`

	// Quality metrics
	ValidationResults map[string]interface{} `json:"validation_results"`
	CriticReview      map[string]interface{} `json:"critic_review"`
	TestResults       map[string]interface{} `json:"test_results"`

	// Status tracking
	Errors       []string   `json:"errors"`
	Warnings     []string   `json:"warnings"`
	PhaseHistory []string   `json:"phase_history"`
	CompletedAt  *time.Time `json:"completed_at"`

	// Brownfield import status
	BrownfieldStatus *BrownfieldStatus `json:"brownfield_status,omitempty"`

	// Multi-agent orchestration (real-time)
	DAG      *DAGState      `json:"dag,omitempty"`
	Agents   []ActiveAgent  `json:"agents,omitempty"`
	CIStatus *CIStatusState `json:"ci_status,omitempty"`
}

// DAGState represents the real-time state of the task graph
type DAGState struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Tasks     []DAGTask `json:"tasks"`
	Total     int       `json:"total"`
	Completed int       `json:"completed"`
	Failed    int       `json:"failed"`
	Running   int       `json:"running"`
	Pending   int       `json:"pending"`
}

// DAGTask represents a single task in the DAG
type DAGTask struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description,omitempty"`
	Type         string   `json:"type"`
	Status       string   `json:"status"` // PENDING, READY, RUNNING, COMPLETED, FAILED, BLOCKED
	Dependencies []string `json:"dependencies"`
	AssignedTo   string   `json:"assigned_to,omitempty"`
	Error        string   `json:"error,omitempty"`
	RetryCount   int      `json:"retry_count,omitempty"`
}

// ActiveAgent represents a currently running agent
type ActiveAgent struct {
	TaskID    string `json:"task_id"`
	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
	TaskName  string `json:"task_name"`
	StartedAt string `json:"started_at"`
}

// CIStatusState represents real-time CI status
type CIStatusState struct {
	Stages []CIStageResult `json:"stages"`
}

// CIStageResult represents the result of a single CI stage
type CIStageResult struct {
	Stage      string `json:"stage"` // LINT, BUILD, TEST
	Success    bool   `json:"success"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
	DurationMS int64  `json:"duration_ms,omitempty"`
}

// ConsultancyWorkflow is the main workflow for end-to-end project delivery
func ConsultancyWorkflow(ctx workflow.Context, config ConsultancyConfig) (*ConsultancyState, error) {
	logger := workflow.GetLogger(ctx)

	// Set defaults
	if config.MaxCodeAttempts <= 0 {
		config.MaxCodeAttempts = 3
	}
	if config.MaxTestAttempts <= 0 {
		config.MaxTestAttempts = 3
	}

	// Initialize state
	state := &ConsultancyState{
		ProjectID:    config.ProjectID, // Use database UUID
		WorkflowID:   workflow.GetInfo(ctx).WorkflowExecution.ID,
		ProjectName:  config.ProjectName,
		ClientID:     config.ClientID,
		Phase:        PhaseIntake,
		StartedAt:    workflow.Now(ctx),
		UpdatedAt:    workflow.Now(ctx),
		AllCodeFiles: make(map[string]string),
		PhaseHistory: []string{string(PhaseIntake)},
		DAG: &DAGState{
			ID:        "dag_" + workflow.GetInfo(ctx).WorkflowExecution.ID,
			ProjectID: workflow.GetInfo(ctx).WorkflowExecution.ID,
			Tasks: []DAGTask{
				{ID: "intake", Name: "Project Intake", Type: "discovery", Status: "RUNNING"},
				{ID: "sizing", Name: "Complexity Sizing", Type: "planning", Status: "PENDING", Dependencies: []string{"intake"}},
				{ID: "planning", Name: "Project Planning", Type: "planning", Status: "PENDING", Dependencies: []string{"sizing"}},
				{ID: "architecture", Name: "System Architecture", Type: "design", Status: "PENDING", Dependencies: []string{"planning"}},
				{ID: "development", Name: "Code Generation", Type: "development", Status: "PENDING", Dependencies: []string{"architecture"}},
				{ID: "testing", Name: "Quality Assurance", Type: "testing", Status: "PENDING", Dependencies: []string{"development"}},
				{ID: "deployment", Name: "Cloud Deployment", Type: "devops", Status: "PENDING", Dependencies: []string{"testing"}},
			},
			Total:   7,
			Pending: 6,
			Running: 1,
		},
	}

	// Activity options
	shortAO := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 2,
	}
	mediumAO := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
	}
	longAO := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
	}

	ctxShort := workflow.WithActivityOptions(ctx, shortAO)
	ctxMedium := workflow.WithActivityOptions(ctx, mediumAO)
	ctxLong := workflow.WithActivityOptions(ctx, longAO)

	// Register query handler for full state
	err := workflow.SetQueryHandler(ctx, "get_state", func() (*ConsultancyState, error) {
		return state, nil
	})
	if err != nil {
		return nil, err
	}

	// Register query handler for brownfield status
	err = workflow.SetQueryHandler(ctx, "get_brownfield_status", func() (*BrownfieldStatus, error) {
		if state.BrownfieldStatus != nil {
			return state.BrownfieldStatus, nil
		}
		return &BrownfieldStatus{Status: "not_started"}, nil
	})
	if err != nil {
		return nil, err
	}

	// Signal channel for user messages
	msgChan := workflow.GetSignalChannel(ctx, "USER_MESSAGE")

	// ============================================================
	// BROWNFIELD ANALYSIS (if configured)
	// ============================================================
	if config.BrownfieldConfig != nil && (config.BrownfieldConfig.RepoURL != "" || config.BrownfieldConfig.LocalPath != "") {
		logger.Info("Starting brownfield analysis")

		// Initialize brownfield status
		state.BrownfieldStatus = &BrownfieldStatus{Status: "analyzing"}

		// Prepare brownfield params
		brownfieldInput := map[string]interface{}{
			"project_id": state.ProjectID,
			"client_id":  state.ClientID,
		}
		if config.BrownfieldConfig.RepoURL != "" {
			brownfieldInput["repo_url"] = config.BrownfieldConfig.RepoURL
		}
		if config.BrownfieldConfig.LocalPath != "" {
			brownfieldInput["local_path"] = config.BrownfieldConfig.LocalPath
		}
		if config.BrownfieldConfig.MaxFiles > 0 {
			brownfieldInput["max_files"] = config.BrownfieldConfig.MaxFiles
		}
		if len(config.BrownfieldConfig.SkipPatterns) > 0 {
			brownfieldInput["skip_patterns"] = config.BrownfieldConfig.SkipPatterns
		}
		brownfieldInput["include_tests"] = config.BrownfieldConfig.IncludeTests

		// Execute brownfield analysis activity
		var brownfieldResult map[string]interface{}
		if err := workflow.ExecuteActivity(ctxLong, "BrownfieldAnalyzeProject", brownfieldInput).Get(ctx, &brownfieldResult); err != nil {
			logger.Error("Brownfield analysis failed", "Error", err)
			state.BrownfieldStatus = &BrownfieldStatus{
				Status: "error",
				Error:  err.Error(),
			}
			state.Errors = append(state.Errors, fmt.Sprintf("Brownfield analysis: %v", err))
		} else {
			// Update brownfield status from result
			filesIndexed := 0
			if v, ok := brownfieldResult["files_indexed"].(float64); ok {
				filesIndexed = int(v)
			} else if v, ok := brownfieldResult["files_indexed"].(int); ok {
				filesIndexed = v
			}
			chunksCreated := 0
			if v, ok := brownfieldResult["chunks_created"].(float64); ok {
				chunksCreated = int(v)
			} else if v, ok := brownfieldResult["chunks_created"].(int); ok {
				chunksCreated = v
			}
			symbolsFound := 0
			if v, ok := brownfieldResult["symbols_found"].(float64); ok {
				symbolsFound = int(v)
			} else if v, ok := brownfieldResult["symbols_found"].(int); ok {
				symbolsFound = v
			}

			state.BrownfieldStatus = &BrownfieldStatus{
				Status:        "complete",
				FilesIndexed:  filesIndexed,
				ChunksCreated: chunksCreated,
				SymbolsFound:  symbolsFound,
			}

			// Extract tech stack from analysis if available
			if stack, ok := brownfieldResult["stack"].(map[string]interface{}); ok {
				if state.TechStack == nil {
					state.TechStack = make(map[string]interface{})
				}
				for k, v := range stack {
					state.TechStack[k] = v
				}
			}

			// Store recommendations
			if recommendations, ok := brownfieldResult["recommendations"].([]interface{}); ok {
				for _, rec := range recommendations {
					if r, ok := rec.(string); ok {
						state.Warnings = append(state.Warnings, r)
					}
				}
			}

			logger.Info("Brownfield analysis completed",
				"FilesIndexed", filesIndexed,
				"ChunksCreated", chunksCreated,
				"SymbolsFound", symbolsFound)
		}

		state.UpdatedAt = workflow.Now(ctx)
	}

	// ============================================================
	// PHASE 1: INTAKE (Discovery)
	// ============================================================
	logger.Info("Starting INTAKE phase")
	state.Phase = PhaseIntake

	// If InitialMessage is provided, use it directly (skip interactive chat)
	if config.InitialMessage != "" {
		logger.Info("Using provided initial message, skipping interactive chat")
		state.ChatHistory = "User: " + config.InitialMessage

		var reply string
		chatInput := activities.ChatInput{
			ProjectID:   state.ProjectID,
			ChatHistory: state.ChatHistory,
		}
		err := workflow.ExecuteActivity(ctxShort, "PMAgentChat", chatInput).Get(ctx, &reply)
		if err != nil {
			logger.Warn("PM Agent chat failed", "Error", err)
			state.Errors = append(state.Errors, fmt.Sprintf("PM chat error: %v", err))
		} else {
			state.ChatHistory += "\nPM: " + reply
		}
		state.UpdatedAt = workflow.Now(ctx)
	} else {
		// Interactive mode - wait for signals
		for {
			var signal UserMessageSignal
			msgChan.Receive(ctx, &signal)

			if signal.Message == "/approve" {
				break
			}

			state.ChatHistory += "\nUser: " + signal.Message

			var reply string
			chatInput := activities.ChatInput{
				ProjectID:   state.ProjectID,
				ChatHistory: state.ChatHistory,
			}
			err := workflow.ExecuteActivity(ctxShort, "PMAgentChat", chatInput).Get(ctx, &reply)
			if err != nil {
				logger.Error("PM Agent chat failed", "Error", err)
				state.Errors = append(state.Errors, fmt.Sprintf("PM chat error: %v", err))
				continue
			}

			state.ChatHistory += "\nPM: " + reply
			state.UpdatedAt = workflow.Now(ctx)
		}
	}

	// Gather structured requirements
	logger.Info("Gathering structured requirements")
	reqInput := map[string]interface{}{
		"chat_history": state.ChatHistory,
		"client_id":    state.ClientID,
	}

	var reqResult map[string]interface{}
	if err := workflow.ExecuteActivity(ctxMedium, "PMGatherRequirements", reqInput).Get(ctx, &reqResult); err != nil {
		logger.Warn("Failed to gather structured requirements", "Error", err)
		// Use chat history as requirements fallback
		state.Requirements = state.ChatHistory
	} else {
		state.Requirements = state.ChatHistory // Keep as string for reference
	}

	transitionPhaseWithDB(ctx, state, PhaseSizing)

	// ============================================================
	// PHASE 2: SIZING (Complexity & Effort Estimation)
	// ============================================================
	logger.Info("Starting SIZING phase")

	// Estimate complexity - pass requirements doc if available, otherwise description
	complexityInput := map[string]interface{}{
		"description": state.ChatHistory, // Use chat history as description
	}
	if reqResult != nil {
		complexityInput["requirements"] = reqResult
	}

	if err := workflow.ExecuteActivity(ctxMedium, "PMEstimateComplexity", complexityInput).Get(ctx, &state.ComplexityScore); err != nil {
		logger.Warn("Complexity estimation failed", "Error", err)
	}

	// Estimate effort
	effortInput := map[string]interface{}{
		"description": state.ChatHistory,
	}
	if reqResult != nil {
		effortInput["requirements"] = reqResult
	}
	if state.ComplexityScore != nil {
		effortInput["complexity_score"] = state.ComplexityScore
	}

	if err := workflow.ExecuteActivity(ctxMedium, "PMEstimateEffort", effortInput).Get(ctx, &state.EffortEstimate); err != nil {
		logger.Warn("Effort estimation failed", "Error", err)
	}

	transitionPhaseWithDB(ctx, state, PhasePlanning)

	// ============================================================
	// PHASE 3: PLANNING (Project Plan & Proposal)
	// ============================================================
	logger.Info("Starting PLANNING phase")

	// Create project plan
	planInput := map[string]interface{}{
		"description":   state.ChatHistory,
		"effort":        state.EffortEstimate,
		"methodology":   "agile",
		"sprint_length": 14,
	}
	if reqResult != nil {
		planInput["requirements"] = reqResult
	}

	if err := workflow.ExecuteActivity(ctxMedium, "PMCreateProjectPlan", planInput).Get(ctx, &state.ProjectPlan); err != nil {
		logger.Warn("Project plan creation failed", "Error", err)
	}

	// Generate proposal
	proposalInput := map[string]interface{}{
		"description":      state.ChatHistory,
		"complexity_score": state.ComplexityScore,
		"effort_estimate":  state.EffortEstimate,
		"project_plan":     state.ProjectPlan,
		"client_name":      config.ClientID,
	}
	if reqResult != nil {
		proposalInput["requirements"] = reqResult
	}

	if err := workflow.ExecuteActivity(ctxMedium, "PMGenerateProposal", proposalInput).Get(ctx, &state.Proposal); err != nil {
		logger.Warn("Proposal generation failed", "Error", err)
	}

	transitionPhaseWithDB(ctx, state, PhaseArchitecture)

	// ============================================================
	// PHASE 4: ARCHITECTURE (System Design)
	// ============================================================
	logger.Info("Starting ARCHITECTURE phase")

	// Analyze requirements for architecture
	archInput := map[string]interface{}{
		"description":  state.ChatHistory,
		"chat_history": state.ChatHistory,
	}
	if reqResult != nil {
		archInput["requirements"] = reqResult
	}

	var systemDesign map[string]interface{}
	if err := workflow.ExecuteActivity(ctxMedium, "ArchitectAnalyzeRequirements", archInput).Get(ctx, &systemDesign); err != nil {
		logger.Warn("Architecture analysis failed", "Error", err)
	} else {
		state.SystemDesign = systemDesign
	}

	// Select tech stack
	techInput := map[string]interface{}{
		"requirements":       state.Requirements,
		"system_design":      state.SystemDesign,
		"preferred_frontend": config.PreferredFrontend,
		"preferred_backend":  config.PreferredBackend,
		"preferred_database": config.PreferredDatabase,
		"preferred_cloud":    config.PreferredCloud,
	}

	if err := workflow.ExecuteActivity(ctxMedium, "ArchitectSelectTechStack", techInput).Get(ctx, &state.TechStack); err != nil {
		logger.Warn("Tech stack selection failed", "Error", err)
		// Default tech stack
		state.TechStack = map[string]interface{}{
			"frontend": firstNonEmpty(config.PreferredFrontend, "react"),
			"backend":  firstNonEmpty(config.PreferredBackend, "nodejs"),
			"database": firstNonEmpty(config.PreferredDatabase, "postgresql"),
			"cloud":    firstNonEmpty(config.PreferredCloud, "aws"),
		}
	}

	// Design database schema (if full stack)
	if config.EnableFullStack {
		dbInput := map[string]interface{}{
			"requirements":  state.Requirements,
			"system_design": state.SystemDesign,
			"database_type": state.TechStack["database"],
		}

		if err := workflow.ExecuteActivity(ctxMedium, "ArchitectDesignDatabase", dbInput).Get(ctx, &state.DatabaseSchema); err != nil {
			logger.Warn("Database design failed", "Error", err)
		}

		// Design API spec
		apiInput := map[string]interface{}{
			"requirements":    state.Requirements,
			"system_design":   state.SystemDesign,
			"database_schema": state.DatabaseSchema,
		}

		if err := workflow.ExecuteActivity(ctxMedium, "ArchitectDesignAPI", apiInput).Get(ctx, &state.APISpec); err != nil {
			logger.Warn("API design failed", "Error", err)
		}
	}

	transitionPhaseWithDB(ctx, state, PhaseDevelopment)

	// ============================================================
	// PHASE 5: DEVELOPMENT (Code Generation)
	// ============================================================
	logger.Info("Starting DEVELOPMENT phase")

	// Determine what to build based on tech stack
	frontendTech := getString(state.TechStack, "frontend", "react")
	backendTech := getString(state.TechStack, "backend", "nodejs")
	databaseTech := getString(state.TechStack, "database", "postgresql")

	// === PARALLEL DEVELOPMENT ===
	// We'll use Temporal's ability to run activities in parallel

	// Create a selector for parallel execution
	selector := workflow.NewSelector(ctx)
	var frontendErr, backendErr, databaseErr error

	// Frontend development - DevAgentGenerate expects a string spec
	frontendSpec := fmt.Sprintf(`Project: %s
Tech Stack: %s

Requirements:
%s

API Specification:
%v

Build a modern, responsive frontend application.`, config.ProjectName, frontendTech, state.ChatHistory, state.APISpec)

	genInput := activities.GenerateCodeInput{
		ProjectID: state.ProjectID,
		Spec:      frontendSpec,
	}

	// Track agents
	state.Agents = append(state.Agents, ActiveAgent{
		TaskID:    "frontend",
		AgentID:   "dev-frontend",
		AgentName: "Dev Agent",
		TaskName:  "Frontend Development",
		StartedAt: workflow.Now(ctx).Format(time.RFC3339),
	})

	frontendFuture := workflow.ExecuteActivity(ctxLong, "DevAgentGenerate", genInput)
	selector.AddFuture(frontendFuture, func(f workflow.Future) {
		var result map[string]string
		frontendErr = f.Get(ctx, &result)
		if frontendErr == nil {
			state.FrontendCode = result
			for k, v := range result {
				state.AllCodeFiles[k] = v
			}
		}
		// Remove agent
		state.Agents = filterAgents(state.Agents, "frontend")
	})

	// Backend development (if full stack)
	if config.EnableFullStack {
		backendInput := map[string]interface{}{
			"project_name": config.ProjectName,
			"requirements": state.ChatHistory,
			"tech_stack": map[string]interface{}{
				"backend":  backendTech,
				"database": databaseTech,
			},
			"database_schema": state.DatabaseSchema,
			"api_spec":        state.APISpec,
		}

		// Track agents
		state.Agents = append(state.Agents, ActiveAgent{
			TaskID:    "backend",
			AgentID:   "dev-backend",
			AgentName: "Backend Agent",
			TaskName:  "Backend Development",
			StartedAt: workflow.Now(ctx).Format(time.RFC3339),
		})
		state.Agents = append(state.Agents, ActiveAgent{
			TaskID:    "database",
			AgentID:   "dev-db",
			AgentName: "Database Agent",
			TaskName:  "Schema Migrations",
			StartedAt: workflow.Now(ctx).Format(time.RFC3339),
		})

		backendFuture := workflow.ExecuteActivity(ctxLong, "BackendGenerate", backendInput)
		selector.AddFuture(backendFuture, func(f workflow.Future) {
			var result CodeBundle
			backendErr = f.Get(ctx, &result)
			if backendErr == nil && result.Files != nil {
				state.BackendCode = result.Files
				for k, v := range result.Files {
					state.AllCodeFiles["backend/"+k] = v
				}
			}
			state.Agents = filterAgents(state.Agents, "backend")
		})

		// Database migrations
		orm := "prisma" // default ORM
		if backendTech == "python" {
			orm = "sqlalchemy"
		} else if backendTech == "go" {
			orm = "gorm"
		}
		dbMigInput := map[string]interface{}{
			"schema":        state.DatabaseSchema,
			"database_type": databaseTech,
			"orm":           orm,
		}

		dbFuture := workflow.ExecuteActivity(ctxMedium, "DatabaseGenerateMigrations", dbMigInput)
		selector.AddFuture(dbFuture, func(f workflow.Future) {
			var result DatabaseBundle
			databaseErr = f.Get(ctx, &result)
			if databaseErr == nil && result.Files != nil {
				state.DatabaseCode = result.Files
				for k, v := range result.Files {
					state.AllCodeFiles["database/"+k] = v
				}
			}
			state.Agents = filterAgents(state.Agents, "database")
		})
	}

	// Wait for all development activities
	for i := 0; i < 1+boolToInt(config.EnableFullStack)*2; i++ {
		selector.Select(ctx)
	}

	// Log any errors
	if frontendErr != nil {
		logger.Error("Frontend development failed", "Error", frontendErr)
		state.Errors = append(state.Errors, fmt.Sprintf("Frontend: %v", frontendErr))
	}
	if backendErr != nil {
		logger.Error("Backend development failed", "Error", backendErr)
		state.Errors = append(state.Errors, fmt.Sprintf("Backend: %v", backendErr))
	}
	if databaseErr != nil {
		logger.Error("Database development failed", "Error", databaseErr)
		state.Errors = append(state.Errors, fmt.Sprintf("Database: %v", databaseErr))
	}

	// === VALIDATION LOOP ===
	validationFeedback := ""
	for attempt := 1; attempt <= config.MaxCodeAttempts; attempt++ {
		if validationFeedback != "" {
			// Refine code based on feedback
			refineInput := map[string]interface{}{
				"current_code":        state.AllCodeFiles,
				"chat_history":        state.ChatHistory,
				"validation_feedback": validationFeedback,
			}

			var refinedCode map[string]string
			if err := workflow.ExecuteActivity(ctxLong, "DevAgentRefine", refineInput).Get(ctx, &refinedCode); err != nil {
				logger.Error("Code refinement failed", "Error", err)
			} else {
				for k, v := range refinedCode {
					state.AllCodeFiles[k] = v
				}
			}
		}

		// Run validation
		validationInput := map[string]interface{}{
			"code_files": state.AllCodeFiles,
		}

		var validationResult map[string]interface{}
		if err := workflow.ExecuteActivity(ctxMedium, "ValidateCode", validationInput).Get(ctx, &validationResult); err != nil {
			logger.Warn("Validation failed", "Error", err)
			break // Skip validation on error
		}

		state.ValidationResults = validationResult

		if success, ok := validationResult["success"].(bool); ok && success {
			logger.Info("Code validation passed", "Attempt", attempt)
			break
		}

		if feedback, ok := validationResult["feedback"].(string); ok {
			validationFeedback = feedback
		}

		if attempt == config.MaxCodeAttempts {
			logger.Warn("Validation failed after max attempts - proceeding to testing")
		}
	}

	transitionPhaseWithDB(ctx, state, PhaseTesting)

	// ============================================================
	// PHASE 6: TESTING (Full Test Pyramid)
	// ============================================================
	logger.Info("Starting TESTING phase")

	// Generate unit tests (frontend)
	unitTestInput := map[string]interface{}{
		"spec":       state.Requirements,
		"code_files": state.FrontendCode,
	}

	var unitTests map[string]string
	if err := workflow.ExecuteActivity(ctxMedium, "QAAgentGenerateTests", unitTestInput).Get(ctx, &unitTests); err != nil {
		logger.Warn("Unit test generation failed", "Error", err)
	} else {
		state.UnitTests = unitTests
		for k, v := range unitTests {
			state.AllCodeFiles[k] = v
		}
	}

	// Generate backend tests (if full stack)
	if config.EnableFullStack && len(state.BackendCode) > 0 {
		backendTestInput := map[string]interface{}{
			"code_files":      state.BackendCode,
			"backend_type":    backendTech,
			"database_schema": state.DatabaseSchema,
			"api_spec":        state.APISpec,
		}

		var backendTests map[string]string
		if err := workflow.ExecuteActivity(ctxMedium, "QAGenerateBackendTests", backendTestInput).Get(ctx, &backendTests); err != nil {
			logger.Warn("Backend test generation failed", "Error", err)
		} else {
			for k, v := range backendTests {
				state.AllCodeFiles["backend/"+k] = v
			}
		}
	}

	// Generate integration tests
	integrationInput := map[string]interface{}{
		"stack":         backendTech,
		"framework":     getString(state.TechStack, "backend_framework", ""),
		"database_type": databaseTech,
		"api_spec":      state.APISpec,
		"code_files":    state.BackendCode,
	}

	var integrationBundle activities.TestBundle
	if err := workflow.ExecuteActivity(ctxMedium, "QAGenerateIntegrationTests", integrationInput).Get(ctx, &integrationBundle); err != nil {
		logger.Warn("Integration test generation failed", "Error", err)
	} else {
		state.IntegrationTests = integrationBundle.Files
		for k, v := range integrationBundle.Files {
			state.AllCodeFiles["tests/integration/"+k] = v
		}
	}

	// Generate E2E tests
	e2eInput := map[string]interface{}{
		"framework":  "playwright",
		"base_url":   "http://localhost:3000",
		"code_files": state.FrontendCode,
		"user_flows": []string{"Landing page renders", "Core functionality works"},
		"api_spec":   state.APISpec,
	}

	var e2eBundle activities.TestBundle
	if err := workflow.ExecuteActivity(ctxMedium, "QAGenerateE2ETests", e2eInput).Get(ctx, &e2eBundle); err != nil {
		logger.Warn("E2E test generation failed", "Error", err)
	} else {
		state.E2ETests = e2eBundle.Files
		for k, v := range e2eBundle.Files {
			state.AllCodeFiles["tests/e2e/"+k] = v
		}
	}

	// Run tests with Three-Strike Rule
	testsPassed := false
	for attempt := 1; attempt <= config.MaxTestAttempts && !testsPassed; attempt++ {
		logger.Info("Running tests", "Attempt", attempt)

		testRunInput := map[string]interface{}{
			"code_files": state.AllCodeFiles,
		}

		var testResult map[string]interface{}
		if err := workflow.ExecuteActivity(ctxLong, "RunTests", testRunInput).Get(ctx, &testResult); err != nil {
			logger.Error("Test runner failed", "Error", err)
			continue
		}

		state.TestResults = testResult

		if success, ok := testResult["success"].(bool); ok && success {
			testsPassed = true
			logger.Info("Tests passed!", "Attempt", attempt)
		} else if attempt < config.MaxTestAttempts {
			// Regenerate tests with feedback
			testOutput := ""
			if output, ok := testResult["output"].(string); ok {
				testOutput = output
			}

			regenerateInput := map[string]interface{}{
				"spec":        state.Requirements,
				"code_files":  filterNonTestFiles(state.AllCodeFiles),
				"test_files":  filterTestFiles(state.AllCodeFiles),
				"test_output": testOutput,
				"attempt":     attempt + 1,
			}

			var regeneratedTests map[string]string
			if err := workflow.ExecuteActivity(ctxMedium, "QAAgentRegenerateTests", regenerateInput).Get(ctx, &regeneratedTests); err != nil {
				logger.Warn("Test regeneration failed", "Error", err)
			} else {
				// Replace test files
				for k := range filterTestFiles(state.AllCodeFiles) {
					delete(state.AllCodeFiles, k)
				}
				for k, v := range regeneratedTests {
					state.AllCodeFiles[k] = v
				}
			}
		}
	}

	// Analyze test coverage
	coverageInput := map[string]interface{}{
		"code_files":        filterNonTestFiles(state.AllCodeFiles),
		"test_files":        filterTestFiles(state.AllCodeFiles),
		"integration_tests": state.IntegrationTests,
		"e2e_tests":         state.E2ETests,
	}

	if err := workflow.ExecuteActivity(ctxMedium, "QAAnalyzeTestCoverage", coverageInput).Get(ctx, &state.TestCoverage); err != nil {
		logger.Warn("Coverage analysis failed", "Error", err)
	}

	// Code Critic Review
	criticInput := map[string]interface{}{
		"code_files": state.AllCodeFiles,
		"spec":       state.Requirements,
	}

	if err := workflow.ExecuteActivity(ctxMedium, "CodeCriticReview", criticInput).Get(ctx, &state.CriticReview); err != nil {
		logger.Warn("Code Critic review failed", "Error", err)
	}

	// ============================================================
	// PHASE 7: DEPLOYMENT (DevOps Artifacts)
	// ============================================================
	if config.EnableDeployment {
		transitionPhaseWithDB(ctx, state, PhaseDeployment)
		logger.Info("Starting DEPLOYMENT phase")

		// Generate Dockerfile
		dockerInput := map[string]interface{}{
			"project_name": state.ProjectName,
			"stack":        backendTech,
			"framework":    backendTech,
			"port":         3000,
			"has_database": config.EnableFullStack,
		}

		var dockerResult DevOpsBundle
		if err := workflow.ExecuteActivity(ctxMedium, "DevOpsGenerateDockerfile", dockerInput).Get(ctx, &dockerResult); err != nil {
			logger.Warn("Dockerfile generation failed", "Error", err)
			state.Errors = append(state.Errors, fmt.Sprintf("Dockerfile: %v", err))
		} else if dockerResult.Files != nil {
			// Extract Dockerfile from files
			for name, content := range dockerResult.Files {
				if strings.Contains(strings.ToLower(name), "dockerfile") {
					state.Dockerfile = content
					break
				}
			}
		}

		// Generate Docker Compose
		composeInput := map[string]interface{}{
			"project_name": state.ProjectName,
			"services":     []string{"api", "web", "db"},
			"tech_stack": map[string]interface{}{
				"backend":  map[string]string{"language": backendTech},
				"frontend": map[string]string{"framework": frontendTech},
				"database": map[string]string{"type": databaseTech},
			},
			"environment": "production",
		}

		var composeResult DevOpsBundle
		if err := workflow.ExecuteActivity(ctxMedium, "DevOpsGenerateDockerCompose", composeInput).Get(ctx, &composeResult); err != nil {
			logger.Warn("Docker Compose generation failed", "Error", err)
			state.Errors = append(state.Errors, fmt.Sprintf("DockerCompose: %v", err))
		} else if composeResult.Files != nil {
			// Extract docker-compose.yml from files
			for name, content := range composeResult.Files {
				if strings.Contains(strings.ToLower(name), "docker-compose") || strings.Contains(strings.ToLower(name), "compose.yml") {
					state.DockerCompose = content
					break
				}
			}
		}

		// Generate CI Pipeline
		ciInput := map[string]interface{}{
			"project_name":  state.ProjectName,
			"platform":      "github",
			"stack":         backendTech,
			"has_tests":     true,
			"has_linting":   true,
			"deploy_target": "kubernetes",
			"environments":  []string{"staging", "production"},
		}

		var ciResult DevOpsBundle
		if err := workflow.ExecuteActivity(ctxMedium, "DevOpsGenerateCIPipeline", ciInput).Get(ctx, &ciResult); err != nil {
			logger.Warn("CI pipeline generation failed", "Error", err)
			state.Errors = append(state.Errors, fmt.Sprintf("CIPipeline: %v", err))
		} else if ciResult.Files != nil {
			// Extract CI pipeline from files
			for name, content := range ciResult.Files {
				if strings.Contains(name, ".github") || strings.Contains(name, "gitlab-ci") || strings.Contains(name, "pipeline") {
					state.CIPipeline = content
					break
				}
			}
		}

		// Generate Kubernetes manifests
		k8sInput := map[string]interface{}{
			"project_name": state.ProjectName,
			"namespace":    strings.ToLower(strings.ReplaceAll(state.ProjectName, " ", "-")),
			"replicas":     3,
			"environment":  "production",
		}

		var k8sResult DevOpsBundle
		if err := workflow.ExecuteActivity(ctxMedium, "DevOpsGenerateK8sManifests", k8sInput).Get(ctx, &k8sResult); err != nil {
			logger.Warn("K8s manifest generation failed", "Error", err)
			state.Errors = append(state.Errors, fmt.Sprintf("K8sManifests: %v", err))
		} else if k8sResult.Files != nil {
			state.KubeManifests = k8sResult.Files
		}

		// Generate Helm chart
		helmInput := map[string]interface{}{
			"project_name": state.ProjectName,
			"environments": []string{"staging", "production"},
		}

		var helmResult DevOpsBundle
		if err := workflow.ExecuteActivity(ctxMedium, "DevOpsGenerateHelmChart", helmInput).Get(ctx, &helmResult); err != nil {
			logger.Warn("Helm chart generation failed", "Error", err)
			state.Errors = append(state.Errors, fmt.Sprintf("HelmChart: %v", err))
		} else if helmResult.Files != nil {
			state.HelmChart = helmResult.Files
		}

		// Generate Terraform
		terraformInput := map[string]interface{}{
			"project_name": state.ProjectName,
			"provider":     getString(state.TechStack, "cloud", "aws"),
			"region":       "us-east-1",
			"environment":  "production",
			"components":   []string{"vpc", "ecs", "rds", "elasticache"},
			"modular":      true,
		}

		var tfResult DevOpsBundle
		if err := workflow.ExecuteActivity(ctxMedium, "InfraGenerateTerraform", terraformInput).Get(ctx, &tfResult); err != nil {
			logger.Warn("Terraform generation failed", "Error", err)
			state.Errors = append(state.Errors, fmt.Sprintf("Terraform: %v", err))
		} else if tfResult.Files != nil {
			state.InfraCode = tfResult.Files
		}
	}

	// ============================================================
	// PHASE 8: OPERATIONS (SRE Artifacts)
	// ============================================================
	if config.EnableSRE {
		transitionPhaseWithDB(ctx, state, PhaseOperations)
		logger.Info("Starting OPERATIONS phase")

		// Generate monitoring config
		monitorInput := map[string]interface{}{
			"app_name": state.ProjectName,
			"metrics":  []string{"http_requests", "latency", "errors", "cpu", "memory"},
			"platform": "prometheus",
		}

		if err := workflow.ExecuteActivity(ctxMedium, "SREGenerateMonitoring", monitorInput).Get(ctx, &state.MonitoringConfig); err != nil {
			logger.Warn("Monitoring config generation failed", "Error", err)
		}

		// Generate alert rules
		alertInput := map[string]interface{}{
			"app_name": state.ProjectName,
			"slos": []map[string]interface{}{
				{"name": "availability", "target": 99.9},
				{"name": "latency_p99", "target_ms": 500},
			},
			"platform": "prometheus",
		}

		if err := workflow.ExecuteActivity(ctxMedium, "SREGenerateAlerts", alertInput).Get(ctx, &state.AlertRules); err != nil {
			logger.Warn("Alert rules generation failed", "Error", err)
		}

		// Generate dashboards
		dashInput := map[string]interface{}{
			"app_name":       state.ProjectName,
			"dashboard_type": "overview",
			"platform":       "grafana",
		}

		if err := workflow.ExecuteActivity(ctxMedium, "SREGenerateDashboards", dashInput).Get(ctx, &state.Dashboards); err != nil {
			logger.Warn("Dashboard generation failed", "Error", err)
		}

		// Generate runbooks
		runbookInput := map[string]interface{}{
			"services":  []string{state.ProjectName + "-api", state.ProjectName + "-frontend"},
			"scenarios": []string{"high_latency", "high_error_rate", "database_connection", "out_of_memory"},
		}

		var runbookResult map[string]interface{}
		if err := workflow.ExecuteActivity(ctxMedium, "SREGenerateRunbooks", runbookInput).Get(ctx, &runbookResult); err != nil {
			logger.Warn("Runbook generation failed", "Error", err)
		} else if runbooks, ok := runbookResult["runbooks"].(map[string]string); ok {
			state.Runbooks = runbooks
		}
	}

	transitionPhaseWithDB(ctx, state, PhaseHandoff)

	// ============================================================
	// PHASE 9: HANDOFF (Review & Deliver)
	// ============================================================
	logger.Info("Starting HANDOFF phase")
	transitionPhaseWithDB(ctx, state, PhaseReview)

	// Wait for final approval
	for {
		var signal UserMessageSignal
		msgChan.Receive(ctx, &signal)

		if signal.Message == "/approve" {
			break
		}

		if signal.Message == "/reject" {
			transitionPhaseWithDB(ctx, state, PhaseFailed)
			state.Errors = append(state.Errors, "Project rejected by user")
			return state, nil
		}

		// Additional feedback
		state.ChatHistory += "\nUser (Feedback): " + signal.Message
		state.UpdatedAt = workflow.Now(ctx)
	}

	// Mark complete
	now := workflow.Now(ctx)
	state.CompletedAt = &now
	transitionPhaseWithDB(ctx, state, PhaseComplete)

	logger.Info("Consultancy workflow completed successfully")
	return state, nil
}

// Helper functions

func transitionPhase(state *ConsultancyState, phase ConsultancyPhase) {
	state.Phase = phase
	state.PhaseHistory = append(state.PhaseHistory, string(phase))
	state.UpdatedAt = time.Now()

	// Update DAG if present
	if state.DAG != nil {
		currentTaskID := strings.ToLower(string(phase))
		found := false
		state.DAG.Completed = 0
		state.DAG.Running = 0
		state.DAG.Pending = 0

		for i := range state.DAG.Tasks {
			t := &state.DAG.Tasks[i]
			if t.ID == currentTaskID {
				t.Status = "RUNNING"
				state.DAG.Running++
				found = true
			} else if !found {
				t.Status = "COMPLETED"
				state.DAG.Completed++
			} else {
				t.Status = "PENDING"
				state.DAG.Pending++
			}
		}
	}
}

// transitionPhaseWithDB transitions the phase and updates the database
func transitionPhaseWithDB(ctx workflow.Context, state *ConsultancyState, phase ConsultancyPhase) {
	transitionPhase(state, phase)

	// Update database - wait for activity to complete (quick operation)
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
	}
	actCtx := workflow.WithActivityOptions(ctx, ao)

	// Execute and wait for result (database update is fast)
	// Use the explicit UpdatePhaseParams matching the activity definition
	params := activities.UpdatePhaseParams{
		ProjectID:  state.ProjectID,
		WorkflowID: state.WorkflowID,
		Phase:      string(phase),
	}

	err := workflow.ExecuteActivity(actCtx, "ProjectUpdatePhase", params).Get(ctx, nil)
	if err != nil {
		// Log but don't fail the workflow for DB update issues
		workflow.GetLogger(ctx).Warn("Failed to update project phase in database", "error", err)
	}
}

func getString(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return defaultVal
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func filterTestFiles(files map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range files {
		if isTestFile(k) {
			result[k] = v
		}
	}
	return result
}

func filterNonTestFiles(files map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range files {
		if !isTestFile(k) {
			result[k] = v
		}
	}
	return result
}

func filterAgents(agents []ActiveAgent, taskID string) []ActiveAgent {
	result := make([]ActiveAgent, 0)
	for _, a := range agents {
		if a.TaskID != taskID {
			result = append(result, a)
		}
	}
	return result
}
