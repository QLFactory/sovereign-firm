package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// SOCType defines the expected output type
type SOCType string

const (
	SOCTypeJSON SOCType = "json"
	SOCTypeCode SOCType = "code"
	SOCTypeText SOCType = "text"
)

// SOCContract defines a Structured Output Contract
type SOCContract struct {
	Type       SOCType                `json:"type"`
	Schema     map[string]interface{} `json:"schema,omitempty"`
	Validation []string               `json:"validation,omitempty"`
	Examples   []SOCExample           `json:"examples,omitempty"`
}

// SOCExample provides examples for the contract
type SOCExample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// SOCResult contains the validation result
type SOCResult struct {
	Valid    bool                   `json:"valid"`
	Output   interface{}            `json:"output,omitempty"`
	Errors   []string               `json:"errors,omitempty"`
	Parsed   map[string]interface{} `json:"parsed,omitempty"`
	RawInput string                 `json:"raw_input,omitempty"`
}

// SOCValidator validates LLM outputs against contracts
type SOCValidator struct {
	MaxRetries  int
	RetryPrompt string
	llmClient   llm.Client
}

// NewSOCValidator creates a new SOC validator
func NewSOCValidator(llmClient llm.Client, maxRetries int) *SOCValidator {
	return &SOCValidator{
		MaxRetries:  maxRetries,
		RetryPrompt: defaultRetryPrompt,
		llmClient:   llmClient,
	}
}

const defaultRetryPrompt = `Your previous response was invalid. Please fix the following issues and try again:

VALIDATION ERRORS:
%s

ORIGINAL REQUEST:
%s

Please provide a valid response following the exact format specified.`

// ValidateAndRetry validates output and retries if invalid
func (v *SOCValidator) ValidateAndRetry(
	ctx context.Context,
	contract *SOCContract,
	output string,
	originalPrompt string,
	systemPrompt string,
) (*SOCResult, error) {

	// First validation attempt
	result := v.Validate(contract, output)
	if result.Valid {
		return result, nil
	}

	// Retry loop
	for attempt := 1; attempt <= v.MaxRetries; attempt++ {
		// Build retry prompt with error feedback
		errorList := strings.Join(result.Errors, "\n- ")
		retryPrompt := fmt.Sprintf(v.RetryPrompt, errorList, originalPrompt)

		// Call LLM again
		resp, err := v.llmClient.Generate(ctx, llm.GenerateRequest{
			Prompt: retryPrompt,
			System: systemPrompt,
			Format: "json",
		})
		if err != nil {
			return nil, fmt.Errorf("retry attempt %d failed: %w", attempt, err)
		}

		// Validate new response
		result = v.Validate(contract, resp.Response)
		if result.Valid {
			return result, nil
		}
	}

	// Exhausted retries
	result.Errors = append(result.Errors, fmt.Sprintf("exhausted %d retry attempts", v.MaxRetries))
	return result, ErrMaxRetriesExceeded
}

// Validate checks if output matches the contract
func (v *SOCValidator) Validate(contract *SOCContract, output string) *SOCResult {
	result := &SOCResult{
		Valid:    true,
		RawInput: output,
		Errors:   make([]string, 0),
	}

	switch contract.Type {
	case SOCTypeJSON:
		result = v.validateJSON(contract, output)
	case SOCTypeCode:
		result = v.validateCode(contract, output)
	case SOCTypeText:
		result = v.validateText(contract, output)
	default:
		result.Valid = false
		result.Errors = append(result.Errors, "unknown contract type: "+string(contract.Type))
	}

	// Run custom validations
	if result.Valid && len(contract.Validation) > 0 {
		for _, validation := range contract.Validation {
			if err := v.runCustomValidation(validation, output, result.Parsed); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, err.Error())
			}
		}
	}

	return result
}

// validateJSON validates JSON output
func (v *SOCValidator) validateJSON(contract *SOCContract, output string) *SOCResult {
	result := &SOCResult{
		Valid:    true,
		RawInput: output,
		Errors:   make([]string, 0),
	}

	// Try to extract JSON from output (handles markdown code blocks)
	jsonStr := extractJSON(output)

	// Parse JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		// Try as array
		var parsedArray []interface{}
		if arrErr := json.Unmarshal([]byte(jsonStr), &parsedArray); arrErr != nil {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("invalid JSON: %s", err.Error()))
			return result
		}
		result.Output = parsedArray
	} else {
		result.Parsed = parsed
		result.Output = parsed
	}

	// Validate against schema if provided
	if contract.Schema != nil && result.Parsed != nil {
		schemaErrors := v.validateSchema(contract.Schema, result.Parsed)
		if len(schemaErrors) > 0 {
			result.Valid = false
			result.Errors = append(result.Errors, schemaErrors...)
		}
	}

	return result
}

// validateCode validates code output (extracts from markdown)
func (v *SOCValidator) validateCode(contract *SOCContract, output string) *SOCResult {
	result := &SOCResult{
		Valid:    true,
		RawInput: output,
		Errors:   make([]string, 0),
	}

	// Code output is expected to be a JSON map of file paths to content
	// or markdown code blocks

	// First try as JSON
	jsonResult := v.validateJSON(contract, output)
	if jsonResult.Valid {
		return jsonResult
	}

	// Otherwise, try to extract code blocks
	codeBlocks := extractCodeBlocks(output)
	if len(codeBlocks) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "no code blocks found in output")
		return result
	}

	result.Output = codeBlocks
	return result
}

// validateText validates text output
func (v *SOCValidator) validateText(contract *SOCContract, output string) *SOCResult {
	result := &SOCResult{
		Valid:    true,
		RawInput: output,
		Output:   output,
		Errors:   make([]string, 0),
	}

	if strings.TrimSpace(output) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "output is empty")
	}

	return result
}

// validateSchema performs basic schema validation
func (v *SOCValidator) validateSchema(schema map[string]interface{}, data map[string]interface{}) []string {
	errors := make([]string, 0)

	// Check for required properties
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		if required, ok := schema["required"].([]interface{}); ok {
			for _, req := range required {
				reqStr := fmt.Sprintf("%v", req)
				if _, exists := data[reqStr]; !exists {
					errors = append(errors, fmt.Sprintf("missing required property: %s", reqStr))
				}
			}
		}

		// Type validation for existing properties
		for propName, propSchema := range props {
			if value, exists := data[propName]; exists {
				if propMap, ok := propSchema.(map[string]interface{}); ok {
					if expectedType, ok := propMap["type"].(string); ok {
						if err := v.checkType(propName, value, expectedType); err != nil {
							errors = append(errors, err.Error())
						}
					}
				}
			}
		}
	}

	return errors
}

// checkType validates that a value matches the expected type
func (v *SOCValidator) checkType(name string, value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s: expected string, got %T", name, value)
		}
	case "number", "integer":
		switch value.(type) {
		case float64, int, int64:
			// OK
		default:
			return fmt.Errorf("%s: expected number, got %T", name, value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s: expected boolean, got %T", name, value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("%s: expected array, got %T", name, value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("%s: expected object, got %T", name, value)
		}
	}
	return nil
}

// runCustomValidation runs a custom validation rule
func (v *SOCValidator) runCustomValidation(rule string, output string, parsed map[string]interface{}) error {
	// Custom validation rules can be defined as patterns
	// e.g., "contains:import React" or "regex:^{.*}$"

	parts := strings.SplitN(rule, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid validation rule format: %s", rule)
	}

	ruleType := parts[0]
	ruleValue := parts[1]

	switch ruleType {
	case "contains":
		if !strings.Contains(output, ruleValue) {
			return fmt.Errorf("output must contain: %s", ruleValue)
		}
	case "not_contains":
		if strings.Contains(output, ruleValue) {
			return fmt.Errorf("output must not contain: %s", ruleValue)
		}
	case "regex":
		re, err := regexp.Compile(ruleValue)
		if err != nil {
			return fmt.Errorf("invalid regex: %s", err.Error())
		}
		if !re.MatchString(output) {
			return fmt.Errorf("output must match pattern: %s", ruleValue)
		}
	case "min_length":
		var minLen int
		fmt.Sscanf(ruleValue, "%d", &minLen)
		if len(output) < minLen {
			return fmt.Errorf("output too short (min %d chars)", minLen)
		}
	case "max_length":
		var maxLen int
		fmt.Sscanf(ruleValue, "%d", &maxLen)
		if len(output) > maxLen {
			return fmt.Errorf("output too long (max %d chars)", maxLen)
		}
	}

	return nil
}

// extractJSON tries to extract JSON from text (handles markdown code blocks)
func extractJSON(text string) string {
	text = strings.TrimSpace(text)

	// Check for markdown JSON block
	jsonBlockPattern := regexp.MustCompile("(?s)```json\\s*\\n(.+?)\\n```")
	if matches := jsonBlockPattern.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Check for generic code block
	codeBlockPattern := regexp.MustCompile("(?s)```\\s*\\n(.+?)\\n```")
	if matches := codeBlockPattern.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Assume raw JSON
	return text
}

// extractCodeBlocks extracts all code blocks from markdown
func extractCodeBlocks(text string) map[string]string {
	blocks := make(map[string]string)

	// Pattern: ```language:filename\ncode\n```
	pattern := regexp.MustCompile("(?s)```(\\w+)?(?::([^\\n]+))?\\s*\\n(.+?)\\n```")
	matches := pattern.FindAllStringSubmatch(text, -1)

	for i, match := range matches {
		lang := match[1]
		filename := match[2]
		code := match[3]

		if filename == "" {
			// Generate filename based on language
			ext := langToExtension(lang)
			filename = fmt.Sprintf("file%d%s", i+1, ext)
		}

		blocks[filename] = code
	}

	return blocks
}

// langToExtension maps language names to file extensions
func langToExtension(lang string) string {
	extensions := map[string]string{
		"javascript": ".js",
		"js":         ".js",
		"typescript": ".ts",
		"ts":         ".ts",
		"jsx":        ".jsx",
		"tsx":        ".tsx",
		"go":         ".go",
		"python":     ".py",
		"py":         ".py",
		"rust":       ".rs",
		"java":       ".java",
		"css":        ".css",
		"html":       ".html",
		"json":       ".json",
		"yaml":       ".yaml",
		"yml":        ".yml",
		"sql":        ".sql",
		"sh":         ".sh",
		"bash":       ".sh",
	}

	if ext, ok := extensions[strings.ToLower(lang)]; ok {
		return ext
	}
	return ".txt"
}
