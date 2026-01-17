package activities

import (
	"testing"
)

// ============================================================================
// Constructor Tests
// ============================================================================

func TestNewSREAgent(t *testing.T) {
	agent := NewSREAgent()
	if agent == nil {
		t.Fatal("NewSREAgent returned nil")
	}
	if agent.llmClient == nil {
		t.Fatal("LLM client not initialized")
	}
}

// ============================================================================
// Data Structure Tests
// ============================================================================

func TestMonitoringConfigStructure(t *testing.T) {
	config := MonitoringConfig{
		Platform:    "prometheus",
		ServiceName: "test-service",
		Metrics: []MetricDefinition{
			{
				Name:        "http_requests_total",
				Type:        "counter",
				Description: "Total HTTP requests",
				Labels:      []string{"method", "status"},
			},
			{
				Name:        "http_request_duration_seconds",
				Type:        "histogram",
				Description: "HTTP request latency",
				Buckets:     []float64{0.01, 0.05, 0.1, 0.5, 1.0},
			},
		},
		ScrapeConfigs: []ScrapeConfig{
			{
				JobName:        "app",
				ScrapeInterval: "15s",
				MetricsPath:    "/metrics",
				StaticConfigs: []StaticConfig{
					{
						Targets: []string{"localhost:8080"},
						Labels:  map[string]string{"env": "production"},
					},
				},
			},
		},
		RecordingRules: []RecordingRule{
			{
				Record: "job:http_requests:rate5m",
				Expr:   "sum(rate(http_requests_total[5m])) by (job)",
			},
		},
		Labels: map[string]string{
			"team": "platform",
		},
		Files: make(map[string]string),
	}

	if config.Platform != "prometheus" {
		t.Errorf("Expected platform prometheus, got %s", config.Platform)
	}
	if len(config.Metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(config.Metrics))
	}
	if config.Metrics[0].Type != "counter" {
		t.Errorf("Expected counter type, got %s", config.Metrics[0].Type)
	}
	if len(config.Metrics[1].Buckets) != 5 {
		t.Errorf("Expected 5 buckets, got %d", len(config.Metrics[1].Buckets))
	}
}

func TestAlertRulesStructure(t *testing.T) {
	rules := AlertRules{
		Platform:    "prometheus",
		ServiceName: "test-service",
		SLOs: []SLO{
			{
				Name:      "availability",
				Target:    99.9,
				Window:    "30d",
				Indicator: "availability",
				Query:     "sum(rate(http_requests_total{status!~\"5..\"}[30d])) / sum(rate(http_requests_total[30d]))",
			},
		},
		Alerts: []AlertRule{
			{
				Name:        "HighErrorRate",
				Severity:    "critical",
				Condition:   "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m])) > 0.01",
				Duration:    "5m",
				Summary:     "High error rate detected",
				Description: "Error rate is above 1%",
				Runbook:     "https://runbooks.example.com/high-error-rate",
				Labels:      map[string]string{"team": "platform"},
			},
		},
	}

	if len(rules.SLOs) != 1 {
		t.Errorf("Expected 1 SLO, got %d", len(rules.SLOs))
	}
	if rules.SLOs[0].Target != 99.9 {
		t.Errorf("Expected SLO target 99.9, got %f", rules.SLOs[0].Target)
	}
	if len(rules.Alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(rules.Alerts))
	}
	if rules.Alerts[0].Severity != "critical" {
		t.Errorf("Expected critical severity, got %s", rules.Alerts[0].Severity)
	}
}

func TestDashboardBundleStructure(t *testing.T) {
	bundle := DashboardBundle{
		Platform:    "grafana",
		ServiceName: "test-service",
		Dashboards: []Dashboard{
			{
				Name:        "service-overview",
				Title:       "Service Overview",
				Description: "Overview dashboard for service metrics",
				Tags:        []string{"sre", "overview"},
				Panels: []DashboardPanel{
					{
						Title:       "Request Rate",
						Type:        "graph",
						Query:       "sum(rate(http_requests_total[5m]))",
						Description: "HTTP requests per second",
						Unit:        "reqps",
						Thresholds: []Threshold{
							{Value: 100, Color: "yellow"},
							{Value: 500, Color: "red"},
						},
					},
				},
				Variables: []DashboardVar{
					{
						Name:    "env",
						Label:   "Environment",
						Type:    "custom",
						Options: []string{"production", "staging", "development"},
						Default: "production",
					},
				},
				Refresh:   "30s",
				TimeRange: "1h",
			},
		},
	}

	if len(bundle.Dashboards) != 1 {
		t.Errorf("Expected 1 dashboard, got %d", len(bundle.Dashboards))
	}
	if len(bundle.Dashboards[0].Panels) != 1 {
		t.Errorf("Expected 1 panel, got %d", len(bundle.Dashboards[0].Panels))
	}
	if len(bundle.Dashboards[0].Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(bundle.Dashboards[0].Variables))
	}
}

func TestRunbookStructure(t *testing.T) {
	runbook := Runbook{
		Name:        "high-error-rate",
		Title:       "High Error Rate Runbook",
		Description: "Steps to diagnose and resolve high error rates",
		AlertName:   "HighErrorRate",
		Severity:    "critical",
		Impact:      "Users may experience errors when accessing the service",
		Symptoms: []string{
			"Alert firing for HighErrorRate",
			"Users reporting 500 errors",
			"Increased latency",
		},
		PossibleCauses: []string{
			"Backend service failure",
			"Database connection issues",
			"Resource exhaustion",
		},
		DiagnosticSteps: []RunbookStep{
			{
				Order:       1,
				Description: "Check service logs for errors",
				Commands:    []string{"kubectl logs -l app=myservice --tail=100"},
				Expected:    "Look for error messages or stack traces",
			},
			{
				Order:       2,
				Description: "Check database connectivity",
				Commands:    []string{"psql -h $DB_HOST -U $DB_USER -c 'SELECT 1'"},
				Expected:    "Should return 1 if database is accessible",
			},
		},
		ResolutionSteps: []RunbookStep{
			{
				Order:       1,
				Description: "Restart the service if logs show OOM",
				Commands:    []string{"kubectl rollout restart deployment/myservice"},
				Notes:       "This will cause brief downtime",
			},
		},
		EscalationPath: []string{
			"Primary on-call engineer (0-15 min)",
			"Secondary on-call (15-30 min)",
			"Team lead (30+ min)",
		},
		Prevention: []string{
			"Implement circuit breakers",
			"Add more comprehensive health checks",
		},
	}

	if runbook.Severity != "critical" {
		t.Errorf("Expected critical severity, got %s", runbook.Severity)
	}
	if len(runbook.Symptoms) != 3 {
		t.Errorf("Expected 3 symptoms, got %d", len(runbook.Symptoms))
	}
	if len(runbook.DiagnosticSteps) != 2 {
		t.Errorf("Expected 2 diagnostic steps, got %d", len(runbook.DiagnosticSteps))
	}
	if len(runbook.EscalationPath) != 3 {
		t.Errorf("Expected 3 escalation levels, got %d", len(runbook.EscalationPath))
	}
}

func TestIncidentAnalysisStructure(t *testing.T) {
	analysis := IncidentAnalysis{
		IncidentID: "INC-2024-001",
		Summary:    "Service outage due to database connection pool exhaustion",
		Severity:   "SEV1",
		Timeline: []TimelineEvent{
			{
				Timestamp:   "2024-01-15T10:00:00Z",
				Description: "First alerts fired for high latency",
				Source:      "alert",
				Severity:    "warning",
			},
			{
				Timestamp:   "2024-01-15T10:05:00Z",
				Description: "Service started returning 503 errors",
				Source:      "metric",
				Severity:    "critical",
			},
		},
		RootCause:    "Database connection pool was exhausted due to slow queries",
		Contributing: []string{"Missing connection timeout", "No circuit breaker"},
		Impact: ImpactAssessment{
			Duration:        "25 minutes",
			AffectedUsers:   "~10,000 users",
			AffectedRegions: []string{"us-east-1", "us-west-2"},
			SLOImpact:       "0.1% error budget consumed",
			FinancialImpact: "$5,000 estimated",
		},
		Resolution: "Increased connection pool size and added connection timeout",
		Recommendations: []Recommendation{
			{
				Priority:    "P0",
				Category:    "architecture",
				Description: "Implement circuit breaker for database calls",
				Owner:       "Platform Team",
				DueDate:     "2024-01-22",
			},
		},
		LessonsLearned: []string{
			"Need better monitoring on connection pool metrics",
			"Should have load tested with realistic traffic",
		},
	}

	if analysis.Severity != "SEV1" {
		t.Errorf("Expected SEV1 severity, got %s", analysis.Severity)
	}
	if len(analysis.Timeline) != 2 {
		t.Errorf("Expected 2 timeline events, got %d", len(analysis.Timeline))
	}
	if len(analysis.Recommendations) != 1 {
		t.Errorf("Expected 1 recommendation, got %d", len(analysis.Recommendations))
	}
	if analysis.Recommendations[0].Priority != "P0" {
		t.Errorf("Expected P0 priority, got %s", analysis.Recommendations[0].Priority)
	}
}

func TestOnCallConfigStructure(t *testing.T) {
	config := OnCallConfig{
		Platform:    "pagerduty",
		ServiceName: "test-service",
		Schedules: []OnCallSchedule{
			{
				Name:         "Primary On-Call",
				TimeZone:     "America/Los_Angeles",
				RotationType: "weekly",
				StartTime:    "2024-01-15T09:00:00",
				Duration:     "7d",
				Participants: []string{"alice", "bob", "charlie"},
			},
		},
		Policies: []EscalationPolicy{
			{
				Name:        "Standard Escalation",
				Description: "Default escalation policy",
				Rules: []EscalationRule{
					{Order: 1, DelayMinutes: 0, Targets: []string{"primary"}, Type: "schedule"},
					{Order: 2, DelayMinutes: 15, Targets: []string{"secondary"}, Type: "schedule"},
					{Order: 3, DelayMinutes: 30, Targets: []string{"team-lead"}, Type: "user"},
				},
			},
		},
	}

	if len(config.Schedules) != 1 {
		t.Errorf("Expected 1 schedule, got %d", len(config.Schedules))
	}
	if len(config.Schedules[0].Participants) != 3 {
		t.Errorf("Expected 3 participants, got %d", len(config.Schedules[0].Participants))
	}
	if len(config.Policies[0].Rules) != 3 {
		t.Errorf("Expected 3 escalation rules, got %d", len(config.Policies[0].Rules))
	}
}

// ============================================================================
// Input Types Tests
// ============================================================================

func TestMonitoringInput(t *testing.T) {
	input := MonitoringInput{
		ServiceName:   "api-gateway",
		Platform:      "prometheus",
		TechStack:     &TechStack{},
		Endpoints:     []string{"/api/v1/users", "/api/v1/orders"},
		CustomMetrics: []string{"business_transactions_total", "payment_success_rate"},
		Labels:        map[string]string{"env": "production"},
	}

	if input.ServiceName != "api-gateway" {
		t.Errorf("Expected api-gateway, got %s", input.ServiceName)
	}
	if len(input.Endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(input.Endpoints))
	}
}

func TestAlertInput(t *testing.T) {
	input := AlertInput{
		ServiceName: "payment-service",
		Platform:    "prometheus",
		SLOs: []SLO{
			{Name: "availability", Target: 99.95, Indicator: "availability"},
			{Name: "latency", Target: 99.0, Indicator: "latency"},
		},
		Thresholds: map[string]float64{
			"error_rate":     0.01,
			"latency_p99_ms": 500,
		},
	}

	if len(input.SLOs) != 2 {
		t.Errorf("Expected 2 SLOs, got %d", len(input.SLOs))
	}
	if input.Thresholds["error_rate"] != 0.01 {
		t.Errorf("Expected error_rate 0.01, got %f", input.Thresholds["error_rate"])
	}
}

func TestDashboardInput(t *testing.T) {
	input := DashboardInput{
		ServiceName: "user-service",
		Platform:    "grafana",
		Metrics:     []string{"http_requests_total", "http_request_duration_seconds"},
		SLOs: []SLO{
			{Name: "availability", Target: 99.9},
		},
		Tags: []string{"backend", "user"},
	}

	if input.Platform != "grafana" {
		t.Errorf("Expected grafana, got %s", input.Platform)
	}
	if len(input.Metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(input.Metrics))
	}
}

func TestRunbookInput(t *testing.T) {
	input := RunbookInput{
		ServiceName: "order-service",
		Alerts: []AlertRule{
			{Name: "HighErrorRate", Severity: "critical"},
			{Name: "HighLatency", Severity: "warning"},
		},
		Architecture: "Microservices with PostgreSQL and Redis",
		Dependencies: []string{"payment-service", "inventory-service", "postgres", "redis"},
		CommonIssues: []string{"Database connection timeout", "Redis memory full"},
	}

	if len(input.Alerts) != 2 {
		t.Errorf("Expected 2 alerts, got %d", len(input.Alerts))
	}
	if len(input.Dependencies) != 4 {
		t.Errorf("Expected 4 dependencies, got %d", len(input.Dependencies))
	}
}

func TestIncidentInput(t *testing.T) {
	input := IncidentInput{
		IncidentID:  "INC-001",
		Description: "Service outage affecting all users",
		Logs:        []string{"ERROR: Connection refused", "WARN: Retry limit exceeded"},
		Metrics:     map[string]string{"error_rate": "0.5", "latency_p99": "5000ms"},
		Alerts:      []string{"HighErrorRate", "ServiceDown"},
		Timeline: []TimelineEvent{
			{Timestamp: "10:00", Description: "First alert", Source: "alert"},
		},
	}

	if input.IncidentID != "INC-001" {
		t.Errorf("Expected INC-001, got %s", input.IncidentID)
	}
	if len(input.Logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(input.Logs))
	}
}

func TestOnCallInput(t *testing.T) {
	input := OnCallInput{
		ServiceName:   "platform",
		Platform:      "pagerduty",
		TeamMembers:   []string{"alice", "bob", "charlie", "dave"},
		TimeZone:      "America/New_York",
		RotationType:  "weekly",
		BusinessHours: false,
	}

	if len(input.TeamMembers) != 4 {
		t.Errorf("Expected 4 team members, got %d", len(input.TeamMembers))
	}
	if input.BusinessHours != false {
		t.Error("Expected BusinessHours to be false")
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestMapPanelType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"graph", "timeseries"},
		{"stat", "stat"},
		{"gauge", "gauge"},
		{"table", "table"},
		{"heatmap", "heatmap"},
		{"logs", "logs"},
		{"unknown", "timeseries"},
	}

	for _, tc := range tests {
		result := mapPanelType(tc.input)
		if result != tc.expected {
			t.Errorf("mapPanelType(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestMapDatadogWidgetType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"graph", "timeseries"},
		{"stat", "query_value"},
		{"gauge", "gauge"},
		{"table", "query_table"},
		{"heatmap", "heatmap"},
		{"logs", "log_stream"},
		{"unknown", "timeseries"},
	}

	for _, tc := range tests {
		result := mapDatadogWidgetType(tc.input)
		if result != tc.expected {
			t.Errorf("mapDatadogWidgetType(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestParseDurationToSeconds(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1h", 3600},
		{"8h", 28800},
		{"1d", 86400},
		{"7d", 604800},
		{"1w", 604800},
		{"invalid", 604800}, // defaults to 1 week
		{"", 604800},        // defaults to 1 week
	}

	for _, tc := range tests {
		result := parseDurationToSeconds(tc.input)
		if result != tc.expected {
			t.Errorf("parseDurationToSeconds(%s) = %d, expected %d", tc.input, result, tc.expected)
		}
	}
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{0, 0, 0},
		{-1, 1, -1},
		{10, 10, 10},
	}

	for _, tc := range tests {
		result := min(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tc.a, tc.b, result, tc.expected)
		}
	}
}

// ============================================================================
// File Generation Tests
// ============================================================================

func TestGeneratePrometheusConfig(t *testing.T) {
	agent := NewSREAgent()

	config := MonitoringConfig{
		Platform:    "prometheus",
		ServiceName: "test-service",
		ScrapeConfigs: []ScrapeConfig{
			{
				JobName:        "app",
				ScrapeInterval: "15s",
				MetricsPath:    "/metrics",
				StaticConfigs: []StaticConfig{
					{
						Targets: []string{"localhost:8080"},
						Labels:  map[string]string{"env": "prod"},
					},
				},
			},
		},
		Labels: map[string]string{"team": "platform"},
	}

	result := agent.generatePrometheusConfig(config)

	if !containsString(result, "scrape_interval: 15s") {
		t.Error("Expected global scrape interval")
	}
	if !containsString(result, "job_name: 'app'") {
		t.Error("Expected job_name in scrape config")
	}
	if !containsString(result, "localhost:8080") {
		t.Error("Expected target in static config")
	}
	if !containsString(result, "team: platform") {
		t.Error("Expected external label")
	}
}

func TestGenerateRecordingRules(t *testing.T) {
	agent := NewSREAgent()

	config := MonitoringConfig{
		ServiceName: "test-service",
		RecordingRules: []RecordingRule{
			{
				Record: "job:http_requests:rate5m",
				Expr:   "sum(rate(http_requests_total[5m])) by (job)",
				Labels: map[string]string{"aggregation": "sum"},
			},
		},
	}

	result := agent.generateRecordingRules(config)

	if !containsString(result, "record: job:http_requests:rate5m") {
		t.Error("Expected recording rule name")
	}
	if !containsString(result, "expr: sum(rate(http_requests_total[5m])) by (job)") {
		t.Error("Expected recording rule expression")
	}
}

func TestGeneratePrometheusAlerts(t *testing.T) {
	agent := NewSREAgent()

	rules := AlertRules{
		ServiceName: "test-service",
		Alerts: []AlertRule{
			{
				Name:        "HighErrorRate",
				Severity:    "critical",
				Condition:   "error_rate > 0.01",
				Duration:    "5m",
				Summary:     "High error rate",
				Description: "Error rate exceeds 1%",
				Runbook:     "https://runbooks.example.com/high-error",
			},
		},
	}

	result := agent.generatePrometheusAlerts(rules)

	if !containsString(result, "alert: HighErrorRate") {
		t.Error("Expected alert name")
	}
	if !containsString(result, "severity: critical") {
		t.Error("Expected severity label")
	}
	if !containsString(result, "for: 5m") {
		t.Error("Expected duration")
	}
	if !containsString(result, "runbook_url:") {
		t.Error("Expected runbook URL annotation")
	}
}

func TestGenerateGrafanaDashboard(t *testing.T) {
	agent := NewSREAgent()

	dashboard := Dashboard{
		Name:        "overview",
		Title:       "Service Overview",
		Description: "Main dashboard",
		Tags:        []string{"sre"},
		Panels: []DashboardPanel{
			{
				Title: "Request Rate",
				Type:  "graph",
				Query: "rate(requests[5m])",
				Unit:  "reqps",
			},
		},
		Variables: []DashboardVar{
			{
				Name:    "env",
				Label:   "Environment",
				Type:    "custom",
				Default: "prod",
			},
		},
		Refresh: "30s",
	}

	result := agent.generateGrafanaDashboard(dashboard, "test-service")

	if !containsString(result, `"title": "Service Overview"`) {
		t.Error("Expected dashboard title")
	}
	if !containsString(result, `"type": "timeseries"`) {
		t.Error("Expected panel type")
	}
	if !containsString(result, `"expr": "rate(requests[5m])"`) {
		t.Error("Expected panel query")
	}
}

func TestGenerateRunbookMarkdown(t *testing.T) {
	agent := NewSREAgent()

	runbook := Runbook{
		Name:        "high-error-rate",
		Title:       "High Error Rate",
		Description: "Handle high error rates",
		AlertName:   "HighErrorRate",
		Severity:    "critical",
		Impact:      "Users affected",
		Symptoms:    []string{"Alert firing", "User complaints"},
		PossibleCauses: []string{"Database issue", "Dependency failure"},
		DiagnosticSteps: []RunbookStep{
			{
				Order:       1,
				Description: "Check logs",
				Commands:    []string{"kubectl logs -l app=service"},
				Expected:    "Look for errors",
			},
		},
		ResolutionSteps: []RunbookStep{
			{
				Order:       1,
				Description: "Restart service",
				Commands:    []string{"kubectl rollout restart deployment/service"},
			},
		},
		EscalationPath: []string{"Primary on-call", "Team lead"},
		Prevention:     []string{"Add circuit breaker"},
	}

	result := agent.generateRunbookMarkdown(runbook)

	if !containsString(result, "# High Error Rate") {
		t.Error("Expected runbook title")
	}
	if !containsString(result, "**Severity:** critical") {
		t.Error("Expected severity")
	}
	if !containsString(result, "## Diagnostic Steps") {
		t.Error("Expected diagnostic steps section")
	}
	if !containsString(result, "kubectl logs -l app=service") {
		t.Error("Expected command in diagnostic step")
	}
	if !containsString(result, "## Escalation Path") {
		t.Error("Expected escalation path section")
	}
}

func TestGenerateRunbookIndex(t *testing.T) {
	agent := NewSREAgent()

	bundle := RunbookBundle{
		ServiceName: "test-service",
		Runbooks: []Runbook{
			{Name: "high-error-rate", Title: "High Error Rate", Severity: "critical", AlertName: "HighErrorRate"},
			{Name: "high-latency", Title: "High Latency", Severity: "warning", AlertName: "HighLatency"},
		},
	}

	result := agent.generateRunbookIndex(bundle)

	if !containsString(result, "# Runbooks for test-service") {
		t.Error("Expected index title")
	}
	if !containsString(result, "| Runbook | Severity | Alert |") {
		t.Error("Expected table header")
	}
	if !containsString(result, "High Error Rate") {
		t.Error("Expected first runbook in index")
	}
	if !containsString(result, "High Latency") {
		t.Error("Expected second runbook in index")
	}
}

// ============================================================================
// Schema Tests
// ============================================================================

func TestGetMonitoringConfigSchema(t *testing.T) {
	schema := GetMonitoringConfigSchema()

	if !containsString(schema, `"platform"`) {
		t.Error("Schema should contain platform")
	}
	if !containsString(schema, `"metrics"`) {
		t.Error("Schema should contain metrics")
	}
	if !containsString(schema, `"scrape_configs"`) {
		t.Error("Schema should contain scrape_configs")
	}
}

func TestGetAlertRulesSchema(t *testing.T) {
	schema := GetAlertRulesSchema()

	if !containsString(schema, `"alerts"`) {
		t.Error("Schema should contain alerts")
	}
	if !containsString(schema, `"severity"`) {
		t.Error("Schema should contain severity")
	}
	if !containsString(schema, `"condition"`) {
		t.Error("Schema should contain condition")
	}
}

func TestGetDashboardBundleSchema(t *testing.T) {
	schema := GetDashboardBundleSchema()

	if !containsString(schema, `"dashboards"`) {
		t.Error("Schema should contain dashboards")
	}
	if !containsString(schema, `"panels"`) {
		t.Error("Schema should contain panels")
	}
	if !containsString(schema, `"variables"`) {
		t.Error("Schema should contain variables")
	}
}

func TestGetRunbookBundleSchema(t *testing.T) {
	schema := GetRunbookBundleSchema()

	if !containsString(schema, `"runbooks"`) {
		t.Error("Schema should contain runbooks")
	}
	if !containsString(schema, `"diagnostic_steps"`) {
		t.Error("Schema should contain diagnostic_steps")
	}
	if !containsString(schema, `"resolution_steps"`) {
		t.Error("Schema should contain resolution_steps")
	}
}

func TestGetIncidentAnalysisSchema(t *testing.T) {
	schema := GetIncidentAnalysisSchema()

	if !containsString(schema, `"root_cause"`) {
		t.Error("Schema should contain root_cause")
	}
	if !containsString(schema, `"timeline"`) {
		t.Error("Schema should contain timeline")
	}
	if !containsString(schema, `"recommendations"`) {
		t.Error("Schema should contain recommendations")
	}
}

func TestGetOnCallConfigSchema(t *testing.T) {
	schema := GetOnCallConfigSchema()

	if !containsString(schema, `"schedules"`) {
		t.Error("Schema should contain schedules")
	}
	if !containsString(schema, `"policies"`) {
		t.Error("Schema should contain policies")
	}
	if !containsString(schema, `"rules"`) {
		t.Error("Schema should contain rules")
	}
}

// ============================================================================
// PagerDuty/Opsgenie Generation Tests
// ============================================================================

func TestGeneratePagerDutySchedules(t *testing.T) {
	agent := NewSREAgent()

	config := OnCallConfig{
		ServiceName: "test-service",
		Schedules: []OnCallSchedule{
			{
				Name:         "Primary",
				TimeZone:     "America/Los_Angeles",
				RotationType: "weekly",
				StartTime:    "2024-01-15T09:00:00",
				Duration:     "7d",
				Participants: []string{"alice", "bob"},
			},
		},
	}

	result := agent.generatePagerDutySchedules(config)

	if !containsString(result, `"name": "Primary"`) {
		t.Error("Expected schedule name")
	}
	if !containsString(result, `"time_zone": "America/Los_Angeles"`) {
		t.Error("Expected timezone")
	}
}

func TestGeneratePagerDutyEscalation(t *testing.T) {
	agent := NewSREAgent()

	config := OnCallConfig{
		Policies: []EscalationPolicy{
			{
				Name:        "Default",
				Description: "Default escalation",
				Rules: []EscalationRule{
					{Order: 1, DelayMinutes: 0, Targets: []string{"primary"}, Type: "schedule"},
					{Order: 2, DelayMinutes: 15, Targets: []string{"secondary"}, Type: "schedule"},
				},
			},
		},
	}

	result := agent.generatePagerDutyEscalation(config)

	if !containsString(result, `"name": "Default"`) {
		t.Error("Expected policy name")
	}
	if !containsString(result, `"escalation_delay_in_minutes"`) {
		t.Error("Expected escalation delay")
	}
}

func TestGenerateOpsgenieSchedules(t *testing.T) {
	agent := NewSREAgent()

	config := OnCallConfig{
		Schedules: []OnCallSchedule{
			{
				Name:         "Primary",
				TimeZone:     "UTC",
				RotationType: "weekly",
				StartTime:    "2024-01-15T00:00:00",
				Participants: []string{"alice", "bob"},
			},
		},
	}

	result := agent.generateOpsgenieSchedules(config)

	if !containsString(result, `"name": "Primary"`) {
		t.Error("Expected schedule name")
	}
	if !containsString(result, `"enabled": true`) {
		t.Error("Expected enabled flag")
	}
}

// ============================================================================
// Datadog Generation Tests
// ============================================================================

func TestGenerateDatadogConfig(t *testing.T) {
	agent := NewSREAgent()

	config := MonitoringConfig{
		ServiceName: "test-service",
		Labels: map[string]string{
			"env":  "production",
			"team": "platform",
		},
	}

	result := agent.generateDatadogConfig(config)

	if !containsString(result, "api_key: ${DD_API_KEY}") {
		t.Error("Expected API key placeholder")
	}
	if !containsString(result, "logs_enabled: true") {
		t.Error("Expected logs enabled")
	}
	if !containsString(result, "service:test-service") {
		t.Error("Expected service tag")
	}
}

func TestGenerateDatadogMonitors(t *testing.T) {
	agent := NewSREAgent()

	rules := AlertRules{
		ServiceName: "test-service",
		Alerts: []AlertRule{
			{
				Name:        "HighErrorRate",
				Severity:    "critical",
				Condition:   "avg:error_rate{service:test} > 0.01",
				Summary:     "Error rate too high",
				Description: "Error rate exceeds threshold",
			},
		},
	}

	result := agent.generateDatadogMonitors(rules)

	if !containsString(result, `"name": "HighErrorRate"`) {
		t.Error("Expected monitor name")
	}
	if !containsString(result, `"type": "metric alert"`) {
		t.Error("Expected metric alert type")
	}
	if !containsString(result, "severity:critical") {
		t.Error("Expected severity tag")
	}
}

func TestGenerateDatadogDashboard(t *testing.T) {
	agent := NewSREAgent()

	dashboard := Dashboard{
		Name:        "overview",
		Title:       "Service Overview",
		Description: "Main dashboard",
		Panels: []DashboardPanel{
			{
				Title: "Request Rate",
				Type:  "graph",
				Query: "sum:http.requests{service:test}.as_rate()",
			},
		},
	}

	result := agent.generateDatadogDashboard(dashboard)

	if !containsString(result, `"title": "Service Overview"`) {
		t.Error("Expected dashboard title")
	}
	if !containsString(result, `"layout_type": "ordered"`) {
		t.Error("Expected layout type")
	}
}

// ============================================================================
// Integration Tests (Mocked)
// ============================================================================

func TestMonitoringFilesGeneration(t *testing.T) {
	agent := NewSREAgent()

	config := MonitoringConfig{
		Platform:    "prometheus",
		ServiceName: "test-service",
		ScrapeConfigs: []ScrapeConfig{
			{JobName: "app", ScrapeInterval: "15s"},
		},
		RecordingRules: []RecordingRule{
			{Record: "test:metric", Expr: "sum(rate(test[5m]))"},
		},
	}
	req := MonitoringInput{
		ServiceName: "test-service",
		Platform:    "prometheus",
	}

	files := agent.generateMonitoringFiles(config, req)

	if _, ok := files["prometheus/prometheus.yml"]; !ok {
		t.Error("Expected prometheus.yml file")
	}
	if _, ok := files["prometheus/recording_rules.yml"]; !ok {
		t.Error("Expected recording_rules.yml file")
	}
}

func TestAlertFilesGeneration(t *testing.T) {
	agent := NewSREAgent()

	rules := AlertRules{
		ServiceName: "test-service",
		Alerts: []AlertRule{
			{Name: "TestAlert", Severity: "warning", Condition: "test > 1", Duration: "5m", Summary: "Test"},
		},
	}
	req := AlertInput{
		ServiceName: "test-service",
		Platform:    "prometheus",
	}

	files := agent.generateAlertFiles(rules, req)

	if _, ok := files["prometheus/alerts.yml"]; !ok {
		t.Error("Expected alerts.yml file for prometheus")
	}
}

func TestDashboardFilesGeneration(t *testing.T) {
	agent := NewSREAgent()

	bundle := DashboardBundle{
		Platform:    "grafana",
		ServiceName: "test-service",
		Dashboards: []Dashboard{
			{
				Name:   "overview",
				Title:  "Overview",
				Panels: []DashboardPanel{{Title: "Test", Type: "graph", Query: "test"}},
			},
		},
	}
	req := DashboardInput{
		ServiceName: "test-service",
		Platform:    "grafana",
	}

	files := agent.generateDashboardFiles(bundle, req)

	if _, ok := files["grafana/dashboards/overview.json"]; !ok {
		t.Error("Expected Grafana dashboard JSON file")
	}
}

func TestRunbookFilesGeneration(t *testing.T) {
	agent := NewSREAgent()

	bundle := RunbookBundle{
		ServiceName: "test-service",
		Runbooks: []Runbook{
			{
				Name:            "high-error-rate",
				Title:           "High Error Rate",
				Description:     "Test",
				Severity:        "critical",
				Impact:          "High",
				Symptoms:        []string{"Test"},
				DiagnosticSteps: []RunbookStep{{Order: 1, Description: "Test"}},
				ResolutionSteps: []RunbookStep{{Order: 1, Description: "Test"}},
				EscalationPath:  []string{"Test"},
			},
		},
	}

	files := agent.generateRunbookFiles(bundle)

	if _, ok := files["runbooks/high-error-rate.md"]; !ok {
		t.Error("Expected runbook markdown file")
	}
	if _, ok := files["runbooks/README.md"]; !ok {
		t.Error("Expected runbook index file")
	}
}

func TestOnCallFilesGeneration(t *testing.T) {
	agent := NewSREAgent()

	config := OnCallConfig{
		Platform:    "pagerduty",
		ServiceName: "test-service",
		Schedules: []OnCallSchedule{
			{Name: "Primary", TimeZone: "UTC", RotationType: "weekly", Participants: []string{"alice"}},
		},
		Policies: []EscalationPolicy{
			{Name: "Default", Rules: []EscalationRule{{Order: 1, DelayMinutes: 0, Targets: []string{"primary"}, Type: "schedule"}}},
		},
	}
	req := OnCallInput{
		ServiceName: "test-service",
		Platform:    "pagerduty",
	}

	files := agent.generateOnCallFiles(config, req)

	if _, ok := files["pagerduty/schedules.json"]; !ok {
		t.Error("Expected PagerDuty schedules file")
	}
	if _, ok := files["pagerduty/escalation_policies.json"]; !ok {
		t.Error("Expected PagerDuty escalation policies file")
	}
}

// Note: containsString and containsSubstring helpers are defined in brownfield_test.go
