package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TestRunner struct{}

func NewTestRunner() *TestRunner {
	return &TestRunner{}
}

type TestRunInput struct {
	CodeFiles map[string]string `json:"code_files"`
}

type TestResult struct {
	Success   bool   `json:"success"`
	Output    string `json:"output"`
	ExitCode  int    `json:"exit_code"`
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

	// Create temporary directory for test execution
	tmpDir, err := os.MkdirTemp("", "test-runner-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write all code files
	for filename, content := range req.CodeFiles {
		// Normalize path - remove leading slash
		normalizedPath := filename
		if strings.HasPrefix(normalizedPath, "/") {
			normalizedPath = normalizedPath[1:]
		}

		filePath := filepath.Join(tmpDir, normalizedPath)

		// Create directory structure
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Write file
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
	}

	// Always use our known-working package.json (overwrite any LLM-generated one)
	packageJSONBytes, _ := json.MarshalIndent(DefaultPackageJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), packageJSONBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write package.json: %w", err)
	}

	// Always use our vite.config.js
	if err := os.WriteFile(filepath.Join(tmpDir, "vite.config.js"), []byte(DefaultViteConfig), 0644); err != nil {
		return nil, fmt.Errorf("failed to write vite.config.js: %w", err)
	}

	// Ensure setupTests.js exists
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	setupTestsPath := filepath.Join(srcDir, "setupTests.js")
	if _, err := os.Stat(setupTestsPath); os.IsNotExist(err) {
		if err := os.WriteFile(setupTestsPath, []byte(DefaultSetupTests), 0644); err != nil {
			return nil, fmt.Errorf("failed to write setupTests.js: %w", err)
		}
	}

	// Run npm install
	var installOut bytes.Buffer
	installCmd := exec.CommandContext(ctx, "npm", "install", "--prefer-offline")
	installCmd.Dir = tmpDir
	installCmd.Stdout = &installOut
	installCmd.Stderr = &installOut

	if err := installCmd.Run(); err != nil {
		return &TestResult{
			Success:  false,
			Output:   fmt.Sprintf("npm install failed:\n%s", installOut.String()),
			ExitCode: 1,
		}, nil
	}

	// Run tests
	var testOut bytes.Buffer
	testCmd := exec.CommandContext(ctx, "npm", "test")
	testCmd.Dir = tmpDir
	testCmd.Stdout = &testOut
	testCmd.Stderr = &testOut

	exitCode := 0
	err = testCmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to run tests: %w", err)
		}
	}

	output := testOut.String()
	success := exitCode == 0

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
