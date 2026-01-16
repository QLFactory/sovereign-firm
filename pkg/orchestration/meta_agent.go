package orchestration

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/qlfactory/sovereign-firm/pkg/agent"
)

// ProjectPlan represents a high-level plan for a project
type ProjectPlan struct {
	ID           string            `json:"id"`
	ProjectID    string            `json:"project_id"`
	Description  string            `json:"description"`
	Requirements []string          `json:"requirements"`
	Components   []ComponentPlan   `json:"components"`
	Constraints  map[string]string `json:"constraints"`
	CreatedAt    time.Time         `json:"created_at"`
}

// ComponentPlan represents a component to be built
type ComponentPlan struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`        // "frontend", "backend", "api", "database", etc.
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"` // Other component names
	Files        []string `json:"files"`        // Expected output files
	AgentRole    string   `json:"agent_role"`   // Role needed to build this
}

// MetaAgentConfig configures the meta-agent
type MetaAgentConfig struct {
	MaxConcurrentAgents int           `json:"max_concurrent_agents"`
	TaskTimeout         time.Duration `json:"task_timeout"`
	RetryOnFailure      bool          `json:"retry_on_failure"`
	MaxRetries          int           `json:"max_retries"`
}

// DefaultMetaAgentConfig returns sensible defaults
func DefaultMetaAgentConfig() MetaAgentConfig {
	return MetaAgentConfig{
		MaxConcurrentAgents: 5,
		TaskTimeout:         10 * time.Minute,
		RetryOnFailure:      true,
		MaxRetries:          3,
	}
}

// ExecutionEvent represents an event during execution
type ExecutionEvent struct {
	Type      string      `json:"type"`
	TaskID    string      `json:"task_id,omitempty"`
	AgentID   string      `json:"agent_id,omitempty"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// MetaAgent orchestrates multiple agents to complete a project
type MetaAgent struct {
	config     MetaAgentConfig
	dag        *TaskDAG
	agentPool  *agent.AgentPool
	fileLocks  *FileLockManager
	gitManager *GitBranchManager

	// Execution state
	activeAgents map[string]string // taskID -> agentID
	artifacts    map[string]map[string]interface{} // taskID -> outputs

	// Event streaming
	events     chan ExecutionEvent
	eventSubs  []chan ExecutionEvent

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	wg     sync.WaitGroup
}

// NewMetaAgent creates a new meta-agent orchestrator
func NewMetaAgent(
	config MetaAgentConfig,
	agentPool *agent.AgentPool,
	fileLocks *FileLockManager,
	gitManager *GitBranchManager,
) *MetaAgent {
	ctx, cancel := context.WithCancel(context.Background())

	ma := &MetaAgent{
		config:       config,
		agentPool:    agentPool,
		fileLocks:    fileLocks,
		gitManager:   gitManager,
		activeAgents: make(map[string]string),
		artifacts:    make(map[string]map[string]interface{}),
		events:       make(chan ExecutionEvent, 100),
		eventSubs:    make([]chan ExecutionEvent, 0),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start event dispatcher
	go ma.dispatchEvents()

	return ma
}

// CreateDAGFromPlan creates a task DAG from a project plan
func (m *MetaAgent) CreateDAGFromPlan(plan *ProjectPlan) (*TaskDAG, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	dag := NewTaskDAG(uuid.New().String(), plan.ProjectID)

	// Create tasks for each component
	componentTasks := make(map[string]string) // component name -> task ID

	for _, comp := range plan.Components {
		taskID := uuid.New().String()
		componentTasks[comp.Name] = taskID

		// Map dependencies to task IDs
		var deps []string
		for _, depName := range comp.Dependencies {
			if depTaskID, exists := componentTasks[depName]; exists {
				deps = append(deps, depTaskID)
			}
		}

		task := &DAGTask{
			ID:           taskID,
			Name:         fmt.Sprintf("Build %s", comp.Name),
			Description:  comp.Description,
			Type:         comp.Type,
			AgentRole:    comp.AgentRole,
			Dependencies: deps,
			Produces:     comp.Files,
			Input: map[string]interface{}{
				"component":    comp.Name,
				"requirements": plan.Requirements,
			},
			Priority: PriorityNormal,
		}

		if err := dag.AddTask(task); err != nil {
			return nil, fmt.Errorf("failed to add task for %s: %w", comp.Name, err)
		}
	}

	// Validate the DAG
	if err := dag.ValidateDAG(); err != nil {
		return nil, fmt.Errorf("invalid DAG: %w", err)
	}

	m.dag = dag
	return dag, nil
}

// SetDAG sets the task DAG for execution
func (m *MetaAgent) SetDAG(dag *TaskDAG) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dag = dag
}

// Execute runs the DAG to completion
func (m *MetaAgent) Execute(ctx context.Context) error {
	m.mu.Lock()
	if m.dag == nil {
		m.mu.Unlock()
		return fmt.Errorf("no DAG set")
	}
	dag := m.dag
	m.mu.Unlock()

	m.emitEvent("EXECUTION_STARTED", "", "", "Starting DAG execution", nil)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.ctx.Done():
			return m.ctx.Err()
		default:
		}

		// Check if complete or failed
		if dag.IsComplete() {
			m.emitEvent("EXECUTION_COMPLETED", "", "", "All tasks completed", nil)
			return nil
		}

		if dag.HasFailed() {
			m.emitEvent("EXECUTION_FAILED", "", "", "DAG execution failed", nil)
			return fmt.Errorf("DAG execution failed")
		}

		// Get ready tasks
		readyTasks := dag.GetReadyTasks()
		if len(readyTasks) == 0 {
			// No ready tasks, wait for running tasks
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Start tasks up to max concurrent limit
		m.mu.RLock()
		activeCount := len(m.activeAgents)
		m.mu.RUnlock()

		for _, task := range readyTasks {
			if activeCount >= m.config.MaxConcurrentAgents {
				break
			}

			// Check file locks for consumed files
			canStart := true
			for _, file := range task.Consumes {
				if m.fileLocks.IsLocked(file) {
					canStart = false
					break
				}
			}

			if !canStart {
				continue
			}

			// Start task
			m.wg.Add(1)
			go m.executeTask(ctx, task)
			activeCount++
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// executeTask executes a single task with an agent
func (m *MetaAgent) executeTask(ctx context.Context, task *DAGTask) {
	defer m.wg.Done()

	taskCtx, cancel := context.WithTimeout(ctx, m.config.TaskTimeout)
	defer cancel()

	m.emitEvent("TASK_STARTED", task.ID, "", fmt.Sprintf("Starting task: %s", task.Name), task)

	// Acquire file locks for produced files
	for _, file := range task.Produces {
		err := m.fileLocks.AcquireLock(taskCtx, file, LockWrite, task.ID)
		if err != nil {
			m.handleTaskFailure(task, fmt.Sprintf("failed to acquire lock on %s: %v", file, err))
			return
		}
	}
	defer m.fileLocks.ReleaseMultiple(task.Produces, task.ID)

	// Spawn an agent for this task
	agentInstance, err := m.agentPool.SpawnAgent(
		taskCtx,
		fmt.Sprintf("Agent-%s", task.ID[:8]),
		task.AgentRole,
		m.dag.ProjectID,
		"gpt-4",
		[]string{task.AgentRole},
	)
	if err != nil {
		m.handleTaskFailure(task, fmt.Sprintf("failed to spawn agent: %v", err))
		return
	}

	// Track active agent
	m.mu.Lock()
	m.activeAgents[task.ID] = agentInstance.ID
	m.dag.MarkTaskRunning(task.ID, agentInstance.ID)
	m.mu.Unlock()

	m.emitEvent("AGENT_ASSIGNED", task.ID, agentInstance.ID, 
		fmt.Sprintf("Agent %s assigned to task", agentInstance.Name), nil)

	// Create branch for agent (if git manager available)
	var branchName string
	if m.gitManager != nil {
		branch, err := m.gitManager.CreateAgentBranch(agentInstance.ID, task.ID)
		if err != nil {
			log.Printf("Warning: failed to create branch for agent: %v", err)
		} else {
			branchName = branch.Name
			m.gitManager.CheckoutBranch(branchName)
		}
	}

	// Inject artifact context from dependencies
	m.injectArtifactContext(task)

	// Execute the task (simplified - in reality this would call the agent)
	// For now, we simulate the execution
	output, err := m.simulateAgentExecution(taskCtx, agentInstance, task)

	// Cleanup
	m.mu.Lock()
	delete(m.activeAgents, task.ID)
	m.mu.Unlock()

	if err != nil {
		m.handleTaskFailure(task, err.Error())
		
		// Retry if configured
		if m.config.RetryOnFailure && task.RetryCount < m.config.MaxRetries {
			m.dag.RetryTask(task.ID)
			m.emitEvent("TASK_RETRY", task.ID, agentInstance.ID,
				fmt.Sprintf("Retrying task (attempt %d)", task.RetryCount+1), nil)
		}
		return
	}

	// Merge branch if successful
	if m.gitManager != nil && branchName != "" {
		result, err := m.gitManager.MergeBranch(taskCtx, branchName)
		if err != nil || !result.Success {
			m.emitEvent("MERGE_FAILED", task.ID, agentInstance.ID,
				fmt.Sprintf("Branch merge failed: %v", err), result)
		} else {
			m.emitEvent("BRANCH_MERGED", task.ID, agentInstance.ID,
				fmt.Sprintf("Branch %s merged", branchName), result)
		}
	}

	// Store artifacts
	m.mu.Lock()
	m.artifacts[task.ID] = output
	m.dag.MarkTaskCompleted(task.ID, output)
	m.mu.Unlock()

	m.emitEvent("TASK_COMPLETED", task.ID, agentInstance.ID,
		fmt.Sprintf("Task completed: %s", task.Name), output)

	// Terminate agent
	m.agentPool.TerminateAgent(agentInstance.ID)
}

// simulateAgentExecution simulates agent work (replace with real execution)
func (m *MetaAgent) simulateAgentExecution(
	ctx context.Context,
	agentInstance *agent.AgentInstance,
	task *DAGTask,
) (map[string]interface{}, error) {
	// In a real implementation, this would:
	// 1. Create a prompt with task details and context
	// 2. Call the LLM through the agent
	// 3. Parse and validate the output
	// 4. Write files to the repository
	// 5. Return the artifacts

	// For now, return a simulated output
	time.Sleep(100 * time.Millisecond) // Simulate work

	return map[string]interface{}{
		"status":  "completed",
		"files":   task.Produces,
		"task_id": task.ID,
	}, nil
}

// injectArtifactContext adds outputs from dependencies to task input
func (m *MetaAgent) injectArtifactContext(task *DAGTask) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	depOutputs := make(map[string]interface{})
	for _, depID := range task.Dependencies {
		if output, exists := m.artifacts[depID]; exists {
			depTask, _ := m.dag.GetTask(depID)
			if depTask != nil {
				depOutputs[depTask.Name] = output
			}
		}
	}

	if len(depOutputs) > 0 {
		if task.Input == nil {
			task.Input = make(map[string]interface{})
		}
		task.Input["dependencies"] = depOutputs
	}
}

// handleTaskFailure handles a failed task
func (m *MetaAgent) handleTaskFailure(task *DAGTask, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.dag.MarkTaskFailed(task.ID, errMsg, true)
	m.emitEvent("TASK_FAILED", task.ID, "", fmt.Sprintf("Task failed: %s", errMsg), nil)
}

// emitEvent sends an event to subscribers
func (m *MetaAgent) emitEvent(eventType, taskID, agentID, message string, data interface{}) {
	event := ExecutionEvent{
		Type:      eventType,
		TaskID:    taskID,
		AgentID:   agentID,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case m.events <- event:
	default:
		// Channel full, drop event
		log.Printf("Event dropped: %s", message)
	}
}

// dispatchEvents sends events to subscribers
func (m *MetaAgent) dispatchEvents() {
	for event := range m.events {
		m.mu.RLock()
		for _, sub := range m.eventSubs {
			select {
			case sub <- event:
			default:
				// Subscriber channel full
			}
		}
		m.mu.RUnlock()

		// Also log
		log.Printf("[%s] %s: %s", event.Type, event.TaskID, event.Message)
	}
}

// SubscribeToEvents returns a channel for receiving events
func (m *MetaAgent) SubscribeToEvents() chan ExecutionEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan ExecutionEvent, 100)
	m.eventSubs = append(m.eventSubs, ch)
	return ch
}

// GetProgress returns current execution progress
func (m *MetaAgent) GetProgress() (total, completed, failed, running, pending int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.dag == nil {
		return 0, 0, 0, 0, 0
	}
	return m.dag.GetProgress()
}

// GetArtifacts returns all produced artifacts
func (m *MetaAgent) GetArtifacts() map[string]map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]map[string]interface{})
	for k, v := range m.artifacts {
		result[k] = v
	}
	return result
}

// Cancel cancels the execution
func (m *MetaAgent) Cancel() {
	m.cancel()
	m.wg.Wait()
	close(m.events)
}

// Wait waits for all tasks to complete
func (m *MetaAgent) Wait() {
	m.wg.Wait()
}

// GetDAG returns the current DAG
func (m *MetaAgent) GetDAG() *TaskDAG {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dag
}

// GetActiveAgents returns currently active agent assignments
func (m *MetaAgent) GetActiveAgents() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range m.activeAgents {
		result[k] = v
	}
	return result
}
