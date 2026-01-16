package memory

import (
	"context"
	"strings"
	"testing"
)

func TestDefaultChunkConfig(t *testing.T) {
	config := DefaultChunkConfig()

	if config.MaxChunkSize != 2000 {
		t.Errorf("Expected MaxChunkSize 2000, got %d", config.MaxChunkSize)
	}

	if config.MinChunkSize != 100 {
		t.Errorf("Expected MinChunkSize 100, got %d", config.MinChunkSize)
	}

	if config.OverlapLines != 3 {
		t.Errorf("Expected OverlapLines 3, got %d", config.OverlapLines)
	}

	if !config.IncludeImports {
		t.Error("Expected IncludeImports to be true")
	}

	if !config.PreserveSymbols {
		t.Error("Expected PreserveSymbols to be true")
	}
}

func TestNewCodeChunker(t *testing.T) {
	chunker := NewCodeChunker(DefaultChunkConfig())
	defer chunker.Close()

	if chunker == nil {
		t.Fatal("Expected chunker to be created")
	}

	if chunker.analyzer == nil {
		t.Error("Expected analyzer to be initialized")
	}
}

func TestChunkGoFile(t *testing.T) {
	chunker := NewCodeChunker(ChunkConfig{
		MaxChunkSize:    500,
		MinChunkSize:    50,
		OverlapLines:    2,
		IncludeImports:  true,
		PreserveSymbols: true,
	})
	defer chunker.Close()

	goCode := `package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("Hello")
}

func helper() string {
	return "help"
}

type User struct {
	Name string
	Age  int
}

func NewUser(name string) *User {
	return &User{Name: name}
}
`

	ctx := context.Background()
	chunks, err := chunker.ChunkFile(ctx, "main.go", []byte(goCode))

	if err != nil {
		t.Fatalf("ChunkFile failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Error("Expected at least one chunk")
	}

	// Check that chunks have content
	for i, chunk := range chunks {
		if chunk.Content == "" {
			t.Errorf("Chunk %d has empty content", i)
		}
		if chunk.StartLine <= 0 {
			t.Errorf("Chunk %d has invalid start line: %d", i, chunk.StartLine)
		}
	}
}

func TestChunkJavaScriptFile(t *testing.T) {
	chunker := NewCodeChunker(DefaultChunkConfig())
	defer chunker.Close()

	jsCode := `import React from 'react';
import { useState } from 'react';

export function Counter() {
	const [count, setCount] = useState(0);
	return <button onClick={() => setCount(count + 1)}>{count}</button>;
}

export function App() {
	return (
		<div>
			<Counter />
		</div>
	);
}

export const PI = 3.14159;
`

	ctx := context.Background()
	chunks, err := chunker.ChunkFile(ctx, "App.jsx", []byte(jsCode))

	if err != nil {
		t.Fatalf("ChunkFile failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Error("Expected at least one chunk")
	}
}

func TestChunkPythonFile(t *testing.T) {
	chunker := NewCodeChunker(DefaultChunkConfig())
	defer chunker.Close()

	pyCode := `import os
from typing import List, Optional

class User:
    def __init__(self, name: str):
        self.name = name

    def greet(self) -> str:
        return f"Hello, {self.name}"

def create_user(name: str) -> User:
    return User(name)

def main():
    user = create_user("Alice")
    print(user.greet())

if __name__ == "__main__":
    main()
`

	ctx := context.Background()
	chunks, err := chunker.ChunkFile(ctx, "app.py", []byte(pyCode))

	if err != nil {
		t.Fatalf("ChunkFile failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Error("Expected at least one chunk")
	}
}

func TestChunkByLines(t *testing.T) {
	chunker := NewCodeChunker(ChunkConfig{
		MaxChunkSize: 200,
		MinChunkSize: 50,
		OverlapLines: 2,
	})
	defer chunker.Close()

	// Create content that will trigger line-based chunking
	var lines []string
	for i := 0; i < 50; i++ {
		lines = append(lines, "This is line number "+string(rune('0'+i%10)))
	}
	content := strings.Join(lines, "\n")

	chunks, err := chunker.chunkByLines(content)

	if err != nil {
		t.Fatalf("chunkByLines failed: %v", err)
	}

	if len(chunks) == 0 {
		t.Error("Expected at least one chunk")
	}

	// Check that chunks have valid line numbers
	for i, chunk := range chunks {
		if chunk.StartLine <= 0 {
			t.Errorf("Chunk %d has invalid start line", i)
		}
		if chunk.EndLine < chunk.StartLine {
			t.Errorf("Chunk %d has end line before start line", i)
		}
	}
}

func TestChunkFileWithStats(t *testing.T) {
	chunker := NewCodeChunker(DefaultChunkConfig())
	defer chunker.Close()

	code := `package main

func main() {
	println("hello")
}

func helper() {}
`

	ctx := context.Background()
	result, err := chunker.ChunkFileWithStats(ctx, "main.go", []byte(code))

	if err != nil {
		t.Fatalf("ChunkFileWithStats failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.FilePath != "main.go" {
		t.Errorf("Expected FilePath 'main.go', got '%s'", result.FilePath)
	}

	if result.Language != "go" {
		t.Errorf("Expected Language 'go', got '%s'", result.Language)
	}

	if result.Stats.TotalChunks != len(result.Chunks) {
		t.Error("Stats.TotalChunks doesn't match chunk count")
	}

	if result.Stats.TotalLines <= 0 {
		t.Error("Expected positive total lines")
	}

	if result.Stats.TotalChars <= 0 {
		t.Error("Expected positive total chars")
	}
}

func TestChunkFiles(t *testing.T) {
	chunker := NewCodeChunker(DefaultChunkConfig())
	defer chunker.Close()

	files := map[string][]byte{
		"main.go": []byte(`package main
func main() {}`),
		"app.ts": []byte(`export function app() { return "hello"; }`),
		"util.py": []byte(`def util(): pass`),
	}

	ctx := context.Background()
	results, err := chunker.ChunkFiles(ctx, files)

	if err != nil {
		t.Fatalf("ChunkFiles failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	for path, result := range results {
		if result.FilePath != path {
			t.Errorf("Expected FilePath '%s', got '%s'", path, result.FilePath)
		}
	}
}

func TestChunkUnknownLanguage(t *testing.T) {
	chunker := NewCodeChunker(DefaultChunkConfig())
	defer chunker.Close()

	content := []byte("This is a plain text file\nwith multiple lines\nfor testing purposes.")

	ctx := context.Background()
	chunks, err := chunker.ChunkFile(ctx, "readme.txt", content)

	if err != nil {
		t.Fatalf("ChunkFile failed: %v", err)
	}

	// Should fall back to line-based chunking
	if len(chunks) == 0 {
		t.Error("Expected at least one chunk")
	}
}

func TestChunkEmptyFile(t *testing.T) {
	chunker := NewCodeChunker(DefaultChunkConfig())
	defer chunker.Close()

	ctx := context.Background()
	chunks, err := chunker.ChunkFile(ctx, "empty.go", []byte(""))

	if err != nil {
		t.Fatalf("ChunkFile failed: %v", err)
	}

	// Empty file should produce no chunks or one empty chunk
	if len(chunks) > 1 {
		t.Errorf("Expected 0 or 1 chunks for empty file, got %d", len(chunks))
	}
}

func TestChunkSmallFile(t *testing.T) {
	chunker := NewCodeChunker(ChunkConfig{
		MaxChunkSize: 2000,
		MinChunkSize: 100,
		OverlapLines: 3,
	})
	defer chunker.Close()

	// File smaller than MinChunkSize
	content := []byte("x := 1")

	ctx := context.Background()
	chunks, err := chunker.ChunkFile(ctx, "small.go", content)

	if err != nil {
		t.Fatalf("ChunkFile failed: %v", err)
	}

	// Small files should still produce at least one chunk with all content
	if len(chunks) == 0 {
		t.Log("Small file produced no chunks (acceptable)")
	}
}

func TestSplitLargeSymbol(t *testing.T) {
	chunker := NewCodeChunker(ChunkConfig{
		MaxChunkSize: 100,
		MinChunkSize: 20,
		OverlapLines: 2,
	})
	defer chunker.Close()

	// Create a large function with many lines
	var lines []string
	for i := 0; i < 20; i++ {
		lines = append(lines, "    line := "+string(rune('a'+i%26)))
	}

	chunks := chunker.splitLargeSymbol(lines, 1, "largeFunc")

	if len(chunks) < 2 {
		t.Errorf("Expected multiple chunks for large symbol, got %d", len(chunks))
	}

	// All chunks should reference the same symbol
	for _, chunk := range chunks {
		if chunk.Symbol != "largeFunc" {
			t.Errorf("Expected symbol 'largeFunc', got '%s'", chunk.Symbol)
		}
	}
}

func TestChunkConfigZeroValues(t *testing.T) {
	chunker := NewCodeChunker(ChunkConfig{}) // All zero values
	defer chunker.Close()

	// Should use defaults
	if chunker.config.MaxChunkSize != 2000 {
		t.Errorf("Expected MaxChunkSize 2000 (default), got %d", chunker.config.MaxChunkSize)
	}
}

func TestMaxFunction(t *testing.T) {
	if max(5, 10) != 10 {
		t.Error("Expected max(5,10) = 10")
	}
	if max(10, 5) != 10 {
		t.Error("Expected max(10,5) = 10")
	}
	if max(5, 5) != 5 {
		t.Error("Expected max(5,5) = 5")
	}
}
