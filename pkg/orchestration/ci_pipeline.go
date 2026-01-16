package orchestration

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// CIStage represents a stage in the CI pipeline
type CIStage string

const (
	StageLint     CIStage = "LINT"
	StageBuild    CIStage = "BUILD"
	StageTest     CIStage = "TEST"
	StageValidate CIStage = "VALIDATE"
)

// CIResult represents the result of a CI pipeline run
type CIResult struct {
	Success     bool          `json:"success"`
	Stage       CIStage       `json:"stage"`
	Output      string        `json:"output"`
	Error       string        `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
	FailedFiles []string      `json:"failed_files,omitempty"`
}

// CIPipelineResult represents the overall pipeline result
type CIPipelineResult struct {
	Success      bool        `json:"success"`
	StageResults []CIResult  `json:"stage_results"`
	TotalTime    time.Duration `json:"total_time"`
	FailedStage  CIStage     `json:"failed_stage,omitempty"`
}

// CIPipelineConfig configures the CI pipeline
type CIPipelineConfig struct {
	WorkDir       string            `json:"work_dir"`
	LintCommand   string            `json:"lint_command"`
	BuildCommand  string            `json:"build_command"`
	TestCommand   string            `json:"test_command"`
	Timeout       time.Duration     `json:"timeout"`
	Environment   map[string]string `json:"environment"`
	SkipStages    []CIStage         `json:"skip_stages"`
}

// DefaultCIPipelineConfig returns sensible defaults for Go projects
func DefaultCIPipelineConfig(workDir string) CIPipelineConfig {
	return CIPipelineConfig{
		WorkDir:      workDir,
		LintCommand:  "golangci-lint run --timeout 5m",
		BuildCommand: "go build ./...",
		TestCommand:  "go test -v -race ./...",
		Timeout:      15 * time.Minute,
		Environment:  make(map[string]string),
		SkipStages:   make([]CIStage, 0),
	}
}

// CIPipeline runs CI checks on agent-generated code
type CIPipeline struct {
	config CIPipelineConfig
	mu     sync.Mutex
}

// NewCIPipeline creates a new CI pipeline
func NewCIPipeline(config CIPipelineConfig) *CIPipeline {
	return &CIPipeline{
		config: config,
	}
}

// RunPipeline executes the full CI pipeline
func (p *CIPipeline) RunPipeline(ctx context.Context, files []string) (*CIPipelineResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	start := time.Now()
	result := &CIPipelineResult{
		StageResults: make([]CIResult, 0),
	}

	// Define pipeline stages in order
	stages := []struct {
		stage   CIStage
		command string
	}{
		{StageLint, p.config.LintCommand},
		{StageBuild, p.config.BuildCommand},
		{StageTest, p.config.TestCommand},
	}

	for _, s := range stages {
		if p.shouldSkipStage(s.stage) {
			continue
		}

		stageResult := p.runStage(ctx, s.stage, s.command, files)
		result.StageResults = append(result.StageResults, stageResult)

		if !stageResult.Success {
			result.Success = false
			result.FailedStage = s.stage
			result.TotalTime = time.Since(start)
			return result, nil
		}
	}

	result.Success = true
	result.TotalTime = time.Since(start)
	return result, nil
}

// RunStage runs a single CI stage
func (p *CIPipeline) RunStage(ctx context.Context, stage CIStage, files []string) CIResult {
	p.mu.Lock()
	defer p.mu.Unlock()

	var command string
	switch stage {
	case StageLint:
		command = p.config.LintCommand
	case StageBuild:
		command = p.config.BuildCommand
	case StageTest:
		command = p.config.TestCommand
	default:
		return CIResult{
			Success: false,
			Stage:   stage,
			Error:   fmt.Sprintf("unknown stage: %s", stage),
		}
	}

	return p.runStage(ctx, stage, command, files)
}

// runStage executes a single stage
func (p *CIPipeline) runStage(ctx context.Context, stage CIStage, command string, files []string) CIResult {
	start := time.Now()
	result := CIResult{
		Stage: stage,
	}

	// Create command with timeout
	stageCtx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	// Split command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		result.Error = "empty command"
		return result
	}

	cmd := exec.CommandContext(stageCtx, parts[0], parts[1:]...)
	cmd.Dir = p.config.WorkDir

	// Set environment
	for k, v := range p.config.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Run command and capture output
	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	result.Duration = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.FailedFiles = p.extractFailedFiles(string(output), files)
	} else {
		result.Success = true
	}

	return result
}

// shouldSkipStage checks if a stage should be skipped
func (p *CIPipeline) shouldSkipStage(stage CIStage) bool {
	for _, s := range p.config.SkipStages {
		if s == stage {
			return true
		}
	}
	return false
}

// extractFailedFiles attempts to extract failed file names from output
func (p *CIPipeline) extractFailedFiles(output string, changedFiles []string) []string {
	var failed []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		for _, file := range changedFiles {
			if strings.Contains(line, file) {
				failed = append(failed, file)
				break
			}
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, f := range failed {
		if !seen[f] {
			seen[f] = true
			unique = append(unique, f)
		}
	}

	return unique
}

// SelfCorrectionConfig configures the self-correction loop
type SelfCorrectionConfig struct {
	MaxAttempts       int           `json:"max_attempts"`
	RetryDelay        time.Duration `json:"retry_delay"`
	AnalyzeFailures   bool          `json:"analyze_failures"`
	IncrementalFix    bool          `json:"incremental_fix"`
}

// DefaultSelfCorrectionConfig returns sensible defaults
func DefaultSelfCorrectionConfig() SelfCorrectionConfig {
	return SelfCorrectionConfig{
		MaxAttempts:     3,
		RetryDelay:      5 * time.Second,
		AnalyzeFailures: true,
		IncrementalFix:  true,
	}
}

// CorrectionFeedback provides feedback to the agent for correction
type CorrectionFeedback struct {
	Attempt       int      `json:"attempt"`
	MaxAttempts   int      `json:"max_attempts"`
	FailedStage   CIStage  `json:"failed_stage"`
	ErrorOutput   string   `json:"error_output"`
	FailedFiles   []string `json:"failed_files"`
	Suggestions   []string `json:"suggestions"`
	PreviousFixes []string `json:"previous_fixes,omitempty"`
}

// SelfCorrectionLoop handles agent self-correction based on CI failures
type SelfCorrectionLoop struct {
	config   SelfCorrectionConfig
	pipeline *CIPipeline
	mu       sync.Mutex
}

// NewSelfCorrectionLoop creates a new self-correction loop
func NewSelfCorrectionLoop(config SelfCorrectionConfig, pipeline *CIPipeline) *SelfCorrectionLoop {
	return &SelfCorrectionLoop{
		config:   config,
		pipeline: pipeline,
	}
}

// CorrectionResult represents the result of self-correction
type CorrectionResult struct {
	Success       bool              `json:"success"`
	Attempts      int               `json:"attempts"`
	FinalResult   *CIPipelineResult `json:"final_result"`
	FixesApplied  []string          `json:"fixes_applied"`
}

// RunWithCorrection runs CI and attempts self-correction on failure
func (s *SelfCorrectionLoop) RunWithCorrection(
	ctx context.Context,
	files []string,
	applyFix func(CorrectionFeedback) ([]string, error),
) (*CorrectionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := &CorrectionResult{
		FixesApplied: make([]string, 0),
	}

	var previousFixes []string

	for attempt := 1; attempt <= s.config.MaxAttempts; attempt++ {
		result.Attempts = attempt

		// Run CI pipeline
		ciResult, err := s.pipeline.RunPipeline(ctx, files)
		if err != nil {
			return result, fmt.Errorf("CI pipeline error: %w", err)
		}

		result.FinalResult = ciResult

		if ciResult.Success {
			result.Success = true
			return result, nil
		}

		// Check if we have more attempts
		if attempt >= s.config.MaxAttempts {
			break
		}

		// Create feedback for agent
		feedback := s.createFeedback(ciResult, attempt, previousFixes)

		// Apply fix through callback
		fixes, err := applyFix(feedback)
		if err != nil {
			return result, fmt.Errorf("failed to apply fix: %w", err)
		}

		result.FixesApplied = append(result.FixesApplied, fixes...)
		previousFixes = append(previousFixes, fixes...)

		// Wait before retry
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(s.config.RetryDelay):
		}
	}

	return result, nil
}

// createFeedback creates correction feedback from CI result
func (s *SelfCorrectionLoop) createFeedback(result *CIPipelineResult, attempt int, previousFixes []string) CorrectionFeedback {
	feedback := CorrectionFeedback{
		Attempt:       attempt,
		MaxAttempts:   s.config.MaxAttempts,
		FailedStage:   result.FailedStage,
		PreviousFixes: previousFixes,
	}

	// Find the failed stage result
	for _, stageResult := range result.StageResults {
		if stageResult.Stage == result.FailedStage {
			feedback.ErrorOutput = stageResult.Output
			feedback.FailedFiles = stageResult.FailedFiles
			break
		}
	}

	// Generate suggestions based on stage
	feedback.Suggestions = s.generateSuggestions(result.FailedStage, feedback.ErrorOutput)

	return feedback
}

// generateSuggestions creates fix suggestions based on failure
func (s *SelfCorrectionLoop) generateSuggestions(stage CIStage, output string) []string {
	var suggestions []string

	switch stage {
	case StageLint:
		suggestions = append(suggestions,
			"Review linting errors and fix code style issues",
			"Run 'go fmt' to fix formatting",
			"Check for unused variables and imports",
		)
		if strings.Contains(output, "ineffectual assignment") {
			suggestions = append(suggestions, "Remove or use the assigned variable")
		}
		if strings.Contains(output, "undeclared name") {
			suggestions = append(suggestions, "Add missing imports or declarations")
		}

	case StageBuild:
		suggestions = append(suggestions,
			"Fix compilation errors in the reported files",
			"Check for missing imports",
			"Verify type compatibility",
		)
		if strings.Contains(output, "undefined:") {
			suggestions = append(suggestions, "Add missing function or variable definition")
		}
		if strings.Contains(output, "cannot convert") {
			suggestions = append(suggestions, "Fix type conversion issues")
		}

	case StageTest:
		suggestions = append(suggestions,
			"Review failing tests and fix the code or test assertions",
			"Check for race conditions if using -race flag",
			"Verify test data and mocks are correct",
		)
		if strings.Contains(output, "panic:") {
			suggestions = append(suggestions, "Fix nil pointer dereference or other panic causes")
		}
		if strings.Contains(output, "timeout") {
			suggestions = append(suggestions, "Check for deadlocks or infinite loops")
		}
	}

	return suggestions
}

// ValidateTaskOutput validates the output of a task against expectations
func ValidateTaskOutput(output map[string]interface{}, expectedFiles []string) (bool, []string) {
	var missing []string

	files, ok := output["files"].([]string)
	if !ok {
		return false, expectedFiles
	}

	fileSet := make(map[string]bool)
	for _, f := range files {
		fileSet[f] = true
	}

	for _, expected := range expectedFiles {
		if !fileSet[expected] {
			missing = append(missing, expected)
		}
	}

	return len(missing) == 0, missing
}
