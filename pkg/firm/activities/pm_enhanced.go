package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// EnhancedPMAgent handles professional project management including
// requirements analysis, sizing, planning, tracking, and reporting
type EnhancedPMAgent struct {
	llmClient llm.Client
}

// NewEnhancedPMAgent creates a new enhanced PM agent
func NewEnhancedPMAgent() *EnhancedPMAgent {
	return &EnhancedPMAgent{
		llmClient: llm.NewClient(),
	}
}

// ============================================================================
// Data Structures
// ============================================================================

// RequirementsDocument represents a structured requirements document
type RequirementsDocument struct {
	ProjectName       string            `json:"project_name"`
	Version           string            `json:"version"`
	CreatedAt         string            `json:"created_at"`
	Overview          string            `json:"overview"`
	Objectives        []string          `json:"objectives"`
	Stakeholders      []Stakeholder     `json:"stakeholders"`
	Scope             ProjectScope      `json:"scope"`
	FunctionalReqs    []Requirement     `json:"functional_requirements"`
	NonFunctionalReqs []Requirement     `json:"non_functional_requirements"`
	Constraints       []string          `json:"constraints"`
	Assumptions       []string          `json:"assumptions"`
	Risks             []Risk            `json:"risks"`
	Glossary          map[string]string `json:"glossary,omitempty"`
}

// Stakeholder represents a project stakeholder
type Stakeholder struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	Influence    string `json:"influence"` // "high", "medium", "low"
	Interest     string `json:"interest"`  // "high", "medium", "low"
	Expectations string `json:"expectations"`
}

// ProjectScope defines what's in and out of scope
type ProjectScope struct {
	InScope      []string `json:"in_scope"`
	OutOfScope   []string `json:"out_of_scope"`
	Deliverables []string `json:"deliverables"`
}

// Requirement represents a single requirement
type Requirement struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Priority           string   `json:"priority"` // "must", "should", "could", "wont"
	Category           string   `json:"category"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Dependencies       []string `json:"dependencies,omitempty"`
}

// Risk represents a project risk
type Risk struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Probability string `json:"probability"` // "high", "medium", "low"
	Impact      string `json:"impact"`      // "high", "medium", "low"
	Mitigation  string `json:"mitigation"`
	Owner       string `json:"owner,omitempty"`
}

// ComplexityScore represents the complexity analysis of a project
type ComplexityScore struct {
	Overall         string               `json:"overall"` // "trivial", "simple", "moderate", "complex", "very_complex"
	Score           int                  `json:"score"`   // 1-10
	Dimensions      ComplexityDimensions `json:"dimensions"`
	Factors         []ComplexityFactor   `json:"factors"`
	Recommendations []string             `json:"recommendations"`
	RiskLevel       string               `json:"risk_level"` // "low", "medium", "high", "critical"
}

// ComplexityDimensions breaks down complexity by category
type ComplexityDimensions struct {
	Technical      int `json:"technical"`       // 1-10
	Business       int `json:"business"`        // 1-10
	Integration    int `json:"integration"`     // 1-10
	DataComplexity int `json:"data_complexity"` // 1-10
	TeamSize       int `json:"team_size"`       // 1-10
	Timeline       int `json:"timeline"`        // 1-10
}

// ComplexityFactor represents a specific factor affecting complexity
type ComplexityFactor struct {
	Factor      string `json:"factor"`
	Impact      string `json:"impact"`   // "increases", "decreases"
	Severity    string `json:"severity"` // "minor", "moderate", "major"
	Description string `json:"description"`
}

// EffortEstimate represents the effort estimation for a project
type EffortEstimate struct {
	TotalStoryPoints   int                `json:"total_story_points"`
	TotalHours         int                `json:"total_hours"`
	TotalDays          int                `json:"total_days"`
	Confidence         string             `json:"confidence"` // "low", "medium", "high"
	Breakdown          []EffortBreakdown  `json:"breakdown"`
	TeamRecommendation TeamRecommendation `json:"team_recommendation"`
	Assumptions        []string           `json:"assumptions"`
	Caveats            []string           `json:"caveats"`
}

// EffortBreakdown shows effort by component/feature
type EffortBreakdown struct {
	Component   string `json:"component"`
	StoryPoints int    `json:"story_points"`
	Hours       int    `json:"hours"`
	Complexity  string `json:"complexity"`
	Notes       string `json:"notes,omitempty"`
}

// TeamRecommendation suggests team composition
type TeamRecommendation struct {
	MinTeamSize     int      `json:"min_team_size"`
	OptimalTeamSize int      `json:"optimal_team_size"`
	Roles           []string `json:"roles"`
	AgentTypes      []string `json:"agent_types"` // AI agents needed
}

// ProjectPlan represents a complete project plan
type ProjectPlan struct {
	ProjectName    string           `json:"project_name"`
	Version        string           `json:"version"`
	CreatedAt      string           `json:"created_at"`
	StartDate      string           `json:"start_date"`
	EndDate        string           `json:"end_date"`
	Methodology    string           `json:"methodology"`     // "agile", "waterfall", "hybrid"
	SprintDuration int              `json:"sprint_duration"` // days
	Phases         []ProjectPhase   `json:"phases"`
	Milestones     []Milestone      `json:"milestones"`
	Sprints        []Sprint         `json:"sprints,omitempty"`
	Dependencies   []TaskDependency `json:"dependencies"`
	CriticalPath   []string         `json:"critical_path"`
	Resources      ResourcePlan     `json:"resources"`
}

// ProjectPhase represents a major phase
type ProjectPhase struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	Deliverables []string `json:"deliverables"`
	Tasks        []Task   `json:"tasks"`
}

// Milestone represents a project milestone
type Milestone struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Date        string   `json:"date"`
	Description string   `json:"description"`
	Criteria    []string `json:"acceptance_criteria"`
	Phase       string   `json:"phase"`
}

// Sprint represents an agile sprint
type Sprint struct {
	Number      int      `json:"number"`
	Name        string   `json:"name"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
	Goal        string   `json:"goal"`
	Capacity    int      `json:"capacity"`     // story points
	PlannedWork []string `json:"planned_work"` // task IDs
}

// Task represents a project task
type Task struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Type           string   `json:"type"` // "development", "testing", "design", etc.
	Priority       string   `json:"priority"`
	StoryPoints    int      `json:"story_points"`
	EstimatedHours int      `json:"estimated_hours"`
	AssignedTo     string   `json:"assigned_to,omitempty"`
	Status         string   `json:"status"` // "todo", "in_progress", "done"
	Dependencies   []string `json:"dependencies"`
}

// TaskDependency represents a dependency between tasks
type TaskDependency struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // "finish_to_start", "start_to_start", etc.
}

// ResourcePlan defines resource allocation
type ResourcePlan struct {
	AgentAllocation []AgentAllocation `json:"agent_allocation"`
	TotalCapacity   int               `json:"total_capacity"`
	Utilization     float64           `json:"utilization"`
}

// AgentAllocation maps agents to work
type AgentAllocation struct {
	AgentType  string   `json:"agent_type"`
	Allocation float64  `json:"allocation"` // percentage
	Tasks      []string `json:"tasks"`
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
}

// ProgressReport represents project progress
type ProgressReport struct {
	ProjectName   string          `json:"project_name"`
	ReportDate    string          `json:"report_date"`
	ReportPeriod  string          `json:"report_period"`
	OverallStatus string          `json:"overall_status"` // "on_track", "at_risk", "delayed", "blocked"
	HealthScore   int             `json:"health_score"`   // 1-100
	Completion    ProgressMetrics `json:"completion"`
	Velocity      VelocityMetrics `json:"velocity"`
	BurndownData  []BurndownPoint `json:"burndown_data"`
	Blockers      []Blocker       `json:"blockers"`
	Highlights    []string        `json:"highlights"`
	Concerns      []string        `json:"concerns"`
	NextSteps     []string        `json:"next_steps"`
}

// ProgressMetrics tracks completion metrics
type ProgressMetrics struct {
	TasksTotal      int     `json:"tasks_total"`
	TasksCompleted  int     `json:"tasks_completed"`
	TasksInProgress int     `json:"tasks_in_progress"`
	TasksBlocked    int     `json:"tasks_blocked"`
	PercentComplete float64 `json:"percent_complete"`
	PointsTotal     int     `json:"points_total"`
	PointsCompleted int     `json:"points_completed"`
}

// VelocityMetrics tracks team velocity
type VelocityMetrics struct {
	CurrentSprint   int     `json:"current_sprint"`
	PlannedVelocity int     `json:"planned_velocity"`
	ActualVelocity  int     `json:"actual_velocity"`
	AverageVelocity float64 `json:"average_velocity"`
	Trend           string  `json:"trend"` // "improving", "stable", "declining"
}

// BurndownPoint represents a point on the burndown chart
type BurndownPoint struct {
	Date            string `json:"date"`
	RemainingPoints int    `json:"remaining_points"`
	IdealPoints     int    `json:"ideal_points"`
}

// Blocker represents a blocking issue
type Blocker struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Owner       string `json:"owner"`
	ETA         string `json:"eta,omitempty"`
	Status      string `json:"status"` // "open", "in_progress", "resolved"
}

// StatusReport is a client-facing status report
type StatusReport struct {
	ProjectName      string            `json:"project_name"`
	ReportDate       string            `json:"report_date"`
	ExecutiveSummary string            `json:"executive_summary"`
	OverallHealth    string            `json:"overall_health"` // "green", "yellow", "red"
	KeyMetrics       map[string]string `json:"key_metrics"`
	Accomplishments  []string          `json:"accomplishments"`
	PlannedWork      []string          `json:"planned_work"`
	Risks            []RiskUpdate      `json:"risks"`
	Decisions        []Decision        `json:"decisions_needed,omitempty"`
	Timeline         TimelineUpdate    `json:"timeline"`
}

// RiskUpdate represents a risk status update
type RiskUpdate struct {
	Risk       string `json:"risk"`
	Status     string `json:"status"` // "new", "ongoing", "mitigated", "closed"
	Mitigation string `json:"mitigation"`
}

// Decision represents a decision needed from stakeholders
type Decision struct {
	Topic       string   `json:"topic"`
	Description string   `json:"description"`
	Options     []string `json:"options"`
	DueDate     string   `json:"due_date"`
	Impact      string   `json:"impact"`
}

// TimelineUpdate shows schedule status
type TimelineUpdate struct {
	OriginalEndDate string `json:"original_end_date"`
	CurrentEndDate  string `json:"current_end_date"`
	Variance        string `json:"variance"`
	OnSchedule      bool   `json:"on_schedule"`
	Explanation     string `json:"explanation,omitempty"`
}

// Proposal represents a Statement of Work / Project Proposal
type Proposal struct {
	Title            string           `json:"title"`
	Version          string           `json:"version"`
	Date             string           `json:"date"`
	PreparedFor      string           `json:"prepared_for"`
	PreparedBy       string           `json:"prepared_by"`
	ExecutiveSummary string           `json:"executive_summary"`
	Background       string           `json:"background"`
	Objectives       []string         `json:"objectives"`
	Scope            ProposalScope    `json:"scope"`
	Approach         ProposalApproach `json:"approach"`
	Timeline         ProposalTimeline `json:"timeline"`
	Team             ProposalTeam     `json:"team"`
	Investment       ProposalCost     `json:"investment"`
	Assumptions      []string         `json:"assumptions"`
	Terms            []string         `json:"terms"`
}

// ProposalScope defines the scope for a proposal
type ProposalScope struct {
	Overview     string   `json:"overview"`
	Deliverables []string `json:"deliverables"`
	InScope      []string `json:"in_scope"`
	OutOfScope   []string `json:"out_of_scope"`
}

// ProposalApproach describes the methodology
type ProposalApproach struct {
	Methodology   string   `json:"methodology"`
	Description   string   `json:"description"`
	Phases        []string `json:"phases"`
	KeyActivities []string `json:"key_activities"`
}

// ProposalTimeline shows the schedule
type ProposalTimeline struct {
	StartDate  string              `json:"start_date"`
	EndDate    string              `json:"end_date"`
	Duration   string              `json:"duration"`
	Milestones []ProposalMilestone `json:"milestones"`
}

// ProposalMilestone for the proposal
type ProposalMilestone struct {
	Name    string `json:"name"`
	Date    string `json:"date"`
	Payment string `json:"payment,omitempty"`
}

// ProposalTeam describes the team
type ProposalTeam struct {
	Description string     `json:"description"`
	Roles       []TeamRole `json:"roles"`
}

// TeamRole describes a role on the team
type TeamRole struct {
	Role        string `json:"role"`
	Allocation  string `json:"allocation"`
	Description string `json:"description"`
}

// ProposalCost shows the investment
type ProposalCost struct {
	TotalCost    string             `json:"total_cost"`
	PaymentTerms string             `json:"payment_terms"`
	Breakdown    []ProposalLineItem `json:"breakdown"`
	Notes        []string           `json:"notes,omitempty"`
}

// ProposalLineItem represents a cost component
type ProposalLineItem struct {
	Item        string `json:"item"`
	Description string `json:"description"`
	Cost        string `json:"cost"`
}

// ============================================================================
// Input Types
// ============================================================================

// RequirementsInput for GatherRequirements
type RequirementsInput struct {
	ProjectDescription string   `json:"project_description"`
	Conversation       string   `json:"conversation,omitempty"`
	Industry           string   `json:"industry,omitempty"`
	TargetUsers        string   `json:"target_users,omitempty"`
	Constraints        []string `json:"constraints,omitempty"`
}

// ComplexityInput for EstimateComplexity
type ComplexityInput struct {
	Requirements *RequirementsDocument `json:"requirements,omitempty"`
	Description  string                `json:"description,omitempty"`
	TechStack    *TechStack            `json:"tech_stack,omitempty"`
	TeamSize     int                   `json:"team_size,omitempty"`
	Timeline     string                `json:"timeline,omitempty"`
}

// EffortInput for EstimateEffort
type EffortInput struct {
	Requirements *RequirementsDocument `json:"requirements,omitempty"`
	Complexity   *ComplexityScore      `json:"complexity,omitempty"`
	Description  string                `json:"description,omitempty"`
	TechStack    *TechStack            `json:"tech_stack,omitempty"`
	TeamSize     int                   `json:"team_size,omitempty"`
}

// PlanInput for CreateProjectPlan
type PlanInput struct {
	ProjectName  string                `json:"project_name"`
	Requirements *RequirementsDocument `json:"requirements,omitempty"`
	Effort       *EffortEstimate       `json:"effort,omitempty"`
	StartDate    string                `json:"start_date,omitempty"`
	Methodology  string                `json:"methodology,omitempty"` // "agile", "waterfall"
	SprintLength int                   `json:"sprint_length,omitempty"`
	Constraints  []string              `json:"constraints,omitempty"`
}

// ProgressInput for TrackProgress
type ProgressInput struct {
	ProjectPlan    *ProjectPlan `json:"project_plan"`
	TaskUpdates    []TaskUpdate `json:"task_updates"`
	BlockerUpdates []Blocker    `json:"blocker_updates,omitempty"`
	SprintNumber   int          `json:"sprint_number,omitempty"`
}

// TaskUpdate represents an update to a task
type TaskUpdate struct {
	TaskID          string `json:"task_id"`
	Status          string `json:"status"`
	PercentComplete int    `json:"percent_complete,omitempty"`
	ActualHours     int    `json:"actual_hours,omitempty"`
	Notes           string `json:"notes,omitempty"`
}

// StatusReportInput for GenerateStatusReport
type StatusReportInput struct {
	Progress    *ProgressReport `json:"progress"`
	Audience    string          `json:"audience,omitempty"` // "executive", "technical", "client"
	Period      string          `json:"period,omitempty"`
	CustomNotes []string        `json:"custom_notes,omitempty"`
}

// ProposalInput for GenerateProposal
type ProposalInput struct {
	ClientName   string                `json:"client_name"`
	ProjectName  string                `json:"project_name"`
	Requirements *RequirementsDocument `json:"requirements,omitempty"`
	Description  string                `json:"description,omitempty"`
	Effort       *EffortEstimate       `json:"effort,omitempty"`
	StartDate    string                `json:"start_date,omitempty"`
	HourlyRate   float64               `json:"hourly_rate,omitempty"`
}

// ============================================================================
// Activity Methods
// ============================================================================

// GatherRequirements transforms conversation into structured requirements
func (a *EnhancedPMAgent) GatherRequirements(ctx context.Context, input map[string]interface{}) (*RequirementsDocument, error) {
	inputJSON, _ := json.Marshal(input)
	var req RequirementsInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	prompt := a.buildRequirementsPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an expert Business Analyst and Product Manager. Create comprehensive requirements documents. Return valid JSON matching the schema: " + GetRequirementsSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var doc RequirementsDocument
	if err := json.Unmarshal([]byte(resp.Response), &doc); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	doc.CreatedAt = time.Now().Format(time.RFC3339)

	return &doc, nil
}

// EstimateComplexity analyzes project complexity
func (a *EnhancedPMAgent) EstimateComplexity(ctx context.Context, input map[string]interface{}) (*ComplexityScore, error) {
	inputJSON, _ := json.Marshal(input)
	var req ComplexityInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	prompt := a.buildComplexityPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an expert at estimating software project complexity. Analyze all factors and provide a detailed complexity assessment. Return valid JSON matching the schema: " + GetComplexitySchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var score ComplexityScore
	if err := json.Unmarshal([]byte(resp.Response), &score); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &score, nil
}

// EstimateEffort estimates effort in story points and hours
func (a *EnhancedPMAgent) EstimateEffort(ctx context.Context, input map[string]interface{}) (*EffortEstimate, error) {
	inputJSON, _ := json.Marshal(input)
	var req EffortInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	prompt := a.buildEffortPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an expert at software effort estimation. Use industry-standard techniques (planning poker, three-point estimation) to provide accurate estimates. Return valid JSON matching the schema: " + GetEffortSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var estimate EffortEstimate
	if err := json.Unmarshal([]byte(resp.Response), &estimate); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &estimate, nil
}

// CreateProjectPlan creates a detailed project plan
func (a *EnhancedPMAgent) CreateProjectPlan(ctx context.Context, input map[string]interface{}) (*ProjectPlan, error) {
	inputJSON, _ := json.Marshal(input)
	var req PlanInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if req.Methodology == "" {
		req.Methodology = "agile"
	}
	if req.SprintLength == 0 {
		req.SprintLength = 14 // 2 weeks
	}
	if req.StartDate == "" {
		req.StartDate = time.Now().Format("2006-01-02")
	}

	prompt := a.buildPlanPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an expert Project Manager. Create detailed, actionable project plans with proper task breakdown, dependencies, and resource allocation. Return valid JSON matching the schema: " + GetProjectPlanSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var plan ProjectPlan
	if err := json.Unmarshal([]byte(resp.Response), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	plan.ProjectName = req.ProjectName
	plan.CreatedAt = time.Now().Format(time.RFC3339)
	plan.StartDate = req.StartDate
	plan.Methodology = req.Methodology
	plan.SprintDuration = req.SprintLength

	return &plan, nil
}

// AllocateResources determines which AI agents are needed
func (a *EnhancedPMAgent) AllocateResources(ctx context.Context, input map[string]interface{}) (*ResourcePlan, error) {
	inputJSON, _ := json.Marshal(input)
	var planInput map[string]interface{}
	if err := json.Unmarshal(inputJSON, &planInput); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	prompt := a.buildResourcePrompt(planInput)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an expert at resource allocation for software projects. Determine the optimal AI agent composition. Return valid JSON matching the schema: " + GetResourcePlanSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var plan ResourcePlan
	if err := json.Unmarshal([]byte(resp.Response), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &plan, nil
}

// TrackProgress tracks project progress
func (a *EnhancedPMAgent) TrackProgress(ctx context.Context, input map[string]interface{}) (*ProgressReport, error) {
	inputJSON, _ := json.Marshal(input)
	var req ProgressInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Calculate metrics from task updates
	report := a.calculateProgressMetrics(req)

	prompt := a.buildProgressAnalysisPrompt(req, report)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an expert Project Manager analyzing project progress. Identify concerns, blockers, and provide actionable recommendations. Return valid JSON matching the schema: " + GetProgressReportSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var analysis ProgressReport
	if err := json.Unmarshal([]byte(resp.Response), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Merge calculated metrics with LLM analysis
	analysis.Completion = report.Completion
	analysis.ReportDate = time.Now().Format(time.RFC3339)

	return &analysis, nil
}

// GenerateStatusReport creates a client-facing status report
func (a *EnhancedPMAgent) GenerateStatusReport(ctx context.Context, input map[string]interface{}) (*StatusReport, error) {
	inputJSON, _ := json.Marshal(input)
	var req StatusReportInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if req.Audience == "" {
		req.Audience = "client"
	}

	prompt := a.buildStatusReportPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: fmt.Sprintf("You are an expert at communicating project status to %s stakeholders. Be clear, concise, and professional. Return valid JSON matching the schema: %s", req.Audience, GetStatusReportSchema()),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var report StatusReport
	if err := json.Unmarshal([]byte(resp.Response), &report); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	report.ReportDate = time.Now().Format(time.RFC3339)

	return &report, nil
}

// GenerateProposal creates a Statement of Work / Project Proposal
func (a *EnhancedPMAgent) GenerateProposal(ctx context.Context, input map[string]interface{}) (*Proposal, error) {
	inputJSON, _ := json.Marshal(input)
	var req ProposalInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	prompt := a.buildProposalPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an expert at writing professional project proposals and Statements of Work. Be persuasive yet accurate. Return valid JSON matching the schema: " + GetProposalSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var proposal Proposal
	if err := json.Unmarshal([]byte(resp.Response), &proposal); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	proposal.Date = time.Now().Format("2006-01-02")
	proposal.PreparedFor = req.ClientName

	return &proposal, nil
}

// RefinePM refines PM outputs based on feedback
func (a *EnhancedPMAgent) RefinePM(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	docType, _ := input["document_type"].(string)
	currentDoc, _ := input["current_document"].(map[string]interface{})
	feedback, _ := input["feedback"].(string)

	docJSON, _ := json.MarshalIndent(currentDoc, "", "  ")

	prompt := fmt.Sprintf(`Refine the following %s document based on the feedback.

Current Document:
%s

Feedback:
%s

Provide an improved document that addresses the feedback while maintaining professional quality.
Return the refined document in the same JSON structure.`, docType, string(docJSON), feedback)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an expert PM. Refine the document based on the feedback and return valid JSON.",
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM refinement failed: %w", err)
	}

	var refined map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Response), &refined); err != nil {
		return nil, fmt.Errorf("failed to parse refined document: %w", err)
	}

	return refined, nil
}

// ============================================================================
// Prompt Builders
// ============================================================================

func (a *EnhancedPMAgent) buildRequirementsPrompt(req RequirementsInput) string {
	var sb strings.Builder

	sb.WriteString("Create a comprehensive requirements document for the following project:\n\n")

	sb.WriteString(fmt.Sprintf("Project Description:\n%s\n\n", req.ProjectDescription))

	if req.Conversation != "" {
		sb.WriteString(fmt.Sprintf("Discovery Conversation:\n%s\n\n", req.Conversation))
	}

	if req.Industry != "" {
		sb.WriteString(fmt.Sprintf("Industry: %s\n", req.Industry))
	}

	if req.TargetUsers != "" {
		sb.WriteString(fmt.Sprintf("Target Users: %s\n", req.TargetUsers))
	}

	if len(req.Constraints) > 0 {
		sb.WriteString("Known Constraints:\n")
		for _, c := range req.Constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", c))
		}
	}

	sb.WriteString(`
Generate a complete requirements document including:
1. Clear project overview and objectives
2. Stakeholder identification
3. Scope definition (in/out of scope, deliverables)
4. Functional requirements with acceptance criteria (MoSCoW prioritization)
5. Non-functional requirements (performance, security, scalability)
6. Constraints and assumptions
7. Risk assessment with mitigation strategies

Use requirement IDs (REQ-001, REQ-002, etc.) and be specific with acceptance criteria.`)

	return sb.String()
}

func (a *EnhancedPMAgent) buildComplexityPrompt(req ComplexityInput) string {
	var sb strings.Builder

	sb.WriteString("Analyze the complexity of the following software project:\n\n")

	if req.Requirements != nil {
		sb.WriteString(fmt.Sprintf("Project: %s\n", req.Requirements.ProjectName))
		sb.WriteString(fmt.Sprintf("Overview: %s\n\n", req.Requirements.Overview))

		sb.WriteString("Functional Requirements:\n")
		for _, r := range req.Requirements.FunctionalReqs {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", r.Priority, r.Title, r.Description))
		}

		sb.WriteString("\nNon-Functional Requirements:\n")
		for _, r := range req.Requirements.NonFunctionalReqs {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", r.Priority, r.Title, r.Description))
		}
	} else if req.Description != "" {
		sb.WriteString(fmt.Sprintf("Description:\n%s\n\n", req.Description))
	}

	if req.TechStack != nil {
		sb.WriteString(fmt.Sprintf("\nTech Stack: %s + %s\n", req.TechStack.Backend.Framework, req.TechStack.Database.Primary))
	}

	if req.TeamSize > 0 {
		sb.WriteString(fmt.Sprintf("Team Size: %d\n", req.TeamSize))
	}

	if req.Timeline != "" {
		sb.WriteString(fmt.Sprintf("Timeline Constraint: %s\n", req.Timeline))
	}

	sb.WriteString(`
Assess complexity across these dimensions (1-10 scale):
1. Technical Complexity - Architecture, algorithms, new technologies
2. Business Complexity - Domain knowledge, rules, workflows
3. Integration Complexity - External systems, APIs, data migration
4. Data Complexity - Volume, variety, transformations
5. Team/Coordination - Communication overhead, skill gaps
6. Timeline Pressure - Deadline constraints

Provide:
- Overall complexity rating (trivial/simple/moderate/complex/very_complex)
- Numeric score (1-10)
- Key factors increasing/decreasing complexity
- Risk level assessment
- Specific recommendations`)

	return sb.String()
}

func (a *EnhancedPMAgent) buildEffortPrompt(req EffortInput) string {
	var sb strings.Builder

	sb.WriteString("Estimate the effort required for the following software project:\n\n")

	if req.Requirements != nil {
		sb.WriteString(fmt.Sprintf("Project: %s\n\n", req.Requirements.ProjectName))

		sb.WriteString("Requirements Summary:\n")
		mustCount := 0
		shouldCount := 0
		for _, r := range req.Requirements.FunctionalReqs {
			if r.Priority == "must" {
				mustCount++
			} else if r.Priority == "should" {
				shouldCount++
			}
		}
		sb.WriteString(fmt.Sprintf("- %d MUST-have requirements\n", mustCount))
		sb.WriteString(fmt.Sprintf("- %d SHOULD-have requirements\n", shouldCount))
		sb.WriteString(fmt.Sprintf("- %d total functional requirements\n", len(req.Requirements.FunctionalReqs)))
		sb.WriteString(fmt.Sprintf("- %d non-functional requirements\n\n", len(req.Requirements.NonFunctionalReqs)))
	}

	if req.Complexity != nil {
		sb.WriteString("Complexity Assessment:\n")
		sb.WriteString(fmt.Sprintf("- Overall: %s (Score: %d/10)\n", req.Complexity.Overall, req.Complexity.Score))
		sb.WriteString(fmt.Sprintf("- Risk Level: %s\n\n", req.Complexity.RiskLevel))
	}

	if req.Description != "" {
		sb.WriteString(fmt.Sprintf("Description:\n%s\n\n", req.Description))
	}

	if req.TechStack != nil {
		sb.WriteString(fmt.Sprintf("Tech Stack: %s backend + %s database\n\n", req.TechStack.Backend.Framework, req.TechStack.Database.Primary))
	}

	sb.WriteString(`
Provide effort estimates including:
1. Total story points (use Fibonacci: 1,2,3,5,8,13,21)
2. Total hours (assume 6 productive hours per day)
3. Total days
4. Confidence level (low/medium/high)

Break down by component/feature:
- Component name
- Story points
- Hours
- Complexity level

Include:
- Team size recommendation (min and optimal)
- Required roles/agent types
- Key assumptions
- Important caveats

Use industry-standard estimation techniques.`)

	return sb.String()
}

func (a *EnhancedPMAgent) buildPlanPrompt(req PlanInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Create a %s project plan for: %s\n\n", req.Methodology, req.ProjectName))
	sb.WriteString(fmt.Sprintf("Start Date: %s\n", req.StartDate))

	if req.Methodology == "agile" {
		sb.WriteString(fmt.Sprintf("Sprint Length: %d days\n", req.SprintLength))
	}

	if req.Requirements != nil {
		sb.WriteString("\nRequirements to implement:\n")
		for _, r := range req.Requirements.FunctionalReqs {
			sb.WriteString(fmt.Sprintf("- [%s] %s\n", r.Priority, r.Title))
		}
	}

	if req.Effort != nil {
		sb.WriteString(fmt.Sprintf("\nEffort Estimate: %d story points, %d days\n", req.Effort.TotalStoryPoints, req.Effort.TotalDays))
	}

	if len(req.Constraints) > 0 {
		sb.WriteString("\nConstraints:\n")
		for _, c := range req.Constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", c))
		}
	}

	sb.WriteString(`
Create a detailed project plan including:

1. Project Phases with:
   - Start and end dates
   - Deliverables for each phase
   - Tasks with IDs, story points, and hours

2. Milestones with:
   - Clear acceptance criteria
   - Target dates

3. Sprints (if Agile) with:
   - Sprint goals
   - Planned work (task IDs)
   - Capacity

4. Task Dependencies:
   - Which tasks block others
   - Critical path identification

5. Resource Allocation:
   - AI agent types needed
   - Allocation percentages
   - Timeline for each agent

Ensure tasks have unique IDs (TASK-001, etc.) and are properly sequenced.`)

	return sb.String()
}

func (a *EnhancedPMAgent) buildResourcePrompt(planInput map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("Determine the optimal AI agent allocation for this project:\n\n")

	planJSON, _ := json.MarshalIndent(planInput, "", "  ")
	sb.WriteString(string(planJSON))

	sb.WriteString(`

Available AI Agent Types:
- Architect Agent: System design, tech stack selection, API design
- Backend Agent: API services, business logic, database integration
- Frontend Agent: UI components, user experience, accessibility
- Database Agent: Schema design, migrations, optimization
- DevOps Agent: CI/CD, containers, deployment
- Infrastructure Agent: Cloud resources, IaC
- QA Agent: Test generation, coverage analysis
- SRE Agent: Monitoring, alerting, runbooks
- PM Agent: Planning, tracking, reporting

For each required agent, specify:
1. Agent type
2. Allocation percentage (0-100%)
3. Assigned tasks (task IDs)
4. Start and end dates

Calculate:
- Total capacity needed
- Overall utilization

Optimize for parallel execution where possible.`)

	return sb.String()
}

func (a *EnhancedPMAgent) calculateProgressMetrics(req ProgressInput) *ProgressReport {
	report := &ProgressReport{}

	if req.ProjectPlan == nil {
		return report
	}

	// Count tasks by status
	totalTasks := 0
	completedTasks := 0
	inProgressTasks := 0
	blockedTasks := 0
	totalPoints := 0
	completedPoints := 0

	for _, phase := range req.ProjectPlan.Phases {
		for _, task := range phase.Tasks {
			totalTasks++
			totalPoints += task.StoryPoints

			// Check for updates
			for _, update := range req.TaskUpdates {
				if update.TaskID == task.ID {
					switch update.Status {
					case "done", "completed":
						completedTasks++
						completedPoints += task.StoryPoints
					case "in_progress":
						inProgressTasks++
					case "blocked":
						blockedTasks++
					}
					break
				}
			}
		}
	}

	percentComplete := 0.0
	if totalTasks > 0 {
		percentComplete = float64(completedTasks) / float64(totalTasks) * 100
	}

	report.Completion = ProgressMetrics{
		TasksTotal:      totalTasks,
		TasksCompleted:  completedTasks,
		TasksInProgress: inProgressTasks,
		TasksBlocked:    blockedTasks,
		PercentComplete: percentComplete,
		PointsTotal:     totalPoints,
		PointsCompleted: completedPoints,
	}

	return report
}

func (a *EnhancedPMAgent) buildProgressAnalysisPrompt(req ProgressInput, metrics *ProgressReport) string {
	var sb strings.Builder

	sb.WriteString("Analyze the following project progress:\n\n")

	sb.WriteString("Progress Metrics:\n")
	sb.WriteString(fmt.Sprintf("- Tasks: %d/%d completed (%.1f%%)\n",
		metrics.Completion.TasksCompleted, metrics.Completion.TasksTotal, metrics.Completion.PercentComplete))
	sb.WriteString(fmt.Sprintf("- Story Points: %d/%d completed\n",
		metrics.Completion.PointsCompleted, metrics.Completion.PointsTotal))
	sb.WriteString(fmt.Sprintf("- In Progress: %d tasks\n", metrics.Completion.TasksInProgress))
	sb.WriteString(fmt.Sprintf("- Blocked: %d tasks\n\n", metrics.Completion.TasksBlocked))

	if len(req.BlockerUpdates) > 0 {
		sb.WriteString("Current Blockers:\n")
		for _, b := range req.BlockerUpdates {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", b.ID, b.Description))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`
Provide analysis including:
1. Overall status (on_track, at_risk, delayed, blocked)
2. Health score (1-100)
3. Velocity metrics and trend
4. Key highlights and accomplishments
5. Concerns and risks
6. Recommended next steps

Be specific and actionable in your recommendations.`)

	return sb.String()
}

func (a *EnhancedPMAgent) buildStatusReportPrompt(req StatusReportInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate a %s status report based on the following progress:\n\n", req.Audience))

	if req.Progress != nil {
		sb.WriteString(fmt.Sprintf("Project: %s\n", req.Progress.ProjectName))
		sb.WriteString(fmt.Sprintf("Overall Status: %s\n", req.Progress.OverallStatus))
		sb.WriteString(fmt.Sprintf("Health Score: %d/100\n\n", req.Progress.HealthScore))

		sb.WriteString("Completion:\n")
		sb.WriteString(fmt.Sprintf("- Tasks: %d/%d (%.1f%%)\n",
			req.Progress.Completion.TasksCompleted,
			req.Progress.Completion.TasksTotal,
			req.Progress.Completion.PercentComplete))

		if len(req.Progress.Blockers) > 0 {
			sb.WriteString("\nBlockers:\n")
			for _, b := range req.Progress.Blockers {
				sb.WriteString(fmt.Sprintf("- %s\n", b.Description))
			}
		}
	}

	if len(req.CustomNotes) > 0 {
		sb.WriteString("\nAdditional Notes:\n")
		for _, note := range req.CustomNotes {
			sb.WriteString(fmt.Sprintf("- %s\n", note))
		}
	}

	audienceGuidance := map[string]string{
		"executive": "Focus on business impact, ROI, and high-level progress. Avoid technical details.",
		"technical": "Include technical accomplishments, architecture decisions, and technical debt.",
		"client":    "Balance business value with transparency. Highlight deliverables and next milestones.",
	}

	sb.WriteString(fmt.Sprintf("\nAudience Guidance: %s\n", audienceGuidance[req.Audience]))

	sb.WriteString(`
Generate a professional status report including:
1. Executive summary (2-3 sentences)
2. Overall health indicator (green/yellow/red)
3. Key metrics
4. Top accomplishments this period
5. Planned work for next period
6. Risk updates with mitigation status
7. Timeline update (on schedule or variance)
8. Decisions needed (if any)`)

	return sb.String()
}

func (a *EnhancedPMAgent) buildProposalPrompt(req ProposalInput) string {
	var sb strings.Builder

	sb.WriteString("Create a professional project proposal/Statement of Work for:\n\n")
	sb.WriteString(fmt.Sprintf("Client: %s\n", req.ClientName))
	sb.WriteString(fmt.Sprintf("Project: %s\n\n", req.ProjectName))

	if req.Description != "" {
		sb.WriteString(fmt.Sprintf("Project Description:\n%s\n\n", req.Description))
	}

	if req.Requirements != nil {
		sb.WriteString("Key Requirements:\n")
		for i, r := range req.Requirements.FunctionalReqs {
			if i >= 5 {
				sb.WriteString(fmt.Sprintf("- ... and %d more requirements\n", len(req.Requirements.FunctionalReqs)-5))
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", r.Title))
		}
		sb.WriteString("\n")
	}

	if req.Effort != nil {
		sb.WriteString("Effort Estimate:\n")
		sb.WriteString(fmt.Sprintf("- Story Points: %d\n", req.Effort.TotalStoryPoints))
		sb.WriteString(fmt.Sprintf("- Hours: %d\n", req.Effort.TotalHours))
		sb.WriteString(fmt.Sprintf("- Days: %d\n\n", req.Effort.TotalDays))
	}

	if req.HourlyRate > 0 {
		sb.WriteString(fmt.Sprintf("Hourly Rate: $%.2f\n\n", req.HourlyRate))
	}

	if req.StartDate != "" {
		sb.WriteString(fmt.Sprintf("Proposed Start Date: %s\n\n", req.StartDate))
	}

	sb.WriteString(`
Create a professional proposal including:
1. Executive Summary (compelling value proposition)
2. Background (client context and opportunity)
3. Objectives (measurable goals)
4. Scope:
   - Overview
   - Key deliverables
   - What's in/out of scope
5. Approach:
   - Methodology (Agile recommended)
   - Key phases and activities
6. Timeline with milestones (and payment schedule if applicable)
7. Team description with roles
8. Investment (cost breakdown)
9. Key assumptions
10. Terms and conditions

Make it persuasive yet professional. Be specific about deliverables.`)

	return sb.String()
}

// ============================================================================
// JSON Schemas
// ============================================================================

// GetRequirementsSchema returns JSON schema for requirements document
func GetRequirementsSchema() string {
	return `{
  "type": "object",
  "properties": {
    "project_name": { "type": "string" },
    "version": { "type": "string" },
    "overview": { "type": "string" },
    "objectives": { "type": "array", "items": { "type": "string" } },
    "stakeholders": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "role": { "type": "string" },
          "influence": { "type": "string" },
          "interest": { "type": "string" },
          "expectations": { "type": "string" }
        }
      }
    },
    "scope": {
      "type": "object",
      "properties": {
        "in_scope": { "type": "array", "items": { "type": "string" } },
        "out_of_scope": { "type": "array", "items": { "type": "string" } },
        "deliverables": { "type": "array", "items": { "type": "string" } }
      }
    },
    "functional_requirements": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "title": { "type": "string" },
          "description": { "type": "string" },
          "priority": { "type": "string" },
          "category": { "type": "string" },
          "acceptance_criteria": { "type": "array", "items": { "type": "string" } }
        },
        "required": ["id", "title", "description", "priority"]
      }
    },
    "non_functional_requirements": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "title": { "type": "string" },
          "description": { "type": "string" },
          "priority": { "type": "string" },
          "category": { "type": "string" }
        }
      }
    },
    "constraints": { "type": "array", "items": { "type": "string" } },
    "assumptions": { "type": "array", "items": { "type": "string" } },
    "risks": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "description": { "type": "string" },
          "probability": { "type": "string" },
          "impact": { "type": "string" },
          "mitigation": { "type": "string" }
        }
      }
    }
  },
  "required": ["project_name", "overview", "objectives", "scope", "functional_requirements"]
}`
}

// GetComplexitySchema returns JSON schema for complexity score
func GetComplexitySchema() string {
	return `{
  "type": "object",
  "properties": {
    "overall": { "type": "string", "enum": ["trivial", "simple", "moderate", "complex", "very_complex"] },
    "score": { "type": "integer", "minimum": 1, "maximum": 10 },
    "dimensions": {
      "type": "object",
      "properties": {
        "technical": { "type": "integer" },
        "business": { "type": "integer" },
        "integration": { "type": "integer" },
        "data_complexity": { "type": "integer" },
        "team_size": { "type": "integer" },
        "timeline": { "type": "integer" }
      }
    },
    "factors": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "factor": { "type": "string" },
          "impact": { "type": "string" },
          "severity": { "type": "string" },
          "description": { "type": "string" }
        }
      }
    },
    "recommendations": { "type": "array", "items": { "type": "string" } },
    "risk_level": { "type": "string", "enum": ["low", "medium", "high", "critical"] }
  },
  "required": ["overall", "score", "dimensions", "risk_level"]
}`
}

// GetEffortSchema returns JSON schema for effort estimate
func GetEffortSchema() string {
	return `{
  "type": "object",
  "properties": {
    "total_story_points": { "type": "integer" },
    "total_hours": { "type": "integer" },
    "total_days": { "type": "integer" },
    "confidence": { "type": "string", "enum": ["low", "medium", "high"] },
    "breakdown": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "component": { "type": "string" },
          "story_points": { "type": "integer" },
          "hours": { "type": "integer" },
          "complexity": { "type": "string" },
          "notes": { "type": "string" }
        },
        "required": ["component", "story_points", "hours"]
      }
    },
    "team_recommendation": {
      "type": "object",
      "properties": {
        "min_team_size": { "type": "integer" },
        "optimal_team_size": { "type": "integer" },
        "roles": { "type": "array", "items": { "type": "string" } },
        "agent_types": { "type": "array", "items": { "type": "string" } }
      }
    },
    "assumptions": { "type": "array", "items": { "type": "string" } },
    "caveats": { "type": "array", "items": { "type": "string" } }
  },
  "required": ["total_story_points", "total_hours", "total_days", "confidence", "breakdown"]
}`
}

// GetProjectPlanSchema returns JSON schema for project plan
func GetProjectPlanSchema() string {
	return `{
  "type": "object",
  "properties": {
    "end_date": { "type": "string" },
    "phases": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "description": { "type": "string" },
          "start_date": { "type": "string" },
          "end_date": { "type": "string" },
          "deliverables": { "type": "array", "items": { "type": "string" } },
          "tasks": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "id": { "type": "string" },
                "name": { "type": "string" },
                "description": { "type": "string" },
                "type": { "type": "string" },
                "priority": { "type": "string" },
                "story_points": { "type": "integer" },
                "estimated_hours": { "type": "integer" },
                "status": { "type": "string" },
                "dependencies": { "type": "array", "items": { "type": "string" } }
              },
              "required": ["id", "name", "story_points"]
            }
          }
        },
        "required": ["id", "name", "tasks"]
      }
    },
    "milestones": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "date": { "type": "string" },
          "description": { "type": "string" },
          "acceptance_criteria": { "type": "array", "items": { "type": "string" } },
          "phase": { "type": "string" }
        },
        "required": ["id", "name", "date"]
      }
    },
    "sprints": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "number": { "type": "integer" },
          "name": { "type": "string" },
          "start_date": { "type": "string" },
          "end_date": { "type": "string" },
          "goal": { "type": "string" },
          "capacity": { "type": "integer" },
          "planned_work": { "type": "array", "items": { "type": "string" } }
        }
      }
    },
    "dependencies": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "from": { "type": "string" },
          "to": { "type": "string" },
          "type": { "type": "string" }
        }
      }
    },
    "critical_path": { "type": "array", "items": { "type": "string" } },
    "resources": {
      "type": "object",
      "properties": {
        "agent_allocation": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "agent_type": { "type": "string" },
              "allocation": { "type": "number" },
              "tasks": { "type": "array", "items": { "type": "string" } },
              "start_date": { "type": "string" },
              "end_date": { "type": "string" }
            }
          }
        },
        "total_capacity": { "type": "integer" },
        "utilization": { "type": "number" }
      }
    }
  },
  "required": ["phases", "milestones"]
}`
}

// GetResourcePlanSchema returns JSON schema for resource plan
func GetResourcePlanSchema() string {
	return `{
  "type": "object",
  "properties": {
    "agent_allocation": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "agent_type": { "type": "string" },
          "allocation": { "type": "number" },
          "tasks": { "type": "array", "items": { "type": "string" } },
          "start_date": { "type": "string" },
          "end_date": { "type": "string" }
        },
        "required": ["agent_type", "allocation"]
      }
    },
    "total_capacity": { "type": "integer" },
    "utilization": { "type": "number" }
  },
  "required": ["agent_allocation"]
}`
}

// GetProgressReportSchema returns JSON schema for progress report
func GetProgressReportSchema() string {
	return `{
  "type": "object",
  "properties": {
    "project_name": { "type": "string" },
    "report_period": { "type": "string" },
    "overall_status": { "type": "string", "enum": ["on_track", "at_risk", "delayed", "blocked"] },
    "health_score": { "type": "integer", "minimum": 1, "maximum": 100 },
    "velocity": {
      "type": "object",
      "properties": {
        "current_sprint": { "type": "integer" },
        "planned_velocity": { "type": "integer" },
        "actual_velocity": { "type": "integer" },
        "average_velocity": { "type": "number" },
        "trend": { "type": "string" }
      }
    },
    "burndown_data": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "date": { "type": "string" },
          "remaining_points": { "type": "integer" },
          "ideal_points": { "type": "integer" }
        }
      }
    },
    "blockers": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "description": { "type": "string" },
          "impact": { "type": "string" },
          "owner": { "type": "string" },
          "status": { "type": "string" }
        }
      }
    },
    "highlights": { "type": "array", "items": { "type": "string" } },
    "concerns": { "type": "array", "items": { "type": "string" } },
    "next_steps": { "type": "array", "items": { "type": "string" } }
  },
  "required": ["overall_status", "health_score"]
}`
}

// GetStatusReportSchema returns JSON schema for status report
func GetStatusReportSchema() string {
	return `{
  "type": "object",
  "properties": {
    "project_name": { "type": "string" },
    "executive_summary": { "type": "string" },
    "overall_health": { "type": "string", "enum": ["green", "yellow", "red"] },
    "key_metrics": { "type": "object" },
    "accomplishments": { "type": "array", "items": { "type": "string" } },
    "planned_work": { "type": "array", "items": { "type": "string" } },
    "risks": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "risk": { "type": "string" },
          "status": { "type": "string" },
          "mitigation": { "type": "string" }
        }
      }
    },
    "decisions_needed": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "topic": { "type": "string" },
          "description": { "type": "string" },
          "options": { "type": "array", "items": { "type": "string" } },
          "due_date": { "type": "string" },
          "impact": { "type": "string" }
        }
      }
    },
    "timeline": {
      "type": "object",
      "properties": {
        "original_end_date": { "type": "string" },
        "current_end_date": { "type": "string" },
        "variance": { "type": "string" },
        "on_schedule": { "type": "boolean" },
        "explanation": { "type": "string" }
      }
    }
  },
  "required": ["executive_summary", "overall_health", "accomplishments", "planned_work"]
}`
}

// GetProposalSchema returns JSON schema for proposal
func GetProposalSchema() string {
	return `{
  "type": "object",
  "properties": {
    "title": { "type": "string" },
    "version": { "type": "string" },
    "prepared_by": { "type": "string" },
    "executive_summary": { "type": "string" },
    "background": { "type": "string" },
    "objectives": { "type": "array", "items": { "type": "string" } },
    "scope": {
      "type": "object",
      "properties": {
        "overview": { "type": "string" },
        "deliverables": { "type": "array", "items": { "type": "string" } },
        "in_scope": { "type": "array", "items": { "type": "string" } },
        "out_of_scope": { "type": "array", "items": { "type": "string" } }
      }
    },
    "approach": {
      "type": "object",
      "properties": {
        "methodology": { "type": "string" },
        "description": { "type": "string" },
        "phases": { "type": "array", "items": { "type": "string" } },
        "key_activities": { "type": "array", "items": { "type": "string" } }
      }
    },
    "timeline": {
      "type": "object",
      "properties": {
        "start_date": { "type": "string" },
        "end_date": { "type": "string" },
        "duration": { "type": "string" },
        "milestones": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "name": { "type": "string" },
              "date": { "type": "string" },
              "payment": { "type": "string" }
            }
          }
        }
      }
    },
    "team": {
      "type": "object",
      "properties": {
        "description": { "type": "string" },
        "roles": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "role": { "type": "string" },
              "allocation": { "type": "string" },
              "description": { "type": "string" }
            }
          }
        }
      }
    },
    "investment": {
      "type": "object",
      "properties": {
        "total_cost": { "type": "string" },
        "payment_terms": { "type": "string" },
        "breakdown": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "item": { "type": "string" },
              "description": { "type": "string" },
              "cost": { "type": "string" }
            }
          }
        },
        "notes": { "type": "array", "items": { "type": "string" } }
      }
    },
    "assumptions": { "type": "array", "items": { "type": "string" } },
    "terms": { "type": "array", "items": { "type": "string" } }
  },
  "required": ["title", "executive_summary", "objectives", "scope", "approach", "timeline", "investment"]
}`
}
