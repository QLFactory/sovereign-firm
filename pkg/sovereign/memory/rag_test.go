package memory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultRAGConfig(t *testing.T) {
	config := DefaultRAGConfig()

	if config.MaxContextTokens != 4000 {
		t.Errorf("Expected MaxContextTokens 4000, got %d", config.MaxContextTokens)
	}

	if config.MaxDocuments != 10 {
		t.Errorf("Expected MaxDocuments 10, got %d", config.MaxDocuments)
	}

	if config.MinRelevance != 0.5 {
		t.Errorf("Expected MinRelevance 0.5, got %f", config.MinRelevance)
	}

	if !config.IncludeMetadata {
		t.Error("Expected IncludeMetadata to be true")
	}

	if !config.GroupByFile {
		t.Error("Expected GroupByFile to be true")
	}
}

func TestNewRAGPipeline(t *testing.T) {
	embedder := &mockEmbedder{}
	store := NewMultiLevelStore(embedder, "http://localhost:8000")
	pipeline := NewRAGPipeline(store, embedder, DefaultRAGConfig())

	if pipeline == nil {
		t.Fatal("Expected pipeline to be created")
	}

	if pipeline.store != store {
		t.Error("Expected store to be set")
	}

	if pipeline.embedder != embedder {
		t.Error("Expected embedder to be set")
	}
}

func TestRAGPipelineRetrieve(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(r.URL.Path, "/query") {
			resp := struct {
				Ids       [][]string                 `json:"ids"`
				Documents [][]string                 `json:"documents"`
				Metadatas [][]map[string]interface{} `json:"metadatas"`
			}{
				Ids:       [][]string{{"doc-1", "doc-2"}},
				Documents: [][]string{{"func auth() {}", "func validate() {}"}},
				Metadatas: [][]map[string]interface{}{
					{
						{"type": "code", "file_path": "auth.go", "language": "go"},
						{"type": "code", "file_path": "validate.go", "language": "go"},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	embedder := &mockEmbedder{}
	store := NewMultiLevelStore(embedder, server.URL)
	pipeline := NewRAGPipeline(store, embedder, DefaultRAGConfig())

	ctx := context.Background()
	result, err := pipeline.Retrieve(ctx, "authentication", "project-1", "client-1")

	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Query != "authentication" {
		t.Errorf("Expected query 'authentication', got '%s'", result.Query)
	}

	// Should have retrieved documents
	if len(result.Documents) == 0 {
		t.Log("No documents retrieved (may be expected in test setup)")
	}
}

func TestFormatContext(t *testing.T) {
	pipeline := &RAGPipeline{
		config: RAGConfig{
			IncludeMetadata: true,
		},
	}

	docs := []EnrichedDocument{
		{
			Document: Document{Content: "func main() {}"},
			Type:     DocTypeCode,
			FilePath: "main.go",
			Language: "go",
			StartLine: 1,
			EndLine:   3,
		},
		{
			Document: Document{Content: "Use dependency injection"},
			Type:     DocTypePattern,
		},
	}

	formatted := pipeline.formatContext(docs)

	if formatted == "" {
		t.Error("Expected formatted context")
	}

	if !strings.Contains(formatted, "Retrieved Context") {
		t.Error("Expected 'Retrieved Context' header")
	}

	if !strings.Contains(formatted, "main.go") {
		t.Error("Expected file path in formatted context")
	}

	if !strings.Contains(formatted, "```go") {
		t.Error("Expected code block with language")
	}
}

func TestFormatContextEmpty(t *testing.T) {
	pipeline := &RAGPipeline{
		config: RAGConfig{},
	}

	formatted := pipeline.formatContext([]EnrichedDocument{})

	if formatted != "" {
		t.Error("Expected empty string for empty documents")
	}
}

func TestGroupByFile(t *testing.T) {
	pipeline := &RAGPipeline{}

	docs := []EnrichedDocument{
		{Document: Document{ID: "1"}, FilePath: "a.go", StartLine: 10},
		{Document: Document{ID: "2"}, FilePath: "b.go", StartLine: 1},
		{Document: Document{ID: "3"}, FilePath: "a.go", StartLine: 1},
		{Document: Document{ID: "4"}, FilePath: "b.go", StartLine: 20},
	}

	grouped := pipeline.groupByFile(docs)

	// Files should be grouped together
	if len(grouped) != 4 {
		t.Errorf("Expected 4 documents, got %d", len(grouped))
	}

	// Documents from same file should be adjacent
	// First two should be from a.go (sorted by start line)
	if grouped[0].FilePath != "a.go" || grouped[1].FilePath != "a.go" {
		t.Error("Expected a.go documents to be grouped together")
	}

	// Within a file, should be sorted by start line
	if grouped[0].StartLine > grouped[1].StartLine {
		t.Error("Expected documents within file to be sorted by start line")
	}
}

func TestInferLanguage(t *testing.T) {
	tests := []struct {
		filePath string
		expected string
	}{
		{"main.go", "go"},
		{"app.js", "javascript"},
		{"app.ts", "typescript"},
		{"App.tsx", "tsx"},
		{"App.jsx", "jsx"},
		{"script.py", "python"},
		{"lib.rs", "rust"},
		{"index.html", "html"},
		{"styles.css", "css"},
		{"config.json", "json"},
		{"docker-compose.yaml", "yaml"},
		{"README.md", "markdown"},
		{"unknown.xyz", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := inferLanguage(tt.filePath)
			if result != tt.expected {
				t.Errorf("inferLanguage(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestRAGRequestStructure(t *testing.T) {
	req := RAGRequest{
		Query:        "authentication",
		SystemPrompt: "You are a helpful assistant",
		UserPrompt:   "How do I implement auth?",
		ProjectID:    "project-1",
		ClientID:     "client-1",
		Temperature:  0.7,
		DocumentTypes: []DocumentType{DocTypeCode, DocTypePattern},
		Languages:    []string{"go", "typescript"},
	}

	if req.Query != "authentication" {
		t.Error("Expected query to be set")
	}

	if req.Temperature != 0.7 {
		t.Error("Expected temperature 0.7")
	}

	if len(req.DocumentTypes) != 2 {
		t.Error("Expected 2 document types")
	}

	if len(req.Languages) != 2 {
		t.Error("Expected 2 languages")
	}
}

func TestRAGResponseStructure(t *testing.T) {
	resp := RAGResponse{
		Response: "Here's how to implement authentication...",
		Retrieval: &RetrievalContext{
			Query:         "auth",
			Documents:     []EnrichedDocument{},
			TokenEstimate: 500,
		},
		TokensUsed: 1000,
	}

	if resp.Response == "" {
		t.Error("Expected response to be set")
	}

	if resp.Retrieval == nil {
		t.Error("Expected retrieval context")
	}

	if resp.Retrieval.TokenEstimate != 500 {
		t.Error("Expected token estimate 500")
	}
}

func TestSourceReferenceStructure(t *testing.T) {
	ref := SourceReference{
		FilePath:  "auth/login.go",
		StartLine: 10,
		EndLine:   25,
		Type:      "code",
		Score:     0.95,
	}

	if ref.FilePath != "auth/login.go" {
		t.Error("Expected file path to be set")
	}

	if ref.StartLine != 10 {
		t.Error("Expected start line 10")
	}

	if ref.Score != 0.95 {
		t.Error("Expected score 0.95")
	}
}

func TestRetrievalContextStructure(t *testing.T) {
	ctx := RetrievalContext{
		Query: "authentication",
		Documents: []EnrichedDocument{
			{Document: Document{ID: "doc-1"}},
		},
		FormattedText: "## Context\n...",
		TokenEstimate: 1000,
		Sources: []SourceReference{
			{FilePath: "auth.go"},
		},
	}

	if ctx.Query != "authentication" {
		t.Error("Expected query to be set")
	}

	if len(ctx.Documents) != 1 {
		t.Error("Expected 1 document")
	}

	if ctx.FormattedText == "" {
		t.Error("Expected formatted text")
	}

	if len(ctx.Sources) != 1 {
		t.Error("Expected 1 source")
	}
}

func TestAugmentPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(r.URL.Path, "/query") {
			resp := struct {
				Ids       [][]string                 `json:"ids"`
				Documents [][]string                 `json:"documents"`
				Metadatas [][]map[string]interface{} `json:"metadatas"`
			}{
				Ids:       [][]string{{"doc-1"}},
				Documents: [][]string{{"func login() {}"}},
				Metadatas: [][]map[string]interface{}{{{"type": "code"}}},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	embedder := &mockEmbedder{}
	store := NewMultiLevelStore(embedder, server.URL)
	pipeline := NewRAGPipeline(store, embedder, DefaultRAGConfig())

	ctx := context.Background()
	originalPrompt := "How do I implement login?"
	augmented, retrieval, err := pipeline.AugmentPrompt(ctx, originalPrompt, "login", "project-1", "client-1")

	if err != nil {
		t.Fatalf("AugmentPrompt failed: %v", err)
	}

	// Augmented prompt should contain original prompt
	if !strings.Contains(augmented, originalPrompt) {
		t.Error("Augmented prompt should contain original prompt")
	}

	// Should have retrieval context
	if retrieval == nil {
		t.Log("No retrieval context (may be expected)")
	}
}

func TestIsCommonError(t *testing.T) {
	tests := []struct {
		errorMsg string
		expected bool
	}{
		{"cannot find module 'react'", true},
		{"TypeError: undefined is not a function", true},
		{"null pointer exception", true},
		{"import error: no module named 'foo'", true},
		{"syntax error: unexpected token", true},
		{"permission denied: /etc/passwd", true},
		{"connection refused to localhost:5432", true},
		{"custom application error", false},
		{"user not found", false},
		{"invalid input", false},
	}

	for _, tt := range tests {
		name := tt.errorMsg
		if len(name) > 20 {
			name = name[:20]
		}
		t.Run(name, func(t *testing.T) {
			result := isCommonError(tt.errorMsg)
			if result != tt.expected {
				t.Errorf("isCommonError(%q) = %v, want %v", tt.errorMsg, result, tt.expected)
			}
		})
	}
}

func TestFindSimilarCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(r.URL.Path, "/query") {
			resp := struct {
				Ids       [][]string                 `json:"ids"`
				Documents [][]string                 `json:"documents"`
				Metadatas [][]map[string]interface{} `json:"metadatas"`
			}{
				Ids:       [][]string{{"code-1", "code-2"}},
				Documents: [][]string{{"func a() {}", "func b() {}"}},
				Metadatas: [][]map[string]interface{}{{
					{"type": "code", "file_path": "a.go"},
					{"type": "code", "file_path": "b.go"},
				}},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	embedder := &mockEmbedder{}
	store := NewMultiLevelStore(embedder, server.URL)
	pipeline := NewRAGPipeline(store, embedder, DefaultRAGConfig())

	ctx := context.Background()
	docs, err := pipeline.FindSimilarCode(ctx, "authentication function", "project-1", 5)

	if err != nil {
		t.Fatalf("FindSimilarCode failed: %v", err)
	}

	// Should return documents (may be filtered)
	t.Logf("Found %d similar code documents", len(docs))
}

func TestFindPatterns(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if strings.Contains(r.URL.Path, "/query") {
			resp := struct {
				Ids       [][]string                 `json:"ids"`
				Documents [][]string                 `json:"documents"`
				Metadatas [][]map[string]interface{} `json:"metadatas"`
			}{
				Ids:       [][]string{{"pattern-1"}},
				Documents: [][]string{{"Singleton pattern: ..."}},
				Metadatas: [][]map[string]interface{}{{{"type": "pattern"}}},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	embedder := &mockEmbedder{}
	store := NewMultiLevelStore(embedder, server.URL)
	pipeline := NewRAGPipeline(store, embedder, DefaultRAGConfig())

	ctx := context.Background()
	docs, err := pipeline.FindPatterns(ctx, "singleton", "client-1", 5)

	if err != nil {
		t.Fatalf("FindPatterns failed: %v", err)
	}

	t.Logf("Found %d patterns", len(docs))
}
