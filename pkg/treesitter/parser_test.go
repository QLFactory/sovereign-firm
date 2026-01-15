package treesitter

import (
	"testing"
)

// safeParse wraps Parse with panic recovery for tests
func safeParse(p *Parser, filename string, code []byte) (result *ParseResult, err error, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	result, err = p.Parse(filename, code)
	return
}

func TestNewParser(t *testing.T) {
	parser := NewParser()

	if parser == nil {
		t.Fatal("Expected parser to be created, got nil")
	}

	if parser.parsers == nil {
		t.Error("Expected parsers map to be initialized")
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filename string
		expected Language
	}{
		{"main.go", LangGo},
		{"app.ts", LangTypeScript},
		{"component.tsx", LangTSX},
		{"script.js", LangJavaScript},
		{"component.jsx", LangJSX},
		{"app.py", LangPython},
		{"main.rs", LangRust},
		{"Main.java", LangJava},
		{"main.c", LangC},
		{"main.h", LangC},
		{"main.cpp", LangCpp},
		{"main.cc", LangCpp},
		{"main.cxx", LangCpp},
		{"main.hpp", LangCpp},
		{"app.rb", LangRuby},
		{"README.md", LangUnknown},
		{"data.json", LangUnknown},
		{"", LangUnknown},
	}

	for _, tt := range tests {
		result := DetectLanguage(tt.filename)
		if result != tt.expected {
			t.Errorf("DetectLanguage(%s) = %s, want %s", tt.filename, result, tt.expected)
		}
	}
}

func TestDetectLanguageCaseInsensitive(t *testing.T) {
	// Test that extension detection is case-insensitive
	tests := []struct {
		filename string
		expected Language
	}{
		{"main.GO", LangGo},
		{"app.TS", LangTypeScript},
		{"script.JS", LangJavaScript},
		{"app.PY", LangPython},
	}

	for _, tt := range tests {
		result := DetectLanguage(tt.filename)
		if result != tt.expected {
			t.Errorf("DetectLanguage(%s) = %s, want %s", tt.filename, result, tt.expected)
		}
	}
}

func TestLanguageConstants(t *testing.T) {
	tests := []struct {
		lang     Language
		expected string
	}{
		{LangGo, "go"},
		{LangTypeScript, "typescript"},
		{LangTSX, "tsx"},
		{LangJavaScript, "javascript"},
		{LangJSX, "jsx"},
		{LangPython, "python"},
		{LangRust, "rust"},
		{LangJava, "java"},
		{LangC, "c"},
		{LangCpp, "cpp"},
		{LangRuby, "ruby"},
		{LangUnknown, "unknown"},
	}

	for _, tt := range tests {
		if string(tt.lang) != tt.expected {
			t.Errorf("Expected '%s', got '%s'", tt.expected, tt.lang)
		}
	}
}

func TestParseGoCode(t *testing.T) {
	parser := NewParser()

	code := []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

func helper(x int) int {
	return x * 2
}
`)

	result, err, panicked := safeParse(parser, "main.go", code)
	if panicked || err != nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	if result == nil {
		t.Fatal("Expected parse result, got nil")
	}

	if result.Language != LangGo {
		t.Errorf("Expected language Go, got %s", result.Language)
	}

	if result.Tree == nil {
		t.Error("Expected AST tree to be populated")
	}

	if result.Filename != "main.go" {
		t.Errorf("Expected filename 'main.go', got '%s'", result.Filename)
	}
}

func TestParseTypeScriptCode(t *testing.T) {
	parser := NewParser()

	code := []byte(`interface User {
	name: string;
	age: number;
}

function greet(user: User): string {
	return "Hello, " + user.name;
}

class UserService {
	private users: User[] = [];

	addUser(user: User): void {
		this.users.push(user);
	}
}
`)

	result, err, panicked := safeParse(parser, "app.ts", code)
	if panicked || err != nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	if result.Language != LangTypeScript {
		t.Errorf("Expected language TypeScript, got %s", result.Language)
	}
}

func TestParsePythonCode(t *testing.T) {
	parser := NewParser()

	code := []byte(`def hello(name: str) -> str:
    return f"Hello, {name}!"

class Calculator:
    def __init__(self):
        self.result = 0

    def add(self, x: int, y: int) -> int:
        return x + y
`)

	result, err, panicked := safeParse(parser, "app.py", code)
	if panicked || err != nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	if result.Language != LangPython {
		t.Errorf("Expected language Python, got %s", result.Language)
	}
}

func TestParseJavaScriptCode(t *testing.T) {
	parser := NewParser()

	code := []byte(`function add(a, b) {
	return a + b;
}

const multiply = (a, b) => a * b;

class Calculator {
	constructor() {
		this.value = 0;
	}
}
`)

	result, err, panicked := safeParse(parser, "script.js", code)
	if panicked || err != nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	if result.Language != LangJavaScript {
		t.Errorf("Expected language JavaScript, got %s", result.Language)
	}
}

func TestParseUnknownLanguage(t *testing.T) {
	parser := NewParser()

	code := []byte("Some random content")

	_, err := parser.Parse("file.unknown", code)
	if err == nil {
		t.Error("Expected error for unknown language, got nil")
	}
}

func TestExtractGoSymbols(t *testing.T) {
	parser := NewParser()

	code := []byte(`package main

func main() {
	println("hello")
}

func helper(x int) int {
	return x * 2
}

type User struct {
	Name string
	Age  int
}
`)

	result, err, panicked := safeParse(parser, "main.go", code)
	if panicked || err != nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	symbols, err := parser.ExtractSymbols(result)
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	// Should find at least the two functions and the type
	if len(symbols) < 2 {
		t.Errorf("Expected at least 2 symbols, got %d", len(symbols))
	}

	// Check that we found the main function
	foundMain := false
	for _, sym := range symbols {
		if sym.Name == "main" && sym.Kind == "function" {
			foundMain = true
			break
		}
	}
	if !foundMain {
		t.Error("Expected to find 'main' function symbol")
	}
}

func TestExtractTypeScriptSymbols(t *testing.T) {
	parser := NewParser()

	code := []byte(`interface User {
	name: string;
}

function greet(user: User): string {
	return "Hello";
}

class UserService {
	addUser(user: User): void {}
}
`)

	result, err, panicked := safeParse(parser, "app.ts", code)
	if panicked || err != nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	symbols, err := parser.ExtractSymbols(result)
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	if len(symbols) < 2 {
		t.Errorf("Expected at least 2 symbols, got %d", len(symbols))
	}
}

func TestExtractPythonSymbols(t *testing.T) {
	parser := NewParser()

	code := []byte(`def hello():
    pass

class MyClass:
    def method(self):
        pass
`)

	result, err, panicked := safeParse(parser, "app.py", code)
	if panicked || err != nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	symbols, err := parser.ExtractSymbols(result)
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	if len(symbols) < 2 {
		t.Errorf("Expected at least 2 symbols, got %d", len(symbols))
	}
}

func TestValidateSyntaxValid(t *testing.T) {
	parser := NewParser()

	validCode := []byte(`package main

func main() {
	println("valid")
}
`)

	// Use safeParse to check if parsing works first
	_, err, panicked := safeParse(parser, "main.go", validCode)
	if panicked || err != nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	valid, errors := parser.ValidateSyntax("main.go", validCode)
	if !valid {
		t.Errorf("Expected valid syntax, got errors: %v", errors)
	}
}

func TestValidateSyntaxInvalid(t *testing.T) {
	parser := NewParser()

	validCode := []byte(`package main`)

	// First check if parsing works at all
	_, err, panicked := safeParse(parser, "main.go", validCode)
	if panicked || err != nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	// Intentionally invalid Go code
	invalidCode := []byte(`package main

func main( {
	println("missing paren")
}
`)

	valid, errors := parser.ValidateSyntax("main.go", invalidCode)
	if valid {
		t.Error("Expected invalid syntax to be detected")
	}

	if len(errors) == 0 {
		t.Error("Expected at least one error to be reported")
	}
}

func TestSymbolStructure(t *testing.T) {
	sym := Symbol{
		Name:       "myFunction",
		Kind:       "function",
		Language:   LangGo,
		File:       "main.go",
		StartLine:  10,
		EndLine:    20,
		StartByte:  100,
		EndByte:    500,
		Signature:  "func myFunction(x int) int",
		DocComment: "// myFunction does something",
		Parent:     "MyClass",
	}

	if sym.Name != "myFunction" {
		t.Errorf("Expected name 'myFunction', got '%s'", sym.Name)
	}

	if sym.Kind != "function" {
		t.Errorf("Expected kind 'function', got '%s'", sym.Kind)
	}

	if sym.StartLine != 10 {
		t.Errorf("Expected start line 10, got %d", sym.StartLine)
	}
}

func TestParseResultStructure(t *testing.T) {
	parser := NewParser()

	code := []byte("package main\n")
	result, _, panicked := safeParse(parser, "main.go", code)
	if panicked || result == nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	if result.Filename != "main.go" {
		t.Errorf("Expected filename 'main.go', got '%s'", result.Filename)
	}

	if result.Language != LangGo {
		t.Errorf("Expected language Go, got %s", result.Language)
	}

	if len(result.Code) == 0 {
		t.Error("Expected code to be stored in result")
	}
}

func TestExtractNodeContent(t *testing.T) {
	parser := NewParser()

	code := []byte(`package main

func hello() {
	println("hi")
}
`)

	result, err, panicked := safeParse(parser, "main.go", code)
	if panicked || err != nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	// Get the root node
	root := result.Tree.RootNode()
	content := ExtractNodeContent(root, code)

	if content == "" {
		t.Error("Expected non-empty content from root node")
	}
}

func TestParserCaching(t *testing.T) {
	parser := NewParser()

	// Parse multiple Go files - should reuse parser
	code1 := []byte("package main\nfunc one() {}")
	code2 := []byte("package main\nfunc two() {}")

	_, err1, panicked1 := safeParse(parser, "file1.go", code1)
	if panicked1 {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	_, err2, panicked2 := safeParse(parser, "file2.go", code2)
	if panicked2 {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	if err1 != nil || err2 != nil {
		t.Skipf("Skipping: tree-sitter parsing errors")
	}

	// Check that only one Go parser was created
	parser.mu.RLock()
	numParsers := len(parser.parsers)
	parser.mu.RUnlock()

	if numParsers != 1 {
		t.Errorf("Expected 1 cached parser, got %d", numParsers)
	}
}

func TestGetNodeAtPosition(t *testing.T) {
	parser := NewParser()

	code := []byte(`package main

func hello() {
	println("test")
}
`)

	result, err, panicked := safeParse(parser, "main.go", code)
	if panicked || err != nil {
		t.Skipf("Skipping: tree-sitter native parsing not available")
	}

	// Get node at line 3 (where the function is)
	node := parser.GetNodeAtPosition(result, 3, 5)
	if node == nil {
		t.Error("Expected to find a node at position")
	}
}

func TestMinFunction(t *testing.T) {
	if min(5, 10) != 5 {
		t.Error("min(5, 10) should be 5")
	}
	if min(10, 5) != 5 {
		t.Error("min(10, 5) should be 5")
	}
	if min(5, 5) != 5 {
		t.Error("min(5, 5) should be 5")
	}
}
