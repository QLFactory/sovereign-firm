package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGenerateRequestStructure(t *testing.T) {
	req := GenerateRequest{
		Model:       "gpt-4",
		Prompt:      "Hello, world!",
		System:      "You are a helpful assistant.",
		Stream:      false,
		Temperature: 0.7,
		Format:      "json",
	}

	if req.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", req.Model)
	}

	if req.Prompt != "Hello, world!" {
		t.Errorf("Expected prompt 'Hello, world!', got '%s'", req.Prompt)
	}

	if req.System != "You are a helpful assistant." {
		t.Errorf("Expected system prompt, got '%s'", req.System)
	}

	if req.Stream {
		t.Error("Expected Stream to be false")
	}

	if req.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", req.Temperature)
	}

	if req.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", req.Format)
	}
}

func TestGenerateResponseStructure(t *testing.T) {
	resp := GenerateResponse{
		Response: "Hello! How can I help you today?",
		Done:     true,
	}

	if resp.Response != "Hello! How can I help you today?" {
		t.Errorf("Expected response message, got '%s'", resp.Response)
	}

	if !resp.Done {
		t.Error("Expected Done to be true")
	}
}

func TestGenerateRequestJSONMarshal(t *testing.T) {
	req := GenerateRequest{
		Model:       "llama3",
		Prompt:      "Test prompt",
		System:      "System prompt",
		Stream:      false,
		Temperature: 0.5,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal GenerateRequest: %v", err)
	}

	var unmarshaled GenerateRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal GenerateRequest: %v", err)
	}

	if unmarshaled.Model != req.Model {
		t.Errorf("Expected model '%s', got '%s'", req.Model, unmarshaled.Model)
	}

	if unmarshaled.Prompt != req.Prompt {
		t.Errorf("Expected prompt '%s', got '%s'", req.Prompt, unmarshaled.Prompt)
	}
}

func TestGenerateResponseJSONMarshal(t *testing.T) {
	resp := GenerateResponse{
		Response: "Test response",
		Done:     true,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal GenerateResponse: %v", err)
	}

	var unmarshaled GenerateResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal GenerateResponse: %v", err)
	}

	if unmarshaled.Response != resp.Response {
		t.Errorf("Expected response '%s', got '%s'", resp.Response, unmarshaled.Response)
	}

	if unmarshaled.Done != resp.Done {
		t.Errorf("Expected done %v, got %v", resp.Done, unmarshaled.Done)
	}
}

func TestNewClientDefaultsToOllama(t *testing.T) {
	// Clear any existing LLM_PROVIDER env var
	os.Unsetenv("LLM_PROVIDER")

	client := NewClient()

	_, ok := client.(*OllamaClient)
	if !ok {
		t.Error("Expected NewClient to return OllamaClient by default")
	}
}

func TestNewClientWithAzureProvider(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "azure")
	defer os.Unsetenv("LLM_PROVIDER")

	client := NewClient()

	_, ok := client.(*AzureOpenAIClient)
	if !ok {
		t.Error("Expected NewClient to return AzureOpenAIClient when LLM_PROVIDER=azure")
	}
}

func TestNewOllamaClientDefaults(t *testing.T) {
	// Clear env vars
	os.Unsetenv("OLLAMA_HOST")
	os.Unsetenv("OLLAMA_MODEL")
	os.Unsetenv("OLLAMA_EMBED_MODEL")

	client := NewOllamaClient()

	if client.BaseURL != "http://localhost:11434" {
		t.Errorf("Expected default BaseURL 'http://localhost:11434', got '%s'", client.BaseURL)
	}

	if client.Model != "llama3" {
		t.Errorf("Expected default Model 'llama3', got '%s'", client.Model)
	}

	if client.EmbedModel != "nomic-embed-text" {
		t.Errorf("Expected default EmbedModel 'nomic-embed-text', got '%s'", client.EmbedModel)
	}
}

func TestNewOllamaClientWithEnvVars(t *testing.T) {
	os.Setenv("OLLAMA_HOST", "http://custom:11434")
	os.Setenv("OLLAMA_MODEL", "codellama")
	os.Setenv("OLLAMA_EMBED_MODEL", "mxbai-embed-large")
	defer func() {
		os.Unsetenv("OLLAMA_HOST")
		os.Unsetenv("OLLAMA_MODEL")
		os.Unsetenv("OLLAMA_EMBED_MODEL")
	}()

	client := NewOllamaClient()

	if client.BaseURL != "http://custom:11434" {
		t.Errorf("Expected BaseURL 'http://custom:11434', got '%s'", client.BaseURL)
	}

	if client.Model != "codellama" {
		t.Errorf("Expected Model 'codellama', got '%s'", client.Model)
	}

	if client.EmbedModel != "mxbai-embed-large" {
		t.Errorf("Expected EmbedModel 'mxbai-embed-large', got '%s'", client.EmbedModel)
	}
}

func TestNewAzureOpenAIClientDefaults(t *testing.T) {
	// Clear env vars
	os.Unsetenv("AZURE_OPENAI_ENDPOINT")
	os.Unsetenv("AZURE_OPENAI_API_KEY")
	os.Unsetenv("AZURE_OPENAI_DEPLOYMENT")
	os.Unsetenv("AZURE_OPENAI_API_VERSION")

	client := NewAzureOpenAIClient()

	if client.Endpoint != "https://myazurellm.openai.azure.com/" {
		t.Errorf("Expected default Endpoint, got '%s'", client.Endpoint)
	}

	if client.DeploymentName != "gpt-4" {
		t.Errorf("Expected default DeploymentName 'gpt-4', got '%s'", client.DeploymentName)
	}

	if client.APIVersion != "2024-02-15-preview" {
		t.Errorf("Expected default APIVersion '2024-02-15-preview', got '%s'", client.APIVersion)
	}

	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}

func TestNewAzureOpenAIClientWithEnvVars(t *testing.T) {
	os.Setenv("AZURE_OPENAI_ENDPOINT", "https://custom.openai.azure.com/")
	os.Setenv("AZURE_OPENAI_API_KEY", "test-api-key")
	os.Setenv("AZURE_OPENAI_DEPLOYMENT", "gpt-4-turbo")
	os.Setenv("AZURE_OPENAI_API_VERSION", "2024-03-01")
	defer func() {
		os.Unsetenv("AZURE_OPENAI_ENDPOINT")
		os.Unsetenv("AZURE_OPENAI_API_KEY")
		os.Unsetenv("AZURE_OPENAI_DEPLOYMENT")
		os.Unsetenv("AZURE_OPENAI_API_VERSION")
	}()

	client := NewAzureOpenAIClient()

	if client.Endpoint != "https://custom.openai.azure.com/" {
		t.Errorf("Expected custom Endpoint, got '%s'", client.Endpoint)
	}

	if client.APIKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", client.APIKey)
	}

	if client.DeploymentName != "gpt-4-turbo" {
		t.Errorf("Expected deployment 'gpt-4-turbo', got '%s'", client.DeploymentName)
	}

	if client.APIVersion != "2024-03-01" {
		t.Errorf("Expected API version '2024-03-01', got '%s'", client.APIVersion)
	}
}

func TestOllamaClientGenerate(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("Expected path '/api/generate', got '%s'", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected method POST, got '%s'", r.Method)
		}

		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify stream is set to false
		if req.Stream {
			t.Error("Expected Stream to be false")
		}

		resp := GenerateResponse{
			Response: "This is a test response from the mock server.",
			Done:     true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &OllamaClient{
		BaseURL:    server.URL,
		Model:      "llama3",
		EmbedModel: "nomic-embed-text",
	}

	ctx := context.Background()
	resp, err := client.Generate(ctx, GenerateRequest{
		Prompt: "Test prompt",
	})

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Response != "This is a test response from the mock server." {
		t.Errorf("Unexpected response: '%s'", resp.Response)
	}

	if !resp.Done {
		t.Error("Expected Done to be true")
	}
}

func TestOllamaClientGenerateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	client := &OllamaClient{
		BaseURL: server.URL,
		Model:   "llama3",
	}

	ctx := context.Background()
	_, err := client.Generate(ctx, GenerateRequest{Prompt: "Test"})

	if err == nil {
		t.Error("Expected error for 500 response")
	}
}

func TestOllamaClientEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("Expected path '/api/embeddings', got '%s'", r.URL.Path)
		}

		resp := struct {
			Embedding []float64 `json:"embedding"`
		}{
			Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &OllamaClient{
		BaseURL:    server.URL,
		EmbedModel: "nomic-embed-text",
	}

	ctx := context.Background()
	embedding, err := client.Embed(ctx, "Test text")

	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(embedding) != 5 {
		t.Errorf("Expected 5 embedding values, got %d", len(embedding))
	}

	expected := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	for i, v := range embedding {
		if v != expected[i] {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, expected[i], v)
		}
	}
}

func TestOllamaClientEmbedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}))
	defer server.Close()

	client := &OllamaClient{
		BaseURL:    server.URL,
		EmbedModel: "nomic-embed-text",
	}

	ctx := context.Background()
	_, err := client.Embed(ctx, "Test")

	if err == nil {
		t.Error("Expected error for 400 response")
	}
}

func TestOllamaClientPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("Expected path '/api/tags', got '%s'", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("Expected method GET, got '%s'", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"models": []}`))
	}))
	defer server.Close()

	client := &OllamaClient{
		BaseURL: server.URL,
	}

	ctx := context.Background()
	err := client.Ping(ctx)

	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestOllamaClientPingError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := &OllamaClient{
		BaseURL: server.URL,
	}

	ctx := context.Background()
	err := client.Ping(ctx)

	if err == nil {
		t.Error("Expected error for 503 response")
	}
}

func TestAzureOpenAIClientGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the API key header
		if r.Header.Get("api-key") != "test-key" {
			t.Errorf("Expected api-key header 'test-key', got '%s'", r.Header.Get("api-key"))
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}

		var req azureChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if len(req.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(req.Messages))
		}

		resp := azureChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "Azure test response"}},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &AzureOpenAIClient{
		Endpoint:       server.URL + "/",
		APIKey:         "test-key",
		DeploymentName: "gpt-4",
		APIVersion:     "2024-02-15-preview",
		httpClient:     http.DefaultClient,
	}

	ctx := context.Background()
	resp, err := client.Generate(ctx, GenerateRequest{
		Prompt: "Test prompt",
		System: "System prompt",
	})

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Response != "Azure test response" {
		t.Errorf("Unexpected response: '%s'", resp.Response)
	}
}

func TestAzureOpenAIClientGenerateWithJSONFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req azureChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.ResponseFormat == nil {
			t.Error("Expected ResponseFormat to be set for JSON format")
		} else if req.ResponseFormat.Type != "json_object" {
			t.Errorf("Expected ResponseFormat.Type 'json_object', got '%s'", req.ResponseFormat.Type)
		}

		resp := azureChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: `{"result": "test"}`}},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &AzureOpenAIClient{
		Endpoint:       server.URL + "/",
		APIKey:         "test-key",
		DeploymentName: "gpt-4",
		APIVersion:     "2024-02-15-preview",
		httpClient:     http.DefaultClient,
	}

	ctx := context.Background()
	_, err := client.Generate(ctx, GenerateRequest{
		Prompt: "Test",
		Format: "json",
	})

	if err != nil {
		t.Fatalf("Generate with JSON format failed: %v", err)
	}
}

func TestAzureOpenAIClientGenerateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := &AzureOpenAIClient{
		Endpoint:       server.URL + "/",
		APIKey:         "invalid-key",
		DeploymentName: "gpt-4",
		APIVersion:     "2024-02-15-preview",
		httpClient:     http.DefaultClient,
	}

	ctx := context.Background()
	_, err := client.Generate(ctx, GenerateRequest{Prompt: "Test"})

	if err == nil {
		t.Error("Expected error for 401 response")
	}
}

func TestAzureOpenAIClientGenerateAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := azureChatResponse{
			Error: &struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			}{
				Message: "Rate limit exceeded",
				Code:    "429",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &AzureOpenAIClient{
		Endpoint:       server.URL + "/",
		APIKey:         "test-key",
		DeploymentName: "gpt-4",
		APIVersion:     "2024-02-15-preview",
		httpClient:     http.DefaultClient,
	}

	ctx := context.Background()
	_, err := client.Generate(ctx, GenerateRequest{Prompt: "Test"})

	if err == nil {
		t.Error("Expected error for API error response")
	}
}

func TestAzureOpenAIClientGenerateNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := azureChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &AzureOpenAIClient{
		Endpoint:       server.URL + "/",
		APIKey:         "test-key",
		DeploymentName: "gpt-4",
		APIVersion:     "2024-02-15-preview",
		httpClient:     http.DefaultClient,
	}

	ctx := context.Background()
	_, err := client.Generate(ctx, GenerateRequest{Prompt: "Test"})

	if err == nil {
		t.Error("Expected error for empty choices")
	}
}

func TestAzureOpenAIClientEmbed(t *testing.T) {
	os.Setenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT", "ada-002")
	defer os.Unsetenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Data []struct {
				Embedding []float32 `json:"embedding"`
			} `json:"data"`
		}{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
			}{
				{Embedding: []float32{0.1, 0.2, 0.3}},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &AzureOpenAIClient{
		Endpoint:       server.URL + "/",
		APIKey:         "test-key",
		DeploymentName: "gpt-4",
		APIVersion:     "2024-02-15-preview",
		httpClient:     http.DefaultClient,
	}

	ctx := context.Background()
	embedding, err := client.Embed(ctx, "Test text")

	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(embedding) != 3 {
		t.Errorf("Expected 3 embedding values, got %d", len(embedding))
	}
}

func TestAzureOpenAIClientEmbedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}))
	defer server.Close()

	client := &AzureOpenAIClient{
		Endpoint:       server.URL + "/",
		APIKey:         "test-key",
		DeploymentName: "gpt-4",
		APIVersion:     "2024-02-15-preview",
		httpClient:     http.DefaultClient,
	}

	ctx := context.Background()
	_, err := client.Embed(ctx, "Test")

	if err == nil {
		t.Error("Expected error for 400 response")
	}
}

func TestAzureOpenAIClientEmbedEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Data []struct {
				Embedding []float32 `json:"embedding"`
			} `json:"data"`
		}{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
			}{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &AzureOpenAIClient{
		Endpoint:       server.URL + "/",
		APIKey:         "test-key",
		DeploymentName: "gpt-4",
		APIVersion:     "2024-02-15-preview",
		httpClient:     http.DefaultClient,
	}

	ctx := context.Background()
	_, err := client.Embed(ctx, "Test")

	if err == nil {
		t.Error("Expected error for empty embedding response")
	}
}

func TestClientInterfaceCompliance(t *testing.T) {
	// Verify that both client types implement the Client interface
	var _ Client = &OllamaClient{}
	var _ Client = &AzureOpenAIClient{}
}

func TestOllamaClientGenerateUsesDefaultModel(t *testing.T) {
	var receivedModel string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GenerateRequest
		json.NewDecoder(r.Body).Decode(&req)
		receivedModel = req.Model

		resp := GenerateResponse{Response: "test", Done: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &OllamaClient{
		BaseURL: server.URL,
		Model:   "codellama",
	}

	ctx := context.Background()
	_, _ = client.Generate(ctx, GenerateRequest{Prompt: "test"})

	if receivedModel != "codellama" {
		t.Errorf("Expected default model 'codellama', got '%s'", receivedModel)
	}
}

func TestOllamaClientGenerateOverridesModel(t *testing.T) {
	var receivedModel string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GenerateRequest
		json.NewDecoder(r.Body).Decode(&req)
		receivedModel = req.Model

		resp := GenerateResponse{Response: "test", Done: true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &OllamaClient{
		BaseURL: server.URL,
		Model:   "codellama",
	}

	ctx := context.Background()
	_, _ = client.Generate(ctx, GenerateRequest{
		Model:  "mistral",
		Prompt: "test",
	})

	if receivedModel != "mistral" {
		t.Errorf("Expected overridden model 'mistral', got '%s'", receivedModel)
	}
}
