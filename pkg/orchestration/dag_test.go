package orchestration

import (
	"testing"
)

func TestNewTaskDAG(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	if dag.ID != "dag-1" {
		t.Errorf("expected ID 'dag-1', got '%s'", dag.ID)
	}
	if dag.ProjectID != "project-1" {
		t.Errorf("expected ProjectID 'project-1', got '%s'", dag.ProjectID)
	}
	if len(dag.Tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(dag.Tasks))
	}
}

func TestAddTask(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	task := &DAGTask{
		ID:          "task-1",
		Name:        "Test Task",
		Description: "A test task",
		Type:        "code",
		AgentRole:   "developer",
	}

	err := dag.AddTask(task)
	if err != nil {
		t.Fatalf("failed to add task: %v", err)
	}

	if len(dag.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(dag.Tasks))
	}

	// Verify defaults are set
	if task.Status != TaskPending {
		t.Errorf("expected status PENDING, got %s", task.Status)
	}
	if task.Priority != PriorityNormal {
		t.Errorf("expected priority %d, got %d", PriorityNormal, task.Priority)
	}
}

func TestAddDuplicateTask(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	task := &DAGTask{ID: "task-1", Name: "Task 1"}
	dag.AddTask(task)

	err := dag.AddTask(&DAGTask{ID: "task-1", Name: "Duplicate"})
	if err == nil {
		t.Error("expected error for duplicate task")
	}
}

func TestRemoveTask(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	task1 := &DAGTask{ID: "task-1", Name: "Task 1"}
	task2 := &DAGTask{ID: "task-2", Name: "Task 2", Dependencies: []string{"task-1"}}

	dag.AddTask(task1)
	dag.AddTask(task2)

	err := dag.RemoveTask("task-1")
	if err != nil {
		t.Fatalf("failed to remove task: %v", err)
	}

	if len(dag.Tasks) != 1 {
		t.Errorf("expected 1 task after removal, got %d", len(dag.Tasks))
	}

	// Try to remove non-existent task
	err = dag.RemoveTask("task-nonexistent")
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestGetTask(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	task := &DAGTask{ID: "task-1", Name: "Task 1"}
	dag.AddTask(task)

	retrieved, err := dag.GetTask("task-1")
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if retrieved.Name != "Task 1" {
		t.Errorf("expected name 'Task 1', got '%s'", retrieved.Name)
	}

	// Try to get non-existent task
	_, err = dag.GetTask("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent task")
	}
}

func TestValidateDAG_ValidGraph(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	// Create a valid DAG: task-1 -> task-2 -> task-3
	dag.AddTask(&DAGTask{ID: "task-1", Name: "Task 1"})
	dag.AddTask(&DAGTask{ID: "task-2", Name: "Task 2", Dependencies: []string{"task-1"}})
	dag.AddTask(&DAGTask{ID: "task-3", Name: "Task 3", Dependencies: []string{"task-2"}})

	err := dag.ValidateDAG()
	if err != nil {
		t.Errorf("valid DAG should not return error: %v", err)
	}
}

func TestValidateDAG_CycleDetection(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	// Create a cycle: task-1 -> task-2 -> task-3 -> task-1
	dag.AddTask(&DAGTask{ID: "task-1", Name: "Task 1", Dependencies: []string{"task-3"}})
	dag.AddTask(&DAGTask{ID: "task-2", Name: "Task 2", Dependencies: []string{"task-1"}})
	dag.AddTask(&DAGTask{ID: "task-3", Name: "Task 3", Dependencies: []string{"task-2"}})

	err := dag.ValidateDAG()
	if err == nil {
		t.Error("expected error for cyclic DAG")
	}
}

func TestValidateDAG_MissingDependency(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	dag.AddTask(&DAGTask{ID: "task-1", Name: "Task 1", Dependencies: []string{"nonexistent"}})

	err := dag.ValidateDAG()
	if err == nil {
		t.Error("expected error for missing dependency")
	}
}

func TestGetReadyTasks(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	// task-1 and task-2 have no dependencies (ready)
	// task-3 depends on task-1
	dag.AddTask(&DAGTask{ID: "task-1", Name: "Task 1", Priority: PriorityHigh})
	dag.AddTask(&DAGTask{ID: "task-2", Name: "Task 2", Priority: PriorityLow})
	dag.AddTask(&DAGTask{ID: "task-3", Name: "Task 3", Dependencies: []string{"task-1"}})

	ready := dag.GetReadyTasks()

	if len(ready) != 2 {
		t.Fatalf("expected 2 ready tasks, got %d", len(ready))
	}

	// Should be sorted by priority (high first)
	if ready[0].ID != "task-1" {
		t.Errorf("expected task-1 first (high priority), got %s", ready[0].ID)
	}
	if ready[1].ID != "task-2" {
		t.Errorf("expected task-2 second (low priority), got %s", ready[1].ID)
	}
}

func TestGetParallelTasks(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	// Layer 1: task-1, task-2 (no deps)
	// Layer 2: task-3, task-4 (depend on layer 1)
	// Layer 3: task-5 (depends on layer 2)
	dag.AddTask(&DAGTask{ID: "task-1", Name: "Task 1"})
	dag.AddTask(&DAGTask{ID: "task-2", Name: "Task 2"})
	dag.AddTask(&DAGTask{ID: "task-3", Name: "Task 3", Dependencies: []string{"task-1"}})
	dag.AddTask(&DAGTask{ID: "task-4", Name: "Task 4", Dependencies: []string{"task-2"}})
	dag.AddTask(&DAGTask{ID: "task-5", Name: "Task 5", Dependencies: []string{"task-3", "task-4"}})

	layers := dag.GetParallelTasks()

	if len(layers) != 3 {
		t.Fatalf("expected 3 layers, got %d", len(layers))
	}

	// First layer should have 2 tasks
	if len(layers[0]) != 2 {
		t.Errorf("expected 2 tasks in layer 1, got %d", len(layers[0]))
	}

	// Second layer should have 2 tasks
	if len(layers[1]) != 2 {
		t.Errorf("expected 2 tasks in layer 2, got %d", len(layers[1]))
	}

	// Third layer should have 1 task
	if len(layers[2]) != 1 {
		t.Errorf("expected 1 task in layer 3, got %d", len(layers[2]))
	}
}

func TestMarkTaskRunning(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	task := &DAGTask{ID: "task-1", Name: "Task 1"}
	dag.AddTask(task)

	err := dag.MarkTaskRunning("task-1", "agent-1")
	if err != nil {
		t.Fatalf("failed to mark task running: %v", err)
	}

	if task.Status != TaskRunning {
		t.Errorf("expected status RUNNING, got %s", task.Status)
	}
	if task.AssignedTo != "agent-1" {
		t.Errorf("expected assigned to 'agent-1', got '%s'", task.AssignedTo)
	}
}

func TestMarkTaskRunning_InvalidState(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	task := &DAGTask{ID: "task-1", Name: "Task 1", Status: TaskCompleted}
	dag.AddTask(task)

	err := dag.MarkTaskRunning("task-1", "agent-1")
	if err == nil {
		t.Error("expected error when starting completed task")
	}
}

func TestMarkTaskCompleted(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	// task-2 depends on task-1
	dag.AddTask(&DAGTask{ID: "task-1", Name: "Task 1"})
	dag.AddTask(&DAGTask{ID: "task-2", Name: "Task 2", Dependencies: []string{"task-1"}})

	output := map[string]interface{}{"result": "success"}
	err := dag.MarkTaskCompleted("task-1", output)
	if err != nil {
		t.Fatalf("failed to mark task completed: %v", err)
	}

	task, _ := dag.GetTask("task-1")
	if task.Status != TaskCompleted {
		t.Errorf("expected status COMPLETED, got %s", task.Status)
	}
	if task.Output["result"] != "success" {
		t.Error("output not stored correctly")
	}

	// task-2 should now be ready
	ready := dag.GetReadyTasks()
	if len(ready) != 1 || ready[0].ID != "task-2" {
		t.Error("task-2 should be ready after task-1 completion")
	}
}

func TestMarkTaskFailed(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	// task-2 depends on task-1
	dag.AddTask(&DAGTask{ID: "task-1", Name: "Task 1"})
	dag.AddTask(&DAGTask{ID: "task-2", Name: "Task 2", Dependencies: []string{"task-1"}})
	dag.AddTask(&DAGTask{ID: "task-3", Name: "Task 3", Dependencies: []string{"task-2"}})

	err := dag.MarkTaskFailed("task-1", "test error", true)
	if err != nil {
		t.Fatalf("failed to mark task failed: %v", err)
	}

	task, _ := dag.GetTask("task-1")
	if task.Status != TaskFailed {
		t.Errorf("expected status FAILED, got %s", task.Status)
	}
	if task.Error != "test error" {
		t.Errorf("expected error 'test error', got '%s'", task.Error)
	}

	// Dependents should be blocked
	task2, _ := dag.GetTask("task-2")
	task3, _ := dag.GetTask("task-3")
	if task2.Status != TaskBlocked {
		t.Errorf("task-2 should be BLOCKED, got %s", task2.Status)
	}
	if task3.Status != TaskBlocked {
		t.Errorf("task-3 should be BLOCKED, got %s", task3.Status)
	}
}

func TestRetryTask(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	task := &DAGTask{ID: "task-1", Name: "Task 1", Status: TaskFailed, MaxRetries: 3}
	dag.AddTask(task)

	err := dag.RetryTask("task-1")
	if err != nil {
		t.Fatalf("failed to retry task: %v", err)
	}

	if task.Status != TaskPending {
		t.Errorf("expected status PENDING after retry, got %s", task.Status)
	}
	if task.RetryCount != 1 {
		t.Errorf("expected retry count 1, got %d", task.RetryCount)
	}
}

func TestRetryTask_MaxRetriesExceeded(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	task := &DAGTask{ID: "task-1", Name: "Task 1", Status: TaskFailed, RetryCount: 3, MaxRetries: 3}
	dag.AddTask(task)

	err := dag.RetryTask("task-1")
	if err == nil {
		t.Error("expected error when max retries exceeded")
	}
}

func TestGetProgress(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	dag.AddTask(&DAGTask{ID: "task-1", Status: TaskCompleted})
	dag.AddTask(&DAGTask{ID: "task-2", Status: TaskRunning})
	dag.AddTask(&DAGTask{ID: "task-3", Status: TaskFailed})
	dag.AddTask(&DAGTask{ID: "task-4", Status: TaskPending})
	dag.AddTask(&DAGTask{ID: "task-5", Status: TaskPending})

	total, completed, failed, running, pending := dag.GetProgress()

	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if completed != 1 {
		t.Errorf("expected completed 1, got %d", completed)
	}
	if failed != 1 {
		t.Errorf("expected failed 1, got %d", failed)
	}
	if running != 1 {
		t.Errorf("expected running 1, got %d", running)
	}
	if pending != 2 {
		t.Errorf("expected pending 2, got %d", pending)
	}
}

func TestIsComplete(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	dag.AddTask(&DAGTask{ID: "task-1", Status: TaskCompleted})
	dag.AddTask(&DAGTask{ID: "task-2", Status: TaskCompleted})

	if !dag.IsComplete() {
		t.Error("DAG should be complete when all tasks completed")
	}

	dag.AddTask(&DAGTask{ID: "task-3", Status: TaskPending})

	if dag.IsComplete() {
		t.Error("DAG should not be complete with pending tasks")
	}
}

func TestHasFailed(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	dag.AddTask(&DAGTask{ID: "task-1", Status: TaskCompleted})

	if dag.HasFailed() {
		t.Error("DAG should not have failed")
	}

	// Add a failed task with retries remaining
	dag.AddTask(&DAGTask{ID: "task-2", Status: TaskFailed, RetryCount: 1, MaxRetries: 3})

	if dag.HasFailed() {
		t.Error("DAG should not have failed with retries remaining")
	}

	// Exhaust retries
	task2, _ := dag.GetTask("task-2")
	task2.RetryCount = 3

	if !dag.HasFailed() {
		t.Error("DAG should have failed with max retries exceeded")
	}
}

func TestTopologicalSort(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	dag.AddTask(&DAGTask{ID: "task-1", Name: "Task 1"})
	dag.AddTask(&DAGTask{ID: "task-2", Name: "Task 2", Dependencies: []string{"task-1"}})
	dag.AddTask(&DAGTask{ID: "task-3", Name: "Task 3", Dependencies: []string{"task-1"}})
	dag.AddTask(&DAGTask{ID: "task-4", Name: "Task 4", Dependencies: []string{"task-2", "task-3"}})

	sorted, err := dag.TopologicalSort()
	if err != nil {
		t.Fatalf("topological sort failed: %v", err)
	}

	if len(sorted) != 4 {
		t.Fatalf("expected 4 tasks in sorted order, got %d", len(sorted))
	}

	// task-1 should come before task-2 and task-3
	// task-2 and task-3 should come before task-4
	taskIndex := make(map[string]int)
	for i, task := range sorted {
		taskIndex[task.ID] = i
	}

	if taskIndex["task-1"] > taskIndex["task-2"] {
		t.Error("task-1 should come before task-2")
	}
	if taskIndex["task-1"] > taskIndex["task-3"] {
		t.Error("task-1 should come before task-3")
	}
	if taskIndex["task-2"] > taskIndex["task-4"] {
		t.Error("task-2 should come before task-4")
	}
	if taskIndex["task-3"] > taskIndex["task-4"] {
		t.Error("task-3 should come before task-4")
	}
}

func TestGetDependencies(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	dag.AddTask(&DAGTask{ID: "task-1", Name: "Task 1"})
	dag.AddTask(&DAGTask{ID: "task-2", Name: "Task 2"})
	dag.AddTask(&DAGTask{ID: "task-3", Name: "Task 3", Dependencies: []string{"task-1", "task-2"}})

	deps, err := dag.GetDependencies("task-3")
	if err != nil {
		t.Fatalf("failed to get dependencies: %v", err)
	}

	if len(deps) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(deps))
	}
}

func TestGetDependents(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	dag.AddTask(&DAGTask{ID: "task-1", Name: "Task 1"})
	dag.AddTask(&DAGTask{ID: "task-2", Name: "Task 2", Dependencies: []string{"task-1"}})
	dag.AddTask(&DAGTask{ID: "task-3", Name: "Task 3", Dependencies: []string{"task-1"}})

	dependents, err := dag.GetDependents("task-1")
	if err != nil {
		t.Fatalf("failed to get dependents: %v", err)
	}

	if len(dependents) != 2 {
		t.Errorf("expected 2 dependents, got %d", len(dependents))
	}
}

func TestClone(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	dag.AddTask(&DAGTask{
		ID:           "task-1",
		Name:         "Task 1",
		Dependencies: []string{},
		Produces:     []string{"file1.go"},
		Input:        map[string]interface{}{"key": "value"},
	})
	dag.AddTask(&DAGTask{ID: "task-2", Name: "Task 2", Dependencies: []string{"task-1"}})

	clone := dag.Clone()

	if clone.ID != "dag-1-clone" {
		t.Errorf("clone ID should be 'dag-1-clone', got '%s'", clone.ID)
	}

	if len(clone.Tasks) != 2 {
		t.Errorf("expected 2 tasks in clone, got %d", len(clone.Tasks))
	}

	// Modify original should not affect clone
	dag.Tasks["task-1"].Name = "Modified"
	if clone.Tasks["task-1"].Name == "Modified" {
		t.Error("clone should be independent of original")
	}
}

func TestGetTasksByStatus(t *testing.T) {
	dag := NewTaskDAG("dag-1", "project-1")

	dag.AddTask(&DAGTask{ID: "task-1", Status: TaskPending})
	dag.AddTask(&DAGTask{ID: "task-2", Status: TaskRunning})
	dag.AddTask(&DAGTask{ID: "task-3", Status: TaskPending})

	pending := dag.GetTasksByStatus(TaskPending)
	if len(pending) != 2 {
		t.Errorf("expected 2 pending tasks, got %d", len(pending))
	}

	running := dag.GetTasksByStatus(TaskRunning)
	if len(running) != 1 {
		t.Errorf("expected 1 running task, got %d", len(running))
	}
}
