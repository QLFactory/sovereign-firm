// Package orchestration provides multi-agent coordination capabilities
// including task DAG scheduling, file locking, and Git branch management.
package orchestration

import (
	"errors"
	"fmt"
	"sync"
)

// TaskStatus represents the execution state of a task
type TaskStatus string

const (
	TaskPending    TaskStatus = "PENDING"
	TaskReady      TaskStatus = "READY"      // All dependencies satisfied
	TaskRunning    TaskStatus = "RUNNING"
	TaskCompleted  TaskStatus = "COMPLETED"
	TaskFailed     TaskStatus = "FAILED"
	TaskBlocked    TaskStatus = "BLOCKED"    // Dependency failed
	TaskCancelled  TaskStatus = "CANCELLED"
)

// TaskPriority defines execution priority
type TaskPriority int

const (
	PriorityLow    TaskPriority = 1
	PriorityNormal TaskPriority = 5
	PriorityHigh   TaskPriority = 10
)

// DAGTask represents a task in the dependency graph
type DAGTask struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         string                 `json:"type"`          // "code", "test", "review", "deploy"
	AgentRole    string                 `json:"agent_role"`    // Role of agent to execute
	Dependencies []string               `json:"dependencies"`  // Task IDs that must complete first
	Produces     []string               `json:"produces"`      // File paths this task produces
	Consumes     []string               `json:"consumes"`      // File paths this task needs
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output,omitempty"`
	Status       TaskStatus             `json:"status"`
	Priority     TaskPriority           `json:"priority"`
	AssignedTo   string                 `json:"assigned_to,omitempty"` // Agent ID
	Error        string                 `json:"error,omitempty"`
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
}

// TaskDAG represents a Directed Acyclic Graph of tasks
type TaskDAG struct {
	ID          string              `json:"id"`
	ProjectID   string              `json:"project_id"`
	Tasks       map[string]*DAGTask `json:"tasks"`
	adjacency   map[string][]string // task -> dependent tasks (reverse edges for notification)
	inDegree    map[string]int      // number of unsatisfied dependencies
	mu          sync.RWMutex
}

// NewTaskDAG creates a new task DAG
func NewTaskDAG(id, projectID string) *TaskDAG {
	return &TaskDAG{
		ID:        id,
		ProjectID: projectID,
		Tasks:     make(map[string]*DAGTask),
		adjacency: make(map[string][]string),
		inDegree:  make(map[string]int),
	}
}

// AddTask adds a task to the DAG
func (d *TaskDAG) AddTask(task *DAGTask) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.Tasks[task.ID]; exists {
		return fmt.Errorf("task %s already exists", task.ID)
	}

	// Set defaults
	if task.Status == "" {
		task.Status = TaskPending
	}
	if task.Priority == 0 {
		task.Priority = PriorityNormal
	}
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}

	d.Tasks[task.ID] = task
	d.inDegree[task.ID] = len(task.Dependencies)

	// Build adjacency list (reverse edges)
	for _, depID := range task.Dependencies {
		d.adjacency[depID] = append(d.adjacency[depID], task.ID)
	}

	return nil
}

// RemoveTask removes a task from the DAG
func (d *TaskDAG) RemoveTask(taskID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	task, exists := d.Tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Remove from adjacency lists
	for _, depID := range task.Dependencies {
		deps := d.adjacency[depID]
		for i, id := range deps {
			if id == taskID {
				d.adjacency[depID] = append(deps[:i], deps[i+1:]...)
				break
			}
		}
	}

	delete(d.Tasks, taskID)
	delete(d.inDegree, taskID)
	delete(d.adjacency, taskID)

	return nil
}

// GetTask retrieves a task by ID
func (d *TaskDAG) GetTask(taskID string) (*DAGTask, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	task, exists := d.Tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}
	return task, nil
}

// ValidateDAG checks if the DAG is valid (no cycles, all dependencies exist)
func (d *TaskDAG) ValidateDAG() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Check all dependencies exist
	for _, task := range d.Tasks {
		for _, depID := range task.Dependencies {
			if _, exists := d.Tasks[depID]; !exists {
				return fmt.Errorf("task %s depends on non-existent task %s", task.ID, depID)
			}
		}
	}

	// Check for cycles using Kahn's algorithm
	inDegree := make(map[string]int)
	for id, deg := range d.inDegree {
		inDegree[id] = deg
	}

	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	visited := 0
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		visited++

		for _, next := range d.adjacency[current] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if visited != len(d.Tasks) {
		return errors.New("DAG contains a cycle")
	}

	return nil
}

// GetReadyTasks returns all tasks that are ready to execute
func (d *TaskDAG) GetReadyTasks() []*DAGTask {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var ready []*DAGTask
	for _, task := range d.Tasks {
		if task.Status == TaskPending && d.inDegree[task.ID] == 0 {
			ready = append(ready, task)
		}
	}

	// Sort by priority (highest first)
	for i := 0; i < len(ready)-1; i++ {
		for j := i + 1; j < len(ready); j++ {
			if ready[j].Priority > ready[i].Priority {
				ready[i], ready[j] = ready[j], ready[i]
			}
		}
	}

	return ready
}

// GetParallelTasks returns tasks that can run in parallel at the current layer
func (d *TaskDAG) GetParallelTasks() [][]*DAGTask {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Use topological sort to find parallel layers
	inDegree := make(map[string]int)
	for id := range d.Tasks {
		inDegree[id] = 0
	}
	for _, task := range d.Tasks {
		for _, depID := range task.Dependencies {
			inDegree[task.ID]++
			_ = depID // dependency exists
		}
	}

	var layers [][]*DAGTask
	remaining := len(d.Tasks)

	for remaining > 0 {
		var layer []*DAGTask
		var toRemove []string

		for id, deg := range inDegree {
			if deg == 0 {
				task := d.Tasks[id]
				if task.Status == TaskPending || task.Status == TaskReady {
					layer = append(layer, task)
				}
				toRemove = append(toRemove, id)
			}
		}

		if len(toRemove) == 0 {
			break // No progress, might have cycle or all done
		}

		for _, id := range toRemove {
			delete(inDegree, id)
			remaining--
			// Decrease in-degree of dependent tasks
			for _, depTask := range d.adjacency[id] {
				if _, exists := inDegree[depTask]; exists {
					inDegree[depTask]--
				}
			}
		}

		if len(layer) > 0 {
			layers = append(layers, layer)
		}
	}

	return layers
}

// MarkTaskRunning marks a task as running
func (d *TaskDAG) MarkTaskRunning(taskID, agentID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	task, exists := d.Tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	if task.Status != TaskPending && task.Status != TaskReady {
		return fmt.Errorf("task %s cannot start from status %s", taskID, task.Status)
	}

	task.Status = TaskRunning
	task.AssignedTo = agentID
	return nil
}

// MarkTaskCompleted marks a task as completed and updates dependents
func (d *TaskDAG) MarkTaskCompleted(taskID string, output map[string]interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	task, exists := d.Tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	task.Status = TaskCompleted
	task.Output = output

	// Update in-degree of dependent tasks
	for _, depTaskID := range d.adjacency[taskID] {
		d.inDegree[depTaskID]--
	}

	return nil
}

// MarkTaskFailed marks a task as failed and optionally blocks dependents
func (d *TaskDAG) MarkTaskFailed(taskID string, err string, blockDependents bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	task, exists := d.Tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	task.Status = TaskFailed
	task.Error = err

	if blockDependents {
		// Block all dependent tasks
		d.blockDependentsRecursive(taskID)
	}

	return nil
}

// blockDependentsRecursive marks all dependent tasks as blocked
func (d *TaskDAG) blockDependentsRecursive(taskID string) {
	for _, depTaskID := range d.adjacency[taskID] {
		task := d.Tasks[depTaskID]
		if task.Status == TaskPending || task.Status == TaskReady {
			task.Status = TaskBlocked
			d.blockDependentsRecursive(depTaskID)
		}
	}
}

// RetryTask resets a failed task for retry
func (d *TaskDAG) RetryTask(taskID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	task, exists := d.Tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	if task.Status != TaskFailed {
		return fmt.Errorf("task %s is not failed, cannot retry", taskID)
	}

	if task.RetryCount >= task.MaxRetries {
		return fmt.Errorf("task %s has exceeded max retries (%d)", taskID, task.MaxRetries)
	}

	task.RetryCount++
	task.Status = TaskPending
	task.Error = ""
	task.AssignedTo = ""
	task.Output = nil

	return nil
}

// GetProgress returns DAG execution progress
func (d *TaskDAG) GetProgress() (total, completed, failed, running, pending int) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	total = len(d.Tasks)
	for _, task := range d.Tasks {
		switch task.Status {
		case TaskCompleted:
			completed++
		case TaskFailed:
			failed++
		case TaskRunning:
			running++
		case TaskPending, TaskReady:
			pending++
		}
	}
	return
}

// IsComplete returns true if all tasks are completed
func (d *TaskDAG) IsComplete() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, task := range d.Tasks {
		if task.Status != TaskCompleted && task.Status != TaskCancelled {
			return false
		}
	}
	return true
}

// HasFailed returns true if any task has failed (and not retryable)
func (d *TaskDAG) HasFailed() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, task := range d.Tasks {
		if task.Status == TaskFailed && task.RetryCount >= task.MaxRetries {
			return true
		}
	}
	return false
}

// GetTasksByStatus returns all tasks with a specific status
func (d *TaskDAG) GetTasksByStatus(status TaskStatus) []*DAGTask {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var tasks []*DAGTask
	for _, task := range d.Tasks {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// GetDependencies returns the dependencies of a task
func (d *TaskDAG) GetDependencies(taskID string) ([]*DAGTask, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	task, exists := d.Tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	var deps []*DAGTask
	for _, depID := range task.Dependencies {
		if dep, exists := d.Tasks[depID]; exists {
			deps = append(deps, dep)
		}
	}
	return deps, nil
}

// GetDependents returns tasks that depend on this task
func (d *TaskDAG) GetDependents(taskID string) ([]*DAGTask, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if _, exists := d.Tasks[taskID]; !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	var dependents []*DAGTask
	for _, depID := range d.adjacency[taskID] {
		if dep, exists := d.Tasks[depID]; exists {
			dependents = append(dependents, dep)
		}
	}
	return dependents, nil
}

// TopologicalSort returns tasks in execution order
func (d *TaskDAG) TopologicalSort() ([]*DAGTask, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	inDegree := make(map[string]int)
	for id, task := range d.Tasks {
		inDegree[id] = len(task.Dependencies)
	}

	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var sorted []*DAGTask
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, d.Tasks[current])

		for _, next := range d.adjacency[current] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(sorted) != len(d.Tasks) {
		return nil, errors.New("DAG contains a cycle")
	}

	return sorted, nil
}

// Clone creates a deep copy of the DAG
func (d *TaskDAG) Clone() *TaskDAG {
	d.mu.RLock()
	defer d.mu.RUnlock()

	clone := NewTaskDAG(d.ID+"-clone", d.ProjectID)
	for id, task := range d.Tasks {
		clonedTask := &DAGTask{
			ID:           task.ID,
			Name:         task.Name,
			Description:  task.Description,
			Type:         task.Type,
			AgentRole:    task.AgentRole,
			Dependencies: make([]string, len(task.Dependencies)),
			Produces:     make([]string, len(task.Produces)),
			Consumes:     make([]string, len(task.Consumes)),
			Status:       task.Status,
			Priority:     task.Priority,
			MaxRetries:   task.MaxRetries,
		}
		copy(clonedTask.Dependencies, task.Dependencies)
		copy(clonedTask.Produces, task.Produces)
		copy(clonedTask.Consumes, task.Consumes)

		if task.Input != nil {
			clonedTask.Input = make(map[string]interface{})
			for k, v := range task.Input {
				clonedTask.Input[k] = v
			}
		}

		clone.Tasks[id] = clonedTask
		clone.inDegree[id] = d.inDegree[id]
	}

	for k, v := range d.adjacency {
		clone.adjacency[k] = make([]string, len(v))
		copy(clone.adjacency[k], v)
	}

	return clone
}
