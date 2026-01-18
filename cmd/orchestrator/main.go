package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/qlfactory/sovereign-firm/pkg/api"
	"github.com/qlfactory/sovereign-firm/pkg/auth"
	"github.com/qlfactory/sovereign-firm/pkg/cache"
	"github.com/qlfactory/sovereign-firm/pkg/database"
	"github.com/qlfactory/sovereign-firm/pkg/firm/workflows"
	"github.com/qlfactory/sovereign-firm/pkg/storage"
	"github.com/qlfactory/sovereign-firm/pkg/streaming"
	"go.temporal.io/sdk/client"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize components
	db, temporalClient, streamHub, redisCache, minioStorage, jwtManager := initComponents(ctx)
	if db != nil {
		defer db.Close()
	}
	defer temporalClient.Close()
	if redisCache != nil {
		defer redisCache.Close()
	}

	// Create caches and stores
	var projectCache *cache.ProjectCache
	var artifactStore *storage.ArtifactStore
	if redisCache != nil {
		projectCache = cache.NewProjectCache(redisCache)
	}
	if minioStorage != nil {
		artifactStore = storage.NewArtifactStore(minioStorage)
	}

	// Create handlers
	var dbPool interface{}
	if db != nil {
		dbPool = db.Pool
	}
	apiHandler := api.NewHandler(dbPool, temporalClient, streamHub, projectCache, artifactStore)
	authHandler := auth.NewHandler(dbPool, jwtManager)
	authMiddleware := auth.NewMiddleware(jwtManager)

	// Build router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting (only if Redis available)
	if redisCache != nil {
		rateLimiter := cache.NewRateLimiter(redisCache, 100, time.Minute)
		r.Use(rateLimiter.Middleware(cache.ByIP))
	}

	// Health endpoints (no auth)
	r.Get("/health", apiHandler.HealthCheck)
	r.Get("/health/ready", apiHandler.ReadinessCheck)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes (no auth required)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.With(authMiddleware.OptionalAuth).Post("/logout", authHandler.Logout)
			r.With(authMiddleware.Authenticate).Get("/me", authHandler.Me)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			// Projects
			r.Route("/projects", func(r chi.Router) {
				r.Get("/", apiHandler.ListProjects)
				r.Post("/", apiHandler.CreateProject)

				// Brownfield import
				r.Post("/import", apiHandler.ImportBrownfield)

				r.Route("/{projectId}", func(r chi.Router) {
					r.Get("/", apiHandler.GetProject)
					r.Delete("/", apiHandler.ArchiveProject)
					r.Get("/state", apiHandler.GetProjectState)
					r.Post("/message", apiHandler.SendMessage)

					// Import status
					r.Get("/import/status", apiHandler.GetImportStatus)

					// Artifacts
					r.Get("/artifacts", apiHandler.ListArtifacts)
					r.Get("/artifacts/{artifactId}", apiHandler.GetArtifact)
				})
			})
		})

		// WebSocket (auth via query param for browser compatibility)
		r.Get("/projects/{projectId}/stream", apiHandler.StreamProject)
	})

	// Legacy endpoints for backward compatibility
	r.Route("/api/pods", func(r chi.Router) {
		r.Post("/", legacyCreatePod(temporalClient, streamHub))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", legacyGetPodStatus(temporalClient))
			r.Post("/message", legacySendMessage(temporalClient, streamHub))
			r.Get("/stream", legacyStream(streamHub))
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
		cancel()
	}()

	log.Printf("Orchestrator listening on :%s", port)
	log.Printf("API: http://localhost:%s/api", port)
	log.Printf("Health: http://localhost:%s/health", port)
	log.Printf("Legacy API: http://localhost:%s/api/pods (backward compatible)", port)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}

func initComponents(ctx context.Context) (*database.DB, client.Client, *streaming.Hub, *cache.Cache, *storage.Storage, *auth.JWTManager) {
	// Initialize database
	db, err := database.New(ctx, database.DefaultConfig())
	if err != nil {
		log.Printf("Warning: Database not available: %v", err)
		db = nil
	}

	// Initialize Temporal client
	temporalHost := os.Getenv("TEMPORAL_HOST")
	if temporalHost == "" {
		temporalHost = "localhost:7233"
	}
	temporalClient, err := client.Dial(client.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		log.Fatalf("Unable to create Temporal client: %v", err)
	}

	// Initialize streaming hub
	streamHub := streaming.NewHub()
	go streamHub.Run()
	log.Println("Streaming hub started")

	// Initialize Redis cache
	var redisCache *cache.Cache
	redisCache, err = cache.New(ctx, cache.DefaultConfig())
	if err != nil {
		log.Printf("Warning: Redis not available: %v", err)
		redisCache = nil
	}

	// Initialize MinIO storage
	var minioStorage *storage.Storage
	minioStorage, err = storage.New(ctx, storage.DefaultConfig())
	if err != nil {
		log.Printf("Warning: MinIO not available: %v", err)
		minioStorage = nil
	}

	// Initialize JWT manager
	jwtManager, err := auth.NewJWTManager(auth.DefaultJWTConfig())
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	return db, temporalClient, streamHub, redisCache, minioStorage, jwtManager
}

// Legacy handlers for backward compatibility with existing frontend

func legacyCreatePod(temporalClient client.Client, streamHub *streaming.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
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

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

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

		we, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflows.ConsultancyWorkflow, config)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to start workflow: %v", err), http.StatusInternalServerError)
			return
		}

		streamHub.BroadcastPhaseChange(workflowID, "", "INTAKE", "Consultancy workflow started")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"workflow_id": workflowID,
			"run_id":      we.GetRunID(),
		})
	}
}

func legacyGetPodStatus(temporalClient client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID := chi.URLParam(r, "id")

		resp, err := temporalClient.QueryWorkflow(context.Background(), workflowID, "", "get_state")
		if err != nil {
			http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
			return
		}

		var state workflows.ConsultancyState
		if err := resp.Get(&state); err != nil {
			http.Error(w, "Failed to decode state", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(state)
	}
}

func legacySendMessage(temporalClient client.Client, streamHub *streaming.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID := chi.URLParam(r, "id")

		var req struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		streamHub.BroadcastChatMessage(workflowID, "user", "", req.Message)

		signal := workflows.UserMessageSignal{Message: req.Message}
		err := temporalClient.SignalWorkflow(context.Background(), workflowID, "", "USER_MESSAGE", signal)
		if err != nil {
			streamHub.BroadcastError(workflowID, "SIGNAL_FAILED", "Failed to send message", err.Error())
			http.Error(w, fmt.Sprintf("Failed to signal workflow: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"sent"}`))
	}
}

func legacyStream(streamHub *streaming.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID := chi.URLParam(r, "id")
		log.Printf("WebSocket connection request for workflow: %s", workflowID)
		streamHub.ServeWs(w, r, workflowID)
	}
}
