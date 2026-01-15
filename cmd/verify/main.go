package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/memory"
	"go.temporal.io/sdk/client"
)

func main() {
	fmt.Println("🔍 Verifying Sovereign Infrastructure...")

	// 1. Check Temporal
	checkTemporal()

	// 2. Check Ollama
	checkOllama()

	// 3. Check Chroma
	checkChroma()

	fmt.Println("✅ Verification Complete")
}

func checkTemporal() {
	host := os.Getenv("TEMPORAL_HOST")
	if host == "" {
		host = "localhost:7233"
	}
	fmt.Printf("1. Checking Temporal at %s... ", host)

	c, err := client.Dial(client.Options{
		HostPort: host,
	})
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	defer c.Close()
	fmt.Println("✅ OK")
}

func checkOllama() {
	fmt.Print("2. Checking Ollama... ")
	client := llm.NewOllamaClient()

	// Test Connection/Generation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Generate(ctx, llm.GenerateRequest{
		Prompt: "Hello",
		Model:  client.Model, // Uses env OLLAMA_MODEL or default llama3
	})
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		fmt.Println("   --> Ensure 'ollama serve' is running and you have pulled the model (e.g. 'ollama pull llama3')")
		return
	}
	fmt.Printf("✅ OK (Model: %s, Response: %q)\n", client.Model, resp.Response)
}

func checkChroma() {
	fmt.Print("3. Checking Chroma... ")
	// We need an embedder for the client, reuse Ollama
	embedder := llm.NewOllamaClient()
	c := memory.NewChromaClient(embedder)

	// Just check if we can create/get collection
	// This tests connection to Chroma
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.SimilaritySearch(ctx, []float32{0.1, 0.2}, 1)

	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
	} else {
		fmt.Println("✅ OK")
	}
}
