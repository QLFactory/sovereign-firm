package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/memory"
)

type PMAgent struct {
	llmClient llm.Client
	store     memory.Store
}

func NewPMAgent() *PMAgent {
	llmClient := llm.NewClient()
	return &PMAgent{
		llmClient: llmClient,
		store:     memory.NewChromaClient(llmClient),
	}
}

// Chat handles the conversation logic
func (a *PMAgent) Chat(ctx context.Context, history string) (string, error) {
	sysPrompt := "You are a pragmatic Product Manager. Ask clarifying questions to define the MVP. Be concise."

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

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: history,
		System: sysPrompt,
	})
	if err != nil {
		return "", fmt.Errorf("PM brain failed: %w", err)
	}

	return resp.Response, nil
}
