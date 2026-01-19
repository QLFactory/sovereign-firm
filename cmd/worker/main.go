package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/joho/godotenv"
	"github.com/qlfactory/sovereign-firm/pkg/agent" // Added
	"github.com/qlfactory/sovereign-firm/pkg/firm/activities"
	"github.com/qlfactory/sovereign-firm/pkg/firm/workflows"
	"github.com/qlfactory/sovereign-firm/pkg/sandbox"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

func main() {
	// Load .env file
	var err error
	if err = godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Verify LLM Connectivity (Fail Fast)
	llmClient := llm.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := llmClient.Ping(ctx); err != nil {
		log.Fatalf("CRITICAL: Cannot connect to LLM: %v", err)
	}
	log.Println("✅ Connected to LLM")

	// Initialize database connection (optional - worker still works without it)
	var dbPool *pgxpool.Pool
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://sovereign:sovereign123@localhost:5432/sovereign_firm?sslmode=disable"
	}
	dbPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Printf("Warning: Database not available: %v (phase updates will be skipped)", err)
		dbPool = nil
	} else {
		if err := dbPool.Ping(ctx); err != nil {
			log.Printf("Warning: Database ping failed: %v (phase updates will be skipped)", err)
			dbPool.Close()
			dbPool = nil
		} else {
			log.Println("✅ Connected to Database")
			defer dbPool.Close()
		}
	}

	temporalHost := os.Getenv("TEMPORAL_HOST")
	if temporalHost == "" {
		temporalHost = "localhost:7233"
	}

	c, err := client.Dial(client.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "sovereign-firm-tasks", worker.Options{})

	// Register Workflows
	w.RegisterWorkflow(workflows.ProjectLifecycle)
	w.RegisterWorkflow(workflows.ConsultancyWorkflow)

	// Initialize Sandbox Manager
	sandboxMgr := sandbox.NewManager()
	defer sandboxMgr.DestroyAll(ctx)

	// Initialize Agent System
	skillsDir := os.Getenv("SKILLS_DIR")
	if skillsDir == "" {
		skillsDir = "skills"
	}
	skillRegistry := agent.NewSkillRegistry(skillsDir)
	if err := skillRegistry.LoadSkills(); err != nil {
		log.Printf("Warning: Failed to load skills from %s: %v", skillsDir, err)
	} else {
		log.Printf("✅ Loaded %d skills from %s", skillRegistry.Count(), skillsDir)
	}

	agentPool := agent.NewAgentPool(agent.DefaultPoolConfig(), skillRegistry, llmClient)
	defer agentPool.Shutdown()

	// Register Activities
	pmAgent := activities.NewPMAgent(agentPool)
	w.RegisterActivityWithOptions(pmAgent.Chat, activity.RegisterOptions{Name: "PMAgentChat"})

	indexer := activities.NewIndexer(agentPool)
	w.RegisterActivityWithOptions(indexer.IndexRepository, activity.RegisterOptions{Name: "IndexRepo"})

	devAgent := activities.NewDevAgent(agentPool)
	w.RegisterActivityWithOptions(devAgent.GenerateCode, activity.RegisterOptions{Name: "DevAgentGenerate"})
	w.RegisterActivityWithOptions(devAgent.RefineCode, activity.RegisterOptions{Name: "DevAgentRefine"})

	qaAgent := activities.NewQAAgent(agentPool)
	w.RegisterActivityWithOptions(qaAgent.GenerateTests, activity.RegisterOptions{Name: "QAAgentGenerateTests"})
	w.RegisterActivityWithOptions(qaAgent.RegenerateTests, activity.RegisterOptions{Name: "QAAgentRegenerateTests"})

	testRunner := activities.NewTestRunner(sandboxMgr)
	w.RegisterActivityWithOptions(testRunner.RunTests, activity.RegisterOptions{Name: "RunTests"})

	validator := activities.NewValidator(sandboxMgr)
	w.RegisterActivityWithOptions(validator.ValidateCode, activity.RegisterOptions{Name: "ValidateCode"})

	critic := activities.NewCodeCritic()
	w.RegisterActivityWithOptions(critic.ReviewCode, activity.RegisterOptions{Name: "CodeCriticReview"})

	// Architect Agent - System Design, Tech Stack, Database Schema, API Design
	architect := activities.NewArchitectAgent()
	w.RegisterActivityWithOptions(architect.AnalyzeRequirements, activity.RegisterOptions{Name: "ArchitectAnalyzeRequirements"})
	w.RegisterActivityWithOptions(architect.SelectTechStack, activity.RegisterOptions{Name: "ArchitectSelectTechStack"})
	w.RegisterActivityWithOptions(architect.DesignDatabase, activity.RegisterOptions{Name: "ArchitectDesignDatabase"})
	w.RegisterActivityWithOptions(architect.DesignAPI, activity.RegisterOptions{Name: "ArchitectDesignAPI"})
	w.RegisterActivityWithOptions(architect.GenerateFullDesign, activity.RegisterOptions{Name: "ArchitectFullDesign"})

	// Backend Agent - Backend Services, APIs, Models (Node.js, Python, Go)
	backend := activities.NewBackendAgent()
	w.RegisterActivityWithOptions(backend.GenerateBackend, activity.RegisterOptions{Name: "BackendGenerate"})
	w.RegisterActivityWithOptions(backend.RefineBackend, activity.RegisterOptions{Name: "BackendRefine"})
	w.RegisterActivityWithOptions(backend.GenerateModels, activity.RegisterOptions{Name: "BackendGenerateModels"})
	w.RegisterActivityWithOptions(backend.GenerateAPIRoutes, activity.RegisterOptions{Name: "BackendGenerateRoutes"})

	// Database Agent - Schema Design, Migrations, Seed Data, Optimization
	database := activities.NewDatabaseAgent()
	w.RegisterActivityWithOptions(database.DesignSchema, activity.RegisterOptions{Name: "DatabaseDesignSchema"})
	w.RegisterActivityWithOptions(database.GenerateMigrations, activity.RegisterOptions{Name: "DatabaseGenerateMigrations"})
	w.RegisterActivityWithOptions(database.GenerateSeedData, activity.RegisterOptions{Name: "DatabaseGenerateSeedData"})
	w.RegisterActivityWithOptions(database.OptimizeSchema, activity.RegisterOptions{Name: "DatabaseOptimizeSchema"})

	// DevOps Agent - Docker, CI/CD, Kubernetes, Helm
	devops := activities.NewDevOpsAgent()
	w.RegisterActivityWithOptions(devops.GenerateDockerfile, activity.RegisterOptions{Name: "DevOpsGenerateDockerfile"})
	w.RegisterActivityWithOptions(devops.GenerateDockerCompose, activity.RegisterOptions{Name: "DevOpsGenerateDockerCompose"})
	w.RegisterActivityWithOptions(devops.GenerateCIPipeline, activity.RegisterOptions{Name: "DevOpsGenerateCIPipeline"})
	w.RegisterActivityWithOptions(devops.GenerateKubernetesManifests, activity.RegisterOptions{Name: "DevOpsGenerateK8sManifests"})
	w.RegisterActivityWithOptions(devops.GenerateHelmChart, activity.RegisterOptions{Name: "DevOpsGenerateHelmChart"})
	w.RegisterActivityWithOptions(devops.RefineDevOps, activity.RegisterOptions{Name: "DevOpsRefine"})

	// Infrastructure Agent - Terraform, CloudFormation, Pulumi, Cost Estimation
	infra := activities.NewInfraAgent()
	w.RegisterActivityWithOptions(infra.GenerateTerraform, activity.RegisterOptions{Name: "InfraGenerateTerraform"})
	w.RegisterActivityWithOptions(infra.GenerateCloudFormation, activity.RegisterOptions{Name: "InfraGenerateCloudFormation"})
	w.RegisterActivityWithOptions(infra.GeneratePulumi, activity.RegisterOptions{Name: "InfraGeneratePulumi"})
	w.RegisterActivityWithOptions(infra.EstimateCost, activity.RegisterOptions{Name: "InfraEstimateCost"})
	w.RegisterActivityWithOptions(infra.RefineInfra, activity.RegisterOptions{Name: "InfraRefine"})

	// Enhanced QA Agent - Full Test Pyramid (Integration, E2E, Performance, Security, Accessibility)
	enhancedQA := activities.NewEnhancedQAAgent()
	w.RegisterActivityWithOptions(enhancedQA.GenerateIntegrationTests, activity.RegisterOptions{Name: "QAGenerateIntegrationTests"})
	w.RegisterActivityWithOptions(enhancedQA.GenerateE2ETests, activity.RegisterOptions{Name: "QAGenerateE2ETests"})
	w.RegisterActivityWithOptions(enhancedQA.GeneratePerformanceTests, activity.RegisterOptions{Name: "QAGeneratePerformanceTests"})
	w.RegisterActivityWithOptions(enhancedQA.GenerateSecurityTests, activity.RegisterOptions{Name: "QAGenerateSecurityTests"})
	w.RegisterActivityWithOptions(enhancedQA.GenerateAccessibilityTests, activity.RegisterOptions{Name: "QAGenerateAccessibilityTests"})
	w.RegisterActivityWithOptions(enhancedQA.GenerateBackendTests, activity.RegisterOptions{Name: "QAGenerateBackendTests"})
	w.RegisterActivityWithOptions(enhancedQA.AnalyzeTestCoverage, activity.RegisterOptions{Name: "QAAnalyzeTestCoverage"})

	// SRE Agent - Monitoring, Alerting, Dashboards, Runbooks, Incident Analysis, On-Call
	sre := activities.NewSREAgent()
	w.RegisterActivityWithOptions(sre.GenerateMonitoringConfig, activity.RegisterOptions{Name: "SREGenerateMonitoring"})
	w.RegisterActivityWithOptions(sre.GenerateAlertRules, activity.RegisterOptions{Name: "SREGenerateAlerts"})
	w.RegisterActivityWithOptions(sre.GenerateDashboards, activity.RegisterOptions{Name: "SREGenerateDashboards"})
	w.RegisterActivityWithOptions(sre.GenerateRunbooks, activity.RegisterOptions{Name: "SREGenerateRunbooks"})
	w.RegisterActivityWithOptions(sre.AnalyzeIncident, activity.RegisterOptions{Name: "SREAnalyzeIncident"})
	w.RegisterActivityWithOptions(sre.GenerateOnCallConfig, activity.RegisterOptions{Name: "SREGenerateOnCall"})
	w.RegisterActivityWithOptions(sre.RefineSRE, activity.RegisterOptions{Name: "SRERefine"})

	// Enhanced PM Agent - Requirements, Sizing, Planning, Tracking, Reporting
	enhancedPM := activities.NewEnhancedPMAgent()
	w.RegisterActivityWithOptions(enhancedPM.GatherRequirements, activity.RegisterOptions{Name: "PMGatherRequirements"})
	w.RegisterActivityWithOptions(enhancedPM.EstimateComplexity, activity.RegisterOptions{Name: "PMEstimateComplexity"})
	w.RegisterActivityWithOptions(enhancedPM.EstimateEffort, activity.RegisterOptions{Name: "PMEstimateEffort"})
	w.RegisterActivityWithOptions(enhancedPM.CreateProjectPlan, activity.RegisterOptions{Name: "PMCreateProjectPlan"})
	w.RegisterActivityWithOptions(enhancedPM.AllocateResources, activity.RegisterOptions{Name: "PMAllocateResources"})
	w.RegisterActivityWithOptions(enhancedPM.TrackProgress, activity.RegisterOptions{Name: "PMTrackProgress"})
	w.RegisterActivityWithOptions(enhancedPM.GenerateStatusReport, activity.RegisterOptions{Name: "PMGenerateStatusReport"})
	w.RegisterActivityWithOptions(enhancedPM.GenerateProposal, activity.RegisterOptions{Name: "PMGenerateProposal"})
	w.RegisterActivityWithOptions(enhancedPM.RefinePM, activity.RegisterOptions{Name: "PMRefine"})

	// Skill Injection System - Dynamic Skill Selection, Agent Spawning, Skill Composition
	skillInjector := activities.NewSkillInjectorWithRegistry(skillRegistry)
	w.RegisterActivityWithOptions(skillInjector.SelectSkills, activity.RegisterOptions{Name: "SkillSelectSkills"})
	w.RegisterActivityWithOptions(skillInjector.SpawnSkilledAgent, activity.RegisterOptions{Name: "SkillSpawnAgent"})
	w.RegisterActivityWithOptions(skillInjector.ComposeSkills, activity.RegisterOptions{Name: "SkillComposeSkills"})
	w.RegisterActivityWithOptions(skillInjector.AnalyzeProjectForSkills, activity.RegisterOptions{Name: "SkillAnalyzeProject"})
	w.RegisterActivityWithOptions(skillInjector.ListSkills, activity.RegisterOptions{Name: "SkillListSkills"})
	w.RegisterActivityWithOptions(skillInjector.GetSkillDependencies, activity.RegisterOptions{Name: "SkillGetDependencies"})
	w.RegisterActivityWithOptions(skillInjector.RegisterSkill, activity.RegisterOptions{Name: "SkillRegister"})

	// Project Agent - Database Operations for Project Tracking
	projectAgent := activities.NewProjectAgent(dbPool)
	w.RegisterActivityWithOptions(projectAgent.UpdatePhase, activity.RegisterOptions{Name: "ProjectUpdatePhase"})

	// Brownfield Analyzer - Codebase Analysis for Existing Projects
	chromaURL := os.Getenv("CHROMA_URL")
	if chromaURL == "" {
		chromaURL = "http://localhost:8000"
	}
	brownfieldAnalyzer := activities.NewBrownfieldAnalyzer(chromaURL)
	w.RegisterActivityWithOptions(brownfieldAnalyzer.AnalyzeProject, activity.RegisterOptions{Name: "BrownfieldAnalyzeProject"})
	w.RegisterActivityWithOptions(brownfieldAnalyzer.QuickAnalyze, activity.RegisterOptions{Name: "BrownfieldQuickAnalyze"})
	log.Println("✅ Registered Brownfield Analyzer activities")

	log.Println("Worker started...")
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
