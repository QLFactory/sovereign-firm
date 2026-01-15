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

	sysPrompt := `You are a Senior QA Engineer specialized in React Testing Library and Vitest.
Your goal is to write comprehensive unit tests for the provided React application.
Output ONLY valid JSON.
The JSON must be a map where keys are filenames (MUST start with "/src/", e.g., "/src/App.test.jsx") and values are the file content.
Do not include markdown backticks.

CRITICAL RULES:
1. Use Vitest imports: import { describe, it, expect } from 'vitest';
2. Use Testing Library: import { render, screen, fireEvent } from '@testing-library/react';
3. Use jest-dom matchers: import '@testing-library/jest-dom';
4. IMPORTANT: Read the actual component code carefully - test what's actually rendered, not what you assume.
5. Look at the actual text, classNames, and element types in the components.
6. **ONLY generate test files for components that ACTUALLY EXIST in the provided file list.**
7. **DO NOT assume separate component files exist. Check the EXISTING FILES list below.**
8. **If a component is defined INSIDE App.jsx, test it via App.jsx - do NOT create a separate test file for it.**
9. Return ONLY valid JSON.
10. Use .jsx extension for test files (e.g., '/src/App.test.jsx').
11. Keep tests simple - test that components render and basic interactions work.

EXAMPLE TEST STRUCTURE:
import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import App from './App';

describe('App', () => {
  it('renders without crashing', () => {
    render(<App />);
    expect(document.body).toBeInTheDocument();
  });
});
`

	// Build list of existing files for the prompt
	fileList := "EXISTING FILES (ONLY test these files):\n"
	for name := range req.CodeFiles {
		fileList += fmt.Sprintf("- %s\n", name)
	}

	codeContext := "\nAPPLICATION CODE:\n"
	for name, content := range req.CodeFiles {
		codeContext += fmt.Sprintf("File: %s\n```\n%s\n```\n", name, content)
	}

	prompt := fmt.Sprintf("%s\n%s\n\nSPECIFICATION:\n%s\n\nGenerate test files ONLY for the files listed above. Do NOT create tests for files that don't exist.", fileList, codeContext, req.Spec)

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

// QARegenerateInput contains the context for regenerating failed tests
type QARegenerateInput struct {
	Spec        string            `json:"spec"`
	CodeFiles   map[string]string `json:"code_files"`
	TestFiles   map[string]string `json:"test_files"`
	TestOutput  string            `json:"test_output"`
	Attempt     int               `json:"attempt"`
}

// RegenerateTests takes failed test output and regenerates better tests
// This implements the feedback loop for the Three-Strike Rule
func (a *QAAgent) RegenerateTests(ctx context.Context, input map[string]interface{}) (map[string]string, error) {
	// Manual unmarshal
	inputBytes, _ := json.Marshal(input)
	var req QARegenerateInput
	json.Unmarshal(inputBytes, &req)

	sysPrompt := `You are a Senior QA Engineer specialized in React Testing Library and Vitest.
Your previous tests FAILED. You must analyze the error output and fix the tests.

CRITICAL: You are on attempt %d of 3. If you fail again, the project will be escalated.

Output ONLY valid JSON.
The JSON must be a map where keys are filenames (MUST start with "/src/", e.g., "/src/App.test.jsx") and values are the file content.
Do not include markdown backticks.

RULES:
1. Use Vitest imports: import { describe, it, expect } from 'vitest';
2. Use Testing Library: import { render, screen, fireEvent } from '@testing-library/react';
3. Use jest-dom matchers: import '@testing-library/jest-dom';
4. CAREFULLY read the ERROR OUTPUT to understand what went wrong.
5. CAREFULLY read the actual component code - test what's actually rendered, not what you assume.
6. Common issues to check:
   - Wrong text content (check exact strings in components)
   - Wrong selectors (use getByRole, getByText with exact matches)
   - Missing async handling (use waitFor for async operations)
   - Component not importing correctly (check file paths)
7. Return ONLY valid JSON with CORRECTED test files.
8. Use .jsx extension for test files (e.g., '/src/App.test.jsx').
`

	sysPrompt = fmt.Sprintf(sysPrompt, req.Attempt)

	codeContext := "APPLICATION CODE:\n"
	for name, content := range req.CodeFiles {
		codeContext += fmt.Sprintf("File: %s\n```\n%s\n```\n", name, content)
	}

	prevTestsContext := "\nPREVIOUS FAILING TESTS:\n"
	for name, content := range req.TestFiles {
		prevTestsContext += fmt.Sprintf("File: %s\n```\n%s\n```\n", name, content)
	}

	errorContext := fmt.Sprintf("\nTEST ERROR OUTPUT:\n```\n%s\n```\n", req.TestOutput)

	prompt := fmt.Sprintf("%s%s%s\n\nSPECIFICATION:\n%s\n\nFix the tests and return ONLY valid JSON.", codeContext, prevTestsContext, errorContext, req.Spec)

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
		// Enforce /src directory for tests
		filename := k
		if idx := len(filename) - 1; idx >= 0 {
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
