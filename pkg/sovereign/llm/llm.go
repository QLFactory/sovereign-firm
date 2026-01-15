package llm

import "context"

// GenerateRequest defines the input for an LLM generation
type GenerateRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	System      string  `json:"system,omitempty"`
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature,omitempty"`
	Format      string  `json:"format,omitempty"` // "json" or empty
}

// GenerateResponse defines the output from an LLM
type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Client is the sovereign interface for AI providers
type Client interface {
	// Generate returns a completion for the given request
	Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)

	// Embed returns vector embeddings (for RAG)
	Embed(ctx context.Context, text string) ([]float32, error)

	// Ping checks if the LLM service is available
	Ping(ctx context.Context) error
}
