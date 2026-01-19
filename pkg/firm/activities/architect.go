package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// ArchitectAgent designs system architecture, selects tech stacks, and creates schemas
// This is the foundation agent that informs all other agents' work
type ArchitectAgent struct {
	llmClient llm.Client
}

func NewArchitectAgent() *ArchitectAgent {
	return &ArchitectAgent{
		llmClient: llm.NewClient(),
	}
}

// ============================================================================
// Data Structures
// ============================================================================

// SystemDesign represents the complete architecture of a system
type SystemDesign struct {
	ProjectName    string           `json:"project_name"`
	Overview       string           `json:"overview"`
	Components     []Component      `json:"components"`
	DataFlow       []DataFlowEdge   `json:"data_flow"`
	TechStack      TechStack        `json:"tech_stack"`
	DatabaseSchema *DatabaseSchema  `json:"database_schema,omitempty"`
	APISpec        *APISpec         `json:"api_spec,omitempty"`
	Diagram        string           `json:"diagram,omitempty"` // Mermaid diagram
}

// Component represents a system component (service, database, etc.)
type Component struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "frontend", "backend", "database", "cache", "queue", "external"
	Description string   `json:"description"`
	Technology  string   `json:"technology"`
	Ports       []int    `json:"ports,omitempty"`
	DependsOn   []string `json:"depends_on,omitempty"`
}

// DataFlowEdge represents data flow between components
type DataFlowEdge struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Protocol    string `json:"protocol"` // "HTTP", "gRPC", "WebSocket", "SQL", "Redis"
	Description string `json:"description"`
}

// TechStack represents the selected technologies for the project
type TechStack struct {
	Frontend     FrontendStack     `json:"frontend"`
	Backend      BackendStack      `json:"backend"`
	Database     DatabaseStack     `json:"database"`
	Infrastructure InfraStack      `json:"infrastructure"`
	Rationale    string            `json:"rationale"`
}

type FrontendStack struct {
	Framework   string   `json:"framework"`    // "React", "Vue", "Angular", "Next.js", "None"
	Language    string   `json:"language"`     // "TypeScript", "JavaScript"
	Styling     string   `json:"styling"`      // "Tailwind", "CSS Modules", "Styled Components"
	StateManagement string `json:"state_management,omitempty"` // "Redux", "Zustand", "Context"
	BuildTool   string   `json:"build_tool"`   // "Vite", "Webpack", "Turbopack"
}

type BackendStack struct {
	Framework   string   `json:"framework"`    // "Express", "FastAPI", "Gin", "Spring Boot", "NestJS"
	Language    string   `json:"language"`     // "TypeScript", "Python", "Go", "Java"
	Runtime     string   `json:"runtime"`      // "Node.js", "Python 3.11", "Go 1.21"
	ORM         string   `json:"orm,omitempty"` // "Prisma", "SQLAlchemy", "GORM", "TypeORM"
}

type DatabaseStack struct {
	Primary     string   `json:"primary"`      // "PostgreSQL", "MySQL", "MongoDB"
	Cache       string   `json:"cache,omitempty"` // "Redis", "Memcached", ""
	Search      string   `json:"search,omitempty"` // "Elasticsearch", "Meilisearch", ""
	Queue       string   `json:"queue,omitempty"` // "RabbitMQ", "Redis", "SQS", ""
}

type InfraStack struct {
	Cloud       string   `json:"cloud"`        // "AWS", "Azure", "GCP", "Self-hosted"
	Container   string   `json:"container"`    // "Docker", "Podman"
	Orchestration string `json:"orchestration"` // "Kubernetes", "ECS", "Docker Compose"
	CI          string   `json:"ci"`           // "GitHub Actions", "GitLab CI", "Jenkins"
	IaC         string   `json:"iac"`          // "Terraform", "Pulumi", "CloudFormation"
}

// DatabaseSchema represents the database design
type DatabaseSchema struct {
	DatabaseType string        `json:"database_type"` // "PostgreSQL", "MongoDB"
	Tables       []TableSchema `json:"tables"`
	Indexes      []IndexDef    `json:"indexes,omitempty"`
	Constraints  []Constraint  `json:"constraints,omitempty"`
}

type TableSchema struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Columns     []ColumnDef   `json:"columns"`
	PrimaryKey  []string      `json:"primary_key"`
	Timestamps  bool          `json:"timestamps"` // created_at, updated_at
}

type ColumnDef struct {
	Name       string `json:"name"`
	Type       string `json:"type"`       // "uuid", "varchar(255)", "text", "integer", "boolean", "timestamp", "jsonb"
	Nullable   bool   `json:"nullable"`
	Default    string `json:"default,omitempty"`
	References string `json:"references,omitempty"` // "table_name.column"
}

type IndexDef struct {
	Name    string   `json:"name"`
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

type Constraint struct {
	Name       string `json:"name"`
	Type       string `json:"type"` // "foreign_key", "unique", "check"
	Table      string `json:"table"`
	Definition string `json:"definition"`
}

// APISpec represents the API design
type APISpec struct {
	Type      string        `json:"type"` // "REST", "GraphQL", "gRPC"
	Version   string        `json:"version"`
	BasePath  string        `json:"base_path"`
	Endpoints []APIEndpoint `json:"endpoints,omitempty"`
	Types     []GraphQLType `json:"types,omitempty"` // For GraphQL
}

type APIEndpoint struct {
	Method      string            `json:"method"` // GET, POST, PUT, DELETE, PATCH
	Path        string            `json:"path"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	RequestBody *RequestBody      `json:"request_body,omitempty"`
	Parameters  []Parameter       `json:"parameters,omitempty"`
	Responses   map[string]Response `json:"responses"`
	Auth        string            `json:"auth,omitempty"` // "bearer", "api_key", "none"
}

type RequestBody struct {
	ContentType string                 `json:"content_type"`
	Schema      map[string]interface{} `json:"schema"`
	Required    bool                   `json:"required"`
}

type Parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"` // "path", "query", "header"
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type Response struct {
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
}

type GraphQLType struct {
	Name   string         `json:"name"`
	Kind   string         `json:"kind"` // "type", "input", "enum", "interface"
	Fields []GraphQLField `json:"fields"`
}

type GraphQLField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

// ============================================================================
// Input/Output Types for Activities
// ============================================================================

type AnalyzeRequirementsInput struct {
	ProjectName  string `json:"project_name"`
	Requirements string `json:"requirements"` // The chat history or spec
}

type SelectTechStackInput struct {
	Requirements string   `json:"requirements"`
	Constraints  []string `json:"constraints,omitempty"` // e.g., "must use Python", "AWS only"
	Preferences  []string `json:"preferences,omitempty"` // e.g., "prefer TypeScript"
}

type DesignDatabaseInput struct {
	Requirements   string    `json:"requirements"`
	TechStack      TechStack `json:"tech_stack"`
	Entities       []string  `json:"entities,omitempty"` // Optional: specific entities to model
}

type DesignAPIInput struct {
	Requirements   string         `json:"requirements"`
	TechStack      TechStack      `json:"tech_stack"`
	DatabaseSchema *DatabaseSchema `json:"database_schema,omitempty"`
	APIType        string         `json:"api_type"` // "REST" or "GraphQL"
}

// ============================================================================
// Activity Methods
// ============================================================================

// AnalyzeRequirements parses requirements and produces a high-level system design
func (a *ArchitectAgent) AnalyzeRequirements(ctx context.Context, input map[string]interface{}) (*SystemDesign, error) {
	inputBytes, _ := json.Marshal(input)
	var req AnalyzeRequirementsInput
	json.Unmarshal(inputBytes, &req)

	sysPrompt := `You are a Senior Solution Architect designing software systems.
Analyze the requirements and produce a comprehensive system design.

Output ONLY valid JSON with this structure:
{
  "project_name": "string",
  "overview": "2-3 sentence description of the system",
  "components": [
    {
      "name": "component name",
      "type": "frontend|backend|database|cache|queue|external",
      "description": "what this component does",
      "technology": "specific tech (e.g., React, PostgreSQL)",
      "depends_on": ["other component names"]
    }
  ],
  "data_flow": [
    {
      "from": "component name",
      "to": "component name",
      "protocol": "HTTP|gRPC|WebSocket|SQL|Redis",
      "description": "what data flows"
    }
  ],
  "tech_stack": {
    "frontend": {
      "framework": "React|Vue|Angular|Next.js|None",
      "language": "TypeScript|JavaScript",
      "styling": "Tailwind|CSS Modules|Styled Components",
      "build_tool": "Vite|Webpack"
    },
    "backend": {
      "framework": "Express|FastAPI|Gin|NestJS",
      "language": "TypeScript|Python|Go",
      "runtime": "Node.js|Python 3.11|Go 1.21",
      "orm": "Prisma|SQLAlchemy|GORM|TypeORM"
    },
    "database": {
      "primary": "PostgreSQL|MySQL|MongoDB",
      "cache": "Redis|None",
      "queue": "RabbitMQ|Redis|None"
    },
    "infrastructure": {
      "cloud": "AWS|Azure|GCP|Self-hosted",
      "container": "Docker",
      "orchestration": "Kubernetes|Docker Compose",
      "ci": "GitHub Actions|GitLab CI",
      "iac": "Terraform|Pulumi"
    },
    "rationale": "Why these technologies were chosen"
  }
}

Guidelines:
1. Keep it simple - don't over-engineer
2. Choose mature, well-supported technologies
3. Consider the team's likely expertise (prefer popular stacks)
4. For small projects, use simpler architectures (monolith over microservices)
5. Always include rationale for tech choices`

	// ISS-024: Sanitize user input to prevent prompt injection
	prompt := fmt.Sprintf("Project: %s\n\n%s\n\nDesign the system architecture.",
		req.ProjectName, SanitizeUserInput("user-requirements", req.Requirements))

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("architect LLM call failed: %w", err)
	}

	var design SystemDesign
	if err := extractJSON(resp.Response, &design); err != nil {
		return nil, fmt.Errorf("failed to parse system design: %w", err)
	}

	// Generate Mermaid diagram
	design.Diagram = a.generateMermaidDiagram(&design)

	return &design, nil
}

// SelectTechStack recommends a technology stack based on requirements
func (a *ArchitectAgent) SelectTechStack(ctx context.Context, input map[string]interface{}) (*TechStack, error) {
	inputBytes, _ := json.Marshal(input)
	var req SelectTechStackInput
	json.Unmarshal(inputBytes, &req)

	constraintsStr := ""
	if len(req.Constraints) > 0 {
		constraintsStr = "\n\nHard Constraints (MUST follow):\n- " + strings.Join(req.Constraints, "\n- ")
	}
	preferencesStr := ""
	if len(req.Preferences) > 0 {
		preferencesStr = "\n\nPreferences (follow if reasonable):\n- " + strings.Join(req.Preferences, "\n- ")
	}

	sysPrompt := `You are a Senior Solution Architect selecting technology stacks.
Based on the requirements, select the most appropriate technologies.

Output ONLY valid JSON with this structure:
{
  "frontend": {
    "framework": "React|Vue|Angular|Next.js|None",
    "language": "TypeScript|JavaScript",
    "styling": "Tailwind|CSS Modules|Styled Components",
    "state_management": "Redux|Zustand|Context|None",
    "build_tool": "Vite|Webpack|Turbopack"
  },
  "backend": {
    "framework": "Express|FastAPI|Gin|NestJS|Django|Spring Boot",
    "language": "TypeScript|Python|Go|Java",
    "runtime": "Node.js 20|Python 3.11|Go 1.21|Java 21",
    "orm": "Prisma|SQLAlchemy|GORM|TypeORM|Drizzle"
  },
  "database": {
    "primary": "PostgreSQL|MySQL|MongoDB|SQLite",
    "cache": "Redis|Memcached|None",
    "search": "Elasticsearch|Meilisearch|None",
    "queue": "RabbitMQ|Redis|SQS|None"
  },
  "infrastructure": {
    "cloud": "AWS|Azure|GCP|Self-hosted",
    "container": "Docker|Podman",
    "orchestration": "Kubernetes|ECS|Docker Compose",
    "ci": "GitHub Actions|GitLab CI|Jenkins|CircleCI",
    "iac": "Terraform|Pulumi|CloudFormation|None"
  },
  "rationale": "Detailed explanation of why each technology was chosen"
}

Selection Guidelines:
1. Match complexity to project size (don't suggest K8s for a todo app)
2. Prioritize developer productivity and ecosystem maturity
3. Consider deployment simplicity
4. TypeScript + React + Node.js is a good default for web apps
5. Python + FastAPI is great for API-heavy or ML projects
6. Go is excellent for high-performance services
7. PostgreSQL is the safe default database choice
8. Redis is useful for caching and simple queues
9. Docker Compose for development, K8s only for production scale`

	prompt := fmt.Sprintf("Requirements:\n%s%s%s\n\nSelect the optimal technology stack.",
		req.Requirements, constraintsStr, preferencesStr)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("tech stack LLM call failed: %w", err)
	}

	var stack TechStack
	if err := extractJSON(resp.Response, &stack); err != nil {
		return nil, fmt.Errorf("failed to parse tech stack: %w", err)
	}

	return &stack, nil
}

// DesignDatabase creates a database schema from requirements
func (a *ArchitectAgent) DesignDatabase(ctx context.Context, input map[string]interface{}) (*DatabaseSchema, error) {
	inputBytes, _ := json.Marshal(input)
	var req DesignDatabaseInput
	json.Unmarshal(inputBytes, &req)

	dbType := req.TechStack.Database.Primary
	if dbType == "" {
		dbType = "PostgreSQL"
	}

	entitiesHint := ""
	if len(req.Entities) > 0 {
		entitiesHint = "\n\nKey entities to model: " + strings.Join(req.Entities, ", ")
	}

	sysPrompt := fmt.Sprintf(`You are a Database Architect designing schemas for %s.
Design a normalized database schema based on the requirements.

Output ONLY valid JSON with this structure:
{
  "database_type": "%s",
  "tables": [
    {
      "name": "table_name (snake_case, plural)",
      "description": "what this table stores",
      "columns": [
        {
          "name": "column_name",
          "type": "uuid|varchar(255)|text|integer|bigint|boolean|timestamp|timestamptz|jsonb|decimal(10,2)",
          "nullable": false,
          "default": "optional default value",
          "references": "other_table.column (for foreign keys)"
        }
      ],
      "primary_key": ["id"],
      "timestamps": true
    }
  ],
  "indexes": [
    {
      "name": "idx_table_column",
      "table": "table_name",
      "columns": ["column1", "column2"],
      "unique": false
    }
  ],
  "constraints": [
    {
      "name": "fk_table_reference",
      "type": "foreign_key|unique|check",
      "table": "table_name",
      "definition": "FOREIGN KEY (column) REFERENCES other_table(id)"
    }
  ]
}

Database Design Guidelines:
1. Use UUIDs for primary keys (better for distributed systems)
2. Always include created_at and updated_at timestamps
3. Use snake_case for table and column names
4. Create indexes for frequently queried columns
5. Create indexes for foreign key columns
6. Use appropriate column types (don't use TEXT for short strings)
7. Add foreign key constraints for data integrity
8. Consider soft deletes (deleted_at) for important tables
9. Normalize to 3NF but denormalize for read-heavy queries
10. Use JSONB sparingly - prefer proper columns`, dbType, dbType)

	prompt := fmt.Sprintf("Requirements:\n%s%s\n\nDesign the database schema.",
		req.Requirements, entitiesHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("database design LLM call failed: %w", err)
	}

	var schema DatabaseSchema
	if err := extractJSON(resp.Response, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse database schema: %w", err)
	}

	return &schema, nil
}

// DesignAPI creates an API specification from requirements
func (a *ArchitectAgent) DesignAPI(ctx context.Context, input map[string]interface{}) (*APISpec, error) {
	inputBytes, _ := json.Marshal(input)
	var req DesignAPIInput
	json.Unmarshal(inputBytes, &req)

	apiType := req.APIType
	if apiType == "" {
		apiType = "REST"
	}

	schemaContext := ""
	if req.DatabaseSchema != nil {
		schemaBytes, _ := json.MarshalIndent(req.DatabaseSchema.Tables, "", "  ")
		schemaContext = fmt.Sprintf("\n\nDatabase Tables:\n%s", string(schemaBytes))
	}

	var sysPrompt string
	if apiType == "REST" {
		sysPrompt = `You are an API Architect designing RESTful APIs.
Design a comprehensive REST API based on the requirements.

Output ONLY valid JSON with this structure:
{
  "type": "REST",
  "version": "v1",
  "base_path": "/api/v1",
  "endpoints": [
    {
      "method": "GET|POST|PUT|PATCH|DELETE",
      "path": "/resources or /resources/{id}",
      "summary": "Short description",
      "description": "Detailed description",
      "parameters": [
        {
          "name": "id",
          "in": "path|query|header",
          "type": "string|integer|boolean",
          "required": true
        }
      ],
      "request_body": {
        "content_type": "application/json",
        "schema": {"field": "type"},
        "required": true
      },
      "responses": {
        "200": {"description": "Success", "schema": {}},
        "400": {"description": "Bad Request"},
        "401": {"description": "Unauthorized"},
        "404": {"description": "Not Found"}
      },
      "auth": "bearer|api_key|none"
    }
  ]
}

REST API Design Guidelines:
1. Use plural nouns for resources (/users, /orders)
2. Use HTTP methods correctly (GET=read, POST=create, PUT/PATCH=update, DELETE=delete)
3. Use path parameters for resource IDs (/users/{id})
4. Use query parameters for filtering, sorting, pagination
5. Return appropriate status codes (200, 201, 400, 401, 403, 404, 500)
6. Include pagination for list endpoints (limit, offset or cursor)
7. Use consistent response formats
8. Version your API (/api/v1)
9. Require authentication for sensitive endpoints
10. Include proper error responses`
	} else {
		sysPrompt = `You are an API Architect designing GraphQL APIs.
Design a comprehensive GraphQL schema based on the requirements.

Output ONLY valid JSON with this structure:
{
  "type": "GraphQL",
  "version": "v1",
  "base_path": "/graphql",
  "types": [
    {
      "name": "User",
      "kind": "type|input|enum|interface",
      "fields": [
        {"name": "id", "type": "ID!", "nullable": false},
        {"name": "email", "type": "String!", "nullable": false},
        {"name": "posts", "type": "[Post!]!", "nullable": false}
      ]
    }
  ]
}

GraphQL Design Guidelines:
1. Use PascalCase for type names
2. Use camelCase for field names
3. Mark required fields with !
4. Create input types for mutations
5. Use connections pattern for pagination (edges, nodes, pageInfo)
6. Include proper error handling
7. Design mutations to be specific (createUser, updateUserEmail)
8. Use interfaces for shared fields
9. Keep queries flat where possible`
	}

	prompt := fmt.Sprintf("Requirements:\n%s%s\n\nDesign the %s API.",
		req.Requirements, schemaContext, apiType)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("API design LLM call failed: %w", err)
	}

	var spec APISpec
	if err := extractJSON(resp.Response, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse API spec: %w", err)
	}

	return &spec, nil
}

// GenerateFullDesign creates a complete system design including all components
func (a *ArchitectAgent) GenerateFullDesign(ctx context.Context, input map[string]interface{}) (*SystemDesign, error) {
	inputBytes, _ := json.Marshal(input)
	var req AnalyzeRequirementsInput
	json.Unmarshal(inputBytes, &req)

	// Step 1: Analyze requirements and create high-level design
	design, err := a.AnalyzeRequirements(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("requirements analysis failed: %w", err)
	}

	// Step 2: Design database schema
	dbInput := map[string]interface{}{
		"requirements": req.Requirements,
		"tech_stack":   design.TechStack,
	}
	schema, err := a.DesignDatabase(ctx, dbInput)
	if err != nil {
		// Non-fatal - continue without schema
		fmt.Printf("Warning: database design failed: %v\n", err)
	} else {
		design.DatabaseSchema = schema
	}

	// Step 3: Design API
	apiInput := map[string]interface{}{
		"requirements":    req.Requirements,
		"tech_stack":      design.TechStack,
		"database_schema": design.DatabaseSchema,
		"api_type":        "REST",
	}
	apiSpec, err := a.DesignAPI(ctx, apiInput)
	if err != nil {
		// Non-fatal - continue without API spec
		fmt.Printf("Warning: API design failed: %v\n", err)
	} else {
		design.APISpec = apiSpec
	}

	// Regenerate diagram with full context
	design.Diagram = a.generateMermaidDiagram(design)

	return design, nil
}

// generateMermaidDiagram creates a Mermaid diagram from the system design
func (a *ArchitectAgent) generateMermaidDiagram(design *SystemDesign) string {
	var sb strings.Builder
	sb.WriteString("```mermaid\nflowchart TB\n")

	// Group components by type
	frontends := []Component{}
	backends := []Component{}
	databases := []Component{}
	others := []Component{}

	for _, c := range design.Components {
		switch c.Type {
		case "frontend":
			frontends = append(frontends, c)
		case "backend":
			backends = append(backends, c)
		case "database", "cache", "queue":
			databases = append(databases, c)
		default:
			others = append(others, c)
		}
	}

	// Create subgraphs
	if len(frontends) > 0 {
		sb.WriteString("    subgraph Frontend\n")
		for _, c := range frontends {
			sb.WriteString(fmt.Sprintf("        %s[%s<br/>%s]\n", sanitizeMermaidID(c.Name), c.Name, c.Technology))
		}
		sb.WriteString("    end\n")
	}

	if len(backends) > 0 {
		sb.WriteString("    subgraph Backend\n")
		for _, c := range backends {
			sb.WriteString(fmt.Sprintf("        %s[%s<br/>%s]\n", sanitizeMermaidID(c.Name), c.Name, c.Technology))
		}
		sb.WriteString("    end\n")
	}

	if len(databases) > 0 {
		sb.WriteString("    subgraph Data\n")
		for _, c := range databases {
			shape := "(%s<br/>%s)" // Cylinder for databases
			if c.Type == "cache" || c.Type == "queue" {
				shape = "[%s<br/>%s]"
			}
			sb.WriteString(fmt.Sprintf("        %s"+shape+"\n", sanitizeMermaidID(c.Name), c.Name, c.Technology))
		}
		sb.WriteString("    end\n")
	}

	for _, c := range others {
		sb.WriteString(fmt.Sprintf("    %s{{%s<br/>%s}}\n", sanitizeMermaidID(c.Name), c.Name, c.Technology))
	}

	// Add data flow edges
	for _, flow := range design.DataFlow {
		sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n",
			sanitizeMermaidID(flow.From),
			flow.Protocol,
			sanitizeMermaidID(flow.To)))
	}

	sb.WriteString("```")
	return sb.String()
}

func sanitizeMermaidID(name string) string {
	// Replace spaces and special characters with underscores
	result := strings.ReplaceAll(name, " ", "_")
	result = strings.ReplaceAll(result, "-", "_")
	result = strings.ReplaceAll(result, ".", "_")
	return result
}
