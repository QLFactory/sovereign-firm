package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// EnhancedQAAgent provides comprehensive test generation across the full test pyramid
// Supports unit tests, integration tests, E2E tests, performance tests, security tests, and accessibility tests
type EnhancedQAAgent struct {
	llmClient llm.Client
}

func NewEnhancedQAAgent() *EnhancedQAAgent {
	return &EnhancedQAAgent{
		llmClient: llm.NewClient(),
	}
}

// ============================================================================
// Data Structures
// ============================================================================

// TestBundle represents generated test files
type TestBundle struct {
	TestType    string            `json:"test_type"`    // "unit", "integration", "e2e", "performance", "security", "accessibility"
	Framework   string            `json:"framework"`    // "vitest", "jest", "pytest", "playwright", "k6", etc.
	Files       map[string]string `json:"files"`        // filename -> content
	SetupFiles  map[string]string `json:"setup_files"`  // Config files (playwright.config.ts, etc.)
	RunCmd      string            `json:"run_cmd"`      // Command to run tests
	CIConfig    string            `json:"ci_config"`    // CI/CD step for this test type
}

// TestCoverage represents test coverage analysis
type TestCoverage struct {
	TotalFiles     int                `json:"total_files"`
	TestedFiles    int                `json:"tested_files"`
	CoveragePercent float64           `json:"coverage_percent"`
	UncoveredFiles []string           `json:"uncovered_files"`
	Recommendations []string          `json:"recommendations"`
}

// SecurityScanResult represents security scan configuration and findings
type SecurityScanResult struct {
	ScanType       string             `json:"scan_type"`    // "sast", "dast", "dependency", "secrets"
	Tool           string             `json:"tool"`         // "trivy", "snyk", "gitleaks", "zap"
	ConfigFiles    map[string]string  `json:"config_files"`
	RunCmd         string             `json:"run_cmd"`
	CIConfig       string             `json:"ci_config"`
}

// ============================================================================
// Input Types
// ============================================================================

// IntegrationTestInput for generating API integration tests
type IntegrationTestInput struct {
	Stack       string            `json:"stack"`        // "node", "python", "go"
	Framework   string            `json:"framework"`    // "express", "fastapi", "gin"
	APISpec     *APISpec          `json:"api_spec,omitempty"`
	CodeFiles   map[string]string `json:"code_files"`
	DatabaseType string           `json:"database_type,omitempty"`
}

// E2ETestInput for generating end-to-end tests
type E2ETestInput struct {
	Framework    string            `json:"framework"`    // "playwright", "cypress"
	BaseURL      string            `json:"base_url"`
	UserFlows    []string          `json:"user_flows"`   // Key user journeys to test
	CodeFiles    map[string]string `json:"code_files"`
	AuthRequired bool              `json:"auth_required"`
}

// PerformanceTestInput for generating load tests
type PerformanceTestInput struct {
	Tool         string   `json:"tool"`          // "k6", "artillery"
	BaseURL      string   `json:"base_url"`
	Endpoints    []string `json:"endpoints"`     // Endpoints to test
	TargetRPS    int      `json:"target_rps"`    // Target requests per second
	Duration     string   `json:"duration"`      // Test duration (e.g., "5m")
	VirtualUsers int      `json:"virtual_users"` // Concurrent users
}

// SecurityTestInput for generating security test configurations
type SecurityTestInput struct {
	ScanTypes    []string          `json:"scan_types"`   // "sast", "dast", "dependency", "secrets"
	Stack        string            `json:"stack"`
	CodeFiles    map[string]string `json:"code_files,omitempty"`
	BaseURL      string            `json:"base_url,omitempty"` // For DAST
}

// AccessibilityTestInput for generating a11y tests
type AccessibilityTestInput struct {
	Tool        string   `json:"tool"`         // "axe", "pa11y"
	BaseURL     string   `json:"base_url"`
	Pages       []string `json:"pages"`        // Pages to test
	Standard    string   `json:"standard"`     // "WCAG2A", "WCAG2AA", "WCAG2AAA"
}

// BackendTestInput for generating backend unit tests
type BackendTestInput struct {
	Stack       string            `json:"stack"`       // "node", "python", "go"
	Framework   string            `json:"framework"`   // "express", "fastapi", "gin"
	CodeFiles   map[string]string `json:"code_files"`
	TestFramework string          `json:"test_framework,omitempty"` // "jest", "pytest", "go test"
}

// ============================================================================
// Activity Methods
// ============================================================================

// GenerateIntegrationTests creates API integration tests
func (a *EnhancedQAAgent) GenerateIntegrationTests(ctx context.Context, input map[string]interface{}) (*TestBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req IntegrationTestInput
	json.Unmarshal(inputBytes, &req)

	stack := strings.ToLower(req.Stack)
	if stack == "" {
		stack = "node"
	}

	sysPrompt := a.getIntegrationTestPrompt(stack, req.Framework)

	codeContext := buildCodeContext(req.CodeFiles)

	apiSpecHint := ""
	if req.APISpec != nil {
		specBytes, _ := json.MarshalIndent(req.APISpec, "", "  ")
		apiSpecHint = fmt.Sprintf("\n\nAPI SPECIFICATION:\n%s", string(specBytes))
	}

	prompt := fmt.Sprintf(`Stack: %s
Framework: %s
Database: %s%s

%s

Generate comprehensive integration tests for the API endpoints.`,
		stack, req.Framework, req.DatabaseType, apiSpecHint, codeContext)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("integration test generation failed: %w", err)
	}

	var bundle TestBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse integration tests: %w", err)
	}

	bundle.TestType = "integration"
	return &bundle, nil
}

// GenerateE2ETests creates end-to-end tests using Playwright or Cypress
func (a *EnhancedQAAgent) GenerateE2ETests(ctx context.Context, input map[string]interface{}) (*TestBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req E2ETestInput
	json.Unmarshal(inputBytes, &req)

	framework := strings.ToLower(req.Framework)
	if framework == "" {
		framework = "playwright"
	}

	sysPrompt := a.getE2ETestPrompt(framework)

	userFlowsHint := ""
	if len(req.UserFlows) > 0 {
		userFlowsHint = "\n\nUser Flows to Test:\n- " + strings.Join(req.UserFlows, "\n- ")
	}

	codeContext := buildCodeContext(req.CodeFiles)

	prompt := fmt.Sprintf(`Framework: %s
Base URL: %s
Auth Required: %v%s

%s

Generate comprehensive E2E tests covering critical user journeys.`,
		framework, req.BaseURL, req.AuthRequired, userFlowsHint, codeContext)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("E2E test generation failed: %w", err)
	}

	var bundle TestBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse E2E tests: %w", err)
	}

	bundle.TestType = "e2e"
	bundle.Framework = framework
	return &bundle, nil
}

// GeneratePerformanceTests creates load and performance tests
func (a *EnhancedQAAgent) GeneratePerformanceTests(ctx context.Context, input map[string]interface{}) (*TestBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req PerformanceTestInput
	json.Unmarshal(inputBytes, &req)

	tool := strings.ToLower(req.Tool)
	if tool == "" {
		tool = "k6"
	}

	targetRPS := req.TargetRPS
	if targetRPS <= 0 {
		targetRPS = 100
	}

	duration := req.Duration
	if duration == "" {
		duration = "5m"
	}

	vus := req.VirtualUsers
	if vus <= 0 {
		vus = 50
	}

	sysPrompt := a.getPerformanceTestPrompt(tool)

	endpointsHint := ""
	if len(req.Endpoints) > 0 {
		endpointsHint = "\n\nEndpoints to test:\n- " + strings.Join(req.Endpoints, "\n- ")
	}

	prompt := fmt.Sprintf(`Tool: %s
Base URL: %s
Target RPS: %d
Duration: %s
Virtual Users: %d%s

Generate performance tests with realistic load patterns.`,
		tool, req.BaseURL, targetRPS, duration, vus, endpointsHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("performance test generation failed: %w", err)
	}

	var bundle TestBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse performance tests: %w", err)
	}

	bundle.TestType = "performance"
	bundle.Framework = tool
	return &bundle, nil
}

// GenerateSecurityTests creates security scanning configurations
func (a *EnhancedQAAgent) GenerateSecurityTests(ctx context.Context, input map[string]interface{}) (*SecurityScanResult, error) {
	inputBytes, _ := json.Marshal(input)
	var req SecurityTestInput
	json.Unmarshal(inputBytes, &req)

	if len(req.ScanTypes) == 0 {
		req.ScanTypes = []string{"dependency", "sast", "secrets"}
	}

	sysPrompt := a.getSecurityTestPrompt(req.ScanTypes, req.Stack)

	codeContext := ""
	if len(req.CodeFiles) > 0 {
		codeContext = buildCodeContext(req.CodeFiles)
	}

	prompt := fmt.Sprintf(`Stack: %s
Scan Types: %s
Base URL: %s

%s

Generate security scanning configuration and scripts.`,
		req.Stack, strings.Join(req.ScanTypes, ", "), req.BaseURL, codeContext)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("security test generation failed: %w", err)
	}

	var result SecurityScanResult
	if err := extractJSON(resp.Response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse security tests: %w", err)
	}

	return &result, nil
}

// GenerateAccessibilityTests creates accessibility testing configurations
func (a *EnhancedQAAgent) GenerateAccessibilityTests(ctx context.Context, input map[string]interface{}) (*TestBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req AccessibilityTestInput
	json.Unmarshal(inputBytes, &req)

	tool := strings.ToLower(req.Tool)
	if tool == "" {
		tool = "axe"
	}

	standard := req.Standard
	if standard == "" {
		standard = "WCAG2AA"
	}

	sysPrompt := a.getAccessibilityTestPrompt(tool, standard)

	pagesHint := ""
	if len(req.Pages) > 0 {
		pagesHint = "\n\nPages to test:\n- " + strings.Join(req.Pages, "\n- ")
	}

	prompt := fmt.Sprintf(`Tool: %s
Base URL: %s
Standard: %s%s

Generate accessibility tests for all specified pages.`,
		tool, req.BaseURL, standard, pagesHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("accessibility test generation failed: %w", err)
	}

	var bundle TestBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse accessibility tests: %w", err)
	}

	bundle.TestType = "accessibility"
	bundle.Framework = tool
	return &bundle, nil
}

// GenerateBackendTests creates backend unit tests for Node.js, Python, or Go
func (a *EnhancedQAAgent) GenerateBackendTests(ctx context.Context, input map[string]interface{}) (*TestBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req BackendTestInput
	json.Unmarshal(inputBytes, &req)

	stack := strings.ToLower(req.Stack)
	if stack == "" {
		stack = "node"
	}

	testFramework := req.TestFramework
	if testFramework == "" {
		testFramework = getDefaultTestFramework(stack)
	}

	sysPrompt := a.getBackendTestPrompt(stack, testFramework)

	codeContext := buildCodeContext(req.CodeFiles)

	prompt := fmt.Sprintf(`Stack: %s
Framework: %s
Test Framework: %s

%s

Generate comprehensive unit tests for all services, controllers, and utilities.`,
		stack, req.Framework, testFramework, codeContext)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("backend test generation failed: %w", err)
	}

	var bundle TestBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse backend tests: %w", err)
	}

	bundle.TestType = "unit"
	bundle.Framework = testFramework
	return &bundle, nil
}

// AnalyzeTestCoverage analyzes test coverage and provides recommendations
func (a *EnhancedQAAgent) AnalyzeTestCoverage(ctx context.Context, input map[string]interface{}) (*TestCoverage, error) {
	codeFiles, _ := input["code_files"].(map[string]interface{})
	testFiles, _ := input["test_files"].(map[string]interface{})

	sysPrompt := `You are a QA architect analyzing test coverage.
Analyze the provided code and test files to determine coverage.

Output ONLY valid JSON with this structure:
{
  "total_files": 10,
  "tested_files": 7,
  "coverage_percent": 70.0,
  "uncovered_files": ["path/to/file1.js", "path/to/file2.js"],
  "recommendations": [
    "Add tests for user authentication flows",
    "Add edge case tests for error handling"
  ]
}

Coverage Analysis Guidelines:
1. Count source files (excluding tests, configs, types)
2. Identify which source files have corresponding tests
3. Calculate coverage percentage
4. List uncovered files
5. Provide actionable recommendations`

	codeFilesStr, _ := json.MarshalIndent(codeFiles, "", "  ")
	testFilesStr, _ := json.MarshalIndent(testFiles, "", "  ")

	prompt := fmt.Sprintf(`SOURCE FILES:
%s

TEST FILES:
%s

Analyze test coverage and provide recommendations.`, string(codeFilesStr), string(testFilesStr))

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("coverage analysis failed: %w", err)
	}

	var coverage TestCoverage
	if err := extractJSON(resp.Response, &coverage); err != nil {
		return nil, fmt.Errorf("failed to parse coverage analysis: %w", err)
	}

	return &coverage, nil
}

// ============================================================================
// Prompt Generators
// ============================================================================

func (a *EnhancedQAAgent) getIntegrationTestPrompt(stack, framework string) string {
	basePrompt := `You are a Senior QA Engineer creating API integration tests.

Output ONLY valid JSON with this structure:
{
  "test_type": "integration",
  "framework": "supertest",
  "files": {
    "tests/integration/users.test.js": "test content"
  },
  "setup_files": {
    "tests/setup.js": "test setup content"
  },
  "run_cmd": "npm run test:integration",
  "ci_config": "steps for CI"
}

Integration Test Best Practices:
1. Test all HTTP methods (GET, POST, PUT, DELETE)
2. Test success and error responses
3. Test authentication and authorization
4. Test input validation
5. Test edge cases and boundary conditions
6. Use fixtures for test data
7. Clean up test data after each test
8. Test database transactions
9. Mock external services
10. Test rate limiting if applicable
`

	switch stack {
	case "node":
		basePrompt += `
Node.js Integration Test Guidelines:
- Use supertest for HTTP assertions
- Use Jest or Mocha as test runner
- Create test database connection
- Use beforeAll/afterAll for setup/teardown
- Mock external APIs with nock
`
	case "python":
		basePrompt += `
Python Integration Test Guidelines:
- Use pytest with pytest-asyncio
- Use httpx or TestClient for FastAPI
- Use fixtures for database setup
- Use factory_boy for test data
- Mock external services with responses
`
	case "go":
		basePrompt += `
Go Integration Test Guidelines:
- Use testing package with httptest
- Create test server with actual routes
- Use testify for assertions
- Use sqlmock for database mocking
- Use gock for HTTP mocking
`
	}

	return basePrompt
}

func (a *EnhancedQAAgent) getE2ETestPrompt(framework string) string {
	basePrompt := `You are a Senior QA Engineer creating end-to-end tests.

Output ONLY valid JSON with this structure:
{
  "test_type": "e2e",
  "framework": "playwright",
  "files": {
    "tests/e2e/auth.spec.ts": "test content",
    "tests/e2e/checkout.spec.ts": "test content"
  },
  "setup_files": {
    "playwright.config.ts": "playwright config",
    "tests/fixtures/auth.ts": "auth fixture"
  },
  "run_cmd": "npx playwright test",
  "ci_config": "CI steps"
}

E2E Test Best Practices:
1. Test critical user journeys end-to-end
2. Use page object model for maintainability
3. Handle authentication with fixtures
4. Use data-testid for stable selectors
5. Test across multiple browsers
6. Handle flaky tests with retries
7. Take screenshots on failure
8. Test responsive design
9. Test error states
10. Keep tests independent
`

	switch framework {
	case "playwright":
		basePrompt += `
Playwright Guidelines:
- Use TypeScript for type safety
- Create page objects in tests/pages/
- Use test.describe for grouping
- Use expect assertions
- Use test.beforeEach for setup
- Configure multiple projects for browsers
- Use fixtures for authentication
- Enable tracing for debugging
`
	case "cypress":
		basePrompt += `
Cypress Guidelines:
- Use TypeScript with cypress types
- Create commands in cypress/support/
- Use cy.intercept for API mocking
- Use data-cy attributes for selectors
- Configure retries for flaky tests
- Use cypress-axe for a11y
- Take screenshots and videos
`
	}

	return basePrompt
}

func (a *EnhancedQAAgent) getPerformanceTestPrompt(tool string) string {
	basePrompt := `You are a Performance Engineer creating load tests.

Output ONLY valid JSON with this structure:
{
  "test_type": "performance",
  "framework": "k6",
  "files": {
    "tests/load/api-load.js": "k6 script",
    "tests/load/stress.js": "stress test script"
  },
  "setup_files": {
    "tests/load/config.json": "test config"
  },
  "run_cmd": "k6 run tests/load/api-load.js",
  "ci_config": "CI steps"
}

Performance Test Best Practices:
1. Define realistic load patterns
2. Include ramp-up and ramp-down periods
3. Test multiple scenarios (smoke, load, stress, spike)
4. Set appropriate thresholds
5. Monitor response times (p50, p95, p99)
6. Test with realistic data
7. Include think time between requests
8. Test authentication flows
9. Monitor error rates
10. Generate HTML reports
`

	switch tool {
	case "k6":
		basePrompt += `
k6 Guidelines:
- Use scenarios for different load patterns
- Define thresholds for SLOs
- Use checks for assertions
- Use groups for organizing requests
- Export metrics to InfluxDB/Prometheus
- Use __ENV for configuration
- Implement data parameterization
- Use sleep() for think time
`
	case "artillery":
		basePrompt += `
Artillery Guidelines:
- Use YAML for configuration
- Define phases for load pattern
- Use CSV for test data
- Configure capture for dynamic data
- Set response time thresholds
- Use plugins for extended functionality
- Generate HTML reports
`
	}

	return basePrompt
}

func (a *EnhancedQAAgent) getSecurityTestPrompt(scanTypes []string, stack string) string {
	basePrompt := `You are a Security Engineer creating security scanning configurations.

Output ONLY valid JSON with this structure:
{
  "scan_type": "sast",
  "tool": "trivy",
  "config_files": {
    ".trivyignore": "ignore rules",
    "trivy.yaml": "trivy config"
  },
  "run_cmd": "trivy fs --config trivy.yaml .",
  "ci_config": "CI steps for security scanning"
}

Security Scanning Best Practices:
1. Scan dependencies for vulnerabilities
2. Scan code for security issues (SAST)
3. Scan containers for vulnerabilities
4. Scan for secrets in code
5. Run DAST against running application
6. Set severity thresholds
7. Generate SARIF reports for GitHub
8. Fail builds on critical vulnerabilities
9. Configure ignore rules carefully
10. Run scans in CI/CD pipeline
`

	// Add scan-type specific guidance
	for _, scanType := range scanTypes {
		switch scanType {
		case "dependency":
			basePrompt += `
Dependency Scanning:
- Use npm audit / pip-audit / govulncheck
- Use Snyk or Trivy for comprehensive scanning
- Set up automatic PRs for updates
- Configure severity thresholds
`
		case "sast":
			basePrompt += `
SAST (Static Analysis):
- Use Semgrep for custom rules
- Use language-specific linters (eslint-plugin-security, bandit)
- Configure rule severity levels
- Integrate with IDE for early detection
`
		case "secrets":
			basePrompt += `
Secrets Detection:
- Use gitleaks or trufflehog
- Scan entire git history
- Configure .gitleaksignore for false positives
- Run as pre-commit hook
`
		case "dast":
			basePrompt += `
DAST (Dynamic Analysis):
- Use OWASP ZAP for automated scanning
- Configure authentication
- Set appropriate scan intensity
- Generate compliance reports
`
		}
	}

	return basePrompt
}

func (a *EnhancedQAAgent) getAccessibilityTestPrompt(tool, standard string) string {
	basePrompt := fmt.Sprintf(`You are an Accessibility Expert creating a11y tests.

Output ONLY valid JSON with this structure:
{
  "test_type": "accessibility",
  "framework": "%s",
  "files": {
    "tests/a11y/pages.test.js": "a11y test content"
  },
  "setup_files": {
    "tests/a11y/config.js": "axe config"
  },
  "run_cmd": "npm run test:a11y",
  "ci_config": "CI steps"
}

Accessibility Testing for %s:
1. Test all pages against %s standard
2. Test keyboard navigation
3. Test screen reader compatibility
4. Test color contrast
5. Test focus management
6. Test form labels and errors
7. Test ARIA attributes
8. Test responsive accessibility
9. Test media alternatives
10. Generate compliance reports
`, tool, standard, standard)

	switch tool {
	case "axe":
		basePrompt += `
axe-core Guidelines:
- Use @axe-core/playwright or cypress-axe
- Configure rules for your standard
- Exclude known issues with selectors
- Test different viewport sizes
- Generate JSON reports
- Integrate with CI/CD
`
	case "pa11y":
		basePrompt += `
Pa11y Guidelines:
- Configure with pa11y.json
- Use pa11y-ci for CI/CD
- Set standard to WCAG2AA
- Configure timeout for SPAs
- Generate HTML reports
- Test sitemap URLs
`
	}

	return basePrompt
}

func (a *EnhancedQAAgent) getBackendTestPrompt(stack, testFramework string) string {
	basePrompt := `You are a Senior Backend Engineer creating comprehensive unit tests.

Output ONLY valid JSON with this structure:
{
  "test_type": "unit",
  "framework": "jest",
  "files": {
    "tests/services/user.test.js": "test content",
    "tests/controllers/auth.test.js": "test content"
  },
  "setup_files": {
    "tests/setup.js": "test setup"
  },
  "run_cmd": "npm test",
  "ci_config": "CI steps"
}

Unit Test Best Practices:
1. Test one thing per test
2. Use descriptive test names
3. Follow AAA pattern (Arrange, Act, Assert)
4. Test edge cases and error conditions
5. Mock external dependencies
6. Test both success and failure paths
7. Use fixtures for test data
8. Aim for high coverage (>80%)
9. Keep tests fast and isolated
10. Test public interfaces, not implementation
`

	switch stack {
	case "node":
		basePrompt += fmt.Sprintf(`
Node.js (%s) Testing Guidelines:
- Use describe/it blocks for organization
- Use beforeEach/afterEach for setup/teardown
- Mock modules with jest.mock()
- Use expect assertions
- Test async functions with async/await
- Mock database with jest-mock
- Test middleware functions
- Test error handling
`, testFramework)
	case "python":
		basePrompt += `
Python (pytest) Testing Guidelines:
- Use fixtures for setup
- Use parametrize for multiple inputs
- Use pytest-mock for mocking
- Use pytest-asyncio for async tests
- Test with pytest-cov for coverage
- Use conftest.py for shared fixtures
- Test exception handling with pytest.raises
`
	case "go":
		basePrompt += `
Go (testing) Testing Guidelines:
- Use table-driven tests
- Use testify for assertions
- Use gomock for mocking interfaces
- Test with -race flag
- Use t.Parallel() for parallel tests
- Test error returns
- Use subtests with t.Run()
- Generate coverage with -coverprofile
`
	}

	return basePrompt
}

// ============================================================================
// Helper Functions
// ============================================================================

func buildCodeContext(codeFiles map[string]string) string {
	if len(codeFiles) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("APPLICATION CODE:\n")

	for name, content := range codeFiles {
		// Truncate long files
		truncated := content
		if len(content) > 2000 {
			truncated = content[:2000] + "\n... (truncated)"
		}
		sb.WriteString(fmt.Sprintf("File: %s\n```\n%s\n```\n\n", name, truncated))
	}

	return sb.String()
}

func getDefaultTestFramework(stack string) string {
	switch stack {
	case "node":
		return "jest"
	case "python":
		return "pytest"
	case "go":
		return "go test"
	default:
		return "jest"
	}
}

// GetTestPyramidRecommendations returns recommended test distribution
func GetTestPyramidRecommendations() map[string]int {
	return map[string]int{
		"unit":          70, // 70% unit tests
		"integration":   20, // 20% integration tests
		"e2e":           10, // 10% E2E tests
	}
}

// GetTestFrameworksByStack returns recommended test frameworks per stack
func GetTestFrameworksByStack(stack string) map[string]string {
	switch stack {
	case "node":
		return map[string]string{
			"unit":        "jest",
			"integration": "supertest",
			"e2e":         "playwright",
			"performance": "k6",
		}
	case "python":
		return map[string]string{
			"unit":        "pytest",
			"integration": "pytest + httpx",
			"e2e":         "playwright",
			"performance": "locust",
		}
	case "go":
		return map[string]string{
			"unit":        "go test + testify",
			"integration": "go test + httptest",
			"e2e":         "playwright",
			"performance": "k6",
		}
	default:
		return map[string]string{
			"unit":        "jest",
			"integration": "supertest",
			"e2e":         "playwright",
			"performance": "k6",
		}
	}
}

// GenerateTestConfig generates a test configuration file for common setups
func GenerateTestConfig(testType, framework string) string {
	switch testType {
	case "e2e":
		if framework == "playwright" {
			return `import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
});
`
		}
	case "performance":
		if framework == "k6" {
			return `export const options = {
  scenarios: {
    smoke: {
      executor: 'constant-vus',
      vus: 1,
      duration: '1m',
    },
    load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 50 },
        { duration: '5m', target: 50 },
        { duration: '2m', target: 0 },
      ],
    },
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 100 },
        { duration: '5m', target: 100 },
        { duration: '2m', target: 200 },
        { duration: '5m', target: 200 },
        { duration: '2m', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],
  },
};
`
		}
	}

	return ""
}
