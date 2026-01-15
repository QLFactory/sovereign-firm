package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

type ChromaClient struct {
	BaseURL    string
	Collection string
	Embedder   llm.Client
}

func NewChromaClient(embedder llm.Client) *ChromaClient {
	url := os.Getenv("CHROMA_HOST")
	if url == "" {
		url = "http://localhost:8000"
	}
	coll := os.Getenv("CHROMA_COLLECTION")
	if coll == "" {
		coll = "sovereign-firm-memory-v2"
	}
	return &ChromaClient{
		BaseURL:    url,
		Collection: coll,
		Embedder:   embedder,
	}
}

// AddDocuments embeds texts using the LLM client and pushes to Chroma
func (c *ChromaClient) AddDocuments(ctx context.Context, docs []Document) error {
	// 1. Ensure Collection Exists
	collID, err := c.getOrCreateCollection(ctx)
	if err != nil {
		return err
	}

	// 2. Prepare payload
	payload := map[string]interface{}{
		"ids":        make([]string, len(docs)),
		"embeddings": make([][]float32, len(docs)),
		"documents":  make([]string, len(docs)),
		"metadatas":  make([]map[string]interface{}, len(docs)),
	}

	for i, doc := range docs {
		// Embed if missing
		if len(doc.Vector) == 0 {
			emb, err := c.Embedder.Embed(ctx, doc.Content)
			if err != nil {
				return fmt.Errorf("embedding failed for doc %s: %w", doc.ID, err)
			}
			doc.Vector = emb
		}

		payload["ids"].([]string)[i] = doc.ID
		payload["embeddings"].([][]float32)[i] = doc.Vector
		payload["documents"].([]string)[i] = doc.Content
		payload["metadatas"].([]map[string]interface{})[i] = doc.Metadata
	}

	// 3. POST /api/v1/collections/{id}/add
	return c.postJSON(ctx, "/api/v1/collections/"+collID+"/add", payload, nil)
}

func (c *ChromaClient) SimilaritySearch(ctx context.Context, vector []float32, limit int) ([]Document, error) {
	collID, err := c.getOrCreateCollection(ctx)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"query_embeddings": [][]float32{vector},
		"n_results":        limit,
	}

	var resp struct {
		Ids       [][]string                 `json:"ids"`
		Documents [][]string                 `json:"documents"`
		Metadatas [][]map[string]interface{} `json:"metadatas"`
	}

	if err := c.postJSON(ctx, "/api/v1/collections/"+collID+"/query", payload, &resp); err != nil {
		return nil, err
	}

	if len(resp.Ids) == 0 || len(resp.Ids[0]) == 0 {
		return []Document{}, nil
	}

	results := make([]Document, len(resp.Ids[0]))
	for i := range results {
		results[i] = Document{
			ID:       resp.Ids[0][i],
			Content:  resp.Documents[0][i],
			Metadata: resp.Metadatas[0][i],
		}
	}
	return results, nil
}

// Helper: Get Collection ID by name, create if not exists
func (c *ChromaClient) getOrCreateCollection(ctx context.Context) (string, error) {
	// Try Create
	payload := map[string]string{"name": c.Collection, "get_or_create": "true"}
	var resp struct {
		ID string `json:"id"`
	}
	// Note: Recent Chroma versions use POST /api/v1/collections
	if err := c.postJSON(ctx, "/api/v1/collections", payload, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *ChromaClient) postJSON(ctx context.Context, path string, reqData interface{}, respData interface{}) error {
	body, _ := json.Marshal(reqData)
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	cl := &http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("chroma error %d", resp.StatusCode)
	}

	if respData != nil {
		return json.NewDecoder(resp.Body).Decode(respData)
	}
	return nil
}
