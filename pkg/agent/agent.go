// Package agent provides the core agent system for Sovereign Firm.
// It implements dynamic agent spawning, lifecycle management, and inter-agent communication.
package agent

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// AgentStatus represents the current lifecycle state of an agent
type AgentStatus string

const (
	StatusSpawning   AgentStatus = "SPAWNING"   // Agent is being initialized
	StatusReady      AgentStatus = "READY"      // Agent is ready for work
	StatusWorking    AgentStatus = "WORKING"    // Agent is processing a task
	StatusBlocked    AgentStatus = "BLOCKED"    // Agent is waiting on another agent
	StatusReviewing  AgentStatus = "REVIEWING"  // Agent is reviewing work
	StatusCompleted  AgentStatus = "COMPLETED"  // Agent has finished all tasks
	StatusTerminated AgentStatus = "TERMINATED" // Agent has been terminated
	StatusError      AgentStatus = "ERROR"      // Agent encountered an error
)

// AgentContext holds the contextual information available to an agent
type AgentContext struct {
	// Project context
	ProjectID      string `json:"project_id"`
	ProjectSpec    string `json:"project_spec"`
	ConversationID string `json:"conversation_id"`

	// Knowledge
	CodebaseIndex   string            `json:"codebase_index,omitempty"` // Reference to indexed codebase
	RelevantFiles   map[string]string `json:"relevant_files,omitempty"` // Key files for this agent
	SharedArtifacts []Artifact        `json:"shared_artifacts,omitempty"`

	// Constraints
	MaxTokens int        `json:"max_tokens,omitempty"`
	Deadline  *time.Time `json:"deadline,omitempty"`
}

// Artifact represents a file or other deliverable produced by an agent
type Artifact struct {
	ID          string    `json:"id"`
	AgentID     string    `json:"agent_id"`
	Type        string    `json:"type"`               // "file", "schema", "config", etc.
	Path        string    `json:"path"`               // File path or identifier
	Content     string    `json:"content"`            // The actual content
	Language    string    `json:"language,omitempty"` // Programming language if applicable
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Task represents a unit of work assigned to an agent
type Task struct {
	ID           string                 `json:"id"`
	Description  string                 `json:"description"`
	Type         string                 `json:"type"` // "code_generate", "code_review", "test", etc.
	Priority     int                    `json:"priority"`
	Dependencies []string               `json:"dependencies,omitempty"` // Task IDs that must complete first
	Status       string                 `json:"status"`
	AssignedTo   string                 `json:"assigned_to,omitempty"`
	Input        map[string]interface{} `json:"input,omitempty"`
	Output       map[string]interface{} `json:"output,omitempty"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
}

// MessageType defines the type of inter-agent message
type MessageType string

const (
	MessageQuestion  MessageType = "QUESTION"
	MessageAnswer    MessageType = "ANSWER"
	MessageArtifact  MessageType = "ARTIFACT"
	MessageStatus    MessageType = "STATUS"
	MessageRequest   MessageType = "REQUEST"
	MessageReview    MessageType = "REVIEW"
	MessageBroadcast MessageType = "BROADCAST"
)

// Message represents an inter-agent communication
type Message struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	From      string                 `json:"from"` // Agent ID
	To        string                 `json:"to"`   // Agent ID or "*" for broadcast
	Subject   string                 `json:"subject"`
	Content   interface{}            `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	ReplyTo   string                 `json:"reply_to,omitempty"` // Message ID if this is a reply
}

// Skill represents an injectable capability for an agent
type Skill struct {
	Name         string   `json:"name" yaml:"name"`
	Version      string   `json:"version" yaml:"version"`
	Description  string   `json:"description" yaml:"description"`
	Capabilities []string `json:"capabilities" yaml:"capabilities"`
	Tools        []string `json:"tools" yaml:"tools"`
	SystemPrompt string   `json:"system_prompt" yaml:"system_prompt"`
}

// AgentInstance represents a running agent in the system
type AgentInstance struct {
	// Identity
	ID   string `json:"id"`
	Name string `json:"name"` // Human-friendly name (e.g., "Alice")
	Role string `json:"role"` // Role identifier (e.g., "senior-frontend-dev")

	// Configuration
	Skills       []Skill `json:"skills"`
	SystemPrompt string  `json:"-"`           // Compiled from role + skills + context
	Model        string  `json:"model"`       // Which LLM model to use
	Temperature  float64 `json:"temperature"` // LLM temperature parameter

	// Project Context
	ProjectID string        `json:"project_id"`
	Context   *AgentContext `json:"context"`

	// Lifecycle
	Status       AgentStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	TerminatedAt *time.Time  `json:"terminated_at,omitempty"`

	// Work
	CurrentTask *Task      `json:"current_task,omitempty"`
	Artifacts   []Artifact `json:"artifacts"`

	// Communication channels
	Inbox  chan Message `json:"-"` // Receive messages
	Outbox chan Message `json:"-"` // Send messages

	// Metrics
	TokensUsed    int64 `json:"tokens_used"`
	TasksComplete int   `json:"tasks_complete"`

	// Internal
	llmClient llm.Client
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewAgentInstance creates a new agent instance with the given configuration
func NewAgentInstance(
	name, role, projectID, model string,
	skills []Skill,
	llmClient llm.Client,
) *AgentInstance {
	ctx, cancel := context.WithCancel(context.Background())

	agent := &AgentInstance{
		ID:            uuid.New().String(),
		Name:          name,
		Role:          role,
		Skills:        skills,
		Model:         model,
		Temperature:   0.7, // Default temperature
		ProjectID:     projectID,
		Context:       &AgentContext{ProjectID: projectID},
		Status:        StatusSpawning,
		CreatedAt:     time.Now(),
		Artifacts:     make([]Artifact, 0),
		Inbox:         make(chan Message, 100),
		Outbox:        make(chan Message, 100),
		TokensUsed:    0,
		TasksComplete: 0,
		llmClient:     llmClient,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Compile system prompt from role and skills
	agent.compileSystemPrompt()

	return agent
}

// compileSystemPrompt builds the system prompt from role and skills
func (a *AgentInstance) compileSystemPrompt() {
	prompt := ""

	// Add role introduction
	prompt += "You are " + a.Name + ", a " + a.Role + " on this project.\n\n"

	// Add skill-specific prompts
	for _, skill := range a.Skills {
		if skill.SystemPrompt != "" {
			prompt += skill.SystemPrompt + "\n\n"
		}
	}

	a.SystemPrompt = prompt
}

// Spawn initializes and starts the agent
func (a *AgentInstance) Spawn(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Status != StatusSpawning {
		return ErrInvalidStatusTransition
	}

	// Start message processing
	go a.processMessages()

	a.Status = StatusReady
	return nil
}

// AssignTask assigns a task to the agent
func (a *AgentInstance) AssignTask(task *Task) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Status != StatusReady {
		return ErrAgentBusy
	}

	a.CurrentTask = task
	a.Status = StatusWorking
	now := time.Now()
	task.StartedAt = &now
	task.Status = "IN_PROGRESS"

	return nil
}

// CompleteTask marks the current task as complete
func (a *AgentInstance) CompleteTask(output map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.CurrentTask == nil {
		return ErrNoActiveTask
	}

	now := time.Now()
	a.CurrentTask.CompletedAt = &now
	a.CurrentTask.Status = "COMPLETED"
	a.CurrentTask.Output = output

	a.TasksComplete++
	a.CurrentTask = nil
	a.Status = StatusReady

	return nil
}

// AddArtifact records an artifact produced by this agent
func (a *AgentInstance) AddArtifact(artifactType, path, content, language, description string) *Artifact {
	a.mu.Lock()
	defer a.mu.Unlock()

	artifact := Artifact{
		ID:          uuid.New().String(),
		AgentID:     a.ID,
		Type:        artifactType,
		Path:        path,
		Content:     content,
		Language:    language,
		Description: description,
		CreatedAt:   time.Now(),
	}

	a.Artifacts = append(a.Artifacts, artifact)
	return &artifact
}

// SendMessage sends a message to another agent
func (a *AgentInstance) SendMessage(msgType MessageType, to, subject string, content interface{}) {
	msg := Message{
		ID:        uuid.New().String(),
		Type:      msgType,
		From:      a.ID,
		To:        to,
		Subject:   subject,
		Content:   content,
		Timestamp: time.Now(),
	}

	select {
	case a.Outbox <- msg:
	default:
		// Outbox full, log warning
	}
}

// processMessages handles incoming messages
func (a *AgentInstance) processMessages() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case msg := <-a.Inbox:
			a.handleMessage(msg)
		}
	}
}

// handleMessage processes a single incoming message
func (a *AgentInstance) handleMessage(msg Message) {
	// TODO: Implement message handling logic based on type
	switch msg.Type {
	case MessageQuestion:
		// Handle question
	case MessageRequest:
		// Handle request
	case MessageReview:
		// Handle review request
	case MessageArtifact:
		// Store shared artifact
	default:
		// Log unknown message type
	}
}

// Terminate shuts down the agent
func (a *AgentInstance) Terminate() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Status == StatusTerminated {
		return
	}

	a.cancel()
	close(a.Inbox)
	close(a.Outbox)

	now := time.Now()
	a.TerminatedAt = &now
	a.Status = StatusTerminated
}

// GetStatus returns the current agent status
func (a *AgentInstance) GetStatus() AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Status
}

// GetMetrics returns agent metrics
func (a *AgentInstance) GetMetrics() (tokensUsed int64, tasksComplete int) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.TokensUsed, a.TasksComplete
}

// IsAlive returns true if the agent is not terminated
func (a *AgentInstance) IsAlive() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Status != StatusTerminated && a.Status != StatusError
}
