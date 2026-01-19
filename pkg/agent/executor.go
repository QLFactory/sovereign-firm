package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/mcp"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// AgentExecutor handles task execution with tool usage
type AgentExecutor struct {
	agent        *AgentInstance
	toolRegistry *mcp.ToolRegistry
	socValidator *SOCValidator
	maxToolCalls int
}

// NewAgentExecutor creates an executor for an agent
func NewAgentExecutor(
	agent *AgentInstance,
	toolRegistry *mcp.ToolRegistry,
	llmClient llm.Client,
) *AgentExecutor {
	return &AgentExecutor{
		agent:        agent,
		toolRegistry: toolRegistry,
		socValidator: NewSOCValidator(llmClient, 3),
		maxToolCalls: 10, // Prevent infinite tool loops
	}
}

// ExecuteTask runs a task with tool usage support
func (e *AgentExecutor) ExecuteTask(ctx context.Context, task *Task) (*TaskResult, error) {
	// Assign task to agent
	if err := e.agent.AssignTask(task); err != nil {
		return nil, err
	}

	// Build the prompt with tool instructions
	prompt := e.buildTaskPrompt(task)

	// Execution loop - agent can use tools or provide final answer
	var finalResult *TaskResult
	toolCallCount := 0
	conversationHistory := []string{prompt}

	for toolCallCount < e.maxToolCalls {
		// Call LLM
		resp, err := e.agent.llmClient.Generate(ctx, llm.GenerateRequest{
			Prompt: strings.Join(conversationHistory, "\n\n"),
			System: e.agent.SystemPrompt + "\n\n" + e.toolRegistry.GenerateToolPrompt(),
		})
		if err != nil {
			return nil, fmt.Errorf("LLM generation failed: %w", err)
		}

		// Track token usage
		e.agent.mu.Lock()
		e.agent.TokensUsed += int64(len(resp.Response) / 4) // Rough approximation
		e.agent.mu.Unlock()

		// Check if this is a tool call
		if toolCall := e.extractToolCall(resp.Response); toolCall != nil {
			toolCallCount++

			// Check for new messages before executing tool
			e.checkForNewMessages(&conversationHistory)

			// Execute the tool
			toolResp := e.toolRegistry.Execute(ctx, toolCall)

			// Add tool result to conversation
			toolResultStr := mcp.FormatToolResult(toolResp)
			conversationHistory = append(conversationHistory,
				fmt.Sprintf("Tool: %s\nResult:\n%s", toolCall.Tool, toolResultStr))

			continue
		}

		// This is the final answer
		finalResult = &TaskResult{
			Success:   true,
			Output:    resp.Response,
			ToolCalls: toolCallCount,
		}
		break
	}

	if finalResult == nil {
		finalResult = &TaskResult{
			Success:   false,
			Error:     "max tool calls exceeded",
			ToolCalls: toolCallCount,
		}
	}

	// Complete the task
	output := map[string]interface{}{
		"result":     finalResult.Output,
		"tool_calls": finalResult.ToolCalls,
	}
	if err := e.agent.CompleteTask(output); err != nil {
		return nil, err
	}

	return finalResult, nil
}

// WaitForReply waits for a reply to a message sent to another agent
func (e *AgentExecutor) WaitForReply(ctx context.Context, replyToID string) (*Message, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case msg := <-e.agent.Inbox:
			if msg.ReplyTo == replyToID {
				return &msg, nil
			}
			// Re-inject other messages into handleMessage
			e.agent.handleMessage(msg)
		}
	}
}

// TaskResult holds the outcome of task execution
type TaskResult struct {
	Success   bool        `json:"success"`
	Output    interface{} `json:"output,omitempty"`
	Error     string      `json:"error,omitempty"`
	ToolCalls int         `json:"tool_calls"`
	Artifacts []Artifact  `json:"artifacts,omitempty"`
}

// buildTaskPrompt constructs the prompt for a task
func (e *AgentExecutor) buildTaskPrompt(task *Task) string {
	prompt := fmt.Sprintf("## Task: %s\n\n", task.Description)

	if task.Type != "" {
		prompt += fmt.Sprintf("Task Type: %s\n\n", task.Type)
	}

	if task.Input != nil {
		inputJSON, _ := json.MarshalIndent(task.Input, "", "  ")
		prompt += fmt.Sprintf("Input:\n```json\n%s\n```\n\n", string(inputJSON))
	}

	// Add context from agent
	if e.agent.Context != nil {
		if e.agent.Context.ProjectSpec != "" {
			prompt += fmt.Sprintf("Project Specification:\n%s\n\n", e.agent.Context.ProjectSpec)
		}

		if len(e.agent.Context.RelevantFiles) > 0 {
			prompt += "Relevant Files:\n"
			for path, content := range e.agent.Context.RelevantFiles {
				prompt += fmt.Sprintf("### %s\n```\n%s\n```\n\n", path, content)
			}
		}
	}

	// Add pending inter-agent messages
	e.agent.mu.RLock()
	if len(e.agent.PendingMessages) > 0 {
		prompt += "## Pending Inter-Agent Requests\n"
		prompt += "The following agents have requested information or action from you:\n\n"
		for _, msg := range e.agent.PendingMessages {
			prompt += fmt.Sprintf("- From: %s\n  Subject: %s\n  Message: %v\n\n", msg.From, msg.Subject, msg.Content)
		}
		prompt += "Please address these requests if they are relevant to your current task or if you can answer them quickly.\n\n"
	}
	e.agent.mu.RUnlock()

	prompt += `Please complete this task. You can use the available tools if needed.
When you are done, provide your final answer.

If you need to create or modify files, use the file_write tool.
If you need to read existing files, use the file_read tool.
If you need to run commands, use the shell_exec or npm_run tools.
`

	return prompt
}

// checkForNewMessages checks for and injects new inter-agent messages into the conversation history.
// Uses non-blocking receives to avoid deadlock - never holds mutex while waiting on channel.
func (e *AgentExecutor) checkForNewMessages(history *[]string) {
	// Drain inbox using non-blocking receives to prevent deadlock.
	// The lock is only held briefly per message, not during channel operations.
	for {
		select {
		case msg := <-e.agent.Inbox:
			// Got a message - now acquire lock briefly to update state
			e.agent.mu.Lock()
			log.Printf("Agent %s (%s) received asynchronous message: %s from %s", e.agent.Name, e.agent.ID, msg.Type, msg.From)
			switch msg.Type {
			case MessageQuestion, MessageRequest:
				e.agent.PendingMessages = append(e.agent.PendingMessages, msg)
				// Inject immediately into history
				*history = append(*history, fmt.Sprintf("System Notification: New message received from %s\nSubject: %s\nContent: %v", msg.From, msg.Subject, msg.Content))
			case MessageArtifact:
				if artifact, ok := msg.Content.(Artifact); ok {
					e.agent.Artifacts = append(e.agent.Artifacts, artifact)
					if e.agent.Context != nil {
						e.agent.Context.SharedArtifacts = append(e.agent.Context.SharedArtifacts, artifact)
					}
					*history = append(*history, fmt.Sprintf("System Notification: Agent %s shared an artifact: %s", msg.From, artifact.Path))
				}
			}
			e.agent.mu.Unlock()
		default:
			// No more messages available, exit without blocking
			return
		}
	}
}

// extractToolCall tries to parse a tool call from LLM output
func (e *AgentExecutor) extractToolCall(output string) *mcp.ToolRequest {
	// Look for JSON blocks that might be tool calls
	call, err := mcp.ParseToolCall(output)
	if err != nil {
		return nil
	}

	// Verify the tool exists
	if _, exists := e.toolRegistry.GetTool(call.Tool); !exists {
		return nil
	}

	return call
}

// ExecuteWithSOC executes a task and validates output against a contract
func (e *AgentExecutor) ExecuteWithSOC(
	ctx context.Context,
	task *Task,
	contract *SOCContract,
) (*TaskResult, *SOCResult, error) {
	// Execute the task
	result, err := e.ExecuteTask(ctx, task)
	if err != nil {
		return nil, nil, err
	}

	if !result.Success {
		return result, nil, nil
	}

	// Validate against SOC
	outputStr, ok := result.Output.(string)
	if !ok {
		outputJSON, _ := json.Marshal(result.Output)
		outputStr = string(outputJSON)
	}

	socResult, err := e.socValidator.ValidateAndRetry(
		ctx,
		contract,
		outputStr,
		e.buildTaskPrompt(task),
		e.agent.SystemPrompt,
	)

	if err != nil {
		return result, socResult, err
	}

	// Update result with validated output
	if socResult.Valid && socResult.Output != nil {
		result.Output = socResult.Output
	}

	return result, socResult, nil
}

// GetAllowedTools returns the list of tools this agent is allowed to use
func (e *AgentExecutor) GetAllowedTools() []string {
	allowed := make(map[string]bool)

	for _, skill := range e.agent.Skills {
		for _, tool := range skill.Tools {
			allowed[tool] = true
		}
	}

	result := make([]string, 0, len(allowed))
	for tool := range allowed {
		result = append(result, tool)
	}
	return result
}
