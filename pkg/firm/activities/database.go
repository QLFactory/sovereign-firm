package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// DatabaseAgent handles database schema design, migrations, seed data, and optimization
// Supports PostgreSQL, MySQL, MongoDB, and various ORMs (Prisma, SQLAlchemy, GORM, Knex)
type DatabaseAgent struct {
	llmClient llm.Client
}

func NewDatabaseAgent() *DatabaseAgent {
	return &DatabaseAgent{
		llmClient: llm.NewClient(),
	}
}

// ============================================================================
// Data Structures
// ============================================================================

// DatabaseCodeBundle represents generated database files
type DatabaseCodeBundle struct {
	DatabaseType string            `json:"database_type"` // "postgresql", "mysql", "mongodb"
	ORM          string            `json:"orm"`           // "prisma", "sqlalchemy", "gorm", "knex", "mongoose"
	Files        map[string]string `json:"files"`         // filename -> content
	MigrateCmd   string            `json:"migrate_cmd"`   // Command to run migrations
	SeedCmd      string            `json:"seed_cmd"`      // Command to run seeds
	ResetCmd     string            `json:"reset_cmd"`     // Command to reset database
}

// MigrationFile represents a single migration
type MigrationFile struct {
	Version     string `json:"version"`     // e.g., "001", "20240115120000"
	Name        string `json:"name"`        // e.g., "create_users_table"
	UpSQL       string `json:"up_sql"`      // SQL to apply migration
	DownSQL     string `json:"down_sql"`    // SQL to rollback migration
	Description string `json:"description"` // What this migration does
}

// SeedData represents seed data for a table
type SeedData struct {
	TableName string                   `json:"table_name"`
	Records   []map[string]interface{} `json:"records"`
}

// SchemaOptimization represents a suggested optimization
type SchemaOptimization struct {
	Type        string `json:"type"`        // "index", "denormalize", "partition", "constraint"
	Table       string `json:"table"`
	Description string `json:"description"`
	SQL         string `json:"sql"`
	Impact      string `json:"impact"` // "high", "medium", "low"
	Reason      string `json:"reason"`
}

// ============================================================================
// Input Types
// ============================================================================

// DesignSchemaInput for designing a new schema
type DesignSchemaInput struct {
	Requirements string   `json:"requirements"`
	Entities     []string `json:"entities,omitempty"` // Optional: specific entities
	DatabaseType string   `json:"database_type"`      // "postgresql", "mysql", "mongodb"
}

// GenerateMigrationsInput for generating migrations
type GenerateMigrationsInput struct {
	Schema       *DatabaseSchema `json:"schema"`
	DatabaseType string          `json:"database_type"`
	ORM          string          `json:"orm"` // "prisma", "knex", "sqlalchemy", "gorm", "raw"
	ExistingSchema *DatabaseSchema `json:"existing_schema,omitempty"` // For diff migrations
}

// GenerateSeedInput for generating seed data
type GenerateSeedInput struct {
	Schema       *DatabaseSchema `json:"schema"`
	DatabaseType string          `json:"database_type"`
	ORM          string          `json:"orm"`
	RecordsPerTable int          `json:"records_per_table"` // How many records to generate
}

// OptimizeSchemaInput for schema optimization
type OptimizeSchemaInput struct {
	Schema       *DatabaseSchema `json:"schema"`
	QueryPatterns []string       `json:"query_patterns,omitempty"` // Common queries to optimize for
	DatabaseType string          `json:"database_type"`
}

// ============================================================================
// Activity Methods
// ============================================================================

// DesignSchema creates a database schema from requirements
func (a *DatabaseAgent) DesignSchema(ctx context.Context, input map[string]interface{}) (*DatabaseSchema, error) {
	inputBytes, _ := json.Marshal(input)
	var req DesignSchemaInput
	json.Unmarshal(inputBytes, &req)

	dbType := req.DatabaseType
	if dbType == "" {
		dbType = "postgresql"
	}

	entitiesHint := ""
	if len(req.Entities) > 0 {
		entitiesHint = "\n\nKey entities to model: " + strings.Join(req.Entities, ", ")
	}

	sysPrompt := fmt.Sprintf(`You are a Database Architect designing schemas for %s.
Design a well-normalized database schema based on the requirements.

Output ONLY valid JSON with this structure:
{
  "database_type": "%s",
  "tables": [
    {
      "name": "table_name_plural_snake_case",
      "description": "what this table stores",
      "columns": [
        {
          "name": "column_name",
          "type": "appropriate_type",
          "nullable": false,
          "default": "optional_default",
          "references": "other_table.column"
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
      "columns": ["column1"],
      "unique": false
    }
  ],
  "constraints": [
    {
      "name": "fk_table_ref",
      "type": "foreign_key",
      "table": "table_name",
      "definition": "FOREIGN KEY (col) REFERENCES other(id)"
    }
  ]
}

Column Types for %s:
- uuid: For primary keys and references
- varchar(n): For short strings (emails, names)
- text: For long text (descriptions, content)
- integer/bigint: For numbers
- boolean: For true/false
- timestamp/timestamptz: For dates
- jsonb: For flexible JSON data (use sparingly)
- decimal(p,s): For money/precise numbers

Design Guidelines:
1. Use UUIDs for primary keys
2. Always include created_at, updated_at
3. Use snake_case for names
4. Create indexes for foreign keys
5. Create indexes for frequently queried columns
6. Normalize to 3NF
7. Add soft delete (deleted_at) for important tables
8. Use appropriate constraints`, dbType, dbType, dbType)

	prompt := fmt.Sprintf("Requirements:\n%s%s\n\nDesign the database schema.", req.Requirements, entitiesHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("schema design LLM call failed: %w", err)
	}

	var schema DatabaseSchema
	if err := extractJSON(resp.Response, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	return &schema, nil
}

// GenerateMigrations creates migration files from a schema
func (a *DatabaseAgent) GenerateMigrations(ctx context.Context, input map[string]interface{}) (*DatabaseCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req GenerateMigrationsInput
	json.Unmarshal(inputBytes, &req)

	if req.Schema == nil {
		return nil, fmt.Errorf("schema is required")
	}

	orm := strings.ToLower(req.ORM)
	if orm == "" {
		orm = "prisma"
	}

	dbType := strings.ToLower(req.DatabaseType)
	if dbType == "" {
		dbType = "postgresql"
	}

	switch orm {
	case "prisma":
		return a.generatePrismaMigrations(ctx, req.Schema, dbType)
	case "knex":
		return a.generateKnexMigrations(ctx, req.Schema, dbType)
	case "sqlalchemy", "alembic":
		return a.generateAlembicMigrations(ctx, req.Schema, dbType)
	case "gorm":
		return a.generateGORMMigrations(ctx, req.Schema, dbType)
	case "mongoose":
		return a.generateMongooseMigrations(ctx, req.Schema)
	default:
		return a.generateRawSQLMigrations(ctx, req.Schema, dbType)
	}
}

// GenerateSeedData creates seed data for testing
func (a *DatabaseAgent) GenerateSeedData(ctx context.Context, input map[string]interface{}) (*DatabaseCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req GenerateSeedInput
	json.Unmarshal(inputBytes, &req)

	if req.Schema == nil {
		return nil, fmt.Errorf("schema is required")
	}

	recordsPerTable := req.RecordsPerTable
	if recordsPerTable <= 0 {
		recordsPerTable = 10
	}

	orm := strings.ToLower(req.ORM)
	if orm == "" {
		orm = "prisma"
	}

	sysPrompt := fmt.Sprintf(`You are a Database Developer generating realistic seed data.
Generate seed data for the provided schema.

Output ONLY valid JSON with this structure:
{
  "database_type": "%s",
  "orm": "%s",
  "files": {
    "path/to/seed/file": "seed file content"
  },
  "seed_cmd": "command to run seeds"
}

Requirements:
1. Generate %d realistic records per table
2. Ensure foreign key relationships are valid
3. Use realistic data (proper emails, names, etc.)
4. Include variety in the data
5. Respect constraints (unique, not null)
6. Use proper date formats
7. Generate UUIDs for id fields`, req.DatabaseType, orm, recordsPerTable)

	schemaBytes, _ := json.MarshalIndent(req.Schema, "", "  ")
	prompt := fmt.Sprintf("DATABASE SCHEMA:\n%s\n\nGenerate seed data with %d records per table.",
		string(schemaBytes), recordsPerTable)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("seed generation LLM call failed: %w", err)
	}

	var bundle DatabaseCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse seed data: %w", err)
	}

	return &bundle, nil
}

// OptimizeSchema suggests optimizations for a schema
func (a *DatabaseAgent) OptimizeSchema(ctx context.Context, input map[string]interface{}) ([]SchemaOptimization, error) {
	inputBytes, _ := json.Marshal(input)
	var req OptimizeSchemaInput
	json.Unmarshal(inputBytes, &req)

	if req.Schema == nil {
		return nil, fmt.Errorf("schema is required")
	}

	dbType := req.DatabaseType
	if dbType == "" {
		dbType = "postgresql"
	}

	queryPatternsHint := ""
	if len(req.QueryPatterns) > 0 {
		queryPatternsHint = "\n\nCommon Query Patterns:\n- " + strings.Join(req.QueryPatterns, "\n- ")
	}

	sysPrompt := fmt.Sprintf(`You are a Database Performance Expert for %s.
Analyze the schema and suggest optimizations.

Output ONLY valid JSON array:
[
  {
    "type": "index|denormalize|partition|constraint|type_change",
    "table": "table_name",
    "description": "What to do",
    "sql": "SQL statement to implement",
    "impact": "high|medium|low",
    "reason": "Why this helps"
  }
]

Optimization Types:
1. index: Add indexes for query performance
2. denormalize: Add redundant data to avoid joins
3. partition: Partition large tables
4. constraint: Add constraints for data integrity
5. type_change: Change column types for efficiency

Consider:
- Missing indexes on foreign keys
- Missing indexes on commonly queried columns
- Composite indexes for multi-column queries
- Partial indexes for filtered queries
- Table partitioning for large tables
- Denormalization for read-heavy tables`, dbType)

	schemaBytes, _ := json.MarshalIndent(req.Schema, "", "  ")
	prompt := fmt.Sprintf("DATABASE SCHEMA:\n%s%s\n\nSuggest optimizations.",
		string(schemaBytes), queryPatternsHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("optimization LLM call failed: %w", err)
	}

	var optimizations []SchemaOptimization
	if err := extractJSON(resp.Response, &optimizations); err != nil {
		return nil, fmt.Errorf("failed to parse optimizations: %w", err)
	}

	return optimizations, nil
}

// ============================================================================
// ORM-Specific Migration Generators
// ============================================================================

func (a *DatabaseAgent) generatePrismaMigrations(ctx context.Context, schema *DatabaseSchema, dbType string) (*DatabaseCodeBundle, error) {
	sysPrompt := `You are a Prisma expert generating schema files.
Output ONLY valid JSON with this structure:
{
  "database_type": "postgresql",
  "orm": "prisma",
  "files": {
    "prisma/schema.prisma": "prisma schema content"
  },
  "migrate_cmd": "npx prisma migrate dev",
  "seed_cmd": "npx prisma db seed",
  "reset_cmd": "npx prisma migrate reset"
}

Prisma Schema Guidelines:
1. Use @id for primary keys
2. Use @default(uuid()) for UUID fields
3. Use @default(now()) for timestamps
4. Use @relation for foreign keys
5. Use @unique for unique constraints
6. Use @map for column name mapping
7. Use @@map for table name mapping
8. Add proper indexes with @@index`

	schemaBytes, _ := json.MarshalIndent(schema, "", "  ")
	prompt := fmt.Sprintf("DATABASE: %s\n\nSCHEMA:\n%s\n\nGenerate Prisma schema.", dbType, string(schemaBytes))

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Prisma generation failed: %w", err)
	}

	var bundle DatabaseCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Prisma output: %w", err)
	}

	bundle.DatabaseType = dbType
	bundle.ORM = "prisma"
	return &bundle, nil
}

func (a *DatabaseAgent) generateKnexMigrations(ctx context.Context, schema *DatabaseSchema, dbType string) (*DatabaseCodeBundle, error) {
	sysPrompt := `You are a Knex.js expert generating migrations.
Output ONLY valid JSON with this structure:
{
  "database_type": "postgresql",
  "orm": "knex",
  "files": {
    "migrations/20240115120000_initial.js": "migration content",
    "knexfile.js": "knex config"
  },
  "migrate_cmd": "npx knex migrate:latest",
  "seed_cmd": "npx knex seed:run",
  "reset_cmd": "npx knex migrate:rollback --all && npx knex migrate:latest"
}

Knex Migration Guidelines:
1. Use table.uuid('id').primary() for UUIDs
2. Use table.timestamps(true, true) for created_at/updated_at
3. Use table.foreign() for foreign keys
4. Use proper column types: text, varchar, integer, boolean
5. Create indexes with table.index()
6. Use async/await syntax`

	schemaBytes, _ := json.MarshalIndent(schema, "", "  ")
	timestamp := time.Now().Format("20060102150405")
	prompt := fmt.Sprintf("DATABASE: %s\nTIMESTAMP: %s\n\nSCHEMA:\n%s\n\nGenerate Knex migrations.",
		dbType, timestamp, string(schemaBytes))

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Knex generation failed: %w", err)
	}

	var bundle DatabaseCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Knex output: %w", err)
	}

	bundle.DatabaseType = dbType
	bundle.ORM = "knex"
	return &bundle, nil
}

func (a *DatabaseAgent) generateAlembicMigrations(ctx context.Context, schema *DatabaseSchema, dbType string) (*DatabaseCodeBundle, error) {
	sysPrompt := `You are a SQLAlchemy/Alembic expert generating migrations.
Output ONLY valid JSON with this structure:
{
  "database_type": "postgresql",
  "orm": "sqlalchemy",
  "files": {
    "alembic/versions/001_initial.py": "migration content",
    "alembic/env.py": "alembic env config",
    "alembic.ini": "alembic config",
    "app/models/__init__.py": "models init",
    "app/models/base.py": "base model",
    "app/models/user.py": "model file per table"
  },
  "migrate_cmd": "alembic upgrade head",
  "seed_cmd": "python -m app.seeds",
  "reset_cmd": "alembic downgrade base && alembic upgrade head"
}

SQLAlchemy/Alembic Guidelines:
1. Use mapped_column() for SQLAlchemy 2.0
2. Use Mapped[] type hints
3. Use UUID type from sqlalchemy.dialects.postgresql
4. Use relationship() for foreign keys
5. Use proper naming conventions
6. Include __tablename__ in models`

	schemaBytes, _ := json.MarshalIndent(schema, "", "  ")
	prompt := fmt.Sprintf("DATABASE: %s\n\nSCHEMA:\n%s\n\nGenerate SQLAlchemy models and Alembic migrations.",
		dbType, string(schemaBytes))

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Alembic generation failed: %w", err)
	}

	var bundle DatabaseCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Alembic output: %w", err)
	}

	bundle.DatabaseType = dbType
	bundle.ORM = "sqlalchemy"
	return &bundle, nil
}

func (a *DatabaseAgent) generateGORMMigrations(ctx context.Context, schema *DatabaseSchema, dbType string) (*DatabaseCodeBundle, error) {
	sysPrompt := `You are a GORM expert generating models and migrations.
Output ONLY valid JSON with this structure:
{
  "database_type": "postgresql",
  "orm": "gorm",
  "files": {
    "internal/models/user.go": "model file",
    "internal/models/models.go": "models init with AutoMigrate",
    "internal/db/db.go": "database connection",
    "cmd/migrate/main.go": "migration command"
  },
  "migrate_cmd": "go run cmd/migrate/main.go",
  "seed_cmd": "go run cmd/seed/main.go",
  "reset_cmd": "go run cmd/migrate/main.go --reset"
}

GORM Guidelines:
1. Use gorm.Model for base fields (ID, CreatedAt, UpdatedAt, DeletedAt)
2. Use uuid.UUID with gorm:"type:uuid;primaryKey"
3. Use proper struct tags for column definitions
4. Use gorm:"foreignKey:FieldID" for relations
5. Use gorm:"index" for indexes
6. Use gorm:"uniqueIndex" for unique indexes
7. Follow Go naming conventions (PascalCase)`

	schemaBytes, _ := json.MarshalIndent(schema, "", "  ")
	prompt := fmt.Sprintf("DATABASE: %s\n\nSCHEMA:\n%s\n\nGenerate GORM models.", dbType, string(schemaBytes))

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("GORM generation failed: %w", err)
	}

	var bundle DatabaseCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse GORM output: %w", err)
	}

	bundle.DatabaseType = dbType
	bundle.ORM = "gorm"
	return &bundle, nil
}

func (a *DatabaseAgent) generateMongooseMigrations(ctx context.Context, schema *DatabaseSchema) (*DatabaseCodeBundle, error) {
	sysPrompt := `You are a Mongoose expert generating schemas.
Output ONLY valid JSON with this structure:
{
  "database_type": "mongodb",
  "orm": "mongoose",
  "files": {
    "src/models/user.model.js": "mongoose schema",
    "src/models/index.js": "models export",
    "src/db/connection.js": "mongodb connection"
  },
  "migrate_cmd": "npm run migrate",
  "seed_cmd": "npm run seed",
  "reset_cmd": "npm run db:reset"
}

Mongoose Guidelines:
1. Use Schema.Types.ObjectId for references
2. Use ref for population
3. Add timestamps: true option
4. Use proper validators
5. Add indexes with index: true
6. Use unique: true for unique fields
7. Define virtual fields if needed`

	schemaBytes, _ := json.MarshalIndent(schema, "", "  ")
	prompt := fmt.Sprintf("SCHEMA:\n%s\n\nGenerate Mongoose schemas.", string(schemaBytes))

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Mongoose generation failed: %w", err)
	}

	var bundle DatabaseCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Mongoose output: %w", err)
	}

	bundle.DatabaseType = "mongodb"
	bundle.ORM = "mongoose"
	return &bundle, nil
}

func (a *DatabaseAgent) generateRawSQLMigrations(ctx context.Context, schema *DatabaseSchema, dbType string) (*DatabaseCodeBundle, error) {
	sysPrompt := fmt.Sprintf(`You are a SQL expert generating raw migration files for %s.
Output ONLY valid JSON with this structure:
{
  "database_type": "%s",
  "orm": "raw",
  "files": {
    "migrations/001_initial_up.sql": "CREATE TABLE statements",
    "migrations/001_initial_down.sql": "DROP TABLE statements"
  },
  "migrate_cmd": "psql -f migrations/001_initial_up.sql",
  "seed_cmd": "psql -f seeds/seed.sql",
  "reset_cmd": "psql -f migrations/001_initial_down.sql && psql -f migrations/001_initial_up.sql"
}

SQL Guidelines for %s:
1. Use proper %s types (uuid, varchar, text, etc.)
2. Add PRIMARY KEY constraints
3. Add FOREIGN KEY constraints with ON DELETE
4. Add NOT NULL constraints
5. Add DEFAULT values
6. Create indexes separately
7. Order tables by dependencies (referenced tables first)
8. Include CREATE INDEX statements`, dbType, dbType, dbType, dbType)

	schemaBytes, _ := json.MarshalIndent(schema, "", "  ")
	prompt := fmt.Sprintf("SCHEMA:\n%s\n\nGenerate raw SQL migrations.", string(schemaBytes))

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("SQL generation failed: %w", err)
	}

	var bundle DatabaseCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse SQL output: %w", err)
	}

	bundle.DatabaseType = dbType
	bundle.ORM = "raw"
	return &bundle, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// ValidateSchema checks a schema for common issues
func ValidateSchema(schema *DatabaseSchema) []string {
	var issues []string

	if len(schema.Tables) == 0 {
		issues = append(issues, "Schema has no tables")
		return issues
	}

	tableNames := make(map[string]bool)
	for _, table := range schema.Tables {
		// Check for duplicate table names
		if tableNames[table.Name] {
			issues = append(issues, fmt.Sprintf("Duplicate table name: %s", table.Name))
		}
		tableNames[table.Name] = true

		// Check for primary key
		if len(table.PrimaryKey) == 0 {
			issues = append(issues, fmt.Sprintf("Table %s has no primary key", table.Name))
		}

		// Check for columns
		if len(table.Columns) == 0 {
			issues = append(issues, fmt.Sprintf("Table %s has no columns", table.Name))
		}

		columnNames := make(map[string]bool)
		for _, col := range table.Columns {
			// Check for duplicate column names
			if columnNames[col.Name] {
				issues = append(issues, fmt.Sprintf("Table %s has duplicate column: %s", table.Name, col.Name))
			}
			columnNames[col.Name] = true

			// Check foreign key references
			if col.References != "" {
				parts := strings.Split(col.References, ".")
				if len(parts) != 2 {
					issues = append(issues, fmt.Sprintf("Invalid reference format in %s.%s: %s", table.Name, col.Name, col.References))
				} else {
					refTable := parts[0]
					if !tableNames[refTable] && refTable != table.Name {
						// Check if referenced table exists (might be defined later)
						found := false
						for _, t := range schema.Tables {
							if t.Name == refTable {
								found = true
								break
							}
						}
						if !found {
							issues = append(issues, fmt.Sprintf("Table %s.%s references non-existent table: %s", table.Name, col.Name, refTable))
						}
					}
				}
			}
		}
	}

	return issues
}

// GetTableDependencyOrder returns tables in order of dependencies (referenced tables first)
func GetTableDependencyOrder(schema *DatabaseSchema) []string {
	// Build dependency graph
	deps := make(map[string][]string)
	for _, table := range schema.Tables {
		deps[table.Name] = []string{}
		for _, col := range table.Columns {
			if col.References != "" {
				parts := strings.Split(col.References, ".")
				if len(parts) == 2 {
					refTable := parts[0]
					if refTable != table.Name { // Skip self-references
						deps[table.Name] = append(deps[table.Name], refTable)
					}
				}
			}
		}
	}

	// Topological sort
	var result []string
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(name string) bool
	visit = func(name string) bool {
		if visiting[name] {
			return false // Cycle detected
		}
		if visited[name] {
			return true
		}

		visiting[name] = true
		for _, dep := range deps[name] {
			if !visit(dep) {
				return false
			}
		}
		visiting[name] = false
		visited[name] = true
		result = append(result, name)
		return true
	}

	for _, table := range schema.Tables {
		visit(table.Name)
	}

	return result
}

// GenerateCreateTableSQL generates CREATE TABLE SQL for a table
func GenerateCreateTableSQL(table *TableSchema, dbType string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table.Name))

	for i, col := range table.Columns {
		sb.WriteString(fmt.Sprintf("  %s %s", col.Name, col.Type))

		if !col.Nullable {
			sb.WriteString(" NOT NULL")
		}

		if col.Default != "" {
			sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.Default))
		}

		if i < len(table.Columns)-1 || len(table.PrimaryKey) > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	if len(table.PrimaryKey) > 0 {
		sb.WriteString(fmt.Sprintf("  PRIMARY KEY (%s)\n", strings.Join(table.PrimaryKey, ", ")))
	}

	sb.WriteString(");")

	return sb.String()
}
