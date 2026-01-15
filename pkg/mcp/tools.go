// Package mcp implements the Model Context Protocol for tool integration.
// MCP provides a standardized way for LLM agents to interact with tools.
// Based on Anthropic's MCP specification (open-sourced Nov 2024, donated to Agentic AI Foundation Dec 2025).
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// ToolType defines the category of tool
type ToolType string

const (
	ToolTypeFile     ToolType = "file"
	ToolTypeCommand  ToolType = "command"
	ToolTypeSearch   ToolType = "search"
	ToolTypeGit      ToolType = "git"
	ToolTypeAnalysis ToolType = "analysis"
	ToolTypeCustom   ToolType = "custom"
)

// ParameterType defines the type of a parameter
type ParameterType string

const (
	ParamTypeString  ParameterType = "string"
	ParamTypeNumber  ParameterType = "number"
	ParamTypeBoolean ParameterType = "boolean"
	ParamTypeArray   ParameterType = "array"
	ParamTypeObject  ParameterType = "object"
)

// Parameter defines a tool parameter
type Parameter struct {
	Name        string        `json:"name"`
	Type        ParameterType `json:"type"`
	Description string        `json:"description"`
	Required    bool          `json:"required"`
	Default     interface{}   `json:"default,omitempty"`
	Enum        []string      `json:"enum,omitempty"`
}

// ToolDefinition defines a tool according to MCP specification
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        ToolType    `json:"type"`
	Parameters  []Parameter `json:"parameters"`
	Returns     string      `json:"returns"`
	Examples    []Example   `json:"examples,omitempty"`
}

// Example shows how to use a tool
type Example struct {
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Output      string                 `json:"output"`
}

// ToolRequest is a request to execute a tool (JSON-RPC style)
type ToolRequest struct {
	ID      string                 `json:"id"`
	Tool    string                 `json:"tool"`
	Params  map[string]interface{} `json:"params"`
	Context *ToolContext           `json:"context,omitempty"`
}

// ToolContext provides execution context for the tool
type ToolContext struct {
	ProjectID    string `json:"project_id,omitempty"`
	WorkDir      string `json:"work_dir,omitempty"`
	AgentID      string `json:"agent_id,omitempty"`
	SandboxLevel string `json:"sandbox_level,omitempty"`
}

// ToolResponse is the result of a tool execution
type ToolResponse struct {
	ID      string      `json:"id"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   *ToolError  `json:"error,omitempty"`
}

// ToolError represents an error from tool execution
type ToolError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Standard MCP error codes
const (
	ErrCodeInvalidParams    = -32602
	ErrCodeToolNotFound     = -32601
	ErrCodeExecutionFailed  = -32000
	ErrCodePermissionDenied = -32001
	ErrCodeTimeout          = -32002
)

// ToolHandler is the function signature for tool implementation
type ToolHandler func(ctx context.Context, req *ToolRequest) *ToolResponse

// ToolRegistry manages tool registration and execution
type ToolRegistry struct {
	tools map[string]*registeredTool
	mu    sync.RWMutex
}

type registeredTool struct {
	definition *ToolDefinition
	handler    ToolHandler
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]*registeredTool),
	}
}

// Register adds a tool to the registry
func (r *ToolRegistry) Register(def *ToolDefinition, handler ToolHandler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if def.Name == "" {
		return fmt.Errorf("tool name is required")
	}

	if handler == nil {
		return fmt.Errorf("tool handler is required")
	}

	r.tools[def.Name] = &registeredTool{
		definition: def,
		handler:    handler,
	}

	return nil
}

// Execute runs a tool by name
func (r *ToolRegistry) Execute(ctx context.Context, req *ToolRequest) *ToolResponse {
	r.mu.RLock()
	tool, ok := r.tools[req.Tool]
	r.mu.RUnlock()

	if !ok {
		return &ToolResponse{
			ID:      req.ID,
			Success: false,
			Error: &ToolError{
				Code:    ErrCodeToolNotFound,
				Message: fmt.Sprintf("tool not found: %s", req.Tool),
			},
		}
	}

	// Validate parameters
	if err := r.validateParams(tool.definition, req.Params); err != nil {
		return &ToolResponse{
			ID:      req.ID,
			Success: false,
			Error: &ToolError{
				Code:    ErrCodeInvalidParams,
				Message: err.Error(),
			},
		}
	}

	// Execute the handler
	return tool.handler(ctx, req)
}

// validateParams checks that required parameters are present
func (r *ToolRegistry) validateParams(def *ToolDefinition, params map[string]interface{}) error {
	for _, param := range def.Parameters {
		if param.Required {
			if _, ok := params[param.Name]; !ok {
				return fmt.Errorf("missing required parameter: %s", param.Name)
			}
		}
	}
	return nil
}

// GetTool returns a tool definition by name
func (r *ToolRegistry) GetTool(name string) (*ToolDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	if !ok {
		return nil, false
	}
	return tool.definition, true
}

// ListTools returns all registered tool definitions
func (r *ToolRegistry) ListTools() []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]*ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		defs = append(defs, tool.definition)
	}
	return defs
}

// ListToolNames returns names of all registered tools
func (r *ToolRegistry) ListToolNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// GenerateToolPrompt creates a prompt listing available tools for the LLM
func (r *ToolRegistry) GenerateToolPrompt() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prompt := "You have access to the following tools:\n\n"

	for name, tool := range r.tools {
		prompt += fmt.Sprintf("## %s\n", name)
		prompt += fmt.Sprintf("%s\n\n", tool.definition.Description)

		if len(tool.definition.Parameters) > 0 {
			prompt += "Parameters:\n"
			for _, param := range tool.definition.Parameters {
				required := ""
				if param.Required {
					required = " (required)"
				}
				prompt += fmt.Sprintf("- %s (%s)%s: %s\n",
					param.Name, param.Type, required, param.Description)
			}
		}
		prompt += "\n"
	}

	prompt += `To use a tool, respond with a JSON block like:
{"tool": "tool_name", "params": {"param1": "value1"}}
`

	return prompt
}

// GenerateOpenAIFunctions converts tools to OpenAI function calling format
func (r *ToolRegistry) GenerateOpenAIFunctions() []map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	functions := make([]map[string]interface{}, 0, len(r.tools))

	for _, tool := range r.tools {
		properties := make(map[string]interface{})
		required := make([]string, 0)

		for _, param := range tool.definition.Parameters {
			propDef := map[string]interface{}{
				"type":        string(param.Type),
				"description": param.Description,
			}
			if len(param.Enum) > 0 {
				propDef["enum"] = param.Enum
			}
			properties[param.Name] = propDef

			if param.Required {
				required = append(required, param.Name)
			}
		}

		fn := map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.definition.Name,
				"description": tool.definition.Description,
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": properties,
					"required":   required,
				},
			},
		}

		functions = append(functions, fn)
	}

	return functions
}

// ParseToolCall parses a tool call from LLM output
func ParseToolCall(output string) (*ToolRequest, error) {
	// Try to extract JSON from the output
	output = extractJSONFromText(output)

	var call struct {
		Tool   string                 `json:"tool"`
		Params map[string]interface{} `json:"params"`
	}

	if err := json.Unmarshal([]byte(output), &call); err != nil {
		return nil, fmt.Errorf("failed to parse tool call: %w", err)
	}

	if call.Tool == "" {
		return nil, fmt.Errorf("tool name is required")
	}

	return &ToolRequest{
		Tool:   call.Tool,
		Params: call.Params,
	}, nil
}

// extractJSONFromText tries to find and extract JSON from text
func extractJSONFromText(text string) string {
	// Find first { and last }
	start := -1
	end := -1
	depth := 0

	for i, c := range text {
		if c == '{' {
			if start == -1 {
				start = i
			}
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}

	if start >= 0 && end > start {
		return text[start:end]
	}

	return text
}
