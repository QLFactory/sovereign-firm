package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

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
