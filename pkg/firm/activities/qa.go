package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

type QAAgent struct {
	llmClient llm.Client
}

func NewQAAgent() *QAAgent {
	return &QAAgent{
		llmClient: llm.NewClient(),
	}
}

type QAGenerateInput struct {
	Spec      string            `json:"spec"`
	CodeFiles map[string]string `json:"code_files"`
}

// GenerateTests analyzes code and writes Jest tests
func (a *QAAgent) GenerateTests(ctx context.Context, input map[string]interface{}) (map[string]string, error) {
	// Manual unmarshal
	inputBytes, _ := json.Marshal(input)
	var req QAGenerateInput
	json.Unmarshal(inputBytes, &req)

	sysPrompt := `You are a Senior QA Engineer specialized in React Testing Library and Jest.
Your goal is to write comprehensive unit tests for the provided React application.
Output ONLY valid JSON.
The JSON must be a map where keys are filenames (MUST start with "/src/", e.g., "/src/App.test.js") and values are the file content.
Do not include markdown backticks.

RULES:
1. Use 'import { render, screen, fireEvent } from "@testing-library/react";'
2. Use 'import "@testing-library/jest-dom";'
3. Assume standard Jest environment.
4. Test for presence of key elements and basic interactions (clicks, etc).
5. Only write tests for components that exist in the input.
6. Return ONLY valid JSON.
7. When matching text with Regex, CAREFULLY ESCAPE special characters. Example: use /\-/i instead of /-/i.
8. Generate test files in the /src/ directory (e.g., '/src/App.test.js'), NOT in the root or subfolders. This ensures react-scripts/Sandpack finds them.
`

	codeContext := "APPLICATION CODE:\n"
	for name, content := range req.CodeFiles {
		codeContext += fmt.Sprintf("File: %s\n```\n%s\n```\n", name, content)
	}

	prompt := fmt.Sprintf("%s\n\nSPECIFICATION:\n%s\n\nGenerate the JSON of test files now.", codeContext, req.Spec)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("QA brain failed: %w", err)
	}

	var rawTests map[string]interface{}
	if err := a.extractAndUnmarshalJSON(resp.Response, &rawTests); err != nil {
		return nil, fmt.Errorf("failed to parse tests: %w", err)
	}

	tests := make(map[string]string)
	for k, v := range rawTests {
		// Enforce /src directory for tests (CRA requirement)
		filename := k
		// Simple basename logic
		if idx := len(filename) - 1; idx >= 0 {
			// find last slash
			lastSlash := -1
			for i := len(filename) - 1; i >= 0; i-- {
				if filename[i] == '/' {
					lastSlash = i
					break
				}
			}
			if lastSlash != -1 {
				filename = "/src/" + filename[lastSlash+1:]
			} else {
				// No slash, but ensure it's in /src/
				if len(filename) > 5 && filename[:5] == "/src/" {
					// already has prefix
				} else if filename[0] == '/' {
					filename = "/src" + filename
				} else {
					filename = "/src/" + filename
				}
			}
		}

		var content string
		switch val := v.(type) {
		case string:
			content = val
		default:
			marshaled, _ := json.MarshalIndent(val, "", "  ")
			content = string(marshaled)
		}
		tests[filename] = content
	}

	return tests, nil
}

// extractAndUnmarshalJSON attempts to find JSON within a string (handling markdown blocks)
func (a *QAAgent) extractAndUnmarshalJSON(input string, target interface{}) error {
	// 1. Try direct unmarshal
	if err := json.Unmarshal([]byte(input), target); err == nil {
		return nil
	}

	// 2. Try stripping markdown code blocks
	// Find first '{' and last '}'
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

	return fmt.Errorf("could not extract valid JSON from response")
}
