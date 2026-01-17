package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// BackendAgent generates backend services, APIs, and models
// Supports Node.js (Express, Fastify, NestJS), Python (FastAPI, Django), Go (Gin, Echo)
type BackendAgent struct {
	llmClient llm.Client
}

func NewBackendAgent() *BackendAgent {
	return &BackendAgent{
		llmClient: llm.NewClient(),
	}
}

// ============================================================================
// Data Structures
// ============================================================================

// BackendCodeBundle represents generated backend code files
type BackendCodeBundle struct {
	Stack       string            `json:"stack"`        // "node-express", "python-fastapi", "go-gin"
	Files       map[string]string `json:"files"`        // filename -> content
	EntryPoint  string            `json:"entry_point"`  // Main file to run
	StartCmd    string            `json:"start_cmd"`    // Command to start server
	InstallCmd  string            `json:"install_cmd"`  // Command to install deps
	TestCmd     string            `json:"test_cmd"`     // Command to run tests
	BuildCmd    string            `json:"build_cmd"`    // Command to build (if applicable)
}

// BackendGenerateInput contains all info needed to generate a backend
type BackendGenerateInput struct {
	ProjectName    string          `json:"project_name"`
	Requirements   string          `json:"requirements"`
	TechStack      TechStack       `json:"tech_stack"`
	DatabaseSchema *DatabaseSchema `json:"database_schema,omitempty"`
	APISpec        *APISpec        `json:"api_spec,omitempty"`
}

// BackendRefineInput for refining existing backend code
type BackendRefineInput struct {
	CurrentCode        map[string]string `json:"current_code"`
	TechStack          TechStack         `json:"tech_stack"`
	Feedback           string            `json:"feedback"`
	ValidationErrors   string            `json:"validation_errors,omitempty"`
}

// ============================================================================
// Activity Methods
// ============================================================================

// GenerateBackend creates a complete backend service based on the tech stack
func (a *BackendAgent) GenerateBackend(ctx context.Context, input map[string]interface{}) (*BackendCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req BackendGenerateInput
	json.Unmarshal(inputBytes, &req)

	// Determine which stack to generate
	stack := determineBackendStack(req.TechStack)

	switch stack {
	case "node-express", "node-fastify", "node-nestjs":
		return a.generateNodeBackend(ctx, req, stack)
	case "python-fastapi", "python-django", "python-flask":
		return a.generatePythonBackend(ctx, req, stack)
	case "go-gin", "go-echo", "go-fiber":
		return a.generateGoBackend(ctx, req, stack)
	default:
		// Default to Node.js Express
		return a.generateNodeBackend(ctx, req, "node-express")
	}
}

// RefineBackend modifies existing backend code based on feedback
func (a *BackendAgent) RefineBackend(ctx context.Context, input map[string]interface{}) (*BackendCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req BackendRefineInput
	json.Unmarshal(inputBytes, &req)

	stack := determineBackendStack(req.TechStack)

	sysPrompt := fmt.Sprintf(`You are a Senior Backend Developer specializing in %s.
Modify the existing code based on the feedback provided.

Output ONLY valid JSON with this structure:
{
  "stack": "%s",
  "files": {
    "path/to/file.ext": "file content"
  },
  "entry_point": "main file",
  "start_cmd": "command to start",
  "install_cmd": "command to install deps",
  "test_cmd": "command to run tests"
}

Rules:
1. Only include files that need to be changed
2. Maintain existing code style and patterns
3. Fix all validation errors first
4. Keep the same project structure
5. Preserve existing functionality unless explicitly asked to change`, stack, stack)

	codeContext := "EXISTING CODE:\n"
	for name, content := range req.CurrentCode {
		codeContext += fmt.Sprintf("File: %s\n```\n%s\n```\n\n", name, content)
	}

	prompt := codeContext
	if req.ValidationErrors != "" {
		prompt += fmt.Sprintf("\nVALIDATION ERRORS TO FIX:\n%s\n", req.ValidationErrors)
	}
	prompt += fmt.Sprintf("\nFEEDBACK/CHANGES REQUESTED:\n%s\n\nGenerate the modified files.", req.Feedback)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("backend refine LLM call failed: %w", err)
	}

	var bundle BackendCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse backend code: %w", err)
	}

	// Merge with existing code
	for k, v := range req.CurrentCode {
		if _, exists := bundle.Files[k]; !exists {
			bundle.Files[k] = v
		}
	}

	return &bundle, nil
}

// GenerateModels creates ORM models from database schema
func (a *BackendAgent) GenerateModels(ctx context.Context, input map[string]interface{}) (*BackendCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req struct {
		TechStack      TechStack       `json:"tech_stack"`
		DatabaseSchema *DatabaseSchema `json:"database_schema"`
	}
	json.Unmarshal(inputBytes, &req)

	if req.DatabaseSchema == nil {
		return nil, fmt.Errorf("database schema is required")
	}

	stack := determineBackendStack(req.TechStack)
	orm := req.TechStack.Backend.ORM

	var sysPrompt string
	switch {
	case strings.Contains(stack, "node"):
		sysPrompt = a.getNodeModelPrompt(orm)
	case strings.Contains(stack, "python"):
		sysPrompt = a.getPythonModelPrompt(orm)
	case strings.Contains(stack, "go"):
		sysPrompt = a.getGoModelPrompt(orm)
	default:
		sysPrompt = a.getNodeModelPrompt("prisma")
	}

	schemaBytes, _ := json.MarshalIndent(req.DatabaseSchema, "", "  ")
	prompt := fmt.Sprintf("DATABASE SCHEMA:\n%s\n\nGenerate the ORM models.", string(schemaBytes))

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("model generation LLM call failed: %w", err)
	}

	var bundle BackendCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse models: %w", err)
	}

	bundle.Stack = stack
	return &bundle, nil
}

// GenerateAPIRoutes creates route handlers from API spec
func (a *BackendAgent) GenerateAPIRoutes(ctx context.Context, input map[string]interface{}) (*BackendCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req struct {
		TechStack TechStack `json:"tech_stack"`
		APISpec   *APISpec  `json:"api_spec"`
	}
	json.Unmarshal(inputBytes, &req)

	if req.APISpec == nil {
		return nil, fmt.Errorf("API spec is required")
	}

	stack := determineBackendStack(req.TechStack)

	var sysPrompt string
	switch {
	case strings.Contains(stack, "node-express"):
		sysPrompt = a.getExpressRoutePrompt()
	case strings.Contains(stack, "node-fastify"):
		sysPrompt = a.getFastifyRoutePrompt()
	case strings.Contains(stack, "python-fastapi"):
		sysPrompt = a.getFastAPIRoutePrompt()
	case strings.Contains(stack, "go-gin"):
		sysPrompt = a.getGinRoutePrompt()
	default:
		sysPrompt = a.getExpressRoutePrompt()
	}

	specBytes, _ := json.MarshalIndent(req.APISpec, "", "  ")
	prompt := fmt.Sprintf("API SPECIFICATION:\n%s\n\nGenerate the route handlers.", string(specBytes))

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("route generation LLM call failed: %w", err)
	}

	var bundle BackendCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse routes: %w", err)
	}

	bundle.Stack = stack
	return &bundle, nil
}

// ============================================================================
// Stack-Specific Generators
// ============================================================================

func (a *BackendAgent) generateNodeBackend(ctx context.Context, req BackendGenerateInput, stack string) (*BackendCodeBundle, error) {
	framework := "Express"
	if stack == "node-fastify" {
		framework = "Fastify"
	} else if stack == "node-nestjs" {
		framework = "NestJS"
	}

	sysPrompt := fmt.Sprintf(`You are a Senior Node.js Developer using %s and TypeScript.
Generate a complete, production-ready backend service.

Output ONLY valid JSON with this structure:
{
  "stack": "%s",
  "files": {
    "src/index.ts": "main entry point",
    "src/routes/index.ts": "route definitions",
    "src/controllers/[resource].controller.ts": "controller logic",
    "src/models/[model].ts": "data models",
    "src/middleware/auth.ts": "authentication middleware",
    "src/middleware/errorHandler.ts": "error handling",
    "src/config/index.ts": "configuration",
    "src/utils/logger.ts": "logging utility",
    "package.json": "dependencies",
    "tsconfig.json": "TypeScript config",
    ".env.example": "environment variables template"
  },
  "entry_point": "src/index.ts",
  "start_cmd": "npm run dev",
  "install_cmd": "npm install",
  "test_cmd": "npm test",
  "build_cmd": "npm run build"
}

Requirements:
1. Use TypeScript with strict mode
2. Implement proper error handling with custom error classes
3. Add request validation (using zod or joi)
4. Include authentication middleware (JWT)
5. Add logging (using pino or winston)
6. Use async/await with proper error catching
7. Include health check endpoint
8. Add CORS configuration
9. Use environment variables for config
10. Follow REST best practices
11. Include proper TypeScript types for all functions`, framework, stack)

	prompt := a.buildBackendPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Node backend LLM call failed: %w", err)
	}

	var bundle BackendCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Node backend: %w", err)
	}

	// Ensure required files exist with defaults
	bundle = a.ensureNodeDefaults(bundle, stack)

	return &bundle, nil
}

func (a *BackendAgent) generatePythonBackend(ctx context.Context, req BackendGenerateInput, stack string) (*BackendCodeBundle, error) {
	framework := "FastAPI"
	if stack == "python-django" {
		framework = "Django"
	} else if stack == "python-flask" {
		framework = "Flask"
	}

	sysPrompt := fmt.Sprintf(`You are a Senior Python Developer using %s.
Generate a complete, production-ready backend service.

Output ONLY valid JSON with this structure:
{
  "stack": "%s",
  "files": {
    "app/main.py": "main entry point",
    "app/api/routes.py": "route definitions",
    "app/api/endpoints/[resource].py": "endpoint handlers",
    "app/models/[model].py": "SQLAlchemy/Pydantic models",
    "app/core/config.py": "configuration",
    "app/core/security.py": "authentication/security",
    "app/db/session.py": "database session",
    "app/schemas/[resource].py": "Pydantic schemas",
    "requirements.txt": "dependencies",
    "pyproject.toml": "project config",
    ".env.example": "environment variables template"
  },
  "entry_point": "app/main.py",
  "start_cmd": "uvicorn app.main:app --reload",
  "install_cmd": "pip install -r requirements.txt",
  "test_cmd": "pytest",
  "build_cmd": ""
}

Requirements:
1. Use Python 3.11+ features
2. Implement proper exception handling
3. Add request validation with Pydantic
4. Include JWT authentication
5. Add logging with structlog or loguru
6. Use async/await where beneficial
7. Include health check endpoint
8. Add CORS middleware
9. Use environment variables (python-dotenv)
10. Follow Python best practices (PEP 8)
11. Include type hints for all functions`, framework, stack)

	prompt := a.buildBackendPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Python backend LLM call failed: %w", err)
	}

	var bundle BackendCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Python backend: %w", err)
	}

	// Ensure required files exist with defaults
	bundle = a.ensurePythonDefaults(bundle, stack)

	return &bundle, nil
}

func (a *BackendAgent) generateGoBackend(ctx context.Context, req BackendGenerateInput, stack string) (*BackendCodeBundle, error) {
	framework := "Gin"
	if stack == "go-echo" {
		framework = "Echo"
	} else if stack == "go-fiber" {
		framework = "Fiber"
	}

	sysPrompt := fmt.Sprintf(`You are a Senior Go Developer using %s framework.
Generate a complete, production-ready backend service.

Output ONLY valid JSON with this structure:
{
  "stack": "%s",
  "files": {
    "cmd/server/main.go": "main entry point",
    "internal/api/routes.go": "route definitions",
    "internal/api/handlers/[resource].go": "request handlers",
    "internal/models/[model].go": "data models",
    "internal/middleware/auth.go": "authentication middleware",
    "internal/middleware/logger.go": "logging middleware",
    "internal/config/config.go": "configuration",
    "internal/db/db.go": "database connection",
    "pkg/utils/response.go": "response utilities",
    "go.mod": "module definition",
    ".env.example": "environment variables template"
  },
  "entry_point": "cmd/server/main.go",
  "start_cmd": "go run cmd/server/main.go",
  "install_cmd": "go mod download",
  "test_cmd": "go test ./...",
  "build_cmd": "go build -o bin/server cmd/server/main.go"
}

Requirements:
1. Follow Go project layout standards
2. Implement proper error handling (no panic)
3. Add request validation
4. Include JWT authentication middleware
5. Add structured logging (zerolog or zap)
6. Use context for cancellation
7. Include health check endpoint
8. Add CORS middleware
9. Use environment variables (godotenv or viper)
10. Follow Go best practices (effective Go)
11. Include proper documentation comments`, framework, stack)

	prompt := a.buildBackendPrompt(req)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Go backend LLM call failed: %w", err)
	}

	var bundle BackendCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Go backend: %w", err)
	}

	// Ensure required files exist with defaults
	bundle = a.ensureGoDefaults(bundle, stack, req.ProjectName)

	return &bundle, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func determineBackendStack(stack TechStack) string {
	lang := strings.ToLower(stack.Backend.Language)
	framework := strings.ToLower(stack.Backend.Framework)

	switch {
	case strings.Contains(lang, "typescript") || strings.Contains(lang, "javascript") || strings.Contains(lang, "node"):
		switch {
		case strings.Contains(framework, "fastify"):
			return "node-fastify"
		case strings.Contains(framework, "nest"):
			return "node-nestjs"
		default:
			return "node-express"
		}
	case strings.Contains(lang, "python"):
		switch {
		case strings.Contains(framework, "django"):
			return "python-django"
		case strings.Contains(framework, "flask"):
			return "python-flask"
		default:
			return "python-fastapi"
		}
	case strings.Contains(lang, "go"):
		switch {
		case strings.Contains(framework, "echo"):
			return "go-echo"
		case strings.Contains(framework, "fiber"):
			return "go-fiber"
		default:
			return "go-gin"
		}
	default:
		return "node-express"
	}
}

func (a *BackendAgent) buildBackendPrompt(req BackendGenerateInput) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("PROJECT: %s\n\n", req.ProjectName))
	sb.WriteString(fmt.Sprintf("REQUIREMENTS:\n%s\n\n", req.Requirements))

	if req.DatabaseSchema != nil {
		schemaBytes, _ := json.MarshalIndent(req.DatabaseSchema, "", "  ")
		sb.WriteString(fmt.Sprintf("DATABASE SCHEMA:\n%s\n\n", string(schemaBytes)))
	}

	if req.APISpec != nil {
		specBytes, _ := json.MarshalIndent(req.APISpec, "", "  ")
		sb.WriteString(fmt.Sprintf("API SPECIFICATION:\n%s\n\n", string(specBytes)))
	}

	sb.WriteString("Generate the complete backend service.")
	return sb.String()
}

// ============================================================================
// Model Generation Prompts
// ============================================================================

func (a *BackendAgent) getNodeModelPrompt(orm string) string {
	switch strings.ToLower(orm) {
	case "prisma":
		return `You are a Node.js Developer generating Prisma models.
Output JSON with files: { "prisma/schema.prisma": "prisma schema content" }
Follow Prisma schema conventions with proper relations.`
	case "typeorm":
		return `You are a Node.js Developer generating TypeORM entities.
Output JSON with files for each entity in src/entities/.
Use decorators and proper TypeScript types.`
	default:
		return `You are a Node.js Developer generating Prisma models.
Output JSON with files: { "prisma/schema.prisma": "prisma schema content" }`
	}
}

func (a *BackendAgent) getPythonModelPrompt(orm string) string {
	switch strings.ToLower(orm) {
	case "sqlalchemy":
		return `You are a Python Developer generating SQLAlchemy models.
Output JSON with files for each model in app/models/.
Use SQLAlchemy 2.0 style with proper type hints.`
	case "django":
		return `You are a Python Developer generating Django models.
Output JSON with files for each app's models.py.
Use Django model conventions with proper Meta classes.`
	default:
		return `You are a Python Developer generating SQLAlchemy models.
Output JSON with files for each model in app/models/.
Use SQLAlchemy 2.0 style with proper type hints.`
	}
}

func (a *BackendAgent) getGoModelPrompt(orm string) string {
	switch strings.ToLower(orm) {
	case "gorm":
		return `You are a Go Developer generating GORM models.
Output JSON with files for each model in internal/models/.
Use GORM conventions with proper struct tags.`
	case "sqlx":
		return `You are a Go Developer generating sqlx-compatible structs.
Output JSON with files for each model in internal/models/.
Use db struct tags for column mapping.`
	default:
		return `You are a Go Developer generating GORM models.
Output JSON with files for each model in internal/models/.
Use GORM conventions with proper struct tags.`
	}
}

// ============================================================================
// Route Generation Prompts
// ============================================================================

func (a *BackendAgent) getExpressRoutePrompt() string {
	return `You are a Node.js Developer generating Express.js routes.
Output JSON with route files in src/routes/ and controllers in src/controllers/.
Use async handlers with proper error handling.
Include validation middleware using zod or express-validator.`
}

func (a *BackendAgent) getFastifyRoutePrompt() string {
	return `You are a Node.js Developer generating Fastify routes.
Output JSON with route files using Fastify's schema validation.
Use async handlers with proper error handling.
Include JSON schema for request/response validation.`
}

func (a *BackendAgent) getFastAPIRoutePrompt() string {
	return `You are a Python Developer generating FastAPI routes.
Output JSON with router files in app/api/endpoints/.
Use Pydantic for request/response models.
Include proper dependency injection and exception handling.`
}

func (a *BackendAgent) getGinRoutePrompt() string {
	return `You are a Go Developer generating Gin routes.
Output JSON with handler files in internal/api/handlers/.
Use proper context handling and error responses.
Include request binding and validation.`
}

// ============================================================================
// Default File Generators
// ============================================================================

func (a *BackendAgent) ensureNodeDefaults(bundle BackendCodeBundle, stack string) BackendCodeBundle {
	if bundle.Files == nil {
		bundle.Files = make(map[string]string)
	}

	bundle.Stack = stack
	bundle.EntryPoint = "src/index.ts"
	bundle.StartCmd = "npm run dev"
	bundle.InstallCmd = "npm install"
	bundle.TestCmd = "npm test"
	bundle.BuildCmd = "npm run build"

	// Ensure package.json exists
	if _, exists := bundle.Files["package.json"]; !exists {
		bundle.Files["package.json"] = a.getDefaultNodePackageJSON()
	}

	// Ensure tsconfig.json exists
	if _, exists := bundle.Files["tsconfig.json"]; !exists {
		bundle.Files["tsconfig.json"] = a.getDefaultTSConfig()
	}

	return bundle
}

func (a *BackendAgent) ensurePythonDefaults(bundle BackendCodeBundle, stack string) BackendCodeBundle {
	if bundle.Files == nil {
		bundle.Files = make(map[string]string)
	}

	bundle.Stack = stack
	bundle.EntryPoint = "app/main.py"
	bundle.StartCmd = "uvicorn app.main:app --reload --host 0.0.0.0 --port 8000"
	bundle.InstallCmd = "pip install -r requirements.txt"
	bundle.TestCmd = "pytest"
	bundle.BuildCmd = ""

	// Ensure requirements.txt exists
	if _, exists := bundle.Files["requirements.txt"]; !exists {
		bundle.Files["requirements.txt"] = a.getDefaultPythonRequirements(stack)
	}

	return bundle
}

func (a *BackendAgent) ensureGoDefaults(bundle BackendCodeBundle, stack string, projectName string) BackendCodeBundle {
	if bundle.Files == nil {
		bundle.Files = make(map[string]string)
	}

	bundle.Stack = stack
	bundle.EntryPoint = "cmd/server/main.go"
	bundle.StartCmd = "go run cmd/server/main.go"
	bundle.InstallCmd = "go mod download"
	bundle.TestCmd = "go test ./..."
	bundle.BuildCmd = "go build -o bin/server cmd/server/main.go"

	// Ensure go.mod exists
	if _, exists := bundle.Files["go.mod"]; !exists {
		moduleName := sanitizeModuleName(projectName)
		if moduleName == "" {
			moduleName = "myapp"
		}
		bundle.Files["go.mod"] = a.getDefaultGoMod(moduleName, stack)
	}

	return bundle
}

// sanitizeModuleName converts project name to valid Go module name
func sanitizeModuleName(name string) string {
	// Convert to lowercase
	result := strings.ToLower(name)
	// Replace spaces with hyphens
	result = strings.ReplaceAll(result, " ", "-")
	// Replace underscores with hyphens
	result = strings.ReplaceAll(result, "_", "-")
	// Remove any characters that aren't alphanumeric or hyphens
	var cleaned strings.Builder
	for _, r := range result {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned.WriteRune(r)
		}
	}
	return cleaned.String()
}

func (a *BackendAgent) getDefaultNodePackageJSON() string {
	return `{
  "name": "backend-service",
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "tsx watch src/index.ts",
    "build": "tsc",
    "start": "node dist/index.js",
    "test": "vitest run",
    "lint": "eslint src/"
  },
  "dependencies": {
    "express": "^4.18.2",
    "cors": "^2.8.5",
    "helmet": "^7.1.0",
    "jsonwebtoken": "^9.0.2",
    "bcryptjs": "^2.4.3",
    "zod": "^3.22.4",
    "pino": "^8.17.2",
    "dotenv": "^16.3.1"
  },
  "devDependencies": {
    "@types/express": "^4.17.21",
    "@types/cors": "^2.8.17",
    "@types/jsonwebtoken": "^9.0.5",
    "@types/bcryptjs": "^2.4.6",
    "@types/node": "^20.10.6",
    "typescript": "^5.3.3",
    "tsx": "^4.7.0",
    "vitest": "^1.1.3",
    "eslint": "^8.56.0",
    "@typescript-eslint/eslint-plugin": "^6.18.1",
    "@typescript-eslint/parser": "^6.18.1"
  }
}`
}

func (a *BackendAgent) getDefaultTSConfig() string {
	return `{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "lib": ["ES2022"],
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}`
}

func (a *BackendAgent) getDefaultPythonRequirements(stack string) string {
	base := `fastapi>=0.109.0
uvicorn[standard]>=0.27.0
pydantic>=2.5.3
pydantic-settings>=2.1.0
python-jose[cryptography]>=3.3.0
passlib[bcrypt]>=1.7.4
sqlalchemy>=2.0.25
alembic>=1.13.1
asyncpg>=0.29.0
python-dotenv>=1.0.0
structlog>=24.1.0
httpx>=0.26.0
pytest>=7.4.4
pytest-asyncio>=0.23.3
`
	if stack == "python-django" {
		return `django>=5.0
djangorestframework>=3.14.0
django-cors-headers>=4.3.1
djangorestframework-simplejwt>=5.3.1
psycopg2-binary>=2.9.9
python-dotenv>=1.0.0
pytest-django>=4.7.0
`
	}
	return base
}

func (a *BackendAgent) getDefaultGoMod(moduleName string, stack string) string {
	framework := "github.com/gin-gonic/gin"
	if stack == "go-echo" {
		framework = "github.com/labstack/echo/v4"
	} else if stack == "go-fiber" {
		framework = "github.com/gofiber/fiber/v2"
	}

	return fmt.Sprintf(`module %s

go 1.21

require (
	%s v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/joho/godotenv v1.5.1
	github.com/rs/zerolog v1.31.0
	gorm.io/gorm v1.25.5
	gorm.io/driver/postgres v1.5.4
)
`, moduleName, framework)
}
