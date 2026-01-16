package memory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// MemoryLevel represents the scope of memory
type MemoryLevel string

const (
	// GlobalMemory - shared across all projects and clients
	// Contains: best practices, common patterns, reusable templates
	GlobalMemory MemoryLevel = "global"

	// ClientMemory - shared across all projects for a specific client
	// Contains: client preferences, tech stack standards, domain knowledge
	ClientMemory MemoryLevel = "client"

	// ProjectMemory - specific to a single project
	// Contains: codebase context, decisions made, file contents
	ProjectMemory MemoryLevel = "project"
)

// DocumentType classifies the kind of content stored
type DocumentType string

const (
	DocTypeCode        DocumentType = "code"         // Source code
	DocTypePattern     DocumentType = "pattern"      // Design pattern or best practice
	DocTypeDecision    DocumentType = "decision"     // Architectural decision
	DocTypeSpec        DocumentType = "spec"         // Specification or requirement
	DocTypeConversation DocumentType = "conversation" // Chat history
	DocTypeError       DocumentType = "error"        // Error and resolution
	DocTypeTemplate    DocumentType = "template"     // Code template
)

// EnrichedDocument extends Document with additional metadata
type EnrichedDocument struct {
	Document
	Level       MemoryLevel  `json:"level"`
	Type        DocumentType `json:"type"`
	ClientID    string       `json:"client_id,omitempty"`
	ProjectID   string       `json:"project_id,omitempty"`
	FilePath    string       `json:"file_path,omitempty"`
	Language    string       `json:"language,omitempty"`
	StartLine   int          `json:"start_line,omitempty"`
	EndLine     int          `json:"end_line,omitempty"`
	Score       float32      `json:"score,omitempty"`       // Relevance score from search
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	AccessCount int          `json:"access_count,omitempty"`
}

// MultiLevelStore provides hierarchical memory storage
type MultiLevelStore struct {
	embedder llm.Client
	clients  map[MemoryLevel]*ChromaClient
}

// NewMultiLevelStore creates a new multi-level memory store
func NewMultiLevelStore(embedder llm.Client, baseURL string) *MultiLevelStore {
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	return &MultiLevelStore{
		embedder: embedder,
		clients: map[MemoryLevel]*ChromaClient{
			GlobalMemory: {
				BaseURL:    baseURL,
				Collection: "sovereign-global",
				Embedder:   embedder,
			},
			ClientMemory: {
				BaseURL:    baseURL,
				Collection: "sovereign-clients",
				Embedder:   embedder,
			},
			ProjectMemory: {
				BaseURL:    baseURL,
				Collection: "sovereign-projects",
				Embedder:   embedder,
			},
		},
	}
}

// Add stores a document at the specified level
func (m *MultiLevelStore) Add(ctx context.Context, doc EnrichedDocument) error {
	client, ok := m.clients[doc.Level]
	if !ok {
		return fmt.Errorf("unknown memory level: %s", doc.Level)
	}

	// Generate ID if not provided
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}
	doc.UpdatedAt = now

	// Build metadata
	metadata := map[string]interface{}{
		"level":      string(doc.Level),
		"type":       string(doc.Type),
		"created_at": doc.CreatedAt.Unix(),
		"updated_at": doc.UpdatedAt.Unix(),
	}

	if doc.ClientID != "" {
		metadata["client_id"] = doc.ClientID
	}
	if doc.ProjectID != "" {
		metadata["project_id"] = doc.ProjectID
	}
	if doc.FilePath != "" {
		metadata["file_path"] = doc.FilePath
	}
	if doc.Language != "" {
		metadata["language"] = doc.Language
	}
	if doc.StartLine > 0 {
		metadata["start_line"] = doc.StartLine
	}
	if doc.EndLine > 0 {
		metadata["end_line"] = doc.EndLine
	}

	// Merge with existing metadata
	for k, v := range doc.Metadata {
		metadata[k] = v
	}

	return client.AddDocuments(ctx, []Document{{
		ID:       doc.ID,
		Content:  doc.Content,
		Metadata: metadata,
		Vector:   doc.Vector,
	}})
}

// AddBatch stores multiple documents at the specified level
func (m *MultiLevelStore) AddBatch(ctx context.Context, docs []EnrichedDocument) error {
	// Group by level
	byLevel := make(map[MemoryLevel][]Document)

	for _, doc := range docs {
		if doc.ID == "" {
			doc.ID = uuid.New().String()
		}

		now := time.Now()
		if doc.CreatedAt.IsZero() {
			doc.CreatedAt = now
		}
		doc.UpdatedAt = now

		metadata := map[string]interface{}{
			"level":      string(doc.Level),
			"type":       string(doc.Type),
			"created_at": doc.CreatedAt.Unix(),
			"updated_at": doc.UpdatedAt.Unix(),
		}

		if doc.ClientID != "" {
			metadata["client_id"] = doc.ClientID
		}
		if doc.ProjectID != "" {
			metadata["project_id"] = doc.ProjectID
		}
		if doc.FilePath != "" {
			metadata["file_path"] = doc.FilePath
		}
		if doc.Language != "" {
			metadata["language"] = doc.Language
		}

		for k, v := range doc.Metadata {
			metadata[k] = v
		}

		byLevel[doc.Level] = append(byLevel[doc.Level], Document{
			ID:       doc.ID,
			Content:  doc.Content,
			Metadata: metadata,
			Vector:   doc.Vector,
		})
	}

	// Store each level
	for level, docs := range byLevel {
		client, ok := m.clients[level]
		if !ok {
			continue
		}
		if err := client.AddDocuments(ctx, docs); err != nil {
			return fmt.Errorf("failed to add docs to %s: %w", level, err)
		}
	}

	return nil
}

// SearchOptions configures memory search behavior
type SearchOptions struct {
	Query      string       // Text query to embed and search
	Vector     []float32    // Pre-computed embedding (optional)
	Limit      int          // Max results per level
	Levels     []MemoryLevel // Which levels to search (default: all)
	ClientID   string       // Filter by client
	ProjectID  string       // Filter by project
	Types      []DocumentType // Filter by document type
	Languages  []string     // Filter by programming language
	MinScore   float32      // Minimum relevance score
}

// SearchResult contains results from multi-level search
type SearchResult struct {
	Documents []EnrichedDocument `json:"documents"`
	Query     string             `json:"query"`
	Levels    []MemoryLevel      `json:"levels_searched"`
	TotalHits int                `json:"total_hits"`
}

// Search queries across memory levels with optional filtering
func (m *MultiLevelStore) Search(ctx context.Context, opts SearchOptions) (*SearchResult, error) {
	// Default limit
	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	// Default to all levels
	if len(opts.Levels) == 0 {
		opts.Levels = []MemoryLevel{GlobalMemory, ClientMemory, ProjectMemory}
	}

	// Get or compute embedding
	var vector []float32
	var err error
	if len(opts.Vector) > 0 {
		vector = opts.Vector
	} else if opts.Query != "" {
		vector, err = m.embedder.Embed(ctx, opts.Query)
		if err != nil {
			return nil, fmt.Errorf("failed to embed query: %w", err)
		}
	} else {
		return nil, fmt.Errorf("either Query or Vector must be provided")
	}

	result := &SearchResult{
		Documents: make([]EnrichedDocument, 0),
		Query:     opts.Query,
		Levels:    opts.Levels,
	}

	// Search each level
	for _, level := range opts.Levels {
		client, ok := m.clients[level]
		if !ok {
			continue
		}

		docs, err := client.SimilaritySearch(ctx, vector, opts.Limit*2) // Fetch extra for filtering
		if err != nil {
			continue // Don't fail completely if one level errors
		}

		// Convert and filter results
		for _, doc := range docs {
			enriched := m.enrichDocument(doc, level)

			// Apply filters
			if !m.matchesFilters(enriched, opts) {
				continue
			}

			result.Documents = append(result.Documents, enriched)
		}
	}

	// Sort by relevance (already sorted from ChromaDB, but we may want to re-rank)
	// For now, just limit results
	if len(result.Documents) > opts.Limit {
		result.Documents = result.Documents[:opts.Limit]
	}

	result.TotalHits = len(result.Documents)
	return result, nil
}

// enrichDocument converts a basic Document to EnrichedDocument
func (m *MultiLevelStore) enrichDocument(doc Document, level MemoryLevel) EnrichedDocument {
	enriched := EnrichedDocument{
		Document: doc,
		Level:    level,
	}

	// Extract metadata
	if doc.Metadata != nil {
		if t, ok := doc.Metadata["type"].(string); ok {
			enriched.Type = DocumentType(t)
		}
		if cid, ok := doc.Metadata["client_id"].(string); ok {
			enriched.ClientID = cid
		}
		if pid, ok := doc.Metadata["project_id"].(string); ok {
			enriched.ProjectID = pid
		}
		if fp, ok := doc.Metadata["file_path"].(string); ok {
			enriched.FilePath = fp
		}
		if lang, ok := doc.Metadata["language"].(string); ok {
			enriched.Language = lang
		}
		if sl, ok := doc.Metadata["start_line"].(float64); ok {
			enriched.StartLine = int(sl)
		}
		if el, ok := doc.Metadata["end_line"].(float64); ok {
			enriched.EndLine = int(el)
		}
		if ts, ok := doc.Metadata["created_at"].(float64); ok {
			enriched.CreatedAt = time.Unix(int64(ts), 0)
		}
		if ts, ok := doc.Metadata["updated_at"].(float64); ok {
			enriched.UpdatedAt = time.Unix(int64(ts), 0)
		}
	}

	return enriched
}

// matchesFilters checks if a document matches the search filters
func (m *MultiLevelStore) matchesFilters(doc EnrichedDocument, opts SearchOptions) bool {
	// Client filter
	if opts.ClientID != "" && doc.ClientID != opts.ClientID {
		return false
	}

	// Project filter
	if opts.ProjectID != "" && doc.ProjectID != opts.ProjectID {
		return false
	}

	// Type filter
	if len(opts.Types) > 0 {
		found := false
		for _, t := range opts.Types {
			if doc.Type == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Language filter
	if len(opts.Languages) > 0 {
		found := false
		for _, lang := range opts.Languages {
			if strings.EqualFold(doc.Language, lang) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// GetProjectContext retrieves all relevant context for a project
func (m *MultiLevelStore) GetProjectContext(ctx context.Context, projectID, clientID, query string, limit int) (*SearchResult, error) {
	return m.Search(ctx, SearchOptions{
		Query:     query,
		Limit:     limit,
		Levels:    []MemoryLevel{ProjectMemory, ClientMemory, GlobalMemory},
		ClientID:  clientID,
		ProjectID: projectID,
	})
}

// StoreCodeContext stores code from a file with proper chunking metadata
func (m *MultiLevelStore) StoreCodeContext(ctx context.Context, projectID, clientID, filePath, language, content string, chunks []CodeChunk) error {
	docs := make([]EnrichedDocument, len(chunks))

	for i, chunk := range chunks {
		docs[i] = EnrichedDocument{
			Document: Document{
				ID:      fmt.Sprintf("%s:%s:%d", projectID, filePath, chunk.StartLine),
				Content: chunk.Content,
			},
			Level:     ProjectMemory,
			Type:      DocTypeCode,
			ClientID:  clientID,
			ProjectID: projectID,
			FilePath:  filePath,
			Language:  language,
			StartLine: chunk.StartLine,
			EndLine:   chunk.EndLine,
		}
	}

	return m.AddBatch(ctx, docs)
}

// StoreDecision records an architectural decision
func (m *MultiLevelStore) StoreDecision(ctx context.Context, projectID, clientID, title, rationale string, metadata map[string]interface{}) error {
	content := fmt.Sprintf("Decision: %s\n\nRationale: %s", title, rationale)

	doc := EnrichedDocument{
		Document: Document{
			Content:  content,
			Metadata: metadata,
		},
		Level:     ProjectMemory,
		Type:      DocTypeDecision,
		ClientID:  clientID,
		ProjectID: projectID,
	}

	return m.Add(ctx, doc)
}

// StorePattern stores a reusable code pattern at global or client level
func (m *MultiLevelStore) StorePattern(ctx context.Context, level MemoryLevel, clientID, name, description, example string) error {
	content := fmt.Sprintf("Pattern: %s\n\n%s\n\nExample:\n%s", name, description, example)

	doc := EnrichedDocument{
		Document: Document{
			Content: content,
			Metadata: map[string]interface{}{
				"pattern_name": name,
			},
		},
		Level:    level,
		Type:     DocTypePattern,
		ClientID: clientID,
	}

	return m.Add(ctx, doc)
}

// StoreConversation stores chat history for context
func (m *MultiLevelStore) StoreConversation(ctx context.Context, projectID, clientID, conversationID, content string) error {
	doc := EnrichedDocument{
		Document: Document{
			ID:      conversationID,
			Content: content,
			Metadata: map[string]interface{}{
				"conversation_id": conversationID,
			},
		},
		Level:     ProjectMemory,
		Type:      DocTypeConversation,
		ClientID:  clientID,
		ProjectID: projectID,
	}

	return m.Add(ctx, doc)
}

// StoreError records an error and its resolution for future reference
func (m *MultiLevelStore) StoreError(ctx context.Context, level MemoryLevel, clientID, projectID, errorMsg, resolution string) error {
	content := fmt.Sprintf("Error: %s\n\nResolution: %s", errorMsg, resolution)

	doc := EnrichedDocument{
		Document: Document{
			Content: content,
			Metadata: map[string]interface{}{
				"error_message": errorMsg,
			},
		},
		Level:     level,
		Type:      DocTypeError,
		ClientID:  clientID,
		ProjectID: projectID,
	}

	return m.Add(ctx, doc)
}

// CodeChunk represents a chunk of code for storage
type CodeChunk struct {
	Content   string
	StartLine int
	EndLine   int
	Symbol    string // Function/class name if applicable
}

// DeleteProjectMemory removes all documents for a project
func (m *MultiLevelStore) DeleteProjectMemory(ctx context.Context, projectID string) error {
	// Note: ChromaDB doesn't have a direct "delete by metadata" API
	// This would need to be implemented with a query + delete by IDs
	// For now, this is a placeholder
	return nil
}

// GetStats returns statistics about stored documents
func (m *MultiLevelStore) GetStats(ctx context.Context) (map[MemoryLevel]int, error) {
	// This would need to be implemented with collection info queries
	// For now, return placeholder
	return map[MemoryLevel]int{
		GlobalMemory:  0,
		ClientMemory:  0,
		ProjectMemory: 0,
	}, nil
}
