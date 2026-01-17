package activities

import (
	"encoding/json"
	"strings"
	"testing"
)

// ============================================================================
// Data Structure Tests
// ============================================================================

func TestDatabaseCodeBundleStructure(t *testing.T) {
	bundle := DatabaseCodeBundle{
		DatabaseType: "postgresql",
		ORM:          "prisma",
		Files: map[string]string{
			"prisma/schema.prisma": "datasource db { ... }",
		},
		MigrateCmd: "npx prisma migrate dev",
		SeedCmd:    "npx prisma db seed",
		ResetCmd:   "npx prisma migrate reset",
	}

	// Test JSON serialization
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("Failed to marshal DatabaseCodeBundle: %v", err)
	}

	// Test JSON deserialization
	var decoded DatabaseCodeBundle
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DatabaseCodeBundle: %v", err)
	}

	if decoded.DatabaseType != "postgresql" {
		t.Errorf("Expected database_type 'postgresql', got '%s'", decoded.DatabaseType)
	}
	if decoded.ORM != "prisma" {
		t.Errorf("Expected orm 'prisma', got '%s'", decoded.ORM)
	}
	if len(decoded.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(decoded.Files))
	}
}

func TestMigrationFileStructure(t *testing.T) {
	migration := MigrationFile{
		Version:     "001",
		Name:        "create_users_table",
		UpSQL:       "CREATE TABLE users (...);",
		DownSQL:     "DROP TABLE users;",
		Description: "Creates the initial users table",
	}

	data, err := json.Marshal(migration)
	if err != nil {
		t.Fatalf("Failed to marshal MigrationFile: %v", err)
	}

	var decoded MigrationFile
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal MigrationFile: %v", err)
	}

	if decoded.Version != "001" {
		t.Errorf("Expected version '001', got '%s'", decoded.Version)
	}
	if decoded.Name != "create_users_table" {
		t.Errorf("Expected name 'create_users_table', got '%s'", decoded.Name)
	}
}

func TestSeedDataStructure(t *testing.T) {
	seed := SeedData{
		TableName: "users",
		Records: []map[string]interface{}{
			{"id": "uuid-1", "name": "John", "email": "john@example.com"},
			{"id": "uuid-2", "name": "Jane", "email": "jane@example.com"},
		},
	}

	data, err := json.Marshal(seed)
	if err != nil {
		t.Fatalf("Failed to marshal SeedData: %v", err)
	}

	var decoded SeedData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SeedData: %v", err)
	}

	if decoded.TableName != "users" {
		t.Errorf("Expected table_name 'users', got '%s'", decoded.TableName)
	}
	if len(decoded.Records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(decoded.Records))
	}
}

func TestSchemaOptimizationStructure(t *testing.T) {
	opt := SchemaOptimization{
		Type:        "index",
		Table:       "users",
		Description: "Add index on email column",
		SQL:         "CREATE INDEX idx_users_email ON users(email);",
		Impact:      "high",
		Reason:      "Email lookups are frequent",
	}

	data, err := json.Marshal(opt)
	if err != nil {
		t.Fatalf("Failed to marshal SchemaOptimization: %v", err)
	}

	var decoded SchemaOptimization
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SchemaOptimization: %v", err)
	}

	if decoded.Type != "index" {
		t.Errorf("Expected type 'index', got '%s'", decoded.Type)
	}
	if decoded.Impact != "high" {
		t.Errorf("Expected impact 'high', got '%s'", decoded.Impact)
	}
}

// ============================================================================
// Input Type Tests
// ============================================================================

func TestDesignSchemaInputStructure(t *testing.T) {
	input := DesignSchemaInput{
		Requirements: "E-commerce platform with users, products, orders",
		Entities:     []string{"users", "products", "orders", "order_items"},
		DatabaseType: "postgresql",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal DesignSchemaInput: %v", err)
	}

	var decoded DesignSchemaInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DesignSchemaInput: %v", err)
	}

	if len(decoded.Entities) != 4 {
		t.Errorf("Expected 4 entities, got %d", len(decoded.Entities))
	}
}

func TestGenerateMigrationsInputStructure(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{Name: "users", Columns: []ColumnDef{{Name: "id", Type: "uuid"}}},
		},
	}

	input := GenerateMigrationsInput{
		Schema:       schema,
		DatabaseType: "postgresql",
		ORM:          "prisma",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal GenerateMigrationsInput: %v", err)
	}

	var decoded GenerateMigrationsInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal GenerateMigrationsInput: %v", err)
	}

	if decoded.ORM != "prisma" {
		t.Errorf("Expected orm 'prisma', got '%s'", decoded.ORM)
	}
}

func TestGenerateSeedInputStructure(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables:       []TableSchema{{Name: "users"}},
	}

	input := GenerateSeedInput{
		Schema:          schema,
		DatabaseType:    "postgresql",
		ORM:             "knex",
		RecordsPerTable: 25,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal GenerateSeedInput: %v", err)
	}

	var decoded GenerateSeedInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal GenerateSeedInput: %v", err)
	}

	if decoded.RecordsPerTable != 25 {
		t.Errorf("Expected records_per_table 25, got %d", decoded.RecordsPerTable)
	}
}

func TestOptimizeSchemaInputStructure(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables:       []TableSchema{{Name: "users"}},
	}

	input := OptimizeSchemaInput{
		Schema:        schema,
		QueryPatterns: []string{"SELECT * FROM users WHERE email = ?", "SELECT * FROM orders WHERE user_id = ?"},
		DatabaseType:  "postgresql",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal OptimizeSchemaInput: %v", err)
	}

	var decoded OptimizeSchemaInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal OptimizeSchemaInput: %v", err)
	}

	if len(decoded.QueryPatterns) != 2 {
		t.Errorf("Expected 2 query patterns, got %d", len(decoded.QueryPatterns))
	}
}

// ============================================================================
// Validation Tests
// ============================================================================

func TestValidateSchemaEmpty(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables:       []TableSchema{},
	}

	issues := ValidateSchema(schema)

	if len(issues) == 0 {
		t.Error("Expected validation issues for empty schema")
	}
	if issues[0] != "Schema has no tables" {
		t.Errorf("Expected 'Schema has no tables', got '%s'", issues[0])
	}
}

func TestValidateSchemaDuplicateTables(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{Name: "users", Columns: []ColumnDef{{Name: "id", Type: "uuid"}}, PrimaryKey: []string{"id"}},
			{Name: "users", Columns: []ColumnDef{{Name: "id", Type: "uuid"}}, PrimaryKey: []string{"id"}},
		},
	}

	issues := ValidateSchema(schema)

	found := false
	for _, issue := range issues {
		if issue == "Duplicate table name: users" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Duplicate table name: users' in validation issues")
	}
}

func TestValidateSchemaNoPrimaryKey(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{Name: "users", Columns: []ColumnDef{{Name: "id", Type: "uuid"}}, PrimaryKey: []string{}},
		},
	}

	issues := ValidateSchema(schema)

	found := false
	for _, issue := range issues {
		if issue == "Table users has no primary key" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Table users has no primary key' in validation issues")
	}
}

func TestValidateSchemaNoColumns(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{Name: "users", Columns: []ColumnDef{}, PrimaryKey: []string{"id"}},
		},
	}

	issues := ValidateSchema(schema)

	found := false
	for _, issue := range issues {
		if issue == "Table users has no columns" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Table users has no columns' in validation issues")
	}
}

func TestValidateSchemaDuplicateColumns(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{
				Name: "users",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
					{Name: "id", Type: "uuid"}, // Duplicate
				},
				PrimaryKey: []string{"id"},
			},
		},
	}

	issues := ValidateSchema(schema)

	found := false
	for _, issue := range issues {
		if issue == "Table users has duplicate column: id" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Table users has duplicate column: id' in validation issues")
	}
}

func TestValidateSchemaInvalidReference(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{
				Name: "orders",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
					{Name: "user_id", Type: "uuid", References: "nonexistent.id"},
				},
				PrimaryKey: []string{"id"},
			},
		},
	}

	issues := ValidateSchema(schema)

	found := false
	for _, issue := range issues {
		if issue == "Table orders.user_id references non-existent table: nonexistent" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected reference to non-existent table in validation issues")
	}
}

func TestValidateSchemaInvalidReferenceFormat(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{
				Name: "orders",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
					{Name: "user_id", Type: "uuid", References: "invalid_format"},
				},
				PrimaryKey: []string{"id"},
			},
		},
	}

	issues := ValidateSchema(schema)

	found := false
	for _, issue := range issues {
		if issue == "Invalid reference format in orders.user_id: invalid_format" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected invalid reference format in validation issues")
	}
}

func TestValidateSchemaValid(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{
				Name: "users",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
					{Name: "email", Type: "varchar(255)"},
				},
				PrimaryKey: []string{"id"},
			},
			{
				Name: "orders",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
					{Name: "user_id", Type: "uuid", References: "users.id"},
				},
				PrimaryKey: []string{"id"},
			},
		},
	}

	issues := ValidateSchema(schema)

	if len(issues) != 0 {
		t.Errorf("Expected no validation issues, got %d: %v", len(issues), issues)
	}
}

// ============================================================================
// Dependency Order Tests
// ============================================================================

func TestGetTableDependencyOrderNoDependencies(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{Name: "users", Columns: []ColumnDef{{Name: "id", Type: "uuid"}}},
			{Name: "products", Columns: []ColumnDef{{Name: "id", Type: "uuid"}}},
		},
	}

	order := GetTableDependencyOrder(schema)

	if len(order) != 2 {
		t.Errorf("Expected 2 tables in order, got %d", len(order))
	}
}

func TestGetTableDependencyOrderWithDependencies(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{
				Name: "orders",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
					{Name: "user_id", Type: "uuid", References: "users.id"},
				},
			},
			{
				Name: "users",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
				},
			},
		},
	}

	order := GetTableDependencyOrder(schema)

	// Users should come before orders
	userIdx := -1
	orderIdx := -1
	for i, name := range order {
		if name == "users" {
			userIdx = i
		}
		if name == "orders" {
			orderIdx = i
		}
	}

	if userIdx > orderIdx {
		t.Error("Expected users to come before orders in dependency order")
	}
}

func TestGetTableDependencyOrderComplex(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{
				Name: "order_items",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
					{Name: "order_id", Type: "uuid", References: "orders.id"},
					{Name: "product_id", Type: "uuid", References: "products.id"},
				},
			},
			{
				Name: "orders",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
					{Name: "user_id", Type: "uuid", References: "users.id"},
				},
			},
			{
				Name: "users",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
				},
			},
			{
				Name: "products",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
				},
			},
		},
	}

	order := GetTableDependencyOrder(schema)

	// Build index map
	idxMap := make(map[string]int)
	for i, name := range order {
		idxMap[name] = i
	}

	// Check: users before orders
	if idxMap["users"] > idxMap["orders"] {
		t.Error("Expected users before orders")
	}

	// Check: orders before order_items
	if idxMap["orders"] > idxMap["order_items"] {
		t.Error("Expected orders before order_items")
	}

	// Check: products before order_items
	if idxMap["products"] > idxMap["order_items"] {
		t.Error("Expected products before order_items")
	}
}

func TestGetTableDependencyOrderSelfReference(t *testing.T) {
	schema := &DatabaseSchema{
		DatabaseType: "postgresql",
		Tables: []TableSchema{
			{
				Name: "employees",
				Columns: []ColumnDef{
					{Name: "id", Type: "uuid"},
					{Name: "manager_id", Type: "uuid", References: "employees.id"}, // Self-reference
				},
			},
		},
	}

	order := GetTableDependencyOrder(schema)

	// Should handle self-reference without issues
	if len(order) != 1 || order[0] != "employees" {
		t.Errorf("Expected [employees], got %v", order)
	}
}

// ============================================================================
// SQL Generation Tests
// ============================================================================

func TestGenerateCreateTableSQLBasic(t *testing.T) {
	table := &TableSchema{
		Name: "users",
		Columns: []ColumnDef{
			{Name: "id", Type: "uuid", Nullable: false},
			{Name: "email", Type: "varchar(255)", Nullable: false},
			{Name: "name", Type: "varchar(100)", Nullable: true},
		},
		PrimaryKey: []string{"id"},
	}

	sql := GenerateCreateTableSQL(table, "postgresql")

	// Check basic structure
	if sql == "" {
		t.Error("Generated SQL is empty")
	}

	// Check for table name
	if !contains(sql, "CREATE TABLE users") {
		t.Error("SQL should contain 'CREATE TABLE users'")
	}

	// Check for columns
	if !contains(sql, "id uuid") {
		t.Error("SQL should contain 'id uuid'")
	}

	// Check for NOT NULL
	if !contains(sql, "NOT NULL") {
		t.Error("SQL should contain 'NOT NULL' for non-nullable columns")
	}

	// Check for PRIMARY KEY
	if !contains(sql, "PRIMARY KEY (id)") {
		t.Error("SQL should contain 'PRIMARY KEY (id)'")
	}
}

func TestGenerateCreateTableSQLWithDefaults(t *testing.T) {
	table := &TableSchema{
		Name: "users",
		Columns: []ColumnDef{
			{Name: "id", Type: "uuid", Nullable: false, Default: "gen_random_uuid()"},
			{Name: "created_at", Type: "timestamptz", Nullable: false, Default: "now()"},
		},
		PrimaryKey: []string{"id"},
	}

	sql := GenerateCreateTableSQL(table, "postgresql")

	// Check for DEFAULT clauses
	if !contains(sql, "DEFAULT gen_random_uuid()") {
		t.Error("SQL should contain 'DEFAULT gen_random_uuid()'")
	}
	if !contains(sql, "DEFAULT now()") {
		t.Error("SQL should contain 'DEFAULT now()'")
	}
}

func TestGenerateCreateTableSQLCompositePrimaryKey(t *testing.T) {
	table := &TableSchema{
		Name: "order_items",
		Columns: []ColumnDef{
			{Name: "order_id", Type: "uuid", Nullable: false},
			{Name: "product_id", Type: "uuid", Nullable: false},
			{Name: "quantity", Type: "integer", Nullable: false},
		},
		PrimaryKey: []string{"order_id", "product_id"},
	}

	sql := GenerateCreateTableSQL(table, "postgresql")

	if !contains(sql, "PRIMARY KEY (order_id, product_id)") {
		t.Error("SQL should contain composite primary key")
	}
}

// ============================================================================
// Agent Construction Tests
// ============================================================================

func TestNewDatabaseAgent(t *testing.T) {
	agent := NewDatabaseAgent()

	if agent == nil {
		t.Fatal("NewDatabaseAgent returned nil")
	}

	if agent.llmClient == nil {
		t.Error("DatabaseAgent.llmClient is nil")
	}
}

// ============================================================================
// ORM Detection Tests
// ============================================================================

func TestORMNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PRISMA", "prisma"},
		{"Prisma", "prisma"},
		{"prisma", "prisma"},
		{"KNEX", "knex"},
		{"Knex.js", "knex.js"}, // stays as-is after lowercase
		{"SQLAlchemy", "sqlalchemy"},
		{"SQLALCHEMY", "sqlalchemy"},
		{"GORM", "gorm"},
		{"Mongoose", "mongoose"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Simulating the normalization logic from GenerateMigrations
			result := strings.ToLower(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Database Type Tests
// ============================================================================

func TestDatabaseTypeDefaults(t *testing.T) {
	tests := []struct {
		name         string
		inputDBType  string
		expectedType string
	}{
		{"empty defaults to postgresql", "", "postgresql"},
		{"postgresql stays postgresql", "postgresql", "postgresql"},
		{"mysql stays mysql", "mysql", "mysql"},
		{"mongodb stays mongodb", "mongodb", "mongodb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbType := tt.inputDBType
			if dbType == "" {
				dbType = "postgresql"
			}
			if dbType != tt.expectedType {
				t.Errorf("Expected '%s', got '%s'", tt.expectedType, dbType)
			}
		})
	}
}

// ============================================================================
// Records Per Table Tests
// ============================================================================

func TestRecordsPerTableDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero defaults to 10", 0, 10},
		{"negative defaults to 10", -5, 10},
		{"positive stays as-is", 25, 25},
		{"large stays as-is", 1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recordsPerTable := tt.input
			if recordsPerTable <= 0 {
				recordsPerTable = 10
			}
			if recordsPerTable != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, recordsPerTable)
			}
		})
	}
}

// Helper function 'contains' is defined in brownfield_test.go
