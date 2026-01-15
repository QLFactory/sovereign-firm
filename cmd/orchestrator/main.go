package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/qlfactory/sovereign-firm/pkg/firm/workflows"
	"go.temporal.io/sdk/client"
)

var temporalClient client.Client

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	var err error
	temporalHost := os.Getenv("TEMPORAL_HOST")
	if temporalHost == "" {
		temporalHost = "localhost:7233"
	}

	temporalClient, err = client.Dial(client.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer temporalClient.Close()

	http.HandleFunc("/api/pods", handlePods)       // POST start
	http.HandleFunc("/api/pods/", handlePodAction) // POST message, GET status
	http.HandleFunc("/health", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Orchestrator listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handlePods(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		startPodHandler(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handlePodAction(w http.ResponseWriter, r *http.Request) {
	// /api/pods/{id} or /api/pods/{id}/message
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	workflowID := pathParts[3]

	if len(pathParts) == 4 && r.Method == http.MethodGet {
		getPodStatusHandler(w, r, workflowID)
		return
	}
	if len(pathParts) == 5 && pathParts[4] == "message" && r.Method == http.MethodPost {
		sendMessageHandler(w, r, workflowID)
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

type StartPodRequest struct {
	ProjectName string `json:"project_name"`
}

func startPodHandler(w http.ResponseWriter, r *http.Request) {
	var req StartPodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        "project_" + req.ProjectName,
		TaskQueue: "sovereign-firm-tasks",
	}

	we, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflows.ProjectLifecycle, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"workflow_id": we.GetID(),
		"run_id":      we.GetRunID(),
	})
}

type MessageRequest struct {
	Message string `json:"message"`
}

func sendMessageHandler(w http.ResponseWriter, r *http.Request, workflowID string) {
	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Send Signal to Temporal
	signal := workflows.UserMessageSignal{Message: req.Message}
	err := temporalClient.SignalWorkflow(context.Background(), workflowID, "", "USER_MESSAGE", signal)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to signal workflow: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"sent"}`))
}

func getPodStatusHandler(w http.ResponseWriter, r *http.Request, workflowID string) {
	// Query the workflow state

	resp, err := temporalClient.QueryWorkflow(context.Background(), workflowID, "", "get_state")
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed (workflow running?): %v", err), http.StatusInternalServerError)
		return
	}

	var state workflows.ProjectState
	if err := resp.Get(&state); err != nil {
		http.Error(w, "Failed to decode state", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(state)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
