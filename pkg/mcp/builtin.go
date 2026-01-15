package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RegisterBuiltinTools registers all built-in tools
func RegisterBuiltinTools(registry *ToolRegistry, workDir string) {
	// File operations
	registry.Register(fileReadDef(), fileReadHandler(workDir))
	registry.Register(fileWriteDef(), fileWriteHandler(workDir))
	registry.Register(fileListDef(), fileListHandler(workDir))
	registry.Register(fileDeleteDef(), fileDeleteHandler(workDir))

	// Command execution
	registry.Register(shellExecDef(), shellExecHandler(workDir))
	registry.Register(npmRunDef(), npmRunHandler(workDir))

	// Git operations
	registry.Register(gitStatusDef(), gitStatusHandler(workDir))
	registry.Register(gitDiffDef(), gitDiffHandler(workDir))

	// Search and analysis
	registry.Register(semanticSearchDef(), semanticSearchHandler())
	registry.Register(grepSearchDef(), grepSearchHandler(workDir))
}

// =====================
// FILE OPERATIONS
// =====================

func fileReadDef() *ToolDefinition {
	return &ToolDefinition{
		Name:        "file_read",
		Description: "Read the contents of a file",
		Type:        ToolTypeFile,
		Parameters: []Parameter{
			{Name: "path", Type: ParamTypeString, Description: "Path to the file", Required: true},
		},
		Returns: "File contents as a string",
	}
}

func fileReadHandler(workDir string) ToolHandler {
	return func(ctx context.Context, req *ToolRequest) *ToolResponse {
		path, _ := req.Params["path"].(string)
		fullPath := resolvePath(workDir, path)

		content, err := os.ReadFile(fullPath)
		if err != nil {
			return errorResponse(req.ID, ErrCodeExecutionFailed, "failed to read file", err.Error())
		}

		return &ToolResponse{
			ID:      req.ID,
			Success: true,
			Result:  string(content),
		}
	}
}

func fileWriteDef() *ToolDefinition {
	return &ToolDefinition{
		Name:        "file_write",
		Description: "Write content to a file (creates parent directories if needed)",
		Type:        ToolTypeFile,
		Parameters: []Parameter{
			{Name: "path", Type: ParamTypeString, Description: "Path to the file", Required: true},
			{Name: "content", Type: ParamTypeString, Description: "Content to write", Required: true},
			{Name: "append", Type: ParamTypeBoolean, Description: "Append instead of overwrite", Required: false, Default: false},
		},
		Returns: "Success message with file path",
	}
}

func fileWriteHandler(workDir string) ToolHandler {
	return func(ctx context.Context, req *ToolRequest) *ToolResponse {
		path, _ := req.Params["path"].(string)
		content, _ := req.Params["content"].(string)
		appendMode, _ := req.Params["append"].(bool)

		fullPath := resolvePath(workDir, path)

		// Create parent directories
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errorResponse(req.ID, ErrCodeExecutionFailed, "failed to create directory", err.Error())
		}

		var err error
		if appendMode {
			f, ferr := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if ferr != nil {
				err = ferr
			} else {
				_, err = f.WriteString(content)
				f.Close()
			}
		} else {
			err = os.WriteFile(fullPath, []byte(content), 0644)
		}

		if err != nil {
			return errorResponse(req.ID, ErrCodeExecutionFailed, "failed to write file", err.Error())
		}

		return &ToolResponse{
			ID:      req.ID,
			Success: true,
			Result:  fmt.Sprintf("Successfully wrote to %s", path),
		}
	}
}

func fileListDef() *ToolDefinition {
	return &ToolDefinition{
		Name:        "file_list",
		Description: "List files in a directory",
		Type:        ToolTypeFile,
		Parameters: []Parameter{
			{Name: "path", Type: ParamTypeString, Description: "Directory path", Required: true},
			{Name: "recursive", Type: ParamTypeBoolean, Description: "List recursively", Required: false, Default: false},
		},
		Returns: "Array of file paths",
	}
}

func fileListHandler(workDir string) ToolHandler {
	return func(ctx context.Context, req *ToolRequest) *ToolResponse {
		path, _ := req.Params["path"].(string)
		recursive, _ := req.Params["recursive"].(bool)

		fullPath := resolvePath(workDir, path)

		var files []string

		if recursive {
			filepath.Walk(fullPath, func(p string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					relPath, _ := filepath.Rel(fullPath, p)
					files = append(files, relPath)
				}
				return nil
			})
		} else {
			entries, err := os.ReadDir(fullPath)
			if err != nil {
				return errorResponse(req.ID, ErrCodeExecutionFailed, "failed to list directory", err.Error())
			}
			for _, e := range entries {
				files = append(files, e.Name())
			}
		}

		return &ToolResponse{
			ID:      req.ID,
			Success: true,
			Result:  files,
		}
	}
}

func fileDeleteDef() *ToolDefinition {
	return &ToolDefinition{
		Name:        "file_delete",
		Description: "Delete a file or empty directory",
		Type:        ToolTypeFile,
		Parameters: []Parameter{
			{Name: "path", Type: ParamTypeString, Description: "Path to delete", Required: true},
		},
		Returns: "Success message",
	}
}

func fileDeleteHandler(workDir string) ToolHandler {
	return func(ctx context.Context, req *ToolRequest) *ToolResponse {
		path, _ := req.Params["path"].(string)
		fullPath := resolvePath(workDir, path)

		// Security: don't allow deleting outside workDir
		if !strings.HasPrefix(fullPath, workDir) {
			return errorResponse(req.ID, ErrCodePermissionDenied, "cannot delete files outside work directory", "")
		}

		if err := os.Remove(fullPath); err != nil {
			return errorResponse(req.ID, ErrCodeExecutionFailed, "failed to delete", err.Error())
		}

		return &ToolResponse{
			ID:      req.ID,
			Success: true,
			Result:  fmt.Sprintf("Successfully deleted %s", path),
		}
	}
}

// =====================
// COMMAND EXECUTION
// =====================

func shellExecDef() *ToolDefinition {
	return &ToolDefinition{
		Name:        "shell_exec",
		Description: "Execute a shell command (use with caution)",
		Type:        ToolTypeCommand,
		Parameters: []Parameter{
			{Name: "command", Type: ParamTypeString, Description: "Command to execute", Required: true},
			{Name: "timeout", Type: ParamTypeNumber, Description: "Timeout in seconds", Required: false, Default: 30},
		},
		Returns: "Command output (stdout and stderr)",
	}
}

func shellExecHandler(workDir string) ToolHandler {
	return func(ctx context.Context, req *ToolRequest) *ToolResponse {
		command, _ := req.Params["command"].(string)
		timeout := 30.0
		if t, ok := req.Params["timeout"].(float64); ok {
			timeout = t
		}

		ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "sh", "-c", command)
		cmd.Dir = workDir

		output, err := cmd.CombinedOutput()
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return errorResponse(req.ID, ErrCodeTimeout, "command timed out", string(output))
			}
			return errorResponse(req.ID, ErrCodeExecutionFailed, "command failed", string(output))
		}

		return &ToolResponse{
			ID:      req.ID,
			Success: true,
			Result:  string(output),
		}
	}
}

func npmRunDef() *ToolDefinition {
	return &ToolDefinition{
		Name:        "npm_run",
		Description: "Run an npm script",
		Type:        ToolTypeCommand,
		Parameters: []Parameter{
			{Name: "script", Type: ParamTypeString, Description: "npm script to run (e.g., 'test', 'build')", Required: true},
			{Name: "args", Type: ParamTypeString, Description: "Additional arguments", Required: false},
		},
		Returns: "npm command output",
	}
}

func npmRunHandler(workDir string) ToolHandler {
	return func(ctx context.Context, req *ToolRequest) *ToolResponse {
		script, _ := req.Params["script"].(string)
		args, _ := req.Params["args"].(string)

		command := "npm run " + script
		if args != "" {
			command += " -- " + args
		}

		ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		cmd := exec.CommandContext(ctx, "sh", "-c", command)
		cmd.Dir = workDir

		output, err := cmd.CombinedOutput()
		if err != nil {
			return errorResponse(req.ID, ErrCodeExecutionFailed, "npm run failed", string(output))
		}

		return &ToolResponse{
			ID:      req.ID,
			Success: true,
			Result:  string(output),
		}
	}
}

// =====================
// GIT OPERATIONS
// =====================

func gitStatusDef() *ToolDefinition {
	return &ToolDefinition{
		Name:        "git_status",
		Description: "Get git repository status",
		Type:        ToolTypeGit,
		Parameters:  []Parameter{},
		Returns:     "Git status output",
	}
}

func gitStatusHandler(workDir string) ToolHandler {
	return func(ctx context.Context, req *ToolRequest) *ToolResponse {
		cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
		cmd.Dir = workDir

		output, err := cmd.CombinedOutput()
		if err != nil {
			return errorResponse(req.ID, ErrCodeExecutionFailed, "git status failed", string(output))
		}

		return &ToolResponse{
			ID:      req.ID,
			Success: true,
			Result:  string(output),
		}
	}
}

func gitDiffDef() *ToolDefinition {
	return &ToolDefinition{
		Name:        "git_diff",
		Description: "Show git diff for a file or all changes",
		Type:        ToolTypeGit,
		Parameters: []Parameter{
			{Name: "path", Type: ParamTypeString, Description: "File path (optional, shows all if empty)", Required: false},
			{Name: "staged", Type: ParamTypeBoolean, Description: "Show staged changes only", Required: false, Default: false},
		},
		Returns: "Git diff output",
	}
}

func gitDiffHandler(workDir string) ToolHandler {
	return func(ctx context.Context, req *ToolRequest) *ToolResponse {
		path, _ := req.Params["path"].(string)
		staged, _ := req.Params["staged"].(bool)

		args := []string{"diff"}
		if staged {
			args = append(args, "--staged")
		}
		if path != "" {
			args = append(args, "--", path)
		}

		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = workDir

		output, err := cmd.CombinedOutput()
		if err != nil {
			return errorResponse(req.ID, ErrCodeExecutionFailed, "git diff failed", string(output))
		}

		return &ToolResponse{
			ID:      req.ID,
			Success: true,
			Result:  string(output),
		}
	}
}

// =====================
// SEARCH AND ANALYSIS
// =====================

func semanticSearchDef() *ToolDefinition {
	return &ToolDefinition{
		Name:        "semantic_search",
		Description: "Search for relevant code or documentation using semantic similarity",
		Type:        ToolTypeSearch,
		Parameters: []Parameter{
			{Name: "query", Type: ParamTypeString, Description: "Search query", Required: true},
			{Name: "top_k", Type: ParamTypeNumber, Description: "Number of results", Required: false, Default: 5},
		},
		Returns: "Array of relevant documents/code snippets",
	}
}

func semanticSearchHandler() ToolHandler {
	return func(ctx context.Context, req *ToolRequest) *ToolResponse {
		query, _ := req.Params["query"].(string)
		_ = query

		// TODO: Integrate with memory store (ChromaDB)
		// For now, return placeholder
		return &ToolResponse{
			ID:      req.ID,
			Success: true,
			Result: []map[string]interface{}{
				{
					"content":  "// Semantic search not yet connected to memory store",
					"score":    0.0,
					"metadata": map[string]string{"note": "implement ChromaDB integration"},
				},
			},
		}
	}
}

func grepSearchDef() *ToolDefinition {
	return &ToolDefinition{
		Name:        "grep_search",
		Description: "Search for text patterns in files using grep",
		Type:        ToolTypeSearch,
		Parameters: []Parameter{
			{Name: "pattern", Type: ParamTypeString, Description: "Pattern to search for", Required: true},
			{Name: "path", Type: ParamTypeString, Description: "Path to search in", Required: false, Default: "."},
			{Name: "regex", Type: ParamTypeBoolean, Description: "Use regex pattern", Required: false, Default: false},
		},
		Returns: "Matching lines with file paths",
	}
}

func grepSearchHandler(workDir string) ToolHandler {
	return func(ctx context.Context, req *ToolRequest) *ToolResponse {
		pattern, _ := req.Params["pattern"].(string)
		path := "."
		if p, ok := req.Params["path"].(string); ok && p != "" {
			path = p
		}
		useRegex, _ := req.Params["regex"].(bool)

		fullPath := resolvePath(workDir, path)

		args := []string{"-r", "-n"}
		if !useRegex {
			args = append(args, "-F") // Fixed string (not regex)
		}
		args = append(args, pattern, fullPath)

		cmd := exec.CommandContext(ctx, "grep", args...)

		output, _ := cmd.CombinedOutput() // grep returns error on no match

		// Parse grep output
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		results := make([]map[string]interface{}, 0)

		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, ":", 3)
			if len(parts) >= 3 {
				results = append(results, map[string]interface{}{
					"file":    parts[0],
					"line":    parts[1],
					"content": parts[2],
				})
			}
		}

		return &ToolResponse{
			ID:      req.ID,
			Success: true,
			Result:  results,
		}
	}
}

// =====================
// HELPER FUNCTIONS
// =====================

func resolvePath(workDir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(workDir, path)
}

func errorResponse(id string, code int, message, details string) *ToolResponse {
	return &ToolResponse{
		ID:      id,
		Success: false,
		Error: &ToolError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// FormatToolResult formats a tool result for LLM context
func FormatToolResult(resp *ToolResponse) string {
	if !resp.Success {
		return fmt.Sprintf("Tool Error: %s\n%s", resp.Error.Message, resp.Error.Details)
	}

	switch result := resp.Result.(type) {
	case string:
		return result
	case []interface{}, []string, []map[string]interface{}:
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return string(jsonBytes)
	default:
		jsonBytes, _ := json.Marshal(result)
		return string(jsonBytes)
	}
}
