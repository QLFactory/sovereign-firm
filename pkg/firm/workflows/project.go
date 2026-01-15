package workflows

import (
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"
)

// isTestFile checks if a filename is a test file
func isTestFile(filename string) bool {
	return strings.Contains(filename, ".test.") || strings.Contains(filename, "_test.") || strings.Contains(filename, ".spec.")
}

type ProjectState struct {
	Name        string            `json:"name"`
	Phase       string            `json:"phase"`
	ChatHistory string            `json:"chat_history"`
	Spec        string            `json:"spec"`
	CodeFiles   map[string]string `json:"code_files"`
}

type UserMessageSignal struct {
	Message string
}

// ProjectLifecycle manages the entire lifecycle of a consultancy project
func ProjectLifecycle(ctx workflow.Context, input interface{}) (*ProjectState, error) {
	logger := workflow.GetLogger(ctx)
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	state := &ProjectState{
		Phase:     "DISCOVERY",
		CodeFiles: make(map[string]string),
	}

	// Register Query Handler
	err := workflow.SetQueryHandler(ctx, "get_state", func() (*ProjectState, error) {
		return state, nil
	})
	if err != nil {
		return nil, err
	}

	logger.Info("Starting Discovery Phase")
	msgChan := workflow.GetSignalChannel(ctx, "USER_MESSAGE")

	// 1. Discovery Loop
	for {
		var signal UserMessageSignal
		msgChan.Receive(ctx, &signal)

		if signal.Message == "/approve" {
			break
		}

		state.ChatHistory += "\nUser: " + signal.Message

		var reply string
		err := workflow.ExecuteActivity(ctx, "PMAgentChat", state.ChatHistory).Get(ctx, &reply)
		if err != nil {
			logger.Error("PM Agent failed", "Error", err)
			continue
		}

		state.ChatHistory += "\nPM: " + reply
	}

	// 2. Implementation Loop
	for {
		state.Phase = "IMPLEMENTATION"

		var codeBundle map[string]string
		// Use "RefineCode" activity if we have existing code, or just "Generate"
		// For now, let's reuse Generate but pass the specific feedback.
		// Actually, standardizing on a "DevAgentGenerate" that takes context is fine.
		// Detailed refinement might need a new activity, but let's stick to the plan:
		// Plan says "Implement DevAgent.RefineCode activity".
		// Let's use "DevAgentRefine" if state.CodeFiles is not empty, else "DevAgentGenerate".

		var actName string
		var input interface{}

		if len(state.CodeFiles) > 0 {
			actName = "DevAgentRefine"
			// We need a complex input for Refine: Code + Changes
			// For simplicity in this step, let's serialize arguments or create a struct in activities.
			// Let's pass a struct. But we are in workflows package.
			// Let's pass a map for flexibility since we don't want strict coupling yet
			input = map[string]interface{}{
				"current_code": state.CodeFiles,
				"chat_history": state.ChatHistory,
			}
		} else {
			actName = "DevAgentGenerate"
			input = state.ChatHistory // or Spec. The original code used ChatHistory.
		}

		err = workflow.ExecuteActivity(ctx, actName, input).Get(ctx, &codeBundle)
		if err != nil {
			logger.Error("Dev Agent failed attempting retry...", "Error", err)
			// In a real refined workflow, we might ask user to retry or fail.
			// For now, let's break or return error.
			return nil, err
		}

		state.CodeFiles = codeBundle

		// Trigger QA Agent with Three-Strike Rule Feedback Loop
		var testBundle map[string]string

		specToUse := state.Spec
		if specToUse == "" {
			specToUse = state.ChatHistory
		}

		qaInput := map[string]interface{}{
			"spec":       specToUse,
			"code_files": state.CodeFiles,
		}
		aoQA := workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute * 2,
		}
		ctxQA := workflow.WithActivityOptions(ctx, aoQA)
		if err := workflow.ExecuteActivity(ctxQA, "QAAgentGenerateTests", qaInput).Get(ctx, &testBundle); err != nil {
			logger.Error("QA Agent failed", "Error", err)
		} else {
			// Add tests to code files
			for k, v := range testBundle {
				state.CodeFiles[k] = v
			}

			// Three-Strike Rule: Run tests and retry up to 3 times if they fail
			const maxTestAttempts = 3
			testsPassed := false

			for attempt := 1; attempt <= maxTestAttempts && !testsPassed; attempt++ {
				logger.Info("Running tests", "Attempt", attempt)

				// Prepare test input with all code files (including tests)
				testRunInput := map[string]interface{}{
					"code_files": state.CodeFiles,
				}

				aoTest := workflow.ActivityOptions{
					StartToCloseTimeout: time.Minute * 5, // Tests may take longer
				}
				ctxTest := workflow.WithActivityOptions(ctx, aoTest)

				var testResult map[string]interface{}
				if err := workflow.ExecuteActivity(ctxTest, "RunTests", testRunInput).Get(ctx, &testResult); err != nil {
					logger.Error("Test runner failed", "Error", err, "Attempt", attempt)
					continue
				}

				// Check if tests passed
				if success, ok := testResult["success"].(bool); ok && success {
					testsPassed = true
					logger.Info("Tests passed!", "Attempt", attempt)
				} else {
					// Tests failed - extract error output
					testOutput := ""
					if output, ok := testResult["output"].(string); ok {
						testOutput = output
					}

					logger.Warn("Tests failed", "Attempt", attempt, "Output", testOutput)

					// If not the last attempt, regenerate tests with feedback
					if attempt < maxTestAttempts {
						logger.Info("Regenerating tests with feedback", "Attempt", attempt+1)

						// Extract current test files from state.CodeFiles
						currentTestFiles := make(map[string]string)
						codeOnlyFiles := make(map[string]string)
						for k, v := range state.CodeFiles {
							if isTestFile(k) {
								currentTestFiles[k] = v
							} else {
								codeOnlyFiles[k] = v
							}
						}

						// Call QA Agent regenerate with error feedback
						regenerateInput := map[string]interface{}{
							"spec":        specToUse,
							"code_files":  codeOnlyFiles,
							"test_files":  currentTestFiles,
							"test_output": testOutput,
							"attempt":     attempt + 1,
						}

						var regeneratedTests map[string]string
						if err := workflow.ExecuteActivity(ctxQA, "QAAgentRegenerateTests", regenerateInput).Get(ctx, &regeneratedTests); err != nil {
							logger.Error("QA Agent regeneration failed", "Error", err, "Attempt", attempt)
						} else {
							// Replace test files with regenerated ones
							for k := range currentTestFiles {
								delete(state.CodeFiles, k)
							}
							for k, v := range regeneratedTests {
								state.CodeFiles[k] = v
							}
						}
					}
				}
			}

			if !testsPassed {
				logger.Warn("Tests failed after 3 attempts - continuing to review phase for human intervention")
			}
		}

		state.Phase = "REVIEW"
		// Wait for feedback
		var signal UserMessageSignal
		msgChan.Receive(ctx, &signal)

		if signal.Message == "/approve" {
			break
		}

		// If not approved, add feedback to history and loop
		state.ChatHistory += "\nUser (Feedback): " + signal.Message
		state.Phase = "REFINEMENT"
	}

	state.Phase = "DONE"
	return state, nil
}
