package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type OllamaClient struct {
	BaseURL    string
	Model      string
	EmbedModel string
}

func NewOllamaClient() *OllamaClient {
	url := os.Getenv("OLLAMA_HOST")
	if url == "" {
		url = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3"
	}
	embedModel := os.Getenv("OLLAMA_EMBED_MODEL")
	if embedModel == "" {
		embedModel = "nomic-embed-text"
	}
	return &OllamaClient{
		BaseURL:    url,
		Model:      model,
		EmbedModel: embedModel,
	}
}

// Generate calls POST /api/generate
func (c *OllamaClient) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	if req.Model == "" {
		req.Model = c.Model
	}
	req.Stream = false // Enforce non-streaming for simple Request/Response

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// ENFORCE TIMEOUT
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &genResp, nil
}

// Embed calls POST /api/embeddings
func (c *OllamaClient) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model":  c.EmbedModel,
		"prompt": text,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// USE TIMEOUT
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	type EmbedResponse struct {
		Embedding []float64 `json:"embedding"`
	}
	var embedResp EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, err
	}

	// Convert float64 to float32
	result := make([]float32, len(embedResp.Embedding))
	for i, v := range embedResp.Embedding {
		result[i] = float32(v)
	}

	return result, nil
}

// Ping checks if the Ollama service is reachable
func (c *OllamaClient) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/api/tags", nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed with status: %d", resp.StatusCode)
	}
	return nil
}
