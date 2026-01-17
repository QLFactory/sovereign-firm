package activities

import (
	"testing"
)

// ============================================================================
// Constructor Tests
// ============================================================================

func TestNewEnhancedPMAgent(t *testing.T) {
	agent := NewEnhancedPMAgent()
	if agent == nil {
		t.Fatal("NewEnhancedPMAgent returned nil")
	}
	if agent.llmClient == nil {
		t.Fatal("LLM client not initialized")
	}
}

// ============================================================================
// Data Structure Tests
// ============================================================================

func TestRequirementsDocumentStructure(t *testing.T) {
	doc := RequirementsDocument{
		ProjectName: "E-Commerce Platform",
		Version:     "1.0",
		Overview:    "Build a modern e-commerce platform",
		Objectives: []string{
			"Enable online product sales",
			"Process payments securely",
		},
		Stakeholders: []Stakeholder{
			{
				Name:        "Product Owner",
				Role:        "Decision maker",
				Influence:   "high",
				Interest:    "high",
				Expectations: "On-time delivery",
			},
		},
		Scope: ProjectScope{
			InScope:      []string{"Product catalog", "Shopping cart"},
			OutOfScope:   []string{"Mobile app"},
			Deliverables: []string{"Web application", "Admin panel"},
		},
		FunctionalReqs: []Requirement{
			{
				ID:          "REQ-001",
				Title:       "User Registration",
				Description: "Users can register with email",
				Priority:    "must",
				Category:    "Authentication",
				AcceptanceCriteria: []string{
					"User can enter email and password",
					"Email verification is sent",
				},
			},
		},
		NonFunctionalReqs: []Requirement{
			{
				ID:          "NFR-001",
				Title:       "Page Load Time",
				Description: "Pages load within 2 seconds",
				Priority:    "must",
				Category:    "Performance",
			},
		},
		Constraints:  []string{"Budget: $50,000"},
		Assumptions:  []string{"Users have modern browsers"},
		Risks: []Risk{
			{
				ID:          "RISK-001",
				Description: "Third-party API unavailable",
				Probability: "low",
				Impact:      "high",
				Mitigation:  "Implement fallback mechanism",
			},
		},
	}

	if doc.ProjectName != "E-Commerce Platform" {
		t.Errorf("Expected project name 'E-Commerce Platform', got %s", doc.ProjectName)
	}
	if len(doc.Objectives) != 2 {
		t.Errorf("Expected 2 objectives, got %d", len(doc.Objectives))
	}
	if len(doc.FunctionalReqs) != 1 {
		t.Errorf("Expected 1 functional requirement, got %d", len(doc.FunctionalReqs))
	}
	if doc.FunctionalReqs[0].Priority != "must" {
		t.Errorf("Expected priority 'must', got %s", doc.FunctionalReqs[0].Priority)
	}
}

func TestComplexityScoreStructure(t *testing.T) {
	score := ComplexityScore{
		Overall: "moderate",
		Score:   6,
		Dimensions: ComplexityDimensions{
			Technical:      7,
			Business:       5,
			Integration:    6,
			DataComplexity: 4,
			TeamSize:       3,
			Timeline:       7,
		},
		Factors: []ComplexityFactor{
			{
				Factor:      "Multiple external integrations",
				Impact:      "increases",
				Severity:    "moderate",
				Description: "Need to integrate with 3 payment providers",
			},
		},
		Recommendations: []string{
			"Consider using API gateway",
			"Plan integration testing early",
		},
		RiskLevel: "medium",
	}

	if score.Overall != "moderate" {
		t.Errorf("Expected overall 'moderate', got %s", score.Overall)
	}
	if score.Score != 6 {
		t.Errorf("Expected score 6, got %d", score.Score)
	}
	if score.Dimensions.Technical != 7 {
		t.Errorf("Expected technical dimension 7, got %d", score.Dimensions.Technical)
	}
	if len(score.Factors) != 1 {
		t.Errorf("Expected 1 factor, got %d", len(score.Factors))
	}
}

func TestEffortEstimateStructure(t *testing.T) {
	estimate := EffortEstimate{
		TotalStoryPoints: 89,
		TotalHours:       534,
		TotalDays:        89,
		Confidence:       "medium",
		Breakdown: []EffortBreakdown{
			{
				Component:   "User Authentication",
				StoryPoints: 13,
				Hours:       78,
				Complexity:  "moderate",
				Notes:       "Includes OAuth integration",
			},
			{
				Component:   "Product Catalog",
				StoryPoints: 21,
				Hours:       126,
				Complexity:  "high",
			},
		},
		TeamRecommendation: TeamRecommendation{
			MinTeamSize:     3,
			OptimalTeamSize: 5,
			Roles:           []string{"Backend Developer", "Frontend Developer", "QA"},
			AgentTypes:      []string{"Backend Agent", "Frontend Agent", "QA Agent"},
		},
		Assumptions: []string{"Team is experienced with the tech stack"},
		Caveats:     []string{"Estimate does not include deployment"},
	}

	if estimate.TotalStoryPoints != 89 {
		t.Errorf("Expected 89 story points, got %d", estimate.TotalStoryPoints)
	}
	if estimate.Confidence != "medium" {
		t.Errorf("Expected confidence 'medium', got %s", estimate.Confidence)
	}
	if len(estimate.Breakdown) != 2 {
		t.Errorf("Expected 2 breakdown items, got %d", len(estimate.Breakdown))
	}
	if estimate.TeamRecommendation.OptimalTeamSize != 5 {
		t.Errorf("Expected optimal team size 5, got %d", estimate.TeamRecommendation.OptimalTeamSize)
	}
}

func TestProjectPlanStructure(t *testing.T) {
	plan := ProjectPlan{
		ProjectName:    "E-Commerce Platform",
		Version:        "1.0",
		StartDate:      "2024-01-15",
		EndDate:        "2024-04-15",
		Methodology:    "agile",
		SprintDuration: 14,
		Phases: []ProjectPhase{
			{
				ID:          "PHASE-001",
				Name:        "Foundation",
				Description: "Set up infrastructure",
				StartDate:   "2024-01-15",
				EndDate:     "2024-02-01",
				Deliverables: []string{"CI/CD Pipeline", "Development environment"},
				Tasks: []Task{
					{
						ID:             "TASK-001",
						Name:           "Set up CI/CD",
						Description:    "Configure GitHub Actions",
						Type:           "devops",
						Priority:       "high",
						StoryPoints:    5,
						EstimatedHours: 30,
						Status:         "todo",
						Dependencies:   []string{},
					},
				},
			},
		},
		Milestones: []Milestone{
			{
				ID:          "MS-001",
				Name:        "MVP Release",
				Date:        "2024-03-01",
				Description: "Minimum viable product ready",
				Criteria:    []string{"Core features complete", "All tests pass"},
				Phase:       "PHASE-002",
			},
		},
		Sprints: []Sprint{
			{
				Number:      1,
				Name:        "Sprint 1",
				StartDate:   "2024-01-15",
				EndDate:     "2024-01-29",
				Goal:        "Foundation setup",
				Capacity:    21,
				PlannedWork: []string{"TASK-001", "TASK-002"},
			},
		},
		CriticalPath: []string{"TASK-001", "TASK-003", "TASK-007"},
	}

	if plan.Methodology != "agile" {
		t.Errorf("Expected methodology 'agile', got %s", plan.Methodology)
	}
	if plan.SprintDuration != 14 {
		t.Errorf("Expected sprint duration 14, got %d", plan.SprintDuration)
	}
	if len(plan.Phases) != 1 {
		t.Errorf("Expected 1 phase, got %d", len(plan.Phases))
	}
	if len(plan.Milestones) != 1 {
		t.Errorf("Expected 1 milestone, got %d", len(plan.Milestones))
	}
	if len(plan.CriticalPath) != 3 {
		t.Errorf("Expected 3 tasks in critical path, got %d", len(plan.CriticalPath))
	}
}

func TestProgressReportStructure(t *testing.T) {
	report := ProgressReport{
		ProjectName:   "E-Commerce Platform",
		ReportDate:    "2024-02-15",
		ReportPeriod:  "Sprint 3",
		OverallStatus: "on_track",
		HealthScore:   85,
		Completion: ProgressMetrics{
			TasksTotal:      50,
			TasksCompleted:  30,
			TasksInProgress: 10,
			TasksBlocked:    2,
			PercentComplete: 60.0,
			PointsTotal:     100,
			PointsCompleted: 60,
		},
		Velocity: VelocityMetrics{
			CurrentSprint:   3,
			PlannedVelocity: 21,
			ActualVelocity:  20,
			AverageVelocity: 19.5,
			Trend:           "stable",
		},
		Blockers: []Blocker{
			{
				ID:          "BLOCK-001",
				Description: "API documentation incomplete",
				Impact:      "Delays integration work",
				Owner:       "John",
				Status:      "in_progress",
			},
		},
		Highlights: []string{"Authentication complete", "Payment integration started"},
		Concerns:   []string{"Third-party API delays"},
		NextSteps:  []string{"Complete payment flow", "Start testing"},
	}

	if report.OverallStatus != "on_track" {
		t.Errorf("Expected status 'on_track', got %s", report.OverallStatus)
	}
	if report.HealthScore != 85 {
		t.Errorf("Expected health score 85, got %d", report.HealthScore)
	}
	if report.Completion.PercentComplete != 60.0 {
		t.Errorf("Expected 60%% complete, got %.1f%%", report.Completion.PercentComplete)
	}
	if report.Velocity.Trend != "stable" {
		t.Errorf("Expected trend 'stable', got %s", report.Velocity.Trend)
	}
}

func TestStatusReportStructure(t *testing.T) {
	report := StatusReport{
		ProjectName:      "E-Commerce Platform",
		ReportDate:       "2024-02-15",
		ExecutiveSummary: "Project is progressing well with minor delays.",
		OverallHealth:    "yellow",
		KeyMetrics: map[string]string{
			"completion": "60%",
			"velocity":   "20 pts/sprint",
		},
		Accomplishments: []string{"Authentication complete"},
		PlannedWork:     []string{"Payment integration"},
		Risks: []RiskUpdate{
			{
				Risk:       "Third-party delays",
				Status:     "ongoing",
				Mitigation: "Added buffer time",
			},
		},
		Decisions: []Decision{
			{
				Topic:       "Payment provider",
				Description: "Choose between Stripe and PayPal",
				Options:     []string{"Stripe", "PayPal", "Both"},
				DueDate:     "2024-02-20",
				Impact:      "High - affects integration timeline",
			},
		},
		Timeline: TimelineUpdate{
			OriginalEndDate: "2024-04-15",
			CurrentEndDate:  "2024-04-22",
			Variance:        "+7 days",
			OnSchedule:      false,
			Explanation:     "Third-party API delays",
		},
	}

	if report.OverallHealth != "yellow" {
		t.Errorf("Expected health 'yellow', got %s", report.OverallHealth)
	}
	if !report.Timeline.OnSchedule {
		// Expected, so no error
	}
	if len(report.Decisions) != 1 {
		t.Errorf("Expected 1 decision, got %d", len(report.Decisions))
	}
}

func TestProposalStructure(t *testing.T) {
	proposal := Proposal{
		Title:            "E-Commerce Platform Development",
		Version:          "1.0",
		Date:             "2024-01-10",
		PreparedFor:      "Acme Corp",
		PreparedBy:       "AI Software Consultancy",
		ExecutiveSummary: "We propose building a modern e-commerce platform.",
		Background:       "Acme Corp needs to expand online presence.",
		Objectives:       []string{"Launch online store", "Increase revenue 20%"},
		Scope: ProposalScope{
			Overview:     "Full-stack e-commerce development",
			Deliverables: []string{"Web application", "Admin panel", "Documentation"},
			InScope:      []string{"Product catalog", "Checkout"},
			OutOfScope:   []string{"Mobile app", "Warehouse integration"},
		},
		Approach: ProposalApproach{
			Methodology:   "Agile Scrum",
			Description:   "2-week sprints with regular demos",
			Phases:        []string{"Discovery", "Development", "Testing", "Launch"},
			KeyActivities: []string{"Sprint planning", "Daily standups", "Retrospectives"},
		},
		Timeline: ProposalTimeline{
			StartDate: "2024-01-15",
			EndDate:   "2024-04-15",
			Duration:  "3 months",
			Milestones: []ProposalMilestone{
				{Name: "MVP", Date: "2024-03-01", Payment: "40%"},
				{Name: "Launch", Date: "2024-04-15", Payment: "40%"},
			},
		},
		Team: ProposalTeam{
			Description: "Cross-functional AI agent team",
			Roles: []TeamRole{
				{Role: "Architect Agent", Allocation: "25%", Description: "System design"},
				{Role: "Backend Agent", Allocation: "50%", Description: "API development"},
			},
		},
		Investment: ProposalCost{
			TotalCost:    "$75,000",
			PaymentTerms: "20% upfront, 40% at MVP, 40% at launch",
			Breakdown: []ProposalLineItem{
				{Item: "Development", Description: "Backend + Frontend", Cost: "$50,000"},
				{Item: "Testing", Description: "QA and UAT", Cost: "$15,000"},
				{Item: "DevOps", Description: "Infrastructure", Cost: "$10,000"},
			},
		},
		Assumptions: []string{"Client provides content"},
		Terms:       []string{"30-day payment terms"},
	}

	if proposal.PreparedFor != "Acme Corp" {
		t.Errorf("Expected prepared for 'Acme Corp', got %s", proposal.PreparedFor)
	}
	if len(proposal.Timeline.Milestones) != 2 {
		t.Errorf("Expected 2 milestones, got %d", len(proposal.Timeline.Milestones))
	}
	if len(proposal.Investment.Breakdown) != 3 {
		t.Errorf("Expected 3 cost items, got %d", len(proposal.Investment.Breakdown))
	}
}

func TestResourcePlanStructure(t *testing.T) {
	plan := ResourcePlan{
		AgentAllocation: []AgentAllocation{
			{
				AgentType:  "Backend Agent",
				Allocation: 60.0,
				Tasks:      []string{"TASK-001", "TASK-002", "TASK-003"},
				StartDate:  "2024-01-15",
				EndDate:    "2024-03-15",
			},
			{
				AgentType:  "Frontend Agent",
				Allocation: 40.0,
				Tasks:      []string{"TASK-004", "TASK-005"},
				StartDate:  "2024-02-01",
				EndDate:    "2024-04-01",
			},
		},
		TotalCapacity: 100,
		Utilization:   85.5,
	}

	if len(plan.AgentAllocation) != 2 {
		t.Errorf("Expected 2 agent allocations, got %d", len(plan.AgentAllocation))
	}
	if plan.AgentAllocation[0].Allocation != 60.0 {
		t.Errorf("Expected 60%% allocation, got %.1f%%", plan.AgentAllocation[0].Allocation)
	}
	if plan.Utilization != 85.5 {
		t.Errorf("Expected 85.5%% utilization, got %.1f%%", plan.Utilization)
	}
}

// ============================================================================
// Input Types Tests
// ============================================================================

func TestRequirementsInput(t *testing.T) {
	input := RequirementsInput{
		ProjectDescription: "Build an e-commerce platform",
		Conversation:       "User: I need an online store...",
		Industry:           "Retail",
		TargetUsers:        "Small businesses",
		Constraints:        []string{"Budget: $50k", "Timeline: 3 months"},
	}

	if input.Industry != "Retail" {
		t.Errorf("Expected industry 'Retail', got %s", input.Industry)
	}
	if len(input.Constraints) != 2 {
		t.Errorf("Expected 2 constraints, got %d", len(input.Constraints))
	}
}

func TestComplexityInput(t *testing.T) {
	input := ComplexityInput{
		Description: "E-commerce platform with payment processing",
		TechStack: &TechStack{
			Backend: BackendStack{Framework: "Node.js"},
			Database: DatabaseStack{Primary: "PostgreSQL"},
		},
		TeamSize: 5,
		Timeline: "3 months",
	}

	if input.TeamSize != 5 {
		t.Errorf("Expected team size 5, got %d", input.TeamSize)
	}
	if input.Timeline != "3 months" {
		t.Errorf("Expected timeline '3 months', got %s", input.Timeline)
	}
}

func TestEffortInput(t *testing.T) {
	input := EffortInput{
		Description: "Build e-commerce platform",
		TeamSize:    5,
		Complexity: &ComplexityScore{
			Overall:   "moderate",
			Score:     6,
			RiskLevel: "medium",
		},
	}

	if input.Complexity.Score != 6 {
		t.Errorf("Expected complexity score 6, got %d", input.Complexity.Score)
	}
}

func TestPlanInput(t *testing.T) {
	input := PlanInput{
		ProjectName:  "E-Commerce Platform",
		StartDate:    "2024-01-15",
		Methodology:  "agile",
		SprintLength: 14,
		Constraints:  []string{"Hard deadline April 15"},
	}

	if input.Methodology != "agile" {
		t.Errorf("Expected methodology 'agile', got %s", input.Methodology)
	}
	if input.SprintLength != 14 {
		t.Errorf("Expected sprint length 14, got %d", input.SprintLength)
	}
}

func TestProgressInput(t *testing.T) {
	input := ProgressInput{
		TaskUpdates: []TaskUpdate{
			{TaskID: "TASK-001", Status: "done", PercentComplete: 100},
			{TaskID: "TASK-002", Status: "in_progress", PercentComplete: 50},
		},
		BlockerUpdates: []Blocker{
			{ID: "BLOCK-001", Description: "API delay", Status: "open"},
		},
		SprintNumber: 3,
	}

	if len(input.TaskUpdates) != 2 {
		t.Errorf("Expected 2 task updates, got %d", len(input.TaskUpdates))
	}
	if input.SprintNumber != 3 {
		t.Errorf("Expected sprint number 3, got %d", input.SprintNumber)
	}
}

func TestStatusReportInput(t *testing.T) {
	input := StatusReportInput{
		Progress: &ProgressReport{
			ProjectName:   "E-Commerce",
			OverallStatus: "on_track",
			HealthScore:   85,
		},
		Audience:    "executive",
		Period:      "Week 6",
		CustomNotes: []string{"Client very satisfied"},
	}

	if input.Audience != "executive" {
		t.Errorf("Expected audience 'executive', got %s", input.Audience)
	}
}

func TestProposalInput(t *testing.T) {
	input := ProposalInput{
		ClientName:  "Acme Corp",
		ProjectName: "E-Commerce Platform",
		Description: "Build online store",
		StartDate:   "2024-01-15",
		HourlyRate:  150.0,
	}

	if input.HourlyRate != 150.0 {
		t.Errorf("Expected hourly rate 150.0, got %.2f", input.HourlyRate)
	}
}

// ============================================================================
// Schema Tests
// ============================================================================

func TestGetRequirementsSchema(t *testing.T) {
	schema := GetRequirementsSchema()

	if !containsString(schema, `"project_name"`) {
		t.Error("Schema should contain project_name")
	}
	if !containsString(schema, `"functional_requirements"`) {
		t.Error("Schema should contain functional_requirements")
	}
	if !containsString(schema, `"acceptance_criteria"`) {
		t.Error("Schema should contain acceptance_criteria")
	}
}

func TestGetComplexitySchema(t *testing.T) {
	schema := GetComplexitySchema()

	if !containsString(schema, `"overall"`) {
		t.Error("Schema should contain overall")
	}
	if !containsString(schema, `"dimensions"`) {
		t.Error("Schema should contain dimensions")
	}
	if !containsString(schema, `"risk_level"`) {
		t.Error("Schema should contain risk_level")
	}
}

func TestGetEffortSchema(t *testing.T) {
	schema := GetEffortSchema()

	if !containsString(schema, `"total_story_points"`) {
		t.Error("Schema should contain total_story_points")
	}
	if !containsString(schema, `"breakdown"`) {
		t.Error("Schema should contain breakdown")
	}
	if !containsString(schema, `"team_recommendation"`) {
		t.Error("Schema should contain team_recommendation")
	}
}

func TestGetProjectPlanSchema(t *testing.T) {
	schema := GetProjectPlanSchema()

	if !containsString(schema, `"phases"`) {
		t.Error("Schema should contain phases")
	}
	if !containsString(schema, `"milestones"`) {
		t.Error("Schema should contain milestones")
	}
	if !containsString(schema, `"sprints"`) {
		t.Error("Schema should contain sprints")
	}
}

func TestGetResourcePlanSchema(t *testing.T) {
	schema := GetResourcePlanSchema()

	if !containsString(schema, `"agent_allocation"`) {
		t.Error("Schema should contain agent_allocation")
	}
	if !containsString(schema, `"utilization"`) {
		t.Error("Schema should contain utilization")
	}
}

func TestGetProgressReportSchema(t *testing.T) {
	schema := GetProgressReportSchema()

	if !containsString(schema, `"overall_status"`) {
		t.Error("Schema should contain overall_status")
	}
	if !containsString(schema, `"health_score"`) {
		t.Error("Schema should contain health_score")
	}
	if !containsString(schema, `"velocity"`) {
		t.Error("Schema should contain velocity")
	}
}

func TestGetStatusReportSchema(t *testing.T) {
	schema := GetStatusReportSchema()

	if !containsString(schema, `"executive_summary"`) {
		t.Error("Schema should contain executive_summary")
	}
	if !containsString(schema, `"overall_health"`) {
		t.Error("Schema should contain overall_health")
	}
	if !containsString(schema, `"timeline"`) {
		t.Error("Schema should contain timeline")
	}
}

func TestGetProposalSchema(t *testing.T) {
	schema := GetProposalSchema()

	if !containsString(schema, `"executive_summary"`) {
		t.Error("Schema should contain executive_summary")
	}
	if !containsString(schema, `"scope"`) {
		t.Error("Schema should contain scope")
	}
	if !containsString(schema, `"investment"`) {
		t.Error("Schema should contain investment")
	}
}

// ============================================================================
// Helper Method Tests
// ============================================================================

func TestCalculateProgressMetrics(t *testing.T) {
	agent := NewEnhancedPMAgent()

	input := ProgressInput{
		ProjectPlan: &ProjectPlan{
			Phases: []ProjectPhase{
				{
					Tasks: []Task{
						{ID: "TASK-001", StoryPoints: 5},
						{ID: "TASK-002", StoryPoints: 8},
						{ID: "TASK-003", StoryPoints: 3},
						{ID: "TASK-004", StoryPoints: 5},
					},
				},
			},
		},
		TaskUpdates: []TaskUpdate{
			{TaskID: "TASK-001", Status: "done"},
			{TaskID: "TASK-002", Status: "done"},
			{TaskID: "TASK-003", Status: "in_progress"},
			{TaskID: "TASK-004", Status: "blocked"},
		},
	}

	report := agent.calculateProgressMetrics(input)

	if report.Completion.TasksTotal != 4 {
		t.Errorf("Expected 4 total tasks, got %d", report.Completion.TasksTotal)
	}
	if report.Completion.TasksCompleted != 2 {
		t.Errorf("Expected 2 completed tasks, got %d", report.Completion.TasksCompleted)
	}
	if report.Completion.TasksInProgress != 1 {
		t.Errorf("Expected 1 in-progress task, got %d", report.Completion.TasksInProgress)
	}
	if report.Completion.TasksBlocked != 1 {
		t.Errorf("Expected 1 blocked task, got %d", report.Completion.TasksBlocked)
	}
	if report.Completion.PointsTotal != 21 {
		t.Errorf("Expected 21 total points, got %d", report.Completion.PointsTotal)
	}
	if report.Completion.PointsCompleted != 13 {
		t.Errorf("Expected 13 completed points, got %d", report.Completion.PointsCompleted)
	}
	if report.Completion.PercentComplete != 50.0 {
		t.Errorf("Expected 50%% complete, got %.1f%%", report.Completion.PercentComplete)
	}
}

func TestCalculateProgressMetricsEmptyPlan(t *testing.T) {
	agent := NewEnhancedPMAgent()

	input := ProgressInput{
		ProjectPlan: nil,
	}

	report := agent.calculateProgressMetrics(input)

	if report.Completion.TasksTotal != 0 {
		t.Errorf("Expected 0 total tasks, got %d", report.Completion.TasksTotal)
	}
}

// ============================================================================
// Prompt Builder Tests
// ============================================================================

func TestBuildRequirementsPrompt(t *testing.T) {
	agent := NewEnhancedPMAgent()

	input := RequirementsInput{
		ProjectDescription: "Build e-commerce platform",
		Industry:           "Retail",
		TargetUsers:        "Small businesses",
		Constraints:        []string{"Budget: $50k"},
	}

	prompt := agent.buildRequirementsPrompt(input)

	if !containsString(prompt, "Build e-commerce platform") {
		t.Error("Prompt should contain project description")
	}
	if !containsString(prompt, "Industry: Retail") {
		t.Error("Prompt should contain industry")
	}
	if !containsString(prompt, "MoSCoW") {
		t.Error("Prompt should mention MoSCoW prioritization")
	}
}

func TestBuildComplexityPrompt(t *testing.T) {
	agent := NewEnhancedPMAgent()

	input := ComplexityInput{
		Description: "E-commerce platform",
		TeamSize:    5,
		Timeline:    "3 months",
	}

	prompt := agent.buildComplexityPrompt(input)

	if !containsString(prompt, "E-commerce platform") {
		t.Error("Prompt should contain description")
	}
	if !containsString(prompt, "Team Size: 5") {
		t.Error("Prompt should contain team size")
	}
	if !containsString(prompt, "Technical Complexity") {
		t.Error("Prompt should mention technical complexity dimension")
	}
}

func TestBuildEffortPrompt(t *testing.T) {
	agent := NewEnhancedPMAgent()

	input := EffortInput{
		Description: "E-commerce platform",
		Complexity: &ComplexityScore{
			Overall:   "moderate",
			Score:     6,
			RiskLevel: "medium",
		},
	}

	prompt := agent.buildEffortPrompt(input)

	if !containsString(prompt, "E-commerce platform") {
		t.Error("Prompt should contain description")
	}
	if !containsString(prompt, "moderate") {
		t.Error("Prompt should contain complexity level")
	}
	if !containsString(prompt, "Fibonacci") {
		t.Error("Prompt should mention Fibonacci story points")
	}
}

func TestBuildPlanPrompt(t *testing.T) {
	agent := NewEnhancedPMAgent()

	input := PlanInput{
		ProjectName:  "E-Commerce Platform",
		Methodology:  "agile",
		SprintLength: 14,
		StartDate:    "2024-01-15",
	}

	prompt := agent.buildPlanPrompt(input)

	if !containsString(prompt, "E-Commerce Platform") {
		t.Error("Prompt should contain project name")
	}
	if !containsString(prompt, "agile") {
		t.Error("Prompt should contain methodology")
	}
	if !containsString(prompt, "Sprint Length: 14") {
		t.Error("Prompt should contain sprint length")
	}
}

func TestBuildStatusReportPrompt(t *testing.T) {
	agent := NewEnhancedPMAgent()

	input := StatusReportInput{
		Progress: &ProgressReport{
			ProjectName:   "E-Commerce",
			OverallStatus: "on_track",
			HealthScore:   85,
		},
		Audience: "executive",
	}

	prompt := agent.buildStatusReportPrompt(input)

	if !containsString(prompt, "executive") {
		t.Error("Prompt should contain audience type")
	}
	if !containsString(prompt, "on_track") {
		t.Error("Prompt should contain status")
	}
	if !containsString(prompt, "business impact") {
		t.Error("Prompt should mention executive-specific guidance")
	}
}

func TestBuildProposalPrompt(t *testing.T) {
	agent := NewEnhancedPMAgent()

	input := ProposalInput{
		ClientName:  "Acme Corp",
		ProjectName: "E-Commerce Platform",
		HourlyRate:  150.0,
	}

	prompt := agent.buildProposalPrompt(input)

	if !containsString(prompt, "Acme Corp") {
		t.Error("Prompt should contain client name")
	}
	if !containsString(prompt, "E-Commerce Platform") {
		t.Error("Prompt should contain project name")
	}
	if !containsString(prompt, "150.00") {
		t.Error("Prompt should contain hourly rate")
	}
}
