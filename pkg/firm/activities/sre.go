package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// SREAgent handles Site Reliability Engineering tasks including monitoring,
// alerting, dashboards, runbooks, and incident management
type SREAgent struct {
	llmClient llm.Client
}

// NewSREAgent creates a new SRE agent
func NewSREAgent() *SREAgent {
	return &SREAgent{
		llmClient: llm.NewClient(),
	}
}

// ============================================================================
// Data Structures
// ============================================================================

// MonitoringConfig represents the complete monitoring configuration
type MonitoringConfig struct {
	Platform       string              `json:"platform"`        // "prometheus", "datadog", "cloudwatch"
	ServiceName    string              `json:"service_name"`
	Metrics        []MetricDefinition  `json:"metrics"`
	ScrapeConfigs  []ScrapeConfig      `json:"scrape_configs,omitempty"`
	RecordingRules []RecordingRule     `json:"recording_rules,omitempty"`
	Labels         map[string]string   `json:"labels,omitempty"`
	Files          map[string]string   `json:"files"`
}

// MetricDefinition defines a metric to be collected
type MetricDefinition struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`        // "counter", "gauge", "histogram", "summary"
	Description string            `json:"description"`
	Labels      []string          `json:"labels,omitempty"`
	Buckets     []float64         `json:"buckets,omitempty"` // for histograms
	Objectives  map[float64]float64 `json:"objectives,omitempty"` // for summaries
}

// ScrapeConfig defines Prometheus scrape configuration
type ScrapeConfig struct {
	JobName        string            `json:"job_name"`
	ScrapeInterval string            `json:"scrape_interval"`
	ScrapeTimeout  string            `json:"scrape_timeout,omitempty"`
	MetricsPath    string            `json:"metrics_path,omitempty"`
	StaticConfigs  []StaticConfig    `json:"static_configs,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
}

// StaticConfig defines static target configuration
type StaticConfig struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels,omitempty"`
}

// RecordingRule defines a Prometheus recording rule
type RecordingRule struct {
	Record string            `json:"record"`
	Expr   string            `json:"expr"`
	Labels map[string]string `json:"labels,omitempty"`
}

// AlertRules represents alert configuration
type AlertRules struct {
	Platform    string       `json:"platform"`     // "prometheus", "datadog", "cloudwatch", "pagerduty"
	ServiceName string       `json:"service_name"`
	SLOs        []SLO        `json:"slos"`
	Alerts      []AlertRule  `json:"alerts"`
	Files       map[string]string `json:"files"`
}

// SLO defines a Service Level Objective
type SLO struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Target      float64 `json:"target"`      // e.g., 99.9 for 99.9%
	Window      string  `json:"window"`      // e.g., "30d", "7d"
	Indicator   string  `json:"indicator"`   // "availability", "latency", "error_rate", "throughput"
	Query       string  `json:"query"`       // The metric query
}

// AlertRule defines an alert
type AlertRule struct {
	Name        string            `json:"name"`
	Severity    string            `json:"severity"`    // "critical", "warning", "info"
	Condition   string            `json:"condition"`   // The alert expression
	Duration    string            `json:"duration"`    // How long before firing
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	Runbook     string            `json:"runbook,omitempty"`     // Link to runbook
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// DashboardBundle contains generated dashboards
type DashboardBundle struct {
	Platform    string            `json:"platform"`    // "grafana", "datadog", "cloudwatch"
	ServiceName string            `json:"service_name"`
	Dashboards  []Dashboard       `json:"dashboards"`
	Files       map[string]string `json:"files"`
}

// Dashboard represents a monitoring dashboard
type Dashboard struct {
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Tags        []string        `json:"tags,omitempty"`
	Panels      []DashboardPanel `json:"panels"`
	Variables   []DashboardVar  `json:"variables,omitempty"`
	Refresh     string          `json:"refresh,omitempty"`
	TimeRange   string          `json:"time_range,omitempty"`
}

// DashboardPanel represents a panel in a dashboard
type DashboardPanel struct {
	Title       string   `json:"title"`
	Type        string   `json:"type"`        // "graph", "stat", "gauge", "table", "heatmap", "logs"
	Query       string   `json:"query"`
	Description string   `json:"description,omitempty"`
	Unit        string   `json:"unit,omitempty"`
	Thresholds  []Threshold `json:"thresholds,omitempty"`
	GridPos     GridPos  `json:"grid_pos,omitempty"`
}

// Threshold defines visual thresholds
type Threshold struct {
	Value float64 `json:"value"`
	Color string  `json:"color"`
}

// GridPos defines panel position in Grafana
type GridPos struct {
	H int `json:"h"`
	W int `json:"w"`
	X int `json:"x"`
	Y int `json:"y"`
}

// DashboardVar represents a dashboard variable
type DashboardVar struct {
	Name    string   `json:"name"`
	Label   string   `json:"label"`
	Type    string   `json:"type"`    // "query", "custom", "constant", "interval"
	Query   string   `json:"query,omitempty"`
	Options []string `json:"options,omitempty"`
	Default string   `json:"default,omitempty"`
}

// RunbookBundle contains generated runbooks
type RunbookBundle struct {
	ServiceName string    `json:"service_name"`
	Runbooks    []Runbook `json:"runbooks"`
	Files       map[string]string `json:"files"`
}

// Runbook represents an operational runbook
type Runbook struct {
	Name           string         `json:"name"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	AlertName      string         `json:"alert_name,omitempty"`
	Severity       string         `json:"severity"`
	Impact         string         `json:"impact"`
	Symptoms       []string       `json:"symptoms"`
	PossibleCauses []string       `json:"possible_causes"`
	DiagnosticSteps []RunbookStep `json:"diagnostic_steps"`
	ResolutionSteps []RunbookStep `json:"resolution_steps"`
	EscalationPath  []string      `json:"escalation_path"`
	Prevention      []string      `json:"prevention,omitempty"`
	RelatedRunbooks []string      `json:"related_runbooks,omitempty"`
}

// RunbookStep represents a step in a runbook
type RunbookStep struct {
	Order       int      `json:"order"`
	Description string   `json:"description"`
	Commands    []string `json:"commands,omitempty"`
	Expected    string   `json:"expected,omitempty"`
	Notes       string   `json:"notes,omitempty"`
}

// IncidentAnalysis represents the analysis of an incident
type IncidentAnalysis struct {
	IncidentID      string             `json:"incident_id"`
	Summary         string             `json:"summary"`
	Severity        string             `json:"severity"`
	Timeline        []TimelineEvent    `json:"timeline"`
	RootCause       string             `json:"root_cause"`
	Contributing    []string           `json:"contributing_factors"`
	Impact          ImpactAssessment   `json:"impact"`
	Resolution      string             `json:"resolution"`
	Recommendations []Recommendation   `json:"recommendations"`
	LessonsLearned  []string           `json:"lessons_learned"`
}

// TimelineEvent represents an event in incident timeline
type TimelineEvent struct {
	Timestamp   string `json:"timestamp"`
	Description string `json:"description"`
	Source      string `json:"source"` // "alert", "log", "metric", "user"
	Severity    string `json:"severity,omitempty"`
}

// ImpactAssessment describes the impact of an incident
type ImpactAssessment struct {
	Duration        string   `json:"duration"`
	AffectedUsers   string   `json:"affected_users"`
	AffectedRegions []string `json:"affected_regions,omitempty"`
	SLOImpact       string   `json:"slo_impact,omitempty"`
	FinancialImpact string   `json:"financial_impact,omitempty"`
}

// Recommendation for post-incident improvement
type Recommendation struct {
	Priority    string `json:"priority"` // "P0", "P1", "P2"
	Category    string `json:"category"` // "process", "tooling", "architecture", "monitoring"
	Description string `json:"description"`
	Owner       string `json:"owner,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
}

// OnCallConfig represents on-call configuration
type OnCallConfig struct {
	Platform    string           `json:"platform"` // "pagerduty", "opsgenie", "victorops"
	ServiceName string           `json:"service_name"`
	Schedules   []OnCallSchedule `json:"schedules"`
	Policies    []EscalationPolicy `json:"policies"`
	Files       map[string]string `json:"files"`
}

// OnCallSchedule defines an on-call rotation
type OnCallSchedule struct {
	Name         string   `json:"name"`
	TimeZone     string   `json:"time_zone"`
	RotationType string   `json:"rotation_type"` // "daily", "weekly", "custom"
	StartTime    string   `json:"start_time"`
	Duration     string   `json:"duration"`
	Participants []string `json:"participants"`
}

// EscalationPolicy defines escalation rules
type EscalationPolicy struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Rules       []EscalationRule  `json:"rules"`
}

// EscalationRule defines a single escalation step
type EscalationRule struct {
	Order        int      `json:"order"`
	DelayMinutes int      `json:"delay_minutes"`
	Targets      []string `json:"targets"` // schedules or users
	Type         string   `json:"type"`    // "schedule", "user", "team"
}

// ============================================================================
// Input Types
// ============================================================================

// MonitoringInput for GenerateMonitoringConfig
type MonitoringInput struct {
	ServiceName  string            `json:"service_name"`
	Platform     string            `json:"platform"`     // "prometheus", "datadog", "cloudwatch"
	TechStack    *TechStack        `json:"tech_stack,omitempty"`
	Endpoints    []string          `json:"endpoints,omitempty"`
	CustomMetrics []string         `json:"custom_metrics,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// AlertInput for GenerateAlertRules
type AlertInput struct {
	ServiceName string     `json:"service_name"`
	Platform    string     `json:"platform"`
	SLOs        []SLO      `json:"slos,omitempty"`
	Metrics     []string   `json:"metrics,omitempty"`
	Thresholds  map[string]float64 `json:"thresholds,omitempty"`
}

// DashboardInput for GenerateDashboards
type DashboardInput struct {
	ServiceName string            `json:"service_name"`
	Platform    string            `json:"platform"`    // "grafana", "datadog", "cloudwatch"
	Metrics     []string          `json:"metrics,omitempty"`
	SLOs        []SLO             `json:"slos,omitempty"`
	TechStack   *TechStack        `json:"tech_stack,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// RunbookInput for GenerateRunbooks
type RunbookInput struct {
	ServiceName    string       `json:"service_name"`
	Alerts         []AlertRule  `json:"alerts,omitempty"`
	Architecture   string       `json:"architecture,omitempty"`
	Dependencies   []string     `json:"dependencies,omitempty"`
	CommonIssues   []string     `json:"common_issues,omitempty"`
}

// IncidentInput for AnalyzeIncident
type IncidentInput struct {
	IncidentID  string            `json:"incident_id"`
	Description string            `json:"description"`
	Logs        []string          `json:"logs,omitempty"`
	Metrics     map[string]string `json:"metrics,omitempty"`
	Alerts      []string          `json:"alerts,omitempty"`
	Timeline    []TimelineEvent   `json:"timeline,omitempty"`
}

// OnCallInput for GenerateOnCallConfig
type OnCallInput struct {
	ServiceName   string   `json:"service_name"`
	Platform      string   `json:"platform"` // "pagerduty", "opsgenie"
	TeamMembers   []string `json:"team_members"`
	TimeZone      string   `json:"time_zone"`
	RotationType  string   `json:"rotation_type"`
	BusinessHours bool     `json:"business_hours"`
}

// ============================================================================
// Activity Methods
// ============================================================================

// GenerateMonitoringConfig generates monitoring configuration
func (a *SREAgent) GenerateMonitoringConfig(ctx context.Context, input map[string]interface{}) (*MonitoringConfig, error) {
	// Parse input
	inputJSON, _ := json.Marshal(input)
	var req MonitoringInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Set defaults
	if req.Platform == "" {
		req.Platform = "prometheus"
	}

	prompt := a.buildMonitoringPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an SRE expert. Generate monitoring configuration in valid JSON format matching the schema: " + GetMonitoringConfigSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var config MonitoringConfig
	if err := json.Unmarshal([]byte(resp.Response), &config); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate config files
	config.Files = a.generateMonitoringFiles(config, req)

	return &config, nil
}

// GenerateAlertRules generates alerting rules based on SLOs
func (a *SREAgent) GenerateAlertRules(ctx context.Context, input map[string]interface{}) (*AlertRules, error) {
	inputJSON, _ := json.Marshal(input)
	var req AlertInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if req.Platform == "" {
		req.Platform = "prometheus"
	}

	prompt := a.buildAlertPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an SRE expert. Generate alert rules in valid JSON format matching the schema: " + GetAlertRulesSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var rules AlertRules
	if err := json.Unmarshal([]byte(resp.Response), &rules); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate alert files
	rules.Files = a.generateAlertFiles(rules, req)

	return &rules, nil
}

// GenerateDashboards generates monitoring dashboards
func (a *SREAgent) GenerateDashboards(ctx context.Context, input map[string]interface{}) (*DashboardBundle, error) {
	inputJSON, _ := json.Marshal(input)
	var req DashboardInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if req.Platform == "" {
		req.Platform = "grafana"
	}

	prompt := a.buildDashboardPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an SRE expert. Generate dashboard configuration in valid JSON format matching the schema: " + GetDashboardBundleSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var bundle DashboardBundle
	if err := json.Unmarshal([]byte(resp.Response), &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate dashboard JSON files
	bundle.Files = a.generateDashboardFiles(bundle, req)

	return &bundle, nil
}

// GenerateRunbooks generates operational runbooks
func (a *SREAgent) GenerateRunbooks(ctx context.Context, input map[string]interface{}) (*RunbookBundle, error) {
	inputJSON, _ := json.Marshal(input)
	var req RunbookInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	prompt := a.buildRunbookPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an SRE expert. Generate operational runbooks in valid JSON format matching the schema: " + GetRunbookBundleSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var bundle RunbookBundle
	if err := json.Unmarshal([]byte(resp.Response), &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate runbook markdown files
	bundle.Files = a.generateRunbookFiles(bundle)

	return &bundle, nil
}

// AnalyzeIncident performs root cause analysis on an incident
func (a *SREAgent) AnalyzeIncident(ctx context.Context, input map[string]interface{}) (*IncidentAnalysis, error) {
	inputJSON, _ := json.Marshal(input)
	var req IncidentInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	prompt := a.buildIncidentAnalysisPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an SRE expert. Analyze the incident and provide root cause analysis in valid JSON format matching the schema: " + GetIncidentAnalysisSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var analysis IncidentAnalysis
	if err := json.Unmarshal([]byte(resp.Response), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	analysis.IncidentID = req.IncidentID

	return &analysis, nil
}

// GenerateOnCallConfig generates on-call schedules and escalation policies
func (a *SREAgent) GenerateOnCallConfig(ctx context.Context, input map[string]interface{}) (*OnCallConfig, error) {
	inputJSON, _ := json.Marshal(input)
	var req OnCallInput
	if err := json.Unmarshal(inputJSON, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if req.Platform == "" {
		req.Platform = "pagerduty"
	}
	if req.TimeZone == "" {
		req.TimeZone = "UTC"
	}
	if req.RotationType == "" {
		req.RotationType = "weekly"
	}

	prompt := a.buildOnCallPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an SRE expert. Generate on-call configuration in valid JSON format matching the schema: " + GetOnCallConfigSchema(),
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var config OnCallConfig
	if err := json.Unmarshal([]byte(resp.Response), &config); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate on-call config files
	config.Files = a.generateOnCallFiles(config, req)

	return &config, nil
}

// RefineSRE refines SRE configurations based on feedback
func (a *SREAgent) RefineSRE(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	configType, _ := input["config_type"].(string)
	currentConfig, _ := input["current_config"].(map[string]interface{})
	feedback, _ := input["feedback"].(string)

	configJSON, _ := json.MarshalIndent(currentConfig, "", "  ")

	prompt := fmt.Sprintf(`Refine the following %s configuration based on the feedback.

Current Configuration:
%s

Feedback:
%s

Provide an improved configuration that addresses the feedback while maintaining best practices.
Return the refined configuration in the same JSON structure.`, configType, string(configJSON), feedback)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are an SRE expert. Refine the configuration based on the feedback and return valid JSON.",
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM refinement failed: %w", err)
	}

	var refined map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Response), &refined); err != nil {
		return nil, fmt.Errorf("failed to parse refined config: %w", err)
	}

	return refined, nil
}

// ============================================================================
// Prompt Builders
// ============================================================================

func (a *SREAgent) buildMonitoringPrompt(req MonitoringInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`Generate comprehensive monitoring configuration for service "%s" using %s.

`, req.ServiceName, req.Platform))

	if req.TechStack != nil {
		sb.WriteString(fmt.Sprintf(`Tech Stack:
- Backend: %s
- Database: %s
- Cache: %s

`, req.TechStack.Backend.Framework, req.TechStack.Database.Primary, req.TechStack.Database.Cache))
	}

	if len(req.Endpoints) > 0 {
		sb.WriteString("Endpoints to monitor:\n")
		for _, ep := range req.Endpoints {
			sb.WriteString(fmt.Sprintf("- %s\n", ep))
		}
		sb.WriteString("\n")
	}

	if len(req.CustomMetrics) > 0 {
		sb.WriteString("Custom metrics to include:\n")
		for _, m := range req.CustomMetrics {
			sb.WriteString(fmt.Sprintf("- %s\n", m))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`Include the following metric categories:

1. **RED Metrics** (Request-oriented):
   - Rate: requests per second
   - Errors: error rate/count
   - Duration: request latency (p50, p95, p99)

2. **USE Metrics** (Resource-oriented):
   - Utilization: CPU, memory, disk usage
   - Saturation: queue depth, thread pool usage
   - Errors: system errors, OOM events

3. **Business Metrics**:
   - Active users
   - Transaction success rate
   - Feature usage

4. **Dependency Metrics**:
   - Database connection pool
   - External API latency
   - Cache hit rate

Generate appropriate scrape configs, recording rules, and metric definitions.`)

	return sb.String()
}

func (a *SREAgent) buildAlertPrompt(req AlertInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`Generate alert rules for service "%s" using %s format.

`, req.ServiceName, req.Platform))

	if len(req.SLOs) > 0 {
		sb.WriteString("Service Level Objectives:\n")
		for _, slo := range req.SLOs {
			sb.WriteString(fmt.Sprintf("- %s: %.2f%% %s over %s\n", slo.Name, slo.Target, slo.Indicator, slo.Window))
		}
		sb.WriteString("\n")
	}

	if len(req.Thresholds) > 0 {
		sb.WriteString("Custom thresholds:\n")
		for k, v := range req.Thresholds {
			sb.WriteString(fmt.Sprintf("- %s: %.2f\n", k, v))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`Generate alerts following these principles:

1. **Multi-window, multi-burn rate** for SLO-based alerts
2. **Severity levels**: critical (page), warning (ticket), info (log)
3. **Actionable alerts** with clear runbook links
4. **Avoid alert fatigue** by using appropriate thresholds and durations

Include alerts for:
- Availability: service down, high error rate
- Latency: p99 latency breach, slow endpoints
- Saturation: resource exhaustion warnings
- Dependencies: database, cache, external API issues
- Security: unusual traffic patterns, auth failures

Each alert should have:
- Clear name and description
- Appropriate severity
- Duration before firing
- Runbook reference
- Useful labels and annotations`)

	return sb.String()
}

func (a *SREAgent) buildDashboardPrompt(req DashboardInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`Generate %s dashboards for service "%s".

`, req.Platform, req.ServiceName))

	if len(req.SLOs) > 0 {
		sb.WriteString("SLOs to visualize:\n")
		for _, slo := range req.SLOs {
			sb.WriteString(fmt.Sprintf("- %s: %.2f%% %s\n", slo.Name, slo.Target, slo.Indicator))
		}
		sb.WriteString("\n")
	}

	if req.TechStack != nil {
		sb.WriteString(fmt.Sprintf("Tech stack: %s + %s\n\n",
			req.TechStack.Backend.Framework, req.TechStack.Database.Primary))
	}

	sb.WriteString(`Create the following dashboards:

1. **Service Overview Dashboard**:
   - Request rate, error rate, latency (RED metrics)
   - SLO burn rate and error budget
   - Active instances and health status
   - Key business metrics

2. **Infrastructure Dashboard**:
   - CPU, memory, disk utilization
   - Network I/O and connections
   - Container/pod metrics (if applicable)
   - Auto-scaling metrics

3. **Dependencies Dashboard**:
   - Database query performance
   - Connection pool metrics
   - Cache hit rates
   - External API latency

4. **SLO Dashboard**:
   - Error budget remaining
   - Burn rate visualization
   - Historical SLO compliance
   - Incident impact on SLOs

Each dashboard should have:
- Meaningful title and description
- Template variables for filtering
- Appropriate time ranges
- Thresholds for visual alerts
- Organized layout with rows/sections`)

	return sb.String()
}

func (a *SREAgent) buildRunbookPrompt(req RunbookInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`Generate operational runbooks for service "%s".

`, req.ServiceName))

	if len(req.Alerts) > 0 {
		sb.WriteString("Alerts requiring runbooks:\n")
		for _, alert := range req.Alerts {
			sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", alert.Name, alert.Severity, alert.Summary))
		}
		sb.WriteString("\n")
	}

	if req.Architecture != "" {
		sb.WriteString(fmt.Sprintf("Architecture: %s\n\n", req.Architecture))
	}

	if len(req.Dependencies) > 0 {
		sb.WriteString("Dependencies:\n")
		for _, dep := range req.Dependencies {
			sb.WriteString(fmt.Sprintf("- %s\n", dep))
		}
		sb.WriteString("\n")
	}

	if len(req.CommonIssues) > 0 {
		sb.WriteString("Common issues:\n")
		for _, issue := range req.CommonIssues {
			sb.WriteString(fmt.Sprintf("- %s\n", issue))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`Generate runbooks for:

1. **Service Unavailable** - Complete outage
2. **High Error Rate** - Elevated 5xx errors
3. **High Latency** - Response time degradation
4. **Resource Exhaustion** - CPU/memory/disk full
5. **Database Issues** - Connection failures, slow queries
6. **Dependency Failures** - External service outages

Each runbook should include:
- Clear title and severity
- Impact assessment
- Symptoms to look for
- Possible root causes
- Step-by-step diagnostic commands
- Resolution steps with commands
- Escalation path
- Prevention recommendations

Make diagnostic and resolution steps actionable with specific commands.`)

	return sb.String()
}

func (a *SREAgent) buildIncidentAnalysisPrompt(req IncidentInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`Analyze incident "%s" and provide root cause analysis.

Description: %s

`, req.IncidentID, req.Description))

	if len(req.Timeline) > 0 {
		sb.WriteString("Timeline:\n")
		for _, event := range req.Timeline {
			sb.WriteString(fmt.Sprintf("- %s: %s (%s)\n", event.Timestamp, event.Description, event.Source))
		}
		sb.WriteString("\n")
	}

	if len(req.Logs) > 0 {
		sb.WriteString("Relevant logs:\n```\n")
		for _, log := range req.Logs[:min(10, len(req.Logs))] {
			sb.WriteString(log + "\n")
		}
		sb.WriteString("```\n\n")
	}

	if len(req.Alerts) > 0 {
		sb.WriteString("Triggered alerts:\n")
		for _, alert := range req.Alerts {
			sb.WriteString(fmt.Sprintf("- %s\n", alert))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`Provide:

1. **Summary**: Brief description of what happened
2. **Timeline**: Key events in chronological order
3. **Root Cause**: The primary cause of the incident
4. **Contributing Factors**: Secondary factors that enabled or worsened the incident
5. **Impact Assessment**: Duration, affected users, SLO impact
6. **Resolution**: What fixed the issue
7. **Recommendations**: Prioritized action items (P0, P1, P2) to prevent recurrence
8. **Lessons Learned**: Key takeaways for the team

Focus on actionable recommendations with clear ownership and categories (process, tooling, architecture, monitoring).`)

	return sb.String()
}

func (a *SREAgent) buildOnCallPrompt(req OnCallInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`Generate on-call configuration for service "%s" using %s.

Team members: %s
Time zone: %s
Rotation type: %s
Business hours only: %v

`, req.ServiceName, req.Platform, strings.Join(req.TeamMembers, ", "), req.TimeZone, req.RotationType, req.BusinessHours))

	sb.WriteString(`Generate:

1. **Primary On-Call Schedule**:
   - Regular rotation among team members
   - Handoff times that work for the team

2. **Secondary/Backup Schedule**:
   - Backup coverage if primary doesn't respond
   - Shadow rotation for new team members

3. **Escalation Policy**:
   - Primary → Secondary → Team Lead → Manager
   - Appropriate delays between escalation levels
   - Different policies for critical vs warning alerts

4. **Coverage Rules**:
   - Business hours vs 24/7 coverage
   - Weekend and holiday handling
   - Override support for PTO

Ensure fair rotation and adequate rest time between on-call shifts.`)

	return sb.String()
}

// ============================================================================
// File Generators
// ============================================================================

func (a *SREAgent) generateMonitoringFiles(config MonitoringConfig, req MonitoringInput) map[string]string {
	files := make(map[string]string)

	switch req.Platform {
	case "prometheus":
		files["prometheus/prometheus.yml"] = a.generatePrometheusConfig(config)
		files["prometheus/recording_rules.yml"] = a.generateRecordingRules(config)
		files["prometheus/alerts.yml"] = "" // Placeholder for alerts
	case "datadog":
		files["datadog/datadog.yaml"] = a.generateDatadogConfig(config)
	case "cloudwatch":
		files["cloudwatch/dashboard.json"] = a.generateCloudWatchConfig(config)
	}

	return files
}

func (a *SREAgent) generatePrometheusConfig(config MonitoringConfig) string {
	var sb strings.Builder

	sb.WriteString("# Prometheus Configuration\n")
	sb.WriteString("# Generated by SRE Agent\n\n")

	sb.WriteString("global:\n")
	sb.WriteString("  scrape_interval: 15s\n")
	sb.WriteString("  evaluation_interval: 15s\n\n")

	if len(config.Labels) > 0 {
		sb.WriteString("  external_labels:\n")
		for k, v := range config.Labels {
			sb.WriteString(fmt.Sprintf("    %s: %s\n", k, v))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("rule_files:\n")
	sb.WriteString("  - 'recording_rules.yml'\n")
	sb.WriteString("  - 'alerts.yml'\n\n")

	sb.WriteString("scrape_configs:\n")
	for _, sc := range config.ScrapeConfigs {
		sb.WriteString(fmt.Sprintf("  - job_name: '%s'\n", sc.JobName))
		sb.WriteString(fmt.Sprintf("    scrape_interval: %s\n", sc.ScrapeInterval))
		if sc.MetricsPath != "" {
			sb.WriteString(fmt.Sprintf("    metrics_path: %s\n", sc.MetricsPath))
		}
		sb.WriteString("    static_configs:\n")
		for _, static := range sc.StaticConfigs {
			sb.WriteString("      - targets:\n")
			for _, target := range static.Targets {
				sb.WriteString(fmt.Sprintf("          - '%s'\n", target))
			}
			if len(static.Labels) > 0 {
				sb.WriteString("        labels:\n")
				for k, v := range static.Labels {
					sb.WriteString(fmt.Sprintf("          %s: %s\n", k, v))
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (a *SREAgent) generateRecordingRules(config MonitoringConfig) string {
	var sb strings.Builder

	sb.WriteString("# Recording Rules\n")
	sb.WriteString("# Generated by SRE Agent\n\n")

	sb.WriteString("groups:\n")
	sb.WriteString(fmt.Sprintf("  - name: %s_recording_rules\n", sanitizeK8sName(config.ServiceName)))
	sb.WriteString("    rules:\n")

	for _, rule := range config.RecordingRules {
		sb.WriteString(fmt.Sprintf("      - record: %s\n", rule.Record))
		sb.WriteString(fmt.Sprintf("        expr: %s\n", rule.Expr))
		if len(rule.Labels) > 0 {
			sb.WriteString("        labels:\n")
			for k, v := range rule.Labels {
				sb.WriteString(fmt.Sprintf("          %s: %s\n", k, v))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (a *SREAgent) generateDatadogConfig(config MonitoringConfig) string {
	var sb strings.Builder

	sb.WriteString("# Datadog Agent Configuration\n")
	sb.WriteString("# Generated by SRE Agent\n\n")

	sb.WriteString("api_key: ${DD_API_KEY}\n")
	sb.WriteString("site: datadoghq.com\n\n")

	sb.WriteString("logs_enabled: true\n")
	sb.WriteString("apm_config:\n")
	sb.WriteString("  enabled: true\n\n")

	sb.WriteString("process_config:\n")
	sb.WriteString("  enabled: true\n\n")

	sb.WriteString(fmt.Sprintf("tags:\n"))
	sb.WriteString(fmt.Sprintf("  - service:%s\n", config.ServiceName))
	for k, v := range config.Labels {
		sb.WriteString(fmt.Sprintf("  - %s:%s\n", k, v))
	}

	return sb.String()
}

func (a *SREAgent) generateCloudWatchConfig(config MonitoringConfig) string {
	cwConfig := map[string]interface{}{
		"widgets": []map[string]interface{}{},
	}

	configJSON, _ := json.MarshalIndent(cwConfig, "", "  ")
	return string(configJSON)
}

func (a *SREAgent) generateAlertFiles(rules AlertRules, req AlertInput) map[string]string {
	files := make(map[string]string)

	switch req.Platform {
	case "prometheus":
		files["prometheus/alerts.yml"] = a.generatePrometheusAlerts(rules)
	case "datadog":
		files["datadog/monitors.json"] = a.generateDatadogMonitors(rules)
	case "pagerduty":
		files["pagerduty/services.json"] = a.generatePagerDutyConfig(rules)
	}

	return files
}

func (a *SREAgent) generatePrometheusAlerts(rules AlertRules) string {
	var sb strings.Builder

	sb.WriteString("# Prometheus Alert Rules\n")
	sb.WriteString("# Generated by SRE Agent\n\n")

	sb.WriteString("groups:\n")
	sb.WriteString(fmt.Sprintf("  - name: %s_alerts\n", sanitizeK8sName(rules.ServiceName)))
	sb.WriteString("    rules:\n")

	for _, alert := range rules.Alerts {
		sb.WriteString(fmt.Sprintf("      - alert: %s\n", alert.Name))
		sb.WriteString(fmt.Sprintf("        expr: %s\n", alert.Condition))
		sb.WriteString(fmt.Sprintf("        for: %s\n", alert.Duration))
		sb.WriteString("        labels:\n")
		sb.WriteString(fmt.Sprintf("          severity: %s\n", alert.Severity))
		for k, v := range alert.Labels {
			sb.WriteString(fmt.Sprintf("          %s: %s\n", k, v))
		}
		sb.WriteString("        annotations:\n")
		sb.WriteString(fmt.Sprintf("          summary: %s\n", alert.Summary))
		sb.WriteString(fmt.Sprintf("          description: %s\n", alert.Description))
		if alert.Runbook != "" {
			sb.WriteString(fmt.Sprintf("          runbook_url: %s\n", alert.Runbook))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (a *SREAgent) generateDatadogMonitors(rules AlertRules) string {
	monitors := []map[string]interface{}{}

	for _, alert := range rules.Alerts {
		monitor := map[string]interface{}{
			"name":    alert.Name,
			"type":    "metric alert",
			"query":   alert.Condition,
			"message": fmt.Sprintf("%s\n\n%s", alert.Summary, alert.Description),
			"tags":    []string{fmt.Sprintf("service:%s", rules.ServiceName), fmt.Sprintf("severity:%s", alert.Severity)},
			"options": map[string]interface{}{
				"notify_no_data":    false,
				"renotify_interval": 60,
				"escalation_message": alert.Description,
			},
		}
		monitors = append(monitors, monitor)
	}

	result, _ := json.MarshalIndent(monitors, "", "  ")
	return string(result)
}

func (a *SREAgent) generatePagerDutyConfig(rules AlertRules) string {
	config := map[string]interface{}{
		"service": map[string]interface{}{
			"name":        rules.ServiceName,
			"description": fmt.Sprintf("Alerts for %s", rules.ServiceName),
			"escalation_policy": map[string]interface{}{
				"type": "escalation_policy_reference",
				"id":   "${ESCALATION_POLICY_ID}",
			},
		},
	}

	result, _ := json.MarshalIndent(config, "", "  ")
	return string(result)
}

func (a *SREAgent) generateDashboardFiles(bundle DashboardBundle, req DashboardInput) map[string]string {
	files := make(map[string]string)

	switch req.Platform {
	case "grafana":
		for _, dashboard := range bundle.Dashboards {
			fileName := fmt.Sprintf("grafana/dashboards/%s.json", sanitizeK8sName(dashboard.Name))
			files[fileName] = a.generateGrafanaDashboard(dashboard, req.ServiceName)
		}
	case "datadog":
		for _, dashboard := range bundle.Dashboards {
			fileName := fmt.Sprintf("datadog/dashboards/%s.json", sanitizeK8sName(dashboard.Name))
			files[fileName] = a.generateDatadogDashboard(dashboard)
		}
	}

	return files
}

func (a *SREAgent) generateGrafanaDashboard(dashboard Dashboard, serviceName string) string {
	panels := []map[string]interface{}{}

	yPos := 0
	for i, panel := range dashboard.Panels {
		grafanaPanel := map[string]interface{}{
			"id":    i + 1,
			"title": panel.Title,
			"type":  mapPanelType(panel.Type),
			"gridPos": map[string]int{
				"h": 8,
				"w": 12,
				"x": (i % 2) * 12,
				"y": yPos,
			},
			"targets": []map[string]interface{}{
				{
					"expr":         panel.Query,
					"legendFormat": "{{instance}}",
					"refId":        "A",
				},
			},
		}

		if panel.Unit != "" {
			grafanaPanel["fieldConfig"] = map[string]interface{}{
				"defaults": map[string]interface{}{
					"unit": panel.Unit,
				},
			}
		}

		panels = append(panels, grafanaPanel)

		if i%2 == 1 {
			yPos += 8
		}
	}

	variables := []map[string]interface{}{}
	for _, v := range dashboard.Variables {
		variable := map[string]interface{}{
			"name":  v.Name,
			"label": v.Label,
			"type":  v.Type,
		}
		if v.Query != "" {
			variable["query"] = v.Query
		}
		if v.Default != "" {
			variable["current"] = map[string]string{"value": v.Default}
		}
		variables = append(variables, variable)
	}

	grafanaDashboard := map[string]interface{}{
		"title":       dashboard.Title,
		"description": dashboard.Description,
		"tags":        append(dashboard.Tags, serviceName),
		"timezone":    "browser",
		"refresh":     dashboard.Refresh,
		"time": map[string]string{
			"from": "now-1h",
			"to":   "now",
		},
		"panels": panels,
		"templating": map[string]interface{}{
			"list": variables,
		},
		"schemaVersion": 38,
		"version":       1,
	}

	result, _ := json.MarshalIndent(grafanaDashboard, "", "  ")
	return string(result)
}

func (a *SREAgent) generateDatadogDashboard(dashboard Dashboard) string {
	widgets := []map[string]interface{}{}

	for i, panel := range dashboard.Panels {
		widget := map[string]interface{}{
			"definition": map[string]interface{}{
				"title": panel.Title,
				"type":  mapDatadogWidgetType(panel.Type),
				"requests": []map[string]interface{}{
					{
						"q":            panel.Query,
						"display_type": "line",
					},
				},
			},
			"layout": map[string]int{
				"x":      (i % 3) * 4,
				"y":      (i / 3) * 3,
				"width":  4,
				"height": 3,
			},
		}
		widgets = append(widgets, widget)
	}

	ddDashboard := map[string]interface{}{
		"title":       dashboard.Title,
		"description": dashboard.Description,
		"widgets":     widgets,
		"layout_type": "ordered",
	}

	result, _ := json.MarshalIndent(ddDashboard, "", "  ")
	return string(result)
}

func (a *SREAgent) generateRunbookFiles(bundle RunbookBundle) map[string]string {
	files := make(map[string]string)

	for _, runbook := range bundle.Runbooks {
		fileName := fmt.Sprintf("runbooks/%s.md", sanitizeK8sName(runbook.Name))
		files[fileName] = a.generateRunbookMarkdown(runbook)
	}

	// Generate index
	files["runbooks/README.md"] = a.generateRunbookIndex(bundle)

	return files
}

func (a *SREAgent) generateRunbookMarkdown(runbook Runbook) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", runbook.Title))

	sb.WriteString(fmt.Sprintf("**Severity:** %s\n\n", runbook.Severity))
	sb.WriteString(fmt.Sprintf("**Alert:** %s\n\n", runbook.AlertName))

	sb.WriteString("## Description\n\n")
	sb.WriteString(runbook.Description + "\n\n")

	sb.WriteString("## Impact\n\n")
	sb.WriteString(runbook.Impact + "\n\n")

	sb.WriteString("## Symptoms\n\n")
	for _, symptom := range runbook.Symptoms {
		sb.WriteString(fmt.Sprintf("- %s\n", symptom))
	}
	sb.WriteString("\n")

	sb.WriteString("## Possible Causes\n\n")
	for _, cause := range runbook.PossibleCauses {
		sb.WriteString(fmt.Sprintf("- %s\n", cause))
	}
	sb.WriteString("\n")

	sb.WriteString("## Diagnostic Steps\n\n")
	for _, step := range runbook.DiagnosticSteps {
		sb.WriteString(fmt.Sprintf("### Step %d: %s\n\n", step.Order, step.Description))
		if len(step.Commands) > 0 {
			sb.WriteString("```bash\n")
			for _, cmd := range step.Commands {
				sb.WriteString(cmd + "\n")
			}
			sb.WriteString("```\n\n")
		}
		if step.Expected != "" {
			sb.WriteString(fmt.Sprintf("**Expected:** %s\n\n", step.Expected))
		}
		if step.Notes != "" {
			sb.WriteString(fmt.Sprintf("> **Note:** %s\n\n", step.Notes))
		}
	}

	sb.WriteString("## Resolution Steps\n\n")
	for _, step := range runbook.ResolutionSteps {
		sb.WriteString(fmt.Sprintf("### Step %d: %s\n\n", step.Order, step.Description))
		if len(step.Commands) > 0 {
			sb.WriteString("```bash\n")
			for _, cmd := range step.Commands {
				sb.WriteString(cmd + "\n")
			}
			sb.WriteString("```\n\n")
		}
		if step.Notes != "" {
			sb.WriteString(fmt.Sprintf("> **Note:** %s\n\n", step.Notes))
		}
	}

	sb.WriteString("## Escalation Path\n\n")
	for i, level := range runbook.EscalationPath {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, level))
	}
	sb.WriteString("\n")

	if len(runbook.Prevention) > 0 {
		sb.WriteString("## Prevention\n\n")
		for _, item := range runbook.Prevention {
			sb.WriteString(fmt.Sprintf("- %s\n", item))
		}
		sb.WriteString("\n")
	}

	if len(runbook.RelatedRunbooks) > 0 {
		sb.WriteString("## Related Runbooks\n\n")
		for _, related := range runbook.RelatedRunbooks {
			sb.WriteString(fmt.Sprintf("- [%s](./%s.md)\n", related, sanitizeK8sName(related)))
		}
	}

	return sb.String()
}

func (a *SREAgent) generateRunbookIndex(bundle RunbookBundle) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Runbooks for %s\n\n", bundle.ServiceName))
	sb.WriteString("This directory contains operational runbooks for incident response.\n\n")

	sb.WriteString("## Runbook Index\n\n")
	sb.WriteString("| Runbook | Severity | Alert |\n")
	sb.WriteString("|---------|----------|-------|\n")

	for _, runbook := range bundle.Runbooks {
		sb.WriteString(fmt.Sprintf("| [%s](./%s.md) | %s | %s |\n",
			runbook.Title,
			sanitizeK8sName(runbook.Name),
			runbook.Severity,
			runbook.AlertName))
	}

	sb.WriteString("\n## Usage\n\n")
	sb.WriteString("1. Identify the triggering alert\n")
	sb.WriteString("2. Find the corresponding runbook\n")
	sb.WriteString("3. Follow diagnostic steps to identify root cause\n")
	sb.WriteString("4. Apply resolution steps\n")
	sb.WriteString("5. Escalate if resolution fails\n")
	sb.WriteString("6. Document incident for post-mortem\n")

	return sb.String()
}

func (a *SREAgent) generateOnCallFiles(config OnCallConfig, req OnCallInput) map[string]string {
	files := make(map[string]string)

	switch req.Platform {
	case "pagerduty":
		files["pagerduty/schedules.json"] = a.generatePagerDutySchedules(config)
		files["pagerduty/escalation_policies.json"] = a.generatePagerDutyEscalation(config)
	case "opsgenie":
		files["opsgenie/schedules.json"] = a.generateOpsgenieSchedules(config)
		files["opsgenie/escalation.json"] = a.generateOpsgenieEscalation(config)
	}

	return files
}

func (a *SREAgent) generatePagerDutySchedules(config OnCallConfig) string {
	schedules := []map[string]interface{}{}

	for _, schedule := range config.Schedules {
		pdSchedule := map[string]interface{}{
			"schedule": map[string]interface{}{
				"name":        schedule.Name,
				"time_zone":   schedule.TimeZone,
				"description": fmt.Sprintf("On-call schedule for %s", config.ServiceName),
				"schedule_layers": []map[string]interface{}{
					{
						"name":                         "Primary",
						"start":                        schedule.StartTime,
						"rotation_virtual_start":       schedule.StartTime,
						"rotation_turn_length_seconds": parseDurationToSeconds(schedule.Duration),
						"users":                        formatPagerDutyUsers(schedule.Participants),
					},
				},
			},
		}
		schedules = append(schedules, pdSchedule)
	}

	result, _ := json.MarshalIndent(schedules, "", "  ")
	return string(result)
}

func (a *SREAgent) generatePagerDutyEscalation(config OnCallConfig) string {
	policies := []map[string]interface{}{}

	for _, policy := range config.Policies {
		rules := []map[string]interface{}{}
		for _, rule := range policy.Rules {
			pdRule := map[string]interface{}{
				"escalation_delay_in_minutes": rule.DelayMinutes,
				"targets":                     formatPagerDutyTargets(rule.Targets, rule.Type),
			}
			rules = append(rules, pdRule)
		}

		pdPolicy := map[string]interface{}{
			"escalation_policy": map[string]interface{}{
				"name":             policy.Name,
				"description":      policy.Description,
				"escalation_rules": rules,
				"num_loops":        1,
			},
		}
		policies = append(policies, pdPolicy)
	}

	result, _ := json.MarshalIndent(policies, "", "  ")
	return string(result)
}

func (a *SREAgent) generateOpsgenieSchedules(config OnCallConfig) string {
	schedules := []map[string]interface{}{}

	for _, schedule := range config.Schedules {
		ogSchedule := map[string]interface{}{
			"name":     schedule.Name,
			"timezone": schedule.TimeZone,
			"enabled":  true,
			"rotations": []map[string]interface{}{
				{
					"name":        "Primary Rotation",
					"startDate":   schedule.StartTime,
					"type":        schedule.RotationType,
					"length":      1,
					"participants": formatOpsgenieParticipants(schedule.Participants),
				},
			},
		}
		schedules = append(schedules, ogSchedule)
	}

	result, _ := json.MarshalIndent(schedules, "", "  ")
	return string(result)
}

func (a *SREAgent) generateOpsgenieEscalation(config OnCallConfig) string {
	policies := []map[string]interface{}{}

	for _, policy := range config.Policies {
		rules := []map[string]interface{}{}
		for _, rule := range policy.Rules {
			ogRule := map[string]interface{}{
				"condition":  "if-not-acked",
				"notifyType": "default",
				"delay": map[string]interface{}{
					"timeAmount": rule.DelayMinutes,
					"timeUnit":   "minutes",
				},
				"recipient": map[string]interface{}{
					"type": rule.Type,
					"id":   rule.Targets[0],
				},
			}
			rules = append(rules, ogRule)
		}

		ogPolicy := map[string]interface{}{
			"name":        policy.Name,
			"description": policy.Description,
			"rules":       rules,
		}
		policies = append(policies, ogPolicy)
	}

	result, _ := json.MarshalIndent(policies, "", "  ")
	return string(result)
}

// ============================================================================
// JSON Schemas
// ============================================================================

// GetMonitoringConfigSchema returns JSON schema for monitoring config
func GetMonitoringConfigSchema() string {
	return `{
  "type": "object",
  "properties": {
    "platform": { "type": "string" },
    "service_name": { "type": "string" },
    "metrics": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "type": { "type": "string" },
          "description": { "type": "string" },
          "labels": { "type": "array", "items": { "type": "string" } }
        },
        "required": ["name", "type", "description"]
      }
    },
    "scrape_configs": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "job_name": { "type": "string" },
          "scrape_interval": { "type": "string" },
          "metrics_path": { "type": "string" },
          "static_configs": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "targets": { "type": "array", "items": { "type": "string" } },
                "labels": { "type": "object" }
              }
            }
          }
        },
        "required": ["job_name", "scrape_interval"]
      }
    },
    "recording_rules": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "record": { "type": "string" },
          "expr": { "type": "string" },
          "labels": { "type": "object" }
        },
        "required": ["record", "expr"]
      }
    },
    "labels": { "type": "object" }
  },
  "required": ["platform", "service_name", "metrics"]
}`
}

// GetAlertRulesSchema returns JSON schema for alert rules
func GetAlertRulesSchema() string {
	return `{
  "type": "object",
  "properties": {
    "platform": { "type": "string" },
    "service_name": { "type": "string" },
    "slos": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "description": { "type": "string" },
          "target": { "type": "number" },
          "window": { "type": "string" },
          "indicator": { "type": "string" },
          "query": { "type": "string" }
        },
        "required": ["name", "target", "indicator"]
      }
    },
    "alerts": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "severity": { "type": "string" },
          "condition": { "type": "string" },
          "duration": { "type": "string" },
          "summary": { "type": "string" },
          "description": { "type": "string" },
          "runbook": { "type": "string" },
          "labels": { "type": "object" },
          "annotations": { "type": "object" }
        },
        "required": ["name", "severity", "condition", "duration", "summary"]
      }
    }
  },
  "required": ["platform", "service_name", "alerts"]
}`
}

// GetDashboardBundleSchema returns JSON schema for dashboard bundle
func GetDashboardBundleSchema() string {
	return `{
  "type": "object",
  "properties": {
    "platform": { "type": "string" },
    "service_name": { "type": "string" },
    "dashboards": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "title": { "type": "string" },
          "description": { "type": "string" },
          "tags": { "type": "array", "items": { "type": "string" } },
          "panels": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "title": { "type": "string" },
                "type": { "type": "string" },
                "query": { "type": "string" },
                "description": { "type": "string" },
                "unit": { "type": "string" }
              },
              "required": ["title", "type", "query"]
            }
          },
          "variables": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "name": { "type": "string" },
                "label": { "type": "string" },
                "type": { "type": "string" },
                "query": { "type": "string" },
                "default": { "type": "string" }
              },
              "required": ["name", "type"]
            }
          },
          "refresh": { "type": "string" },
          "time_range": { "type": "string" }
        },
        "required": ["name", "title", "panels"]
      }
    }
  },
  "required": ["platform", "service_name", "dashboards"]
}`
}

// GetRunbookBundleSchema returns JSON schema for runbook bundle
func GetRunbookBundleSchema() string {
	return `{
  "type": "object",
  "properties": {
    "service_name": { "type": "string" },
    "runbooks": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "title": { "type": "string" },
          "description": { "type": "string" },
          "alert_name": { "type": "string" },
          "severity": { "type": "string" },
          "impact": { "type": "string" },
          "symptoms": { "type": "array", "items": { "type": "string" } },
          "possible_causes": { "type": "array", "items": { "type": "string" } },
          "diagnostic_steps": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "order": { "type": "integer" },
                "description": { "type": "string" },
                "commands": { "type": "array", "items": { "type": "string" } },
                "expected": { "type": "string" },
                "notes": { "type": "string" }
              },
              "required": ["order", "description"]
            }
          },
          "resolution_steps": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "order": { "type": "integer" },
                "description": { "type": "string" },
                "commands": { "type": "array", "items": { "type": "string" } },
                "notes": { "type": "string" }
              },
              "required": ["order", "description"]
            }
          },
          "escalation_path": { "type": "array", "items": { "type": "string" } },
          "prevention": { "type": "array", "items": { "type": "string" } },
          "related_runbooks": { "type": "array", "items": { "type": "string" } }
        },
        "required": ["name", "title", "description", "severity", "impact", "symptoms", "diagnostic_steps", "resolution_steps", "escalation_path"]
      }
    }
  },
  "required": ["service_name", "runbooks"]
}`
}

// GetIncidentAnalysisSchema returns JSON schema for incident analysis
func GetIncidentAnalysisSchema() string {
	return `{
  "type": "object",
  "properties": {
    "summary": { "type": "string" },
    "severity": { "type": "string" },
    "timeline": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "timestamp": { "type": "string" },
          "description": { "type": "string" },
          "source": { "type": "string" },
          "severity": { "type": "string" }
        },
        "required": ["timestamp", "description", "source"]
      }
    },
    "root_cause": { "type": "string" },
    "contributing_factors": { "type": "array", "items": { "type": "string" } },
    "impact": {
      "type": "object",
      "properties": {
        "duration": { "type": "string" },
        "affected_users": { "type": "string" },
        "affected_regions": { "type": "array", "items": { "type": "string" } },
        "slo_impact": { "type": "string" },
        "financial_impact": { "type": "string" }
      },
      "required": ["duration", "affected_users"]
    },
    "resolution": { "type": "string" },
    "recommendations": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "priority": { "type": "string" },
          "category": { "type": "string" },
          "description": { "type": "string" },
          "owner": { "type": "string" },
          "due_date": { "type": "string" }
        },
        "required": ["priority", "category", "description"]
      }
    },
    "lessons_learned": { "type": "array", "items": { "type": "string" } }
  },
  "required": ["summary", "severity", "root_cause", "impact", "resolution", "recommendations"]
}`
}

// GetOnCallConfigSchema returns JSON schema for on-call config
func GetOnCallConfigSchema() string {
	return `{
  "type": "object",
  "properties": {
    "platform": { "type": "string" },
    "service_name": { "type": "string" },
    "schedules": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "time_zone": { "type": "string" },
          "rotation_type": { "type": "string" },
          "start_time": { "type": "string" },
          "duration": { "type": "string" },
          "participants": { "type": "array", "items": { "type": "string" } }
        },
        "required": ["name", "time_zone", "rotation_type", "participants"]
      }
    },
    "policies": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "description": { "type": "string" },
          "rules": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "order": { "type": "integer" },
                "delay_minutes": { "type": "integer" },
                "targets": { "type": "array", "items": { "type": "string" } },
                "type": { "type": "string" }
              },
              "required": ["order", "delay_minutes", "targets", "type"]
            }
          }
        },
        "required": ["name", "rules"]
      }
    }
  },
  "required": ["platform", "service_name", "schedules", "policies"]
}`
}

// ============================================================================
// Helper Functions
// ============================================================================

func mapPanelType(panelType string) string {
	mapping := map[string]string{
		"graph":   "timeseries",
		"stat":    "stat",
		"gauge":   "gauge",
		"table":   "table",
		"heatmap": "heatmap",
		"logs":    "logs",
	}
	if mapped, ok := mapping[panelType]; ok {
		return mapped
	}
	return "timeseries"
}

func mapDatadogWidgetType(panelType string) string {
	mapping := map[string]string{
		"graph":   "timeseries",
		"stat":    "query_value",
		"gauge":   "gauge",
		"table":   "query_table",
		"heatmap": "heatmap",
		"logs":    "log_stream",
	}
	if mapped, ok := mapping[panelType]; ok {
		return mapped
	}
	return "timeseries"
}

func parseDurationToSeconds(duration string) int {
	// Simple parsing - handles common formats like "1d", "7d", "8h"
	if len(duration) < 2 {
		return 604800 // default 1 week
	}

	unit := duration[len(duration)-1]
	value := 0
	n, _ := fmt.Sscanf(duration[:len(duration)-1], "%d", &value)

	// If parsing failed or value is 0, return default
	if n == 0 || value <= 0 {
		return 604800 // default 1 week
	}

	switch unit {
	case 'h':
		return value * 3600
	case 'd':
		return value * 86400
	case 'w':
		return value * 604800
	default:
		return 604800
	}
}

func formatPagerDutyUsers(users []string) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, user := range users {
		result = append(result, map[string]interface{}{
			"user": map[string]interface{}{
				"type": "user_reference",
				"id":   fmt.Sprintf("${USER_ID_%s}", sanitizeK8sName(user)),
			},
		})
	}
	return result
}

func formatPagerDutyTargets(targets []string, targetType string) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, target := range targets {
		result = append(result, map[string]interface{}{
			"type": targetType + "_reference",
			"id":   fmt.Sprintf("${%s_ID_%s}", strings.ToUpper(targetType), sanitizeK8sName(target)),
		})
	}
	return result
}

func formatOpsgenieParticipants(users []string) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, user := range users {
		result = append(result, map[string]interface{}{
			"type":     "user",
			"username": user,
		})
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
