package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/sandbox"
)

type TestRunner struct {
	sandboxManager *sandbox.Manager
}

func NewTestRunner(mgr *sandbox.Manager) *TestRunner {
	return &TestRunner{sandboxManager: mgr}
}

type TestRunInput struct {
	CodeFiles map[string]string `json:"code_files"`
}

type TestResult struct {
	Success     bool     `json:"success"`
	Output      string   `json:"output"`
	ExitCode    int      `json:"exit_code"`
	FailedTests []string `json:"failed_tests,omitempty"`
}

// DefaultPackageJSON provides a known-working package.json for Vite+React+Vitest
var DefaultPackageJSON = map[string]interface{}{
	"name":    "test-runner-app",
	"private": true,
	"version": "0.0.0",
	"type":    "module",
	"scripts": map[string]string{
		"test": "vitest run --reporter=verbose",
	},
	"dependencies": map[string]string{
		"react":            "^18.2.0",
		"react-dom":        "^18.2.0",
		"react-router-dom": "^6.22.0",
	},
	"devDependencies": map[string]string{
		"@testing-library/jest-dom": "^6.4.2",
		"@testing-library/react":    "^14.2.1",
		"@vitejs/plugin-react":      "^4.2.1",
		"jsdom":                     "^24.0.0",
		"vite":                      "^5.1.0",
		"vitest":                    "^1.3.1",
	},
}

// DefaultViteConfig provides vitest configuration
var DefaultViteConfig = `import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/setupTests.js',
  },
})`

// DefaultSetupTests for jest-dom matchers
var DefaultSetupTests = `import '@testing-library/jest-dom'`

// RunTests executes Vitest tests on the provided code files
// Returns test results including pass/fail status and output
func (t *TestRunner) RunTests(ctx context.Context, input map[string]interface{}) (*TestResult, error) {
	// Manual unmarshal
	inputBytes, _ := json.Marshal(input)
	var req TestRunInput
	json.Unmarshal(inputBytes, &req)

	// Check if there are any test files
	hasTests := false
	for filename := range req.CodeFiles {
		if strings.Contains(filename, ".test.") {
			hasTests = true
			break
		}
	}

	if !hasTests {
		return &TestResult{
			Success:  true,
			Output:   "No test files found - skipping test execution",
			ExitCode: 0,
		}, nil
	}

	// Create sandbox configuration
	// ISS-026: Network disabled by default for security - tests should not have network access
	// This prevents test code from exfiltrating data or attacking internal services.
	// npm install uses --prefer-offline to use cached packages.
	// For production, ensure npm packages are pre-cached or use a private registry.
	config := sandbox.DefaultConfig(sandbox.LevelContainer)
	config.ID = "test-" + strings.ReplaceAll(time.Now().Format("20060102-150405.000"), ".", "-")
	config.ProjectID = "test-runner-app"
	config.Language = "javascript"
	config.ReadOnlyRoot = false
	config.NetworkEnabled = false // ISS-026: Disabled for security
	config.MaxMemory = 1024

	// Create and start sandbox
	sb, err := t.sandboxManager.Create(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}
	defer t.sandboxManager.Destroy(ctx, config.ID)

	// Prepare files for execution
	files := make(map[string]string)
	for filename, content := range req.CodeFiles {
		files[filename] = content
	}

	// Merge dependencies if package.json exists
	pkgJSON := make(map[string]interface{})
	if content, ok := req.CodeFiles["/package.json"]; ok {
		json.Unmarshal([]byte(content), &pkgJSON)
	} else if content, ok := req.CodeFiles["package.json"]; ok {
		json.Unmarshal([]byte(content), &pkgJSON)
	}

	// Ensure basic structure
	if pkgJSON["dependencies"] == nil {
		pkgJSON["dependencies"] = make(map[string]interface{})
	}
	if pkgJSON["devDependencies"] == nil {
		pkgJSON["devDependencies"] = make(map[string]interface{})
	}
	if pkgJSON["scripts"] == nil {
		pkgJSON["scripts"] = make(map[string]interface{})
	}

	deps := pkgJSON["dependencies"].(map[string]interface{})
	devDeps := pkgJSON["devDependencies"].(map[string]interface{})
	scripts := pkgJSON["scripts"].(map[string]interface{})

	// Add test-specific requirements
	scripts["test"] = "vitest run --reporter=verbose"

	// Merge from DefaultPackageJSON
	for k, v := range DefaultPackageJSON["dependencies"].(map[string]string) {
		deps[k] = v
	}
	for k, v := range DefaultPackageJSON["devDependencies"].(map[string]string) {
		devDeps[k] = v
	}

	// Add common UI and testing libraries that LLMs often use but might forget to list
	commonUI := map[string]string{
		"framer-motion":         "11.0.8",
		"lucide-react":          "0.344.0",
		"clsx":                  "2.1.0",
		"tailwind-merge":        "2.2.1",
		"react-query":           "3.39.3",
		"@tanstack/react-query": "5.28.4",
		"supertest":             "6.34.0",
		"@playwright/test":      "1.42.1",
	}
	for k, v := range commonUI {
		if deps[k] == nil {
			deps[k] = v
		}
	}

	finalPkgJSON, _ := json.MarshalIndent(pkgJSON, "", "  ")
	files["/package.json"] = string(finalPkgJSON)
	files["/vite.config.js"] = DefaultViteConfig
	files["/src/setupTests.js"] = DefaultSetupTests

	// Run npm install in sandbox
	installReq := &sandbox.ExecuteRequest{
		Command: "npm install --prefer-offline --silent",
		Files:   files,
	}

	installResult, err := sb.Execute(ctx, installReq)
	if err != nil || !installResult.Success {
		output := ""
		if installResult != nil {
			output = installResult.Stdout + installResult.Stderr
		}
		return &TestResult{
			Success:  false,
			Output:   fmt.Sprintf("npm install failed:\n%s", output),
			ExitCode: 1,
		}, nil
	}

	// Run tests in sandbox
	testReq := &sandbox.ExecuteRequest{
		Command: "npm test",
	}

	testResult, err := sb.Execute(ctx, testReq)
	if err != nil {
		return nil, fmt.Errorf("failed to run tests in sandbox: %w", err)
	}

	output := testResult.Stdout + testResult.Stderr
	success := testResult.Success
	exitCode := testResult.ExitCode

	// Extract failed test names from output for feedback
	failedTests := extractFailedTests(output)

	return &TestResult{
		Success:     success,
		Output:      output,
		ExitCode:    exitCode,
		FailedTests: failedTests,
	}, nil
}

// extractFailedTests parses Vitest output to find failed test names
func extractFailedTests(output string) []string {
	var failed []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// Vitest marks failures with ❌ or FAIL
		if strings.Contains(line, "❌") || strings.Contains(line, "FAIL") {
			// Clean up the line and extract test name
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "FAIL") {
				failed = append(failed, trimmed)
			}
		}
	}

	return failed
}
