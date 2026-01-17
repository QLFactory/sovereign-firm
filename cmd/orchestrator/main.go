package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/qlfactory/sovereign-firm/pkg/firm/workflows"
	"github.com/qlfactory/sovereign-firm/pkg/streaming"
	"go.temporal.io/sdk/client"
)

var (
	temporalClient client.Client
	streamHub      *streaming.Hub
)

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

	// Initialize streaming hub
	streamHub = streaming.NewHub()
	go streamHub.Run()
	log.Println("Streaming hub started")

	// HTTP routes
	http.HandleFunc("/api/pods", handlePods)       // POST start
	http.HandleFunc("/api/pods/", handlePodAction) // POST message, GET status, WS stream
	http.HandleFunc("/health", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Orchestrator listening on :%s", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/api/pods/{id}/stream", port)
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
	// /api/pods/{id} or /api/pods/{id}/message or /api/pods/{id}/stream
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	workflowID := pathParts[3]

	// GET /api/pods/{id} - Get status
	if len(pathParts) == 4 && r.Method == http.MethodGet {
		getPodStatusHandler(w, r, workflowID)
		return
	}

	// POST /api/pods/{id}/message - Send message
	if len(pathParts) == 5 && pathParts[4] == "message" && r.Method == http.MethodPost {
		sendMessageHandler(w, r, workflowID)
		return
	}

	// GET /api/pods/{id}/stream - WebSocket stream
	if len(pathParts) == 5 && pathParts[4] == "stream" {
		handleStreamWebSocket(w, r, workflowID)
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

// StartPodRequest matches ConsultancyConfig for full-stack project generation
type StartPodRequest struct {
	ProjectName       string `json:"project_name"`
	ClientID          string `json:"client_id,omitempty"`
	InitialMessage    string `json:"initial_message,omitempty"`
	EnableFullStack   bool   `json:"enable_full_stack,omitempty"`
	EnableDeployment  bool   `json:"enable_deployment,omitempty"`
	EnableSRE         bool   `json:"enable_sre,omitempty"`
	PreferredFrontend string `json:"preferred_frontend,omitempty"`
	PreferredBackend  string `json:"preferred_backend,omitempty"`
	PreferredDatabase string `json:"preferred_database,omitempty"`
	PreferredCloud    string `json:"preferred_cloud,omitempty"`
	MaxCodeAttempts   int    `json:"max_code_attempts,omitempty"`
	MaxTestAttempts   int    `json:"max_test_attempts,omitempty"`
}

func startPodHandler(w http.ResponseWriter, r *http.Request) {
	var req StartPodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Convert request to ConsultancyConfig
	config := workflows.ConsultancyConfig{
		ProjectName:       req.ProjectName,
		ClientID:          req.ClientID,
		InitialMessage:    req.InitialMessage,
		EnableFullStack:   req.EnableFullStack,
		EnableDeployment:  req.EnableDeployment,
		EnableSRE:         req.EnableSRE,
		PreferredFrontend: req.PreferredFrontend,
		PreferredBackend:  req.PreferredBackend,
		PreferredDatabase: req.PreferredDatabase,
		PreferredCloud:    req.PreferredCloud,
		MaxCodeAttempts:   req.MaxCodeAttempts,
		MaxTestAttempts:   req.MaxTestAttempts,
	}

	// Set defaults if not provided
	if config.ClientID == "" {
		config.ClientID = "default-client"
	}
	if config.MaxCodeAttempts <= 0 {
		config.MaxCodeAttempts = 3
	}
	if config.MaxTestAttempts <= 0 {
		config.MaxTestAttempts = 3
	}

	workflowID := fmt.Sprintf("consultancy_%s_%d", strings.ReplaceAll(req.ProjectName, " ", "_"), time.Now().Unix())

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "sovereign-firm-tasks",
	}

	// Use ConsultancyWorkflow for full artifact generation
	we, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflows.ConsultancyWorkflow, config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast phase change event
	streamHub.BroadcastPhaseChange(workflowID, "", "INTAKE", "Consultancy workflow started")

	json.NewEncoder(w).Encode(map[string]string{
		"workflow_id": workflowID,
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

	// Broadcast user message to connected clients
	streamHub.BroadcastChatMessage(workflowID, "user", "", req.Message)

	// Send Signal to Temporal
	signal := workflows.UserMessageSignal{Message: req.Message}
	err := temporalClient.SignalWorkflow(context.Background(), workflowID, "", "USER_MESSAGE", signal)
	if err != nil {
		streamHub.BroadcastError(workflowID, "SIGNAL_FAILED", "Failed to send message", err.Error())
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

	// Use ConsultancyState for full artifact access
	var state workflows.ConsultancyState
	if err := resp.Get(&state); err != nil {
		http.Error(w, "Failed to decode state", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(state)
}

func handleStreamWebSocket(w http.ResponseWriter, r *http.Request, workflowID string) {
	log.Printf("WebSocket connection request for workflow: %s", workflowID)
	streamHub.ServeWs(w, r, workflowID)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","streaming":true}`))
}

// GetStreamHub returns the global stream hub (for use by activities)
func GetStreamHub() *streaming.Hub {
	return streamHub
}
