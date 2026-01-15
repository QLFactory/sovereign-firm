package llm

import "os"

// NewClient returns the appropriate LLM client based on LLM_PROVIDER env var
func NewClient() Client {
	provider := os.Getenv("LLM_PROVIDER")
	switch provider {
	case "azure":
		return NewAzureOpenAIClient()
	default:
		return NewOllamaClient()
	}
}
