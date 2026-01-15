package agent

import (
	"context"
	"testing"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// mockLLMClient implements llm.Client for testing
type mockLLMClient struct{}

func (m *mockLLMClient) Generate(ctx context.Context, req llm.GenerateRequest) (*llm.GenerateResponse, error) {
	return &llm.GenerateResponse{
		Response: "mock response",
		Done:     true,
	}, nil
}

func (m *mockLLMClient) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *mockLLMClient) Ping(ctx context.Context) error {
	return nil
}

func TestNewAgentInstance(t *testing.T) {
	skills := []Skill{
		{
			Name:         "test-skill",
			Version:      "1.0",
			Description:  "A test skill",
			Capabilities: []string{"testing"},
			Tools:        []string{"test_tool"},
			SystemPrompt: "You are a test agent.",
		},
	}

	agent := NewAgentInstance("TestBot", "tester", "project-123", "gpt-4", skills, &mockLLMClient{})

	if agent == nil {
		t.Fatal("Expected agent to be created, got nil")
	}

	if agent.Name != "TestBot" {
		t.Errorf("Expected name 'TestBot', got '%s'", agent.Name)
	}

	if agent.Role != "tester" {
		t.Errorf("Expected role 'tester', got '%s'", agent.Role)
	}

	if agent.ProjectID != "project-123" {
		t.Errorf("Expected project ID 'project-123', got '%s'", agent.ProjectID)
	}

	if agent.Status != StatusSpawning {
		t.Errorf("Expected status SPAWNING, got '%s'", agent.Status)
	}

	if agent.ID == "" {
		t.Error("Expected agent ID to be set")
	}

	if len(agent.Skills) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(agent.Skills))
	}
}

func TestAgentSpawn(t *testing.T) {
	agent := NewAgentInstance("TestBot", "tester", "project-123", "gpt-4", nil, &mockLLMClient{})

	err := agent.Spawn(context.Background())
	if err != nil {
		t.Fatalf("Expected spawn to succeed, got error: %v", err)
	}

	if agent.GetStatus() != StatusReady {
		t.Errorf("Expected status READY after spawn, got '%s'", agent.GetStatus())
	}

	// Trying to spawn again should fail
	err = agent.Spawn(context.Background())
	if err != ErrInvalidStatusTransition {
		t.Errorf("Expected ErrInvalidStatusTransition on double spawn, got: %v", err)
	}
}

func TestAgentAssignTask(t *testing.T) {
	agent := NewAgentInstance("TestBot", "tester", "project-123", "gpt-4", nil, &mockLLMClient{})
	agent.Spawn(context.Background())

	task := &Task{
		ID:          "task-1",
		Description: "Test task",
		Type:        "test",
		Priority:    1,
	}

	err := agent.AssignTask(task)
	if err != nil {
		t.Fatalf("Expected task assignment to succeed, got error: %v", err)
	}

	if agent.GetStatus() != StatusWorking {
		t.Errorf("Expected status WORKING after task assignment, got '%s'", agent.GetStatus())
	}

	if agent.CurrentTask == nil {
		t.Error("Expected current task to be set")
	}

	if task.Status != "IN_PROGRESS" {
		t.Errorf("Expected task status 'IN_PROGRESS', got '%s'", task.Status)
	}

	// Trying to assign another task while busy should fail
	task2 := &Task{ID: "task-2", Description: "Another task"}
	err = agent.AssignTask(task2)
	if err != ErrAgentBusy {
		t.Errorf("Expected ErrAgentBusy, got: %v", err)
	}
}

func TestAgentCompleteTask(t *testing.T) {
	agent := NewAgentInstance("TestBot", "tester", "project-123", "gpt-4", nil, &mockLLMClient{})
	agent.Spawn(context.Background())

	task := &Task{
		ID:          "task-1",
		Description: "Test task",
		Type:        "test",
	}
	agent.AssignTask(task)

	output := map[string]interface{}{"result": "success"}
	err := agent.CompleteTask(output)
	if err != nil {
		t.Fatalf("Expected task completion to succeed, got error: %v", err)
	}

	if agent.GetStatus() != StatusReady {
		t.Errorf("Expected status READY after task completion, got '%s'", agent.GetStatus())
	}

	if agent.CurrentTask != nil {
		t.Error("Expected current task to be cleared")
	}

	tokens, tasks := agent.GetMetrics()
	if tasks != 1 {
		t.Errorf("Expected 1 completed task, got %d", tasks)
	}
	if tokens != 0 {
		t.Errorf("Expected 0 tokens used, got %d", tokens)
	}

	// Completing without an active task should fail
	err = agent.CompleteTask(nil)
	if err != ErrNoActiveTask {
		t.Errorf("Expected ErrNoActiveTask, got: %v", err)
	}
}

func TestAgentAddArtifact(t *testing.T) {
	agent := NewAgentInstance("TestBot", "tester", "project-123", "gpt-4", nil, &mockLLMClient{})

	artifact := agent.AddArtifact("file", "/src/main.go", "package main", "go", "Main entry point")

	if artifact == nil {
		t.Fatal("Expected artifact to be created, got nil")
	}

	if artifact.ID == "" {
		t.Error("Expected artifact ID to be set")
	}

	if artifact.AgentID != agent.ID {
		t.Errorf("Expected artifact agent ID '%s', got '%s'", agent.ID, artifact.AgentID)
	}

	if artifact.Type != "file" {
		t.Errorf("Expected artifact type 'file', got '%s'", artifact.Type)
	}

	if len(agent.Artifacts) != 1 {
		t.Errorf("Expected 1 artifact, got %d", len(agent.Artifacts))
	}
}

func TestAgentSendMessage(t *testing.T) {
	agent := NewAgentInstance("TestBot", "tester", "project-123", "gpt-4", nil, &mockLLMClient{})

	agent.SendMessage(MessageQuestion, "other-agent", "Test Subject", "Test Content")

	select {
	case msg := <-agent.Outbox:
		if msg.Type != MessageQuestion {
			t.Errorf("Expected message type QUESTION, got '%s'", msg.Type)
		}
		if msg.From != agent.ID {
			t.Errorf("Expected from '%s', got '%s'", agent.ID, msg.From)
		}
		if msg.To != "other-agent" {
			t.Errorf("Expected to 'other-agent', got '%s'", msg.To)
		}
		if msg.Subject != "Test Subject" {
			t.Errorf("Expected subject 'Test Subject', got '%s'", msg.Subject)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected message in outbox, timed out")
	}
}

func TestAgentTerminate(t *testing.T) {
	agent := NewAgentInstance("TestBot", "tester", "project-123", "gpt-4", nil, &mockLLMClient{})
	agent.Spawn(context.Background())

	if !agent.IsAlive() {
		t.Error("Expected agent to be alive before termination")
	}

	agent.Terminate()

	if agent.GetStatus() != StatusTerminated {
		t.Errorf("Expected status TERMINATED, got '%s'", agent.GetStatus())
	}

	if agent.IsAlive() {
		t.Error("Expected agent to not be alive after termination")
	}

	if agent.TerminatedAt == nil {
		t.Error("Expected terminated timestamp to be set")
	}

	// Terminating again should be a no-op
	agent.Terminate()
	if agent.GetStatus() != StatusTerminated {
		t.Error("Expected status to remain TERMINATED")
	}
}

func TestAgentSystemPromptCompilation(t *testing.T) {
	skills := []Skill{
		{
			Name:         "skill-1",
			SystemPrompt: "First skill instructions.",
		},
		{
			Name:         "skill-2",
			SystemPrompt: "Second skill instructions.",
		},
	}

	agent := NewAgentInstance("Alice", "developer", "project-123", "gpt-4", skills, &mockLLMClient{})

	if agent.SystemPrompt == "" {
		t.Error("Expected system prompt to be compiled")
	}

	// Check that name and role are included
	if !contains(agent.SystemPrompt, "Alice") {
		t.Error("Expected system prompt to contain agent name")
	}

	if !contains(agent.SystemPrompt, "developer") {
		t.Error("Expected system prompt to contain agent role")
	}

	// Check that skill prompts are included
	if !contains(agent.SystemPrompt, "First skill instructions") {
		t.Error("Expected system prompt to contain first skill prompt")
	}

	if !contains(agent.SystemPrompt, "Second skill instructions") {
		t.Error("Expected system prompt to contain second skill prompt")
	}
}

func TestAgentStatusConstants(t *testing.T) {
	// Verify status constants are as expected
	tests := []struct {
		status   AgentStatus
		expected string
	}{
		{StatusSpawning, "SPAWNING"},
		{StatusReady, "READY"},
		{StatusWorking, "WORKING"},
		{StatusBlocked, "BLOCKED"},
		{StatusReviewing, "REVIEWING"},
		{StatusCompleted, "COMPLETED"},
		{StatusTerminated, "TERMINATED"},
		{StatusError, "ERROR"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("Expected status '%s', got '%s'", tt.expected, tt.status)
		}
	}
}

func TestMessageTypeConstants(t *testing.T) {
	tests := []struct {
		msgType  MessageType
		expected string
	}{
		{MessageQuestion, "QUESTION"},
		{MessageAnswer, "ANSWER"},
		{MessageArtifact, "ARTIFACT"},
		{MessageStatus, "STATUS"},
		{MessageRequest, "REQUEST"},
		{MessageReview, "REVIEW"},
		{MessageBroadcast, "BROADCAST"},
	}

	for _, tt := range tests {
		if string(tt.msgType) != tt.expected {
			t.Errorf("Expected message type '%s', got '%s'", tt.expected, tt.msgType)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
