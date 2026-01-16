package memory

import (
	"context"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/treesitter"
)

// ChunkConfig configures code chunking behavior
type ChunkConfig struct {
	MaxChunkSize    int  // Maximum characters per chunk (default: 2000)
	MinChunkSize    int  // Minimum characters per chunk (default: 100)
	OverlapLines    int  // Lines to overlap between chunks (default: 3)
	IncludeImports  bool // Include imports in each chunk (default: true)
	PreserveSymbols bool // Try to keep functions/classes intact (default: true)
}

// DefaultChunkConfig returns sensible defaults
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		MaxChunkSize:    2000,
		MinChunkSize:    100,
		OverlapLines:    3,
		IncludeImports:  true,
		PreserveSymbols: true,
	}
}

// CodeChunker splits code into semantically meaningful chunks
type CodeChunker struct {
	analyzer *treesitter.ProjectAnalyzer
	config   ChunkConfig
}

// NewCodeChunker creates a new chunker with the given config
func NewCodeChunker(config ChunkConfig) *CodeChunker {
	if config.MaxChunkSize == 0 {
		config = DefaultChunkConfig()
	}
	return &CodeChunker{
		analyzer: treesitter.NewProjectAnalyzer(),
		config:   config,
	}
}

// Close releases resources
func (c *CodeChunker) Close() {
	if c.analyzer != nil {
		c.analyzer.Close()
	}
}

// ChunkFile splits a file into chunks using tree-sitter analysis
func (c *CodeChunker) ChunkFile(ctx context.Context, filePath string, content []byte) ([]CodeChunk, error) {
	// Detect language
	lang := treesitter.DetectLanguage(filePath, content)
	if lang == treesitter.LangUnknown || !treesitter.IsSupported(lang) {
		// Fall back to line-based chunking
		return c.chunkByLines(string(content))
	}

	// Parse and extract structure
	structure, err := c.analyzer.AnalyzeFile(ctx, filePath, content)
	if err != nil {
		// Fall back to line-based chunking
		return c.chunkByLines(string(content))
	}

	return c.chunkBySymbols(string(content), structure)
}

// chunkBySymbols creates chunks based on code structure
func (c *CodeChunker) chunkBySymbols(content string, structure *treesitter.CodeStructure) ([]CodeChunk, error) {
	lines := strings.Split(content, "\n")
	chunks := make([]CodeChunk, 0)

	// Extract import section
	var importSection string
	if c.config.IncludeImports && len(structure.Imports) > 0 {
		maxImportLine := uint32(0)
		for _, imp := range structure.Imports {
			if imp.EndLine > maxImportLine {
				maxImportLine = imp.EndLine
			}
		}
		if maxImportLine > 0 && int(maxImportLine) <= len(lines) {
			importLines := lines[:maxImportLine]
			importSection = strings.Join(importLines, "\n") + "\n\n"
		}
	}

	// Collect all symbols with their ranges
	type symbolRange struct {
		name      string
		kind      string
		startLine uint32
		endLine   uint32
	}
	symbols := make([]symbolRange, 0)

	for _, fn := range structure.Functions {
		symbols = append(symbols, symbolRange{
			name:      fn.Name,
			kind:      string(fn.Kind),
			startLine: fn.StartLine,
			endLine:   fn.EndLine,
		})
	}

	for _, cls := range structure.Classes {
		symbols = append(symbols, symbolRange{
			name:      cls.Name,
			kind:      string(cls.Kind),
			startLine: cls.StartLine,
			endLine:   cls.EndLine,
		})
	}

	for _, typ := range structure.Types {
		symbols = append(symbols, symbolRange{
			name:      typ.Name,
			kind:      string(typ.Kind),
			startLine: typ.StartLine,
			endLine:   typ.EndLine,
		})
	}

	// If no symbols found, fall back to line-based chunking
	if len(symbols) == 0 {
		return c.chunkByLines(content)
	}

	// Sort symbols by start line (they should already be, but ensure)
	// Create chunks from symbols
	for _, sym := range symbols {
		startIdx := int(sym.startLine) - 1
		endIdx := int(sym.endLine)

		if startIdx < 0 {
			startIdx = 0
		}
		if endIdx > len(lines) {
			endIdx = len(lines)
		}
		if startIdx >= endIdx {
			continue
		}

		symbolLines := lines[startIdx:endIdx]
		symbolContent := strings.Join(symbolLines, "\n")

		// Check if symbol is too large
		if len(symbolContent) > c.config.MaxChunkSize {
			// Split large symbols into sub-chunks
			subChunks := c.splitLargeSymbol(symbolLines, startIdx+1, sym.name)
			for _, subChunk := range subChunks {
				if c.config.IncludeImports && importSection != "" {
					subChunk.Content = importSection + subChunk.Content
				}
				chunks = append(chunks, subChunk)
			}
		} else if len(symbolContent) >= c.config.MinChunkSize {
			chunk := CodeChunk{
				Content:   symbolContent,
				StartLine: startIdx + 1,
				EndLine:   endIdx,
				Symbol:    sym.name,
			}
			if c.config.IncludeImports && importSection != "" {
				chunk.Content = importSection + chunk.Content
			}
			chunks = append(chunks, chunk)
		}
	}

	// If we got no chunks from symbols, fall back to line-based
	if len(chunks) == 0 {
		return c.chunkByLines(content)
	}

	return chunks, nil
}

// splitLargeSymbol splits a large symbol into smaller chunks
func (c *CodeChunker) splitLargeSymbol(lines []string, baseLineNum int, symbol string) []CodeChunk {
	chunks := make([]CodeChunk, 0)
	currentChunk := make([]string, 0)
	currentSize := 0
	startLine := baseLineNum

	for i, line := range lines {
		lineSize := len(line) + 1 // +1 for newline

		if currentSize+lineSize > c.config.MaxChunkSize && len(currentChunk) > 0 {
			// Save current chunk
			chunks = append(chunks, CodeChunk{
				Content:   strings.Join(currentChunk, "\n"),
				StartLine: startLine,
				EndLine:   baseLineNum + i - 1,
				Symbol:    symbol,
			})

			// Start new chunk with overlap
			overlapStart := len(currentChunk) - c.config.OverlapLines
			if overlapStart < 0 {
				overlapStart = 0
			}
			currentChunk = currentChunk[overlapStart:]
			startLine = baseLineNum + i - (len(currentChunk))
			currentSize = 0
			for _, l := range currentChunk {
				currentSize += len(l) + 1
			}
		}

		currentChunk = append(currentChunk, line)
		currentSize += lineSize
	}

	// Don't forget the last chunk
	if len(currentChunk) > 0 {
		chunks = append(chunks, CodeChunk{
			Content:   strings.Join(currentChunk, "\n"),
			StartLine: startLine,
			EndLine:   baseLineNum + len(lines) - 1,
			Symbol:    symbol,
		})
	}

	return chunks
}

// chunkByLines creates chunks based on line count (fallback)
func (c *CodeChunker) chunkByLines(content string) ([]CodeChunk, error) {
	lines := strings.Split(content, "\n")
	chunks := make([]CodeChunk, 0)

	// Calculate approximate lines per chunk
	avgLineLen := len(content) / max(len(lines), 1)
	linesPerChunk := c.config.MaxChunkSize / max(avgLineLen, 50)
	if linesPerChunk < 10 {
		linesPerChunk = 10
	}

	for i := 0; i < len(lines); i += linesPerChunk - c.config.OverlapLines {
		endIdx := i + linesPerChunk
		if endIdx > len(lines) {
			endIdx = len(lines)
		}

		chunkLines := lines[i:endIdx]
		chunkContent := strings.Join(chunkLines, "\n")

		if len(chunkContent) >= c.config.MinChunkSize {
			chunks = append(chunks, CodeChunk{
				Content:   chunkContent,
				StartLine: i + 1,
				EndLine:   endIdx,
			})
		}
	}

	// If file is too small for chunking, return whole file
	if len(chunks) == 0 && len(content) > 0 {
		chunks = append(chunks, CodeChunk{
			Content:   content,
			StartLine: 1,
			EndLine:   len(lines),
		})
	}

	return chunks, nil
}

// ChunkResult holds the result of chunking a file
type ChunkResult struct {
	FilePath string
	Language string
	Chunks   []CodeChunk
	Stats    ChunkStats
}

// ChunkStats provides statistics about chunking
type ChunkStats struct {
	TotalChunks    int
	TotalLines     int
	TotalChars     int
	AvgChunkSize   int
	SymbolsFound   int
	FallbackToLine bool
}

// ChunkFileWithStats chunks a file and returns detailed stats
func (c *CodeChunker) ChunkFileWithStats(ctx context.Context, filePath string, content []byte) (*ChunkResult, error) {
	lang := treesitter.DetectLanguage(filePath, content)

	chunks, err := c.ChunkFile(ctx, filePath, content)
	if err != nil {
		return nil, err
	}

	totalChars := 0
	for _, chunk := range chunks {
		totalChars += len(chunk.Content)
	}

	avgSize := 0
	if len(chunks) > 0 {
		avgSize = totalChars / len(chunks)
	}

	lines := strings.Split(string(content), "\n")

	return &ChunkResult{
		FilePath: filePath,
		Language: string(lang),
		Chunks:   chunks,
		Stats: ChunkStats{
			TotalChunks:  len(chunks),
			TotalLines:   len(lines),
			TotalChars:   len(content),
			AvgChunkSize: avgSize,
		},
	}, nil
}

// BatchChunkConfig configures batch chunking
type BatchChunkConfig struct {
	ChunkConfig
	MaxConcurrent int      // Max concurrent file processing
	SkipPatterns  []string // File patterns to skip
}

// ChunkFiles chunks multiple files
func (c *CodeChunker) ChunkFiles(ctx context.Context, files map[string][]byte) (map[string]*ChunkResult, error) {
	results := make(map[string]*ChunkResult)

	for path, content := range files {
		result, err := c.ChunkFileWithStats(ctx, path, content)
		if err != nil {
			continue // Skip files that fail
		}
		results[path] = result
	}

	return results, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
