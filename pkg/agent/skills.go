package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
	"gopkg.in/yaml.v3"
)

// SkillCategory represents a category of skills
type SkillCategory string

const (
	CategoryFrontend       SkillCategory = "frontend"
	CategoryBackend        SkillCategory = "backend"
	CategoryDatabase       SkillCategory = "database"
	CategoryDevOps         SkillCategory = "devops"
	CategoryInfrastructure SkillCategory = "infrastructure"
	CategoryQA             SkillCategory = "qa"
	CategorySRE            SkillCategory = "sre"
	CategorySecurity       SkillCategory = "security"
	CategoryArchitect      SkillCategory = "architect"
	CategoryPM             SkillCategory = "pm"
)

// TechStack represents the technology choices for a project
type TechStack struct {
	Frontend       string   `json:"frontend" yaml:"frontend"`               // React, Vue, Angular, Next.js
	Backend        string   `json:"backend" yaml:"backend"`                 // Node.js, Python, Go, Java
	Database       string   `json:"database" yaml:"database"`               // PostgreSQL, MongoDB, Redis, MySQL
	Cloud          string   `json:"cloud" yaml:"cloud"`                     // AWS, Azure, GCP
	Containerization string `json:"containerization" yaml:"containerization"` // Docker, Kubernetes
	CI             string   `json:"ci" yaml:"ci"`                           // GitHub Actions, GitLab CI, Jenkins
	Monitoring     string   `json:"monitoring" yaml:"monitoring"`           // Prometheus, DataDog, CloudWatch
	Additional     []string `json:"additional" yaml:"additional"`           // Additional technologies
}

// SkillRequirements specifies what skills are needed for a project
type SkillRequirements struct {
	TechStack    *TechStack    `json:"tech_stack" yaml:"tech_stack"`
	Categories   []SkillCategory `json:"categories" yaml:"categories"`
	SkillNames   []string      `json:"skill_names" yaml:"skill_names"` // Explicit skill names
	MinVersion   string        `json:"min_version" yaml:"min_version"` // Minimum skill version
}

// SkillDefinition is the full YAML structure for a skill file
type SkillDefinition struct {
	Name         string        `yaml:"name"`
	Version      string        `yaml:"version"`
	Description  string        `yaml:"description"`
	Category     SkillCategory `yaml:"category"`
	Tags         []string      `yaml:"tags"`         // e.g., ["react", "typescript", "frontend"]
	Capabilities []string      `yaml:"capabilities"`
	Tools        []string      `yaml:"tools"`
	Dependencies []string      `yaml:"dependencies"` // Other skills this depends on

	Prompts struct {
		System   string `yaml:"system"`
		Examples []struct {
			Description string `yaml:"description"`
			Input       string `yaml:"input"`
			Output      string `yaml:"output"`
		} `yaml:"examples"`
		Constraints []string `yaml:"constraints"` // Best practices, anti-patterns to avoid
	} `yaml:"prompts"`

	OutputContract struct {
		Type       string                 `yaml:"type"` // "json", "code", "text"
		Schema     map[string]interface{} `yaml:"schema,omitempty"`
		Validation []string               `yaml:"validation,omitempty"`
	} `yaml:"output_contract"`
}

// SkillMatch represents a skill that matches requirements
type SkillMatch struct {
	Skill      *Skill
	Definition *SkillDefinition
	Score      int // Relevance score
	Reason     string
}

// SkillRegistry manages loading and accessing skills
type SkillRegistry struct {
	skills      map[string]*Skill
	definitions map[string]*SkillDefinition
	skillsDir   string
	mu          sync.RWMutex
}

// NewSkillRegistry creates a new skill registry
func NewSkillRegistry(skillsDir string) *SkillRegistry {
	return &SkillRegistry{
		skills:      make(map[string]*Skill),
		definitions: make(map[string]*SkillDefinition),
		skillsDir:   skillsDir,
	}
}

// LoadSkills loads all skills from the skills directory
func (r *SkillRegistry) LoadSkills() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find all skill.yaml files
	pattern := filepath.Join(r.skillsDir, "*", "skill.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find skills: %w", err)
	}

	// Also check for flat structure
	flatPattern := filepath.Join(r.skillsDir, "*.yaml")
	flatMatches, err := filepath.Glob(flatPattern)
	if err == nil {
		matches = append(matches, flatMatches...)
	}

	for _, path := range matches {
		if err := r.loadSkillFile(path); err != nil {
			return fmt.Errorf("failed to load %s: %w", path, err)
		}
	}

	return nil
}

// loadSkillFile loads a single skill from a YAML file
func (r *SkillRegistry) loadSkillFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var def SkillDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if def.Name == "" {
		return fmt.Errorf("skill name is required")
	}

	// Store the full definition
	r.definitions[def.Name] = &def

	// Create the simplified Skill for agents
	skill := &Skill{
		Name:         def.Name,
		Version:      def.Version,
		Description:  def.Description,
		Capabilities: def.Capabilities,
		Tools:        def.Tools,
		SystemPrompt: def.Prompts.System,
	}

	r.skills[def.Name] = skill

	return nil
}

// RegisterSkill manually registers a skill (for testing or built-in skills)
func (r *SkillRegistry) RegisterSkill(skill *Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills[skill.Name] = skill

	// Also create a minimal definition for consistency
	if _, exists := r.definitions[skill.Name]; !exists {
		r.definitions[skill.Name] = &SkillDefinition{
			Name:         skill.Name,
			Version:      skill.Version,
			Description:  skill.Description,
			Capabilities: skill.Capabilities,
			Tools:        skill.Tools,
		}
		r.definitions[skill.Name].Prompts.System = skill.SystemPrompt
	}
}

// GetSkill returns a skill by name
func (r *SkillRegistry) GetSkill(name string) (*Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, ok := r.skills[name]
	if !ok {
		return nil, ErrSkillNotFound
	}
	return skill, nil
}

// GetSkillDefinition returns the full skill definition
func (r *SkillRegistry) GetSkillDefinition(name string) (*SkillDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, ok := r.definitions[name]
	if !ok {
		return nil, ErrSkillNotFound
	}
	return def, nil
}

// ListSkills returns all available skill names
func (r *SkillRegistry) ListSkills() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.skills))
	for name := range r.skills {
		names = append(names, name)
	}
	return names
}

// HasSkill checks if a skill is registered
func (r *SkillRegistry) HasSkill(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.skills[name]
	return ok
}

// Count returns the number of registered skills
func (r *SkillRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.skills)
}

// GetSkillsByCategory returns all skills in a category
func (r *SkillRegistry) GetSkillsByCategory(category SkillCategory) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Skill
	for name, def := range r.definitions {
		if def.Category == category {
			if skill, ok := r.skills[name]; ok {
				result = append(result, skill)
			}
		}
	}
	return result
}

// GetSkillsByTag returns all skills that have a specific tag
func (r *SkillRegistry) GetSkillsByTag(tag string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tag = strings.ToLower(tag)
	var result []*Skill
	for name, def := range r.definitions {
		for _, t := range def.Tags {
			if strings.ToLower(t) == tag {
				if skill, ok := r.skills[name]; ok {
					result = append(result, skill)
				}
				break
			}
		}
	}
	return result
}

// SelectSkillsForTechStack automatically selects appropriate skills based on tech stack
func (r *SkillRegistry) SelectSkillsForTechStack(stack *TechStack) []SkillMatch {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []SkillMatch

	// Map tech stack to skill tags
	techToTags := map[string][]string{
		// Frontend
		"react":     {"react", "frontend", "javascript", "typescript"},
		"vue":       {"vue", "frontend", "javascript", "typescript"},
		"angular":   {"angular", "frontend", "typescript"},
		"nextjs":    {"nextjs", "react", "frontend", "typescript"},
		"next.js":   {"nextjs", "react", "frontend", "typescript"},
		"svelte":    {"svelte", "frontend", "javascript"},

		// Backend
		"node":      {"nodejs", "backend", "javascript", "typescript"},
		"nodejs":    {"nodejs", "backend", "javascript", "typescript"},
		"node.js":   {"nodejs", "backend", "javascript", "typescript"},
		"express":   {"express", "nodejs", "backend"},
		"python":    {"python", "backend"},
		"fastapi":   {"fastapi", "python", "backend"},
		"django":    {"django", "python", "backend"},
		"flask":     {"flask", "python", "backend"},
		"go":        {"go", "golang", "backend"},
		"golang":    {"go", "golang", "backend"},
		"java":      {"java", "backend"},
		"spring":    {"spring", "java", "backend"},

		// Database
		"postgresql": {"postgresql", "postgres", "database", "sql"},
		"postgres":   {"postgresql", "postgres", "database", "sql"},
		"mysql":      {"mysql", "database", "sql"},
		"mongodb":    {"mongodb", "database", "nosql"},
		"redis":      {"redis", "database", "cache"},
		"dynamodb":   {"dynamodb", "database", "nosql", "aws"},

		// Cloud
		"aws":        {"aws", "cloud"},
		"azure":      {"azure", "cloud"},
		"gcp":        {"gcp", "cloud"},

		// Container/Orchestration
		"docker":     {"docker", "container", "devops"},
		"kubernetes": {"kubernetes", "k8s", "devops"},
		"k8s":        {"kubernetes", "k8s", "devops"},

		// CI/CD
		"github-actions": {"github-actions", "ci", "devops"},
		"gitlab-ci":      {"gitlab-ci", "ci", "devops"},
		"jenkins":        {"jenkins", "ci", "devops"},

		// Monitoring
		"prometheus": {"prometheus", "monitoring", "sre"},
		"grafana":    {"grafana", "monitoring", "sre"},
		"datadog":    {"datadog", "monitoring", "sre"},
	}

	// Collect all relevant tags from tech stack
	relevantTags := make(map[string]int) // tag -> score
	addTags := func(tech string, baseScore int) {
		if tech == "" {
			return
		}
		tech = strings.ToLower(tech)
		if tags, ok := techToTags[tech]; ok {
			for i, tag := range tags {
				score := baseScore - i // Primary tag gets higher score
				if relevantTags[tag] < score {
					relevantTags[tag] = score
				}
			}
		}
	}

	addTags(stack.Frontend, 10)
	addTags(stack.Backend, 10)
	addTags(stack.Database, 8)
	addTags(stack.Cloud, 6)
	addTags(stack.Containerization, 6)
	addTags(stack.CI, 5)
	addTags(stack.Monitoring, 5)
	for _, tech := range stack.Additional {
		addTags(tech, 4)
	}

	// Score each skill based on tag matches
	for name, def := range r.definitions {
		score := 0
		var matchedTags []string

		for _, tag := range def.Tags {
			if tagScore, ok := relevantTags[strings.ToLower(tag)]; ok {
				score += tagScore
				matchedTags = append(matchedTags, tag)
			}
		}

		if score > 0 {
			skill := r.skills[name]
			matches = append(matches, SkillMatch{
				Skill:      skill,
				Definition: def,
				Score:      score,
				Reason:     fmt.Sprintf("Matched tags: %v", matchedTags),
			})
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	return matches
}

// SelectSkillsForRequirements selects skills based on comprehensive requirements
func (r *SkillRegistry) SelectSkillsForRequirements(req *SkillRequirements) []SkillMatch {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skillScores := make(map[string]int)
	skillReasons := make(map[string][]string)

	// 1. Explicit skill names have highest priority
	for _, name := range req.SkillNames {
		if _, ok := r.skills[name]; ok {
			skillScores[name] = 100
			skillReasons[name] = append(skillReasons[name], "Explicitly requested")
		}
	}

	// 2. Category matches
	for _, cat := range req.Categories {
		for name, def := range r.definitions {
			if def.Category == cat {
				skillScores[name] += 50
				skillReasons[name] = append(skillReasons[name], fmt.Sprintf("Category: %s", cat))
			}
		}
	}

	// 3. Tech stack matches
	if req.TechStack != nil {
		techMatches := r.SelectSkillsForTechStack(req.TechStack)
		for _, match := range techMatches {
			skillScores[match.Skill.Name] += match.Score
			skillReasons[match.Skill.Name] = append(skillReasons[match.Skill.Name], match.Reason)
		}
	}

	// Build final match list
	var matches []SkillMatch
	for name, score := range skillScores {
		if score > 0 {
			skill, skillOk := r.skills[name]
			def, defOk := r.definitions[name]
			// Skip if skill doesn't exist (shouldn't happen but be safe)
			if !skillOk || skill == nil {
				continue
			}
			// Create minimal definition if not found
			if !defOk || def == nil {
				def = &SkillDefinition{
					Name:        skill.Name,
					Version:     skill.Version,
					Description: skill.Description,
				}
			}
			matches = append(matches, SkillMatch{
				Skill:      skill,
				Definition: def,
				Score:      score,
				Reason:     strings.Join(skillReasons[name], "; "),
			})
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	return matches
}

// GetSkilledAgent creates an agent with skills injected based on requirements.
// The ctx parameter should be a parent context (e.g., pool context) for proper cancellation.
func (r *SkillRegistry) GetSkilledAgent(
	ctx context.Context,
	name, role, projectID, model string,
	requirements *SkillRequirements,
	llmClient llm.Client,
) (*AgentInstance, []SkillMatch) {
	// Select skills based on requirements
	matches := r.SelectSkillsForRequirements(requirements)

	// Extract skills from matches
	var skills []Skill
	for _, match := range matches {
		if match.Skill != nil {
			skills = append(skills, *match.Skill)
		}
	}

	// Create agent with selected skills
	agent := NewAgentInstance(ctx, name, role, projectID, model, skills, llmClient)

	return agent, matches
}

// ComposeSkills combines multiple skills into a merged skill
func (r *SkillRegistry) ComposeSkills(names ...string) (*Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(names) == 0 {
		return nil, fmt.Errorf("no skill names provided")
	}

	composed := &Skill{
		Name:         strings.Join(names, "+"),
		Capabilities: []string{},
		Tools:        []string{},
	}

	seenCaps := make(map[string]bool)
	seenTools := make(map[string]bool)
	var prompts []string

	for _, name := range names {
		skill, ok := r.skills[name]
		if !ok {
			return nil, fmt.Errorf("skill not found: %s", name)
		}

		// Merge capabilities
		for _, cap := range skill.Capabilities {
			if !seenCaps[cap] {
				seenCaps[cap] = true
				composed.Capabilities = append(composed.Capabilities, cap)
			}
		}

		// Merge tools
		for _, tool := range skill.Tools {
			if !seenTools[tool] {
				seenTools[tool] = true
				composed.Tools = append(composed.Tools, tool)
			}
		}

		// Combine system prompts
		if skill.SystemPrompt != "" {
			prompts = append(prompts, skill.SystemPrompt)
		}

		// Use latest version
		if skill.Version > composed.Version {
			composed.Version = skill.Version
		}
	}

	composed.SystemPrompt = strings.Join(prompts, "\n\n---\n\n")
	composed.Description = fmt.Sprintf("Composed skill from: %v", names)

	return composed, nil
}

// GetDependencies returns all dependencies for a skill (recursive)
func (r *SkillRegistry) GetDependencies(skillName string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, ok := r.definitions[skillName]
	if !ok {
		return nil, ErrSkillNotFound
	}

	visited := make(map[string]bool)
	var deps []string

	var collect func(name string) error
	collect = func(name string) error {
		if visited[name] {
			return nil
		}
		visited[name] = true

		d, ok := r.definitions[name]
		if !ok {
			return fmt.Errorf("dependency not found: %s", name)
		}

		for _, depName := range d.Dependencies {
			if err := collect(depName); err != nil {
				return err
			}
			deps = append(deps, depName)
		}
		return nil
	}

	for _, depName := range def.Dependencies {
		if err := collect(depName); err != nil {
			return nil, err
		}
		deps = append(deps, depName)
	}

	return deps, nil
}

// GetSkillWithDependencies returns a skill and all its dependencies
func (r *SkillRegistry) GetSkillWithDependencies(skillName string) ([]*Skill, error) {
	deps, err := r.GetDependencies(skillName)
	if err != nil {
		return nil, err
	}

	// Get main skill
	mainSkill, err := r.GetSkill(skillName)
	if err != nil {
		return nil, err
	}

	skills := []*Skill{mainSkill}

	// Get dependency skills
	for _, depName := range deps {
		skill, err := r.GetSkill(depName)
		if err != nil {
			return nil, fmt.Errorf("failed to get dependency %s: %w", depName, err)
		}
		skills = append(skills, skill)
	}

	return skills, nil
}

// ExportSkillsToPrompt generates a combined system prompt from selected skills
func (r *SkillRegistry) ExportSkillsToPrompt(skillNames []string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sections []string

	for _, name := range skillNames {
		skill, ok := r.skills[name]
		if !ok {
			continue
		}
		def, ok := r.definitions[name]
		if !ok {
			continue
		}

		section := fmt.Sprintf("## %s (v%s)\n\n%s", skill.Name, skill.Version, skill.SystemPrompt)

		// Add constraints if any
		if len(def.Prompts.Constraints) > 0 {
			section += "\n\n### Constraints\n"
			for _, c := range def.Prompts.Constraints {
				section += fmt.Sprintf("- %s\n", c)
			}
		}

		// Add examples if any
		if len(def.Prompts.Examples) > 0 {
			section += "\n\n### Examples\n"
			for _, ex := range def.Prompts.Examples {
				section += fmt.Sprintf("\n**%s**\nInput: %s\nOutput:\n```\n%s\n```\n",
					ex.Description, ex.Input, ex.Output)
			}
		}

		sections = append(sections, section)
	}

	return strings.Join(sections, "\n\n---\n\n")
}

// GetAllCategories returns all unique categories
func (r *SkillRegistry) GetAllCategories() []SkillCategory {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[SkillCategory]bool)
	for _, def := range r.definitions {
		if def.Category != "" {
			seen[def.Category] = true
		}
	}

	cats := make([]SkillCategory, 0, len(seen))
	for cat := range seen {
		cats = append(cats, cat)
	}
	return cats
}

// GetAllTags returns all unique tags
func (r *SkillRegistry) GetAllTags() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	for _, def := range r.definitions {
		for _, tag := range def.Tags {
			seen[tag] = true
		}
	}

	tags := make([]string, 0, len(seen))
	for tag := range seen {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}
