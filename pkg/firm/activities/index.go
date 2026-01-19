package activities

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/qlfactory/sovereign-firm/pkg/agent"
	"github.com/qlfactory/sovereign-firm/pkg/mcp"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/memory"
)

type Indexer struct {
	store memory.Store
	pool  *agent.AgentPool
}

func NewIndexer(pool *agent.AgentPool) *Indexer {
	embedder := llm.NewClient()
	store := memory.NewChromaClient(embedder)
	return &Indexer{
		store: store,
		pool:  pool,
	}
}

// IndexRepository handles the indexing of a repository
func (a *Indexer) IndexRepository(ctx context.Context, params IndexRepoParams) (map[string]interface{}, error) {
	projectID := params.ProjectID
	repoURL := params.RepoURL

	if projectID == "" {
		projectID = "default-project"
	}

	// Spawn or get agent for this project
	agentInstance, err := a.pool.SpawnAgent(ctx, "Indexer", "indexer-agent", projectID, "gpt-4", []string{"project-manager"})
	if err != nil {
		return nil, fmt.Errorf("failed to spawn agent: %w", err)
	}

	// Setup tool registry
	registry := mcp.NewToolRegistry()
	mcp.RegisterBuiltinToolsWithConfig(registry, ".", &mcp.ToolConfig{
		ProjectID: projectID,
		AgentID:   agentInstance.ID,
		Messaging: a.pool,
	})

	// 1. Clone Repo to Temp Dir
	tempDir, err := os.MkdirTemp("", "sovereign-clone-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir) // Cleanup

	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone: %w", err)
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
			ID:      fmt.Sprintf("%s:%s", projectID, relPath),
			Content: string(content),
			Metadata: map[string]interface{}{
				"source":   relPath,
				"project":  projectID,
				"language": ext,
			},
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 3. Store in Memory
	// For production, batch this. For MVP, send all given it's small.
	if len(docs) > 0 {
		if err := a.store.AddDocuments(ctx, docs); err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		"message": fmt.Sprintf("Indexed %d files", len(docs)),
		"count":   len(docs),
	}, nil
}

func isSupportedExt(ext string) bool {
	switch ext {
	case ".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".md", ".json", ".html", ".css":
		return true
	}
	return false
}
