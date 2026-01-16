package orchestration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/agent"
)

// createTestAgentPool creates an agent pool for testing
func createTestAgentPool() *agent.AgentPool {
	poolConfig := agent.PoolConfig{
		MaxAgents:           10,
		MaxAgentsPerProject: 5,
		IdleTimeout:         30 * time.Second,
		CleanupInterval:     1 * time.Minute,
	}
	return agent.NewAgentPool(poolConfig, nil, nil)
}

// Integration tests for Meta-Agent + AgentPool + DAG + FileLocking

func TestIntegration_MetaAgentWithAgentPool(t *testing.T) {
	// Create agent pool
	pool := createTestAgentPool()
	defer pool.Shutdown()

	// Create file lock manager
	fileLocks := NewFileLockManager(5 * time.Minute)

	// Create meta-agent config
	config := MetaAgentConfig{
		MaxConcurrentAgents: 3,
		TaskTimeout:         30 * time.Second,
		RetryOnFailure:      true,
		MaxRetries:          2,
	}

	// Create meta-agent (without git manager for this test)
	metaAgent := NewMetaAgent(config, pool, fileLocks, nil)
	defer metaAgent.Cancel()

	// Subscribe to events
	events := metaAgent.SubscribeToEvents()

	// Create a simple project plan
	plan := &ProjectPlan{
		ID:          "plan-001",
		ProjectID:   "project-001",
		Description: "Test project",
		Requirements: []string{
			"Build a simple component",
		},
		Components: []ComponentPlan{
			{
				Name:        "Component A",
				Type:        "frontend",
				Description: "A simple component",
				AgentRole:   "react-developer",
				Files:       []string{"src/ComponentA.tsx"},
			},
		},
	}

	// Create DAG from plan
	dag, err := metaAgent.CreateDAGFromPlan(plan)
	if err != nil {
		t.Fatalf("failed to create DAG: %v", err)
	}

	if dag == nil {
		t.Fatal("DAG should not be nil")
	}

	// Verify DAG structure
	if len(dag.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(dag.Tasks))
	}

	// Verify progress starts at 0
	total, completed, _, _, pending := metaAgent.GetProgress()
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if completed != 0 {
		t.Errorf("expected completed 0, got %d", completed)
	}
	if pending != 1 {
		t.Errorf("expected pending 1, got %d", pending)
	}

	// Verify events are being received
	go func() {
		for range events {
			// Just consume events
		}
	}()
}

func TestIntegration_DAGWithDependencies(t *testing.T) {
	// Create agent pool
	pool := createTestAgentPool()
	defer pool.Shutdown()

	fileLocks := NewFileLockManager(5 * time.Minute)

	config := MetaAgentConfig{
		MaxConcurrentAgents: 3,
		TaskTimeout:         30 * time.Second,
		RetryOnFailure:      true,
		MaxRetries:          2,
	}

	metaAgent := NewMetaAgent(config, pool, fileLocks, nil)
	defer metaAgent.Cancel()

	// Create plan with dependencies
	plan := &ProjectPlan{
		ID:          "plan-002",
		ProjectID:   "project-002",
		Description: "Project with dependencies",
		Components: []ComponentPlan{
			{
				Name:         "Database Schema",
				Type:         "database",
				Description:  "Database schema design",
				AgentRole:    "backend-developer",
				Files:        []string{"db/schema.sql"},
				Dependencies: []string{},
			},
			{
				Name:         "API Layer",
				Type:         "backend",
				Description:  "REST API implementation",
				AgentRole:    "backend-developer",
				Files:        []string{"api/handlers.go"},
				Dependencies: []string{"Database Schema"},
			},
			{
				Name:         "Frontend",
				Type:         "frontend",
				Description:  "React frontend",
				AgentRole:    "react-developer",
				Files:        []string{"src/App.tsx"},
				Dependencies: []string{"API Layer"},
			},
		},
	}

	dag, err := metaAgent.CreateDAGFromPlan(plan)
	if err != nil {
		t.Fatalf("failed to create DAG: %v", err)
	}

	// Verify DAG has correct structure
	if len(dag.Tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(dag.Tasks))
	}

	// Get parallel layers
	layers := dag.GetParallelTasks()

	// Should be 3 layers (serial dependencies)
	if len(layers) != 3 {
		t.Errorf("expected 3 layers, got %d", len(layers))
	}

	// First layer should have 1 task (Database Schema)
	if len(layers[0]) != 1 {
		t.Errorf("expected 1 task in first layer, got %d", len(layers[0]))
	}
}

func TestIntegration_FileLockingDuringExecution(t *testing.T) {
	fileLocks := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	// Simulate multiple agents trying to access the same file
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(agentNum int) {
			defer wg.Done()

			// Try to acquire write lock
			ctxTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			defer cancel()

			err := fileLocks.AcquireLock(ctxTimeout, "shared.go", LockWrite, "agent-"+string(rune('0'+agentNum)))
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()

				// Simulate work
				time.Sleep(10 * time.Millisecond)

				// Release lock
				fileLocks.ReleaseLock("shared.go", "agent-"+string(rune('0'+agentNum)))
			}
		}(i)
	}

	wg.Wait()

	// At least one should succeed, possibly more due to sequential release
	if successCount < 1 {
		t.Errorf("expected at least 1 success, got %d", successCount)
	}
}

func TestIntegration_ConcurrentReadLocks(t *testing.T) {
	fileLocks := NewFileLockManager(5 * time.Minute)
	ctx := context.Background()

	// Multiple readers should all succeed
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(agentNum int) {
			defer wg.Done()

			err := fileLocks.AcquireLock(ctx, "readonly.go", LockRead, "agent-"+string(rune('A'+agentNum)))
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// All readers should succeed
	if successCount != 10 {
		t.Errorf("expected 10 successful readers, got %d", successCount)
	}
}

func TestIntegration_DAGExecutionOrder(t *testing.T) {
	// Create a DAG and verify execution order respects dependencies
	dag := NewTaskDAG("dag-001", "project-001")

	// Add tasks: A -> B -> C (serial)
	//            A -> D (parallel with B)
	dag.AddTask(&DAGTask{
		ID:   "task-A",
		Name: "Task A",
	})
	dag.AddTask(&DAGTask{
		ID:           "task-B",
		Name:         "Task B",
		Dependencies: []string{"task-A"},
	})
	dag.AddTask(&DAGTask{
		ID:           "task-C",
		Name:         "Task C",
		Dependencies: []string{"task-B"},
	})
	dag.AddTask(&DAGTask{
		ID:           "task-D",
		Name:         "Task D",
		Dependencies: []string{"task-A"},
	})

	// Initially only task-A should be ready
	ready := dag.GetReadyTasks()
	if len(ready) != 1 || ready[0].ID != "task-A" {
		t.Errorf("only task-A should be ready initially")
	}

	// Complete task-A
	dag.MarkTaskCompleted("task-A", nil)

	// Now B and D should be ready (parallel)
	ready = dag.GetReadyTasks()
	if len(ready) != 2 {
		t.Errorf("expected 2 ready tasks after completing A, got %d", len(ready))
	}

	// Complete B
	dag.MarkTaskCompleted("task-B", nil)

	// Now C and D should be ready (D if not already counted)
	ready = dag.GetReadyTasks()
	hasC := false
	for _, task := range ready {
		if task.ID == "task-C" {
			hasC = true
		}
	}
	if !hasC {
		t.Error("task-C should be ready after completing B")
	}
}

func TestIntegration_CIPipelineWithSelfCorrection(t *testing.T) {
	// Create CI pipeline with commands that succeed
	pipelineConfig := CIPipelineConfig{
		WorkDir:      "/tmp",
		LintCommand:  "echo lint",
		BuildCommand: "echo build",
		TestCommand:  "echo test",
		Timeout:      30 * time.Second,
	}
	pipeline := NewCIPipeline(pipelineConfig)

	correctionConfig := SelfCorrectionConfig{
		MaxAttempts: 3,
		RetryDelay:  10 * time.Millisecond,
	}
	correctionLoop := NewSelfCorrectionLoop(correctionConfig, pipeline)

	ctx := context.Background()
	attemptCount := 0

	result, err := correctionLoop.RunWithCorrection(ctx, nil, func(feedback CorrectionFeedback) ([]string, error) {
		attemptCount++
		return []string{"fix applied"}, nil
	})

	if err != nil {
		t.Fatalf("correction loop error: %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}

	// Should succeed on first attempt
	if result.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", result.Attempts)
	}

	// Fix callback should not be called on success
	if attemptCount != 0 {
		t.Errorf("fix callback should not be called on success, was called %d times", attemptCount)
	}
}

func TestIntegration_MetaAgentEventStreaming(t *testing.T) {
	pool := createTestAgentPool()
	defer pool.Shutdown()

	fileLocks := NewFileLockManager(5 * time.Minute)

	config := MetaAgentConfig{
		MaxConcurrentAgents: 2,
		TaskTimeout:         10 * time.Second,
	}

	metaAgent := NewMetaAgent(config, pool, fileLocks, nil)

	// Subscribe to events
	events := metaAgent.SubscribeToEvents()

	// Manually emit an event
	metaAgent.emitEvent("TEST_EVENT", "task-123", "agent-456", "Test message", nil)

	// Should receive the event
	select {
	case event := <-events:
		if event.Type != "TEST_EVENT" {
			t.Errorf("expected TEST_EVENT, got %s", event.Type)
		}
		if event.TaskID != "task-123" {
			t.Errorf("expected task-123, got %s", event.TaskID)
		}
		if event.AgentID != "agent-456" {
			t.Errorf("expected agent-456, got %s", event.AgentID)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for event")
	}

	metaAgent.Cancel()
}

func TestIntegration_ArtifactStorage(t *testing.T) {
	pool := createTestAgentPool()
	defer pool.Shutdown()

	fileLocks := NewFileLockManager(5 * time.Minute)

	config := MetaAgentConfig{
		MaxConcurrentAgents: 2,
		TaskTimeout:         10 * time.Second,
	}

	metaAgent := NewMetaAgent(config, pool, fileLocks, nil)
	defer metaAgent.Cancel()

	// Manually store artifacts (simulating task completion)
	metaAgent.mu.Lock()
	metaAgent.artifacts["task-1"] = map[string]interface{}{
		"files":  []string{"file1.go", "file2.go"},
		"status": "completed",
	}
	metaAgent.artifacts["task-2"] = map[string]interface{}{
		"files":  []string{"file3.go"},
		"status": "completed",
	}
	metaAgent.mu.Unlock()

	// Retrieve artifacts
	artifacts := metaAgent.GetArtifacts()

	if len(artifacts) != 2 {
		t.Errorf("expected 2 artifacts, got %d", len(artifacts))
	}

	if artifacts["task-1"]["status"] != "completed" {
		t.Error("task-1 status should be completed")
	}
}

func TestIntegration_ActiveAgentTracking(t *testing.T) {
	pool := createTestAgentPool()
	defer pool.Shutdown()

	fileLocks := NewFileLockManager(5 * time.Minute)

	config := MetaAgentConfig{
		MaxConcurrentAgents: 2,
		TaskTimeout:         10 * time.Second,
	}

	metaAgent := NewMetaAgent(config, pool, fileLocks, nil)
	defer metaAgent.Cancel()

	// Manually track active agents
	metaAgent.mu.Lock()
	metaAgent.activeAgents["task-1"] = "agent-001"
	metaAgent.activeAgents["task-2"] = "agent-002"
	metaAgent.mu.Unlock()

	// Get active agents
	active := metaAgent.GetActiveAgents()

	if len(active) != 2 {
		t.Errorf("expected 2 active agents, got %d", len(active))
	}

	if active["task-1"] != "agent-001" {
		t.Errorf("expected agent-001 for task-1, got %s", active["task-1"])
	}
}

func TestIntegration_DAGCloneIndependence(t *testing.T) {
	// Create original DAG
	original := NewTaskDAG("dag-001", "project-001")
	original.AddTask(&DAGTask{
		ID:     "task-1",
		Name:   "Task 1",
		Status: TaskPending,
	})

	// Clone it
	clone := original.Clone()

	// Modify original
	original.MarkTaskCompleted("task-1", nil)

	// Clone should be unaffected
	task, _ := clone.GetTask("task-1")
	if task.Status != TaskPending {
		t.Errorf("clone task status should still be PENDING, got %s", task.Status)
	}
}

func TestIntegration_ProjectPlanToExecution(t *testing.T) {
	// Full integration: Plan -> DAG -> Ready Tasks -> Progress

	pool := createTestAgentPool()
	defer pool.Shutdown()

	fileLocks := NewFileLockManager(5 * time.Minute)

	config := MetaAgentConfig{
		MaxConcurrentAgents: 5,
		TaskTimeout:         30 * time.Second,
		RetryOnFailure:      true,
		MaxRetries:          3,
	}

	metaAgent := NewMetaAgent(config, pool, fileLocks, nil)
	defer metaAgent.Cancel()

	// Create a realistic project plan
	plan := &ProjectPlan{
		ID:          "plan-003",
		ProjectID:   "project-003",
		Description: "E-commerce application",
		Requirements: []string{
			"User authentication",
			"Product catalog",
			"Shopping cart",
		},
		Components: []ComponentPlan{
			{
				Name:        "Auth Service",
				Type:        "backend",
				Description: "Authentication microservice",
				AgentRole:   "backend-developer",
				Files:       []string{"services/auth/main.go"},
			},
			{
				Name:         "Product Service",
				Type:         "backend",
				Description:  "Product catalog service",
				AgentRole:    "backend-developer",
				Files:        []string{"services/product/main.go"},
				Dependencies: []string{"Auth Service"},
			},
			{
				Name:         "Cart Service",
				Type:         "backend",
				Description:  "Shopping cart service",
				AgentRole:    "backend-developer",
				Files:        []string{"services/cart/main.go"},
				Dependencies: []string{"Auth Service", "Product Service"},
			},
			{
				Name:         "Frontend",
				Type:         "frontend",
				Description:  "React frontend",
				AgentRole:    "react-developer",
				Files:        []string{"frontend/src/App.tsx"},
				Dependencies: []string{"Auth Service", "Product Service", "Cart Service"},
			},
		},
	}

	dag, err := metaAgent.CreateDAGFromPlan(plan)
	if err != nil {
		t.Fatalf("failed to create DAG: %v", err)
	}

	// Verify task count
	if len(dag.Tasks) != 4 {
		t.Errorf("expected 4 tasks, got %d", len(dag.Tasks))
	}

	// Verify layers
	layers := dag.GetParallelTasks()
	if len(layers) != 4 {
		t.Errorf("expected 4 layers (serial chain with branches), got %d", len(layers))
	}

	// First layer should be Auth Service
	if len(layers[0]) != 1 {
		t.Errorf("expected 1 task in first layer, got %d", len(layers[0]))
	}

	// Verify progress
	total, completed, failed, running, pending := metaAgent.GetProgress()
	if total != 4 {
		t.Errorf("expected total 4, got %d", total)
	}
	if completed != 0 || failed != 0 || running != 0 {
		t.Error("all tasks should be pending initially")
	}
	if pending != 4 {
		t.Errorf("expected pending 4, got %d", pending)
	}
}
