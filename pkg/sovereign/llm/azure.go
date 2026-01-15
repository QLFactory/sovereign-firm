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

type AzureOpenAIClient struct {
	Endpoint       string
	APIKey         string
	DeploymentName string
	APIVersion     string
	httpClient     *http.Client
}

func NewAzureOpenAIClient() *AzureOpenAIClient {
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://myazurellm.openai.azure.com/"
	}

	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")

	deploymentName := os.Getenv("AZURE_OPENAI_DEPLOYMENT")
	if deploymentName == "" {
		deploymentName = "gpt-4"
	}

	apiVersion := os.Getenv("AZURE_OPENAI_API_VERSION")
	if apiVersion == "" {
		apiVersion = "2024-02-15-preview"
	}

	return &AzureOpenAIClient{
		Endpoint:       endpoint,
		APIKey:         apiKey,
		DeploymentName: deploymentName,
		APIVersion:     apiVersion,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type azureChatRequest struct {
	Messages       []azureChatMessage `json:"messages"`
	Temperature    float64            `json:"temperature,omitempty"`
	ResponseFormat *responseFormat    `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type azureChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type azureChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func (c *AzureOpenAIClient) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	url := fmt.Sprintf("%sopenai/deployments/%s/chat/completions?api-version=%s",
		c.Endpoint, c.DeploymentName, c.APIVersion)

	messages := []azureChatMessage{
		{Role: "system", Content: req.System},
		{Role: "user", Content: req.Prompt},
	}

	chatReq := azureChatRequest{
		Messages:    messages,
		Temperature: 0.7,
	}

	if req.Format == "json" {
		chatReq.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-key", c.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Azure OpenAI error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp azureChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("Azure OpenAI API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from Azure OpenAI")
	}

	return &GenerateResponse{
		Response: chatResp.Choices[0].Message.Content,
	}, nil
}

func (c *AzureOpenAIClient) Embed(ctx context.Context, text string) ([]float32, error) {
	embeddingDeployment := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")
	if embeddingDeployment == "" {
		embeddingDeployment = "text-embedding-ada-002"
	}

	url := fmt.Sprintf("%sopenai/deployments/%s/embeddings?api-version=%s",
		c.Endpoint, embeddingDeployment, c.APIVersion)

	reqBody := map[string]interface{}{
		"input": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-key", c.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Azure embedding error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var embResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &embResp); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return embResp.Data[0].Embedding, nil
}

func (c *AzureOpenAIClient) Ping(ctx context.Context) error {
	// Simple ping by making a minimal request
	_, err := c.Generate(ctx, GenerateRequest{
		System: "You are a helpful assistant.",
		Prompt: "Say 'ok'",
	})
	return err
}
