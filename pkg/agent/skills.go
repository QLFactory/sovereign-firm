package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// SkillDefinition is the full YAML structure for a skill file
type SkillDefinition struct {
	Name         string   `yaml:"name"`
	Version      string   `yaml:"version"`
	Description  string   `yaml:"description"`
	Capabilities []string `yaml:"capabilities"`
	Tools        []string `yaml:"tools"`

	Prompts struct {
		System   string `yaml:"system"`
		Examples []struct {
			Description string `yaml:"description"`
			Input       string `yaml:"input"`
			Output      string `yaml:"output"`
		} `yaml:"examples"`
	} `yaml:"prompts"`

	OutputContract struct {
		Type       string                 `yaml:"type"` // "json", "code", "text"
		Schema     map[string]interface{} `yaml:"schema,omitempty"`
		Validation []string               `yaml:"validation,omitempty"`
	} `yaml:"output_contract"`
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
