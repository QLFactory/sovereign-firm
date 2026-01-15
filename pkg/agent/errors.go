package agent

import "errors"

var (
	// ErrAgentNotFound is returned when the requested agent doesn't exist
	ErrAgentNotFound = errors.New("agent not found")

	// ErrAgentBusy is returned when trying to assign work to a busy agent
	ErrAgentBusy = errors.New("agent is busy with another task")

	// ErrNoActiveTask is returned when completing a task but none is active
	ErrNoActiveTask = errors.New("no active task to complete")

	// ErrInvalidStatusTransition is returned for invalid state transitions
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// ErrPoolFull is returned when the agent pool has reached capacity
	ErrPoolFull = errors.New("agent pool is at capacity")

	// ErrSkillNotFound is returned when a required skill is not loaded
	ErrSkillNotFound = errors.New("skill not found")

	// ErrInvalidSkillYAML is returned when skill YAML is malformed
	ErrInvalidSkillYAML = errors.New("invalid skill YAML format")

	// ErrSOCValidationFailed is returned when output fails schema validation
	ErrSOCValidationFailed = errors.New("structured output contract validation failed")

	// ErrMaxRetriesExceeded is returned when SOC retry limit is reached
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded for valid output")
)
