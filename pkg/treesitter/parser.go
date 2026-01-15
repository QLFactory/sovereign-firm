// Package treesitter provides code parsing and analysis using tree-sitter.
package treesitter

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/toml"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
)

// Language represents a supported programming language
type Language string

const (
	LangUnknown    Language = "unknown"
	LangJavaScript Language = "javascript"
	LangTypeScript Language = "typescript"
	LangTSX        Language = "tsx"
	LangJSX        Language = "jsx"
	LangGo         Language = "go"
	LangPython     Language = "python"
	LangRust       Language = "rust"
	LangHTML       Language = "html"
	LangCSS        Language = "css"
	LangJSON       Language = "json"     // Not parsed, just detected
	LangYAML       Language = "yaml"
	LangTOML       Language = "toml"
	LangBash       Language = "bash"
	LangMarkdown   Language = "markdown"
)

// Parser provides multi-language code parsing using tree-sitter
type Parser struct {
	parsers map[Language]*sitter.Parser
	mu      sync.RWMutex
}

// NewParser creates a new multi-language parser
func NewParser() *Parser {
	return &Parser{
		parsers: make(map[Language]*sitter.Parser),
	}
}

// getParser returns a parser for the given language, creating one if needed
func (p *Parser) getParser(lang Language) (*sitter.Parser, error) {
	p.mu.RLock()
	if parser, ok := p.parsers[lang]; ok {
		p.mu.RUnlock()
		return parser, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	if parser, ok := p.parsers[lang]; ok {
		return parser, nil
	}

	parser := sitter.NewParser()
	sitterLang, err := getSitterLanguage(lang)
	if err != nil {
		return nil, err
	}
	parser.SetLanguage(sitterLang)
	p.parsers[lang] = parser

	return parser, nil
}

// getSitterLanguage returns the tree-sitter language for a given Language
func getSitterLanguage(lang Language) (*sitter.Language, error) {
	switch lang {
	case LangJavaScript, LangJSX:
		return javascript.GetLanguage(), nil
	case LangTypeScript:
		return typescript.GetLanguage(), nil
	case LangTSX:
		return tsx.GetLanguage(), nil
	case LangGo:
		return golang.GetLanguage(), nil
	case LangPython:
		return python.GetLanguage(), nil
	case LangRust:
		return rust.GetLanguage(), nil
	case LangHTML:
		return html.GetLanguage(), nil
	case LangCSS:
		return css.GetLanguage(), nil
	case LangYAML:
		return yaml.GetLanguage(), nil
	case LangTOML:
		return toml.GetLanguage(), nil
	case LangBash:
		return bash.GetLanguage(), nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
}

// ParseResult contains the parsed syntax tree and metadata
type ParseResult struct {
	Tree     *sitter.Tree
	Language Language
	FilePath string
	Source   []byte
}

// Parse parses source code with automatic language detection
func (p *Parser) Parse(ctx context.Context, filePath string, source []byte) (*ParseResult, error) {
	lang := DetectLanguage(filePath, source)
	if lang == LangUnknown {
		return nil, fmt.Errorf("could not detect language for file: %s", filePath)
	}

	return p.ParseWithLanguage(ctx, filePath, source, lang)
}

// ParseWithLanguage parses source code with a specific language
func (p *Parser) ParseWithLanguage(ctx context.Context, filePath string, source []byte, lang Language) (*ParseResult, error) {
	parser, err := p.getParser(lang)
	if err != nil {
		return nil, err
	}

	tree, err := parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return &ParseResult{
		Tree:     tree,
		Language: lang,
		FilePath: filePath,
		Source:   source,
	}, nil
}

// DetectLanguage detects the programming language from file path and content
func DetectLanguage(filePath string, source []byte) Language {
	ext := strings.ToLower(filepath.Ext(filePath))
	baseName := strings.ToLower(filepath.Base(filePath))

	switch ext {
	case ".js":
		if containsJSX(source) {
			return LangJSX
		}
		return LangJavaScript
	case ".jsx":
		return LangJSX
	case ".ts":
		return LangTypeScript
	case ".tsx":
		return LangTSX
	case ".go":
		return LangGo
	case ".py":
		return LangPython
	case ".rs":
		return LangRust
	case ".html", ".htm":
		return LangHTML
	case ".css":
		return LangCSS
	case ".json":
		return LangJSON
	case ".yaml", ".yml":
		return LangYAML
	case ".toml":
		return LangTOML
	case ".sh", ".bash":
		return LangBash
	case ".md", ".markdown":
		return LangMarkdown
	}

	switch baseName {
	case "dockerfile":
		return LangBash
	case "makefile", "gnumakefile":
		return LangBash
	case ".bashrc", ".bash_profile", ".zshrc":
		return LangBash
	}

	if len(source) > 2 && source[0] == '#' && source[1] == '!' {
		shebang := strings.ToLower(string(source[:min(100, len(source))]))
		if strings.Contains(shebang, "python") {
			return LangPython
		}
		if strings.Contains(shebang, "node") {
			return LangJavaScript
		}
		if strings.Contains(shebang, "bash") || strings.Contains(shebang, "/sh") {
			return LangBash
		}
	}

	return LangUnknown
}

// containsJSX checks if JavaScript source contains JSX syntax
func containsJSX(source []byte) bool {
	s := string(source)
	return strings.Contains(s, "React") ||
		strings.Contains(s, "jsx") ||
		strings.Contains(s, "</>") ||
		strings.Contains(s, "/>") ||
		(strings.Contains(s, "<") && strings.Contains(s, "className"))
}

// SupportedLanguages returns a list of all supported languages for parsing
func SupportedLanguages() []Language {
	return []Language{
		LangJavaScript, LangTypeScript, LangTSX, LangJSX,
		LangGo, LangPython, LangRust,
		LangHTML, LangCSS, LangYAML, LangTOML, LangBash,
	}
}

// IsSupported returns true if the language is supported for parsing
func IsSupported(lang Language) bool {
	for _, l := range SupportedLanguages() {
		if l == lang {
			return true
		}
	}
	return false
}

// Close releases all parser resources
func (p *Parser) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, parser := range p.parsers {
		parser.Close()
	}
	p.parsers = make(map[Language]*sitter.Parser)
}
