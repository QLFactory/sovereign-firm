package orchestration

import (
	"context"
	"testing"
	"time"
)

func TestDefaultCIPipelineConfig(t *testing.T) {
	config := DefaultCIPipelineConfig("/tmp/test")

	if config.WorkDir != "/tmp/test" {
		t.Errorf("expected WorkDir '/tmp/test', got '%s'", config.WorkDir)
	}
	if config.Timeout != 15*time.Minute {
		t.Errorf("expected timeout 15m, got %v", config.Timeout)
	}
	if config.LintCommand == "" {
		t.Error("lint command should not be empty")
	}
	if config.BuildCommand == "" {
		t.Error("build command should not be empty")
	}
	if config.TestCommand == "" {
		t.Error("test command should not be empty")
	}
}

func TestNewCIPipeline(t *testing.T) {
	config := DefaultCIPipelineConfig("/tmp/test")
	pipeline := NewCIPipeline(config)

	if pipeline == nil {
		t.Fatal("pipeline should not be nil")
	}
}

func TestRunStage_UnknownStage(t *testing.T) {
	config := DefaultCIPipelineConfig("/tmp/test")
	pipeline := NewCIPipeline(config)
	ctx := context.Background()

	result := pipeline.RunStage(ctx, CIStage("UNKNOWN"), nil)

	if result.Success {
		t.Error("unknown stage should fail")
	}
	if result.Error == "" {
		t.Error("error should be set for unknown stage")
	}
}

func TestRunStage_WithEchoCommand(t *testing.T) {
	config := CIPipelineConfig{
		WorkDir:     "/tmp",
		LintCommand: "echo lint success",
		Timeout:     30 * time.Second,
	}
	pipeline := NewCIPipeline(config)
	ctx := context.Background()

	result := pipeline.RunStage(ctx, StageLint, nil)

	if !result.Success {
		t.Errorf("echo command should succeed: %s", result.Error)
	}
	if result.Duration == 0 {
		t.Error("duration should be set")
	}
}

func TestRunStage_FailingCommand(t *testing.T) {
	config := CIPipelineConfig{
		WorkDir:     "/tmp",
		LintCommand: "false",
		Timeout:     30 * time.Second,
	}
	pipeline := NewCIPipeline(config)
	ctx := context.Background()

	result := pipeline.RunStage(ctx, StageLint, nil)

	if result.Success {
		t.Error("false command should fail")
	}
}

func TestRunPipeline_AllSuccess(t *testing.T) {
	config := CIPipelineConfig{
		WorkDir:      "/tmp",
		LintCommand:  "echo lint",
		BuildCommand: "echo build",
		TestCommand:  "echo test",
		Timeout:      30 * time.Second,
	}
	pipeline := NewCIPipeline(config)
	ctx := context.Background()

	result, err := pipeline.RunPipeline(ctx, nil)
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}

	if !result.Success {
		t.Error("pipeline should succeed")
	}
	if len(result.StageResults) != 3 {
		t.Errorf("expected 3 stage results, got %d", len(result.StageResults))
	}
}

func TestRunPipeline_FailsAtLint(t *testing.T) {
	config := CIPipelineConfig{
		WorkDir:      "/tmp",
		LintCommand:  "false",
		BuildCommand: "echo build",
		TestCommand:  "echo test",
		Timeout:      30 * time.Second,
	}
	pipeline := NewCIPipeline(config)
	ctx := context.Background()

	result, err := pipeline.RunPipeline(ctx, nil)
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}

	if result.Success {
		t.Error("pipeline should fail at lint")
	}
	if result.FailedStage != StageLint {
		t.Errorf("failed stage should be LINT, got %s", result.FailedStage)
	}
	// Should only have 1 stage result (stopped at lint)
	if len(result.StageResults) != 1 {
		t.Errorf("expected 1 stage result, got %d", len(result.StageResults))
	}
}

func TestRunPipeline_SkipStages(t *testing.T) {
	config := CIPipelineConfig{
		WorkDir:      "/tmp",
		LintCommand:  "echo lint",
		BuildCommand: "echo build",
		TestCommand:  "echo test",
		Timeout:      30 * time.Second,
		SkipStages:   []CIStage{StageLint},
	}
	pipeline := NewCIPipeline(config)
	ctx := context.Background()

	result, err := pipeline.RunPipeline(ctx, nil)
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}

	if !result.Success {
		t.Error("pipeline should succeed")
	}
	// Should have 2 stage results (lint skipped)
	if len(result.StageResults) != 2 {
		t.Errorf("expected 2 stage results, got %d", len(result.StageResults))
	}
}

func TestExtractFailedFiles(t *testing.T) {
	config := DefaultCIPipelineConfig("/tmp")
	pipeline := NewCIPipeline(config)

	output := `
pkg/api/handler.go:10:5: undefined: foo
pkg/api/handler.go:15:3: unused variable
pkg/service/user.go:20:1: missing return
`
	changedFiles := []string{"pkg/api/handler.go", "pkg/service/user.go", "pkg/db/conn.go"}

	failed := pipeline.extractFailedFiles(output, changedFiles)

	if len(failed) != 2 {
		t.Errorf("expected 2 failed files, got %d", len(failed))
	}
}

func TestDefaultSelfCorrectionConfig(t *testing.T) {
	config := DefaultSelfCorrectionConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("expected max attempts 3, got %d", config.MaxAttempts)
	}
	if config.RetryDelay != 5*time.Second {
		t.Errorf("expected retry delay 5s, got %v", config.RetryDelay)
	}
}

func TestNewSelfCorrectionLoop(t *testing.T) {
	pipelineConfig := DefaultCIPipelineConfig("/tmp")
	pipeline := NewCIPipeline(pipelineConfig)
	config := DefaultSelfCorrectionConfig()

	loop := NewSelfCorrectionLoop(config, pipeline)

	if loop == nil {
		t.Fatal("loop should not be nil")
	}
}

func TestRunWithCorrection_ImmediateSuccess(t *testing.T) {
	pipelineConfig := CIPipelineConfig{
		WorkDir:      "/tmp",
		LintCommand:  "echo lint",
		BuildCommand: "echo build",
		TestCommand:  "echo test",
		Timeout:      30 * time.Second,
	}
	pipeline := NewCIPipeline(pipelineConfig)
	config := DefaultSelfCorrectionConfig()
	loop := NewSelfCorrectionLoop(config, pipeline)

	ctx := context.Background()
	applyFix := func(feedback CorrectionFeedback) ([]string, error) {
		return nil, nil
	}

	result, err := loop.RunWithCorrection(ctx, nil, applyFix)
	if err != nil {
		t.Fatalf("correction loop error: %v", err)
	}

	if !result.Success {
		t.Error("should succeed on first attempt")
	}
	if result.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", result.Attempts)
	}
}

func TestRunWithCorrection_FailsAfterMaxAttempts(t *testing.T) {
	pipelineConfig := CIPipelineConfig{
		WorkDir:      "/tmp",
		LintCommand:  "false",
		BuildCommand: "echo build",
		TestCommand:  "echo test",
		Timeout:      30 * time.Second,
	}
	pipeline := NewCIPipeline(pipelineConfig)
	config := SelfCorrectionConfig{
		MaxAttempts: 2,
		RetryDelay:  10 * time.Millisecond,
	}
	loop := NewSelfCorrectionLoop(config, pipeline)

	ctx := context.Background()
	fixCount := 0
	applyFix := func(feedback CorrectionFeedback) ([]string, error) {
		fixCount++
		return []string{"fix-" + string(rune('0'+fixCount))}, nil
	}

	result, err := loop.RunWithCorrection(ctx, nil, applyFix)
	if err != nil {
		t.Fatalf("correction loop error: %v", err)
	}

	if result.Success {
		t.Error("should fail after max attempts")
	}
	if result.Attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", result.Attempts)
	}
	// Should have 1 fix applied (not called on last attempt)
	if len(result.FixesApplied) != 1 {
		t.Errorf("expected 1 fix applied, got %d", len(result.FixesApplied))
	}
}

func TestGenerateSuggestions_Lint(t *testing.T) {
	config := DefaultSelfCorrectionConfig()
	loop := NewSelfCorrectionLoop(config, nil)

	suggestions := loop.generateSuggestions(StageLint, "ineffectual assignment")

	if len(suggestions) == 0 {
		t.Error("should generate suggestions for lint")
	}

	found := false
	for _, s := range suggestions {
		if s == "Remove or use the assigned variable" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should include specific suggestion for ineffectual assignment")
	}
}

func TestGenerateSuggestions_Build(t *testing.T) {
	config := DefaultSelfCorrectionConfig()
	loop := NewSelfCorrectionLoop(config, nil)

	suggestions := loop.generateSuggestions(StageBuild, "undefined: foo")

	if len(suggestions) == 0 {
		t.Error("should generate suggestions for build")
	}

	found := false
	for _, s := range suggestions {
		if s == "Add missing function or variable definition" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should include specific suggestion for undefined")
	}
}

func TestGenerateSuggestions_Test(t *testing.T) {
	config := DefaultSelfCorrectionConfig()
	loop := NewSelfCorrectionLoop(config, nil)

	suggestions := loop.generateSuggestions(StageTest, "panic: runtime error")

	if len(suggestions) == 0 {
		t.Error("should generate suggestions for test")
	}

	found := false
	for _, s := range suggestions {
		if s == "Fix nil pointer dereference or other panic causes" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should include specific suggestion for panic")
	}
}

func TestValidateTaskOutput(t *testing.T) {
	output := map[string]interface{}{
		"files": []string{"file1.go", "file2.go", "file3.go"},
	}
	expectedFiles := []string{"file1.go", "file2.go"}

	valid, missing := ValidateTaskOutput(output, expectedFiles)

	if !valid {
		t.Error("output should be valid")
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing files, got %v", missing)
	}
}

func TestValidateTaskOutput_MissingFiles(t *testing.T) {
	output := map[string]interface{}{
		"files": []string{"file1.go"},
	}
	expectedFiles := []string{"file1.go", "file2.go", "file3.go"}

	valid, missing := ValidateTaskOutput(output, expectedFiles)

	if valid {
		t.Error("output should be invalid with missing files")
	}
	if len(missing) != 2 {
		t.Errorf("expected 2 missing files, got %d", len(missing))
	}
}

func TestValidateTaskOutput_NoFilesKey(t *testing.T) {
	output := map[string]interface{}{
		"status": "completed",
	}
	expectedFiles := []string{"file1.go"}

	valid, missing := ValidateTaskOutput(output, expectedFiles)

	if valid {
		t.Error("output should be invalid without files key")
	}
	if len(missing) != 1 {
		t.Errorf("expected 1 missing file, got %d", len(missing))
	}
}

func TestCIPipelineResult_TotalTime(t *testing.T) {
	pipelineConfig := CIPipelineConfig{
		WorkDir:      "/tmp",
		LintCommand:  "echo lint",
		BuildCommand: "echo build",
		TestCommand:  "echo test",
		Timeout:      30 * time.Second,
	}
	pipeline := NewCIPipeline(pipelineConfig)
	ctx := context.Background()

	result, err := pipeline.RunPipeline(ctx, nil)
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}

	// Just verify TotalTime is set (non-zero)
	if result.TotalTime == 0 {
		t.Error("total time should be non-zero")
	}
}

func TestCorrectionFeedback(t *testing.T) {
	feedback := CorrectionFeedback{
		Attempt:     1,
		MaxAttempts: 3,
		FailedStage: StageLint,
		ErrorOutput: "error in file.go",
		FailedFiles: []string{"file.go"},
		Suggestions: []string{"Fix the error"},
	}

	if feedback.Attempt != 1 {
		t.Errorf("expected attempt 1, got %d", feedback.Attempt)
	}
	if feedback.FailedStage != StageLint {
		t.Errorf("expected stage LINT, got %s", feedback.FailedStage)
	}
	if len(feedback.FailedFiles) != 1 {
		t.Errorf("expected 1 failed file, got %d", len(feedback.FailedFiles))
	}
}
