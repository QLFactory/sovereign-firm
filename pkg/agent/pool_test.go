package agent

import (
	"context"
	"testing"
	"time"
)

func TestDefaultPoolConfig(t *testing.T) {
	config := DefaultPoolConfig()

	if config.MaxAgents != 100 {
		t.Errorf("Expected MaxAgents 100, got %d", config.MaxAgents)
	}

	if config.MaxAgentsPerProject != 10 {
		t.Errorf("Expected MaxAgentsPerProject 10, got %d", config.MaxAgentsPerProject)
	}

	if config.IdleTimeout != 15*time.Minute {
		t.Errorf("Expected IdleTimeout 15m, got %v", config.IdleTimeout)
	}

	if config.CleanupInterval != 1*time.Minute {
		t.Errorf("Expected CleanupInterval 1m, got %v", config.CleanupInterval)
	}
}

func TestNewAgentPool(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	pool := NewAgentPool(DefaultPoolConfig(), skills, &mockLLMClient{})

	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	total, active, idle, terminated := pool.GetStats()
	if total != 0 || active != 0 || idle != 0 || terminated != 0 {
		t.Error("Expected empty pool on creation")
	}

	pool.Shutdown()
}

func TestAgentPoolSpawnAgent(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	skills.RegisterSkill(&Skill{Name: "test-skill", SystemPrompt: "Test"})

	pool := NewAgentPool(DefaultPoolConfig(), skills, &mockLLMClient{})
	defer pool.Shutdown()

	ctx := context.Background()
	agent, err := pool.SpawnAgent(ctx, "Alice", "developer", "project-1", "gpt-4", []string{"test-skill"})

	if err != nil {
		t.Fatalf("SpawnAgent failed: %v", err)
	}

	if agent == nil {
		t.Fatal("Expected agent to be created")
	}

	if agent.Name != "Alice" {
		t.Errorf("Expected name 'Alice', got '%s'", agent.Name)
	}

	if agent.Role != "developer" {
		t.Errorf("Expected role 'developer', got '%s'", agent.Role)
	}

	if agent.Status != StatusReady {
		t.Errorf("Expected status READY, got '%s'", agent.Status)
	}

	total, _, _, _ := pool.GetStats()
	if total != 1 {
		t.Errorf("Expected 1 agent in pool, got %d", total)
	}
}

func TestAgentPoolSpawnAgentWithInvalidSkill(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	pool := NewAgentPool(DefaultPoolConfig(), skills, &mockLLMClient{})
	defer pool.Shutdown()

	ctx := context.Background()
	_, err := pool.SpawnAgent(ctx, "Bob", "dev", "project-1", "gpt-4", []string{"non-existent-skill"})

	if err != ErrSkillNotFound {
		t.Errorf("Expected ErrSkillNotFound, got %v", err)
	}
}

func TestAgentPoolMaxAgentsLimit(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	skills.RegisterSkill(&Skill{Name: "test-skill"})

	config := PoolConfig{
		MaxAgents:           2,
		MaxAgentsPerProject: 10,
		IdleTimeout:         15 * time.Minute,
		CleanupInterval:     1 * time.Minute,
	}

	pool := NewAgentPool(config, skills, &mockLLMClient{})
	defer pool.Shutdown()

	ctx := context.Background()

	// Spawn 2 agents (at limit)
	_, err := pool.SpawnAgent(ctx, "Agent1", "dev", "project-1", "gpt-4", []string{"test-skill"})
	if err != nil {
		t.Fatalf("First spawn failed: %v", err)
	}

	_, err = pool.SpawnAgent(ctx, "Agent2", "dev", "project-2", "gpt-4", []string{"test-skill"})
	if err != nil {
		t.Fatalf("Second spawn failed: %v", err)
	}

	// Third agent should fail - pool full
	_, err = pool.SpawnAgent(ctx, "Agent3", "dev", "project-3", "gpt-4", []string{"test-skill"})
	if err != ErrPoolFull {
		t.Errorf("Expected ErrPoolFull, got %v", err)
	}
}

func TestAgentPoolMaxAgentsPerProjectLimit(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	skills.RegisterSkill(&Skill{Name: "test-skill"})

	config := PoolConfig{
		MaxAgents:           100,
		MaxAgentsPerProject: 2,
		IdleTimeout:         15 * time.Minute,
		CleanupInterval:     1 * time.Minute,
	}

	pool := NewAgentPool(config, skills, &mockLLMClient{})
	defer pool.Shutdown()

	ctx := context.Background()

	// Spawn 2 agents for same project
	_, err := pool.SpawnAgent(ctx, "Agent1", "dev", "project-1", "gpt-4", []string{"test-skill"})
	if err != nil {
		t.Fatalf("First spawn failed: %v", err)
	}

	_, err = pool.SpawnAgent(ctx, "Agent2", "dev", "project-1", "gpt-4", []string{"test-skill"})
	if err != nil {
		t.Fatalf("Second spawn failed: %v", err)
	}

	// Third agent for same project should fail
	_, err = pool.SpawnAgent(ctx, "Agent3", "dev", "project-1", "gpt-4", []string{"test-skill"})
	if err != ErrPoolFull {
		t.Errorf("Expected ErrPoolFull for project limit, got %v", err)
	}

	// But a different project should work
	_, err = pool.SpawnAgent(ctx, "Agent4", "dev", "project-2", "gpt-4", []string{"test-skill"})
	if err != nil {
		t.Errorf("Spawn for different project should succeed, got %v", err)
	}
}

func TestAgentPoolGetAgent(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	skills.RegisterSkill(&Skill{Name: "test-skill"})

	pool := NewAgentPool(DefaultPoolConfig(), skills, &mockLLMClient{})
	defer pool.Shutdown()

	ctx := context.Background()
	agent, _ := pool.SpawnAgent(ctx, "Alice", "dev", "project-1", "gpt-4", []string{"test-skill"})

	// Get existing agent
	retrieved, err := pool.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("GetAgent failed: %v", err)
	}

	if retrieved.ID != agent.ID {
		t.Error("Retrieved agent ID mismatch")
	}

	// Get non-existent agent
	_, err = pool.GetAgent("non-existent-id")
	if err != ErrAgentNotFound {
		t.Errorf("Expected ErrAgentNotFound, got %v", err)
	}
}

func TestAgentPoolGetProjectAgents(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	skills.RegisterSkill(&Skill{Name: "test-skill"})

	pool := NewAgentPool(DefaultPoolConfig(), skills, &mockLLMClient{})
	defer pool.Shutdown()

	ctx := context.Background()

	// Spawn agents for different projects
	pool.SpawnAgent(ctx, "Agent1", "dev", "project-1", "gpt-4", []string{"test-skill"})
	pool.SpawnAgent(ctx, "Agent2", "dev", "project-1", "gpt-4", []string{"test-skill"})
	pool.SpawnAgent(ctx, "Agent3", "dev", "project-2", "gpt-4", []string{"test-skill"})

	// Get project-1 agents
	agents := pool.GetProjectAgents("project-1")
	if len(agents) != 2 {
		t.Errorf("Expected 2 agents for project-1, got %d", len(agents))
	}

	// Get project-2 agents
	agents = pool.GetProjectAgents("project-2")
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent for project-2, got %d", len(agents))
	}

	// Get non-existent project
	agents = pool.GetProjectAgents("project-3")
	if len(agents) != 0 {
		t.Errorf("Expected 0 agents for project-3, got %d", len(agents))
	}
}

func TestAgentPoolTerminateAgent(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	skills.RegisterSkill(&Skill{Name: "test-skill"})

	pool := NewAgentPool(DefaultPoolConfig(), skills, &mockLLMClient{})
	defer pool.Shutdown()

	ctx := context.Background()
	agent, _ := pool.SpawnAgent(ctx, "Alice", "dev", "project-1", "gpt-4", []string{"test-skill"})

	err := pool.TerminateAgent(agent.ID)
	if err != nil {
		t.Fatalf("TerminateAgent failed: %v", err)
	}

	// Agent should be removed from pool
	total, _, _, _ := pool.GetStats()
	if total != 0 {
		t.Errorf("Expected 0 agents after termination, got %d", total)
	}

	// Getting terminated agent should fail
	_, err = pool.GetAgent(agent.ID)
	if err != ErrAgentNotFound {
		t.Errorf("Expected ErrAgentNotFound after termination, got %v", err)
	}
}

func TestAgentPoolTerminateAgentNotFound(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	pool := NewAgentPool(DefaultPoolConfig(), skills, &mockLLMClient{})
	defer pool.Shutdown()

	err := pool.TerminateAgent("non-existent-id")
	if err != ErrAgentNotFound {
		t.Errorf("Expected ErrAgentNotFound, got %v", err)
	}
}

func TestAgentPoolTerminateProjectAgents(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	skills.RegisterSkill(&Skill{Name: "test-skill"})

	pool := NewAgentPool(DefaultPoolConfig(), skills, &mockLLMClient{})
	defer pool.Shutdown()

	ctx := context.Background()

	// Spawn agents for multiple projects
	pool.SpawnAgent(ctx, "Agent1", "dev", "project-1", "gpt-4", []string{"test-skill"})
	pool.SpawnAgent(ctx, "Agent2", "dev", "project-1", "gpt-4", []string{"test-skill"})
	pool.SpawnAgent(ctx, "Agent3", "dev", "project-2", "gpt-4", []string{"test-skill"})

	// Terminate all project-1 agents
	pool.TerminateProjectAgents("project-1")

	// Should only have project-2 agent left
	total, _, _, _ := pool.GetStats()
	if total != 1 {
		t.Errorf("Expected 1 agent remaining, got %d", total)
	}

	agents := pool.GetProjectAgents("project-1")
	if len(agents) != 0 {
		t.Errorf("Expected 0 agents for project-1, got %d", len(agents))
	}

	agents = pool.GetProjectAgents("project-2")
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent for project-2, got %d", len(agents))
	}
}

func TestAgentPoolGetStats(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	skills.RegisterSkill(&Skill{Name: "test-skill"})

	pool := NewAgentPool(DefaultPoolConfig(), skills, &mockLLMClient{})
	defer pool.Shutdown()

	ctx := context.Background()

	// Spawn agents
	agent1, _ := pool.SpawnAgent(ctx, "Agent1", "dev", "project-1", "gpt-4", []string{"test-skill"})
	agent2, _ := pool.SpawnAgent(ctx, "Agent2", "dev", "project-1", "gpt-4", []string{"test-skill"})

	// One agent working
	agent1.AssignTask(&Task{ID: "task-1", Description: "Work"})

	total, active, idle, terminated := pool.GetStats()

	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}

	if active != 1 {
		t.Errorf("Expected active 1, got %d", active)
	}

	if idle != 1 {
		t.Errorf("Expected idle 1, got %d", idle)
	}

	if terminated != 0 {
		t.Errorf("Expected terminated 0, got %d", terminated)
	}

	// Complete work and terminate one
	agent1.CompleteTask(nil)
	agent2.Terminate()

	// Note: terminated agents are still in pool until cleanup
	total, active, idle, terminated = pool.GetStats()
	if active != 0 {
		t.Errorf("Expected active 0, got %d", active)
	}
}

// Note: Message routing tests are skipped because the agent's internal
// processMessages goroutine consumes messages from Inbox, making it
// impossible to test externally. Message routing is tested at integration level.

func TestAgentPoolShutdown(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	skills.RegisterSkill(&Skill{Name: "test-skill"})

	pool := NewAgentPool(DefaultPoolConfig(), skills, &mockLLMClient{})

	ctx := context.Background()
	pool.SpawnAgent(ctx, "Agent1", "dev", "project-1", "gpt-4", []string{"test-skill"})
	pool.SpawnAgent(ctx, "Agent2", "dev", "project-1", "gpt-4", []string{"test-skill"})

	pool.Shutdown()

	total, _, _, _ := pool.GetStats()
	if total != 0 {
		t.Errorf("Expected 0 agents after shutdown, got %d", total)
	}
}

func TestAgentPoolConcurrentSpawn(t *testing.T) {
	skills := NewSkillRegistry("/tmp/skills")
	skills.RegisterSkill(&Skill{Name: "test-skill"})

	pool := NewAgentPool(DefaultPoolConfig(), skills, &mockLLMClient{})
	defer pool.Shutdown()

	ctx := context.Background()

	// Spawn multiple agents concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			_, err := pool.SpawnAgent(ctx, "Agent", "dev", "project-1", "gpt-4", []string{"test-skill"})
			if err != nil && err != ErrPoolFull {
				t.Errorf("Unexpected error: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all spawns
	for i := 0; i < 10; i++ {
		<-done
	}

	total, _, _, _ := pool.GetStats()
	if total != 10 {
		t.Errorf("Expected 10 agents, got %d", total)
	}
}
