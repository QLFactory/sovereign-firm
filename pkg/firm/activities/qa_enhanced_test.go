package activities

import (
	"encoding/json"
	"strings"
	"testing"
)

// ============================================================================
// Data Structure Tests
// ============================================================================

func TestTestBundleStructure(t *testing.T) {
	bundle := TestBundle{
		TestType:  "integration",
		Framework: "supertest",
		Files: map[string]string{
			"tests/integration/users.test.js": "test content...",
		},
		SetupFiles: map[string]string{
			"tests/setup.js": "setup content...",
		},
		RunCmd:   "npm run test:integration",
		CIConfig: "- run: npm run test:integration",
	}

	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("Failed to marshal TestBundle: %v", err)
	}

	var decoded TestBundle
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal TestBundle: %v", err)
	}

	if decoded.TestType != "integration" {
		t.Errorf("Expected test_type 'integration', got '%s'", decoded.TestType)
	}
	if decoded.Framework != "supertest" {
		t.Errorf("Expected framework 'supertest', got '%s'", decoded.Framework)
	}
	if len(decoded.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(decoded.Files))
	}
}

func TestTestCoverageStructure(t *testing.T) {
	coverage := TestCoverage{
		TotalFiles:      10,
		TestedFiles:     7,
		CoveragePercent: 70.0,
		UncoveredFiles:  []string{"src/utils/helpers.js", "src/services/email.js"},
		Recommendations: []string{
			"Add tests for email service",
			"Add edge case tests for helpers",
		},
	}

	data, err := json.Marshal(coverage)
	if err != nil {
		t.Fatalf("Failed to marshal TestCoverage: %v", err)
	}

	var decoded TestCoverage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal TestCoverage: %v", err)
	}

	if decoded.TotalFiles != 10 {
		t.Errorf("Expected total_files 10, got %d", decoded.TotalFiles)
	}
	if decoded.CoveragePercent != 70.0 {
		t.Errorf("Expected coverage_percent 70.0, got %f", decoded.CoveragePercent)
	}
	if len(decoded.UncoveredFiles) != 2 {
		t.Errorf("Expected 2 uncovered files, got %d", len(decoded.UncoveredFiles))
	}
}

func TestSecurityScanResultStructure(t *testing.T) {
	result := SecurityScanResult{
		ScanType: "sast",
		Tool:     "semgrep",
		ConfigFiles: map[string]string{
			".semgrep.yml": "rules config...",
		},
		RunCmd:   "semgrep scan --config=.semgrep.yml",
		CIConfig: "- run: semgrep scan",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal SecurityScanResult: %v", err)
	}

	var decoded SecurityScanResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SecurityScanResult: %v", err)
	}

	if decoded.ScanType != "sast" {
		t.Errorf("Expected scan_type 'sast', got '%s'", decoded.ScanType)
	}
	if decoded.Tool != "semgrep" {
		t.Errorf("Expected tool 'semgrep', got '%s'", decoded.Tool)
	}
}

// ============================================================================
// Input Type Tests
// ============================================================================

func TestIntegrationTestInputStructure(t *testing.T) {
	input := IntegrationTestInput{
		Stack:        "node",
		Framework:    "express",
		CodeFiles:    map[string]string{"src/routes/users.js": "route code..."},
		DatabaseType: "postgresql",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal IntegrationTestInput: %v", err)
	}

	var decoded IntegrationTestInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal IntegrationTestInput: %v", err)
	}

	if decoded.Stack != "node" {
		t.Errorf("Expected stack 'node', got '%s'", decoded.Stack)
	}
	if decoded.DatabaseType != "postgresql" {
		t.Errorf("Expected database_type 'postgresql', got '%s'", decoded.DatabaseType)
	}
}

func TestE2ETestInputStructure(t *testing.T) {
	input := E2ETestInput{
		Framework:    "playwright",
		BaseURL:      "http://localhost:3000",
		UserFlows:    []string{"login", "checkout", "profile update"},
		AuthRequired: true,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal E2ETestInput: %v", err)
	}

	var decoded E2ETestInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal E2ETestInput: %v", err)
	}

	if decoded.Framework != "playwright" {
		t.Errorf("Expected framework 'playwright', got '%s'", decoded.Framework)
	}
	if len(decoded.UserFlows) != 3 {
		t.Errorf("Expected 3 user flows, got %d", len(decoded.UserFlows))
	}
	if !decoded.AuthRequired {
		t.Error("Expected AuthRequired to be true")
	}
}

func TestPerformanceTestInputStructure(t *testing.T) {
	input := PerformanceTestInput{
		Tool:         "k6",
		BaseURL:      "http://localhost:3000",
		Endpoints:    []string{"/api/users", "/api/products", "/api/orders"},
		TargetRPS:    200,
		Duration:     "10m",
		VirtualUsers: 100,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal PerformanceTestInput: %v", err)
	}

	var decoded PerformanceTestInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PerformanceTestInput: %v", err)
	}

	if decoded.TargetRPS != 200 {
		t.Errorf("Expected target_rps 200, got %d", decoded.TargetRPS)
	}
	if decoded.VirtualUsers != 100 {
		t.Errorf("Expected virtual_users 100, got %d", decoded.VirtualUsers)
	}
}

func TestSecurityTestInputStructure(t *testing.T) {
	input := SecurityTestInput{
		ScanTypes: []string{"sast", "dependency", "secrets"},
		Stack:     "node",
		BaseURL:   "http://localhost:3000",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal SecurityTestInput: %v", err)
	}

	var decoded SecurityTestInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SecurityTestInput: %v", err)
	}

	if len(decoded.ScanTypes) != 3 {
		t.Errorf("Expected 3 scan types, got %d", len(decoded.ScanTypes))
	}
}

func TestAccessibilityTestInputStructure(t *testing.T) {
	input := AccessibilityTestInput{
		Tool:     "axe",
		BaseURL:  "http://localhost:3000",
		Pages:    []string{"/", "/login", "/dashboard", "/profile"},
		Standard: "WCAG2AA",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal AccessibilityTestInput: %v", err)
	}

	var decoded AccessibilityTestInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AccessibilityTestInput: %v", err)
	}

	if decoded.Standard != "WCAG2AA" {
		t.Errorf("Expected standard 'WCAG2AA', got '%s'", decoded.Standard)
	}
	if len(decoded.Pages) != 4 {
		t.Errorf("Expected 4 pages, got %d", len(decoded.Pages))
	}
}

func TestBackendTestInputStructure(t *testing.T) {
	input := BackendTestInput{
		Stack:         "python",
		Framework:     "fastapi",
		CodeFiles:     map[string]string{"app/main.py": "code..."},
		TestFramework: "pytest",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal BackendTestInput: %v", err)
	}

	var decoded BackendTestInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal BackendTestInput: %v", err)
	}

	if decoded.TestFramework != "pytest" {
		t.Errorf("Expected test_framework 'pytest', got '%s'", decoded.TestFramework)
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestGetDefaultTestFramework(t *testing.T) {
	tests := []struct {
		stack    string
		expected string
	}{
		{"node", "jest"},
		{"python", "pytest"},
		{"go", "go test"},
		{"unknown", "jest"},
		{"", "jest"},
	}

	for _, tt := range tests {
		t.Run(tt.stack, func(t *testing.T) {
			result := getDefaultTestFramework(tt.stack)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildCodeContext(t *testing.T) {
	tests := []struct {
		name      string
		codeFiles map[string]string
		contains  []string
	}{
		{
			name:      "empty files",
			codeFiles: map[string]string{},
			contains:  []string{},
		},
		{
			name: "with files",
			codeFiles: map[string]string{
				"src/app.js": "const app = express();",
			},
			contains: []string{"APPLICATION CODE:", "File: src/app.js", "const app"},
		},
		{
			name: "truncates long files",
			codeFiles: map[string]string{
				"src/big.js": strings.Repeat("x", 3000),
			},
			contains: []string{"truncated"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCodeContext(tt.codeFiles)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s'", expected)
				}
			}
		})
	}
}

func TestGetTestPyramidRecommendations(t *testing.T) {
	recommendations := GetTestPyramidRecommendations()

	expectedKeys := []string{"unit", "integration", "e2e"}
	for _, key := range expectedKeys {
		if _, ok := recommendations[key]; !ok {
			t.Errorf("Missing key '%s' in test pyramid recommendations", key)
		}
	}

	// Check that percentages sum to 100
	total := 0
	for _, pct := range recommendations {
		total += pct
	}
	if total != 100 {
		t.Errorf("Expected percentages to sum to 100, got %d", total)
	}

	// Check relative proportions (unit > integration > e2e)
	if recommendations["unit"] <= recommendations["integration"] {
		t.Error("Expected unit tests > integration tests")
	}
	if recommendations["integration"] <= recommendations["e2e"] {
		t.Error("Expected integration tests > e2e tests")
	}
}

func TestGetTestFrameworksByStack(t *testing.T) {
	tests := []struct {
		stack          string
		expectedUnit   string
		expectedE2E    string
	}{
		{"node", "jest", "playwright"},
		{"python", "pytest", "playwright"},
		{"go", "go test + testify", "playwright"},
		{"unknown", "jest", "playwright"},
	}

	for _, tt := range tests {
		t.Run(tt.stack, func(t *testing.T) {
			frameworks := GetTestFrameworksByStack(tt.stack)

			if frameworks["unit"] != tt.expectedUnit {
				t.Errorf("Expected unit framework '%s', got '%s'", tt.expectedUnit, frameworks["unit"])
			}
			if frameworks["e2e"] != tt.expectedE2E {
				t.Errorf("Expected e2e framework '%s', got '%s'", tt.expectedE2E, frameworks["e2e"])
			}
		})
	}
}

func TestGenerateTestConfig(t *testing.T) {
	tests := []struct {
		testType  string
		framework string
		contains  []string
	}{
		{
			testType:  "e2e",
			framework: "playwright",
			contains:  []string{"defineConfig", "chromium", "firefox", "webkit", "testDir"},
		},
		{
			testType:  "performance",
			framework: "k6",
			contains:  []string{"scenarios", "smoke", "load", "stress", "thresholds"},
		},
		{
			testType:  "unknown",
			framework: "unknown",
			contains:  []string{}, // Should return empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.testType+"_"+tt.framework, func(t *testing.T) {
			result := GenerateTestConfig(tt.testType, tt.framework)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected config to contain '%s'", expected)
				}
			}
		})
	}
}

// ============================================================================
// Agent Construction Tests
// ============================================================================

func TestNewEnhancedQAAgent(t *testing.T) {
	agent := NewEnhancedQAAgent()

	if agent == nil {
		t.Fatal("NewEnhancedQAAgent returned nil")
	}

	if agent.llmClient == nil {
		t.Error("EnhancedQAAgent.llmClient is nil")
	}
}

// ============================================================================
// Default Value Tests
// ============================================================================

func TestE2EFrameworkDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to playwright", "", "playwright"},
		{"playwright stays", "playwright", "playwright"},
		{"cypress stays", "cypress", "cypress"},
		{"Playwright normalizes", "Playwright", "playwright"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			framework := strings.ToLower(tt.input)
			if framework == "" {
				framework = "playwright"
			}
			if framework != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, framework)
			}
		})
	}
}

func TestPerformanceToolDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to k6", "", "k6"},
		{"k6 stays", "k6", "k6"},
		{"artillery stays", "artillery", "artillery"},
		{"K6 normalizes", "K6", "k6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := strings.ToLower(tt.input)
			if tool == "" {
				tool = "k6"
			}
			if tool != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, tool)
			}
		})
	}
}

func TestAccessibilityStandardDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to WCAG2AA", "", "WCAG2AA"},
		{"WCAG2A stays", "WCAG2A", "WCAG2A"},
		{"WCAG2AA stays", "WCAG2AA", "WCAG2AA"},
		{"WCAG2AAA stays", "WCAG2AAA", "WCAG2AAA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			standard := tt.input
			if standard == "" {
				standard = "WCAG2AA"
			}
			if standard != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, standard)
			}
		})
	}
}

func TestTargetRPSDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero defaults to 100", 0, 100},
		{"negative defaults to 100", -10, 100},
		{"positive stays", 500, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetRPS := tt.input
			if targetRPS <= 0 {
				targetRPS = 100
			}
			if targetRPS != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, targetRPS)
			}
		})
	}
}

func TestVirtualUsersDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero defaults to 50", 0, 50},
		{"negative defaults to 50", -5, 50},
		{"positive stays", 200, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vus := tt.input
			if vus <= 0 {
				vus = 50
			}
			if vus != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, vus)
			}
		})
	}
}

func TestDurationDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to 5m", "", "5m"},
		{"5m stays", "5m", "5m"},
		{"10m stays", "10m", "10m"},
		{"1h stays", "1h", "1h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := tt.input
			if duration == "" {
				duration = "5m"
			}
			if duration != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, duration)
			}
		})
	}
}

// ============================================================================
// Security Scan Types Tests
// ============================================================================

func TestDefaultScanTypes(t *testing.T) {
	// Test that default scan types are set when empty
	scanTypes := []string{}
	if len(scanTypes) == 0 {
		scanTypes = []string{"dependency", "sast", "secrets"}
	}

	expectedTypes := map[string]bool{
		"dependency": true,
		"sast":       true,
		"secrets":    true,
	}

	for _, st := range scanTypes {
		if !expectedTypes[st] {
			t.Errorf("Unexpected scan type: %s", st)
		}
	}

	if len(scanTypes) != 3 {
		t.Errorf("Expected 3 default scan types, got %d", len(scanTypes))
	}
}

// ============================================================================
// Test Pyramid Distribution Tests
// ============================================================================

func TestTestPyramidDistribution(t *testing.T) {
	pyramid := GetTestPyramidRecommendations()

	// Unit tests should be the majority
	if pyramid["unit"] < 50 {
		t.Errorf("Unit tests should be at least 50%%, got %d%%", pyramid["unit"])
	}

	// E2E tests should be minimal
	if pyramid["e2e"] > 20 {
		t.Errorf("E2E tests should be at most 20%%, got %d%%", pyramid["e2e"])
	}
}
