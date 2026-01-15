package mcp

import (
	"context"
	"strings"
	"testing"
)

func TestNewToolRegistry(t *testing.T) {
	registry := NewToolRegistry()

	if registry == nil {
		t.Fatal("Expected registry to be created, got nil")
	}

	if len(registry.ListTools()) != 0 {
		t.Errorf("Expected empty registry, got %d tools", len(registry.ListTools()))
	}
}

func TestToolRegistryRegister(t *testing.T) {
	registry := NewToolRegistry()

	def := &ToolDefinition{
		Name:        "test_tool",
		Description: "A test tool",
		Type:        ToolTypeCustom,
		Parameters: []Parameter{
			{Name: "input", Type: ParamTypeString, Required: true},
		},
	}

	handler := func(ctx context.Context, req *ToolRequest) *ToolResponse {
		return &ToolResponse{ID: req.ID, Success: true, Result: "ok"}
	}

	err := registry.Register(def, handler)
	if err != nil {
		t.Fatalf("Expected registration to succeed, got error: %v", err)
	}

	tools := registry.ListTools()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
}

func TestToolRegistryRegisterValidation(t *testing.T) {
	registry := NewToolRegistry()

	// Test missing name
	err := registry.Register(&ToolDefinition{}, func(ctx context.Context, req *ToolRequest) *ToolResponse {
		return nil
	})
	if err == nil {
		t.Error("Expected error for missing tool name")
	}

	// Test missing handler
	err = registry.Register(&ToolDefinition{Name: "test"}, nil)
	if err == nil {
		t.Error("Expected error for missing handler")
	}
}

func TestToolRegistryExecute(t *testing.T) {
	registry := NewToolRegistry()

	def := &ToolDefinition{
		Name:        "echo",
		Description: "Echoes input back",
		Type:        ToolTypeCustom,
		Parameters: []Parameter{
			{Name: "message", Type: ParamTypeString, Required: true},
		},
	}

	handler := func(ctx context.Context, req *ToolRequest) *ToolResponse {
		msg := req.Params["message"].(string)
		return &ToolResponse{
			ID:      req.ID,
			Success: true,
			Result:  "Echo: " + msg,
		}
	}

	registry.Register(def, handler)

	req := &ToolRequest{
		ID:     "req-1",
		Tool:   "echo",
		Params: map[string]interface{}{"message": "Hello"},
	}

	resp := registry.Execute(context.Background(), req)

	if !resp.Success {
		t.Errorf("Expected success, got error: %v", resp.Error)
	}

	if resp.Result != "Echo: Hello" {
		t.Errorf("Expected 'Echo: Hello', got '%v'", resp.Result)
	}
}

func TestToolRegistryExecuteNotFound(t *testing.T) {
	registry := NewToolRegistry()

	req := &ToolRequest{
		ID:     "req-1",
		Tool:   "nonexistent",
		Params: map[string]interface{}{},
	}

	resp := registry.Execute(context.Background(), req)

	if resp.Success {
		t.Error("Expected failure for non-existent tool")
	}

	if resp.Error == nil {
		t.Error("Expected error to be set")
	}

	if resp.Error.Code != ErrCodeToolNotFound {
		t.Errorf("Expected error code %d, got %d", ErrCodeToolNotFound, resp.Error.Code)
	}
}

func TestToolRegistryExecuteMissingParam(t *testing.T) {
	registry := NewToolRegistry()

	def := &ToolDefinition{
		Name: "required_params",
		Type: ToolTypeCustom,
		Parameters: []Parameter{
			{Name: "required_field", Type: ParamTypeString, Required: true},
		},
	}

	handler := func(ctx context.Context, req *ToolRequest) *ToolResponse {
		return &ToolResponse{ID: req.ID, Success: true}
	}

	registry.Register(def, handler)

	// Missing required parameter
	req := &ToolRequest{
		ID:     "req-1",
		Tool:   "required_params",
		Params: map[string]interface{}{},
	}

	resp := registry.Execute(context.Background(), req)

	if resp.Success {
		t.Error("Expected failure for missing required parameter")
	}

	if resp.Error.Code != ErrCodeInvalidParams {
		t.Errorf("Expected error code %d, got %d", ErrCodeInvalidParams, resp.Error.Code)
	}
}

func TestToolRegistryGetTool(t *testing.T) {
	registry := NewToolRegistry()

	def := &ToolDefinition{
		Name:        "get_test",
		Description: "Test tool for get",
		Type:        ToolTypeFile,
	}

	registry.Register(def, func(ctx context.Context, req *ToolRequest) *ToolResponse {
		return nil
	})

	// Test existing tool
	retrieved, ok := registry.GetTool("get_test")
	if !ok {
		t.Error("Expected tool to be found")
	}
	if retrieved.Name != "get_test" {
		t.Errorf("Expected name 'get_test', got '%s'", retrieved.Name)
	}

	// Test non-existing tool
	_, ok = registry.GetTool("nonexistent")
	if ok {
		t.Error("Expected tool not to be found")
	}
}

func TestToolRegistryListToolNames(t *testing.T) {
	registry := NewToolRegistry()

	handler := func(ctx context.Context, req *ToolRequest) *ToolResponse { return nil }

	registry.Register(&ToolDefinition{Name: "tool_a"}, handler)
	registry.Register(&ToolDefinition{Name: "tool_b"}, handler)
	registry.Register(&ToolDefinition{Name: "tool_c"}, handler)

	names := registry.ListToolNames()

	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}

	nameMap := make(map[string]bool)
	for _, n := range names {
		nameMap[n] = true
	}

	for _, expected := range []string{"tool_a", "tool_b", "tool_c"} {
		if !nameMap[expected] {
			t.Errorf("Expected '%s' in tool names", expected)
		}
	}
}

func TestToolRegistryGenerateToolPrompt(t *testing.T) {
	registry := NewToolRegistry()

	def := &ToolDefinition{
		Name:        "file_read",
		Description: "Reads a file from disk",
		Type:        ToolTypeFile,
		Parameters: []Parameter{
			{Name: "path", Type: ParamTypeString, Required: true, Description: "File path to read"},
			{Name: "encoding", Type: ParamTypeString, Required: false, Description: "Text encoding"},
		},
	}

	registry.Register(def, func(ctx context.Context, req *ToolRequest) *ToolResponse { return nil })

	prompt := registry.GenerateToolPrompt()

	// Check prompt contains tool name
	if !strings.Contains(prompt, "file_read") {
		t.Error("Expected prompt to contain tool name")
	}

	// Check prompt contains description
	if !strings.Contains(prompt, "Reads a file from disk") {
		t.Error("Expected prompt to contain tool description")
	}

	// Check prompt contains parameters
	if !strings.Contains(prompt, "path") {
		t.Error("Expected prompt to contain parameter name")
	}

	// Check prompt contains required marker
	if !strings.Contains(prompt, "(required)") {
		t.Error("Expected prompt to contain required marker")
	}

	// Check prompt contains usage example
	if !strings.Contains(prompt, `"tool":`) {
		t.Error("Expected prompt to contain usage example")
	}
}

func TestToolRegistryGenerateOpenAIFunctions(t *testing.T) {
	registry := NewToolRegistry()

	def := &ToolDefinition{
		Name:        "search_code",
		Description: "Searches code in repository",
		Type:        ToolTypeSearch,
		Parameters: []Parameter{
			{Name: "query", Type: ParamTypeString, Required: true, Description: "Search query"},
			{Name: "language", Type: ParamTypeString, Required: false, Description: "Language filter", Enum: []string{"go", "python", "javascript"}},
		},
	}

	registry.Register(def, func(ctx context.Context, req *ToolRequest) *ToolResponse { return nil })

	functions := registry.GenerateOpenAIFunctions()

	if len(functions) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(functions))
	}

	fn := functions[0]

	if fn["type"] != "function" {
		t.Errorf("Expected type 'function', got '%v'", fn["type"])
	}

	fnDef := fn["function"].(map[string]interface{})
	if fnDef["name"] != "search_code" {
		t.Errorf("Expected name 'search_code', got '%v'", fnDef["name"])
	}

	params := fnDef["parameters"].(map[string]interface{})
	if params["type"] != "object" {
		t.Errorf("Expected parameters type 'object', got '%v'", params["type"])
	}

	required := params["required"].([]string)
	if len(required) != 1 || required[0] != "query" {
		t.Errorf("Expected required ['query'], got %v", required)
	}
}

func TestParseToolCall(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "simple json",
			input: `{"tool": "file_read", "params": {"path": "/test.txt"}}`,
			want:  "file_read",
		},
		{
			name:  "json with surrounding text",
			input: `I'll read the file for you: {"tool": "file_read", "params": {"path": "/test.txt"}} Done!`,
			want:  "file_read",
		},
		{
			name:  "nested json params",
			input: `{"tool": "api_call", "params": {"data": {"key": "value"}}}`,
			want:  "api_call",
		},
		{
			name:    "missing tool",
			input:   `{"params": {"path": "/test.txt"}}`,
			wantErr: true,
		},
		{
			name:    "invalid json",
			input:   `not json at all`,
			wantErr: true,
		},
		{
			name:    "empty tool name",
			input:   `{"tool": "", "params": {}}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseToolCall(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if req.Tool != tt.want {
				t.Errorf("Expected tool '%s', got '%s'", tt.want, req.Tool)
			}
		})
	}
}

func TestToolTypeConstants(t *testing.T) {
	tests := []struct {
		toolType ToolType
		expected string
	}{
		{ToolTypeFile, "file"},
		{ToolTypeCommand, "command"},
		{ToolTypeSearch, "search"},
		{ToolTypeGit, "git"},
		{ToolTypeAnalysis, "analysis"},
		{ToolTypeCustom, "custom"},
	}

	for _, tt := range tests {
		if string(tt.toolType) != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, tt.toolType)
		}
	}
}

func TestParameterTypeConstants(t *testing.T) {
	tests := []struct {
		paramType ParameterType
		expected  string
	}{
		{ParamTypeString, "string"},
		{ParamTypeNumber, "number"},
		{ParamTypeBoolean, "boolean"},
		{ParamTypeArray, "array"},
		{ParamTypeObject, "object"},
	}

	for _, tt := range tests {
		if string(tt.paramType) != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, tt.paramType)
		}
	}
}

func TestErrorCodeConstants(t *testing.T) {
	tests := []struct {
		code     int
		expected int
	}{
		{ErrCodeInvalidParams, -32602},
		{ErrCodeToolNotFound, -32601},
		{ErrCodeExecutionFailed, -32000},
		{ErrCodePermissionDenied, -32001},
		{ErrCodeTimeout, -32002},
	}

	for _, tt := range tests {
		if tt.code != tt.expected {
			t.Errorf("Expected %d, got %d", tt.expected, tt.code)
		}
	}
}

func TestToolRegistryConcurrentAccess(t *testing.T) {
	registry := NewToolRegistry()

	handler := func(ctx context.Context, req *ToolRequest) *ToolResponse {
		return &ToolResponse{ID: req.ID, Success: true}
	}

	// Register concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			def := &ToolDefinition{Name: string(rune('a'+n)) + "_tool"}
			registry.Register(def, handler)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Execute concurrently
	for i := 0; i < 10; i++ {
		go func(n int) {
			req := &ToolRequest{
				ID:     "req",
				Tool:   string(rune('a'+n)) + "_tool",
				Params: map[string]interface{}{},
			}
			registry.Execute(context.Background(), req)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if len(registry.ListTools()) != 10 {
		t.Errorf("Expected 10 tools after concurrent registration, got %d", len(registry.ListTools()))
	}
}

func TestExtractJSONFromText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "pure json",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "json with prefix",
			input:    `Here is the result: {"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "json with suffix",
			input:    `{"key": "value"} and that's it`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "nested json",
			input:    `{"outer": {"inner": "value"}}`,
			expected: `{"outer": {"inner": "value"}}`,
		},
		{
			name:     "no json",
			input:    `no json here`,
			expected: `no json here`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSONFromText(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
