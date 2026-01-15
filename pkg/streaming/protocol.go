package streaming

import "time"

// StreamEventType defines the type of streaming event
type StreamEventType string

const (
	// Connection events
	EventConnected    StreamEventType = "CONNECTED"
	EventDisconnected StreamEventType = "DISCONNECTED"
	EventError        StreamEventType = "ERROR"

	// Workflow lifecycle events
	EventPhaseChange StreamEventType = "PHASE_CHANGE"
	EventAgentSpawn  StreamEventType = "AGENT_SPAWN"
	EventAgentDone   StreamEventType = "AGENT_DONE"

	// Code generation events (chunked)
	EventCodeChunk    StreamEventType = "CODE_CHUNK"
	EventCodeComplete StreamEventType = "CODE_COMPLETE"
	EventFileStart    StreamEventType = "FILE_START"
	EventFileEnd      StreamEventType = "FILE_END"

	// Chat events
	EventChatMessage StreamEventType = "CHAT_MESSAGE"
	EventChatTyping  StreamEventType = "CHAT_TYPING"

	// Build/Test events
	EventBuildStart   StreamEventType = "BUILD_START"
	EventBuildOutput  StreamEventType = "BUILD_OUTPUT"
	EventBuildSuccess StreamEventType = "BUILD_SUCCESS"
	EventBuildError   StreamEventType = "BUILD_ERROR"
	EventTestStart    StreamEventType = "TEST_START"
	EventTestResult   StreamEventType = "TEST_RESULT"

	// Preview events
	EventPreviewReady StreamEventType = "PREVIEW_READY"
)

// StreamEvent is the base event structure sent over WebSocket
type StreamEvent struct {
	Type      StreamEventType        `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Sequence  int64                  `json:"seq"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// NewEvent creates a new stream event
func NewEvent(eventType StreamEventType, payload map[string]interface{}) *StreamEvent {
	return &StreamEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}

// PhaseChangePayload for PHASE_CHANGE events
type PhaseChangePayload struct {
	PreviousPhase string `json:"previous_phase"`
	CurrentPhase  string `json:"current_phase"`
	Message       string `json:"message,omitempty"`
}

// CodeChunkPayload for CODE_CHUNK events (incremental code streaming)
type CodeChunkPayload struct {
	FilePath    string `json:"file_path"`
	ChunkIndex  int    `json:"chunk_index"`
	Content     string `json:"content"`
	IsComplete  bool   `json:"is_complete"`
	TotalChunks int    `json:"total_chunks,omitempty"`
}

// FileStartPayload for FILE_START events
type FileStartPayload struct {
	FilePath     string `json:"file_path"`
	Language     string `json:"language"`
	EstimatedLOC int    `json:"estimated_loc,omitempty"`
}

// ChatMessagePayload for CHAT_MESSAGE events
type ChatMessagePayload struct {
	Role    string `json:"role"`            // "user", "agent", "system"
	Agent   string `json:"agent,omitempty"` // "PM", "Dev", "QA", etc.
	Content string `json:"content"`
}

// AgentPayload for AGENT_SPAWN and AGENT_DONE events
type AgentPayload struct {
	AgentID   string   `json:"agent_id"`
	AgentType string   `json:"agent_type"`
	Skills    []string `json:"skills,omitempty"`
	TaskID    string   `json:"task_id,omitempty"`
	Result    string   `json:"result,omitempty"`
}

// BuildOutputPayload for BUILD_OUTPUT events
type BuildOutputPayload struct {
	Line     string `json:"line"`
	IsStderr bool   `json:"is_stderr"`
}

// TestResultPayload for TEST_RESULT events
type TestResultPayload struct {
	TestName string `json:"test_name"`
	Passed   bool   `json:"passed"`
	Duration int    `json:"duration_ms"`
	Error    string `json:"error,omitempty"`
}

// ErrorPayload for ERROR events
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
