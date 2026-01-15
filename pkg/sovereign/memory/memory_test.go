package memory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// mockEmbedder implements llm.Client for testing
type mockEmbedder struct {
	embedFunc func(ctx context.Context, text string) ([]float32, error)
}

func (m *mockEmbedder) Generate(ctx context.Context, req llm.GenerateRequest) (*llm.GenerateResponse, error) {
	return &llm.GenerateResponse{Response: "mock response", Done: true}, nil
}

func (m *mockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if m.embedFunc != nil {
		return m.embedFunc(ctx, text)
	}
	return []float32{0.1, 0.2, 0.3, 0.4, 0.5}, nil
}

func (m *mockEmbedder) Ping(ctx context.Context) error {
	return nil
}

func TestDocumentStructure(t *testing.T) {
	doc := Document{
		ID:      "doc-123",
		Content: "This is the document content.",
		Metadata: map[string]interface{}{
			"type":   "code",
			"lang":   "go",
			"file":   "main.go",
			"lineno": 42,
		},
		Vector: []float32{0.1, 0.2, 0.3},
	}

	if doc.ID != "doc-123" {
		t.Errorf("Expected ID 'doc-123', got '%s'", doc.ID)
	}

	if doc.Content != "This is the document content." {
		t.Errorf("Expected content to match, got '%s'", doc.Content)
	}

	if doc.Metadata["type"] != "code" {
		t.Errorf("Expected metadata type 'code', got '%v'", doc.Metadata["type"])
	}

	if doc.Metadata["lang"] != "go" {
		t.Errorf("Expected metadata lang 'go', got '%v'", doc.Metadata["lang"])
	}

	if len(doc.Vector) != 3 {
		t.Errorf("Expected 3 vector values, got %d", len(doc.Vector))
	}
}

func TestDocumentJSONMarshal(t *testing.T) {
	doc := Document{
		ID:      "doc-456",
		Content: "Test content",
		Metadata: map[string]interface{}{
			"key": "value",
		},
		Vector: []float32{0.5, 0.6},
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Failed to marshal Document: %v", err)
	}

	var unmarshaled Document
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Document: %v", err)
	}

	if unmarshaled.ID != doc.ID {
		t.Errorf("Expected ID '%s', got '%s'", doc.ID, unmarshaled.ID)
	}

	if unmarshaled.Content != doc.Content {
		t.Errorf("Expected Content '%s', got '%s'", doc.Content, unmarshaled.Content)
	}
}

func TestDocumentWithEmptyFields(t *testing.T) {
	doc := Document{}

	if doc.ID != "" {
		t.Errorf("Expected empty ID, got '%s'", doc.ID)
	}

	if doc.Content != "" {
		t.Errorf("Expected empty Content, got '%s'", doc.Content)
	}

	if doc.Metadata != nil {
		t.Error("Expected nil Metadata")
	}

	if doc.Vector != nil {
		t.Error("Expected nil Vector")
	}
}

func TestNewChromaClientDefaults(t *testing.T) {
	os.Unsetenv("CHROMA_HOST")
	os.Unsetenv("CHROMA_COLLECTION")

	embedder := &mockEmbedder{}
	client := NewChromaClient(embedder)

	if client.BaseURL != "http://localhost:8000" {
		t.Errorf("Expected default BaseURL 'http://localhost:8000', got '%s'", client.BaseURL)
	}

	if client.Collection != "sovereign-firm-memory-v2" {
		t.Errorf("Expected default Collection 'sovereign-firm-memory-v2', got '%s'", client.Collection)
	}

	if client.Embedder != embedder {
		t.Error("Expected Embedder to be set")
	}
}

func TestNewChromaClientWithEnvVars(t *testing.T) {
	os.Setenv("CHROMA_HOST", "http://custom-chroma:8080")
	os.Setenv("CHROMA_COLLECTION", "custom-collection")
	defer func() {
		os.Unsetenv("CHROMA_HOST")
		os.Unsetenv("CHROMA_COLLECTION")
	}()

	embedder := &mockEmbedder{}
	client := NewChromaClient(embedder)

	if client.BaseURL != "http://custom-chroma:8080" {
		t.Errorf("Expected BaseURL 'http://custom-chroma:8080', got '%s'", client.BaseURL)
	}

	if client.Collection != "custom-collection" {
		t.Errorf("Expected Collection 'custom-collection', got '%s'", client.Collection)
	}
}

func TestChromaClientAddDocuments(t *testing.T) {
	var collectionCreated bool
	var documentsAdded bool
	var receivedDocs struct {
		IDs        []string                   `json:"ids"`
		Embeddings [][]float32                `json:"embeddings"`
		Documents  []string                   `json:"documents"`
		Metadatas  []map[string]interface{}   `json:"metadatas"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			collectionCreated = true
			resp := map[string]string{"id": "collection-uuid-123"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.URL.Path == "/api/v1/collections/collection-uuid-123/add" {
			documentsAdded = true
			json.NewDecoder(r.Body).Decode(&receivedDocs)
			w.WriteHeader(http.StatusOK)
			return
		}

		t.Errorf("Unexpected path: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	embedder := &mockEmbedder{
		embedFunc: func(ctx context.Context, text string) ([]float32, error) {
			return []float32{0.1, 0.2, 0.3}, nil
		},
	}

	client := &ChromaClient{
		BaseURL:    server.URL,
		Collection: "test-collection",
		Embedder:   embedder,
	}

	ctx := context.Background()
	docs := []Document{
		{
			ID:       "doc-1",
			Content:  "First document content",
			Metadata: map[string]interface{}{"type": "test"},
		},
		{
			ID:       "doc-2",
			Content:  "Second document content",
			Metadata: map[string]interface{}{"type": "test"},
		},
	}

	err := client.AddDocuments(ctx, docs)

	if err != nil {
		t.Fatalf("AddDocuments failed: %v", err)
	}

	if !collectionCreated {
		t.Error("Expected collection to be created/fetched")
	}

	if !documentsAdded {
		t.Error("Expected documents to be added")
	}

	if len(receivedDocs.IDs) != 2 {
		t.Errorf("Expected 2 document IDs, got %d", len(receivedDocs.IDs))
	}
}

func TestChromaClientAddDocumentsWithPrecomputedVectors(t *testing.T) {
	var receivedEmbeddings [][]float32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.URL.Path == "/api/v1/collections/coll-id/add" {
			var payload struct {
				Embeddings [][]float32 `json:"embeddings"`
			}
			json.NewDecoder(r.Body).Decode(&payload)
			receivedEmbeddings = payload.Embeddings
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Embedder that should NOT be called since vectors are precomputed
	embedder := &mockEmbedder{
		embedFunc: func(ctx context.Context, text string) ([]float32, error) {
			t.Error("Embedder should not be called when vectors are precomputed")
			return nil, nil
		},
	}

	client := &ChromaClient{
		BaseURL:    server.URL,
		Collection: "test",
		Embedder:   embedder,
	}

	ctx := context.Background()
	docs := []Document{
		{
			ID:      "doc-precomputed",
			Content: "Content",
			Vector:  []float32{1.0, 2.0, 3.0},
		},
	}

	err := client.AddDocuments(ctx, docs)

	if err != nil {
		t.Fatalf("AddDocuments failed: %v", err)
	}

	if len(receivedEmbeddings) != 1 {
		t.Fatalf("Expected 1 embedding, got %d", len(receivedEmbeddings))
	}

	expected := []float32{1.0, 2.0, 3.0}
	for i, v := range receivedEmbeddings[0] {
		if v != expected[i] {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, expected[i], v)
		}
	}
}

func TestChromaClientSimilaritySearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.URL.Path == "/api/v1/collections/coll-id/query" {
			var req struct {
				QueryEmbeddings [][]float32 `json:"query_embeddings"`
				NResults        int         `json:"n_results"`
			}
			json.NewDecoder(r.Body).Decode(&req)

			if req.NResults != 5 {
				t.Errorf("Expected n_results=5, got %d", req.NResults)
			}

			resp := struct {
				Ids       [][]string                 `json:"ids"`
				Documents [][]string                 `json:"documents"`
				Metadatas [][]map[string]interface{} `json:"metadatas"`
			}{
				Ids:       [][]string{{"doc-1", "doc-2", "doc-3"}},
				Documents: [][]string{{"Content 1", "Content 2", "Content 3"}},
				Metadatas: [][]map[string]interface{}{
					{
						{"type": "test"},
						{"type": "test"},
						{"type": "test"},
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

	client := &ChromaClient{
		BaseURL:    server.URL,
		Collection: "test",
		Embedder:   &mockEmbedder{},
	}

	ctx := context.Background()
	queryVector := []float32{0.1, 0.2, 0.3}
	results, err := client.SimilaritySearch(ctx, queryVector, 5)

	if err != nil {
		t.Fatalf("SimilaritySearch failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if results[0].ID != "doc-1" {
		t.Errorf("Expected first result ID 'doc-1', got '%s'", results[0].ID)
	}

	if results[0].Content != "Content 1" {
		t.Errorf("Expected first result content 'Content 1', got '%s'", results[0].Content)
	}
}

func TestChromaClientSimilaritySearchEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.URL.Path == "/api/v1/collections/coll-id/query" {
			resp := struct {
				Ids       [][]string                 `json:"ids"`
				Documents [][]string                 `json:"documents"`
				Metadatas [][]map[string]interface{} `json:"metadatas"`
			}{
				Ids:       [][]string{{}},
				Documents: [][]string{{}},
				Metadatas: [][]map[string]interface{}{{}},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &ChromaClient{
		BaseURL:    server.URL,
		Collection: "test",
		Embedder:   &mockEmbedder{},
	}

	ctx := context.Background()
	results, err := client.SimilaritySearch(ctx, []float32{0.1}, 10)

	if err != nil {
		t.Fatalf("SimilaritySearch failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty search, got %d", len(results))
	}
}

func TestChromaClientCollectionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &ChromaClient{
		BaseURL:    server.URL,
		Collection: "test",
		Embedder:   &mockEmbedder{},
	}

	ctx := context.Background()
	err := client.AddDocuments(ctx, []Document{{ID: "test", Content: "test"}})

	if err == nil {
		t.Error("Expected error for server error response")
	}
}

func TestChromaClientAddDocumentsEmbeddingError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	embedder := &mockEmbedder{
		embedFunc: func(ctx context.Context, text string) ([]float32, error) {
			return nil, &testError{msg: "embedding service unavailable"}
		},
	}

	client := &ChromaClient{
		BaseURL:    server.URL,
		Collection: "test",
		Embedder:   embedder,
	}

	ctx := context.Background()
	docs := []Document{
		{ID: "doc-1", Content: "Content without vector"},
	}

	err := client.AddDocuments(ctx, docs)

	if err == nil {
		t.Error("Expected error when embedding fails")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestStoreInterfaceCompliance(t *testing.T) {
	// Verify that ChromaClient implements the Store interface
	var _ Store = &ChromaClient{}
}

func TestDocumentWithAllMetadataTypes(t *testing.T) {
	doc := Document{
		ID:      "doc-meta",
		Content: "Content",
		Metadata: map[string]interface{}{
			"string":  "value",
			"int":     42,
			"float":   3.14,
			"bool":    true,
			"array":   []string{"a", "b", "c"},
			"nested":  map[string]string{"key": "value"},
		},
	}

	if doc.Metadata["string"] != "value" {
		t.Error("Expected string metadata")
	}

	if doc.Metadata["int"] != 42 {
		t.Error("Expected int metadata")
	}

	if doc.Metadata["float"] != 3.14 {
		t.Error("Expected float metadata")
	}

	if doc.Metadata["bool"] != true {
		t.Error("Expected bool metadata")
	}
}

func TestChromaClientSimilaritySearchNoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.URL.Path == "/api/v1/collections/coll-id/query" {
			resp := struct {
				Ids       [][]string                 `json:"ids"`
				Documents [][]string                 `json:"documents"`
				Metadatas [][]map[string]interface{} `json:"metadatas"`
			}{
				Ids:       [][]string{},
				Documents: [][]string{},
				Metadatas: [][]map[string]interface{}{},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &ChromaClient{
		BaseURL:    server.URL,
		Collection: "test",
		Embedder:   &mockEmbedder{},
	}

	ctx := context.Background()
	results, err := client.SimilaritySearch(ctx, []float32{0.1}, 10)

	if err != nil {
		t.Fatalf("SimilaritySearch failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestChromaClientQueryError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.URL.Path == "/api/v1/collections/coll-id/query" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &ChromaClient{
		BaseURL:    server.URL,
		Collection: "test",
		Embedder:   &mockEmbedder{},
	}

	ctx := context.Background()
	_, err := client.SimilaritySearch(ctx, []float32{0.1}, 10)

	if err == nil {
		t.Error("Expected error for query failure")
	}
}

func TestMultipleDocumentsWithMetadata(t *testing.T) {
	docs := []Document{
		{
			ID:       "code-1",
			Content:  "func main() {}",
			Metadata: map[string]interface{}{"file": "main.go", "type": "function"},
		},
		{
			ID:       "code-2",
			Content:  "type User struct {}",
			Metadata: map[string]interface{}{"file": "types.go", "type": "struct"},
		},
		{
			ID:       "doc-1",
			Content:  "# README",
			Metadata: map[string]interface{}{"file": "README.md", "type": "docs"},
		},
	}

	if len(docs) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(docs))
	}

	for i, doc := range docs {
		if doc.ID == "" {
			t.Errorf("Document %d has empty ID", i)
		}
		if doc.Content == "" {
			t.Errorf("Document %d has empty Content", i)
		}
		if doc.Metadata["file"] == nil {
			t.Errorf("Document %d missing file metadata", i)
		}
	}
}
