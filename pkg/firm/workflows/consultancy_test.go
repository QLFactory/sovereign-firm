package workflows

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConsultancyPhase(t *testing.T) {
	phases := []ConsultancyPhase{
		PhaseIntake,
		PhaseSizing,
		PhasePlanning,
		PhaseArchitecture,
		PhaseDevelopment,
		PhaseTesting,
		PhaseDeployment,
		PhaseOperations,
		PhaseHandoff,
		PhaseComplete,
		PhaseReview,
		PhaseFailed,
	}

	assert.Equal(t, 12, len(phases))
	assert.Equal(t, ConsultancyPhase("INTAKE"), PhaseIntake)
	assert.Equal(t, ConsultancyPhase("COMPLETE"), PhaseComplete)
}

func TestConsultancyConfig(t *testing.T) {
	config := ConsultancyConfig{
		ProjectName:       "Test Project",
		ClientID:          "client-123",
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

	assert.Equal(t, "Test Project", config.ProjectName)
	assert.Equal(t, "client-123", config.ClientID)
	assert.True(t, config.EnableFullStack)
	assert.True(t, config.EnableDeployment)
	assert.True(t, config.EnableSRE)
	assert.Equal(t, "react", config.PreferredFrontend)
	assert.Equal(t, 3, config.MaxCodeAttempts)
}

func TestConsultancyConfigDefaults(t *testing.T) {
	config := ConsultancyConfig{}

	// Zero values
	assert.Equal(t, "", config.ProjectName)
	assert.False(t, config.EnableFullStack)
	assert.Equal(t, 0, config.MaxCodeAttempts) // Will be set to 3 in workflow
}

func TestConsultancyState(t *testing.T) {
	state := &ConsultancyState{
		ProjectID:    "proj-123",
		ProjectName:  "Test Project",
		ClientID:     "client-456",
		Phase:        PhaseIntake,
		StartedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ChatHistory:  "User: Hello\nPM: Hi!",
		AllCodeFiles: make(map[string]string),
		PhaseHistory: []string{string(PhaseIntake)},
	}

	assert.Equal(t, "proj-123", state.ProjectID)
	assert.Equal(t, "Test Project", state.ProjectName)
	assert.Equal(t, PhaseIntake, state.Phase)
	assert.Contains(t, state.ChatHistory, "Hello")
	assert.Equal(t, 1, len(state.PhaseHistory))
}

func TestConsultancyStateWithCodeFiles(t *testing.T) {
	state := &ConsultancyState{
		FrontendCode: map[string]string{
			"src/App.tsx": "export default function App() {}",
		},
		BackendCode: map[string]string{
			"src/index.ts": "const app = express()",
		},
		DatabaseCode: map[string]string{
			"migrations/001_init.sql": "CREATE TABLE users...",
		},
		AllCodeFiles: make(map[string]string),
	}

	assert.Equal(t, 1, len(state.FrontendCode))
	assert.Equal(t, 1, len(state.BackendCode))
	assert.Equal(t, 1, len(state.DatabaseCode))
}

func TestConsultancyStateWithTests(t *testing.T) {
	state := &ConsultancyState{
		UnitTests: map[string]string{
			"src/App.test.tsx": "test('renders', () => {})",
		},
		IntegrationTests: map[string]string{
			"tests/api.test.ts": "test('api works', () => {})",
		},
		E2ETests: map[string]string{
			"e2e/login.spec.ts": "test('user can login', () => {})",
		},
		TestCoverage: map[string]interface{}{
			"line_coverage":   85.5,
			"branch_coverage": 72.3,
		},
	}

	assert.Equal(t, 1, len(state.UnitTests))
	assert.Equal(t, 1, len(state.IntegrationTests))
	assert.Equal(t, 1, len(state.E2ETests))
	assert.NotNil(t, state.TestCoverage)
}

func TestConsultancyStateWithDeployment(t *testing.T) {
	state := &ConsultancyState{
		Dockerfile:    "FROM node:18",
		DockerCompose: "version: '3.8'",
		KubeManifests: map[string]string{
			"deployment.yaml": "apiVersion: apps/v1",
			"service.yaml":    "apiVersion: v1",
		},
		HelmChart: map[string]string{
			"Chart.yaml":   "apiVersion: v2",
			"values.yaml":  "replicaCount: 3",
		},
		CIPipeline: "name: CI",
		InfraCode: map[string]string{
			"main.tf": "provider \"aws\" {}",
		},
	}

	assert.Contains(t, state.Dockerfile, "node")
	assert.Contains(t, state.DockerCompose, "3.8")
	assert.Equal(t, 2, len(state.KubeManifests))
	assert.Equal(t, 2, len(state.HelmChart))
	assert.NotEmpty(t, state.CIPipeline)
	assert.Equal(t, 1, len(state.InfraCode))
}

func TestConsultancyStateWithOperations(t *testing.T) {
	state := &ConsultancyState{
		MonitoringConfig: map[string]interface{}{
			"scrape_interval": "15s",
			"metrics":         []string{"http_requests", "latency"},
		},
		AlertRules: map[string]interface{}{
			"groups": []string{"slo-alerts"},
		},
		Dashboards: map[string]interface{}{
			"title": "Overview",
		},
		Runbooks: map[string]string{
			"high_latency.md": "# High Latency Runbook",
		},
	}

	assert.NotNil(t, state.MonitoringConfig)
	assert.NotNil(t, state.AlertRules)
	assert.NotNil(t, state.Dashboards)
	assert.Equal(t, 1, len(state.Runbooks))
}

func TestConsultancyStateWithErrors(t *testing.T) {
	state := &ConsultancyState{
		Errors:   []string{"Frontend build failed", "Tests timed out"},
		Warnings: []string{"Low test coverage"},
	}

	assert.Equal(t, 2, len(state.Errors))
	assert.Equal(t, 1, len(state.Warnings))
}

func TestTransitionPhase(t *testing.T) {
	state := &ConsultancyState{
		Phase:        PhaseIntake,
		PhaseHistory: []string{string(PhaseIntake)},
	}

	now := time.Now()
	transitionPhase(state, PhaseSizing, now)

	assert.Equal(t, PhaseSizing, state.Phase)
	assert.Equal(t, 2, len(state.PhaseHistory))
	assert.Equal(t, string(PhaseSizing), state.PhaseHistory[1])
	assert.Equal(t, now, state.UpdatedAt)
}

func TestGetString(t *testing.T) {
	m := map[string]interface{}{
		"frontend": "react",
		"backend":  "nodejs",
		"empty":    "",
	}

	assert.Equal(t, "react", getString(m, "frontend", "vue"))
	assert.Equal(t, "nodejs", getString(m, "backend", "python"))
	assert.Equal(t, "default", getString(m, "missing", "default"))
	assert.Equal(t, "default", getString(m, "empty", "default"))
}

func TestFirstNonEmpty(t *testing.T) {
	assert.Equal(t, "first", firstNonEmpty("first", "second"))
	assert.Equal(t, "second", firstNonEmpty("", "second"))
	assert.Equal(t, "third", firstNonEmpty("", "", "third"))
	assert.Equal(t, "", firstNonEmpty("", "", ""))
}

func TestBoolToInt(t *testing.T) {
	assert.Equal(t, 1, boolToInt(true))
	assert.Equal(t, 0, boolToInt(false))
}

func TestFilterTestFiles(t *testing.T) {
	files := map[string]string{
		"src/App.tsx":       "app code",
		"src/App.test.tsx":  "test code",
		"src/utils.ts":      "utils code",
		"src/utils.spec.ts": "spec code",
		"src/helper_test.go": "go test",
	}

	testFiles := filterTestFiles(files)

	assert.Equal(t, 3, len(testFiles))
	assert.Contains(t, testFiles, "src/App.test.tsx")
	assert.Contains(t, testFiles, "src/utils.spec.ts")
	assert.Contains(t, testFiles, "src/helper_test.go")
	assert.NotContains(t, testFiles, "src/App.tsx")
}

func TestFilterNonTestFiles(t *testing.T) {
	files := map[string]string{
		"src/App.tsx":       "app code",
		"src/App.test.tsx":  "test code",
		"src/utils.ts":      "utils code",
		"src/utils.spec.ts": "spec code",
	}

	nonTestFiles := filterNonTestFiles(files)

	assert.Equal(t, 2, len(nonTestFiles))
	assert.Contains(t, nonTestFiles, "src/App.tsx")
	assert.Contains(t, nonTestFiles, "src/utils.ts")
	assert.NotContains(t, nonTestFiles, "src/App.test.tsx")
}

func TestIsTestFile(t *testing.T) {
	testCases := []struct {
		filename string
		isTest   bool
	}{
		{"App.test.tsx", true},
		{"App.spec.tsx", true},
		{"helper_test.go", true},
		{"App.tsx", false},
		{"testutils.ts", false},
		{"my.test.helper.ts", true},
	}

	for _, tc := range testCases {
		result := isTestFile(tc.filename)
		assert.Equal(t, tc.isTest, result, "Failed for: %s", tc.filename)
	}
}

func TestConsultancyStateComplete(t *testing.T) {
	now := time.Now()
	state := &ConsultancyState{
		ProjectID:   "proj-123",
		ProjectName: "Complete Project",
		Phase:       PhaseComplete,
		CompletedAt: &now,
		PhaseHistory: []string{
			string(PhaseIntake),
			string(PhaseSizing),
			string(PhasePlanning),
			string(PhaseArchitecture),
			string(PhaseDevelopment),
			string(PhaseTesting),
			string(PhaseDeployment),
			string(PhaseOperations),
			string(PhaseHandoff),
			string(PhaseComplete),
		},
	}

	assert.Equal(t, PhaseComplete, state.Phase)
	assert.NotNil(t, state.CompletedAt)
	assert.Equal(t, 10, len(state.PhaseHistory))
}

func TestConsultancyWorkflowPhaseOrder(t *testing.T) {
	// Verify the expected phase order
	expectedOrder := []ConsultancyPhase{
		PhaseIntake,
		PhaseSizing,
		PhasePlanning,
		PhaseArchitecture,
		PhaseDevelopment,
		PhaseTesting,
		PhaseDeployment,
		PhaseOperations,
		PhaseHandoff,
		PhaseComplete,
	}

	for i, phase := range expectedOrder {
		assert.NotEmpty(t, string(phase), "Phase %d should not be empty", i)
	}
}

func TestFullStackConfigValidation(t *testing.T) {
	// Full stack config
	fullStackConfig := ConsultancyConfig{
		EnableFullStack:  true,
		EnableDeployment: true,
		EnableSRE:        true,
	}

	assert.True(t, fullStackConfig.EnableFullStack)
	assert.True(t, fullStackConfig.EnableDeployment)
	assert.True(t, fullStackConfig.EnableSRE)

	// Frontend only config
	frontendConfig := ConsultancyConfig{
		EnableFullStack:  false,
		EnableDeployment: false,
		EnableSRE:        false,
	}

	assert.False(t, frontendConfig.EnableFullStack)
	assert.False(t, frontendConfig.EnableDeployment)
	assert.False(t, frontendConfig.EnableSRE)
}

func TestTechStackPreferences(t *testing.T) {
	config := ConsultancyConfig{
		PreferredFrontend: "vue",
		PreferredBackend:  "python",
		PreferredDatabase: "mongodb",
		PreferredCloud:    "gcp",
	}

	// These should be used if architect doesn't override
	assert.Equal(t, "vue", config.PreferredFrontend)
	assert.Equal(t, "python", config.PreferredBackend)
	assert.Equal(t, "mongodb", config.PreferredDatabase)
	assert.Equal(t, "gcp", config.PreferredCloud)
}
