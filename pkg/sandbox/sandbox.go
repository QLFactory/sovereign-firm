// Package sandbox provides multi-layer code execution environments.
// It implements three isolation levels: WebContainers (browser), Docker+gVisor (container), Firecracker (microVM).
package sandbox

import (
	"context"
	"fmt"
	"time"
)

// Level represents the isolation level of a sandbox
type Level int

const (
	// LevelBrowser uses WebContainers for in-browser execution (JS/TS only)
	LevelBrowser Level = 1

	// LevelContainer uses Docker + gVisor for server-side execution
	LevelContainer Level = 2

	// LevelMicroVM uses Firecracker for maximum isolation
	LevelMicroVM Level = 3
)

func (l Level) String() string {
	switch l {
	case LevelBrowser:
		return "browser"
	case LevelContainer:
		return "container"
	case LevelMicroVM:
		return "microvm"
	default:
		return "unknown"
	}
}

// Config holds sandbox configuration
type Config struct {
	// Identification
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`

	// Isolation level
	Level Level `json:"level"`

	// Resource limits
	MaxCPU     int           `json:"max_cpu_percent"` // 0-100
	MaxMemory  int64         `json:"max_memory_mb"`
	MaxDisk    int64         `json:"max_disk_mb"`
	MaxTimeout time.Duration `json:"max_timeout"`

	// Network restrictions
	NetworkEnabled bool     `json:"network_enabled"`
	AllowedHosts   []string `json:"allowed_hosts"`

	// Language/runtime
	Language string `json:"language"`
	Runtime  string `json:"runtime"` // e.g., "node:20", "go:1.21", "python:3.11"

	// Working directory
	WorkDir string `json:"work_dir"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig(level Level) *Config {
	return &Config{
		Level:          level,
		MaxCPU:         50,
		MaxMemory:      512,
		MaxDisk:        1024,
		MaxTimeout:     5 * time.Minute,
		NetworkEnabled: false,
		AllowedHosts:   []string{"registry.npmjs.org", "pypi.org"},
	}
}

// ExecuteRequest represents a code execution request
type ExecuteRequest struct {
	// Code or command to execute
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Stdin   string   `json:"stdin,omitempty"`

	// Files to write before execution
	Files map[string]string `json:"files,omitempty"`

	// Environment variables
	Env map[string]string `json:"env,omitempty"`

	// Timeout for this specific execution
	Timeout time.Duration `json:"timeout,omitempty"`
}

// ExecuteResult holds the result of code execution
type ExecuteResult struct {
	// Exit status
	Success  bool `json:"success"`
	ExitCode int  `json:"exit_code"`

	// Output
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`

	// Timing
	Duration time.Duration `json:"duration"`

	// Resource usage
	CPUTime    time.Duration `json:"cpu_time,omitempty"`
	MemoryUsed int64         `json:"memory_used_mb,omitempty"`

	// Files produced
	OutputFiles map[string]string `json:"output_files,omitempty"`

	// Error if any
	Error string `json:"error,omitempty"`
}

// Sandbox is the interface for code execution environments
type Sandbox interface {
	// Level returns the isolation level
	Level() Level

	// Start initializes the sandbox
	Start(ctx context.Context) error

	// Stop terminates the sandbox
	Stop(ctx context.Context) error

	// IsRunning checks if the sandbox is active
	IsRunning() bool

	// Execute runs code in the sandbox
	Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResult, error)

	// WriteFile writes a file to the sandbox filesystem
	WriteFile(ctx context.Context, path string, content []byte) error

	// ReadFile reads a file from the sandbox filesystem
	ReadFile(ctx context.Context, path string) ([]byte, error)

	// ListFiles lists files in a directory
	ListFiles(ctx context.Context, path string) ([]string, error)

	// GetConfig returns the sandbox configuration
	GetConfig() *Config
}

// Manager manages sandbox lifecycle
type Manager struct {
	sandboxes map[string]Sandbox
}

// NewManager creates a new sandbox manager
func NewManager() *Manager {
	return &Manager{
		sandboxes: make(map[string]Sandbox),
	}
}

// Create creates a new sandbox with the given config
func (m *Manager) Create(ctx context.Context, config *Config) (Sandbox, error) {
	var sandbox Sandbox
	var err error

	switch config.Level {
	case LevelBrowser:
		// WebContainers run in browser, not managed here
		return nil, fmt.Errorf("browser sandboxes are managed client-side")
	case LevelContainer:
		sandbox, err = NewContainerSandbox(config)
	case LevelMicroVM:
		sandbox, err = NewFirecrackerSandbox(config)
	default:
		return nil, fmt.Errorf("unknown sandbox level: %d", config.Level)
	}

	if err != nil {
		return nil, err
	}

	if err := sandbox.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start sandbox: %w", err)
	}

	m.sandboxes[config.ID] = sandbox
	return sandbox, nil
}

// Get retrieves a sandbox by ID
func (m *Manager) Get(id string) (Sandbox, bool) {
	s, ok := m.sandboxes[id]
	return s, ok
}

// Destroy stops and removes a sandbox
func (m *Manager) Destroy(ctx context.Context, id string) error {
	sandbox, ok := m.sandboxes[id]
	if !ok {
		return fmt.Errorf("sandbox not found: %s", id)
	}

	if err := sandbox.Stop(ctx); err != nil {
		return err
	}

	delete(m.sandboxes, id)
	return nil
}

// DestroyAll stops all sandboxes
func (m *Manager) DestroyAll(ctx context.Context) {
	for id := range m.sandboxes {
		m.Destroy(ctx, id)
	}
}

// SelectLevel chooses appropriate sandbox level based on requirements
func SelectLevel(language string, requiresNetwork bool, trustLevel string) Level {
	// Browser-only languages can use WebContainers
	browserLanguages := map[string]bool{
		"javascript": true,
		"typescript": true,
		"jsx":        true,
		"tsx":        true,
	}

	if browserLanguages[language] && !requiresNetwork && trustLevel == "high" {
		return LevelBrowser
	}

	// Untrusted code goes to microVM
	if trustLevel == "low" || trustLevel == "untrusted" {
		return LevelMicroVM
	}

	// Default to container for most use cases
	return LevelContainer
}
