package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ContainerSandbox implements Sandbox using Docker CLI
// This approach avoids SDK version conflicts and is simpler to maintain.
type ContainerSandbox struct {
	config      *Config
	containerID string
	workDir     string
	running     bool
	mu          sync.RWMutex
}

// NewContainerSandbox creates a new Docker-based sandbox
func NewContainerSandbox(config *Config) (*ContainerSandbox, error) {
	// Verify Docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, fmt.Errorf("docker not found in PATH")
	}

	// Create temporary work directory
	workDir := config.WorkDir
	if workDir == "" {
		var err error
		workDir, err = os.MkdirTemp("", "sandbox-"+config.ID+"-")
		if err != nil {
			return nil, fmt.Errorf("failed to create work directory: %w", err)
		}
	}

	return &ContainerSandbox{
		config:  config,
		workDir: workDir,
	}, nil
}

// Level returns the isolation level
func (s *ContainerSandbox) Level() Level {
	return LevelContainer
}

// Start initializes the container
func (s *ContainerSandbox) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	image := s.getImage()

	// Build docker run command
	args := []string{
		"run", "-d",
		"--name", "sandbox-" + s.config.ID,
		"-w", "/workspace",
		"-v", s.workDir + ":/workspace",
		// Resource limits
		"--memory", fmt.Sprintf("%dm", s.config.MaxMemory),
		"--cpu-period", "100000",
		"--cpu-quota", fmt.Sprintf("%d", s.config.MaxCPU*1000),
		// Security
		"--security-opt", "no-new-privileges",
		"--cap-drop", "ALL",
		"--cap-add", "CHOWN",
		"--cap-add", "SETUID",
		"--cap-add", "SETGID",
	}

	// Network restrictions
	if !s.config.NetworkEnabled {
		args = append(args, "--network", "none")
	}

	// Try gVisor runtime if available
	gvisorAvailable := s.checkRuntime("runsc")
	if gvisorAvailable {
		args = append(args, "--runtime", "runsc")
	}

	// Image and command
	args = append(args, image, "tail", "-f", "/dev/null")

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Retry without gVisor if it failed
		if gvisorAvailable && strings.Contains(string(output), "runtime") {
			args = removeRuntimeArg(args)
			cmd = exec.CommandContext(ctx, "docker", args...)
			output, err = cmd.CombinedOutput()
		}
		if err != nil {
			return fmt.Errorf("failed to start container: %s - %w", string(output), err)
		}
	}

	s.containerID = strings.TrimSpace(string(output))
	s.running = true
	return nil
}

// Stop terminates the container
func (s *ContainerSandbox) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.containerID == "" {
		return nil
	}

	// Stop container
	exec.CommandContext(ctx, "docker", "stop", "-t", "5", s.containerID).Run()

	// Remove container
	exec.CommandContext(ctx, "docker", "rm", "-f", s.containerID).Run()

	// Cleanup work directory
	os.RemoveAll(s.workDir)

	s.running = false
	s.containerID = ""
	return nil
}

// IsRunning checks if the container is active
func (s *ContainerSandbox) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// Execute runs a command in the container
func (s *ContainerSandbox) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResult, error) {
	s.mu.RLock()
	if !s.running {
		s.mu.RUnlock()
		return nil, fmt.Errorf("sandbox not running")
	}
	containerID := s.containerID
	s.mu.RUnlock()

	// Write any files first
	for path, content := range req.Files {
		if err := s.WriteFile(ctx, path, []byte(content)); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	// Determine timeout
	timeout := req.Timeout
	if timeout == 0 {
		timeout = s.config.MaxTimeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	startTime := time.Now()

	// Build exec command
	args := []string{"exec"}

	// Add environment variables
	for k, v := range req.Env {
		args = append(args, "-e", k+"="+v)
	}

	args = append(args, containerID, "sh", "-c", req.Command)

	cmd := exec.CommandContext(execCtx, "docker", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if req.Stdin != "" {
		cmd.Stdin = strings.NewReader(req.Stdin)
	}

	err := cmd.Run()
	duration := time.Since(startTime)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if execCtx.Err() == context.DeadlineExceeded {
			return &ExecuteResult{
				Success:  false,
				ExitCode: -1,
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				Duration: duration,
				Error:    "execution timeout",
			}, nil
		}
	}

	return &ExecuteResult{
		Success:  exitCode == 0,
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}, nil
}

// WriteFile writes a file to the sandbox filesystem
func (s *ContainerSandbox) WriteFile(ctx context.Context, path string, content []byte) error {
	fullPath := filepath.Join(s.workDir, path)

	// Ensure parent directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(fullPath, content, 0644)
}

// ReadFile reads a file from the sandbox filesystem
func (s *ContainerSandbox) ReadFile(ctx context.Context, path string) ([]byte, error) {
	fullPath := filepath.Join(s.workDir, path)
	return os.ReadFile(fullPath)
}

// ListFiles lists files in a directory
func (s *ContainerSandbox) ListFiles(ctx context.Context, path string) ([]string, error) {
	fullPath := filepath.Join(s.workDir, path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, e := range entries {
		files = append(files, e.Name())
	}
	return files, nil
}

// GetConfig returns the sandbox configuration
func (s *ContainerSandbox) GetConfig() *Config {
	return s.config
}

// getImage returns the Docker image for the runtime
func (s *ContainerSandbox) getImage() string {
	if s.config.Runtime != "" {
		return s.config.Runtime
	}

	switch s.config.Language {
	case "javascript", "typescript", "jsx", "tsx":
		return "node:20-slim"
	case "go":
		return "golang:1.21-alpine"
	case "python":
		return "python:3.11-slim"
	case "rust":
		return "rust:1.75-slim"
	case "java":
		return "eclipse-temurin:21-jdk"
	default:
		return "ubuntu:22.04"
	}
}

// checkRuntime checks if a Docker runtime is available
func (s *ContainerSandbox) checkRuntime(runtime string) bool {
	cmd := exec.Command("docker", "info", "--format", "{{.Runtimes}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), runtime)
}

// removeRuntimeArg removes the --runtime argument from args
func removeRuntimeArg(args []string) []string {
	result := make([]string, 0, len(args))
	skip := false
	for _, arg := range args {
		if arg == "--runtime" {
			skip = true
			continue
		}
		if skip {
			skip = false
			continue
		}
		result = append(result, arg)
	}
	return result
}

// Ensure ContainerSandbox implements Sandbox
var _ Sandbox = (*ContainerSandbox)(nil)
