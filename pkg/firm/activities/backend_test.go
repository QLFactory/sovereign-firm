package activities

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBackendCodeBundleStructure(t *testing.T) {
	bundle := BackendCodeBundle{
		Stack: "node-express",
		Files: map[string]string{
			"src/index.ts":       "console.log('hello')",
			"package.json":       "{}",
			"tsconfig.json":      "{}",
		},
		EntryPoint: "src/index.ts",
		StartCmd:   "npm run dev",
		InstallCmd: "npm install",
		TestCmd:    "npm test",
		BuildCmd:   "npm run build",
	}

	// Test JSON serialization
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("Failed to marshal BackendCodeBundle: %v", err)
	}

	var unmarshaled BackendCodeBundle
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal BackendCodeBundle: %v", err)
	}

	if unmarshaled.Stack != "node-express" {
		t.Errorf("Stack mismatch: got %s, want node-express", unmarshaled.Stack)
	}

	if len(unmarshaled.Files) != 3 {
		t.Errorf("Files count mismatch: got %d, want 3", len(unmarshaled.Files))
	}

	if unmarshaled.EntryPoint != "src/index.ts" {
		t.Errorf("EntryPoint mismatch: got %s, want src/index.ts", unmarshaled.EntryPoint)
	}

	if unmarshaled.StartCmd != "npm run dev" {
		t.Errorf("StartCmd mismatch: got %s, want npm run dev", unmarshaled.StartCmd)
	}
}

func TestBackendGenerateInputStructure(t *testing.T) {
	input := BackendGenerateInput{
		ProjectName:  "My API",
		Requirements: "Build a REST API for user management",
		TechStack: TechStack{
			Backend: BackendStack{
				Framework: "Express",
				Language:  "TypeScript",
				Runtime:   "Node.js 20",
				ORM:       "Prisma",
			},
			Database: DatabaseStack{
				Primary: "PostgreSQL",
			},
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal BackendGenerateInput: %v", err)
	}

	var unmarshaled BackendGenerateInput
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal BackendGenerateInput: %v", err)
	}

	if unmarshaled.ProjectName != "My API" {
		t.Errorf("ProjectName mismatch")
	}

	if unmarshaled.TechStack.Backend.Framework != "Express" {
		t.Errorf("Backend.Framework mismatch")
	}
}

func TestBackendRefineInputStructure(t *testing.T) {
	input := BackendRefineInput{
		CurrentCode: map[string]string{
			"src/index.ts": "const app = express();",
		},
		TechStack: TechStack{
			Backend: BackendStack{
				Framework: "Express",
				Language:  "TypeScript",
			},
		},
		Feedback:         "Add error handling",
		ValidationErrors: "Missing return type on line 10",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal BackendRefineInput: %v", err)
	}

	var unmarshaled BackendRefineInput
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal BackendRefineInput: %v", err)
	}

	if len(unmarshaled.CurrentCode) != 1 {
		t.Errorf("CurrentCode count mismatch")
	}

	if unmarshaled.Feedback != "Add error handling" {
		t.Errorf("Feedback mismatch")
	}

	if unmarshaled.ValidationErrors != "Missing return type on line 10" {
		t.Errorf("ValidationErrors mismatch")
	}
}

func TestDetermineBackendStack(t *testing.T) {
	tests := []struct {
		name     string
		stack    TechStack
		expected string
	}{
		{
			name: "Node Express from TypeScript",
			stack: TechStack{
				Backend: BackendStack{Language: "TypeScript", Framework: "Express"},
			},
			expected: "node-express",
		},
		{
			name: "Node Fastify",
			stack: TechStack{
				Backend: BackendStack{Language: "TypeScript", Framework: "Fastify"},
			},
			expected: "node-fastify",
		},
		{
			name: "Node NestJS",
			stack: TechStack{
				Backend: BackendStack{Language: "TypeScript", Framework: "NestJS"},
			},
			expected: "node-nestjs",
		},
		{
			name: "Python FastAPI",
			stack: TechStack{
				Backend: BackendStack{Language: "Python", Framework: "FastAPI"},
			},
			expected: "python-fastapi",
		},
		{
			name: "Python Django",
			stack: TechStack{
				Backend: BackendStack{Language: "Python", Framework: "Django"},
			},
			expected: "python-django",
		},
		{
			name: "Python Flask",
			stack: TechStack{
				Backend: BackendStack{Language: "Python", Framework: "Flask"},
			},
			expected: "python-flask",
		},
		{
			name: "Go Gin",
			stack: TechStack{
				Backend: BackendStack{Language: "Go", Framework: "Gin"},
			},
			expected: "go-gin",
		},
		{
			name: "Go Echo",
			stack: TechStack{
				Backend: BackendStack{Language: "Go", Framework: "Echo"},
			},
			expected: "go-echo",
		},
		{
			name: "Go Fiber",
			stack: TechStack{
				Backend: BackendStack{Language: "Go", Framework: "Fiber"},
			},
			expected: "go-fiber",
		},
		{
			name: "Default to Node Express",
			stack: TechStack{
				Backend: BackendStack{Language: "", Framework: ""},
			},
			expected: "node-express",
		},
		{
			name: "JavaScript defaults to Node",
			stack: TechStack{
				Backend: BackendStack{Language: "JavaScript", Framework: ""},
			},
			expected: "node-express",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineBackendStack(tt.stack)
			if result != tt.expected {
				t.Errorf("determineBackendStack() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestNewBackendAgent(t *testing.T) {
	agent := NewBackendAgent()
	if agent == nil {
		t.Fatal("NewBackendAgent returned nil")
	}

	if agent.llmClient == nil {
		t.Error("BackendAgent should have llmClient initialized")
	}
}

func TestEnsureNodeDefaults(t *testing.T) {
	agent := NewBackendAgent()

	// Empty bundle
	bundle := BackendCodeBundle{
		Files: nil,
	}

	result := agent.ensureNodeDefaults(bundle, "node-express")

	if result.Stack != "node-express" {
		t.Errorf("Stack should be node-express, got %s", result.Stack)
	}

	if result.EntryPoint != "src/index.ts" {
		t.Errorf("EntryPoint should be src/index.ts, got %s", result.EntryPoint)
	}

	if result.StartCmd != "npm run dev" {
		t.Errorf("StartCmd should be npm run dev, got %s", result.StartCmd)
	}

	if result.InstallCmd != "npm install" {
		t.Errorf("InstallCmd should be npm install, got %s", result.InstallCmd)
	}

	if result.Files == nil {
		t.Error("Files should not be nil after ensureNodeDefaults")
	}

	// Should have default package.json
	if _, exists := result.Files["package.json"]; !exists {
		t.Error("Should have default package.json")
	}

	// Should have default tsconfig.json
	if _, exists := result.Files["tsconfig.json"]; !exists {
		t.Error("Should have default tsconfig.json")
	}
}

func TestEnsurePythonDefaults(t *testing.T) {
	agent := NewBackendAgent()

	bundle := BackendCodeBundle{
		Files: nil,
	}

	result := agent.ensurePythonDefaults(bundle, "python-fastapi")

	if result.Stack != "python-fastapi" {
		t.Errorf("Stack should be python-fastapi, got %s", result.Stack)
	}

	if result.EntryPoint != "app/main.py" {
		t.Errorf("EntryPoint should be app/main.py, got %s", result.EntryPoint)
	}

	if !strings.Contains(result.StartCmd, "uvicorn") {
		t.Errorf("StartCmd should contain uvicorn, got %s", result.StartCmd)
	}

	if result.InstallCmd != "pip install -r requirements.txt" {
		t.Errorf("InstallCmd should be pip install, got %s", result.InstallCmd)
	}

	// Should have default requirements.txt
	if _, exists := result.Files["requirements.txt"]; !exists {
		t.Error("Should have default requirements.txt")
	}
}

func TestEnsureGoDefaults(t *testing.T) {
	agent := NewBackendAgent()

	bundle := BackendCodeBundle{
		Files: nil,
	}

	result := agent.ensureGoDefaults(bundle, "go-gin", "my-project")

	if result.Stack != "go-gin" {
		t.Errorf("Stack should be go-gin, got %s", result.Stack)
	}

	if result.EntryPoint != "cmd/server/main.go" {
		t.Errorf("EntryPoint should be cmd/server/main.go, got %s", result.EntryPoint)
	}

	if !strings.Contains(result.StartCmd, "go run") {
		t.Errorf("StartCmd should contain 'go run', got %s", result.StartCmd)
	}

	if result.InstallCmd != "go mod download" {
		t.Errorf("InstallCmd should be go mod download, got %s", result.InstallCmd)
	}

	// Should have default go.mod
	if _, exists := result.Files["go.mod"]; !exists {
		t.Error("Should have default go.mod")
	}

	// go.mod should contain project name
	goMod := result.Files["go.mod"]
	if !strings.Contains(goMod, "my-project") {
		t.Error("go.mod should contain project name")
	}
}

func TestGetDefaultNodePackageJSON(t *testing.T) {
	agent := NewBackendAgent()
	pkg := agent.getDefaultNodePackageJSON()

	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(pkg), &parsed); err != nil {
		t.Fatalf("Default package.json is not valid JSON: %v", err)
	}

	// Should have required fields
	if _, exists := parsed["dependencies"]; !exists {
		t.Error("package.json should have dependencies")
	}

	if _, exists := parsed["devDependencies"]; !exists {
		t.Error("package.json should have devDependencies")
	}

	if _, exists := parsed["scripts"]; !exists {
		t.Error("package.json should have scripts")
	}

	// Check for Express
	deps := parsed["dependencies"].(map[string]interface{})
	if _, exists := deps["express"]; !exists {
		t.Error("package.json should include express dependency")
	}
}

func TestGetDefaultTSConfig(t *testing.T) {
	agent := NewBackendAgent()
	tsconfig := agent.getDefaultTSConfig()

	// Should be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(tsconfig), &parsed); err != nil {
		t.Fatalf("Default tsconfig.json is not valid JSON: %v", err)
	}

	// Should have compilerOptions
	if _, exists := parsed["compilerOptions"]; !exists {
		t.Error("tsconfig.json should have compilerOptions")
	}

	options := parsed["compilerOptions"].(map[string]interface{})

	// Check strict mode
	if strict, ok := options["strict"].(bool); !ok || !strict {
		t.Error("tsconfig should have strict: true")
	}
}

func TestGetDefaultPythonRequirements(t *testing.T) {
	agent := NewBackendAgent()

	// Test FastAPI requirements
	fastapiReqs := agent.getDefaultPythonRequirements("python-fastapi")
	if !strings.Contains(fastapiReqs, "fastapi") {
		t.Error("FastAPI requirements should include fastapi")
	}
	if !strings.Contains(fastapiReqs, "uvicorn") {
		t.Error("FastAPI requirements should include uvicorn")
	}
	if !strings.Contains(fastapiReqs, "pydantic") {
		t.Error("FastAPI requirements should include pydantic")
	}

	// Test Django requirements
	djangoReqs := agent.getDefaultPythonRequirements("python-django")
	if !strings.Contains(djangoReqs, "django") {
		t.Error("Django requirements should include django")
	}
	if !strings.Contains(djangoReqs, "djangorestframework") {
		t.Error("Django requirements should include djangorestframework")
	}
}

func TestGetDefaultGoMod(t *testing.T) {
	agent := NewBackendAgent()

	// Test Gin
	ginMod := agent.getDefaultGoMod("myapp", "go-gin")
	if !strings.Contains(ginMod, "module myapp") {
		t.Error("go.mod should contain module name")
	}
	if !strings.Contains(ginMod, "gin-gonic/gin") {
		t.Error("go.mod for Gin should include gin-gonic/gin")
	}

	// Test Echo
	echoMod := agent.getDefaultGoMod("myapp", "go-echo")
	if !strings.Contains(echoMod, "labstack/echo") {
		t.Error("go.mod for Echo should include labstack/echo")
	}

	// Test Fiber
	fiberMod := agent.getDefaultGoMod("myapp", "go-fiber")
	if !strings.Contains(fiberMod, "gofiber/fiber") {
		t.Error("go.mod for Fiber should include gofiber/fiber")
	}
}

func TestBuildBackendPrompt(t *testing.T) {
	agent := NewBackendAgent()

	req := BackendGenerateInput{
		ProjectName:  "Test API",
		Requirements: "Build a user management API",
		TechStack: TechStack{
			Backend: BackendStack{Framework: "Express"},
		},
		DatabaseSchema: &DatabaseSchema{
			DatabaseType: "PostgreSQL",
			Tables: []TableSchema{
				{Name: "users", Columns: []ColumnDef{{Name: "id", Type: "uuid"}}},
			},
		},
		APISpec: &APISpec{
			Type:     "REST",
			BasePath: "/api/v1",
		},
	}

	prompt := agent.buildBackendPrompt(req)

	// Should contain project name
	if !strings.Contains(prompt, "Test API") {
		t.Error("Prompt should contain project name")
	}

	// Should contain requirements
	if !strings.Contains(prompt, "user management API") {
		t.Error("Prompt should contain requirements")
	}

	// Should contain database schema
	if !strings.Contains(prompt, "users") {
		t.Error("Prompt should contain database schema info")
	}

	// Should contain API spec
	if !strings.Contains(prompt, "/api/v1") {
		t.Error("Prompt should contain API spec info")
	}
}

func TestModelPrompts(t *testing.T) {
	agent := NewBackendAgent()

	// Test Node model prompts
	prismaPrompt := agent.getNodeModelPrompt("prisma")
	if !strings.Contains(prismaPrompt, "Prisma") {
		t.Error("Prisma prompt should mention Prisma")
	}

	typeormPrompt := agent.getNodeModelPrompt("typeorm")
	if !strings.Contains(typeormPrompt, "TypeORM") {
		t.Error("TypeORM prompt should mention TypeORM")
	}

	// Test Python model prompts
	sqlalchemyPrompt := agent.getPythonModelPrompt("sqlalchemy")
	if !strings.Contains(sqlalchemyPrompt, "SQLAlchemy") {
		t.Error("SQLAlchemy prompt should mention SQLAlchemy")
	}

	djangoPrompt := agent.getPythonModelPrompt("django")
	if !strings.Contains(djangoPrompt, "Django") {
		t.Error("Django prompt should mention Django")
	}

	// Test Go model prompts
	gormPrompt := agent.getGoModelPrompt("gorm")
	if !strings.Contains(gormPrompt, "GORM") {
		t.Error("GORM prompt should mention GORM")
	}

	sqlxPrompt := agent.getGoModelPrompt("sqlx")
	if !strings.Contains(sqlxPrompt, "sqlx") {
		t.Error("sqlx prompt should mention sqlx")
	}
}

func TestRoutePrompts(t *testing.T) {
	agent := NewBackendAgent()

	expressPrompt := agent.getExpressRoutePrompt()
	if !strings.Contains(expressPrompt, "Express") {
		t.Error("Express prompt should mention Express")
	}

	fastifyPrompt := agent.getFastifyRoutePrompt()
	if !strings.Contains(fastifyPrompt, "Fastify") {
		t.Error("Fastify prompt should mention Fastify")
	}

	fastapiPrompt := agent.getFastAPIRoutePrompt()
	if !strings.Contains(fastapiPrompt, "FastAPI") {
		t.Error("FastAPI prompt should mention FastAPI")
	}

	ginPrompt := agent.getGinRoutePrompt()
	if !strings.Contains(ginPrompt, "Gin") {
		t.Error("Gin prompt should mention Gin")
	}
}

func TestBackendCodeBundleEmpty(t *testing.T) {
	bundle := BackendCodeBundle{}

	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("Failed to marshal empty BackendCodeBundle: %v", err)
	}

	var unmarshaled BackendCodeBundle
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal empty BackendCodeBundle: %v", err)
	}

	if unmarshaled.Files != nil && len(unmarshaled.Files) != 0 {
		t.Error("Empty bundle should have nil or empty files")
	}
}

func TestBackendInputWithDatabaseSchema(t *testing.T) {
	input := BackendGenerateInput{
		ProjectName:  "Blog API",
		Requirements: "Create a blog API",
		TechStack: TechStack{
			Backend: BackendStack{
				Framework: "FastAPI",
				Language:  "Python",
			},
		},
		DatabaseSchema: &DatabaseSchema{
			DatabaseType: "PostgreSQL",
			Tables: []TableSchema{
				{
					Name:        "posts",
					Description: "Blog posts",
					Columns: []ColumnDef{
						{Name: "id", Type: "uuid", Nullable: false},
						{Name: "title", Type: "varchar(255)", Nullable: false},
						{Name: "content", Type: "text", Nullable: true},
						{Name: "author_id", Type: "uuid", References: "users.id"},
					},
					PrimaryKey: []string{"id"},
					Timestamps: true,
				},
			},
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input with schema: %v", err)
	}

	var unmarshaled BackendGenerateInput
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal input with schema: %v", err)
	}

	if unmarshaled.DatabaseSchema == nil {
		t.Fatal("DatabaseSchema should not be nil")
	}

	if len(unmarshaled.DatabaseSchema.Tables) != 1 {
		t.Errorf("Should have 1 table, got %d", len(unmarshaled.DatabaseSchema.Tables))
	}

	if unmarshaled.DatabaseSchema.Tables[0].Name != "posts" {
		t.Errorf("Table name should be posts")
	}
}

func TestBackendInputWithAPISpec(t *testing.T) {
	input := BackendGenerateInput{
		ProjectName:  "User API",
		Requirements: "Create a user API",
		TechStack: TechStack{
			Backend: BackendStack{
				Framework: "Express",
				Language:  "TypeScript",
			},
		},
		APISpec: &APISpec{
			Type:     "REST",
			Version:  "v1",
			BasePath: "/api/v1",
			Endpoints: []APIEndpoint{
				{
					Method:  "GET",
					Path:    "/users",
					Summary: "List users",
				},
				{
					Method:  "POST",
					Path:    "/users",
					Summary: "Create user",
				},
			},
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input with API spec: %v", err)
	}

	var unmarshaled BackendGenerateInput
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal input with API spec: %v", err)
	}

	if unmarshaled.APISpec == nil {
		t.Fatal("APISpec should not be nil")
	}

	if len(unmarshaled.APISpec.Endpoints) != 2 {
		t.Errorf("Should have 2 endpoints, got %d", len(unmarshaled.APISpec.Endpoints))
	}
}

func TestSupportedStacks(t *testing.T) {
	supportedStacks := []string{
		"node-express",
		"node-fastify",
		"node-nestjs",
		"python-fastapi",
		"python-django",
		"python-flask",
		"go-gin",
		"go-echo",
		"go-fiber",
	}

	for _, stack := range supportedStacks {
		bundle := BackendCodeBundle{
			Stack: stack,
		}

		data, err := json.Marshal(bundle)
		if err != nil {
			t.Errorf("Failed to marshal bundle with stack %s: %v", stack, err)
		}

		var unmarshaled BackendCodeBundle
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Errorf("Failed to unmarshal bundle with stack %s: %v", stack, err)
		}

		if unmarshaled.Stack != stack {
			t.Errorf("Stack mismatch for %s: got %s", stack, unmarshaled.Stack)
		}
	}
}

func TestEmptyProjectNameInGoMod(t *testing.T) {
	agent := NewBackendAgent()

	// Test with empty project name
	bundle := BackendCodeBundle{Files: nil}
	result := agent.ensureGoDefaults(bundle, "go-gin", "")

	goMod := result.Files["go.mod"]
	if !strings.Contains(goMod, "module myapp") {
		t.Error("Empty project name should default to 'myapp' in go.mod")
	}
}

func TestProjectNameSanitization(t *testing.T) {
	agent := NewBackendAgent()

	// Test with spaces in project name
	bundle := BackendCodeBundle{Files: nil}
	result := agent.ensureGoDefaults(bundle, "go-gin", "My Cool Project")

	goMod := result.Files["go.mod"]
	// Extract the module name line
	lines := strings.Split(goMod, "\n")
	moduleLine := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			moduleLine = line
			break
		}
	}

	if moduleLine == "" {
		t.Fatal("go.mod should have a module line")
	}

	// Module name should not have spaces
	moduleName := strings.TrimPrefix(moduleLine, "module ")
	if strings.Contains(moduleName, " ") {
		t.Error("go.mod module name should not contain spaces")
	}
	if !strings.Contains(moduleName, "my-cool-project") {
		t.Errorf("Project name should be lowercased and hyphenated, got: %s", moduleName)
	}
}

func TestSanitizeModuleName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My Project", "my-project"},
		{"my_project", "my-project"},
		{"My Cool Project", "my-cool-project"},
		{"Project123", "project123"},
		{"test@project!", "testproject"},
		{"", ""},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeModuleName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeModuleName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
