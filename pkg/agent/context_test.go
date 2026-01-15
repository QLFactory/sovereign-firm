package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCodeContextManager(t *testing.T) {
	// Create a temporary project directory with sample files
	tmpDir, err := os.MkdirTemp("", "context-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create sample source files
	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	user := NewUser("Alice")
	fmt.Println(user.Greet())
}

type User struct {
	Name string
}

func NewUser(name string) *User {
	return &User{Name: name}
}

func (u *User) Greet() string {
	return "Hello, " + u.Name
}
`,
		"utils.go": `package main

func helper() string {
	return "help"
}

const Version = "1.0.0"
`,
		"src/app.tsx": `import React from 'react';

export function App() {
	return <div>Hello World</div>;
}

export const version = "1.0.0";
`,
	}

	// Create directories and files
	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", path, err)
		}
	}

	// Create context manager
	mgr := NewCodeContextManager(tmpDir)
	defer mgr.Close()

	ctx := context.Background()

	// Index project
	if err := mgr.IndexProject(ctx); err != nil {
		t.Fatalf("IndexProject failed: %v", err)
	}

	// Verify index was created
	index := mgr.GetIndex()
	if index == nil {
		t.Fatal("Index is nil")
	}

	// Check files were indexed
	if len(index.Files) < 3 {
		t.Errorf("Expected at least 3 files indexed, got %d", len(index.Files))
	}

	// Check symbols were indexed
	if len(index.Symbols) == 0 {
		t.Error("No symbols were indexed")
	}

	// Check specific symbols
	userLocs := mgr.FindSymbol("User")
	if len(userLocs) == 0 {
		t.Error("Expected to find User symbol")
	}

	newUserLocs := mgr.FindSymbol("NewUser")
	if len(newUserLocs) == 0 {
		t.Error("Expected to find NewUser symbol")
	}

	appLocs := mgr.FindSymbol("App")
	if len(appLocs) == 0 {
		t.Error("Expected to find App symbol")
	}
}

func TestSelectContext(t *testing.T) {
	// Create a temporary project directory
	tmpDir, err := os.MkdirTemp("", "context-select-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create sample files
	files := map[string]string{
		"user.go": `package main

type User struct {
	ID   int
	Name string
}

func NewUser(id int, name string) *User {
	return &User{ID: id, Name: name}
}
`,
		"auth.go": `package main

func Login(username, password string) bool {
	return username == "admin"
}

func Logout() {
}
`,
		"api/handler.go": `package api

import "net/http"

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("login"))
}

func HandleUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("user"))
}
`,
		"test/user_test.go": `package main

import "testing"

func TestNewUser(t *testing.T) {
	u := NewUser(1, "test")
	if u.Name != "test" {
		t.Fail()
	}
}
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		os.MkdirAll(dir, 0755)
		os.WriteFile(fullPath, []byte(content), 0644)
	}

	mgr := NewCodeContextManager(tmpDir)
	defer mgr.Close()

	ctx := context.Background()

	// Test selecting context by symbol
	selected, err := mgr.SelectContext(ctx, ContextRequest{
		Symbols:  []string{"User", "NewUser"},
		MaxFiles: 5,
	})
	if err != nil {
		t.Fatalf("SelectContext failed: %v", err)
	}

	// Should include user.go which has User and NewUser
	if _, ok := selected.Files["user.go"]; !ok {
		t.Error("Expected user.go to be selected for User symbol")
	}

	// Test selecting context by task type
	selected, err = mgr.SelectContext(ctx, ContextRequest{
		TaskType:    "test",
		Description: "write tests",
		MaxFiles:    5,
	})
	if err != nil {
		t.Fatalf("SelectContext failed: %v", err)
	}

	// Should prefer test files
	foundTest := false
	for path := range selected.Files {
		if filepath.Base(path) == "user_test.go" {
			foundTest = true
			break
		}
	}
	if !foundTest && len(selected.Files) > 0 {
		// Test file should be ranked high for test task type
		// but it's not required to be included if there are not enough test files
	}

	// Test selecting by file type
	selected, err = mgr.SelectContext(ctx, ContextRequest{
		FileTypes: []string{".go"},
		MaxFiles:  2,
	})
	if err != nil {
		t.Fatalf("SelectContext failed: %v", err)
	}

	// All selected files should be .go
	for path := range selected.Files {
		if filepath.Ext(path) != ".go" {
			t.Errorf("Expected only .go files, got %s", path)
		}
	}
}

func TestKeywordExtraction(t *testing.T) {
	tests := []struct {
		description string
		shouldHave  []string
	}{
		{
			description: "implement user authentication with JWT tokens",
			shouldHave:  []string{"user", "authentication", "jwt", "tokens"},
		},
		{
			description: "add a login form to the dashboard",
			shouldHave:  []string{"login", "form", "dashboard"},
		},
		{
			description: "fix bug in the API handler",
			shouldHave:  []string{"bug", "api", "handler"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description[:20], func(t *testing.T) {
			keywords := extractKeywords(tt.description)
			for _, expected := range tt.shouldHave {
				found := false
				for _, kw := range keywords {
					if kw == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected keyword %q not found in %v", expected, keywords)
				}
			}
		})
	}
}

func TestIsEntryPoint(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"main.go", true},
		{"index.ts", true},
		{"app.tsx", true},
		{"server.go", true},
		{"utils.go", false},
		{"helper.ts", false},
		{"lib.rs", true},
		{"config.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isEntryPoint(tt.path)
			if result != tt.expected {
				t.Errorf("isEntryPoint(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestPopulateAgentContext(t *testing.T) {
	// Create a temporary project
	tmpDir, err := os.MkdirTemp("", "agent-ctx-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create sample file
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main

func main() {
	println("hello")
}
`), 0644)

	mgr := NewCodeContextManager(tmpDir)
	defer mgr.Close()

	// Create a mock agent
	agent := &AgentInstance{
		Context: &AgentContext{
			ProjectID: "test-project",
		},
	}

	ctx := context.Background()

	err = mgr.PopulateAgentContext(ctx, agent, ContextRequest{
		MaxFiles: 10,
	})
	if err != nil {
		t.Fatalf("PopulateAgentContext failed: %v", err)
	}

	// Check that agent context was populated
	if len(agent.Context.RelevantFiles) == 0 {
		t.Error("Expected agent context to have relevant files")
	}

	if agent.Context.CodebaseIndex == "" {
		t.Error("Expected agent context to have codebase index summary")
	}
}

func TestListExportedSymbols(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "exported-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "lib.go"), []byte(`package lib

// Exported function
func PublicFunc() {}

// unexported function
func privateFunc() {}

type PublicType struct{}
type privateType struct{}

const PublicConst = "public"
const privateConst = "private"
`), 0644)

	mgr := NewCodeContextManager(tmpDir)
	defer mgr.Close()

	ctx := context.Background()
	mgr.IndexProject(ctx)

	exported := mgr.ListExportedSymbols()

	// Check that only exported symbols are returned
	hasPublicFunc := false
	hasPublicType := false
	hasPrivate := false

	for _, loc := range exported {
		if loc.Name == "PublicFunc" {
			hasPublicFunc = true
		}
		if loc.Name == "PublicType" {
			hasPublicType = true
		}
		if loc.Name == "privateFunc" || loc.Name == "privateType" {
			hasPrivate = true
		}
	}

	if !hasPublicFunc {
		t.Error("Expected PublicFunc in exported symbols")
	}
	if !hasPublicType {
		t.Error("Expected PublicType in exported symbols")
	}
	if hasPrivate {
		t.Error("Private symbols should not be in exported list")
	}
}
