package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// PoolConfig configures the agent pool
type PoolConfig struct {
	MaxAgents           int           `json:"max_agents"`
	MaxAgentsPerProject int           `json:"max_agents_per_project"`
	IdleTimeout         time.Duration `json:"idle_timeout"`
	CleanupInterval     time.Duration `json:"cleanup_interval"`
}

// DefaultPoolConfig returns sensible defaults
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxAgents:           100,
		MaxAgentsPerProject: 10,
		IdleTimeout:         15 * time.Minute,
		CleanupInterval:     1 * time.Minute,
	}
}

// AgentPool manages all active agents in the system
type AgentPool struct {
	config    PoolConfig
	agents    map[string]*AgentInstance            // ID -> Agent
	byProject map[string]map[string]*AgentInstance // ProjectID -> ID -> Agent
	skills    *SkillRegistry
	llmClient llm.Client

	// Message routing
	messageBus chan Message

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewAgentPool creates a new agent pool
func NewAgentPool(config PoolConfig, skills *SkillRegistry, llmClient llm.Client) *AgentPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &AgentPool{
		config:     config,
		agents:     make(map[string]*AgentInstance),
		byProject:  make(map[string]map[string]*AgentInstance),
		skills:     skills,
		llmClient:  llmClient,
		messageBus: make(chan Message, 1000),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start background workers
	go pool.messageRouter()
	go pool.cleanup()

	return pool
}

// SpawnAgent creates a new agent with the given role and skills
func (p *AgentPool) SpawnAgent(
	ctx context.Context,
	name, role, projectID, model string,
	skillNames []string,
) (*AgentInstance, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check pool limits
	if len(p.agents) >= p.config.MaxAgents {
		return nil, ErrPoolFull
	}

	// Check per-project limits
	if projectAgents, ok := p.byProject[projectID]; ok {
		if len(projectAgents) >= p.config.MaxAgentsPerProject {
			return nil, ErrPoolFull
		}
	}

	// Load skills
	skills := make([]Skill, 0, len(skillNames))
	for _, skillName := range skillNames {
		skill, err := p.skills.GetSkill(skillName)
		if err != nil {
			return nil, err
		}
		skills = append(skills, *skill)
	}

	// Create agent with pool's context so it's cancelled on pool shutdown
	agent := NewAgentInstance(p.ctx, name, role, projectID, model, skills, p.llmClient)

	// Spawn the agent
	if err := agent.Spawn(ctx); err != nil {
		return nil, err
	}

	// Register in pool
	p.agents[agent.ID] = agent
	if _, ok := p.byProject[projectID]; !ok {
		p.byProject[projectID] = make(map[string]*AgentInstance)
	}
	p.byProject[projectID][agent.ID] = agent

	// Wire up outbox to message bus with proper cancellation
	go func() {
		for {
			select {
			case <-p.ctx.Done():
				// Pool is shutting down
				return
			case msg, ok := <-agent.Outbox:
				if !ok {
					// Channel closed (agent terminated)
					return
				}
				select {
				case p.messageBus <- msg:
				case <-p.ctx.Done():
					return
				}
			}
		}
	}()

	log.Printf("Agent spawned: %s (%s) for project %s", agent.Name, agent.ID, projectID)

	return agent, nil
}

// GetAgent retrieves an agent by ID
func (p *AgentPool) GetAgent(id string) (*AgentInstance, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agent, ok := p.agents[id]
	if !ok {
		return nil, ErrAgentNotFound
	}
	return agent, nil
}

// GetProjectAgents returns all agents for a project
func (p *AgentPool) GetProjectAgents(projectID string) []*AgentInstance {
	p.mu.RLock()
	defer p.mu.RUnlock()

	agents := make([]*AgentInstance, 0)
	if projectAgents, ok := p.byProject[projectID]; ok {
		for _, agent := range projectAgents {
			agents = append(agents, agent)
		}
	}
	return agents
}

// TerminateAgent terminates a specific agent
func (p *AgentPool) TerminateAgent(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	agent, ok := p.agents[id]
	if !ok {
		return ErrAgentNotFound
	}

	agent.Terminate()

	// Remove from pool
	delete(p.agents, id)
	if projectAgents, ok := p.byProject[agent.ProjectID]; ok {
		delete(projectAgents, id)
		if len(projectAgents) == 0 {
			delete(p.byProject, agent.ProjectID)
		}
	}

	log.Printf("Agent terminated: %s (%s)", agent.Name, id)

	return nil
}

// TerminateProjectAgents terminates all agents for a project
func (p *AgentPool) TerminateProjectAgents(projectID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if projectAgents, ok := p.byProject[projectID]; ok {
		for id, agent := range projectAgents {
			agent.Terminate()
			delete(p.agents, id)
		}
		delete(p.byProject, projectID)
	}
}

// messageRouter routes messages between agents
func (p *AgentPool) messageRouter() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case msg := <-p.messageBus:
			p.routeMessage(msg)
		}
	}
}

// routeMessage sends a message to the appropriate agent(s)
func (p *AgentPool) routeMessage(msg Message) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if msg.To == "*" {
		// Broadcast to all agents
		for _, agent := range p.agents {
			if agent.ID != msg.From {
				p.deliverToAgent(agent, msg)
			}
		}
	} else {
		// Send to specific agent
		if agent, ok := p.agents[msg.To]; ok {
			p.deliverToAgent(agent, msg)
		} else {
			log.Printf("WARNING: Message to unknown agent %s dropped (from: %s, subject: %s)",
				msg.To, msg.From, msg.Subject)
		}
	}
}

// deliverToAgent attempts to deliver a message to an agent's inbox with timeout
func (p *AgentPool) deliverToAgent(agent *AgentInstance, msg Message) {
	select {
	case agent.Inbox <- msg:
		// Delivered successfully
	case <-time.After(100 * time.Millisecond):
		log.Printf("WARNING: Agent %s (%s) inbox full, message from %s dropped (subject: %s)",
			agent.Name, agent.ID, msg.From, msg.Subject)
	case <-p.ctx.Done():
		// Pool is shutting down
	}
}

// Satisfy AgentMessaging interface

// GetAgents returns a list of available agents for a project
func (p *AgentPool) GetAgents(projectID string) []map[string]string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	results := make([]map[string]string, 0)
	if projectAgents, ok := p.byProject[projectID]; ok {
		for _, agent := range projectAgents {
			results = append(results, map[string]string{
				"id":   agent.ID,
				"name": agent.Name,
				"role": agent.Role,
			})
		}
	}
	return results
}

// SendMessage sends a message between agents
func (p *AgentPool) SendMessage(from, to, subject string, content interface{}) error {
	msg := Message{
		ID:        uuid.New().String(), // Need to import uuid in pool.go if not there
		Type:      MessageQuestion,     // Default for tool-initiated messaging
		From:      from,
		To:        to,
		Subject:   subject,
		Content:   content,
		Timestamp: time.Now(),
	}

	p.messageBus <- msg
	return nil
}

// cleanup periodically removes terminated and idle agents
func (p *AgentPool) cleanup() {
	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.doCleanup()
		}
	}
}

// doCleanup removes terminated agents
func (p *AgentPool) doCleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	for id, agent := range p.agents {
		// Remove terminated agents
		if agent.Status == StatusTerminated {
			toRemove = append(toRemove, id)
			continue
		}

		// Remove idle agents past timeout
		if agent.Status == StatusReady && agent.CurrentTask == nil {
			idleTime := now.Sub(agent.CreatedAt)
			if agent.TasksComplete > 0 {
				// Use last task completion time if available
				// For now, just use creation time
			}
			if idleTime > p.config.IdleTimeout {
				agent.Terminate()
				toRemove = append(toRemove, id)
			}
		}
	}

	for _, id := range toRemove {
		if agent, ok := p.agents[id]; ok {
			if projectAgents, ok := p.byProject[agent.ProjectID]; ok {
				delete(projectAgents, id)
				if len(projectAgents) == 0 {
					delete(p.byProject, agent.ProjectID)
				}
			}
			delete(p.agents, id)
		}
	}

	if len(toRemove) > 0 {
		log.Printf("Cleaned up %d agents", len(toRemove))
	}
}

// GetStats returns pool statistics
func (p *AgentPool) GetStats() (total, active, idle, terminated int) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total = len(p.agents)
	for _, agent := range p.agents {
		switch agent.Status {
		case StatusWorking, StatusReviewing:
			active++
		case StatusReady, StatusBlocked:
			idle++
		case StatusTerminated:
			terminated++
		}
	}
	return
}

// Shutdown terminates all agents and stops the pool
func (p *AgentPool) Shutdown() {
	p.cancel()

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, agent := range p.agents {
		agent.Terminate()
	}

	p.agents = make(map[string]*AgentInstance)
	p.byProject = make(map[string]map[string]*AgentInstance)

	close(p.messageBus)

	log.Println("Agent pool shutdown complete")
}
