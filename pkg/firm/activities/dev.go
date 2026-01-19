package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/agent"
	"github.com/qlfactory/sovereign-firm/pkg/mcp"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/memory"
)

type DevAgent struct {
	pool      *agent.AgentPool
	llmClient llm.Client
	store     memory.Store
}

func NewDevAgent(pool *agent.AgentPool) *DevAgent {
	llmClient := llm.NewClient()
	return &DevAgent{
		pool:      pool,
		llmClient: llmClient,
		store:     memory.NewChromaClient(llmClient),
	}
}

type CodeBundle map[string]string

// GenerateCode takes the final spec and produces the file system
func (a *DevAgent) GenerateCode(ctx context.Context, input GenerateCodeInput) (CodeBundle, error) {
	spec := input.Spec
	projectID := input.ProjectID

	if projectID == "" {
		projectID = "default-project"
	}
	sysPrompt := `You are a Senior React Developer using Vite. 
Output ONLY valid JSON.
The JSON must be a map where keys are filenames (MUST start with "/src/", e.g., "/src/App.jsx", "/src/components/Header.jsx") and values are the code content.
Do not include markdown backticks.
Use .jsx extension for React components (NOT .js).
Use standard CSS or inline styles. Do NOT use Tailwind CSS.
Build a modern React app using React 18+ standards.
Use 'react-router-dom' for navigation and 'framer-motion' for rich, premium animations.
DO NOT generate package.json - we will provide that.

COLLABORATION: If the requirements are ambiguous or you need clarification on the product vision, you can use the 'get_agents' tool to find the 'product-manager' and 'message_agent' to ask them questions. They are here to help you build the best possible MVP.

IMPORTANT: If you generate '/src/main.jsx', IT MUST BE EXACTLY:
import React from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
const root = createRoot(document.getElementById('root'));
root.render(<BrowserRouter><App /></BrowserRouter>);
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
	// Spawn or get agent for this project
	agentInstance, err := a.pool.SpawnAgent(ctx, "Dev", "senior-react-dev", projectID, "gpt-4", []string{"react-developer"})
	if err != nil {
		return nil, fmt.Errorf("failed to spawn agent: %w", err)
	}

	// Setup tool registry with collaboration tools
	registry := mcp.NewToolRegistry()
	mcp.RegisterBuiltinToolsWithConfig(registry, ".", &mcp.ToolConfig{
		ProjectID: projectID,
		AgentID:   agentInstance.ID,
		Messaging: a.pool,
	})

	// Execute via AgentExecutor
	executor := agent.NewAgentExecutor(agentInstance, registry, a.llmClient)
	// ISS-023: Sanitize user input to prevent prompt injection
	task := &agent.Task{
		Description: sysPrompt + "\n\n" + SanitizeUserInput("user-requirements", spec),
		Type:        "code_generate",
	}

	result, err := executor.ExecuteTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("Dev task failed: %w", err)
	}

	respContent := fmt.Sprintf("%v", result.Output)
	fmt.Printf("DEBUG: DevAgent received response: %s\n", respContent)

	// Robust JSON Extraction
	var rawBundle map[string]interface{}
	if err := a.extractAndUnmarshalJSON(respContent, &rawBundle); err != nil {
		return nil, fmt.Errorf("failed to parse code bundle: %w. Response: %s", err, respContent)
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
import { BrowserRouter } from 'react-router-dom';
import App from './App';

const root = createRoot(document.getElementById('root'));
root.render(<BrowserRouter><App /></BrowserRouter>);`

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
			"react":                 "^18.2.0",
			"react-dom":             "^18.2.0",
			"react-router-dom":      "^6.22.0",
			"framer-motion":         "^11.0.8",
			"lucide-react":          "^0.344.0",
			"react-query":           "^3.39.3",
			"@tanstack/react-query": "^5.28.4",
			"clsx":                  "^2.1.0",
			"tailwind-merge":        "^2.2.1",
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

// RefineCode modifies existing code based on feedback
func (a *DevAgent) RefineCode(ctx context.Context, input RefineCodeInput) (CodeBundle, error) {
	req := input

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

	// ISS-023: Sanitize user input to prevent prompt injection
	var prompt string
	if req.ValidationFeedback != "" {
		// Self-correction mode: focus on fixing validation errors (feedback is system-generated, not user input)
		prompt = fmt.Sprintf("%s\n\nVALIDATION ERRORS TO FIX:\n%s\n\nFix these validation errors and generate the JSON of corrected files.", codeContext, req.ValidationFeedback)
	} else {
		prompt = fmt.Sprintf("%s\n\n%s\n\nGenerate the JSON of modified files now.",
			codeContext, SanitizeUserInput("user-feedback", req.ChatHistory))
	}

	// RAG (Refinement patterns?) - skipped for now to save context, or we could look up "how to change colors".

	// Spawn or get agent for this project
	agentInstance, err := a.pool.SpawnAgent(ctx, "Dev", "senior-react-dev", "project-1", "gpt-4", []string{"react-developer"})
	if err != nil {
		return nil, fmt.Errorf("failed to spawn agent: %w", err)
	}

	// Setup tool registry
	registry := mcp.NewToolRegistry()
	mcp.RegisterBuiltinToolsWithConfig(registry, ".", &mcp.ToolConfig{
		ProjectID: "project-1",
		AgentID:   agentInstance.ID,
		Messaging: a.pool,
	})

	// Execute via AgentExecutor
	executor := agent.NewAgentExecutor(agentInstance, registry, a.llmClient)
	task := &agent.Task{
		Description: sysPrompt + "\n\nInstructions:\n" + prompt,
		Type:        "code_refine",
	}

	result, err := executor.ExecuteTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("Refine task failed: %w", err)
	}

	respContent := fmt.Sprintf("%v", result.Output)

	var rawChanges map[string]interface{}
	if err := a.extractAndUnmarshalJSON(respContent, &rawChanges); err != nil {
		return nil, fmt.Errorf("failed to parse changes: %w. Response: %s", err, respContent)
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
