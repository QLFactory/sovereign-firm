package memory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewMultiLevelStore(t *testing.T) {
	embedder := &mockEmbedder{}
	store := NewMultiLevelStore(embedder, "http://localhost:8000")

	if store == nil {
		t.Fatal("Expected store to be created")
	}

	if len(store.clients) != 3 {
		t.Errorf("Expected 3 clients (global, client, project), got %d", len(store.clients))
	}

	// Check collection names
	if store.clients[GlobalMemory].Collection != "sovereign-global" {
		t.Errorf("Expected global collection 'sovereign-global', got '%s'", store.clients[GlobalMemory].Collection)
	}
	if store.clients[ClientMemory].Collection != "sovereign-clients" {
		t.Errorf("Expected client collection 'sovereign-clients', got '%s'", store.clients[ClientMemory].Collection)
	}
	if store.clients[ProjectMemory].Collection != "sovereign-projects" {
		t.Errorf("Expected project collection 'sovereign-projects', got '%s'", store.clients[ProjectMemory].Collection)
	}
}

func TestMultiLevelStoreAdd(t *testing.T) {
	var collectionRequests int
	var addRequests int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			collectionRequests++
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.Method == "POST" && len(r.URL.Path) > 20 {
			addRequests++
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	embedder := &mockEmbedder{}
	store := NewMultiLevelStore(embedder, server.URL)

	ctx := context.Background()
	doc := EnrichedDocument{
		Document: Document{
			Content: "Test content",
		},
		Level:     ProjectMemory,
		Type:      DocTypeCode,
		ClientID:  "client-123",
		ProjectID: "project-456",
		FilePath:  "main.go",
		Language:  "go",
	}

	err := store.Add(ctx, doc)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if addRequests == 0 {
		t.Error("Expected document to be added")
	}
}

func TestMultiLevelStoreAddBatch(t *testing.T) {
	var addCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.Method == "POST" {
			addCount++
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	embedder := &mockEmbedder{}
	store := NewMultiLevelStore(embedder, server.URL)

	ctx := context.Background()
	docs := []EnrichedDocument{
		{
			Document: Document{Content: "Global pattern"},
			Level:    GlobalMemory,
			Type:     DocTypePattern,
		},
		{
			Document:  Document{Content: "Project code"},
			Level:     ProjectMemory,
			Type:      DocTypeCode,
			ProjectID: "proj-1",
		},
		{
			Document:  Document{Content: "Client standard"},
			Level:     ClientMemory,
			Type:      DocTypePattern,
			ClientID:  "client-1",
		},
	}

	err := store.AddBatch(ctx, docs)
	if err != nil {
		t.Fatalf("AddBatch failed: %v", err)
	}

	// Should have made 3 add requests (one for each level)
	if addCount != 3 {
		t.Errorf("Expected 3 add requests, got %d", addCount)
	}
}

func TestEnrichedDocumentMetadata(t *testing.T) {
	doc := EnrichedDocument{
		Document: Document{
			ID:      "doc-123",
			Content: "function test() {}",
		},
		Level:     ProjectMemory,
		Type:      DocTypeCode,
		ClientID:  "client-abc",
		ProjectID: "project-xyz",
		FilePath:  "src/test.js",
		Language:  "javascript",
		StartLine: 10,
		EndLine:   20,
		CreatedAt: time.Now(),
	}

	if doc.Level != ProjectMemory {
		t.Errorf("Expected level ProjectMemory, got %s", doc.Level)
	}

	if doc.Type != DocTypeCode {
		t.Errorf("Expected type code, got %s", doc.Type)
	}

	if doc.FilePath != "src/test.js" {
		t.Errorf("Expected file path 'src/test.js', got '%s'", doc.FilePath)
	}
}

func TestSearchOptions(t *testing.T) {
	opts := SearchOptions{
		Query:     "authentication",
		Limit:     5,
		Levels:    []MemoryLevel{ProjectMemory, GlobalMemory},
		ClientID:  "client-1",
		ProjectID: "project-1",
		Types:     []DocumentType{DocTypeCode, DocTypePattern},
		Languages: []string{"go", "typescript"},
	}

	if opts.Query != "authentication" {
		t.Error("Expected query to be set")
	}

	if opts.Limit != 5 {
		t.Error("Expected limit 5")
	}

	if len(opts.Levels) != 2 {
		t.Error("Expected 2 levels")
	}

	if len(opts.Types) != 2 {
		t.Error("Expected 2 types")
	}

	if len(opts.Languages) != 2 {
		t.Error("Expected 2 languages")
	}
}

func TestMemoryLevelConstants(t *testing.T) {
	if GlobalMemory != "global" {
		t.Errorf("Expected GlobalMemory to be 'global', got '%s'", GlobalMemory)
	}

	if ClientMemory != "client" {
		t.Errorf("Expected ClientMemory to be 'client', got '%s'", ClientMemory)
	}

	if ProjectMemory != "project" {
		t.Errorf("Expected ProjectMemory to be 'project', got '%s'", ProjectMemory)
	}
}

func TestDocumentTypeConstants(t *testing.T) {
	types := []DocumentType{
		DocTypeCode,
		DocTypePattern,
		DocTypeDecision,
		DocTypeSpec,
		DocTypeConversation,
		DocTypeError,
		DocTypeTemplate,
	}

	expectedValues := []string{
		"code", "pattern", "decision", "spec", "conversation", "error", "template",
	}

	for i, docType := range types {
		if string(docType) != expectedValues[i] {
			t.Errorf("Expected %s, got %s", expectedValues[i], docType)
		}
	}
}

func TestMatchesFilters(t *testing.T) {
	store := &MultiLevelStore{}

	doc := EnrichedDocument{
		Document:  Document{ID: "test"},
		ClientID:  "client-1",
		ProjectID: "project-1",
		Type:      DocTypeCode,
		Language:  "go",
	}

	// Test client filter
	if !store.matchesFilters(doc, SearchOptions{ClientID: "client-1"}) {
		t.Error("Expected doc to match client filter")
	}
	if store.matchesFilters(doc, SearchOptions{ClientID: "client-2"}) {
		t.Error("Expected doc to NOT match wrong client filter")
	}

	// Test project filter
	if !store.matchesFilters(doc, SearchOptions{ProjectID: "project-1"}) {
		t.Error("Expected doc to match project filter")
	}
	if store.matchesFilters(doc, SearchOptions{ProjectID: "project-2"}) {
		t.Error("Expected doc to NOT match wrong project filter")
	}

	// Test type filter
	if !store.matchesFilters(doc, SearchOptions{Types: []DocumentType{DocTypeCode}}) {
		t.Error("Expected doc to match type filter")
	}
	if store.matchesFilters(doc, SearchOptions{Types: []DocumentType{DocTypePattern}}) {
		t.Error("Expected doc to NOT match wrong type filter")
	}

	// Test language filter
	if !store.matchesFilters(doc, SearchOptions{Languages: []string{"go"}}) {
		t.Error("Expected doc to match language filter")
	}
	if store.matchesFilters(doc, SearchOptions{Languages: []string{"python"}}) {
		t.Error("Expected doc to NOT match wrong language filter")
	}
}

func TestEnrichDocument(t *testing.T) {
	store := &MultiLevelStore{}

	doc := Document{
		ID:      "test-doc",
		Content: "content",
		Metadata: map[string]interface{}{
			"type":       "code",
			"client_id":  "client-1",
			"project_id": "project-1",
			"file_path":  "main.go",
			"language":   "go",
			"start_line": float64(10),
			"end_line":   float64(20),
			"created_at": float64(1700000000),
			"updated_at": float64(1700001000),
		},
	}

	enriched := store.enrichDocument(doc, ProjectMemory)

	if enriched.ID != "test-doc" {
		t.Errorf("Expected ID 'test-doc', got '%s'", enriched.ID)
	}

	if enriched.Level != ProjectMemory {
		t.Errorf("Expected level ProjectMemory, got %s", enriched.Level)
	}

	if enriched.Type != DocTypeCode {
		t.Errorf("Expected type code, got %s", enriched.Type)
	}

	if enriched.ClientID != "client-1" {
		t.Errorf("Expected client_id 'client-1', got '%s'", enriched.ClientID)
	}

	if enriched.ProjectID != "project-1" {
		t.Errorf("Expected project_id 'project-1', got '%s'", enriched.ProjectID)
	}

	if enriched.FilePath != "main.go" {
		t.Errorf("Expected file_path 'main.go', got '%s'", enriched.FilePath)
	}

	if enriched.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", enriched.Language)
	}

	if enriched.StartLine != 10 {
		t.Errorf("Expected start_line 10, got %d", enriched.StartLine)
	}

	if enriched.EndLine != 20 {
		t.Errorf("Expected end_line 20, got %d", enriched.EndLine)
	}
}

func TestCodeChunkStructure(t *testing.T) {
	chunk := CodeChunk{
		Content:   "func hello() {}",
		StartLine: 1,
		EndLine:   3,
		Symbol:    "hello",
	}

	if chunk.Content != "func hello() {}" {
		t.Error("Expected content to be set")
	}

	if chunk.StartLine != 1 {
		t.Error("Expected start line 1")
	}

	if chunk.EndLine != 3 {
		t.Error("Expected end line 3")
	}

	if chunk.Symbol != "hello" {
		t.Error("Expected symbol 'hello'")
	}
}

func TestStorePatternHelper(t *testing.T) {
	var addedDocs []map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.Method == "POST" {
			var payload map[string]interface{}
			json.NewDecoder(r.Body).Decode(&payload)
			if docs, ok := payload["documents"].([]interface{}); ok && len(docs) > 0 {
				addedDocs = append(addedDocs, payload)
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	embedder := &mockEmbedder{}
	store := NewMultiLevelStore(embedder, server.URL)

	ctx := context.Background()
	err := store.StorePattern(ctx, GlobalMemory, "", "Singleton Pattern",
		"Ensures a class has only one instance",
		"type singleton struct{}\nvar instance *singleton")

	if err != nil {
		t.Fatalf("StorePattern failed: %v", err)
	}

	if len(addedDocs) == 0 {
		t.Error("Expected pattern to be stored")
	}
}

func TestStoreDecisionHelper(t *testing.T) {
	var addedDocs []map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/collections" {
			resp := map[string]string{"id": "coll-id"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.Method == "POST" {
			var payload map[string]interface{}
			json.NewDecoder(r.Body).Decode(&payload)
			if docs, ok := payload["documents"].([]interface{}); ok && len(docs) > 0 {
				addedDocs = append(addedDocs, payload)
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	embedder := &mockEmbedder{}
	store := NewMultiLevelStore(embedder, server.URL)

	ctx := context.Background()
	err := store.StoreDecision(ctx, "project-1", "client-1",
		"Use PostgreSQL for persistence",
		"PostgreSQL offers better JSON support and full-text search",
		map[string]interface{}{"category": "database"})

	if err != nil {
		t.Fatalf("StoreDecision failed: %v", err)
	}

	if len(addedDocs) == 0 {
		t.Error("Expected decision to be stored")
	}
}
