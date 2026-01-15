package activities

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/memory"
)

type IndexRepoParams struct {
	RepoURL   string
	ProjectID string
}

type Indexer struct {
	store memory.Store
}

func NewIndexer() *Indexer {
	// Initialize with ChromaClient (which uses configurable LLM for embeddings)
	embedder := llm.NewClient()
	store := memory.NewChromaClient(embedder)
	return &Indexer{store: store}
}

func (i *Indexer) IndexRepo(ctx context.Context, params IndexRepoParams) (string, error) {
	// 1. Clone Repo to Temp Dir
	tempDir, err := os.MkdirTemp("", "sovereign-clone-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir) // Cleanup

	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      params.RepoURL,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone: %w", err)
	}

	// 2. Walk and Index
	var docs []memory.Document

	err = filepath.Walk(tempDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") { // Skip .git, .github
				return filepath.SkipDir
			}
			return nil
		}

		// Filter extensions
		ext := filepath.Ext(path)
		if !isSupportedExt(ext) {
			return nil
		}

		// Read content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable
		}

		relPath, _ := filepath.Rel(tempDir, path)

		// Create Document
		docs = append(docs, memory.Document{
			ID:      fmt.Sprintf("%s:%s", params.ProjectID, relPath),
			Content: string(content),
			Metadata: map[string]interface{}{
				"source":   relPath,
				"project":  params.ProjectID,
				"language": ext,
			},
		})
		return nil
	})

	if err != nil {
		return "", err
	}

	// 3. Store in Memory
	// For production, batch this. For MVP, send all given it's small.
	if len(docs) > 0 {
		if err := i.store.AddDocuments(ctx, docs); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("Indexed %d files", len(docs)), nil
}

func isSupportedExt(ext string) bool {
	switch ext {
	case ".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".md", ".json", ".html", ".css":
		return true
	}
	return false
}
