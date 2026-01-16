package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// RAGPipeline provides retrieval-augmented generation
type RAGPipeline struct {
	store    *MultiLevelStore
	embedder llm.Client
	config   RAGConfig
}

// RAGConfig configures RAG behavior
type RAGConfig struct {
	MaxContextTokens int     // Maximum tokens for retrieved context (default: 4000)
	MaxDocuments     int     // Maximum documents to retrieve (default: 10)
	MinRelevance     float32 // Minimum relevance score (default: 0.5)
	IncludeMetadata  bool    // Include file paths, line numbers (default: true)
	GroupByFile      bool    // Group chunks from same file (default: true)
}

// DefaultRAGConfig returns sensible defaults
func DefaultRAGConfig() RAGConfig {
	return RAGConfig{
		MaxContextTokens: 4000,
		MaxDocuments:     10,
		MinRelevance:     0.5,
		IncludeMetadata:  true,
		GroupByFile:      true,
	}
}

// NewRAGPipeline creates a new RAG pipeline
func NewRAGPipeline(store *MultiLevelStore, embedder llm.Client, config RAGConfig) *RAGPipeline {
	if config.MaxContextTokens == 0 {
		config = DefaultRAGConfig()
	}
	return &RAGPipeline{
		store:    store,
		embedder: embedder,
		config:   config,
	}
}

// RetrievalContext holds the result of RAG retrieval
type RetrievalContext struct {
	Query         string              `json:"query"`
	Documents     []EnrichedDocument  `json:"documents"`
	FormattedText string              `json:"formatted_text"`
	TokenEstimate int                 `json:"token_estimate"`
	Sources       []SourceReference   `json:"sources"`
}

// SourceReference tracks where retrieved content came from
type SourceReference struct {
	FilePath  string `json:"file_path,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
	Type      string `json:"type"`
	Score     float32 `json:"score,omitempty"`
}

// Retrieve fetches relevant context for a query
func (r *RAGPipeline) Retrieve(ctx context.Context, query string, projectID, clientID string) (*RetrievalContext, error) {
	// Search across all levels with project/client context
	result, err := r.store.Search(ctx, SearchOptions{
		Query:     query,
		Limit:     r.config.MaxDocuments * 2, // Fetch extra for filtering
		Levels:    []MemoryLevel{ProjectMemory, ClientMemory, GlobalMemory},
		ClientID:  clientID,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Build retrieval context
	retrieval := &RetrievalContext{
		Query:     query,
		Documents: make([]EnrichedDocument, 0),
		Sources:   make([]SourceReference, 0),
	}

	// Group by file if configured
	if r.config.GroupByFile {
		result.Documents = r.groupByFile(result.Documents)
	}

	// Select documents within token budget
	tokenCount := 0
	for _, doc := range result.Documents {
		docTokens := len(doc.Content) / 4 // Rough estimate

		if tokenCount+docTokens > r.config.MaxContextTokens {
			break
		}
		if len(retrieval.Documents) >= r.config.MaxDocuments {
			break
		}

		retrieval.Documents = append(retrieval.Documents, doc)
		retrieval.Sources = append(retrieval.Sources, SourceReference{
			FilePath:  doc.FilePath,
			StartLine: doc.StartLine,
			EndLine:   doc.EndLine,
			Type:      string(doc.Type),
			Score:     doc.Score,
		})
		tokenCount += docTokens
	}

	// Format as text
	retrieval.FormattedText = r.formatContext(retrieval.Documents)
	retrieval.TokenEstimate = tokenCount

	return retrieval, nil
}

// groupByFile groups chunks from the same file together
func (r *RAGPipeline) groupByFile(docs []EnrichedDocument) []EnrichedDocument {
	// Group by file path
	byFile := make(map[string][]EnrichedDocument)
	var order []string

	for _, doc := range docs {
		key := doc.FilePath
		if key == "" {
			key = doc.ID
		}
		if _, exists := byFile[key]; !exists {
			order = append(order, key)
		}
		byFile[key] = append(byFile[key], doc)
	}

	// Flatten while keeping file groups together
	result := make([]EnrichedDocument, 0, len(docs))
	for _, key := range order {
		fileDocs := byFile[key]
		// Sort chunks within file by line number
		sort.Slice(fileDocs, func(i, j int) bool {
			return fileDocs[i].StartLine < fileDocs[j].StartLine
		})
		result = append(result, fileDocs...)
	}

	return result
}

// formatContext formats documents as text for LLM context
func (r *RAGPipeline) formatContext(docs []EnrichedDocument) string {
	if len(docs) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Retrieved Context\n\n")

	currentFile := ""
	for _, doc := range docs {
		if r.config.IncludeMetadata && doc.FilePath != "" && doc.FilePath != currentFile {
			currentFile = doc.FilePath
			sb.WriteString(fmt.Sprintf("### %s", doc.FilePath))
			if doc.StartLine > 0 {
				sb.WriteString(fmt.Sprintf(" (lines %d-%d)", doc.StartLine, doc.EndLine))
			}
			sb.WriteString("\n")
		}

		// Determine language for code block
		lang := doc.Language
		if lang == "" && doc.Type == DocTypeCode {
			lang = inferLanguage(doc.FilePath)
		}

		if doc.Type == DocTypeCode || strings.Contains(doc.Content, "\n") {
			if lang != "" {
				sb.WriteString(fmt.Sprintf("```%s\n", lang))
			} else {
				sb.WriteString("```\n")
			}
			sb.WriteString(doc.Content)
			if !strings.HasSuffix(doc.Content, "\n") {
				sb.WriteString("\n")
			}
			sb.WriteString("```\n\n")
		} else {
			sb.WriteString(doc.Content)
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

// inferLanguage tries to determine language from file extension
func inferLanguage(filePath string) string {
	if filePath == "" {
		return ""
	}

	ext := ""
	if idx := strings.LastIndex(filePath, "."); idx >= 0 {
		ext = filePath[idx:]
	}

	switch strings.ToLower(ext) {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "tsx"
	case ".jsx":
		return "jsx"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md":
		return "markdown"
	default:
		return ""
	}
}

// AugmentPrompt adds retrieved context to a prompt
func (r *RAGPipeline) AugmentPrompt(ctx context.Context, prompt, query, projectID, clientID string) (string, *RetrievalContext, error) {
	// Retrieve relevant context
	retrieval, err := r.Retrieve(ctx, query, projectID, clientID)
	if err != nil {
		// Don't fail completely, just return original prompt
		return prompt, nil, nil
	}

	if len(retrieval.Documents) == 0 {
		return prompt, retrieval, nil
	}

	// Augment the prompt
	augmented := retrieval.FormattedText + "\n\n" + prompt

	return augmented, retrieval, nil
}

// AugmentSystemPrompt adds retrieved context to a system prompt
func (r *RAGPipeline) AugmentSystemPrompt(ctx context.Context, systemPrompt, query, projectID, clientID string) (string, error) {
	retrieval, err := r.Retrieve(ctx, query, projectID, clientID)
	if err != nil {
		return systemPrompt, nil
	}

	if len(retrieval.Documents) == 0 {
		return systemPrompt, nil
	}

	augmented := systemPrompt + "\n\n" + retrieval.FormattedText
	return augmented, nil
}

// RAGRequest configures a RAG-augmented generation
type RAGRequest struct {
	Query         string
	SystemPrompt  string
	UserPrompt    string
	ProjectID     string
	ClientID      string
	Temperature   float64
	DocumentTypes []DocumentType // Filter by type
	Languages     []string       // Filter by language
}

// RAGResponse contains the result of RAG-augmented generation
type RAGResponse struct {
	Response      string            `json:"response"`
	Retrieval     *RetrievalContext `json:"retrieval"`
	TokensUsed    int               `json:"tokens_used"`
}

// Generate performs RAG-augmented generation
func (r *RAGPipeline) Generate(ctx context.Context, req RAGRequest) (*RAGResponse, error) {
	// Build search query from user prompt if not specified
	query := req.Query
	if query == "" {
		query = req.UserPrompt
	}

	// Retrieve context
	retrieval, err := r.Retrieve(ctx, query, req.ProjectID, req.ClientID)
	if err != nil {
		// Continue without context on error
		retrieval = &RetrievalContext{}
	}

	// Augment system prompt
	systemPrompt := req.SystemPrompt
	if len(retrieval.Documents) > 0 {
		systemPrompt = systemPrompt + "\n\n" + retrieval.FormattedText
	}

	// Generate with LLM
	resp, err := r.embedder.Generate(ctx, llm.GenerateRequest{
		Prompt:      req.UserPrompt,
		System:      systemPrompt,
		Temperature: req.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	return &RAGResponse{
		Response:  resp.Response,
		Retrieval: retrieval,
	}, nil
}

// IndexCodebase indexes a codebase for RAG
func (r *RAGPipeline) IndexCodebase(ctx context.Context, projectID, clientID string, files map[string][]byte) error {
	chunker := NewCodeChunker(DefaultChunkConfig())
	defer chunker.Close()

	for filePath, content := range files {
		result, err := chunker.ChunkFileWithStats(ctx, filePath, content)
		if err != nil {
			continue
		}

		err = r.store.StoreCodeContext(ctx, projectID, clientID, filePath, result.Language, string(content), result.Chunks)
		if err != nil {
			continue
		}
	}

	return nil
}

// FindSimilarCode finds code similar to a query
func (r *RAGPipeline) FindSimilarCode(ctx context.Context, query, projectID string, limit int) ([]EnrichedDocument, error) {
	result, err := r.store.Search(ctx, SearchOptions{
		Query:     query,
		Limit:     limit,
		Levels:    []MemoryLevel{ProjectMemory},
		ProjectID: projectID,
		Types:     []DocumentType{DocTypeCode},
	})
	if err != nil {
		return nil, err
	}

	return result.Documents, nil
}

// FindPatterns finds relevant patterns for a query
func (r *RAGPipeline) FindPatterns(ctx context.Context, query string, clientID string, limit int) ([]EnrichedDocument, error) {
	result, err := r.store.Search(ctx, SearchOptions{
		Query:    query,
		Limit:    limit,
		Levels:   []MemoryLevel{GlobalMemory, ClientMemory},
		ClientID: clientID,
		Types:    []DocumentType{DocTypePattern},
	})
	if err != nil {
		return nil, err
	}

	return result.Documents, nil
}

// FindDecisions finds relevant architectural decisions
func (r *RAGPipeline) FindDecisions(ctx context.Context, query, projectID string, limit int) ([]EnrichedDocument, error) {
	result, err := r.store.Search(ctx, SearchOptions{
		Query:     query,
		Limit:     limit,
		Levels:    []MemoryLevel{ProjectMemory},
		ProjectID: projectID,
		Types:     []DocumentType{DocTypeDecision},
	})
	if err != nil {
		return nil, err
	}

	return result.Documents, nil
}

// LearnFromError stores an error and resolution for future reference
func (r *RAGPipeline) LearnFromError(ctx context.Context, projectID, clientID, errorMsg, resolution string) error {
	// Store at project level
	if err := r.store.StoreError(ctx, ProjectMemory, clientID, projectID, errorMsg, resolution); err != nil {
		return err
	}

	// If it's a common error pattern, also store globally
	if isCommonError(errorMsg) {
		return r.store.StoreError(ctx, GlobalMemory, "", "", errorMsg, resolution)
	}

	return nil
}

// isCommonError checks if an error is likely to be common across projects
func isCommonError(errorMsg string) bool {
	commonPatterns := []string{
		"cannot find module",
		"undefined is not",
		"null pointer",
		"type error",
		"import error",
		"syntax error",
		"permission denied",
		"connection refused",
	}

	lowerErr := strings.ToLower(errorMsg)
	for _, pattern := range commonPatterns {
		if strings.Contains(lowerErr, pattern) {
			return true
		}
	}
	return false
}
