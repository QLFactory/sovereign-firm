package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"

	"github.com/qlfactory/sovereign-firm/pkg/firm/workflows"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	temporalHost := os.Getenv("TEMPORAL_HOST")
	if temporalHost == "" {
		temporalHost = "localhost:7233"
	}

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		log.Fatalf("Unable to create Temporal client: %v", err)
	}
	defer c.Close()

	// Define test project configuration
	config := workflows.ConsultancyConfig{
		ProjectName:       "TaskFlow - Project Management App",
		ClientID:          "test-client-001",
		InitialMessage:    `Build a modern project management application called "TaskFlow" with the following features:
1. User authentication (signup, login, password reset)
2. Project creation and management
3. Task boards with drag-and-drop (Kanban style)
4. Team collaboration with comments and mentions
5. Due date tracking with notifications
6. Dashboard with progress analytics
7. REST API for mobile app integration

Target users: Small to medium development teams
Expected scale: Up to 1000 concurrent users initially`,
		EnableFullStack:   true,
		EnableDeployment:  true,
		EnableSRE:         true,
		PreferredFrontend: "react",
		PreferredBackend:  "nodejs",
		PreferredDatabase: "postgresql",
		PreferredCloud:    "aws",
		MaxCodeAttempts:   3,
		MaxTestAttempts:   3,
	}

	// Start the workflow
	workflowID := fmt.Sprintf("consultancy-%s-%d", config.ClientID, time.Now().Unix())

	log.Printf("🚀 Starting Consultancy Workflow")
	log.Printf("   Project: %s", config.ProjectName)
	log.Printf("   Workflow ID: %s", workflowID)
	log.Printf("   Full Stack: %v", config.EnableFullStack)
	log.Printf("   Deployment: %v", config.EnableDeployment)
	log.Printf("   SRE: %v", config.EnableSRE)
	log.Println()

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "sovereign-firm-tasks",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, workflows.ConsultancyWorkflow, config)
	if err != nil {
		log.Fatalf("Unable to start workflow: %v", err)
	}

	log.Printf("✅ Workflow started successfully")
	log.Printf("   Run ID: %s", we.GetRunID())
	log.Println()

	// Monitor workflow progress
	log.Println("📊 Monitoring workflow progress...")
	log.Println("   (Press Ctrl+C to stop monitoring - workflow will continue)")
	log.Println()

	// Poll for workflow status
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	lastPhase := ""

	for {
		select {
		case <-ticker.C:
			// Query workflow state
			resp, err := c.QueryWorkflow(context.Background(), workflowID, "", "get_state")
			if err != nil {
				// Workflow might have completed
				var result workflows.ConsultancyState
				if err := we.Get(context.Background(), &result); err != nil {
					log.Printf("⚠️  Query failed: %v", err)
					continue
				}
				// Workflow completed
				printFinalResult(&result, startTime)
				return
			}

			var state workflows.ConsultancyState
			if err := resp.Get(&state); err != nil {
				log.Printf("⚠️  Failed to decode state: %v", err)
				continue
			}

			// Print phase transitions
			if string(state.Phase) != lastPhase {
				elapsed := time.Since(startTime).Round(time.Second)
				log.Printf("[%s] Phase: %s → %s", elapsed, lastPhase, state.Phase)
				lastPhase = string(state.Phase)

				// Print phase-specific details
				printPhaseDetails(&state)
			}

			// Check if complete
			if state.Phase == workflows.PhaseComplete || state.Phase == workflows.PhaseFailed {
				printFinalResult(&state, startTime)
				return
			}
		}
	}
}

func printPhaseDetails(state *workflows.ConsultancyState) {
	switch state.Phase {
	case workflows.PhaseSizing:
		if state.ComplexityScore != nil {
			log.Printf("   📏 Complexity score calculated")
		}
		if state.EffortEstimate != nil {
			log.Printf("   ⏱️  Effort estimate calculated")
		}

	case workflows.PhaseArchitecture:
		if state.TechStack != nil {
			log.Printf("   🏗️  Tech Stack Selected:")
			if ts, err := json.MarshalIndent(state.TechStack, "      ", "  "); err == nil {
				log.Printf("      %s", string(ts))
			}
		}

	case workflows.PhaseDevelopment:
		log.Printf("   💻 Generating code...")
		if len(state.FrontendCode) > 0 {
			log.Printf("      Frontend files: %d", len(state.FrontendCode))
		}
		if len(state.BackendCode) > 0 {
			log.Printf("      Backend files: %d", len(state.BackendCode))
		}
		if len(state.DatabaseCode) > 0 {
			log.Printf("      Database files: %d", len(state.DatabaseCode))
		}

	case workflows.PhaseTesting:
		log.Printf("   🧪 Running tests...")
		if len(state.UnitTests) > 0 {
			log.Printf("      Unit tests: %d files", len(state.UnitTests))
		}
		if len(state.IntegrationTests) > 0 {
			log.Printf("      Integration tests: %d files", len(state.IntegrationTests))
		}
		if len(state.E2ETests) > 0 {
			log.Printf("      E2E tests: %d files", len(state.E2ETests))
		}

	case workflows.PhaseDeployment:
		log.Printf("   🚀 Generating deployment artifacts...")
		if state.Dockerfile != "" {
			log.Printf("      ✓ Dockerfile")
		}
		if state.DockerCompose != "" {
			log.Printf("      ✓ Docker Compose")
		}
		if len(state.KubeManifests) > 0 {
			log.Printf("      ✓ Kubernetes manifests: %d", len(state.KubeManifests))
		}
		if len(state.HelmChart) > 0 {
			log.Printf("      ✓ Helm chart")
		}
		if state.CIPipeline != "" {
			log.Printf("      ✓ CI/CD pipeline")
		}
		if len(state.InfraCode) > 0 {
			log.Printf("      ✓ Infrastructure code: %d files", len(state.InfraCode))
		}

	case workflows.PhaseOperations:
		log.Printf("   📈 Setting up observability...")
		if state.MonitoringConfig != nil {
			log.Printf("      ✓ Monitoring config")
		}
		if state.AlertRules != nil {
			log.Printf("      ✓ Alert rules")
		}
		if state.Dashboards != nil {
			log.Printf("      ✓ Dashboards")
		}
		if len(state.Runbooks) > 0 {
			log.Printf("      ✓ Runbooks: %d", len(state.Runbooks))
		}
	}
}

func printFinalResult(state *workflows.ConsultancyState, startTime time.Time) {
	elapsed := time.Since(startTime).Round(time.Second)

	log.Println()
	log.Println("═══════════════════════════════════════════════════════════════")
	if state.Phase == workflows.PhaseComplete {
		log.Printf("✅ WORKFLOW COMPLETED SUCCESSFULLY")
	} else {
		log.Printf("❌ WORKFLOW FAILED")
	}
	log.Println("═══════════════════════════════════════════════════════════════")
	log.Println()

	log.Printf("📋 Project: %s", state.ProjectName)
	log.Printf("🆔 Project ID: %s", state.ProjectID)
	log.Printf("⏱️  Total Duration: %s", elapsed)
	log.Println()

	log.Println("📁 Deliverables Summary:")
	log.Println("─────────────────────────────────────────────────────────────────")

	// Code files
	totalCodeFiles := len(state.FrontendCode) + len(state.BackendCode) + len(state.DatabaseCode)
	log.Printf("   Code Files: %d total", totalCodeFiles)
	if len(state.FrontendCode) > 0 {
		log.Printf("      • Frontend: %d files", len(state.FrontendCode))
		for path := range state.FrontendCode {
			log.Printf("        - %s", path)
		}
	}
	if len(state.BackendCode) > 0 {
		log.Printf("      • Backend: %d files", len(state.BackendCode))
		for path := range state.BackendCode {
			log.Printf("        - %s", path)
		}
	}
	if len(state.DatabaseCode) > 0 {
		log.Printf("      • Database: %d files", len(state.DatabaseCode))
		for path := range state.DatabaseCode {
			log.Printf("        - %s", path)
		}
	}
	log.Println()

	// Tests
	totalTests := len(state.UnitTests) + len(state.IntegrationTests) + len(state.E2ETests)
	log.Printf("   Test Files: %d total", totalTests)
	if len(state.UnitTests) > 0 {
		log.Printf("      • Unit Tests: %d files", len(state.UnitTests))
	}
	if len(state.IntegrationTests) > 0 {
		log.Printf("      • Integration Tests: %d files", len(state.IntegrationTests))
	}
	if len(state.E2ETests) > 0 {
		log.Printf("      • E2E Tests: %d files", len(state.E2ETests))
	}
	log.Println()

	// DevOps
	log.Println("   DevOps Artifacts:")
	if state.Dockerfile != "" {
		log.Println("      ✓ Dockerfile")
	}
	if state.DockerCompose != "" {
		log.Println("      ✓ Docker Compose")
	}
	if len(state.KubeManifests) > 0 {
		log.Printf("      ✓ Kubernetes Manifests (%d files)", len(state.KubeManifests))
	}
	if len(state.HelmChart) > 0 {
		log.Println("      ✓ Helm Chart")
	}
	if state.CIPipeline != "" {
		log.Println("      ✓ CI/CD Pipeline")
	}
	if len(state.InfraCode) > 0 {
		log.Printf("      ✓ Infrastructure Code (%d files)", len(state.InfraCode))
	}
	log.Println()

	// SRE
	log.Println("   SRE Artifacts:")
	if state.MonitoringConfig != nil {
		log.Println("      ✓ Prometheus Monitoring Config")
	}
	if state.AlertRules != nil {
		log.Println("      ✓ Alert Rules")
	}
	if state.Dashboards != nil {
		log.Println("      ✓ Grafana Dashboards")
	}
	if len(state.Runbooks) > 0 {
		log.Printf("      ✓ Runbooks (%d)", len(state.Runbooks))
	}
	log.Println()

	// Phase history
	log.Println("📜 Phase History:")
	for i, phase := range state.PhaseHistory {
		log.Printf("   %d. %s", i+1, phase)
	}
	log.Println()

	// Errors/Warnings
	if len(state.Errors) > 0 {
		log.Println("⚠️  Errors:")
		for _, err := range state.Errors {
			log.Printf("   • %s", err)
		}
		log.Println()
	}

	if len(state.Warnings) > 0 {
		log.Println("⚡ Warnings:")
		for _, warn := range state.Warnings {
			log.Printf("   • %s", warn)
		}
		log.Println()
	}

	log.Println("═══════════════════════════════════════════════════════════════")
}
