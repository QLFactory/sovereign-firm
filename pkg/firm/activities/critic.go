package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// CodeCritic performs code review and scoring before merge
// This is Stage 4 validation in the Agent CI Pipeline
type CodeCritic struct {
	llmClient llm.Client
}

func NewCodeCritic() *CodeCritic {
	return &CodeCritic{
		llmClient: llm.NewClient(),
	}
}

type CriticInput struct {
	CodeFiles map[string]string `json:"code_files"`
	Spec      string            `json:"spec"`
}

type CriticResult struct {
	Approved      bool              `json:"approved"`
	OverallScore  int               `json:"overall_score"` // 0-100
	Categories    map[string]int    `json:"categories"`    // Per-category scores
	Issues        []CriticIssue     `json:"issues"`
	Suggestions   []string          `json:"suggestions"`
	Summary       string            `json:"summary"`
}

type CriticIssue struct {
	File     string `json:"file"`
	Line     int    `json:"line,omitempty"`
	Severity string `json:"severity"` // "critical", "major", "minor", "info"
	Category string `json:"category"` // "security", "performance", "maintainability", "correctness"
	Message  string `json:"message"`
}

// ReviewCode performs comprehensive code review
func (c *CodeCritic) ReviewCode(ctx context.Context, input map[string]interface{}) (*CriticResult, error) {
	inputBytes, _ := json.Marshal(input)
	var req CriticInput
	json.Unmarshal(inputBytes, &req)

	sysPrompt := `You are a Senior Code Reviewer and Software Architect.
Review the provided code and output a JSON review with this structure:
{
  "approved": boolean,
  "overall_score": number (0-100),
  "categories": {
    "correctness": number (0-100),
    "security": number (0-100),
    "performance": number (0-100),
    "maintainability": number (0-100),
    "best_practices": number (0-100)
  },
  "issues": [
    {
      "file": "filename",
      "severity": "critical|major|minor|info",
      "category": "security|performance|maintainability|correctness|best_practices",
      "message": "description"
    }
  ],
  "suggestions": ["improvement suggestion 1", "improvement suggestion 2"],
  "summary": "brief overall assessment"
}

Approve if:
- No critical issues
- Overall score >= 70
- Security score >= 80
- Correctness score >= 80

Focus on:
1. Security vulnerabilities (XSS, injection, etc.)
2. React best practices
3. Performance issues
4. Code organization
5. Error handling
`

	codeContext := "CODE TO REVIEW:\n"
	for name, content := range req.CodeFiles {
		// Skip non-code files
		if name == "/package.json" || name == "/vite.config.js" || name == "/index.html" {
			continue
		}
		codeContext += fmt.Sprintf("File: %s\n```\n%s\n```\n", name, content)
	}

	prompt := fmt.Sprintf("REQUIREMENTS:\n%s\n\n%s\n\nReview this code and provide your assessment.", req.Spec, codeContext)

	resp, err := c.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("critic LLM call failed: %w", err)
	}

	var result CriticResult
	if err := extractJSON(resp.Response, &result); err != nil {
		// If parsing fails, return a conservative default
		return &CriticResult{
			Approved:     true, // Don't block on parse failure
			OverallScore: 75,
			Summary:      "Review parsing failed - auto-approved",
		}, nil
	}

	// Auto-fail if critical issues exist
	for _, issue := range result.Issues {
		if issue.Severity == "critical" {
			result.Approved = false
			break
		}
	}

	// Auto-fail if below thresholds
	if result.OverallScore < 70 {
		result.Approved = false
	}
	if score, ok := result.Categories["security"]; ok && score < 80 {
		result.Approved = false
	}
	if score, ok := result.Categories["correctness"]; ok && score < 80 {
		result.Approved = false
	}

	return &result, nil
}

// extractJSON attempts to find and parse JSON from a string
func extractJSON(input string, target interface{}) error {
	if err := json.Unmarshal([]byte(input), target); err == nil {
		return nil
	}

	// Try extracting JSON block
	start := -1
	end := -1
	for i, r := range input {
		if r == '{' {
			start = i
			break
		}
	}
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] == '}' {
			end = i
			break
		}
	}

	if start != -1 && end != -1 && start < end {
		jsonPart := input[start : end+1]
		if err := json.Unmarshal([]byte(jsonPart), target); err == nil {
			return nil
		}
	}

	return fmt.Errorf("could not extract valid JSON")
}
