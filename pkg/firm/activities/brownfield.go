package activities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
	"github.com/qlfactory/sovereign-firm/pkg/sovereign/memory"
	"github.com/qlfactory/sovereign-firm/pkg/treesitter"
)

// BrownfieldAnalyzer performs comprehensive codebase analysis
type BrownfieldAnalyzer struct {
	store    *memory.MultiLevelStore
	chunker  *memory.CodeChunker
	analyzer *treesitter.ProjectAnalyzer
}

// NewBrownfieldAnalyzer creates a new brownfield analyzer
func NewBrownfieldAnalyzer(chromaURL string) *BrownfieldAnalyzer {
	embedder := llm.NewClient()
	return &BrownfieldAnalyzer{
		store:    memory.NewMultiLevelStore(embedder, chromaURL),
		chunker:  memory.NewCodeChunker(memory.DefaultChunkConfig()),
		analyzer: treesitter.NewProjectAnalyzer(),
	}
}

// Close releases resources
func (b *BrownfieldAnalyzer) Close() {
	if b.chunker != nil {
		b.chunker.Close()
	}
	if b.analyzer != nil {
		b.analyzer.Close()
	}
}

// BrownfieldParams configures brownfield analysis
type BrownfieldParams struct {
	// Source - one of these must be provided
	RepoURL    string // Git repository URL to clone
	LocalPath  string // Local directory path

	// Context
	ProjectID  string
	ClientID   string

	// Options
	MaxFiles       int      // Maximum files to index (0 = unlimited)
	SkipPatterns   []string // Glob patterns to skip
	IncludeTests   bool     // Include test files in analysis
	StoreGlobally  bool     // Also store patterns at global level
}

// BrownfieldResult contains analysis results
type BrownfieldResult struct {
	ProjectID       string                    `json:"project_id"`
	Stack           *treesitter.ProjectStack  `json:"stack"`
	FilesIndexed    int                       `json:"files_indexed"`
	ChunksCreated   int                       `json:"chunks_created"`
	SymbolsFound    int                       `json:"symbols_found"`
	Decisions       []string                  `json:"decisions"`
	Recommendations []string                  `json:"recommendations"`
	Errors          []string                  `json:"errors,omitempty"`
}

// AnalyzeProject performs comprehensive brownfield analysis
func (b *BrownfieldAnalyzer) AnalyzeProject(ctx context.Context, params BrownfieldParams) (*BrownfieldResult, error) {
	result := &BrownfieldResult{
		ProjectID:       params.ProjectID,
		Decisions:       make([]string, 0),
		Recommendations: make([]string, 0),
		Errors:          make([]string, 0),
	}

	// Get project directory
	projectDir, cleanup, err := b.getProjectDir(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get project directory: %w", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	// 1. Analyze project stack
	stack, err := b.analyzer.AnalyzeProject(ctx, projectDir)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("stack analysis failed: %v", err))
	} else {
		result.Stack = stack

		// Store stack decision
		stackDecision := b.describeStack(stack)
		if err := b.store.StoreDecision(ctx, params.ProjectID, params.ClientID,
			"Technology Stack",
			stackDecision,
			map[string]interface{}{"type": "stack_analysis"},
		); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to store stack decision: %v", err))
		}
		result.Decisions = append(result.Decisions, stackDecision)
	}

	// 2. Walk and index files
	filesIndexed := 0
	chunksCreated := 0
	symbolsFound := 0

	err = filepath.Walk(projectDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			base := filepath.Base(path)
			if b.shouldSkipDir(base, params.SkipPatterns) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file limits
		if params.MaxFiles > 0 && filesIndexed >= params.MaxFiles {
			return filepath.SkipDir
		}

		// Skip non-source files
		lang := treesitter.DetectLanguage(path, nil)
		if lang == treesitter.LangUnknown {
			return nil
		}

		// Skip test files if not requested
		if !params.IncludeTests && b.isTestFile(path) {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(projectDir, path)

		// Chunk the file
		chunks, err := b.chunker.ChunkFile(ctx, path, content)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("chunking failed for %s: %v", relPath, err))
			return nil
		}

		// Store chunks
		err = b.store.StoreCodeContext(ctx, params.ProjectID, params.ClientID, relPath, string(lang), string(content), chunks)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("storage failed for %s: %v", relPath, err))
			return nil
		}

		filesIndexed++
		chunksCreated += len(chunks)

		// Count symbols from structure
		structure, err := b.analyzer.AnalyzeFile(ctx, path, content)
		if err == nil {
			symbolsFound += len(structure.Functions) + len(structure.Classes) + len(structure.Types)

			// Store notable patterns globally if configured
			if params.StoreGlobally && len(structure.Functions) > 0 {
				for _, fn := range structure.Functions {
					if fn.Exported && len(fn.Signature) > 0 {
						// Store as a pattern example
						if err := b.store.StorePattern(ctx, memory.GlobalMemory, "",
							fn.Name,
							fmt.Sprintf("Example %s from %s", fn.Kind, relPath),
							b.extractFunctionContent(content, fn),
						); err != nil {
							// Don't fail, just log
						}
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("walk failed: %v", err))
	}

	result.FilesIndexed = filesIndexed
	result.ChunksCreated = chunksCreated
	result.SymbolsFound = symbolsFound

	// 3. Generate recommendations
	result.Recommendations = b.generateRecommendations(stack, filesIndexed, symbolsFound)

	return result, nil
}

// getProjectDir returns the project directory, handling both git clone and local paths
func (b *BrownfieldAnalyzer) getProjectDir(params BrownfieldParams) (string, func(), error) {
	if params.LocalPath != "" {
		// Use local path directly
		if _, err := os.Stat(params.LocalPath); err != nil {
			return "", nil, fmt.Errorf("local path not found: %s", params.LocalPath)
		}
		return params.LocalPath, nil, nil
	}

	if params.RepoURL != "" {
		// Clone to temp directory
		tempDir, err := os.MkdirTemp("", "brownfield-*")
		if err != nil {
			return "", nil, err
		}

		_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
			URL:   params.RepoURL,
			Depth: 1,
		})
		if err != nil {
			os.RemoveAll(tempDir)
			return "", nil, fmt.Errorf("clone failed: %w", err)
		}

		cleanup := func() { os.RemoveAll(tempDir) }
		return tempDir, cleanup, nil
	}

	return "", nil, fmt.Errorf("either RepoURL or LocalPath must be provided")
}

// shouldSkipDir checks if a directory should be skipped
func (b *BrownfieldAnalyzer) shouldSkipDir(name string, skipPatterns []string) bool {
	// Always skip these
	skipDirs := []string{
		"node_modules", ".git", "vendor", "__pycache__",
		".venv", "venv", "dist", "build", "target",
		".next", ".nuxt", "coverage", ".cache",
	}

	for _, skip := range skipDirs {
		if name == skip {
			return true
		}
	}

	// Check custom patterns
	for _, pattern := range skipPatterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}

	return false
}

// isTestFile checks if a file is a test file
func (b *BrownfieldAnalyzer) isTestFile(path string) bool {
	base := filepath.Base(path)
	return strings.Contains(base, "_test.") ||
		strings.Contains(base, ".test.") ||
		strings.Contains(base, ".spec.") ||
		strings.HasPrefix(base, "test_")
}

// describeStack creates a human-readable description of the tech stack
func (b *BrownfieldAnalyzer) describeStack(stack *treesitter.ProjectStack) string {
	if stack == nil {
		return "Unable to detect technology stack"
	}

	var parts []string

	// Primary language
	if stack.PrimaryLanguage != treesitter.LangUnknown {
		parts = append(parts, fmt.Sprintf("Primary language: %s", stack.PrimaryLanguage))
	}

	// Frameworks
	if len(stack.Frameworks) > 0 {
		frameworks := make([]string, len(stack.Frameworks))
		for i, f := range stack.Frameworks {
			frameworks[i] = string(f)
		}
		parts = append(parts, fmt.Sprintf("Frameworks: %s", strings.Join(frameworks, ", ")))
	}

	// Build tools
	if len(stack.BuildTools) > 0 {
		tools := make([]string, len(stack.BuildTools))
		for i, t := range stack.BuildTools {
			tools[i] = string(t)
		}
		parts = append(parts, fmt.Sprintf("Build tools: %s", strings.Join(tools, ", ")))
	}

	// Test frameworks
	if len(stack.TestFrameworks) > 0 {
		testFw := make([]string, len(stack.TestFrameworks))
		for i, t := range stack.TestFrameworks {
			testFw[i] = string(t)
		}
		parts = append(parts, fmt.Sprintf("Test frameworks: %s", strings.Join(testFw, ", ")))
	}

	// Package manager
	if stack.PackageManager != "" {
		parts = append(parts, fmt.Sprintf("Package manager: %s", stack.PackageManager))
	}

	// TypeScript
	if stack.HasTypeScript {
		parts = append(parts, "Uses TypeScript")
	}

	// Monorepo
	if stack.IsMonorepo {
		parts = append(parts, "Monorepo structure detected")
	}

	return strings.Join(parts, ". ")
}

// extractFunctionContent extracts function source code
func (b *BrownfieldAnalyzer) extractFunctionContent(content []byte, fn treesitter.Symbol) string {
	lines := strings.Split(string(content), "\n")
	startIdx := int(fn.StartLine) - 1
	endIdx := int(fn.EndLine)

	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	return strings.Join(lines[startIdx:endIdx], "\n")
}

// generateRecommendations creates recommendations based on analysis
func (b *BrownfieldAnalyzer) generateRecommendations(stack *treesitter.ProjectStack, filesIndexed, symbolsFound int) []string {
	recommendations := make([]string, 0)

	if stack == nil {
		recommendations = append(recommendations, "Could not detect project structure - consider adding configuration files")
		return recommendations
	}

	// TypeScript recommendation
	if !stack.HasTypeScript && (stack.PrimaryLanguage == treesitter.LangJavaScript) {
		recommendations = append(recommendations, "Consider migrating to TypeScript for better type safety")
	}

	// Test recommendation
	if !stack.HasTests {
		recommendations = append(recommendations, "No test files detected - consider adding tests")
	}

	// Size-based recommendations
	if filesIndexed > 100 {
		recommendations = append(recommendations, "Large codebase detected - consider incremental indexing for updates")
	}

	if symbolsFound > 500 {
		recommendations = append(recommendations, "Many symbols detected - semantic search will be most effective for navigation")
	}

	// Framework-specific
	for _, fw := range stack.Frameworks {
		switch fw {
		case treesitter.FrameworkReact:
			if !hasFramework(stack.Frameworks, treesitter.FrameworkNextJS) {
				recommendations = append(recommendations, "Consider Next.js for server-side rendering if needed")
			}
		case treesitter.FrameworkExpress:
			recommendations = append(recommendations, "Consider adding request validation middleware")
		}
	}

	return recommendations
}

func hasFramework(frameworks []treesitter.Framework, target treesitter.Framework) bool {
	for _, f := range frameworks {
		if f == target {
			return true
		}
	}
	return false
}

// QuickAnalyze performs a fast analysis without full indexing
func (b *BrownfieldAnalyzer) QuickAnalyze(ctx context.Context, projectDir string) (*treesitter.ProjectStack, error) {
	return b.analyzer.AnalyzeProject(ctx, projectDir)
}
