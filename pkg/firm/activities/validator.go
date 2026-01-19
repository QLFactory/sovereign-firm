package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/sandbox"
)

// ValidationStage1 performs instant syntax, type, and lint checks
// This is Stage 1 of the Agent CI Pipeline (< 5 seconds)
type Validator struct {
	sandboxManager *sandbox.Manager
}

func NewValidator(mgr *sandbox.Manager) *Validator {
	return &Validator{sandboxManager: mgr}
}

type ValidationInput struct {
	CodeFiles map[string]string `json:"code_files"`
}

type ValidationResult struct {
	Success      bool              `json:"success"`
	Stage        string            `json:"stage"`
	SyntaxErrors []ValidationError `json:"syntax_errors,omitempty"`
	TypeErrors   []ValidationError `json:"type_errors,omitempty"`
	LintWarnings []ValidationError `json:"lint_warnings,omitempty"`
	Summary      string            `json:"summary"`
	Feedback     string            `json:"feedback"` // Actionable feedback for agent
}

type ValidationError struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message"`
	Rule    string `json:"rule,omitempty"`
}

// ValidatePackageJSON with ESLint configuration included
var ValidatePackageJSON = map[string]interface{}{
	"name":    "validator-app",
	"private": true,
	"version": "0.0.0",
	"type":    "module",
	"scripts": map[string]string{
		"lint":       "eslint src/ --ext .js,.jsx,.ts,.tsx --format json",
		"type-check": "tsc --noEmit --skipLibCheck",
	},
	"dependencies": map[string]string{
		"react":            "^18.2.0",
		"react-dom":        "^18.2.0",
		"react-router-dom": "^6.22.0",
	},
	"devDependencies": map[string]string{
		"@types/react":              "^18.2.0",
		"@types/react-dom":          "^18.2.0",
		"@vitejs/plugin-react":      "^4.2.1",
		"eslint":                    "^8.57.0",
		"eslint-plugin-react":       "^7.33.2",
		"eslint-plugin-react-hooks": "^4.6.0",
		"typescript":                "^5.3.0",
		"vite":                      "^5.1.0",
	},
}

var ESLintConfig = `{
  "env": {
    "browser": true,
    "es2021": true
  },
  "extends": [
    "eslint:recommended",
    "plugin:react/recommended",
    "plugin:react-hooks/recommended"
  ],
  "parserOptions": {
    "ecmaFeatures": { "jsx": true },
    "ecmaVersion": "latest",
    "sourceType": "module"
  },
  "plugins": ["react", "react-hooks"],
  "settings": {
    "react": { "version": "detect" }
  },
  "rules": {
    "react/react-in-jsx-scope": "off",
    "react/prop-types": "off"
  }
}`

var TSConfig = `{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": false,
    "noUnusedLocals": false,
    "noUnusedParameters": false,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src"]
}`

// ValidateCode performs Stage 1 validation: syntax, types, and lint
func (v *Validator) ValidateCode(ctx context.Context, input map[string]interface{}) (*ValidationResult, error) {
	inputBytes, _ := json.Marshal(input)
	var req ValidationInput
	json.Unmarshal(inputBytes, &req)

	result := &ValidationResult{
		Success: true,
		Stage:   "validation",
	}

	// Create sandbox configuration
	config := sandbox.DefaultConfig(sandbox.LevelContainer)
	config.ID = "test-" + strings.ReplaceAll(time.Now().Format("20060102-150405.000"), ".", "-")
	config.ProjectID = "test-runner-app" // Fixed field name
	config.Language = "javascript"
	config.ReadOnlyRoot = false // Disabled to allow npm install to write to /root/.npm and /tmp
	config.NetworkEnabled = true
	config.MaxMemory = 1024 // Increased from 512 for better reliability

	// Create and start sandbox
	sb, err := v.sandboxManager.Create(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox: %w", err)
	}
	defer v.sandboxManager.Destroy(ctx, config.ID)

	// Prepare files
	files := make(map[string]string)
	for filename, content := range req.CodeFiles {
		files[filename] = content
	}

	// Always use our validation configs
	packageJSONBytes, _ := json.MarshalIndent(ValidatePackageJSON, "", "  ")
	files["/package.json"] = string(packageJSONBytes)
	files["/.eslintrc.json"] = ESLintConfig
	files["/tsconfig.json"] = TSConfig

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
		return &ValidationResult{
			Success:  false,
			Stage:    "install",
			Summary:  "Failed to install dependencies",
			Feedback: fmt.Sprintf("npm install failed: %s\nPlease check package dependencies.", output),
		}, nil
	}

	// Stage 1.1: ESLint (syntax + lint)
	lintReq := &sandbox.ExecuteRequest{
		Command: "npm run lint",
	}
	lintResult, err := sb.Execute(ctx, lintReq)
	if err != nil {
		return nil, fmt.Errorf("failed to run lint in sandbox: %w", err)
	}

	if !lintResult.Success {
		lintErrors := parseLintOutput(lintResult.Stdout + lintResult.Stderr)
		if len(lintErrors) > 0 {
			result.Success = false
			result.LintWarnings = lintErrors
		}
	}

	// Stage 1.2: TypeScript type check (for .ts/.tsx files)
	hasTypeScript := false
	for filename := range req.CodeFiles {
		if strings.HasSuffix(filename, ".ts") || strings.HasSuffix(filename, ".tsx") {
			hasTypeScript = true
			break
		}
	}

	if hasTypeScript {
		typeReq := &sandbox.ExecuteRequest{
			Command: "npm run type-check",
		}
		typeResult, err := sb.Execute(ctx, typeReq)
		if err != nil {
			return nil, fmt.Errorf("failed to run type-check in sandbox: %w", err)
		}

		if !typeResult.Success {
			typeErrors := parseTypeErrors(typeResult.Stdout + typeResult.Stderr)
			if len(typeErrors) > 0 {
				result.Success = false
				result.TypeErrors = typeErrors
			}
		}
	}

	// Generate summary and actionable feedback
	result.Summary, result.Feedback = generateValidationFeedback(result)

	return result, nil
}

// parseLintOutput extracts errors from ESLint output
func parseLintOutput(output string) []ValidationError {
	var errors []ValidationError

	// Try JSON parse first
	var eslintResult []struct {
		FilePath string `json:"filePath"`
		Messages []struct {
			RuleId   string `json:"ruleId"`
			Severity int    `json:"severity"`
			Message  string `json:"message"`
			Line     int    `json:"line"`
			Column   int    `json:"column"`
		} `json:"messages"`
	}

	if err := json.Unmarshal([]byte(output), &eslintResult); err == nil {
		for _, file := range eslintResult {
			for _, msg := range file.Messages {
				if msg.Severity >= 2 { // Error level
					errors = append(errors, ValidationError{
						File:    file.FilePath,
						Line:    msg.Line,
						Column:  msg.Column,
						Message: msg.Message,
						Rule:    msg.RuleId,
					})
				}
			}
		}
	} else {
		// Fallback: parse text output
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "error") {
				errors = append(errors, ValidationError{
					Message: strings.TrimSpace(line),
				})
			}
		}
	}

	return errors
}

// parseTypeErrors extracts TypeScript errors
func parseTypeErrors(output string) []ValidationError {
	var errors []ValidationError
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "error TS") {
			errors = append(errors, ValidationError{
				Message: strings.TrimSpace(line),
			})
		}
	}

	return errors
}

// generateValidationFeedback creates actionable feedback for the agent
func generateValidationFeedback(result *ValidationResult) (string, string) {
	if result.Success {
		return "All validation checks passed", ""
	}

	var issues []string
	var feedback strings.Builder

	if len(result.SyntaxErrors) > 0 {
		issues = append(issues, fmt.Sprintf("%d syntax errors", len(result.SyntaxErrors)))
		feedback.WriteString("SYNTAX ERRORS:\n")
		for _, e := range result.SyntaxErrors {
			feedback.WriteString(fmt.Sprintf("  - %s:%d: %s\n", e.File, e.Line, e.Message))
		}
		feedback.WriteString("\n")
	}

	if len(result.TypeErrors) > 0 {
		issues = append(issues, fmt.Sprintf("%d type errors", len(result.TypeErrors)))
		feedback.WriteString("TYPE ERRORS:\n")
		for _, e := range result.TypeErrors {
			feedback.WriteString(fmt.Sprintf("  - %s\n", e.Message))
		}
		feedback.WriteString("\n")
	}

	if len(result.LintWarnings) > 0 {
		issues = append(issues, fmt.Sprintf("%d lint issues", len(result.LintWarnings)))
		feedback.WriteString("LINT ISSUES:\n")
		for _, e := range result.LintWarnings {
			if e.Rule != "" {
				feedback.WriteString(fmt.Sprintf("  - %s:%d [%s]: %s\n", e.File, e.Line, e.Rule, e.Message))
			} else {
				feedback.WriteString(fmt.Sprintf("  - %s\n", e.Message))
			}
		}
		feedback.WriteString("\n")
	}

	feedback.WriteString("Please fix these issues and regenerate the code.")

	return fmt.Sprintf("Validation failed: %s", strings.Join(issues, ", ")), feedback.String()
}
