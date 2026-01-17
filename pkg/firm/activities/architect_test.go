package activities

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSystemDesignStructure(t *testing.T) {
	design := SystemDesign{
		ProjectName: "Test Project",
		Overview:    "A test project for unit testing",
		Components: []Component{
			{
				Name:        "Frontend",
				Type:        "frontend",
				Description: "React web application",
				Technology:  "React",
				DependsOn:   []string{"Backend"},
			},
			{
				Name:        "Backend",
				Type:        "backend",
				Description: "REST API server",
				Technology:  "Node.js",
				Ports:       []int{3000},
			},
			{
				Name:        "Database",
				Type:        "database",
				Description: "Primary data store",
				Technology:  "PostgreSQL",
				Ports:       []int{5432},
			},
		},
		DataFlow: []DataFlowEdge{
			{From: "Frontend", To: "Backend", Protocol: "HTTP", Description: "API calls"},
			{From: "Backend", To: "Database", Protocol: "SQL", Description: "Data queries"},
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(design)
	if err != nil {
		t.Fatalf("Failed to marshal SystemDesign: %v", err)
	}

	var unmarshaled SystemDesign
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal SystemDesign: %v", err)
	}

	if unmarshaled.ProjectName != design.ProjectName {
		t.Errorf("ProjectName mismatch: got %s, want %s", unmarshaled.ProjectName, design.ProjectName)
	}

	if len(unmarshaled.Components) != 3 {
		t.Errorf("Components count mismatch: got %d, want 3", len(unmarshaled.Components))
	}

	if len(unmarshaled.DataFlow) != 2 {
		t.Errorf("DataFlow count mismatch: got %d, want 2", len(unmarshaled.DataFlow))
	}
}

func TestTechStackStructure(t *testing.T) {
	stack := TechStack{
		Frontend: FrontendStack{
			Framework:       "React",
			Language:        "TypeScript",
			Styling:         "Tailwind",
			StateManagement: "Zustand",
			BuildTool:       "Vite",
		},
		Backend: BackendStack{
			Framework: "Express",
			Language:  "TypeScript",
			Runtime:   "Node.js 20",
			ORM:       "Prisma",
		},
		Database: DatabaseStack{
			Primary: "PostgreSQL",
			Cache:   "Redis",
			Queue:   "RabbitMQ",
		},
		Infrastructure: InfraStack{
			Cloud:         "AWS",
			Container:     "Docker",
			Orchestration: "Kubernetes",
			CI:            "GitHub Actions",
			IaC:           "Terraform",
		},
		Rationale: "Modern, well-supported stack",
	}

	// Test JSON serialization
	data, err := json.Marshal(stack)
	if err != nil {
		t.Fatalf("Failed to marshal TechStack: %v", err)
	}

	var unmarshaled TechStack
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal TechStack: %v", err)
	}

	if unmarshaled.Frontend.Framework != "React" {
		t.Errorf("Frontend.Framework mismatch: got %s, want React", unmarshaled.Frontend.Framework)
	}

	if unmarshaled.Backend.ORM != "Prisma" {
		t.Errorf("Backend.ORM mismatch: got %s, want Prisma", unmarshaled.Backend.ORM)
	}

	if unmarshaled.Database.Primary != "PostgreSQL" {
		t.Errorf("Database.Primary mismatch: got %s, want PostgreSQL", unmarshaled.Database.Primary)
	}

	if unmarshaled.Infrastructure.Cloud != "AWS" {
		t.Errorf("Infrastructure.Cloud mismatch: got %s, want AWS", unmarshaled.Infrastructure.Cloud)
	}
}

func TestDatabaseSchemaStructure(t *testing.T) {
	schema := DatabaseSchema{
		DatabaseType: "PostgreSQL",
		Tables: []TableSchema{
			{
				Name:        "users",
				Description: "User accounts",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid", Nullable: false, Default: "gen_random_uuid()"},
					{Name: "email", Type: "varchar(255)", Nullable: false},
					{Name: "name", Type: "varchar(255)", Nullable: true},
					{Name: "created_at", Type: "timestamptz", Nullable: false, Default: "now()"},
				},
				PrimaryKey: []string{"id"},
				Timestamps: true,
			},
			{
				Name:        "posts",
				Description: "User posts",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid", Nullable: false},
					{Name: "user_id", Type: "uuid", Nullable: false, References: "users.id"},
					{Name: "title", Type: "varchar(255)", Nullable: false},
					{Name: "content", Type: "text", Nullable: true},
				},
				PrimaryKey: []string{"id"},
				Timestamps: true,
			},
		},
		Indexes: []IndexDef{
			{Name: "idx_users_email", Table: "users", Columns: []string{"email"}, Unique: true},
			{Name: "idx_posts_user_id", Table: "posts", Columns: []string{"user_id"}, Unique: false},
		},
		Constraints: []Constraint{
			{
				Name:       "fk_posts_user",
				Type:       "foreign_key",
				Table:      "posts",
				Definition: "FOREIGN KEY (user_id) REFERENCES users(id)",
			},
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Failed to marshal DatabaseSchema: %v", err)
	}

	var unmarshaled DatabaseSchema
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal DatabaseSchema: %v", err)
	}

	if len(unmarshaled.Tables) != 2 {
		t.Errorf("Tables count mismatch: got %d, want 2", len(unmarshaled.Tables))
	}

	if len(unmarshaled.Indexes) != 2 {
		t.Errorf("Indexes count mismatch: got %d, want 2", len(unmarshaled.Indexes))
	}

	// Check users table
	usersTable := unmarshaled.Tables[0]
	if usersTable.Name != "users" {
		t.Errorf("First table name mismatch: got %s, want users", usersTable.Name)
	}

	if len(usersTable.Columns) != 4 {
		t.Errorf("Users columns count mismatch: got %d, want 4", len(usersTable.Columns))
	}

	// Check foreign key reference
	postsTable := unmarshaled.Tables[1]
	userIdCol := postsTable.Columns[1]
	if userIdCol.References != "users.id" {
		t.Errorf("user_id references mismatch: got %s, want users.id", userIdCol.References)
	}
}

func TestAPISpecStructure(t *testing.T) {
	spec := APISpec{
		Type:     "REST",
		Version:  "v1",
		BasePath: "/api/v1",
		Endpoints: []APIEndpoint{
			{
				Method:      "GET",
				Path:        "/users",
				Summary:     "List users",
				Description: "Get a paginated list of users",
				Parameters: []Parameter{
					{Name: "limit", In: "query", Type: "integer", Required: false},
					{Name: "offset", In: "query", Type: "integer", Required: false},
				},
				Responses: map[string]Response{
					"200": {Description: "Success", Schema: map[string]interface{}{"type": "array"}},
					"401": {Description: "Unauthorized"},
				},
				Auth: "bearer",
			},
			{
				Method:      "POST",
				Path:        "/users",
				Summary:     "Create user",
				Description: "Create a new user account",
				RequestBody: &RequestBody{
					ContentType: "application/json",
					Schema:      map[string]interface{}{"email": "string", "name": "string"},
					Required:    true,
				},
				Responses: map[string]Response{
					"201": {Description: "Created"},
					"400": {Description: "Bad Request"},
				},
				Auth: "bearer",
			},
			{
				Method:      "GET",
				Path:        "/users/{id}",
				Summary:     "Get user",
				Description: "Get a user by ID",
				Parameters: []Parameter{
					{Name: "id", In: "path", Type: "string", Required: true},
				},
				Responses: map[string]Response{
					"200": {Description: "Success"},
					"404": {Description: "Not Found"},
				},
				Auth: "bearer",
			},
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal APISpec: %v", err)
	}

	var unmarshaled APISpec
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal APISpec: %v", err)
	}

	if unmarshaled.Type != "REST" {
		t.Errorf("Type mismatch: got %s, want REST", unmarshaled.Type)
	}

	if len(unmarshaled.Endpoints) != 3 {
		t.Errorf("Endpoints count mismatch: got %d, want 3", len(unmarshaled.Endpoints))
	}

	// Check POST endpoint has request body
	postEndpoint := unmarshaled.Endpoints[1]
	if postEndpoint.RequestBody == nil {
		t.Error("POST endpoint should have request body")
	}

	// Check GET by ID endpoint has path parameter
	getByIdEndpoint := unmarshaled.Endpoints[2]
	if len(getByIdEndpoint.Parameters) != 1 || getByIdEndpoint.Parameters[0].In != "path" {
		t.Error("GET by ID endpoint should have path parameter")
	}
}

func TestGraphQLSpecStructure(t *testing.T) {
	spec := APISpec{
		Type:     "GraphQL",
		Version:  "v1",
		BasePath: "/graphql",
		Types: []GraphQLType{
			{
				Name: "User",
				Kind: "type",
				Fields: []GraphQLField{
					{Name: "id", Type: "ID!", Nullable: false},
					{Name: "email", Type: "String!", Nullable: false},
					{Name: "name", Type: "String", Nullable: true},
					{Name: "posts", Type: "[Post!]!", Nullable: false},
				},
			},
			{
				Name: "Post",
				Kind: "type",
				Fields: []GraphQLField{
					{Name: "id", Type: "ID!", Nullable: false},
					{Name: "title", Type: "String!", Nullable: false},
					{Name: "content", Type: "String", Nullable: true},
					{Name: "author", Type: "User!", Nullable: false},
				},
			},
			{
				Name: "CreateUserInput",
				Kind: "input",
				Fields: []GraphQLField{
					{Name: "email", Type: "String!", Nullable: false},
					{Name: "name", Type: "String", Nullable: true},
				},
			},
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal GraphQL APISpec: %v", err)
	}

	var unmarshaled APISpec
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal GraphQL APISpec: %v", err)
	}

	if unmarshaled.Type != "GraphQL" {
		t.Errorf("Type mismatch: got %s, want GraphQL", unmarshaled.Type)
	}

	if len(unmarshaled.Types) != 3 {
		t.Errorf("Types count mismatch: got %d, want 3", len(unmarshaled.Types))
	}

	// Check input type
	inputType := unmarshaled.Types[2]
	if inputType.Kind != "input" {
		t.Errorf("Third type kind mismatch: got %s, want input", inputType.Kind)
	}
}

func TestSanitizeMermaidID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Frontend", "Frontend"},
		{"Backend API", "Backend_API"},
		{"user-service", "user_service"},
		{"PostgreSQL DB", "PostgreSQL_DB"},
		{"auth.service", "auth_service"},
		{"My Cool Service", "My_Cool_Service"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeMermaidID(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeMermaidID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateMermaidDiagram(t *testing.T) {
	agent := NewArchitectAgent()

	design := &SystemDesign{
		ProjectName: "Test App",
		Components: []Component{
			{Name: "Web App", Type: "frontend", Technology: "React"},
			{Name: "API Server", Type: "backend", Technology: "Node.js"},
			{Name: "Database", Type: "database", Technology: "PostgreSQL"},
			{Name: "Cache", Type: "cache", Technology: "Redis"},
		},
		DataFlow: []DataFlowEdge{
			{From: "Web App", To: "API Server", Protocol: "HTTP"},
			{From: "API Server", To: "Database", Protocol: "SQL"},
			{From: "API Server", To: "Cache", Protocol: "Redis"},
		},
	}

	diagram := agent.generateMermaidDiagram(design)

	// Check that it starts with mermaid code block
	if !strings.HasPrefix(diagram, "```mermaid") {
		t.Error("Diagram should start with ```mermaid")
	}

	// Check that it ends with code block
	if !strings.HasSuffix(diagram, "```") {
		t.Error("Diagram should end with ```")
	}

	// Check for subgraphs
	if !strings.Contains(diagram, "subgraph Frontend") {
		t.Error("Diagram should contain Frontend subgraph")
	}

	if !strings.Contains(diagram, "subgraph Backend") {
		t.Error("Diagram should contain Backend subgraph")
	}

	if !strings.Contains(diagram, "subgraph Data") {
		t.Error("Diagram should contain Data subgraph")
	}

	// Check for data flow edges
	if !strings.Contains(diagram, "-->|HTTP|") {
		t.Error("Diagram should contain HTTP edge")
	}

	if !strings.Contains(diagram, "-->|SQL|") {
		t.Error("Diagram should contain SQL edge")
	}
}

func TestInputStructures(t *testing.T) {
	// Test AnalyzeRequirementsInput
	analyzeInput := AnalyzeRequirementsInput{
		ProjectName:  "My Project",
		Requirements: "Build a todo app with user authentication",
	}

	data, err := json.Marshal(analyzeInput)
	if err != nil {
		t.Fatalf("Failed to marshal AnalyzeRequirementsInput: %v", err)
	}

	var input map[string]interface{}
	json.Unmarshal(data, &input)

	if input["project_name"] != "My Project" {
		t.Errorf("project_name mismatch")
	}

	// Test SelectTechStackInput
	selectInput := SelectTechStackInput{
		Requirements: "Build a REST API",
		Constraints:  []string{"Must use Python", "AWS only"},
		Preferences:  []string{"Prefer FastAPI"},
	}

	data, err = json.Marshal(selectInput)
	if err != nil {
		t.Fatalf("Failed to marshal SelectTechStackInput: %v", err)
	}

	json.Unmarshal(data, &input)

	constraints := input["constraints"].([]interface{})
	if len(constraints) != 2 {
		t.Errorf("constraints count mismatch: got %d, want 2", len(constraints))
	}

	// Test DesignDatabaseInput
	dbInput := DesignDatabaseInput{
		Requirements: "User management system",
		TechStack: TechStack{
			Database: DatabaseStack{Primary: "PostgreSQL"},
		},
		Entities: []string{"User", "Role", "Permission"},
	}

	data, err = json.Marshal(dbInput)
	if err != nil {
		t.Fatalf("Failed to marshal DesignDatabaseInput: %v", err)
	}

	json.Unmarshal(data, &input)

	entities := input["entities"].([]interface{})
	if len(entities) != 3 {
		t.Errorf("entities count mismatch: got %d, want 3", len(entities))
	}

	// Test DesignAPIInput
	apiInput := DesignAPIInput{
		Requirements: "CRUD API for users",
		TechStack:    TechStack{},
		APIType:      "REST",
	}

	data, err = json.Marshal(apiInput)
	if err != nil {
		t.Fatalf("Failed to marshal DesignAPIInput: %v", err)
	}

	json.Unmarshal(data, &input)

	if input["api_type"] != "REST" {
		t.Errorf("api_type mismatch: got %v, want REST", input["api_type"])
	}
}

func TestNewArchitectAgent(t *testing.T) {
	agent := NewArchitectAgent()
	if agent == nil {
		t.Fatal("NewArchitectAgent returned nil")
	}

	if agent.llmClient == nil {
		t.Error("ArchitectAgent should have llmClient initialized")
	}
}

func TestComponentTypes(t *testing.T) {
	validTypes := []string{"frontend", "backend", "database", "cache", "queue", "external"}

	for _, componentType := range validTypes {
		c := Component{
			Name:        "Test Component",
			Type:        componentType,
			Description: "Test description",
			Technology:  "Test Tech",
		}

		data, err := json.Marshal(c)
		if err != nil {
			t.Errorf("Failed to marshal component with type %s: %v", componentType, err)
		}

		var unmarshaled Component
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Errorf("Failed to unmarshal component with type %s: %v", componentType, err)
		}

		if unmarshaled.Type != componentType {
			t.Errorf("Type mismatch for %s: got %s", componentType, unmarshaled.Type)
		}
	}
}

func TestDatabaseColumnTypes(t *testing.T) {
	validTypes := []string{
		"uuid",
		"varchar(255)",
		"text",
		"integer",
		"bigint",
		"boolean",
		"timestamp",
		"timestamptz",
		"jsonb",
		"decimal(10,2)",
	}

	for _, colType := range validTypes {
		col := ColumnDef{
			Name:     "test_column",
			Type:     colType,
			Nullable: false,
		}

		data, err := json.Marshal(col)
		if err != nil {
			t.Errorf("Failed to marshal column with type %s: %v", colType, err)
		}

		var unmarshaled ColumnDef
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Errorf("Failed to unmarshal column with type %s: %v", colType, err)
		}

		if unmarshaled.Type != colType {
			t.Errorf("Type mismatch for %s: got %s", colType, unmarshaled.Type)
		}
	}
}

func TestHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	for _, method := range methods {
		endpoint := APIEndpoint{
			Method:  method,
			Path:    "/test",
			Summary: "Test endpoint",
			Responses: map[string]Response{
				"200": {Description: "OK"},
			},
		}

		data, err := json.Marshal(endpoint)
		if err != nil {
			t.Errorf("Failed to marshal endpoint with method %s: %v", method, err)
		}

		var unmarshaled APIEndpoint
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Errorf("Failed to unmarshal endpoint with method %s: %v", method, err)
		}

		if unmarshaled.Method != method {
			t.Errorf("Method mismatch for %s: got %s", method, unmarshaled.Method)
		}
	}
}

func TestParameterLocations(t *testing.T) {
	locations := []string{"path", "query", "header"}

	for _, loc := range locations {
		param := Parameter{
			Name:     "test",
			In:       loc,
			Type:     "string",
			Required: true,
		}

		data, err := json.Marshal(param)
		if err != nil {
			t.Errorf("Failed to marshal parameter with location %s: %v", loc, err)
		}

		var unmarshaled Parameter
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Errorf("Failed to unmarshal parameter with location %s: %v", loc, err)
		}

		if unmarshaled.In != loc {
			t.Errorf("Location mismatch for %s: got %s", loc, unmarshaled.In)
		}
	}
}

func TestGraphQLTypeKinds(t *testing.T) {
	kinds := []string{"type", "input", "enum", "interface"}

	for _, kind := range kinds {
		gqlType := GraphQLType{
			Name:   "TestType",
			Kind:   kind,
			Fields: []GraphQLField{{Name: "id", Type: "ID!"}},
		}

		data, err := json.Marshal(gqlType)
		if err != nil {
			t.Errorf("Failed to marshal GraphQL type with kind %s: %v", kind, err)
		}

		var unmarshaled GraphQLType
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Errorf("Failed to unmarshal GraphQL type with kind %s: %v", kind, err)
		}

		if unmarshaled.Kind != kind {
			t.Errorf("Kind mismatch for %s: got %s", kind, unmarshaled.Kind)
		}
	}
}

func TestConstraintTypes(t *testing.T) {
	types := []string{"foreign_key", "unique", "check"}

	for _, constraintType := range types {
		constraint := Constraint{
			Name:       "test_constraint",
			Type:       constraintType,
			Table:      "test_table",
			Definition: "TEST",
		}

		data, err := json.Marshal(constraint)
		if err != nil {
			t.Errorf("Failed to marshal constraint with type %s: %v", constraintType, err)
		}

		var unmarshaled Constraint
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Errorf("Failed to unmarshal constraint with type %s: %v", constraintType, err)
		}

		if unmarshaled.Type != constraintType {
			t.Errorf("Type mismatch for %s: got %s", constraintType, unmarshaled.Type)
		}
	}
}

func TestEmptyDesign(t *testing.T) {
	design := SystemDesign{}

	data, err := json.Marshal(design)
	if err != nil {
		t.Fatalf("Failed to marshal empty SystemDesign: %v", err)
	}

	var unmarshaled SystemDesign
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal empty SystemDesign: %v", err)
	}

	if unmarshaled.Components != nil && len(unmarshaled.Components) != 0 {
		t.Error("Empty design should have nil or empty components")
	}
}

func TestMermaidDiagramWithNoComponents(t *testing.T) {
	agent := NewArchitectAgent()

	design := &SystemDesign{
		ProjectName: "Empty Project",
		Components:  []Component{},
		DataFlow:    []DataFlowEdge{},
	}

	diagram := agent.generateMermaidDiagram(design)

	// Should still produce valid mermaid
	if !strings.HasPrefix(diagram, "```mermaid") {
		t.Error("Empty diagram should still be valid mermaid")
	}

	// Should not have subgraphs since no components
	if strings.Contains(diagram, "subgraph Frontend") {
		t.Error("Empty diagram should not have Frontend subgraph")
	}
}
