package activities

import (
	"context"
	"testing"

	"github.com/qlfactory/sovereign-firm/pkg/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillInjector_NewSkillInjector(t *testing.T) {
	// Test with non-existent directory (should not panic)
	injector := NewSkillInjector("/non/existent/path")
	assert.NotNil(t, injector)
	assert.NotNil(t, injector.GetRegistry())
}

func TestSkillInjector_WithRegistry(t *testing.T) {
	registry := agent.NewSkillRegistry("/tmp/skills")

	// Register a test skill
	skill := &agent.Skill{
		Name:         "test-skill",
		Version:      "1.0.0",
		Description:  "A test skill",
		Capabilities: []string{"testing"},
		Tools:        []string{"test_tool"},
		SystemPrompt: "You are a tester.",
	}
	registry.RegisterSkill(skill)

	injector := NewSkillInjectorWithRegistry(registry)
	assert.NotNil(t, injector)
	assert.Equal(t, registry, injector.GetRegistry())

	// Verify the skill is accessible
	result, err := injector.ListSkills(context.Background(), ListSkillsInput{})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, "test-skill", result.Skills[0].Name)
}

func TestSkillSelectionInput(t *testing.T) {
	input := SkillSelectionInput{
		TechStack: &agent.TechStack{
			Frontend: "react",
			Backend:  "nodejs",
			Database: "postgresql",
		},
		Categories:         []agent.SkillCategory{agent.CategoryFrontend, agent.CategoryBackend},
		ExplicitSkillNames: []string{"react-developer"},
		MaxSkills:          5,
	}

	assert.Equal(t, "react", input.TechStack.Frontend)
	assert.Equal(t, "nodejs", input.TechStack.Backend)
	assert.Equal(t, 2, len(input.Categories))
	assert.Equal(t, 5, input.MaxSkills)
}

func TestSelectedSkillInfo(t *testing.T) {
	info := SelectedSkillInfo{
		Name:        "react-developer",
		Version:     "1.0.0",
		Category:    "frontend",
		Score:       85,
		Reason:      "Matched tags: [react, frontend]",
		Description: "Build React components",
	}

	assert.Equal(t, "react-developer", info.Name)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "frontend", info.Category)
	assert.Equal(t, 85, info.Score)
	assert.Contains(t, info.Reason, "react")
}

func TestSkillSelectionResult(t *testing.T) {
	result := SkillSelectionResult{
		SelectedSkills: []SelectedSkillInfo{
			{Name: "skill1", Score: 50},
			{Name: "skill2", Score: 30},
		},
		SystemPrompt: "Combined prompt",
		TotalScore:   80,
	}

	assert.Equal(t, 2, len(result.SelectedSkills))
	assert.Equal(t, 80, result.TotalScore)
	assert.NotEmpty(t, result.SystemPrompt)
}

func TestAgentSpawnInput(t *testing.T) {
	input := AgentSpawnInput{
		AgentName:  "Alice",
		AgentRole:  "senior-frontend-dev",
		ProjectID:  "project-123",
		Model:      "gpt-4",
		TechStack:  &agent.TechStack{Frontend: "react"},
		Categories: []agent.SkillCategory{agent.CategoryFrontend},
		SkillNames: []string{"react-developer"},
	}

	assert.Equal(t, "Alice", input.AgentName)
	assert.Equal(t, "senior-frontend-dev", input.AgentRole)
	assert.Equal(t, "project-123", input.ProjectID)
	assert.Equal(t, "gpt-4", input.Model)
}

func TestAgentSpawnResult(t *testing.T) {
	result := AgentSpawnResult{
		AgentID:   "agent-uuid-123",
		AgentName: "Alice",
		AgentRole: "frontend-dev",
		InjectedSkills: []SelectedSkillInfo{
			{Name: "react-developer", Score: 100},
		},
		SystemPrompt: "You are Alice...",
	}

	assert.NotEmpty(t, result.AgentID)
	assert.Equal(t, "Alice", result.AgentName)
	assert.Equal(t, 1, len(result.InjectedSkills))
}

func TestComposeSkillsInput(t *testing.T) {
	input := ComposeSkillsInput{
		SkillNames: []string{"react-developer", "typescript-expert"},
	}

	assert.Equal(t, 2, len(input.SkillNames))
	assert.Contains(t, input.SkillNames, "react-developer")
}

func TestComposeSkillsResult(t *testing.T) {
	result := ComposeSkillsResult{
		Name:         "react-developer+typescript-expert",
		Version:      "1.0.0",
		Description:  "Composed skill from: [react-developer typescript-expert]",
		Capabilities: []string{"react", "typescript"},
		Tools:        []string{"file_write", "npm_run"},
		SystemPrompt: "Combined prompt...",
	}

	assert.Contains(t, result.Name, "+")
	assert.Equal(t, 2, len(result.Capabilities))
}

func TestAnalyzeProjectInput(t *testing.T) {
	input := AnalyzeProjectInput{
		ProjectSpec: "Build a todo app with React frontend and Node.js backend",
	}

	assert.Contains(t, input.ProjectSpec, "todo")
	assert.Contains(t, input.ProjectSpec, "React")
}

func TestAnalyzeProjectResult(t *testing.T) {
	result := AnalyzeProjectResult{
		RecommendedTechStack: &agent.TechStack{
			Frontend: "react",
			Backend:  "nodejs",
			Database: "postgresql",
		},
		RecommendedSkills: []SelectedSkillInfo{
			{Name: "react-developer", Score: 100},
			{Name: "express-developer", Score: 90},
		},
		Categories: []agent.SkillCategory{
			agent.CategoryFrontend,
			agent.CategoryBackend,
		},
		Reasoning: "React is suitable for modern web apps...",
	}

	assert.Equal(t, "react", result.RecommendedTechStack.Frontend)
	assert.Equal(t, 2, len(result.RecommendedSkills))
	assert.Equal(t, 2, len(result.Categories))
	assert.NotEmpty(t, result.Reasoning)
}

func TestListSkillsInput(t *testing.T) {
	// Test with category
	input1 := ListSkillsInput{
		Category: agent.CategoryFrontend,
	}
	assert.Equal(t, agent.CategoryFrontend, input1.Category)

	// Test with tag
	input2 := ListSkillsInput{
		Tag: "react",
	}
	assert.Equal(t, "react", input2.Tag)
}

func TestListSkillsResult(t *testing.T) {
	result := ListSkillsResult{
		Skills: []SkillSummary{
			{Name: "skill1", Version: "1.0.0", Category: "frontend"},
			{Name: "skill2", Version: "2.0.0", Category: "backend"},
		},
		Total: 2,
	}

	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 2, len(result.Skills))
}

func TestSkillSummary(t *testing.T) {
	summary := SkillSummary{
		Name:         "react-developer",
		Version:      "1.0.0",
		Description:  "Build React components",
		Category:     "frontend",
		Tags:         []string{"react", "frontend", "typescript"},
		Capabilities: []string{"components", "hooks", "state"},
	}

	assert.Equal(t, "react-developer", summary.Name)
	assert.Equal(t, 3, len(summary.Tags))
	assert.Equal(t, 3, len(summary.Capabilities))
}

func TestGetSkillDependenciesInput(t *testing.T) {
	input := GetSkillDependenciesInput{
		SkillName: "kubernetes-specialist",
	}

	assert.Equal(t, "kubernetes-specialist", input.SkillName)
}

func TestGetSkillDependenciesResult(t *testing.T) {
	result := GetSkillDependenciesResult{
		SkillName:    "kubernetes-specialist",
		Dependencies: []string{"docker-devops"},
	}

	assert.Equal(t, "kubernetes-specialist", result.SkillName)
	assert.Equal(t, 1, len(result.Dependencies))
	assert.Contains(t, result.Dependencies, "docker-devops")
}

func TestRegisterSkillInput(t *testing.T) {
	input := RegisterSkillInput{
		Name:         "custom-skill",
		Version:      "1.0.0",
		Description:  "A custom skill",
		Capabilities: []string{"custom-cap"},
		Tools:        []string{"custom-tool"},
		SystemPrompt: "You are a custom expert.",
	}

	assert.Equal(t, "custom-skill", input.Name)
	assert.Equal(t, "1.0.0", input.Version)
	assert.Equal(t, 1, len(input.Capabilities))
}

func TestRegisterSkillResult(t *testing.T) {
	result := RegisterSkillResult{
		Success: true,
		Message: "Skill 'custom-skill' registered successfully",
	}

	assert.True(t, result.Success)
	assert.Contains(t, result.Message, "custom-skill")
}

func TestSkillInjector_SelectSkills(t *testing.T) {
	registry := createTestRegistry()
	injector := NewSkillInjectorWithRegistry(registry)

	input := SkillSelectionInput{
		TechStack: &agent.TechStack{
			Frontend: "react",
		},
		MaxSkills: 5,
	}

	result, err := injector.SelectSkills(context.Background(), input)
	require.NoError(t, err)
	assert.NotNil(t, result)
	// The react-test skill should match
	if len(result.SelectedSkills) > 0 {
		assert.True(t, result.TotalScore > 0)
	}
}

func TestSkillInjector_SelectSkillsWithCategories(t *testing.T) {
	registry := createTestRegistry()
	injector := NewSkillInjectorWithRegistry(registry)

	input := SkillSelectionInput{
		Categories: []agent.SkillCategory{agent.CategoryFrontend},
		MaxSkills:  10,
	}

	result, err := injector.SelectSkills(context.Background(), input)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSkillInjector_SelectSkillsWithExplicitNames(t *testing.T) {
	registry := createTestRegistry()
	injector := NewSkillInjectorWithRegistry(registry)

	input := SkillSelectionInput{
		ExplicitSkillNames: []string{"react-test"},
		MaxSkills:          5,
	}

	result, err := injector.SelectSkills(context.Background(), input)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Explicitly named skills should have highest priority
	if len(result.SelectedSkills) > 0 {
		assert.Equal(t, "react-test", result.SelectedSkills[0].Name)
		assert.Equal(t, 100, result.SelectedSkills[0].Score) // Explicit = 100 points
	}
}

func TestSkillInjector_ComposeSkills(t *testing.T) {
	registry := createTestRegistryWithMultipleSkills()
	injector := NewSkillInjectorWithRegistry(registry)

	input := ComposeSkillsInput{
		SkillNames: []string{"react-test", "node-test"},
	}

	result, err := injector.ComposeSkills(context.Background(), input)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Name, "+")
	assert.Contains(t, result.Name, "react-test")
	assert.Contains(t, result.Name, "node-test")
}

func TestSkillInjector_ComposeSkillsNotFound(t *testing.T) {
	registry := agent.NewSkillRegistry("/tmp/skills")
	injector := NewSkillInjectorWithRegistry(registry)

	input := ComposeSkillsInput{
		SkillNames: []string{"non-existent-skill"},
	}

	_, err := injector.ComposeSkills(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "skill not found")
}

func TestSkillInjector_ListSkills(t *testing.T) {
	registry := createTestRegistryWithMultipleSkills()
	injector := NewSkillInjectorWithRegistry(registry)

	result, err := injector.ListSkills(context.Background(), ListSkillsInput{})
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.Total)
}

func TestSkillInjector_ListSkillsByTag(t *testing.T) {
	registry := createTestRegistry()
	injector := NewSkillInjectorWithRegistry(registry)

	result, err := injector.ListSkills(context.Background(), ListSkillsInput{
		Tag: "react",
	})
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSkillInjector_RegisterSkill(t *testing.T) {
	registry := agent.NewSkillRegistry("/tmp/skills")
	injector := NewSkillInjectorWithRegistry(registry)

	input := RegisterSkillInput{
		Name:         "new-skill",
		Version:      "1.0.0",
		Description:  "A new skill",
		Capabilities: []string{"new-cap"},
		Tools:        []string{"new-tool"},
		SystemPrompt: "You are new.",
	}

	result, err := injector.RegisterSkill(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Verify skill was registered
	listResult, err := injector.ListSkills(context.Background(), ListSkillsInput{})
	require.NoError(t, err)
	assert.Equal(t, 1, listResult.Total)
	assert.Equal(t, "new-skill", listResult.Skills[0].Name)
}

func TestSkillInjector_GetSkillDependencies(t *testing.T) {
	registry := agent.NewSkillRegistry("/tmp/skills")
	injector := NewSkillInjectorWithRegistry(registry)

	// Register a skill without dependencies
	skill := &agent.Skill{
		Name:    "base-skill",
		Version: "1.0.0",
	}
	registry.RegisterSkill(skill)

	result, err := injector.GetSkillDependencies(context.Background(), GetSkillDependenciesInput{
		SkillName: "base-skill",
	})
	require.NoError(t, err)
	assert.Equal(t, "base-skill", result.SkillName)
	assert.Empty(t, result.Dependencies)
}

func TestGetSkillSelectionInputSchema(t *testing.T) {
	schema := GetSkillSelectionInputSchema()
	assert.Contains(t, schema, "tech_stack")
	assert.Contains(t, schema, "categories")
	assert.Contains(t, schema, "explicit_skill_names")
	assert.Contains(t, schema, "max_skills")
}

func TestGetAgentSpawnInputSchema(t *testing.T) {
	schema := GetAgentSpawnInputSchema()
	assert.Contains(t, schema, "agent_name")
	assert.Contains(t, schema, "agent_role")
	assert.Contains(t, schema, "project_id")
	assert.Contains(t, schema, "required")
}

func TestTechStack(t *testing.T) {
	stack := agent.TechStack{
		Frontend:         "react",
		Backend:          "nodejs",
		Database:         "postgresql",
		Cloud:            "aws",
		Containerization: "docker",
		CI:               "github-actions",
		Monitoring:       "prometheus",
		Additional:       []string{"redis", "elasticsearch"},
	}

	assert.Equal(t, "react", stack.Frontend)
	assert.Equal(t, "nodejs", stack.Backend)
	assert.Equal(t, "postgresql", stack.Database)
	assert.Equal(t, "aws", stack.Cloud)
	assert.Equal(t, "docker", stack.Containerization)
	assert.Equal(t, "github-actions", stack.CI)
	assert.Equal(t, "prometheus", stack.Monitoring)
	assert.Equal(t, 2, len(stack.Additional))
}

func TestSkillRequirements(t *testing.T) {
	req := agent.SkillRequirements{
		TechStack: &agent.TechStack{
			Frontend: "vue",
		},
		Categories: []agent.SkillCategory{
			agent.CategoryFrontend,
			agent.CategoryBackend,
		},
		SkillNames: []string{"vue-developer"},
		MinVersion: "1.0.0",
	}

	assert.NotNil(t, req.TechStack)
	assert.Equal(t, 2, len(req.Categories))
	assert.Equal(t, 1, len(req.SkillNames))
	assert.Equal(t, "1.0.0", req.MinVersion)
}

func TestSkillCategory(t *testing.T) {
	categories := []agent.SkillCategory{
		agent.CategoryFrontend,
		agent.CategoryBackend,
		agent.CategoryDatabase,
		agent.CategoryDevOps,
		agent.CategoryInfrastructure,
		agent.CategoryQA,
		agent.CategorySRE,
		agent.CategorySecurity,
		agent.CategoryArchitect,
		agent.CategoryPM,
	}

	assert.Equal(t, 10, len(categories))
	assert.Equal(t, agent.SkillCategory("frontend"), agent.CategoryFrontend)
	assert.Equal(t, agent.SkillCategory("backend"), agent.CategoryBackend)
}

// Helper functions to create test registries

func createTestRegistry() *agent.SkillRegistry {
	registry := agent.NewSkillRegistry("/tmp/skills")

	skill := &agent.Skill{
		Name:         "react-test",
		Version:      "1.0.0",
		Description:  "React development",
		Capabilities: []string{"components", "hooks"},
		Tools:        []string{"file_write"},
		SystemPrompt: "You are a React developer.",
	}
	registry.RegisterSkill(skill)

	return registry
}

func createTestRegistryWithMultipleSkills() *agent.SkillRegistry {
	registry := agent.NewSkillRegistry("/tmp/skills")

	reactSkill := &agent.Skill{
		Name:         "react-test",
		Version:      "1.0.0",
		Description:  "React development",
		Capabilities: []string{"components"},
		Tools:        []string{"file_write"},
		SystemPrompt: "React prompt.",
	}
	registry.RegisterSkill(reactSkill)

	nodeSkill := &agent.Skill{
		Name:         "node-test",
		Version:      "1.0.0",
		Description:  "Node.js development",
		Capabilities: []string{"api"},
		Tools:        []string{"npm_run"},
		SystemPrompt: "Node prompt.",
	}
	registry.RegisterSkill(nodeSkill)

	return registry
}
