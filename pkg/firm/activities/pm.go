package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/agent"
	"github.com/qlfactory/sovereign-firm/pkg/mcp"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/memory"
)

type PMAgent struct {
	pool      *agent.AgentPool
	llmClient llm.Client
	store     memory.Store
}

func NewPMAgent(pool *agent.AgentPool) *PMAgent {
	llmClient := llm.NewClient()
	return &PMAgent{
		pool:      pool,
		llmClient: llmClient,
		store:     memory.NewChromaClient(llmClient),
	}
}

// Chat handles the conversation logic
func (a *PMAgent) Chat(ctx context.Context, input ChatInput) (string, error) {
	sysPrompt := "You are a pragmatic Product Manager. Ask clarifying questions to define the MVP. Be concise. " +
		"You are also available to answer technical or product-related questions from the development team. " +
		"If you receive a message from another agent (e.g., 'Dev'), answer it promptly and clearly."
	history := input.ChatHistory
	projectID := input.ProjectID

	if projectID == "" {
		projectID = "default-project"
	}

	// RAG: Retrieve relevant context (with timeout)
	ragCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	embedding, err := a.llmClient.Embed(ragCtx, history)
	if err == nil {
		docs, err := a.store.SimilaritySearch(ctx, embedding, 3)
		if err == nil && len(docs) > 0 {
			sysPrompt += "\n\nRELEVANT CONTEXT FROM MEMORY:\n"
			for _, d := range docs {
				sysPrompt += fmt.Sprintf("- %s\n", d.Content)
			}
		}
	}

	// Spawn or get agent for this project
	agentInstance, err := a.pool.SpawnAgent(ctx, "PM", "product-manager", projectID, "gpt-4", []string{"project-manager"})
	if err != nil {
		return "", fmt.Errorf("failed to spawn agent: %w", err)
	}

	// Setup tool registry
	registry := mcp.NewToolRegistry()
	mcp.RegisterBuiltinToolsWithConfig(registry, ".", &mcp.ToolConfig{
		ProjectID: projectID,
		AgentID:   agentInstance.ID,
		Messaging: a.pool,
	})

	// Execute via AgentExecutor
	executor := agent.NewAgentExecutor(agentInstance, registry, a.llmClient)
	task := &agent.Task{
		Description: sysPrompt + "\n\nCHAT HISTORY:\n" + history,
		Type:        "chat",
	}

	result, err := executor.ExecuteTask(ctx, task)
	if err != nil {
		return "", fmt.Errorf("PM task failed: %w", err)
	}

	return fmt.Sprintf("%v", result.Output), nil
}
