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

// ==================== Client Registration Tests ====================

func TestHubClientRegistration(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	// Create a mock client
	client := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 256),
		workflowID: "test-workflow",
		sequence:   0,
	}

	// Register client
	hub.register <- client

	// Wait for registration
	time.Sleep(10 * time.Millisecond)

	// Verify client is registered
	count := hub.GetClientCount("test-workflow")
	if count != 1 {
		t.Errorf("Expected 1 client, got %d", count)
	}

	// Should receive CONNECTED event
	select {
	case event := <-client.send:
		if event.Type != EventConnected {
			t.Errorf("Expected CONNECTED event, got %s", event.Type)
		}
		if event.Payload["workflow_id"] != "test-workflow" {
			t.Errorf("Expected workflow_id 'test-workflow', got %v", event.Payload["workflow_id"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for CONNECTED event")
	}
}

func TestHubClientUnregistration(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 256),
		workflowID: "test-workflow",
		sequence:   0,
	}

	// Register
	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Drain CONNECTED event
	<-client.send

	if hub.GetClientCount("test-workflow") != 1 {
		t.Error("Client should be registered")
	}

	// Unregister
	hub.unregister <- client
	time.Sleep(10 * time.Millisecond)

	if hub.GetClientCount("test-workflow") != 0 {
		t.Error("Client should be unregistered")
	}
}

func TestHubMultipleClientsOnSameWorkflow(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	workflowID := "shared-workflow"

	// Create 3 clients for same workflow
	clients := make([]*Client, 3)
	for i := 0; i < 3; i++ {
		clients[i] = &Client{
			hub:        hub,
			conn:       nil,
			send:       make(chan *StreamEvent, 256),
			workflowID: workflowID,
			sequence:   0,
		}
		hub.register <- clients[i]
	}

	time.Sleep(20 * time.Millisecond)

	// Verify all connected
	count := hub.GetClientCount(workflowID)
	if count != 3 {
		t.Errorf("Expected 3 clients, got %d", count)
	}

	// Drain CONNECTED events
	for _, c := range clients {
		<-c.send
	}

	// Broadcast a message
	hub.BroadcastChatMessage(workflowID, "agent", "PM", "Hello everyone!")
	time.Sleep(10 * time.Millisecond)

	// All clients should receive the message
	for i, c := range clients {
		select {
		case event := <-c.send:
			if event.Type != EventChatMessage {
				t.Errorf("Client %d: Expected CHAT_MESSAGE, got %s", i, event.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Client %d: Timeout waiting for message", i)
		}
	}
}

func TestHubWorkflowIsolation(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	// Create clients for different workflows
	client1 := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 256),
		workflowID: "workflow-A",
		sequence:   0,
	}
	client2 := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 256),
		workflowID: "workflow-B",
		sequence:   0,
	}

	hub.register <- client1
	hub.register <- client2
	time.Sleep(20 * time.Millisecond)

	// Drain CONNECTED events
	<-client1.send
	<-client2.send

	// Broadcast to workflow-A only
	hub.BroadcastChatMessage("workflow-A", "agent", "PM", "Message for A")
	time.Sleep(10 * time.Millisecond)

	// Client1 should receive message
	select {
	case event := <-client1.send:
		if event.Type != EventChatMessage {
			t.Errorf("Client1: Expected CHAT_MESSAGE, got %s", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client1 should have received message")
	}

	// Client2 should NOT receive message
	select {
	case event := <-client2.send:
		t.Errorf("Client2 should not receive workflow-A message, got %s", event.Type)
	case <-time.After(50 * time.Millisecond):
		// Good - no message received
	}
}

// ==================== Event Sequence Tests ====================

func TestEventSequenceNumbers(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 256),
		workflowID: "seq-test",
		sequence:   0,
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Drain CONNECTED event (seq 0)
	<-client.send

	// Send 5 messages
	for i := 0; i < 5; i++ {
		hub.BroadcastChatMessage("seq-test", "agent", "PM", "Message")
	}
	time.Sleep(20 * time.Millisecond)

	// Verify sequence numbers are incrementing
	var lastSeq int64 = 0
	for i := 0; i < 5; i++ {
		select {
		case event := <-client.send:
			if event.Sequence <= lastSeq {
				t.Errorf("Sequence should increment: got %d after %d", event.Sequence, lastSeq)
			}
			lastSeq = event.Sequence
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timeout waiting for message %d", i)
			return
		}
	}

	if lastSeq != 5 {
		t.Errorf("Expected final sequence 5, got %d", lastSeq)
	}
}

// ==================== Backpressure Tests ====================

func TestSlowClientHandling(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	// Create a client with small buffer (simulates slow reader)
	slowClient := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 2), // Very small buffer
		workflowID: "backpressure-test",
		sequence:   0,
	}

	hub.register <- slowClient
	time.Sleep(10 * time.Millisecond)

	// Don't drain any messages - simulate slow client

	// Flood with messages (more than buffer can hold)
	for i := 0; i < 10; i++ {
		hub.BroadcastChatMessage("backpressure-test", "agent", "PM", "Flood message")
	}
	time.Sleep(50 * time.Millisecond)

	// Slow client should be disconnected (channel closed)
	// After flooding, the client count should be 0
	count := hub.GetClientCount("backpressure-test")
	if count != 0 {
		t.Errorf("Slow client should be disconnected, but count is %d", count)
	}
}

// ==================== Broadcast Helper Tests ====================

func TestBroadcastPhaseChangeContent(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 256),
		workflowID: "phase-test",
		sequence:   0,
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)
	<-client.send // Drain CONNECTED

	hub.BroadcastPhaseChange("phase-test", "DISCOVERY", "IMPLEMENTATION", "Starting implementation")
	time.Sleep(10 * time.Millisecond)

	select {
	case event := <-client.send:
		if event.Type != EventPhaseChange {
			t.Errorf("Expected PHASE_CHANGE, got %s", event.Type)
		}
		if event.Payload["previous_phase"] != "DISCOVERY" {
			t.Errorf("Expected previous_phase DISCOVERY, got %v", event.Payload["previous_phase"])
		}
		if event.Payload["current_phase"] != "IMPLEMENTATION" {
			t.Errorf("Expected current_phase IMPLEMENTATION, got %v", event.Payload["current_phase"])
		}
		if event.Payload["message"] != "Starting implementation" {
			t.Errorf("Expected message, got %v", event.Payload["message"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for phase change")
	}
}

func TestBroadcastCodeChunkContent(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 256),
		workflowID: "code-test",
		sequence:   0,
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)
	<-client.send // Drain CONNECTED

	hub.BroadcastCodeChunk("code-test", "/src/App.jsx", 0, "import React from 'react';", false)
	time.Sleep(10 * time.Millisecond)

	select {
	case event := <-client.send:
		if event.Type != EventCodeChunk {
			t.Errorf("Expected CODE_CHUNK, got %s", event.Type)
		}
		if event.Payload["file_path"] != "/src/App.jsx" {
			t.Errorf("Expected file_path /src/App.jsx, got %v", event.Payload["file_path"])
		}
		// chunk_index comes back as float64 from map[string]interface{}
		if int(event.Payload["chunk_index"].(int)) != 0 {
			t.Errorf("Expected chunk_index 0, got %v", event.Payload["chunk_index"])
		}
		if event.Payload["content"] != "import React from 'react';" {
			t.Errorf("Expected content, got %v", event.Payload["content"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for code chunk")
	}
}

func TestBroadcastFileStartEnd(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 256),
		workflowID: "file-test",
		sequence:   0,
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)
	<-client.send // Drain CONNECTED

	// Test FILE_START
	hub.BroadcastFileStart("file-test", "/src/utils.js", "javascript")
	time.Sleep(10 * time.Millisecond)

	select {
	case event := <-client.send:
		if event.Type != EventFileStart {
			t.Errorf("Expected FILE_START, got %s", event.Type)
		}
		if event.Payload["file_path"] != "/src/utils.js" {
			t.Errorf("Expected file_path /src/utils.js, got %v", event.Payload["file_path"])
		}
		if event.Payload["language"] != "javascript" {
			t.Errorf("Expected language javascript, got %v", event.Payload["language"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for FILE_START")
	}

	// Test FILE_END
	hub.BroadcastFileEnd("file-test", "/src/utils.js")
	time.Sleep(10 * time.Millisecond)

	select {
	case event := <-client.send:
		if event.Type != EventFileEnd {
			t.Errorf("Expected FILE_END, got %s", event.Type)
		}
		if event.Payload["file_path"] != "/src/utils.js" {
			t.Errorf("Expected file_path /src/utils.js, got %v", event.Payload["file_path"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for FILE_END")
	}
}

func TestBroadcastErrorContent(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	client := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 256),
		workflowID: "error-test",
		sequence:   0,
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)
	<-client.send // Drain CONNECTED

	hub.BroadcastError("error-test", "BUILD_FAILED", "Compilation error", "Line 42: undefined variable")
	time.Sleep(10 * time.Millisecond)

	select {
	case event := <-client.send:
		if event.Type != EventError {
			t.Errorf("Expected ERROR, got %s", event.Type)
		}
		if event.Payload["code"] != "BUILD_FAILED" {
			t.Errorf("Expected code BUILD_FAILED, got %v", event.Payload["code"])
		}
		if event.Payload["message"] != "Compilation error" {
			t.Errorf("Expected message, got %v", event.Payload["message"])
		}
		if event.Payload["details"] != "Line 42: undefined variable" {
			t.Errorf("Expected details, got %v", event.Payload["details"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for ERROR")
	}
}

// ==================== Integration Test ====================

func TestFullBroadcastFlow(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	workflowID := "integration-test"

	// Create 2 clients
	client1 := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 256),
		workflowID: workflowID,
		sequence:   0,
	}
	client2 := &Client{
		hub:        hub,
		conn:       nil,
		send:       make(chan *StreamEvent, 256),
		workflowID: workflowID,
		sequence:   0,
	}

	hub.register <- client1
	hub.register <- client2
	time.Sleep(20 * time.Millisecond)

	// Drain CONNECTED events
	<-client1.send
	<-client2.send

	// Simulate a code generation flow
	events := []struct {
		broadcast func()
		eventType StreamEventType
	}{
		{func() { hub.BroadcastPhaseChange(workflowID, "DISCOVERY", "IMPLEMENTATION", "") }, EventPhaseChange},
		{func() { hub.BroadcastFileStart(workflowID, "/src/App.jsx", "jsx") }, EventFileStart},
		{func() { hub.BroadcastCodeChunk(workflowID, "/src/App.jsx", 0, "import React", false) }, EventCodeChunk},
		{func() { hub.BroadcastCodeChunk(workflowID, "/src/App.jsx", 1, "export default", true) }, EventCodeChunk},
		{func() { hub.BroadcastFileEnd(workflowID, "/src/App.jsx") }, EventFileEnd},
	}

	for _, e := range events {
		e.broadcast()
	}
	time.Sleep(30 * time.Millisecond)

	// Verify both clients received all events
	for clientIdx, client := range []*Client{client1, client2} {
		for i, expected := range events {
			select {
			case event := <-client.send:
				if event.Type != expected.eventType {
					t.Errorf("Client %d, Event %d: Expected %s, got %s",
						clientIdx, i, expected.eventType, event.Type)
				}
			case <-time.After(100 * time.Millisecond):
				t.Errorf("Client %d, Event %d: Timeout", clientIdx, i)
			}
		}
	}

	// Verify sequence numbers are correct
	if client1.sequence != 5 {
		t.Errorf("Client1 sequence should be 5, got %d", client1.sequence)
	}
	if client2.sequence != 5 {
		t.Errorf("Client2 sequence should be 5, got %d", client2.sequence)
	}
}
