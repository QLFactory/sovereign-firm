package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/memory"
)

type DevAgent struct {
	llmClient llm.Client
	store     memory.Store
}

func NewDevAgent() *DevAgent {
	llmClient := llm.NewClient()
	return &DevAgent{
		llmClient: llmClient,
		store:     memory.NewChromaClient(llmClient),
	}
}

type CodeBundle map[string]string

// GenerateCode takes the final spec and produces the file system
func (a *DevAgent) GenerateCode(ctx context.Context, spec string) (CodeBundle, error) {
	sysPrompt := `You are a Senior React Developer using Vite. 
Output ONLY valid JSON.
The JSON must be a map where keys are filenames (MUST start with "/src/", e.g., "/src/App.jsx", "/src/components/Header.jsx") and values are the code content.
Do not include markdown backticks.
Use .jsx extension for React components (NOT .js).
Use standard CSS or inline styles. Do NOT use Tailwind CSS.
Build a modern React app using React 18+ standards.
DO NOT generate package.json - we will provide that.
IMPORTANT: If you generate '/src/main.jsx', IT MUST BE EXACTLY:
import React from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';
const root = createRoot(document.getElementById('root'));
root.render(<App />);
`
	// RAG (with timeout)
	ragCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	embedding, err := a.llmClient.Embed(ragCtx, spec)
	if err == nil {
		docs, err := a.store.SimilaritySearch(ctx, embedding, 5) // Retrieve top 5 code patterns
		if err == nil && len(docs) > 0 {
			sysPrompt += "\n\nREFERENCE CODE PATTERNS:\n"
			for _, d := range docs {
				sysPrompt += fmt.Sprintf("File matched: %s\n```\n%s\n```\n", d.Metadata["source"], d.Content)
			}
		}
	}

	// Simple Prompt Engineering for JSON
	// In production, we'd use a constrained grammar or more robust parsing.
	prompt := fmt.Sprintf("Requirements:\n%s\n\nGenerate the JSON filesystem now.", spec)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json", // Ollama supports "format": "json" natively
	})
	if err != nil {
		return nil, fmt.Errorf("Dev brain failed: %w", err)
	}
	fmt.Printf("DEBUG: DevAgent V3-ROBUST received response: %s\n", resp.Response)

	// Robust JSON Extraction
	var rawBundle map[string]interface{}
	if err := a.extractAndUnmarshalJSON(resp.Response, &rawBundle); err != nil {
		return nil, fmt.Errorf("failed to parse code bundle: %w. Response: %s", err, resp.Response)
	}

	bundle := make(CodeBundle)
	for k, v := range rawBundle {
		switch val := v.(type) {
		case string:
			bundle[k] = val
		default:
			// If it's an object (like a nested package.json), marshal it back to a string
			marshaled, _ := json.MarshalIndent(val, "", "  ")
			bundle[k] = string(marshaled)
		}
	}

	return a.postProcessBundle(bundle), nil
}

// extractAndUnmarshalJSON attempts to find JSON within a string (handling markdown blocks)
func (a *DevAgent) extractAndUnmarshalJSON(input string, target interface{}) error {
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

// postProcessBundle ensures required boilerplate and dependencies exist
func (a *DevAgent) postProcessBundle(bundle CodeBundle) CodeBundle {
	// 0. Normalize ALL paths to /src/ (except package.json)
	normalized := make(CodeBundle)
	for path, content := range bundle {
		if path == "/package.json" {
			normalized[path] = content
			continue
		}

		// Ensure file is in /src/
		newPath := path
		if len(path) < 5 || path[:5] != "/src/" {
			// Extract just the filename (last segment after /)
			filename := path
			lastSlash := -1
			for i := len(path) - 1; i >= 0; i-- {
				if path[i] == '/' {
					lastSlash = i
					break
				}
			}
			if lastSlash != -1 {
				filename = path[lastSlash+1:]
			} else if path[0] == '/' {
				filename = path[1:]
			}
			newPath = "/src/" + filename
		}
		normalized[newPath] = content
	}
	bundle = normalized

	// 1. Force valid main.jsx in /src (Vite entry point)
	bundle["/src/main.jsx"] = `import React from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';

const root = createRoot(document.getElementById('root'));
root.render(<App />);`

	// 2. Ensure we have an index.html for Vite
	bundle["/index.html"] = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Sovereign Preview</title>
</head>
<body>
  <div id="root"></div>
  <script type="module" src="/src/main.jsx"></script>
</body>
</html>`

	// 3. Ensure vite.config.js exists
	bundle["/vite.config.js"] = `import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
});`

	// 4. Generate Vite-compatible package.json
	pkg := map[string]interface{}{
		"name": "sovereign-preview",
		"type": "module",
		"scripts": map[string]string{
			"dev":   "vite",
			"build": "vite build",
			"test":  "vitest run",
		},
		"dependencies": map[string]string{
			"react":            "^18.2.0",
			"react-dom":        "^18.2.0",
			"react-router-dom": "^6.22.0",
		},
		"devDependencies": map[string]string{
			"vite":                   "^5.0.0",
			"@vitejs/plugin-react":   "^4.2.0",
			"vitest":                 "^1.0.0",
			"@testing-library/react": "^14.0.0",
			"jsdom":                  "^23.0.0",
		},
	}

	newPkg, _ := json.MarshalIndent(pkg, "", "  ")
	bundle["/package.json"] = string(newPkg)

	return bundle
}

// RefineInput struct to deserialized the map interface{}
type RefineInput struct {
	CurrentCode        map[string]string `json:"current_code"`
	ChatHistory        string            `json:"chat_history"`
	ValidationFeedback string            `json:"validation_feedback,omitempty"`
}

// RefineCode modifies existing code based on feedback
func (a *DevAgent) RefineCode(ctx context.Context, input map[string]interface{}) (CodeBundle, error) {
	// Manual unmarshal since Temporal passes map[string]interface{} for dynamic inputs usually,
	// or we can just cast if it was preserved. Safe way is to marshal/unmarshal.
	inputBytes, _ := json.Marshal(input)
	var req RefineInput
	json.Unmarshal(inputBytes, &req)

	sysPrompt := `You are a Senior React Developer being asked to modify an existing application.
Output ONLY valid JSON of the *modified* files.
You can return a partial list of files - only those that changed.
Keys are filenames, values are new content.
IMPORTANT: Fix any validation errors first before implementing new features.
`
	// Compress code for context (in a real app we'd be smarter due to context window)
	// For now, dump it all in.
	codeContext := "EXISTING CODE:\n"
	for name, content := range req.CurrentCode {
		codeContext += fmt.Sprintf("File: %s\n```\n%s\n```\n", name, content)
	}

	var prompt string
	if req.ValidationFeedback != "" {
		// Self-correction mode: focus on fixing validation errors
		prompt = fmt.Sprintf("%s\n\nVALIDATION ERRORS TO FIX:\n%s\n\nFix these validation errors and generate the JSON of corrected files.", codeContext, req.ValidationFeedback)
	} else {
		prompt = fmt.Sprintf("%s\n\nFEEDBACK/INSTRUCTIONS:\n%s\n\nGenerate the JSON of modified files now.", codeContext, req.ChatHistory)
	}

	// RAG (Refinement patterns?) - skipped for now to save context, or we could look up "how to change colors".

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, err
	}

	var rawChanges map[string]interface{}
	if err := a.extractAndUnmarshalJSON(resp.Response, &rawChanges); err != nil {
		return nil, fmt.Errorf("failed to parse changes: %w", err)
	}

	changes := make(CodeBundle)
	for k, v := range rawChanges {
		switch val := v.(type) {
		case string:
			changes[k] = val
		default:
			marshaled, _ := json.MarshalIndent(val, "", "  ")
			changes[k] = string(marshaled)
		}
	}

	// Merge changes into current code
	// Actually, the new state will just be the new file set.
	// The Activity "RefineCode" should strictly return the *Full New Bundle* or the Workflow orchestrates the merge.
	// Workflow in project.go just says `state.CodeFiles = codeBundle`.
	// So we need to return the FULL bundle.

	finalBundle := make(CodeBundle)
	for k, v := range req.CurrentCode {
		finalBundle[k] = v
	}
	for k, v := range changes {
		finalBundle[k] = v
	}

	return a.postProcessBundle(finalBundle), nil
}
