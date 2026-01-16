package activities

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/qlfactory/sovereign-firm/pkg/treesitter"
)

func TestBrownfieldParams(t *testing.T) {
	params := BrownfieldParams{
		LocalPath:     "/tmp/project",
		ProjectID:     "proj-123",
		ClientID:      "client-456",
		MaxFiles:      100,
		IncludeTests:  true,
		StoreGlobally: false,
	}

	if params.ProjectID != "proj-123" {
		t.Error("Expected ProjectID to be set")
	}

	if params.ClientID != "client-456" {
		t.Error("Expected ClientID to be set")
	}

	if params.MaxFiles != 100 {
		t.Error("Expected MaxFiles 100")
	}

	if !params.IncludeTests {
		t.Error("Expected IncludeTests to be true")
	}
}

func TestBrownfieldResult(t *testing.T) {
	result := BrownfieldResult{
		ProjectID:       "proj-123",
		FilesIndexed:    50,
		ChunksCreated:   200,
		SymbolsFound:    150,
		Decisions:       []string{"Use TypeScript"},
		Recommendations: []string{"Add tests"},
	}

	if result.FilesIndexed != 50 {
		t.Error("Expected FilesIndexed 50")
	}

	if result.ChunksCreated != 200 {
		t.Error("Expected ChunksCreated 200")
	}

	if len(result.Decisions) != 1 {
		t.Error("Expected 1 decision")
	}

	if len(result.Recommendations) != 1 {
		t.Error("Expected 1 recommendation")
	}
}

func TestShouldSkipDir(t *testing.T) {
	b := &BrownfieldAnalyzer{}

	tests := []struct {
		name     string
		expected bool
	}{
		{"node_modules", true},
		{".git", true},
		{"vendor", true},
		{"__pycache__", true},
		{"dist", true},
		{"build", true},
		{"src", false},
		{"lib", false},
		{"components", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := b.shouldSkipDir(tt.name, nil)
			if result != tt.expected {
				t.Errorf("shouldSkipDir(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestShouldSkipDirCustomPatterns(t *testing.T) {
	b := &BrownfieldAnalyzer{}

	skipPatterns := []string{"temp*", "*.bak"}

	if !b.shouldSkipDir("temp123", skipPatterns) {
		t.Error("Expected to skip temp123")
	}

	if b.shouldSkipDir("important", skipPatterns) {
		t.Error("Should not skip 'important'")
	}
}

func TestIsTestFile(t *testing.T) {
	b := &BrownfieldAnalyzer{}

	tests := []struct {
		path     string
		expected bool
	}{
		{"main_test.go", true},
		{"app.test.ts", true},
		{"component.spec.tsx", true},
		{"test_utils.py", true},
		{"main.go", false},
		{"app.ts", false},
		{"testing.go", false}, // Has "test" in name but not a test pattern
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := b.isTestFile(tt.path)
			if result != tt.expected {
				t.Errorf("isTestFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDescribeStack(t *testing.T) {
	b := &BrownfieldAnalyzer{}

	// Test with nil stack
	desc := b.describeStack(nil)
	if desc != "Unable to detect technology stack" {
		t.Error("Expected message for nil stack")
	}

	// Test with populated stack
	stack := &treesitter.ProjectStack{
		PrimaryLanguage: treesitter.LangTypeScript,
		Frameworks:      []treesitter.Framework{treesitter.FrameworkReact, treesitter.FrameworkNextJS},
		BuildTools:      []treesitter.Framework{treesitter.FrameworkVite},
		TestFrameworks:  []treesitter.Framework{treesitter.FrameworkVitest},
		PackageManager:  "npm",
		HasTypeScript:   true,
		IsMonorepo:      false,
	}

	desc = b.describeStack(stack)

	if desc == "" {
		t.Error("Expected non-empty description")
	}

	// Check that key info is included
	if !contains(desc, "typescript") && !contains(desc, "TypeScript") {
		t.Error("Expected TypeScript to be mentioned")
	}

	if !contains(desc, "react") && !contains(desc, "React") {
		t.Error("Expected React to be mentioned")
	}

	if !contains(desc, "npm") {
		t.Error("Expected npm to be mentioned")
	}
}

func TestDescribeStackMonorepo(t *testing.T) {
	b := &BrownfieldAnalyzer{}

	stack := &treesitter.ProjectStack{
		PrimaryLanguage: treesitter.LangTypeScript,
		IsMonorepo:      true,
	}

	desc := b.describeStack(stack)

	if !contains(desc, "Monorepo") {
		t.Error("Expected Monorepo to be mentioned")
	}
}

func TestGenerateRecommendations(t *testing.T) {
	b := &BrownfieldAnalyzer{}

	// Test with nil stack
	recs := b.generateRecommendations(nil, 10, 50)
	if len(recs) == 0 {
		t.Error("Expected at least one recommendation for nil stack")
	}

	// Test with JS project without TypeScript
	stack := &treesitter.ProjectStack{
		PrimaryLanguage: treesitter.LangJavaScript,
		HasTypeScript:   false,
		HasTests:        false,
	}

	recs = b.generateRecommendations(stack, 10, 50)

	foundTSRec := false
	foundTestRec := false
	for _, rec := range recs {
		if contains(rec, "TypeScript") {
			foundTSRec = true
		}
		if contains(rec, "test") {
			foundTestRec = true
		}
	}

	if !foundTSRec {
		t.Error("Expected TypeScript recommendation")
	}

	if !foundTestRec {
		t.Error("Expected test recommendation")
	}

	// Test with large codebase
	recs = b.generateRecommendations(stack, 150, 600)

	foundLargeCodebase := false
	foundManySymbols := false
	for _, rec := range recs {
		if contains(rec, "Large codebase") {
			foundLargeCodebase = true
		}
		if contains(rec, "Many symbols") {
			foundManySymbols = true
		}
	}

	if !foundLargeCodebase {
		t.Error("Expected large codebase recommendation")
	}

	if !foundManySymbols {
		t.Error("Expected many symbols recommendation")
	}
}

func TestHasFramework(t *testing.T) {
	frameworks := []treesitter.Framework{
		treesitter.FrameworkReact,
		treesitter.FrameworkVite,
	}

	if !hasFramework(frameworks, treesitter.FrameworkReact) {
		t.Error("Expected to find React")
	}

	if !hasFramework(frameworks, treesitter.FrameworkVite) {
		t.Error("Expected to find Vite")
	}

	if hasFramework(frameworks, treesitter.FrameworkNextJS) {
		t.Error("Should not find NextJS")
	}
}

func TestGetProjectDirLocal(t *testing.T) {
	b := &BrownfieldAnalyzer{}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "brownfield-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	params := BrownfieldParams{
		LocalPath: tmpDir,
	}

	dir, cleanup, err := b.getProjectDir(params)
	if err != nil {
		t.Fatalf("getProjectDir failed: %v", err)
	}

	if cleanup != nil {
		t.Error("Expected no cleanup for local path")
	}

	if dir != tmpDir {
		t.Errorf("Expected dir %s, got %s", tmpDir, dir)
	}
}

func TestGetProjectDirInvalidLocal(t *testing.T) {
	b := &BrownfieldAnalyzer{}

	params := BrownfieldParams{
		LocalPath: "/nonexistent/path/that/does/not/exist",
	}

	_, _, err := b.getProjectDir(params)
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}

func TestGetProjectDirNoSource(t *testing.T) {
	b := &BrownfieldAnalyzer{}

	params := BrownfieldParams{
		// Neither RepoURL nor LocalPath
		ProjectID: "test",
	}

	_, _, err := b.getProjectDir(params)
	if err == nil {
		t.Error("Expected error when no source provided")
	}
}

func TestExtractFunctionContent(t *testing.T) {
	b := &BrownfieldAnalyzer{}

	content := []byte(`package main

func hello() {
	println("hello")
}

func world() {
	println("world")
}
`)

	fn := treesitter.Symbol{
		Name:      "hello",
		StartLine: 3,
		EndLine:   5,
	}

	extracted := b.extractFunctionContent(content, fn)

	if extracted == "" {
		t.Error("Expected extracted content")
	}

	if !contains(extracted, "func hello()") {
		t.Error("Expected function declaration in extracted content")
	}
}

func TestAnalyzeProjectLocal(t *testing.T) {
	// Skip if no ChromaDB available (integration test)
	if os.Getenv("CHROMA_HOST") == "" {
		t.Skip("Skipping integration test - CHROMA_HOST not set")
	}

	// Create temp project
	tmpDir, err := os.MkdirTemp("", "brownfield-analyze-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create sample files
	mainGo := `package main

func main() {
	println("hello")
}
`
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644)

	b := NewBrownfieldAnalyzer("")
	defer b.Close()

	ctx := context.Background()
	result, err := b.AnalyzeProject(ctx, BrownfieldParams{
		LocalPath: tmpDir,
		ProjectID: "test-project",
		ClientID:  "test-client",
	})

	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	if result.ProjectID != "test-project" {
		t.Error("Expected project ID to be set")
	}

	if result.FilesIndexed < 1 {
		t.Error("Expected at least one file indexed")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
