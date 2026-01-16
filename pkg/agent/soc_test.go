package agent

import (
	"context"
	"testing"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// mockLLMClientWithResponse is a mock that returns configurable responses
type mockLLMClientWithResponse struct {
	responses []string
	callCount int
}

func (m *mockLLMClientWithResponse) Generate(ctx context.Context, req llm.GenerateRequest) (*llm.GenerateResponse, error) {
	resp := "mock"
	if m.callCount < len(m.responses) {
		resp = m.responses[m.callCount]
	}
	m.callCount++
	return &llm.GenerateResponse{Response: resp, Done: true}, nil
}

func (m *mockLLMClientWithResponse) Embed(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2}, nil
}

func (m *mockLLMClientWithResponse) Ping(ctx context.Context) error {
	return nil
}

func TestNewSOCValidator(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	if validator == nil {
		t.Fatal("Expected validator to be created")
	}

	if validator.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", validator.MaxRetries)
	}

	if validator.RetryPrompt == "" {
		t.Error("Expected RetryPrompt to be set")
	}
}

func TestSOCTypeConstants(t *testing.T) {
	tests := []struct {
		socType  SOCType
		expected string
	}{
		{SOCTypeJSON, "json"},
		{SOCTypeCode, "code"},
		{SOCTypeText, "text"},
	}

	for _, tt := range tests {
		if string(tt.socType) != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, tt.socType)
		}
	}
}

func TestValidateJSONValid(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeJSON,
	}

	result := validator.Validate(contract, `{"name": "test", "value": 42}`)

	if !result.Valid {
		t.Errorf("Expected valid JSON, got errors: %v", result.Errors)
	}

	if result.Parsed == nil {
		t.Error("Expected parsed map to be set")
	}

	if result.Parsed["name"] != "test" {
		t.Error("Expected parsed 'name' to be 'test'")
	}
}

func TestValidateJSONInvalid(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeJSON,
	}

	result := validator.Validate(contract, `{invalid json}`)

	if result.Valid {
		t.Error("Expected invalid result for malformed JSON")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for invalid JSON")
	}
}

func TestValidateJSONInMarkdownBlock(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeJSON,
	}

	// JSON wrapped in markdown code block
	output := "```json\n{\"status\": \"success\"}\n```"

	result := validator.Validate(contract, output)

	if !result.Valid {
		t.Errorf("Expected valid JSON from markdown block, got errors: %v", result.Errors)
	}
}

func TestValidateJSONWithSchema(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeJSON,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{"type": "string"},
				"age":  map[string]interface{}{"type": "number"},
			},
			"required": []interface{}{"name", "age"},
		},
	}

	// Valid JSON matching schema
	result := validator.Validate(contract, `{"name": "Alice", "age": 30}`)
	if !result.Valid {
		t.Errorf("Expected valid, got errors: %v", result.Errors)
	}

	// Missing required field
	result = validator.Validate(contract, `{"name": "Bob"}`)
	if result.Valid {
		t.Error("Expected invalid for missing required field")
	}
}

func TestValidateJSONWrongType(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeJSON,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"count": map[string]interface{}{"type": "number"},
			},
		},
	}

	// Wrong type (string instead of number)
	result := validator.Validate(contract, `{"count": "not a number"}`)
	if result.Valid {
		t.Error("Expected invalid for wrong type")
	}
}

func TestValidateJSONArray(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeJSON,
	}

	result := validator.Validate(contract, `[1, 2, 3, "test"]`)

	if !result.Valid {
		t.Errorf("Expected valid JSON array, got errors: %v", result.Errors)
	}

	if result.Output == nil {
		t.Error("Expected output to be set for array")
	}
}

func TestValidateTextValid(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeText,
	}

	result := validator.Validate(contract, "This is some text content.")

	if !result.Valid {
		t.Errorf("Expected valid text, got errors: %v", result.Errors)
	}
}

func TestValidateTextEmpty(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeText,
	}

	result := validator.Validate(contract, "   ")

	if result.Valid {
		t.Error("Expected invalid for empty/whitespace text")
	}
}

func TestValidateCodeFromJSON(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeCode,
	}

	// Code output as JSON map of files
	output := `{"main.go": "package main\n\nfunc main() {}", "util.go": "package main\n\nfunc util() {}"}`

	result := validator.Validate(contract, output)

	if !result.Valid {
		t.Errorf("Expected valid code output, got errors: %v", result.Errors)
	}
}

func TestValidateCodeFromMarkdown(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeCode,
	}

	output := "```go\npackage main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n```"

	result := validator.Validate(contract, output)

	if !result.Valid {
		t.Errorf("Expected valid code blocks, got errors: %v", result.Errors)
	}
}

func TestValidateCodeNoBlocks(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeCode,
	}

	// No code blocks and not valid JSON
	result := validator.Validate(contract, "This is just plain text without code.")

	if result.Valid {
		t.Error("Expected invalid for no code blocks")
	}
}

func TestValidateUnknownType(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCType("unknown"),
	}

	result := validator.Validate(contract, "test")

	if result.Valid {
		t.Error("Expected invalid for unknown contract type")
	}
}

func TestCustomValidationContains(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type:       SOCTypeText,
		Validation: []string{"contains:import React"},
	}

	// Has the required text
	result := validator.Validate(contract, "import React from 'react';")
	if !result.Valid {
		t.Errorf("Expected valid, got errors: %v", result.Errors)
	}

	// Missing required text
	result = validator.Validate(contract, "import Vue from 'vue';")
	if result.Valid {
		t.Error("Expected invalid when missing required text")
	}
}

func TestCustomValidationNotContains(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type:       SOCTypeText,
		Validation: []string{"not_contains:console.log"},
	}

	// Does not have forbidden text
	result := validator.Validate(contract, "function test() { return 42; }")
	if !result.Valid {
		t.Errorf("Expected valid, got errors: %v", result.Errors)
	}

	// Has forbidden text
	result = validator.Validate(contract, "function test() { console.log('debug'); }")
	if result.Valid {
		t.Error("Expected invalid when containing forbidden text")
	}
}

func TestCustomValidationRegex(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type:       SOCTypeText,
		Validation: []string{"regex:^\\{.*\\}$"},
	}

	// Matches regex (starts and ends with braces)
	result := validator.Validate(contract, "{content}")
	if !result.Valid {
		t.Errorf("Expected valid for regex match, got errors: %v", result.Errors)
	}

	// Doesn't match
	result = validator.Validate(contract, "no braces")
	if result.Valid {
		t.Error("Expected invalid for regex mismatch")
	}
}

func TestCustomValidationMinLength(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type:       SOCTypeText,
		Validation: []string{"min_length:10"},
	}

	// Long enough
	result := validator.Validate(contract, "This is long enough text.")
	if !result.Valid {
		t.Errorf("Expected valid, got errors: %v", result.Errors)
	}

	// Too short
	result = validator.Validate(contract, "Short")
	if result.Valid {
		t.Error("Expected invalid for text below min length")
	}
}

func TestCustomValidationMaxLength(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type:       SOCTypeText,
		Validation: []string{"max_length:20"},
	}

	// Short enough
	result := validator.Validate(contract, "Short text")
	if !result.Valid {
		t.Errorf("Expected valid, got errors: %v", result.Errors)
	}

	// Too long
	result = validator.Validate(contract, "This text is way too long to pass validation")
	if result.Valid {
		t.Error("Expected invalid for text above max length")
	}
}

func TestCustomValidationInvalidFormat(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type:       SOCTypeText,
		Validation: []string{"invalid_format_no_colon"},
	}

	result := validator.Validate(contract, "test")
	if result.Valid {
		t.Error("Expected invalid for malformed validation rule")
	}
}

func TestValidateAndRetrySuccess(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	contract := &SOCContract{
		Type: SOCTypeJSON,
	}

	// Valid on first try
	result, err := validator.ValidateAndRetry(
		context.Background(),
		contract,
		`{"status": "ok"}`,
		"original prompt",
		"system prompt",
	)

	if err != nil {
		t.Fatalf("ValidateAndRetry failed: %v", err)
	}

	if !result.Valid {
		t.Error("Expected valid result")
	}
}

func TestValidateAndRetryWithRetries(t *testing.T) {
	// Mock that returns valid JSON on second call
	mockClient := &mockLLMClientWithResponse{
		responses: []string{`{"valid": true}`},
	}

	validator := NewSOCValidator(mockClient, 3)

	contract := &SOCContract{
		Type: SOCTypeJSON,
	}

	// Invalid on first try, retry should succeed
	result, err := validator.ValidateAndRetry(
		context.Background(),
		contract,
		`{invalid}`,
		"original prompt",
		"system prompt",
	)

	if err != nil {
		t.Fatalf("ValidateAndRetry failed: %v", err)
	}

	if !result.Valid {
		t.Error("Expected valid result after retry")
	}

	if mockClient.callCount != 1 {
		t.Errorf("Expected 1 retry call, got %d", mockClient.callCount)
	}
}

func TestValidateAndRetryExhausted(t *testing.T) {
	// Mock that always returns invalid JSON
	mockClient := &mockLLMClientWithResponse{
		responses: []string{"{bad}", "{bad}", "{bad}"},
	}

	validator := NewSOCValidator(mockClient, 2)

	contract := &SOCContract{
		Type: SOCTypeJSON,
	}

	result, err := validator.ValidateAndRetry(
		context.Background(),
		contract,
		`{invalid}`,
		"original prompt",
		"system prompt",
	)

	if err != ErrMaxRetriesExceeded {
		t.Errorf("Expected ErrMaxRetriesExceeded, got %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result after exhausting retries")
	}

	// Should have error about exhausted retries
	foundExhaustedError := false
	for _, e := range result.Errors {
		if contains(e, "exhausted") {
			foundExhaustedError = true
			break
		}
	}
	if !foundExhaustedError {
		t.Error("Expected error message about exhausted retries")
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "pure JSON",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON in markdown block",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON in generic code block",
			input:    "```\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "with whitespace",
			input:    "  {\"key\": \"value\"}  ",
			expected: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestExtractCodeBlocks(t *testing.T) {
	input := "```go:main.go\npackage main\n```\n\n```typescript:app.ts\nexport const x = 1;\n```"

	blocks := extractCodeBlocks(input)

	if len(blocks) != 2 {
		t.Errorf("Expected 2 code blocks, got %d", len(blocks))
	}

	if blocks["main.go"] != "package main" {
		t.Errorf("Unexpected content for main.go: %s", blocks["main.go"])
	}

	if blocks["app.ts"] != "export const x = 1;" {
		t.Errorf("Unexpected content for app.ts: %s", blocks["app.ts"])
	}
}

func TestExtractCodeBlocksWithoutFilenames(t *testing.T) {
	input := "```go\nfunc test() {}\n```\n\n```python\ndef test(): pass\n```"

	blocks := extractCodeBlocks(input)

	if len(blocks) != 2 {
		t.Errorf("Expected 2 code blocks, got %d", len(blocks))
	}

	// Should generate filenames based on language
	hasGoFile := false
	hasPyFile := false
	for filename := range blocks {
		if filename == "file1.go" {
			hasGoFile = true
		}
		if filename == "file2.py" {
			hasPyFile = true
		}
	}

	if !hasGoFile {
		t.Error("Expected .go file to be generated")
	}
	if !hasPyFile {
		t.Error("Expected .py file to be generated")
	}
}

func TestLangToExtension(t *testing.T) {
	tests := []struct {
		lang     string
		expected string
	}{
		{"javascript", ".js"},
		{"js", ".js"},
		{"typescript", ".ts"},
		{"ts", ".ts"},
		{"jsx", ".jsx"},
		{"tsx", ".tsx"},
		{"go", ".go"},
		{"python", ".py"},
		{"py", ".py"},
		{"rust", ".rs"},
		{"java", ".java"},
		{"css", ".css"},
		{"html", ".html"},
		{"json", ".json"},
		{"yaml", ".yaml"},
		{"yml", ".yml"},
		{"sql", ".sql"},
		{"sh", ".sh"},
		{"bash", ".sh"},
		{"unknown", ".txt"},
		{"", ".txt"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := langToExtension(tt.lang)
			if result != tt.expected {
				t.Errorf("langToExtension(%q) = %q, want %q", tt.lang, result, tt.expected)
			}
		})
	}
}

func TestSOCContractStructure(t *testing.T) {
	contract := SOCContract{
		Type: SOCTypeJSON,
		Schema: map[string]interface{}{
			"type": "object",
		},
		Validation: []string{"min_length:10"},
		Examples: []SOCExample{
			{Input: "test input", Output: "test output"},
		},
	}

	if contract.Type != SOCTypeJSON {
		t.Error("Type mismatch")
	}

	if len(contract.Validation) != 1 {
		t.Error("Validation count mismatch")
	}

	if len(contract.Examples) != 1 {
		t.Error("Examples count mismatch")
	}
}

func TestSOCResultStructure(t *testing.T) {
	result := SOCResult{
		Valid:    true,
		Output:   map[string]interface{}{"key": "value"},
		Errors:   []string{},
		Parsed:   map[string]interface{}{"key": "value"},
		RawInput: `{"key": "value"}`,
	}

	if !result.Valid {
		t.Error("Valid mismatch")
	}

	if result.RawInput != `{"key": "value"}` {
		t.Error("RawInput mismatch")
	}
}

func TestCheckTypeString(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	err := validator.checkType("field", "hello", "string")
	if err != nil {
		t.Errorf("Expected no error for string type, got %v", err)
	}

	err = validator.checkType("field", 123, "string")
	if err == nil {
		t.Error("Expected error for number when expecting string")
	}
}

func TestCheckTypeNumber(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	err := validator.checkType("field", 42.5, "number")
	if err != nil {
		t.Errorf("Expected no error for float number, got %v", err)
	}

	err = validator.checkType("field", 42, "integer")
	if err != nil {
		t.Errorf("Expected no error for int, got %v", err)
	}

	err = validator.checkType("field", "42", "number")
	if err == nil {
		t.Error("Expected error for string when expecting number")
	}
}

func TestCheckTypeBoolean(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	err := validator.checkType("field", true, "boolean")
	if err != nil {
		t.Errorf("Expected no error for boolean, got %v", err)
	}

	err = validator.checkType("field", "true", "boolean")
	if err == nil {
		t.Error("Expected error for string when expecting boolean")
	}
}

func TestCheckTypeArray(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	err := validator.checkType("field", []interface{}{1, 2, 3}, "array")
	if err != nil {
		t.Errorf("Expected no error for array, got %v", err)
	}

	err = validator.checkType("field", "not an array", "array")
	if err == nil {
		t.Error("Expected error for string when expecting array")
	}
}

func TestCheckTypeObject(t *testing.T) {
	validator := NewSOCValidator(&mockLLMClient{}, 3)

	err := validator.checkType("field", map[string]interface{}{"key": "value"}, "object")
	if err != nil {
		t.Errorf("Expected no error for object, got %v", err)
	}

	err = validator.checkType("field", "not an object", "object")
	if err == nil {
		t.Error("Expected error for string when expecting object")
	}
}
