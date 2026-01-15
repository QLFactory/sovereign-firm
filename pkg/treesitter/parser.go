// Package treesitter provides language-agnostic code parsing and analysis.
// It wraps go-tree-sitter to provide consistent AST access across 20+ languages.
package treesitter

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// Language represents a detected programming language
type Language string

const (
	LangGo         Language = "go"
	LangTypeScript Language = "typescript"
	LangTSX        Language = "tsx"
	LangJavaScript Language = "javascript"
	LangJSX        Language = "jsx"
	LangPython     Language = "python"
	LangRust       Language = "rust"
	LangJava       Language = "java"
	LangC          Language = "c"
	LangCpp        Language = "cpp"
	LangRuby       Language = "ruby"
	LangUnknown    Language = "unknown"
)

// Parser provides language-agnostic code parsing
type Parser struct {
	parsers map[Language]*sitter.Parser
	mu      sync.RWMutex
}

// NewParser creates a new Tree-sitter parser
func NewParser() *Parser {
	return &Parser{
		parsers: make(map[Language]*sitter.Parser),
	}
}

// getParser returns or creates a parser for the given language
func (p *Parser) getParser(lang Language) (*sitter.Parser, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if parser, ok := p.parsers[lang]; ok {
		return parser, nil
	}

	parser := sitter.NewParser()
	sitterLang, err := p.getSitterLanguage(lang)
	if err != nil {
		return nil, err
	}

	parser.SetLanguage(sitterLang)
	p.parsers[lang] = parser
	return parser, nil
}

// getSitterLanguage returns the Tree-sitter language implementation
func (p *Parser) getSitterLanguage(lang Language) (*sitter.Language, error) {
	switch lang {
	case LangGo:
		return golang.GetLanguage(), nil
	case LangTypeScript:
		return typescript.GetLanguage(), nil
	case LangTSX, LangJSX:
		return tsx.GetLanguage(), nil
	case LangJavaScript:
		return javascript.GetLanguage(), nil
	case LangPython:
		return python.GetLanguage(), nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
}

// DetectLanguage determines the programming language from a filename
func DetectLanguage(filename string) Language {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".go":
		return LangGo
	case ".ts":
		return LangTypeScript
	case ".tsx":
		return LangTSX
	case ".js":
		return LangJavaScript
	case ".jsx":
		return LangJSX
	case ".py":
		return LangPython
	case ".rs":
		return LangRust
	case ".java":
		return LangJava
	case ".c", ".h":
		return LangC
	case ".cpp", ".cc", ".cxx", ".hpp":
		return LangCpp
	case ".rb":
		return LangRuby
	default:
		return LangUnknown
	}
}

// ParseResult contains the parsed AST and metadata
type ParseResult struct {
	Tree     *sitter.Tree
	Language Language
	Filename string
	Code     []byte
}

// Parse parses code and returns the AST
func (p *Parser) Parse(filename string, code []byte) (*ParseResult, error) {
	lang := DetectLanguage(filename)
	if lang == LangUnknown {
		return nil, fmt.Errorf("cannot detect language for %s", filename)
	}

	parser, err := p.getParser(lang)
	if err != nil {
		return nil, err
	}

	tree, err := parser.ParseCtx(nil, nil, code)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return &ParseResult{
		Tree:     tree,
		Language: lang,
		Filename: filename,
		Code:     code,
	}, nil
}

// Symbol represents a code symbol (function, class, variable, etc.)
type Symbol struct {
	Name       string   `json:"name"`
	Kind       string   `json:"kind"` // function, class, method, variable, type
	Language   Language `json:"language"`
	File       string   `json:"file"`
	StartLine  int      `json:"start_line"`
	EndLine    int      `json:"end_line"`
	StartByte  uint32   `json:"start_byte"`
	EndByte    uint32   `json:"end_byte"`
	Signature  string   `json:"signature,omitempty"`
	DocComment string   `json:"doc_comment,omitempty"`
	Parent     string   `json:"parent,omitempty"` // For methods, the class name
}

// ExtractSymbols extracts all symbols from a parsed file
func (p *Parser) ExtractSymbols(result *ParseResult) ([]Symbol, error) {
	var symbols []Symbol

	rootNode := result.Tree.RootNode()

	// Walk the tree and extract symbols based on language
	p.walkTree(rootNode, result, "", &symbols)

	return symbols, nil
}

// walkTree recursively walks the AST extracting symbols
func (p *Parser) walkTree(node *sitter.Node, result *ParseResult, parent string, symbols *[]Symbol) {
	if node == nil {
		return
	}

	// Extract symbol based on node type
	if sym := p.extractSymbol(node, result, parent); sym != nil {
		*symbols = append(*symbols, *sym)

		// Update parent for nested symbols
		if sym.Kind == "class" || sym.Kind == "interface" {
			parent = sym.Name
		}
	}

	// Recurse into children
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		p.walkTree(child, result, parent, symbols)
	}
}

// extractSymbol extracts a symbol from an AST node
func (p *Parser) extractSymbol(node *sitter.Node, result *ParseResult, parent string) *Symbol {
	nodeType := node.Type()

	var kind string
	var nameNode *sitter.Node

	switch result.Language {
	case LangGo:
		kind, nameNode = p.extractGoSymbol(node)
	case LangTypeScript, LangTSX, LangJavaScript, LangJSX:
		kind, nameNode = p.extractJSSymbol(node)
	case LangPython:
		kind, nameNode = p.extractPythonSymbol(node)
	default:
		return nil
	}

	if kind == "" || nameNode == nil {
		return nil
	}

	name := string(result.Code[nameNode.StartByte():nameNode.EndByte()])

	// Get signature (the whole first line)
	startLine := int(node.StartPoint().Row)
	endLine := int(node.EndPoint().Row)

	return &Symbol{
		Name:      name,
		Kind:      kind,
		Language:  result.Language,
		File:      result.Filename,
		StartLine: startLine + 1, // 1-indexed
		EndLine:   endLine + 1,
		StartByte: node.StartByte(),
		EndByte:   node.EndByte(),
		Parent:    parent,
		Signature: p.getSignature(node, result.Code, nodeType),
	}
}

// extractGoSymbol extracts symbol info from Go AST nodes
func (p *Parser) extractGoSymbol(node *sitter.Node) (kind string, nameNode *sitter.Node) {
	switch node.Type() {
	case "function_declaration":
		nameNode = node.ChildByFieldName("name")
		return "function", nameNode
	case "method_declaration":
		nameNode = node.ChildByFieldName("name")
		return "method", nameNode
	case "type_declaration":
		// Look for type_spec child
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "type_spec" {
				nameNode = child.ChildByFieldName("name")
				return "type", nameNode
			}
		}
	case "var_declaration", "const_declaration":
		// Look for var_spec child
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if strings.HasSuffix(child.Type(), "_spec") {
				nameNode = child.ChildByFieldName("name")
				if node.Type() == "const_declaration" {
					return "constant", nameNode
				}
				return "variable", nameNode
			}
		}
	}
	return "", nil
}

// extractJSSymbol extracts symbol info from JavaScript/TypeScript AST nodes
func (p *Parser) extractJSSymbol(node *sitter.Node) (kind string, nameNode *sitter.Node) {
	switch node.Type() {
	case "function_declaration", "function":
		nameNode = node.ChildByFieldName("name")
		return "function", nameNode
	case "arrow_function":
		// Arrow functions in assignments
		parent := node.Parent()
		if parent != nil && parent.Type() == "variable_declarator" {
			nameNode = parent.ChildByFieldName("name")
			return "function", nameNode
		}
	case "class_declaration", "class":
		nameNode = node.ChildByFieldName("name")
		return "class", nameNode
	case "method_definition":
		nameNode = node.ChildByFieldName("name")
		return "method", nameNode
	case "interface_declaration":
		nameNode = node.ChildByFieldName("name")
		return "interface", nameNode
	case "type_alias_declaration":
		nameNode = node.ChildByFieldName("name")
		return "type", nameNode
	case "lexical_declaration":
		// const/let declarations
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "variable_declarator" {
				nameNode = child.ChildByFieldName("name")
				return "variable", nameNode
			}
		}
	}
	return "", nil
}

// extractPythonSymbol extracts symbol info from Python AST nodes
func (p *Parser) extractPythonSymbol(node *sitter.Node) (kind string, nameNode *sitter.Node) {
	switch node.Type() {
	case "function_definition":
		nameNode = node.ChildByFieldName("name")
		return "function", nameNode
	case "class_definition":
		nameNode = node.ChildByFieldName("name")
		return "class", nameNode
	case "decorated_definition":
		// Handle @decorator on functions/classes
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "function_definition" {
				nameNode = child.ChildByFieldName("name")
				return "function", nameNode
			}
			if child.Type() == "class_definition" {
				nameNode = child.ChildByFieldName("name")
				return "class", nameNode
			}
		}
	}
	return "", nil
}

// getSignature extracts the first line of a symbol as its signature
func (p *Parser) getSignature(node *sitter.Node, code []byte, nodeType string) string {
	start := node.StartByte()
	end := node.EndByte()

	// Find first newline
	for i := start; i < end && i < uint32(len(code)); i++ {
		if code[i] == '\n' {
			return strings.TrimSpace(string(code[start:i]))
		}
	}

	// No newline, use whole content (but limit length)
	content := string(code[start:end])
	if len(content) > 100 {
		content = content[:100] + "..."
	}
	return strings.TrimSpace(content)
}

// ValidateSyntax checks if the code is syntactically valid
func (p *Parser) ValidateSyntax(filename string, code []byte) (valid bool, errors []string) {
	result, err := p.Parse(filename, code)
	if err != nil {
		return false, []string{err.Error()}
	}

	// Check for error nodes in the tree
	errors = p.findErrors(result.Tree.RootNode(), code)
	return len(errors) == 0, errors
}

// findErrors recursively finds error nodes in the AST
func (p *Parser) findErrors(node *sitter.Node, code []byte) []string {
	var errors []string

	if node.IsError() || node.IsMissing() {
		line := node.StartPoint().Row + 1
		col := node.StartPoint().Column + 1

		snippet := ""
		if node.EndByte() > node.StartByte() {
			snippet = string(code[node.StartByte():min(node.EndByte(), node.StartByte()+50)])
		}

		errors = append(errors, fmt.Sprintf(
			"Syntax error at line %d, column %d: near '%s'",
			line, col, snippet,
		))
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		childErrors := p.findErrors(node.Child(i), code)
		errors = append(errors, childErrors...)
	}

	return errors
}

// GetNodeAtPosition finds the AST node at a given line/column
func (p *Parser) GetNodeAtPosition(result *ParseResult, line, column int) *sitter.Node {
	point := sitter.Point{Row: uint32(line - 1), Column: uint32(column)}
	return result.Tree.RootNode().NamedDescendantForPointRange(point, point)
}

// ExtractNodeContent gets the code content for a node
func ExtractNodeContent(node *sitter.Node, code []byte) string {
	return string(code[node.StartByte():node.EndByte()])
}

func min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
