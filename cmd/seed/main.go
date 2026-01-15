package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"path/filepath"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/memory"
)

func main() {
	fmt.Println("🌱 Seeding Sovereign Memory...")

	// Configure Env
	os.Setenv("CHROMA_HOST", "http://localhost:8000")
	if os.Getenv("OLLAMA_HOST") == "" {
		os.Setenv("OLLAMA_HOST", "http://192.168.1.131:11434")
	}

	embedder := llm.NewOllamaClient()
	store := memory.NewChromaClient(embedder)

	cwd, _ := os.Getwd()
	fmt.Printf("Indexing local files in: %s\n", cwd)

	var docs []memory.Document
	projectID := "sovereign-firm-core"

	err := filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" || info.Name() == "chroma-data" {
				return filepath.SkipDir
			}
			return nil
		}

		// Filter extensions
		ext := filepath.Ext(path)
		switch ext {
		case ".go", ".js", ".ts", ".tsx", ".md", ".json", ".yaml":
			// accessible
		default:
			return nil
		}

		// Read content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(cwd, path)
		docs = append(docs, memory.Document{
			ID:      fmt.Sprintf("%s:%s", projectID, relPath),
			Content: string(content),
			Metadata: map[string]interface{}{
				"source":   relPath,
				"project":  projectID,
				"language": ext,
			},
		})
		fmt.Printf("Prepared: %s\n", relPath)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sending %d documents to Chroma...\n", len(docs))
	if len(docs) > 0 {
		if err := store.AddDocuments(context.Background(), docs); err != nil {
			log.Fatalf("❌ Failed to index: %v", err)
		}
	}
	fmt.Println("✅ Seeding Complete")
}
