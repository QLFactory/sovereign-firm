package memory

import "context"

type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Vector   []float32              `json:"vector,omitempty"` // Computed embedding
}

// VectorStore abstraction for Air-Gapped RAG
type Store interface {
	// AddDocuments embeds and stores documents
	AddDocuments(ctx context.Context, docs []Document) error

	// SimilaritySearch finds relevant docs
	SimilaritySearch(ctx context.Context, vector []float32, limit int) ([]Document, error)
}
