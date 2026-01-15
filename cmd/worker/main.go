package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/joho/godotenv"
	"github.com/qlfactory/sovereign-firm/pkg/firm/activities"
	"github.com/qlfactory/sovereign-firm/pkg/firm/workflows"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Verify LLM Connectivity (Fail Fast)
	llmClient := llm.NewOllamaClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := llmClient.Ping(ctx); err != nil {
		log.Fatalf("CRITICAL: Cannot connect to Ollama at %s: %v", llmClient.BaseURL, err)
	}
	log.Println("✅ Connected to Ollama")

	temporalHost := os.Getenv("TEMPORAL_HOST")
	if temporalHost == "" {
		temporalHost = "localhost:7233"
	}

	c, err := client.Dial(client.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "sovereign-firm-tasks", worker.Options{})

	// Register Workflow
	w.RegisterWorkflow(workflows.ProjectLifecycle)

	// Register Activities
	pmAgent := activities.NewPMAgent()
	w.RegisterActivityWithOptions(pmAgent.Chat, activity.RegisterOptions{Name: "PMAgentChat"})

	indexer := activities.NewIndexer()
	w.RegisterActivityWithOptions(indexer.IndexRepo, activity.RegisterOptions{Name: "IndexRepo"})

	devAgent := activities.NewDevAgent()
	w.RegisterActivityWithOptions(devAgent.GenerateCode, activity.RegisterOptions{Name: "DevAgentGenerate"})
	w.RegisterActivityWithOptions(devAgent.RefineCode, activity.RegisterOptions{Name: "DevAgentRefine"})

	log.Println("Worker started...")
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
