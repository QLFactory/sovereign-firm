// Package activities provides Temporal activities for the Sovereign Firm.
package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/qlfactory/sovereign-firm/pkg/agent"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// SkillInjector handles skill-related activities
type SkillInjector struct {
	registry  *agent.SkillRegistry
	llmClient llm.Client
}

// NewSkillInjector creates a new skill injector
func NewSkillInjector(skillsDir string) *SkillInjector {
	registry := agent.NewSkillRegistry(skillsDir)
	if err := registry.LoadSkills(); err != nil {
		// Log warning but continue - skills can be loaded later
		fmt.Printf("Warning: failed to load skills: %v\n", err)
	}
	return &SkillInjector{
		registry:  registry,
		llmClient: llm.NewClient(),
	}
}

// NewSkillInjectorWithRegistry creates a skill injector with an existing registry
func NewSkillInjectorWithRegistry(registry *agent.SkillRegistry) *SkillInjector {
	return &SkillInjector{
		registry:  registry,
		llmClient: llm.NewClient(),
	}
}

// SkillSelectionInput contains input for skill selection
type SkillSelectionInput struct {
	TechStack          *agent.TechStack      `json:"tech_stack,omitempty"`
	Categories         []agent.SkillCategory `json:"categories,omitempty"`
	ExplicitSkillNames []string              `json:"explicit_skill_names,omitempty"`
	MaxSkills          int                   `json:"max_skills,omitempty"`
}

// SkillSelectionResult contains the selected skills
type SkillSelectionResult struct {
	SelectedSkills []SelectedSkillInfo `json:"selected_skills"`
	SystemPrompt   string              `json:"system_prompt"`
	TotalScore     int                 `json:"total_score"`
}

// SelectedSkillInfo contains information about a selected skill
type SelectedSkillInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Category    string `json:"category"`
	Score       int    `json:"score"`
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

// SelectSkills selects appropriate skills based on requirements
func (s *SkillInjector) SelectSkills(ctx context.Context, input SkillSelectionInput) (*SkillSelectionResult, error) {
	requirements := &agent.SkillRequirements{
		TechStack:  input.TechStack,
		Categories: input.Categories,
		SkillNames: input.ExplicitSkillNames,
	}

	matches := s.registry.SelectSkillsForRequirements(requirements)

	// Apply max skills limit
	maxSkills := input.MaxSkills
	if maxSkills <= 0 {
		maxSkills = 10 // Default
	}
	if len(matches) > maxSkills {
		matches = matches[:maxSkills]
	}

	result := &SkillSelectionResult{
		SelectedSkills: make([]SelectedSkillInfo, len(matches)),
	}

	var skillNames []string
	for i, match := range matches {
		result.SelectedSkills[i] = SelectedSkillInfo{
			Name:        match.Skill.Name,
			Version:     match.Skill.Version,
			Category:    string(match.Definition.Category),
			Score:       match.Score,
			Reason:      match.Reason,
			Description: match.Skill.Description,
		}
		result.TotalScore += match.Score
		skillNames = append(skillNames, match.Skill.Name)
	}

	// Generate combined system prompt
	result.SystemPrompt = s.registry.ExportSkillsToPrompt(skillNames)

	return result, nil
}

// AgentSpawnInput contains input for spawning a skilled agent
type AgentSpawnInput struct {
	AgentName   string                `json:"agent_name"`
	AgentRole   string                `json:"agent_role"`
	ProjectID   string                `json:"project_id"`
	Model       string                `json:"model,omitempty"`
	TechStack   *agent.TechStack      `json:"tech_stack,omitempty"`
	Categories  []agent.SkillCategory `json:"categories,omitempty"`
	SkillNames  []string              `json:"skill_names,omitempty"`
}

// AgentSpawnResult contains the spawned agent information
type AgentSpawnResult struct {
	AgentID        string              `json:"agent_id"`
	AgentName      string              `json:"agent_name"`
	AgentRole      string              `json:"agent_role"`
	InjectedSkills []SelectedSkillInfo `json:"injected_skills"`
	SystemPrompt   string              `json:"system_prompt"`
}

// SpawnSkilledAgent creates an agent with appropriate skills injected
func (s *SkillInjector) SpawnSkilledAgent(ctx context.Context, input AgentSpawnInput) (*AgentSpawnResult, error) {
	requirements := &agent.SkillRequirements{
		TechStack:  input.TechStack,
		Categories: input.Categories,
		SkillNames: input.SkillNames,
	}

	model := input.Model
	if model == "" {
		model = "default"
	}

	agentInstance, matches := s.registry.GetSkilledAgent(
		ctx,
		input.AgentName,
		input.AgentRole,
		input.ProjectID,
		model,
		requirements,
		s.llmClient,
	)

	// Spawn the agent
	if err := agentInstance.Spawn(ctx); err != nil {
		return nil, fmt.Errorf("failed to spawn agent: %w", err)
	}

	result := &AgentSpawnResult{
		AgentID:        agentInstance.ID,
		AgentName:      agentInstance.Name,
		AgentRole:      agentInstance.Role,
		InjectedSkills: make([]SelectedSkillInfo, len(matches)),
		SystemPrompt:   agentInstance.SystemPrompt,
	}

	for i, match := range matches {
		result.InjectedSkills[i] = SelectedSkillInfo{
			Name:        match.Skill.Name,
			Version:     match.Skill.Version,
			Category:    string(match.Definition.Category),
			Score:       match.Score,
			Reason:      match.Reason,
			Description: match.Skill.Description,
		}
	}

	return result, nil
}

// ComposeSkillsInput contains input for composing skills
type ComposeSkillsInput struct {
	SkillNames []string `json:"skill_names"`
}

// ComposeSkillsResult contains the composed skill
type ComposeSkillsResult struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
	Tools        []string `json:"tools"`
	SystemPrompt string   `json:"system_prompt"`
}

// ComposeSkills combines multiple skills into one
func (s *SkillInjector) ComposeSkills(ctx context.Context, input ComposeSkillsInput) (*ComposeSkillsResult, error) {
	composed, err := s.registry.ComposeSkills(input.SkillNames...)
	if err != nil {
		return nil, fmt.Errorf("failed to compose skills: %w", err)
	}

	return &ComposeSkillsResult{
		Name:         composed.Name,
		Version:      composed.Version,
		Description:  composed.Description,
		Capabilities: composed.Capabilities,
		Tools:        composed.Tools,
		SystemPrompt: composed.SystemPrompt,
	}, nil
}

// AnalyzeProjectInput contains input for analyzing project requirements
type AnalyzeProjectInput struct {
	ProjectSpec string `json:"project_spec"`
}

// AnalyzeProjectResult contains recommended skills for a project
type AnalyzeProjectResult struct {
	RecommendedTechStack *agent.TechStack      `json:"recommended_tech_stack"`
	RecommendedSkills    []SelectedSkillInfo   `json:"recommended_skills"`
	Categories           []agent.SkillCategory `json:"categories"`
	Reasoning            string                `json:"reasoning"`
}

// AnalyzeProjectForSkills uses LLM to analyze project and recommend skills
func (s *SkillInjector) AnalyzeProjectForSkills(ctx context.Context, input AnalyzeProjectInput) (*AnalyzeProjectResult, error) {
	// Get available skills for context
	availableSkills := s.registry.ListSkills()
	availableCategories := s.registry.GetAllCategories()
	availableTags := s.registry.GetAllTags()

	prompt := fmt.Sprintf(`Analyze this project specification and recommend the appropriate technology stack and skills.

PROJECT SPECIFICATION:
%s

AVAILABLE SKILLS:
%v

AVAILABLE CATEGORIES:
%v

AVAILABLE TAGS:
%v

Based on the project requirements, recommend:
1. A technology stack (frontend, backend, database, cloud, etc.)
2. Which skills from the available list would be most helpful
3. Which categories of agents are needed
4. Your reasoning for these choices

Respond in JSON format with this structure:
{
  "recommended_tech_stack": {
    "frontend": "react",
    "backend": "nodejs",
    "database": "postgresql",
    "cloud": "aws",
    "containerization": "docker",
    "ci": "github-actions",
    "monitoring": "prometheus",
    "additional": []
  },
  "recommended_skill_names": ["react-developer", "express-developer", "postgresql-specialist"],
  "categories": ["frontend", "backend", "database", "devops"],
  "reasoning": "Explanation of why these choices were made..."
}`, input.ProjectSpec, availableSkills, availableCategories, availableTags)

	resp, err := s.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: "You are a technical consultant who analyzes project requirements and recommends appropriate technology stacks and development skills. Respond only with valid JSON.",
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Parse LLM response
	var llmResult struct {
		RecommendedTechStack   *agent.TechStack `json:"recommended_tech_stack"`
		RecommendedSkillNames  []string         `json:"recommended_skill_names"`
		Categories             []string         `json:"categories"`
		Reasoning              string           `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(resp.Response), &llmResult); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Convert string categories to SkillCategory
	var categories []agent.SkillCategory
	for _, cat := range llmResult.Categories {
		categories = append(categories, agent.SkillCategory(cat))
	}

	// Select skills based on recommendations
	requirements := &agent.SkillRequirements{
		TechStack:  llmResult.RecommendedTechStack,
		Categories: categories,
		SkillNames: llmResult.RecommendedSkillNames,
	}

	matches := s.registry.SelectSkillsForRequirements(requirements)

	result := &AnalyzeProjectResult{
		RecommendedTechStack: llmResult.RecommendedTechStack,
		RecommendedSkills:    make([]SelectedSkillInfo, len(matches)),
		Categories:           categories,
		Reasoning:            llmResult.Reasoning,
	}

	for i, match := range matches {
		result.RecommendedSkills[i] = SelectedSkillInfo{
			Name:        match.Skill.Name,
			Version:     match.Skill.Version,
			Category:    string(match.Definition.Category),
			Score:       match.Score,
			Reason:      match.Reason,
			Description: match.Skill.Description,
		}
	}

	return result, nil
}

// ListSkillsInput contains input for listing skills
type ListSkillsInput struct {
	Category agent.SkillCategory `json:"category,omitempty"`
	Tag      string              `json:"tag,omitempty"`
}

// ListSkillsResult contains the list of skills
type ListSkillsResult struct {
	Skills []SkillSummary `json:"skills"`
	Total  int            `json:"total"`
}

// SkillSummary contains summary information about a skill
type SkillSummary struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	Capabilities []string `json:"capabilities"`
}

// ListSkills lists available skills with optional filtering
func (s *SkillInjector) ListSkills(ctx context.Context, input ListSkillsInput) (*ListSkillsResult, error) {
	var skills []*agent.Skill

	if input.Category != "" {
		skills = s.registry.GetSkillsByCategory(input.Category)
	} else if input.Tag != "" {
		skills = s.registry.GetSkillsByTag(input.Tag)
	} else {
		// Get all skills
		names := s.registry.ListSkills()
		for _, name := range names {
			skill, err := s.registry.GetSkill(name)
			if err == nil {
				skills = append(skills, skill)
			}
		}
	}

	result := &ListSkillsResult{
		Skills: make([]SkillSummary, len(skills)),
		Total:  len(skills),
	}

	for i, skill := range skills {
		def, _ := s.registry.GetSkillDefinition(skill.Name)
		var tags []string
		var category string
		if def != nil {
			tags = def.Tags
			category = string(def.Category)
		}

		result.Skills[i] = SkillSummary{
			Name:         skill.Name,
			Version:      skill.Version,
			Description:  skill.Description,
			Category:     category,
			Tags:         tags,
			Capabilities: skill.Capabilities,
		}
	}

	return result, nil
}

// GetSkillDependenciesInput contains input for getting skill dependencies
type GetSkillDependenciesInput struct {
	SkillName string `json:"skill_name"`
}

// GetSkillDependenciesResult contains the skill and its dependencies
type GetSkillDependenciesResult struct {
	SkillName    string   `json:"skill_name"`
	Dependencies []string `json:"dependencies"`
}

// GetSkillDependencies returns all dependencies for a skill
func (s *SkillInjector) GetSkillDependencies(ctx context.Context, input GetSkillDependenciesInput) (*GetSkillDependenciesResult, error) {
	deps, err := s.registry.GetDependencies(input.SkillName)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}

	return &GetSkillDependenciesResult{
		SkillName:    input.SkillName,
		Dependencies: deps,
	}, nil
}

// RegisterSkillInput contains input for registering a new skill
type RegisterSkillInput struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
	Tools        []string `json:"tools"`
	SystemPrompt string   `json:"system_prompt"`
}

// RegisterSkillResult contains the result of skill registration
type RegisterSkillResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RegisterSkill registers a new skill at runtime
func (s *SkillInjector) RegisterSkill(ctx context.Context, input RegisterSkillInput) (*RegisterSkillResult, error) {
	skill := &agent.Skill{
		Name:         input.Name,
		Version:      input.Version,
		Description:  input.Description,
		Capabilities: input.Capabilities,
		Tools:        input.Tools,
		SystemPrompt: input.SystemPrompt,
	}

	s.registry.RegisterSkill(skill)

	return &RegisterSkillResult{
		Success: true,
		Message: fmt.Sprintf("Skill '%s' registered successfully", input.Name),
	}, nil
}

// GetRegistry returns the underlying skill registry
func (s *SkillInjector) GetRegistry() *agent.SkillRegistry {
	return s.registry
}

// RefreshSkills reloads skills from disk
func (s *SkillInjector) RefreshSkills(ctx context.Context) error {
	return s.registry.LoadSkills()
}

// GetSchema functions for activity inputs

// GetSkillSelectionInputSchema returns the JSON schema for SkillSelectionInput
func GetSkillSelectionInputSchema() string {
	return `{
		"type": "object",
		"properties": {
			"tech_stack": {
				"type": "object",
				"properties": {
					"frontend": { "type": "string" },
					"backend": { "type": "string" },
					"database": { "type": "string" },
					"cloud": { "type": "string" },
					"containerization": { "type": "string" },
					"ci": { "type": "string" },
					"monitoring": { "type": "string" },
					"additional": { "type": "array", "items": { "type": "string" } }
				}
			},
			"categories": {
				"type": "array",
				"items": { "type": "string", "enum": ["frontend", "backend", "database", "devops", "infrastructure", "qa", "sre", "security", "architect", "pm"] }
			},
			"explicit_skill_names": {
				"type": "array",
				"items": { "type": "string" }
			},
			"max_skills": { "type": "integer", "minimum": 1 }
		}
	}`
}

// GetAgentSpawnInputSchema returns the JSON schema for AgentSpawnInput
func GetAgentSpawnInputSchema() string {
	return `{
		"type": "object",
		"required": ["agent_name", "agent_role", "project_id"],
		"properties": {
			"agent_name": { "type": "string" },
			"agent_role": { "type": "string" },
			"project_id": { "type": "string" },
			"model": { "type": "string" },
			"tech_stack": { "type": "object" },
			"categories": { "type": "array", "items": { "type": "string" } },
			"skill_names": { "type": "array", "items": { "type": "string" } }
		}
	}`
}
