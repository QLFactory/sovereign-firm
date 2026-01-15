package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/treesitter"
)

// CodeContextManager provides intelligent code context selection for agents
// using tree-sitter for code analysis and symbol extraction.
type CodeContextManager struct {
	analyzer   *treesitter.ProjectAnalyzer
	projectDir string
	index      *CodebaseIndex
}

// CodebaseIndex holds the parsed structure of a codebase
type CodebaseIndex struct {
	ProjectDir string                             `json:"project_dir"`
	Stack      *treesitter.ProjectStack           `json:"stack"`
	Files      map[string]*treesitter.CodeStructure `json:"files"`
	Symbols    map[string][]SymbolLocation        `json:"symbols"` // symbol name -> locations
}

// SymbolLocation identifies where a symbol is defined
type SymbolLocation struct {
	FilePath  string `json:"file_path"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	StartLine uint32 `json:"start_line"`
	EndLine   uint32 `json:"end_line"`
	Exported  bool   `json:"exported"`
}

// NewCodeContextManager creates a new context manager for a project
func NewCodeContextManager(projectDir string) *CodeContextManager {
	return &CodeContextManager{
		analyzer:   treesitter.NewProjectAnalyzer(),
		projectDir: projectDir,
	}
}

// Close releases resources
func (m *CodeContextManager) Close() {
	if m.analyzer != nil {
		m.analyzer.Close()
	}
}

// IndexProject analyzes and indexes the entire project
func (m *CodeContextManager) IndexProject(ctx context.Context) error {
	// Analyze project stack (frameworks, languages)
	stack, err := m.analyzer.AnalyzeProject(ctx, m.projectDir)
	if err != nil {
		return err
	}

	m.index = &CodebaseIndex{
		ProjectDir: m.projectDir,
		Stack:      stack,
		Files:      make(map[string]*treesitter.CodeStructure),
		Symbols:    make(map[string][]SymbolLocation),
	}

	// Walk through source files and index them
	err = filepath.Walk(m.projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories and non-source files
		if info.IsDir() {
			// Skip common non-source directories
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" || base == "vendor" ||
				base == "__pycache__" || base == ".venv" || base == "dist" ||
				base == "build" || base == "target" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's a supported language
		lang := treesitter.DetectLanguage(path, nil)
		if lang == treesitter.LangUnknown || !treesitter.IsSupported(lang) {
			return nil
		}

		// Read and parse file
		source, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable files
		}

		structure, err := m.analyzer.AnalyzeFile(ctx, path, source)
		if err != nil {
			return nil // Skip unparseable files
		}

		// Store relative path
		relPath, _ := filepath.Rel(m.projectDir, path)
		m.index.Files[relPath] = structure

		// Index symbols
		m.indexSymbols(relPath, structure)

		return nil
	})

	return err
}

// indexSymbols adds symbols from a file to the symbol index
func (m *CodeContextManager) indexSymbols(filePath string, structure *treesitter.CodeStructure) {
	addSymbol := func(sym treesitter.Symbol) {
		loc := SymbolLocation{
			FilePath:  filePath,
			Name:      sym.Name,
			Kind:      string(sym.Kind),
			StartLine: sym.StartLine,
			EndLine:   sym.EndLine,
			Exported:  sym.Exported,
		}
		m.index.Symbols[sym.Name] = append(m.index.Symbols[sym.Name], loc)
	}

	for _, f := range structure.Functions {
		addSymbol(f)
	}
	for _, c := range structure.Classes {
		addSymbol(c)
	}
	for _, t := range structure.Types {
		addSymbol(t)
	}
	for _, v := range structure.Variables {
		if v.Exported { // Only index exported variables
			addSymbol(v)
		}
	}
}

// GetIndex returns the codebase index
func (m *CodeContextManager) GetIndex() *CodebaseIndex {
	return m.index
}

// ContextRequest specifies what context an agent needs
type ContextRequest struct {
	TaskType    string   // e.g., "code_generate", "code_review", "bug_fix"
	Description string   // Task description for keyword extraction
	Symbols     []string // Specific symbols to include
	FilePaths   []string // Specific files to include
	FileTypes   []string // File extensions to prefer (e.g., ".tsx", ".go")
	MaxFiles    int      // Maximum files to include
	MaxTokens   int      // Approximate max tokens for context
}

// RelevantContext holds the selected context for an agent
type RelevantContext struct {
	Files       map[string]string         // path -> content
	Symbols     []SymbolLocation          // Relevant symbols found
	Summary     string                    // Brief summary of context
	TokenEstimate int                     // Estimated tokens
}

// SelectContext intelligently selects relevant code context for a task
func (m *CodeContextManager) SelectContext(ctx context.Context, req ContextRequest) (*RelevantContext, error) {
	if m.index == nil {
		if err := m.IndexProject(ctx); err != nil {
			return nil, err
		}
	}

	result := &RelevantContext{
		Files:   make(map[string]string),
		Symbols: make([]SymbolLocation, 0),
	}

	// Score and rank files based on relevance
	type scoredFile struct {
		path  string
		score float64
	}
	scoredFiles := make([]scoredFile, 0)

	keywords := extractKeywords(req.Description)

	for filePath, structure := range m.index.Files {
		score := m.scoreFileRelevance(filePath, structure, req, keywords)
		if score > 0 {
			scoredFiles = append(scoredFiles, scoredFile{path: filePath, score: score})
		}
	}

	// Sort by score descending
	sort.Slice(scoredFiles, func(i, j int) bool {
		return scoredFiles[i].score > scoredFiles[j].score
	})

	// Select top files within limits
	maxFiles := req.MaxFiles
	if maxFiles == 0 {
		maxFiles = 10 // Default
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 8000 // Default ~8k tokens
	}

	tokenCount := 0
	for _, sf := range scoredFiles {
		if len(result.Files) >= maxFiles {
			break
		}

		fullPath := filepath.Join(m.projectDir, sf.path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Rough token estimate (4 chars per token)
		fileTokens := len(content) / 4
		if tokenCount+fileTokens > maxTokens && len(result.Files) > 0 {
			break // Skip if would exceed token limit
		}

		result.Files[sf.path] = string(content)
		tokenCount += fileTokens
	}

	// Find matching symbols
	for _, sym := range req.Symbols {
		if locs, ok := m.index.Symbols[sym]; ok {
			result.Symbols = append(result.Symbols, locs...)
		}
	}

	result.TokenEstimate = tokenCount
	result.Summary = m.generateContextSummary(result)

	return result, nil
}

// scoreFileRelevance calculates how relevant a file is to a task
func (m *CodeContextManager) scoreFileRelevance(
	filePath string,
	structure *treesitter.CodeStructure,
	req ContextRequest,
	keywords []string,
) float64 {
	score := 0.0

	// Explicit file path match (highest priority)
	for _, reqPath := range req.FilePaths {
		if strings.Contains(filePath, reqPath) || strings.Contains(reqPath, filePath) {
			score += 100.0
		}
	}

	// File type match
	ext := filepath.Ext(filePath)
	for _, reqExt := range req.FileTypes {
		if ext == reqExt || "."+ext == reqExt {
			score += 20.0
		}
	}

	// Symbol match
	for _, reqSym := range req.Symbols {
		for _, fn := range structure.Functions {
			if strings.EqualFold(fn.Name, reqSym) {
				score += 50.0
			}
		}
		for _, cls := range structure.Classes {
			if strings.EqualFold(cls.Name, reqSym) {
				score += 50.0
			}
		}
		for _, typ := range structure.Types {
			if strings.EqualFold(typ.Name, reqSym) {
				score += 50.0
			}
		}
	}

	// Keyword match in file path
	lowerPath := strings.ToLower(filePath)
	for _, kw := range keywords {
		if strings.Contains(lowerPath, kw) {
			score += 10.0
		}
	}

	// Keyword match in symbols
	for _, fn := range structure.Functions {
		lowerName := strings.ToLower(fn.Name)
		for _, kw := range keywords {
			if strings.Contains(lowerName, kw) {
				score += 5.0
			}
		}
	}
	for _, cls := range structure.Classes {
		lowerName := strings.ToLower(cls.Name)
		for _, kw := range keywords {
			if strings.Contains(lowerName, kw) {
				score += 5.0
			}
		}
	}

	// Task type specific scoring
	switch req.TaskType {
	case "code_generate":
		// Prefer similar files as examples
		if len(structure.Functions) > 0 || len(structure.Classes) > 0 {
			score += 2.0
		}
	case "code_review", "bug_fix":
		// Prefer files with more code
		score += float64(structure.LOC) / 100.0
	case "test":
		// Prefer test files
		if strings.Contains(lowerPath, "test") || strings.Contains(lowerPath, "spec") {
			score += 30.0
		}
	}

	// Entry point files get a boost
	if isEntryPoint(filePath) {
		score += 15.0
	}

	// Exported symbols get a boost
	exportedCount := 0
	for _, fn := range structure.Functions {
		if fn.Exported {
			exportedCount++
		}
	}
	for _, cls := range structure.Classes {
		if cls.Exported {
			exportedCount++
		}
	}
	score += float64(exportedCount) * 2.0

	return score
}

// extractKeywords extracts search keywords from a description
func extractKeywords(description string) []string {
	// Simple keyword extraction - split on common delimiters and filter
	words := strings.FieldsFunc(strings.ToLower(description), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})

	// Filter out common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"to": true, "for": true, "of": true, "in": true, "on": true,
		"with": true, "that": true, "this": true, "is": true, "are": true,
		"be": true, "was": true, "were": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "shall": true, "can": true,
		"need": true, "use": true, "using": true, "used": true,
		"create": true, "add": true, "implement": true, "make": true,
	}

	keywords := make([]string, 0)
	for _, w := range words {
		if len(w) >= 3 && !stopWords[w] {
			keywords = append(keywords, w)
		}
	}

	return keywords
}

// isEntryPoint checks if a file is likely an entry point
func isEntryPoint(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	entryPoints := []string{
		"main.go", "main.py", "main.ts", "main.js",
		"index.ts", "index.js", "index.tsx", "index.jsx",
		"app.ts", "app.js", "app.tsx", "app.jsx", "app.py",
		"server.go", "server.ts", "server.js", "server.py",
		"mod.rs", "lib.rs",
	}
	for _, ep := range entryPoints {
		if base == ep {
			return true
		}
	}
	return false
}

// generateContextSummary creates a brief summary of selected context
func (m *CodeContextManager) generateContextSummary(ctx *RelevantContext) string {
	summary := fmt.Sprintf("Context includes %d files", len(ctx.Files))

	if len(ctx.Symbols) > 0 {
		summary += fmt.Sprintf(" with %d matching symbols", len(ctx.Symbols))
	}

	summary += fmt.Sprintf(" (~%d tokens)", ctx.TokenEstimate)

	return summary
}

// PopulateAgentContext fills an agent's context with relevant code
func (m *CodeContextManager) PopulateAgentContext(
	ctx context.Context,
	agent *AgentInstance,
	req ContextRequest,
) error {
	selected, err := m.SelectContext(ctx, req)
	if err != nil {
		return err
	}

	// Update agent context
	agent.Context.RelevantFiles = selected.Files
	agent.Context.CodebaseIndex = selected.Summary

	return nil
}

// GetProjectStack returns the detected project stack
func (m *CodeContextManager) GetProjectStack() *treesitter.ProjectStack {
	if m.index != nil {
		return m.index.Stack
	}
	return nil
}

// FindSymbol searches for a symbol by name
func (m *CodeContextManager) FindSymbol(name string) []SymbolLocation {
	if m.index == nil {
		return nil
	}
	return m.index.Symbols[name]
}

// GetFileStructure returns the structure of a specific file
func (m *CodeContextManager) GetFileStructure(path string) *treesitter.CodeStructure {
	if m.index == nil {
		return nil
	}
	return m.index.Files[path]
}

// ListExportedSymbols returns all exported symbols in the codebase
func (m *CodeContextManager) ListExportedSymbols() []SymbolLocation {
	if m.index == nil {
		return nil
	}

	exported := make([]SymbolLocation, 0)
	for _, locs := range m.index.Symbols {
		for _, loc := range locs {
			if loc.Exported {
				exported = append(exported, loc)
			}
		}
	}
	return exported
}
