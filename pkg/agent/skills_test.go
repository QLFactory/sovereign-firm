package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSkillRegistry(t *testing.T) {
	registry := NewSkillRegistry("/tmp/skills")

	if registry == nil {
		t.Fatal("Expected registry to be created, got nil")
	}

	if registry.skillsDir != "/tmp/skills" {
		t.Errorf("Expected skills dir '/tmp/skills', got '%s'", registry.skillsDir)
	}

	if registry.Count() != 0 {
		t.Errorf("Expected empty registry, got %d skills", registry.Count())
	}
}

func TestSkillRegistryRegisterSkill(t *testing.T) {
	registry := NewSkillRegistry("/tmp/skills")

	skill := &Skill{
		Name:         "test-skill",
		Version:      "1.0.0",
		Description:  "A test skill",
		Capabilities: []string{"testing", "mocking"},
		Tools:        []string{"test_tool"},
		SystemPrompt: "You are a test skill.",
	}

	registry.RegisterSkill(skill)

	if registry.Count() != 1 {
		t.Errorf("Expected 1 skill, got %d", registry.Count())
	}

	if !registry.HasSkill("test-skill") {
		t.Error("Expected registry to have 'test-skill'")
	}
}

func TestSkillRegistryGetSkill(t *testing.T) {
	registry := NewSkillRegistry("/tmp/skills")

	skill := &Skill{
		Name:        "retrieval-skill",
		Version:     "2.0.0",
		Description: "Skill for retrieval testing",
	}
	registry.RegisterSkill(skill)

	// Test successful retrieval
	retrieved, err := registry.GetSkill("retrieval-skill")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if retrieved.Name != "retrieval-skill" {
		t.Errorf("Expected name 'retrieval-skill', got '%s'", retrieved.Name)
	}

	if retrieved.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", retrieved.Version)
	}

	// Test non-existent skill
	_, err = registry.GetSkill("non-existent")
	if err != ErrSkillNotFound {
		t.Errorf("Expected ErrSkillNotFound, got: %v", err)
	}
}

func TestSkillRegistryListSkills(t *testing.T) {
	registry := NewSkillRegistry("/tmp/skills")

	registry.RegisterSkill(&Skill{Name: "skill-a"})
	registry.RegisterSkill(&Skill{Name: "skill-b"})
	registry.RegisterSkill(&Skill{Name: "skill-c"})

	names := registry.ListSkills()

	if len(names) != 3 {
		t.Errorf("Expected 3 skills, got %d", len(names))
	}

	// Check all names are present (order not guaranteed)
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	for _, expected := range []string{"skill-a", "skill-b", "skill-c"} {
		if !nameMap[expected] {
			t.Errorf("Expected skill '%s' in list", expected)
		}
	}
}

func TestSkillRegistryHasSkill(t *testing.T) {
	registry := NewSkillRegistry("/tmp/skills")

	registry.RegisterSkill(&Skill{Name: "existing-skill"})

	if !registry.HasSkill("existing-skill") {
		t.Error("Expected HasSkill to return true for existing skill")
	}

	if registry.HasSkill("non-existing-skill") {
		t.Error("Expected HasSkill to return false for non-existing skill")
	}
}

func TestSkillRegistryLoadSkillsFromYAML(t *testing.T) {
	// Create a temporary directory with test skill files
	tempDir, err := os.MkdirTemp("", "skills-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a skill subdirectory
	skillDir := filepath.Join(tempDir, "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill dir: %v", err)
	}

	// Create a skill.yaml file
	skillYAML := `name: test-skill
version: "1.0.0"
description: A test skill loaded from YAML
capabilities:
  - testing
  - validation
tools:
  - file_read
  - file_write
prompts:
  system: |
    You are a testing agent.
    Follow best practices.
  examples:
    - description: Example test
      input: "Test this"
      output: "Tested successfully"
output_contract:
  type: json
  schema:
    type: object
    properties:
      result:
        type: string
`

	skillPath := filepath.Join(skillDir, "skill.yaml")
	if err := os.WriteFile(skillPath, []byte(skillYAML), 0644); err != nil {
		t.Fatalf("Failed to write skill file: %v", err)
	}

	// Load skills
	registry := NewSkillRegistry(tempDir)
	if err := registry.LoadSkills(); err != nil {
		t.Fatalf("Failed to load skills: %v", err)
	}

	// Verify skill was loaded
	if registry.Count() != 1 {
		t.Errorf("Expected 1 skill, got %d", registry.Count())
	}

	skill, err := registry.GetSkill("test-skill")
	if err != nil {
		t.Fatalf("Expected skill to be loaded, got error: %v", err)
	}

	if skill.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", skill.Version)
	}

	if skill.Description != "A test skill loaded from YAML" {
		t.Errorf("Unexpected description: %s", skill.Description)
	}

	if len(skill.Capabilities) != 2 {
		t.Errorf("Expected 2 capabilities, got %d", len(skill.Capabilities))
	}

	if len(skill.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(skill.Tools))
	}
}

func TestSkillRegistryGetSkillDefinition(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-def-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	skillDir := filepath.Join(tempDir, "def-skill")
	os.MkdirAll(skillDir, 0755)

	skillYAML := `name: def-skill
version: "2.0.0"
description: Definition test skill
capabilities: []
tools: []
prompts:
  system: System prompt here
output_contract:
  type: json
  schema:
    type: object
`

	os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(skillYAML), 0644)

	registry := NewSkillRegistry(tempDir)
	registry.LoadSkills()

	def, err := registry.GetSkillDefinition("def-skill")
	if err != nil {
		t.Fatalf("Expected definition, got error: %v", err)
	}

	if def.Name != "def-skill" {
		t.Errorf("Expected name 'def-skill', got '%s'", def.Name)
	}

	if def.OutputContract.Type != "json" {
		t.Errorf("Expected output contract type 'json', got '%s'", def.OutputContract.Type)
	}

	// Test non-existent definition
	_, err = registry.GetSkillDefinition("non-existent")
	if err != ErrSkillNotFound {
		t.Errorf("Expected ErrSkillNotFound, got: %v", err)
	}
}

func TestSkillRegistryLoadSkillsInvalidYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-invalid-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	skillDir := filepath.Join(tempDir, "invalid-skill")
	os.MkdirAll(skillDir, 0755)

	// Invalid YAML (missing name which is required)
	invalidYAML := `version: "1.0.0"
description: No name field
`

	os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(invalidYAML), 0644)

	registry := NewSkillRegistry(tempDir)
	err = registry.LoadSkills()

	if err == nil {
		t.Error("Expected error for skill without name, got nil")
	}
}

func TestSkillRegistryLoadSkillsMalformedYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "skills-malformed-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	skillDir := filepath.Join(tempDir, "malformed-skill")
	os.MkdirAll(skillDir, 0755)

	// Malformed YAML
	malformedYAML := `name: test
version: [invalid
description: broken`

	os.WriteFile(filepath.Join(skillDir, "skill.yaml"), []byte(malformedYAML), 0644)

	registry := NewSkillRegistry(tempDir)
	err = registry.LoadSkills()

	if err == nil {
		t.Error("Expected error for malformed YAML, got nil")
	}
}

func TestSkillRegistryConcurrentAccess(t *testing.T) {
	registry := NewSkillRegistry("/tmp/skills")

	// Register skills concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			skill := &Skill{Name: string(rune('a'+n)) + "-skill"}
			registry.RegisterSkill(skill)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Read concurrently
	for i := 0; i < 10; i++ {
		go func() {
			registry.ListSkills()
			registry.Count()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 skills
	if registry.Count() != 10 {
		t.Errorf("Expected 10 skills after concurrent registration, got %d", registry.Count())
	}
}
