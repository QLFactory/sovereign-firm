// Package workflows provides Temporal workflows for the Sovereign Firm.
package workflows

import (
	"fmt"
	"strings"
	"time"

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
	PhaseIntake        ConsultancyPhase = "INTAKE"
	PhaseSizing        ConsultancyPhase = "SIZING"
	PhasePlanning      ConsultancyPhase = "PLANNING"
	PhaseArchitecture  ConsultancyPhase = "ARCHITECTURE"
	PhaseDevelopment   ConsultancyPhase = "DEVELOPMENT"
	PhaseTesting       ConsultancyPhase = "TESTING"
	PhaseDeployment    ConsultancyPhase = "DEPLOYMENT"
	PhaseOperations    ConsultancyPhase = "OPERATIONS"
	PhaseHandoff       ConsultancyPhase = "HANDOFF"
	PhaseComplete      ConsultancyPhase = "COMPLETE"
	PhaseReview        ConsultancyPhase = "REVIEW"
	PhaseFailed        ConsultancyPhase = "FAILED"
)

// ConsultancyConfig contains configuration for the consultancy workflow
type ConsultancyConfig struct {
	// Project settings
	ProjectName      string `json:"project_name"`
	ClientID         string `json:"client_id"`
	InitialMessage   string `json:"initial_message"`    // Initial project description (skips intake chat)
	EnableFullStack  bool   `json:"enable_full_stack"`  // Full stack vs frontend-only
	EnableDeployment bool   `json:"enable_deployment"`  // Generate deployment configs
	EnableSRE        bool   `json:"enable_sre"`         // Generate monitoring/alerting

	// Tech preferences (if empty, architect will decide)
	PreferredFrontend string `json:"preferred_frontend"` // React, Vue, Angular
	PreferredBackend  string `json:"preferred_backend"`  // Node.js, Python, Go
	PreferredDatabase string `json:"preferred_database"` // PostgreSQL, MongoDB
	PreferredCloud    string `json:"preferred_cloud"`    // AWS, Azure, GCP

	// Limits
	MaxCodeAttempts int `json:"max_code_attempts"` // Default 3
	MaxTestAttempts int `json:"max_test_attempts"` // Default 3
}

// ConsultancyState holds the complete state of a consultancy project
type ConsultancyState struct {
	// Project identity
	ProjectID   string           `json:"project_id"`
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
	FrontendCode  map[string]string `json:"frontend_code"`
	BackendCode   map[string]string `json:"backend_code"`
	DatabaseCode  map[string]string `json:"database_code"` // Migrations, schemas
	AllCodeFiles  map[string]string `json:"all_code_files"`

	// Testing artifacts
	UnitTests        map[string]string `json:"unit_tests"`
	IntegrationTests map[string]string `json:"integration_tests"`
	E2ETests         map[string]string `json:"e2e_tests"`
	TestCoverage     map[string]interface{} `json:"test_coverage"`

	// Deployment artifacts
	Dockerfile      string            `json:"dockerfile"`
	DockerCompose   string            `json:"docker_compose"`
	KubeManifests   map[string]string `json:"kube_manifests"`
	HelmChart       map[string]string `json:"helm_chart"`
	CIPipeline      string            `json:"ci_pipeline"`
	InfraCode       map[string]string `json:"infra_code"` // Terraform, etc.

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
	Errors        []string `json:"errors"`
	Warnings      []string `json:"warnings"`
	PhaseHistory  []string `json:"phase_history"`
	CompletedAt   *time.Time `json:"completed_at"`
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
		ProjectID:    workflow.GetInfo(ctx).WorkflowExecution.ID,
		ProjectName:  config.ProjectName,
		ClientID:     config.ClientID,
		Phase:        PhaseIntake,
		StartedAt:    workflow.Now(ctx),
		UpdatedAt:    workflow.Now(ctx),
		AllCodeFiles: make(map[string]string),
		PhaseHistory: []string{string(PhaseIntake)},
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

	// Register query handler
	err := workflow.SetQueryHandler(ctx, "get_state", func() (*ConsultancyState, error) {
		return state, nil
	})
	if err != nil {
		return nil, err
	}

	// Signal channel for user messages
	msgChan := workflow.GetSignalChannel(ctx, "USER_MESSAGE")

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
		err := workflow.ExecuteActivity(ctxShort, "PMAgentChat", state.ChatHistory).Get(ctx, &reply)
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
			err := workflow.ExecuteActivity(ctxShort, "PMAgentChat", state.ChatHistory).Get(ctx, &reply)
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

	transitionPhase(state, PhaseSizing)

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

	transitionPhase(state, PhasePlanning)

	// ============================================================
	// PHASE 3: PLANNING (Project Plan & Proposal)
	// ============================================================
	logger.Info("Starting PLANNING phase")

	// Create project plan
	planInput := map[string]interface{}{
		"description":    state.ChatHistory,
		"effort":         state.EffortEstimate,
		"methodology":    "agile",
		"sprint_length":  14,
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

	transitionPhase(state, PhaseArchitecture)

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
		"requirements":      state.Requirements,
		"system_design":     state.SystemDesign,
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

	transitionPhase(state, PhaseDevelopment)

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

	frontendFuture := workflow.ExecuteActivity(ctxLong, "DevAgentGenerate", frontendSpec)
	selector.AddFuture(frontendFuture, func(f workflow.Future) {
		var result map[string]string
		frontendErr = f.Get(ctx, &result)
		if frontendErr == nil {
			state.FrontendCode = result
			for k, v := range result {
				state.AllCodeFiles[k] = v
			}
		}
	})

	// Backend development (if full stack)
	if config.EnableFullStack {
		backendInput := map[string]interface{}{
			"project_name":    config.ProjectName,
			"requirements":    state.ChatHistory,
			"tech_stack": map[string]interface{}{
				"backend":  backendTech,
				"database": databaseTech,
			},
			"database_schema": state.DatabaseSchema,
			"api_spec":        state.APISpec,
		}

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
		})
	}

	// Wait for all development activities
	for i := 0; i < 1 + boolToInt(config.EnableFullStack)*2; i++ {
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

	transitionPhase(state, PhaseTesting)

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
		"services":        []string{"api", "database"},
		"backend_type":    backendTech,
		"database_type":   databaseTech,
		"api_spec":        state.APISpec,
	}

	var integrationTests map[string]string
	if err := workflow.ExecuteActivity(ctxMedium, "QAGenerateIntegrationTests", integrationInput).Get(ctx, &integrationTests); err != nil {
		logger.Warn("Integration test generation failed", "Error", err)
	} else {
		state.IntegrationTests = integrationTests
	}

	// Generate E2E tests
	e2eInput := map[string]interface{}{
		"requirements":    state.Requirements,
		"frontend_type":   frontendTech,
		"api_spec":        state.APISpec,
		"test_framework":  "playwright",
	}

	var e2eTests map[string]string
	if err := workflow.ExecuteActivity(ctxMedium, "QAGenerateE2ETests", e2eInput).Get(ctx, &e2eTests); err != nil {
		logger.Warn("E2E test generation failed", "Error", err)
	} else {
		state.E2ETests = e2eTests
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
		"code_files":         filterNonTestFiles(state.AllCodeFiles),
		"test_files":         filterTestFiles(state.AllCodeFiles),
		"integration_tests":  state.IntegrationTests,
		"e2e_tests":          state.E2ETests,
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
		transitionPhase(state, PhaseDeployment)
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
		transitionPhase(state, PhaseOperations)
		logger.Info("Starting OPERATIONS phase")

		// Generate monitoring config
		monitorInput := map[string]interface{}{
			"app_name":       state.ProjectName,
			"metrics":        []string{"http_requests", "latency", "errors", "cpu", "memory"},
			"platform":       "prometheus",
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
			"app_name":   state.ProjectName,
			"dashboard_type": "overview",
			"platform":   "grafana",
		}

		if err := workflow.ExecuteActivity(ctxMedium, "SREGenerateDashboards", dashInput).Get(ctx, &state.Dashboards); err != nil {
			logger.Warn("Dashboard generation failed", "Error", err)
		}

		// Generate runbooks
		runbookInput := map[string]interface{}{
			"services": []string{state.ProjectName + "-api", state.ProjectName + "-frontend"},
			"scenarios": []string{"high_latency", "high_error_rate", "database_connection", "out_of_memory"},
		}

		var runbookResult map[string]interface{}
		if err := workflow.ExecuteActivity(ctxMedium, "SREGenerateRunbooks", runbookInput).Get(ctx, &runbookResult); err != nil {
			logger.Warn("Runbook generation failed", "Error", err)
		} else if runbooks, ok := runbookResult["runbooks"].(map[string]string); ok {
			state.Runbooks = runbooks
		}
	}

	transitionPhase(state, PhaseHandoff)

	// ============================================================
	// PHASE 9: HANDOFF (Review & Deliver)
	// ============================================================
	logger.Info("Starting HANDOFF phase")
	state.Phase = PhaseReview

	// Wait for final approval
	for {
		var signal UserMessageSignal
		msgChan.Receive(ctx, &signal)

		if signal.Message == "/approve" {
			break
		}

		if signal.Message == "/reject" {
			state.Phase = PhaseFailed
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
	state.Phase = PhaseComplete
	state.PhaseHistory = append(state.PhaseHistory, string(PhaseComplete))

	logger.Info("Consultancy workflow completed successfully")
	return state, nil
}

// Helper functions

func transitionPhase(state *ConsultancyState, phase ConsultancyPhase) {
	state.Phase = phase
	state.PhaseHistory = append(state.PhaseHistory, string(phase))
	state.UpdatedAt = time.Now()
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
