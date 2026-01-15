package treesitter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeFiles(t *testing.T) {
	analyzer := NewProjectAnalyzer()
	defer analyzer.Close()

	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello")
}`,
		"app.tsx": `import React from 'react';

export function App() {
	return <div>Hello</div>;
}`,
		"utils.py": `def helper():
    pass

class Util:
    pass`,
	}

	ctx := context.Background()
	analysis, err := analyzer.AnalyzeFiles(ctx, files)
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}

	// Check file count
	if analysis.Summary.TotalFiles != 3 {
		t.Errorf("TotalFiles = %d, want 3", analysis.Summary.TotalFiles)
	}

	// Check languages detected
	if analysis.Languages[LangGo] != 1 {
		t.Errorf("Go files = %d, want 1", analysis.Languages[LangGo])
	}
	if analysis.Languages[LangTSX] != 1 {
		t.Errorf("TSX files = %d, want 1", analysis.Languages[LangTSX])
	}
	if analysis.Languages[LangPython] != 1 {
		t.Errorf("Python files = %d, want 1", analysis.Languages[LangPython])
	}

	// Check functions found
	if analysis.Summary.TotalFunctions < 3 {
		t.Errorf("TotalFunctions = %d, want >= 3", analysis.Summary.TotalFunctions)
	}
}

func TestAnalyzeProject(t *testing.T) {
	// Create a temporary project directory
	tmpDir, err := os.MkdirTemp("", "treesitter-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	pkgJSON := `{
		"name": "test-project",
		"dependencies": {
			"react": "^18.0.0",
			"react-dom": "^18.0.0"
		},
		"devDependencies": {
			"vite": "^5.0.0",
			"vitest": "^1.0.0",
			"tailwindcss": "^3.0.0"
		}
	}`
	err = os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create tsconfig.json
	err = os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(`{}`), 0644)
	if err != nil {
		t.Fatalf("Failed to write tsconfig.json: %v", err)
	}

	// Create source files
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	
	err = os.WriteFile(filepath.Join(srcDir, "App.tsx"), []byte(`export function App() { return <div />; }`), 0644)
	if err != nil {
		t.Fatalf("Failed to write App.tsx: %v", err)
	}

	err = os.WriteFile(filepath.Join(srcDir, "App.test.tsx"), []byte(`import { test } from 'vitest';`), 0644)
	if err != nil {
		t.Fatalf("Failed to write App.test.tsx: %v", err)
	}

	// Analyze project
	analyzer := NewProjectAnalyzer()
	defer analyzer.Close()

	ctx := context.Background()
	stack, err := analyzer.AnalyzeProject(ctx, tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	// Check frameworks detected
	hasReact := false
	hasVite := false
	hasTailwind := false
	for _, fw := range stack.Frameworks {
		if fw == FrameworkReact {
			hasReact = true
		}
		if fw == FrameworkTailwind {
			hasTailwind = true
		}
	}
	for _, tool := range stack.BuildTools {
		if tool == FrameworkVite {
			hasVite = true
		}
	}

	if !hasReact {
		t.Error("Expected to detect React framework")
	}
	if !hasVite {
		t.Error("Expected to detect Vite build tool")
	}
	if !hasTailwind {
		t.Error("Expected to detect Tailwind CSS")
	}

	// Check test framework
	hasVitest := false
	for _, tf := range stack.TestFrameworks {
		if tf == FrameworkVitest {
			hasVitest = true
		}
	}
	if !hasVitest {
		t.Error("Expected to detect Vitest test framework")
	}

	// Check TypeScript detection
	if !stack.HasTypeScript {
		t.Error("Expected HasTypeScript to be true")
	}

	// Check test files detected
	if !stack.HasTests {
		t.Error("Expected HasTests to be true")
	}

	// Check languages
	hasTSX := false
	for _, lang := range stack.Languages {
		if lang == LangTypeScript || lang == LangTSX {
			hasTSX = true
		}
	}
	if !hasTSX {
		t.Error("Expected TypeScript/TSX in languages")
	}
}

func TestDetectPackageManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		file     string
		expected string
	}{
		{"npm", "package-lock.json", "npm"},
		{"yarn", "yarn.lock", "yarn"},
		{"pnpm", "pnpm-lock.yaml", "pnpm"},
		{"bun", "bun.lockb", "bun"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up previous files
			for _, f := range []string{"package-lock.json", "yarn.lock", "pnpm-lock.yaml", "bun.lockb"} {
				os.Remove(filepath.Join(tmpDir, f))
			}

			// Create the lock file
			err := os.WriteFile(filepath.Join(tmpDir, tt.file), []byte(""), 0644)
			if err != nil {
				t.Fatalf("Failed to create %s: %v", tt.file, err)
			}

			analyzer := NewProjectAnalyzer()
			defer analyzer.Close()

			ctx := context.Background()
			stack, err := analyzer.AnalyzeProject(ctx, tmpDir)
			if err != nil {
				t.Fatalf("AnalyzeProject failed: %v", err)
			}

			if stack.PackageManager != tt.expected {
				t.Errorf("PackageManager = %q, want %q", stack.PackageManager, tt.expected)
			}
		})
	}
}

func TestDetectGoFrameworks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-fw-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goMod := `module test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.0
)
`
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	analyzer := NewProjectAnalyzer()
	defer analyzer.Close()

	ctx := context.Background()
	stack, err := analyzer.AnalyzeProject(ctx, tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	hasGin := false
	for _, fw := range stack.Frameworks {
		if fw == FrameworkGin {
			hasGin = true
		}
	}
	if !hasGin {
		t.Error("Expected to detect Gin framework")
	}

	hasGo := false
	for _, lang := range stack.Languages {
		if lang == LangGo {
			hasGo = true
		}
	}
	if !hasGo {
		t.Error("Expected Go in languages")
	}
}
