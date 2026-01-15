package treesitter

import (
	"context"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		source   []byte
		expected Language
	}{
		{"JavaScript", "app.js", nil, LangJavaScript},
		{"TypeScript", "app.ts", nil, LangTypeScript},
		{"TSX", "App.tsx", nil, LangTSX},
		{"JSX", "App.jsx", nil, LangJSX},
		{"Go", "main.go", nil, LangGo},
		{"Python", "script.py", nil, LangPython},
		{"Rust", "lib.rs", nil, LangRust},
		{"HTML", "index.html", nil, LangHTML},
		{"CSS", "styles.css", nil, LangCSS},
		{"JSON", "package.json", nil, LangJSON},
		{"YAML", "config.yaml", nil, LangYAML},
		{"YAML alternate", "docker-compose.yml", nil, LangYAML},
		{"TOML", "Cargo.toml", nil, LangTOML},
		{"Bash", "script.sh", nil, LangBash},
		{"Markdown", "README.md", nil, LangMarkdown},
		{"Unknown", "file.xyz", nil, LangUnknown},
		// JSX detection from content
		{"JS with JSX", "app.js", []byte(`import React from 'react'`), LangJSX},
		{"JS with className", "app.js", []byte(`<div className="test" />`), LangJSX},
		// Shebang detection
		{"Python shebang", "script", []byte("#!/usr/bin/env python3\nprint('hello')"), LangPython},
		{"Node shebang", "script", []byte("#!/usr/bin/env node\nconsole.log('hi')"), LangJavaScript},
		{"Bash shebang", "script", []byte("#!/bin/bash\necho hello"), LangBash},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguage(tt.filePath, tt.source)
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestParserParseJavaScript(t *testing.T) {
	p := NewParser()
	defer p.Close()

	source := []byte(`
function hello(name) {
	return "Hello, " + name;
}

const greeting = hello("World");
console.log(greeting);
`)

	ctx := context.Background()
	result, err := p.Parse(ctx, "test.js", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Language != LangJavaScript {
		t.Errorf("Language = %q, want %q", result.Language, LangJavaScript)
	}

	if result.Tree == nil {
		t.Fatal("Tree is nil")
	}

	root := result.Tree.RootNode()
	if root.Type() != "program" {
		t.Errorf("Root type = %q, want %q", root.Type(), "program")
	}
}

func TestParserParseTypeScript(t *testing.T) {
	p := NewParser()
	defer p.Close()

	source := []byte(`
interface User {
	name: string;
	age: number;
}

function greet(user: User): string {
	return "Hello, " + user.name;
}
`)

	ctx := context.Background()
	result, err := p.Parse(ctx, "test.ts", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Language != LangTypeScript {
		t.Errorf("Language = %q, want %q", result.Language, LangTypeScript)
	}
}

func TestParserParseGo(t *testing.T) {
	p := NewParser()
	defer p.Close()

	source := []byte(`
package main

import "fmt"

type User struct {
	Name string
	Age  int
}

func (u *User) Greet() string {
	return fmt.Sprintf("Hello, %s", u.Name)
}

func main() {
	user := &User{Name: "World", Age: 30}
	fmt.Println(user.Greet())
}
`)

	ctx := context.Background()
	result, err := p.Parse(ctx, "main.go", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Language != LangGo {
		t.Errorf("Language = %q, want %q", result.Language, LangGo)
	}

	root := result.Tree.RootNode()
	if root.Type() != "source_file" {
		t.Errorf("Root type = %q, want %q", root.Type(), "source_file")
	}
}

func TestParserParsePython(t *testing.T) {
	p := NewParser()
	defer p.Close()

	source := []byte(`
class User:
    def __init__(self, name: str, age: int):
        self.name = name
        self.age = age
    
    def greet(self) -> str:
        return f"Hello, {self.name}"

def main():
    user = User("World", 30)
    print(user.greet())

if __name__ == "__main__":
    main()
`)

	ctx := context.Background()
	result, err := p.Parse(ctx, "main.py", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Language != LangPython {
		t.Errorf("Language = %q, want %q", result.Language, LangPython)
	}
}

func TestSupportedLanguages(t *testing.T) {
	langs := SupportedLanguages()
	if len(langs) == 0 {
		t.Error("SupportedLanguages returned empty list")
	}

	// Check that common languages are supported
	expected := []Language{LangJavaScript, LangTypeScript, LangGo, LangPython}
	for _, exp := range expected {
		found := false
		for _, lang := range langs {
			if lang == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected language %q not in supported list", exp)
		}
	}
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		lang     Language
		expected bool
	}{
		{LangJavaScript, true},
		{LangTypeScript, true},
		{LangGo, true},
		{LangPython, true},
		{LangUnknown, false},
		{Language("cobol"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			result := IsSupported(tt.lang)
			if result != tt.expected {
				t.Errorf("IsSupported(%q) = %v, want %v", tt.lang, result, tt.expected)
			}
		})
	}
}

func TestParserConcurrency(t *testing.T) {
	p := NewParser()
	defer p.Close()

	// Parse multiple files concurrently
	files := map[string][]byte{
		"test.js": []byte(`function foo() { return 1; }`),
		"test.ts": []byte(`const x: number = 1;`),
		"test.go": []byte(`package main; func main() {}`),
		"test.py": []byte(`def foo(): pass`),
	}

	ctx := context.Background()
	done := make(chan bool)

	for path, source := range files {
		go func(filePath string, src []byte) {
			_, err := p.Parse(ctx, filePath, src)
			if err != nil {
				t.Errorf("Parse(%q) failed: %v", filePath, err)
			}
			done <- true
		}(path, source)
	}

	for range files {
		<-done
	}
}
