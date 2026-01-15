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
	llmClient := llm.NewOllamaClient()
	return &DevAgent{
		llmClient: llmClient,
		store:     memory.NewChromaClient(llmClient),
	}
}

type CodeBundle map[string]string

// GenerateCode takes the final spec and produces the file system
func (a *DevAgent) GenerateCode(ctx context.Context, spec string) (CodeBundle, error) {
	sysPrompt := `You are a Senior React Developer. 
Output ONLY valid JSON.
The JSON must be a map where keys are filenames (e.g., "/App.js", "/components/Header.js") and values are the code content.
Do not include markdown backticks.
Build a modern React app using React 18+ standards (functional components, hooks).
IMPORTANT: If you generate 'index.js', you MUST use 'import { createRoot } from "react-dom/client"'.
Use Tailwind CSS for styling.
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

	var bundle CodeBundle
	if err := json.Unmarshal([]byte(resp.Response), &bundle); err != nil {
		// Fallback: Try to find JSON in markdown if model ignored instruction
		return nil, fmt.Errorf("failed to parse code bundle: %w. Response: %s", err, resp.Response)
	}

	return bundle, nil
}

// RefineInput struct to deserialized the map interface{}
type RefineInput struct {
	CurrentCode map[string]string `json:"current_code"`
	ChatHistory string            `json:"chat_history"`
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
`
	// Compress code for context (in a real app we'd be smarter due to context window)
	// For now, dump it all in.
	codeContext := "EXISTING CODE:\n"
	for name, content := range req.CurrentCode {
		codeContext += fmt.Sprintf("File: %s\n```\n%s\n```\n", name, content)
	}

	prompt := fmt.Sprintf("%s\n\nFEEDBACK/INSTRUCTIONS:\n%s\n\nGenerate the JSON of modified files now.", codeContext, req.ChatHistory)

	// RAG (Refinement patterns?) - skipped for now to save context, or we could look up "how to change colors".

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, err
	}

	var changes CodeBundle
	if err := json.Unmarshal([]byte(resp.Response), &changes); err != nil {
		return nil, fmt.Errorf("failed to parse refinement: %w", err)
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

	return finalBundle, nil
}
