package sandbox

import (
	"context"
	"testing"
	"time"
)

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelBrowser, "browser"},
		{LevelContainer, "container"},
		{LevelMicroVM, "microvm"},
		{Level(99), "unknown"},
	}

	for _, tt := range tests {
		result := tt.level.String()
		if result != tt.expected {
			t.Errorf("Expected '%s' for level %d, got '%s'", tt.expected, tt.level, result)
		}
	}
}

func TestLevelConstants(t *testing.T) {
	if LevelBrowser != 1 {
		t.Errorf("Expected LevelBrowser = 1, got %d", LevelBrowser)
	}
	if LevelContainer != 2 {
		t.Errorf("Expected LevelContainer = 2, got %d", LevelContainer)
	}
	if LevelMicroVM != 3 {
		t.Errorf("Expected LevelMicroVM = 3, got %d", LevelMicroVM)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig(LevelContainer)

	if config == nil {
		t.Fatal("Expected config to be created, got nil")
	}

	if config.Level != LevelContainer {
		t.Errorf("Expected level LevelContainer, got %d", config.Level)
	}

	if config.MaxCPU != 50 {
		t.Errorf("Expected MaxCPU 50, got %d", config.MaxCPU)
	}

	if config.MaxMemory != 512 {
		t.Errorf("Expected MaxMemory 512, got %d", config.MaxMemory)
	}

	if config.MaxDisk != 1024 {
		t.Errorf("Expected MaxDisk 1024, got %d", config.MaxDisk)
	}

	if config.MaxTimeout != 5*time.Minute {
		t.Errorf("Expected MaxTimeout 5m, got %v", config.MaxTimeout)
	}

	if config.NetworkEnabled {
		t.Error("Expected NetworkEnabled false")
	}

	if len(config.AllowedHosts) != 2 {
		t.Errorf("Expected 2 allowed hosts, got %d", len(config.AllowedHosts))
	}
}

func TestConfigStructure(t *testing.T) {
	config := &Config{
		ID:             "sandbox-123",
		ProjectID:      "project-456",
		Level:          LevelContainer,
		MaxCPU:         80,
		MaxMemory:      1024,
		MaxDisk:        2048,
		MaxTimeout:     10 * time.Minute,
		NetworkEnabled: true,
		AllowedHosts:   []string{"api.github.com"},
		Language:       "go",
		Runtime:        "go:1.21",
		WorkDir:        "/app",
	}

	if config.ID != "sandbox-123" {
		t.Errorf("Expected ID 'sandbox-123', got '%s'", config.ID)
	}

	if config.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", config.Language)
	}

	if config.Runtime != "go:1.21" {
		t.Errorf("Expected runtime 'go:1.21', got '%s'", config.Runtime)
	}
}

func TestExecuteRequestStructure(t *testing.T) {
	req := &ExecuteRequest{
		Command: "go",
		Args:    []string{"run", "main.go"},
		Stdin:   "input data",
		Files: map[string]string{
			"main.go": "package main\n\nfunc main() {}",
		},
		Env: map[string]string{
			"GOPATH": "/go",
		},
		Timeout: 30 * time.Second,
	}

	if req.Command != "go" {
		t.Errorf("Expected command 'go', got '%s'", req.Command)
	}

	if len(req.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(req.Args))
	}

	if req.Files["main.go"] == "" {
		t.Error("Expected main.go file to be present")
	}
}

func TestExecuteResultStructure(t *testing.T) {
	result := &ExecuteResult{
		Success:    true,
		ExitCode:   0,
		Stdout:     "Hello, World!",
		Stderr:     "",
		Duration:   150 * time.Millisecond,
		CPUTime:    100 * time.Millisecond,
		MemoryUsed: 64,
		OutputFiles: map[string]string{
			"output.txt": "result",
		},
		Error: "",
	}

	if !result.Success {
		t.Error("Expected Success true")
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected ExitCode 0, got %d", result.ExitCode)
	}

	if result.Stdout != "Hello, World!" {
		t.Errorf("Expected stdout 'Hello, World!', got '%s'", result.Stdout)
	}
}

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("Expected manager to be created, got nil")
	}

	if manager.sandboxes == nil {
		t.Error("Expected sandboxes map to be initialized")
	}
}

func TestManagerGet(t *testing.T) {
	manager := NewManager()

	// Get non-existent sandbox
	_, ok := manager.Get("nonexistent")
	if ok {
		t.Error("Expected Get to return false for non-existent sandbox")
	}
}

func TestManagerCreateBrowserLevel(t *testing.T) {
	manager := NewManager()

	config := &Config{
		ID:    "browser-sandbox",
		Level: LevelBrowser,
	}

	_, err := manager.Create(context.Background(), config)
	if err == nil {
		t.Error("Expected error for browser level sandbox (managed client-side)")
	}
}

func TestManagerCreateUnknownLevel(t *testing.T) {
	manager := NewManager()

	config := &Config{
		ID:    "unknown-sandbox",
		Level: Level(99),
	}

	_, err := manager.Create(context.Background(), config)
	if err == nil {
		t.Error("Expected error for unknown sandbox level")
	}
}

func TestManagerDestroy(t *testing.T) {
	manager := NewManager()

	// Try to destroy non-existent sandbox
	err := manager.Destroy(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error when destroying non-existent sandbox")
	}
}

func TestSelectLevelBrowser(t *testing.T) {
	// JavaScript with high trust and no network should be browser
	level := SelectLevel("javascript", false, "high")
	if level != LevelBrowser {
		t.Errorf("Expected LevelBrowser for JS high trust, got %s", level.String())
	}

	// TypeScript with high trust and no network should be browser
	level = SelectLevel("typescript", false, "high")
	if level != LevelBrowser {
		t.Errorf("Expected LevelBrowser for TS high trust, got %s", level.String())
	}

	// JSX with high trust and no network should be browser
	level = SelectLevel("jsx", false, "high")
	if level != LevelBrowser {
		t.Errorf("Expected LevelBrowser for JSX high trust, got %s", level.String())
	}
}

func TestSelectLevelContainer(t *testing.T) {
	// Go should default to container
	level := SelectLevel("go", false, "medium")
	if level != LevelContainer {
		t.Errorf("Expected LevelContainer for Go, got %s", level.String())
	}

	// Python should default to container
	level = SelectLevel("python", false, "medium")
	if level != LevelContainer {
		t.Errorf("Expected LevelContainer for Python, got %s", level.String())
	}

	// JavaScript needing network should be container
	level = SelectLevel("javascript", true, "high")
	if level != LevelContainer {
		t.Errorf("Expected LevelContainer for JS with network, got %s", level.String())
	}
}

func TestSelectLevelMicroVM(t *testing.T) {
	// Untrusted code should go to microVM
	level := SelectLevel("python", false, "untrusted")
	if level != LevelMicroVM {
		t.Errorf("Expected LevelMicroVM for untrusted code, got %s", level.String())
	}

	// Low trust should go to microVM
	level = SelectLevel("go", true, "low")
	if level != LevelMicroVM {
		t.Errorf("Expected LevelMicroVM for low trust, got %s", level.String())
	}
}

func TestSelectLevelJavaScriptLowTrust(t *testing.T) {
	// Even JavaScript should go to microVM if low trust
	level := SelectLevel("javascript", false, "low")
	if level != LevelMicroVM {
		t.Errorf("Expected LevelMicroVM for JS low trust, got %s", level.String())
	}
}

func TestDefaultConfigDifferentLevels(t *testing.T) {
	// Test that DefaultConfig works for all levels
	levels := []Level{LevelBrowser, LevelContainer, LevelMicroVM}

	for _, level := range levels {
		config := DefaultConfig(level)
		if config.Level != level {
			t.Errorf("Expected level %d, got %d", level, config.Level)
		}
	}
}

func TestExecuteRequestWithFiles(t *testing.T) {
	req := &ExecuteRequest{
		Command: "python",
		Args:    []string{"main.py"},
		Files: map[string]string{
			"main.py":   "print('hello')",
			"helper.py": "def helper(): pass",
			"data.json": `{"key": "value"}`,
		},
	}

	if len(req.Files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(req.Files))
	}
}

func TestExecuteResultWithError(t *testing.T) {
	result := &ExecuteResult{
		Success:  false,
		ExitCode: 1,
		Stdout:   "",
		Stderr:   "Error: file not found",
		Duration: 50 * time.Millisecond,
		Error:    "execution failed",
	}

	if result.Success {
		t.Error("Expected Success false for error result")
	}

	if result.ExitCode != 1 {
		t.Errorf("Expected ExitCode 1, got %d", result.ExitCode)
	}

	if result.Error != "execution failed" {
		t.Errorf("Expected error 'execution failed', got '%s'", result.Error)
	}
}
