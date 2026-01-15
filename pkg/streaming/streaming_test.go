package streaming

import (
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()

	if hub == nil {
		t.Fatal("Expected hub to be created, got nil")
	}

	if hub.clients == nil {
		t.Error("Expected clients map to be initialized")
	}

	if hub.broadcast == nil {
		t.Error("Expected broadcast channel to be initialized")
	}
}

func TestNewEvent(t *testing.T) {
	payload := map[string]interface{}{
		"key": "value",
	}

	event := NewEvent(EventPhaseChange, payload)

	if event == nil {
		t.Fatal("Expected event to be created, got nil")
	}

	if event.Type != EventPhaseChange {
		t.Errorf("Expected type '%s', got '%s'", EventPhaseChange, event.Type)
	}

	if event.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if event.Payload["key"] != "value" {
		t.Errorf("Expected payload key 'value', got '%v'", event.Payload["key"])
	}
}

func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		eventType StreamEventType
		expected  string
	}{
		{EventConnected, "CONNECTED"},
		{EventDisconnected, "DISCONNECTED"},
		{EventError, "ERROR"},
		{EventPhaseChange, "PHASE_CHANGE"},
		{EventAgentSpawn, "AGENT_SPAWN"},
		{EventAgentDone, "AGENT_DONE"},
		{EventCodeChunk, "CODE_CHUNK"},
		{EventCodeComplete, "CODE_COMPLETE"},
		{EventFileStart, "FILE_START"},
		{EventFileEnd, "FILE_END"},
		{EventChatMessage, "CHAT_MESSAGE"},
		{EventChatTyping, "CHAT_TYPING"},
		{EventBuildStart, "BUILD_START"},
		{EventBuildOutput, "BUILD_OUTPUT"},
		{EventBuildSuccess, "BUILD_SUCCESS"},
		{EventBuildError, "BUILD_ERROR"},
		{EventTestStart, "TEST_START"},
		{EventTestResult, "TEST_RESULT"},
		{EventPreviewReady, "PREVIEW_READY"},
	}

	for _, tt := range tests {
		if string(tt.eventType) != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, tt.eventType)
		}
	}
}

func TestHubGetClientCount(t *testing.T) {
	hub := NewHub()

	// Initially no clients
	count := hub.GetClientCount("workflow-1")
	if count != 0 {
		t.Errorf("Expected 0 clients, got %d", count)
	}
}

func TestHubBroadcastHelpers(t *testing.T) {
	hub := NewHub()

	// Start hub in background
	go hub.Run()

	// Give hub time to start
	time.Sleep(10 * time.Millisecond)

	// Test that broadcast helpers don't panic (no clients connected)
	// These should not block or panic even with no clients
	done := make(chan bool)

	go func() {
		hub.BroadcastPhaseChange("workflow-1", "DISCOVERY", "IMPLEMENTATION", "Moving to implementation")
		hub.BroadcastCodeChunk("workflow-1", "/src/main.go", 0, "package main", false)
		hub.BroadcastChatMessage("workflow-1", "agent", "PM", "Hello!")
		hub.BroadcastFileStart("workflow-1", "/src/main.go", "go")
		hub.BroadcastFileEnd("workflow-1", "/src/main.go")
		hub.BroadcastError("workflow-1", "E001", "Test error", "Details here")
		done <- true
	}()

	select {
	case <-done:
		// Success - broadcasts completed without blocking
	case <-time.After(1 * time.Second):
		t.Error("Broadcast helpers blocked - possible deadlock")
	}
}

func TestStreamEventSerialization(t *testing.T) {
	event := NewEvent(EventChatMessage, map[string]interface{}{
		"role":    "agent",
		"agent":   "PM",
		"content": "Hello, let me help you with your project.",
	})

	if event.Payload["role"] != "agent" {
		t.Errorf("Expected role 'agent', got '%v'", event.Payload["role"])
	}

	if event.Payload["agent"] != "PM" {
		t.Errorf("Expected agent 'PM', got '%v'", event.Payload["agent"])
	}
}

func TestPhaseChangePayloadStructure(t *testing.T) {
	payload := PhaseChangePayload{
		PreviousPhase: "DISCOVERY",
		CurrentPhase:  "IMPLEMENTATION",
		Message:       "Moving to next phase",
	}

	if payload.PreviousPhase != "DISCOVERY" {
		t.Errorf("Expected previous phase 'DISCOVERY', got '%s'", payload.PreviousPhase)
	}

	if payload.CurrentPhase != "IMPLEMENTATION" {
		t.Errorf("Expected current phase 'IMPLEMENTATION', got '%s'", payload.CurrentPhase)
	}
}

func TestCodeChunkPayloadStructure(t *testing.T) {
	payload := CodeChunkPayload{
		FilePath:    "/src/main.go",
		ChunkIndex:  0,
		Content:     "package main\n\nfunc main() {",
		IsComplete:  false,
		TotalChunks: 5,
	}

	if payload.FilePath != "/src/main.go" {
		t.Errorf("Expected file path '/src/main.go', got '%s'", payload.FilePath)
	}

	if payload.ChunkIndex != 0 {
		t.Errorf("Expected chunk index 0, got %d", payload.ChunkIndex)
	}

	if payload.IsComplete {
		t.Error("Expected IsComplete to be false")
	}
}

func TestChatMessagePayloadStructure(t *testing.T) {
	payload := ChatMessagePayload{
		Role:    "agent",
		Agent:   "Dev",
		Content: "I'll implement that feature now.",
	}

	if payload.Role != "agent" {
		t.Errorf("Expected role 'agent', got '%s'", payload.Role)
	}

	if payload.Agent != "Dev" {
		t.Errorf("Expected agent 'Dev', got '%s'", payload.Agent)
	}
}

func TestAgentPayloadStructure(t *testing.T) {
	payload := AgentPayload{
		AgentID:   "agent-123",
		AgentType: "backend-developer",
		Skills:    []string{"go", "api-design"},
		TaskID:    "task-456",
		Result:    "completed",
	}

	if payload.AgentID != "agent-123" {
		t.Errorf("Expected agent ID 'agent-123', got '%s'", payload.AgentID)
	}

	if len(payload.Skills) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(payload.Skills))
	}
}

func TestBuildOutputPayloadStructure(t *testing.T) {
	payload := BuildOutputPayload{
		Line:     "Building main.go...",
		IsStderr: false,
	}

	if payload.Line != "Building main.go..." {
		t.Errorf("Expected line 'Building main.go...', got '%s'", payload.Line)
	}

	if payload.IsStderr {
		t.Error("Expected IsStderr to be false")
	}
}

func TestTestResultPayloadStructure(t *testing.T) {
	payload := TestResultPayload{
		TestName: "TestUserAuthentication",
		Passed:   true,
		Duration: 150,
		Error:    "",
	}

	if payload.TestName != "TestUserAuthentication" {
		t.Errorf("Expected test name 'TestUserAuthentication', got '%s'", payload.TestName)
	}

	if !payload.Passed {
		t.Error("Expected Passed to be true")
	}

	if payload.Duration != 150 {
		t.Errorf("Expected duration 150, got %d", payload.Duration)
	}
}

func TestErrorPayloadStructure(t *testing.T) {
	payload := ErrorPayload{
		Code:    "E001",
		Message: "Build failed",
		Details: "Missing dependency: github.com/example/pkg",
	}

	if payload.Code != "E001" {
		t.Errorf("Expected code 'E001', got '%s'", payload.Code)
	}

	if payload.Message != "Build failed" {
		t.Errorf("Expected message 'Build failed', got '%s'", payload.Message)
	}
}

func TestFileStartPayloadStructure(t *testing.T) {
	payload := FileStartPayload{
		FilePath:     "/src/handler.go",
		Language:     "go",
		EstimatedLOC: 200,
	}

	if payload.FilePath != "/src/handler.go" {
		t.Errorf("Expected file path '/src/handler.go', got '%s'", payload.FilePath)
	}

	if payload.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", payload.Language)
	}

	if payload.EstimatedLOC != 200 {
		t.Errorf("Expected estimated LOC 200, got %d", payload.EstimatedLOC)
	}
}

func TestBroadcastMessageStructure(t *testing.T) {
	event := NewEvent(EventPhaseChange, map[string]interface{}{
		"current_phase": "IMPLEMENTATION",
	})

	msg := &BroadcastMessage{
		WorkflowID: "workflow-abc",
		Event:      event,
	}

	if msg.WorkflowID != "workflow-abc" {
		t.Errorf("Expected workflow ID 'workflow-abc', got '%s'", msg.WorkflowID)
	}

	if msg.Event != event {
		t.Error("Expected event to match")
	}
}

func TestClientStructure(t *testing.T) {
	hub := NewHub()

	client := &Client{
		hub:        hub,
		conn:       nil, // Would be a real connection in practice
		send:       make(chan *StreamEvent, 256),
		workflowID: "workflow-xyz",
		sequence:   0,
	}

	if client.hub != hub {
		t.Error("Expected hub to match")
	}

	if client.workflowID != "workflow-xyz" {
		t.Errorf("Expected workflow ID 'workflow-xyz', got '%s'", client.workflowID)
	}

	if client.sequence != 0 {
		t.Errorf("Expected sequence 0, got %d", client.sequence)
	}
}

func TestHubConcurrentBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	time.Sleep(10 * time.Millisecond)

	// Broadcast concurrently from multiple goroutines
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			hub.BroadcastChatMessage("workflow-1", "agent", "Agent", "Message")
			done <- true
		}(i)
	}

	// Wait with timeout
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Good
		case <-time.After(1 * time.Second):
			t.Error("Concurrent broadcast timed out")
			return
		}
	}
}
