package treesitter

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

// SymbolKind represents the type of code symbol
type SymbolKind string

const (
	SymbolFunction  SymbolKind = "function"
	SymbolMethod    SymbolKind = "method"
	SymbolClass     SymbolKind = "class"
	SymbolInterface SymbolKind = "interface"
	SymbolVariable  SymbolKind = "variable"
	SymbolConstant  SymbolKind = "constant"
	SymbolImport    SymbolKind = "import"
	SymbolExport    SymbolKind = "export"
	SymbolType      SymbolKind = "type"
	SymbolStruct    SymbolKind = "struct"
	SymbolEnum      SymbolKind = "enum"
)

// Symbol represents a code symbol (function, class, variable, etc.)
type Symbol struct {
	Name      string     `json:"name"`
	Kind      SymbolKind `json:"kind"`
	StartLine uint32     `json:"start_line"`
	EndLine   uint32     `json:"end_line"`
	StartByte uint32     `json:"start_byte"`
	EndByte   uint32     `json:"end_byte"`
	Signature string     `json:"signature,omitempty"` // Function signature
	Parent    string     `json:"parent,omitempty"`    // Parent class/struct name
	Exported  bool       `json:"exported"`            // Is public/exported
	Children  []Symbol   `json:"children,omitempty"`  // Nested symbols
}

// CodeStructure represents the overall structure of a code file
type CodeStructure struct {
	FilePath  string   `json:"file_path"`
	Language  Language `json:"language"`
	Imports   []Symbol `json:"imports"`
	Exports   []Symbol `json:"exports"`
	Functions []Symbol `json:"functions"`
	Classes   []Symbol `json:"classes"`
	Types     []Symbol `json:"types"`
	Variables []Symbol `json:"variables"`
	LOC       int      `json:"loc"` // Lines of code
}

// ExtractSymbols extracts all symbols from a parse result
func ExtractSymbols(p *Parser, result *ParseResult) (*CodeStructure, error) {
	if result == nil || result.Tree == nil {
		return nil, fmt.Errorf("invalid parse result")
	}

	structure := &CodeStructure{
		FilePath:  result.FilePath,
		Language:  result.Language,
		Imports:   []Symbol{},
		Exports:   []Symbol{},
		Functions: []Symbol{},
		Classes:   []Symbol{},
		Types:     []Symbol{},
		Variables: []Symbol{},
	}

	// Count lines of code
	structure.LOC = countLines(result.Source)

	// Try query-based extraction first
	if err := extractSymbolsWithQueries(p, result, structure); err == nil {
		// Only consider it a success if we found some structural elements
		if len(structure.Imports) > 0 || len(structure.Functions) > 0 || len(structure.Classes) > 0 || len(structure.Types) > 0 || len(structure.Variables) > 0 {
			return structure, nil
		}
	}

	// Fallback to legacy extraction or generic if queries fail
	root := result.Tree.RootNode()
	switch result.Language {
	case LangJavaScript, LangJSX:
		extractJavaScriptSymbols(root, result.Source, structure)
	case LangTypeScript, LangTSX:
		extractTypeScriptSymbols(root, result.Source, structure)
	case LangGo:
		extractGoSymbols(root, result.Source, structure)
	case LangPython:
		extractPythonSymbols(root, result.Source, structure)
	case LangRust:
		extractRustSymbols(root, result.Source, structure)
	case LangPHP:
		extractPHPSymbols(root, result.Source, structure)
	case LangCPP, LangC:
		extractCPPSymbols(root, result.Source, structure)
	default:
		extractGenericSymbols(root, result.Source, structure)
	}

	return structure, nil
}

var languageQueries = map[Language]string{
	LangGo: `
		(import_declaration) @import
		(function_declaration name: (identifier) @name) @function
		(method_declaration name: (field_identifier) @name) @method
		(type_spec name: (type_identifier) @name) @type
		(var_spec name: (identifier_list (identifier) @name)) @variable
		(const_spec name: (identifier_list (identifier) @name)) @constant
	`,
	LangJavaScript: `
		(import_statement) @import
		(export_statement) @export
		(function_declaration name: (identifier) @name) @function
		(class_declaration name: (identifier) @name) @class
		(lexical_declaration) @variable
		(variable_declaration) @variable
	`,
	LangTypeScript: `
		(import_statement) @import
		(export_statement) @export
		(function_declaration name: (identifier) @name) @function
		(class_declaration name: (identifier) @name) @class
		(interface_declaration name: (type_identifier) @name) @interface
		(type_alias_declaration name: (type_identifier) @name) @type
		(enum_declaration name: (identifier) @name) @enum
	`,
	LangPython: `
		(import_statement) @import
		(import_from_statement) @import
		(function_definition name: (identifier) @name) @function
		(class_definition name: (identifier) @name) @class
	`,
	LangRust: `
		(function_item name: (identifier) @name) @function
		(struct_item name: (type_identifier) @name) @struct
		(enum_item name: (type_identifier) @name) @enum
		(trait_item name: (type_identifier) @name) @interface
		(mod_item name: (identifier) @name) @export
	`,
	LangPHP: `
		(function_definition name: (name) @name) @function
		(class_declaration name: (name) @name) @class
		(interface_declaration name: (name) @name) @interface
		(trait_declaration name: (name) @name) @interface
	`,
	LangCPP: `
		(function_definition declarator: (function_declarator declarator: (identifier) @name)) @function
		(class_specifier name: (type_identifier) @name) @class
		(struct_specifier name: (type_identifier) @name) @class
	`,
	LangC: `
		(function_definition declarator: (function_declarator declarator: (identifier) @name)) @function
		(struct_specifier name: (type_identifier) @name) @class
	`,
}

func extractSymbolsWithQueries(p *Parser, result *ParseResult, structure *CodeStructure) error {
	queryStr, ok := languageQueries[result.Language]
	if !ok {
		return fmt.Errorf("no queries for language %s", result.Language)
	}

	matches, err := p.ExecuteQuery(result.Language, result.Tree.RootNode(), queryStr)
	if err != nil {
		return err
	}

	for _, match := range matches {
		var categoryNode *sitter.Node
		var categoryName string
		var nameNode *sitter.Node

		for _, cap := range match.Captures {
			cName := p.queries[fmt.Sprintf("%s:%s", result.Language, queryStr)].CaptureNameForId(cap.Index)
			if cName == "name" {
				nameNode = cap.Node
			} else {
				categoryNode = cap.Node
				categoryName = cName
			}
		}

		if categoryNode == nil {
			continue
		}

		symbol := Symbol{
			Name:      "",
			StartLine: categoryNode.StartPoint().Row + 1,
			EndLine:   categoryNode.EndPoint().Row + 1,
			StartByte: categoryNode.StartByte(),
			EndByte:   categoryNode.EndByte(),
		}

		if nameNode != nil {
			symbol.Name = getNodeText(nameNode, result.Source)
		}

		// Heuristic name resolution if @name capture wasn't used or returned something too broad
		if symbol.Name == "" || (len(symbol.Name) > 100 && categoryName != "import") {
			if n := getChildByType(categoryNode, "identifier"); n != nil {
				symbol.Name = getNodeText(n, result.Source)
			} else if n := getChildByType(categoryNode, "type_identifier"); n != nil {
				symbol.Name = getNodeText(n, result.Source)
			} else if n := getChildByType(categoryNode, "field_identifier"); n != nil {
				symbol.Name = getNodeText(n, result.Source)
			} else if n := getChildByType(categoryNode, "name"); n != nil {
				symbol.Name = getNodeText(n, result.Source)
			} else if list := getChildByType(categoryNode, "identifier_list"); list != nil {
				if ident := getChildByType(list, "identifier"); ident != nil {
					symbol.Name = getNodeText(ident, result.Source)
				}
			}
		}

		// Final fallback: use node type or snippet (only for very specific categories)
		if symbol.Name == "" {
			if categoryName == "import" || categoryName == "export" {
				symbol.Name = getNodeText(categoryNode, result.Source)
			} else {
				continue // Skip if we can't find a decent name
			}
		}

		switch categoryName {
		case "import":
			symbol.Kind = SymbolImport
			structure.Imports = append(structure.Imports, symbol)
		case "export":
			symbol.Kind = SymbolExport
			symbol.Exported = true
			structure.Exports = append(structure.Exports, symbol)
		case "function", "method", "class", "interface", "struct", "type", "enum":
			// Determine base kind
			switch categoryName {
			case "function":
				symbol.Kind = SymbolFunction
			case "method":
				symbol.Kind = SymbolMethod
			case "class":
				symbol.Kind = SymbolClass
			case "interface":
				symbol.Kind = SymbolInterface
			case "struct":
				symbol.Kind = SymbolStruct
			case "type":
				symbol.Kind = SymbolType
			case "enum":
				symbol.Kind = SymbolEnum
			}

			// Language-specific adjustments
			if result.Language == LangPython {
				symbol.Exported = !isPythonPrivate(symbol.Name)
			} else if result.Language == LangGo {
				symbol.Exported = isGoExported(symbol.Name)
				if symbol.Kind == SymbolType {
					if getChildByType(categoryNode, "struct_type") != nil {
						symbol.Kind = SymbolStruct
					} else if getChildByType(categoryNode, "interface_type") != nil {
						symbol.Kind = SymbolInterface
					}
				}
			}

			switch symbol.Kind {
			case SymbolClass, SymbolStruct:
				if (result.Language == LangGo || result.Language == LangRust) && symbol.Kind == SymbolStruct {
					structure.Types = append(structure.Types, symbol)
				} else {
					structure.Classes = append(structure.Classes, symbol)
				}
			case SymbolFunction, SymbolMethod:
				structure.Functions = append(structure.Functions, symbol)
			case SymbolType, SymbolInterface, SymbolEnum:
				structure.Types = append(structure.Types, symbol)
			}
		case "variable":
			symbol.Kind = SymbolVariable
			if result.Language == LangGo {
				symbol.Exported = isGoExported(symbol.Name)
			}
			structure.Variables = append(structure.Variables, symbol)
		case "constant":
			symbol.Kind = SymbolConstant
			if result.Language == LangGo {
				symbol.Exported = isGoExported(symbol.Name)
			}
			structure.Variables = append(structure.Variables, symbol)
		}
	}

	return nil
}

func countLines(source []byte) int {
	lines := 1
	for _, b := range source {
		if b == '\n' {
			lines++
		}
	}
	return lines
}

func extractJavaScriptSymbols(root *sitter.Node, source []byte, structure *CodeStructure) {
	walkTree(root, func(node *sitter.Node) bool {
		nodeType := node.Type()

		switch nodeType {
		case "import_statement":
			structure.Imports = append(structure.Imports, Symbol{
				Name:      getNodeText(node, source),
				Kind:      SymbolImport,
				StartLine: node.StartPoint().Row + 1,
				EndLine:   node.EndPoint().Row + 1,
				StartByte: node.StartByte(),
				EndByte:   node.EndByte(),
			})

		case "export_statement":
			structure.Exports = append(structure.Exports, Symbol{
				Name:      getNodeText(node, source),
				Kind:      SymbolExport,
				StartLine: node.StartPoint().Row + 1,
				EndLine:   node.EndPoint().Row + 1,
				StartByte: node.StartByte(),
				EndByte:   node.EndByte(),
				Exported:  true,
			})

		case "function_declaration", "arrow_function", "function":
			name := getFunctionName(node, source)
			if name != "" {
				structure.Functions = append(structure.Functions, Symbol{
					Name:      name,
					Kind:      SymbolFunction,
					StartLine: node.StartPoint().Row + 1,
					EndLine:   node.EndPoint().Row + 1,
					StartByte: node.StartByte(),
					EndByte:   node.EndByte(),
					Signature: getFunctionSignature(node, source),
					Exported:  isExported(node),
				})
			}

		case "class_declaration":
			name := getClassName(node, source)
			if name != "" {
				structure.Classes = append(structure.Classes, Symbol{
					Name:      name,
					Kind:      SymbolClass,
					StartLine: node.StartPoint().Row + 1,
					EndLine:   node.EndPoint().Row + 1,
					StartByte: node.StartByte(),
					EndByte:   node.EndByte(),
					Exported:  isExported(node),
					Children:  extractClassMethods(node, source),
				})
			}

		case "lexical_declaration", "variable_declaration":
			vars := extractVariables(node, source)
			structure.Variables = append(structure.Variables, vars...)
		}

		return true // continue walking
	})
}

func extractTypeScriptSymbols(root *sitter.Node, source []byte, structure *CodeStructure) {
	// TypeScript includes all JavaScript symbols plus types/interfaces
	extractJavaScriptSymbols(root, source, structure)

	walkTree(root, func(node *sitter.Node) bool {
		nodeType := node.Type()

		switch nodeType {
		case "interface_declaration":
			name := getInterfaceName(node, source)
			if name != "" {
				structure.Types = append(structure.Types, Symbol{
					Name:      name,
					Kind:      SymbolInterface,
					StartLine: node.StartPoint().Row + 1,
					EndLine:   node.EndPoint().Row + 1,
					StartByte: node.StartByte(),
					EndByte:   node.EndByte(),
					Exported:  isExported(node),
				})
			}

		case "type_alias_declaration":
			name := getTypeAliasName(node, source)
			if name != "" {
				structure.Types = append(structure.Types, Symbol{
					Name:      name,
					Kind:      SymbolType,
					StartLine: node.StartPoint().Row + 1,
					EndLine:   node.EndPoint().Row + 1,
					StartByte: node.StartByte(),
					EndByte:   node.EndByte(),
					Exported:  isExported(node),
				})
			}

		case "enum_declaration":
			name := getEnumName(node, source)
			if name != "" {
				structure.Types = append(structure.Types, Symbol{
					Name:      name,
					Kind:      SymbolEnum,
					StartLine: node.StartPoint().Row + 1,
					EndLine:   node.EndPoint().Row + 1,
					StartByte: node.StartByte(),
					EndByte:   node.EndByte(),
					Exported:  isExported(node),
				})
			}
		}

		return true
	})
}

func extractGoSymbols(root *sitter.Node, source []byte, structure *CodeStructure) {
	walkTree(root, func(node *sitter.Node) bool {
		nodeType := node.Type()

		switch nodeType {
		case "import_declaration":
			structure.Imports = append(structure.Imports, Symbol{
				Name:      getNodeText(node, source),
				Kind:      SymbolImport,
				StartLine: node.StartPoint().Row + 1,
				EndLine:   node.EndPoint().Row + 1,
				StartByte: node.StartByte(),
				EndByte:   node.EndByte(),
			})

		case "function_declaration":
			name := getGoFunctionName(node, source)
			if name != "" {
				structure.Functions = append(structure.Functions, Symbol{
					Name:      name,
					Kind:      SymbolFunction,
					StartLine: node.StartPoint().Row + 1,
					EndLine:   node.EndPoint().Row + 1,
					StartByte: node.StartByte(),
					EndByte:   node.EndByte(),
					Signature: getGoFunctionSignature(node, source),
					Exported:  isGoExported(name),
				})
			}

		case "method_declaration":
			name := getGoMethodName(node, source)
			receiver := getGoMethodReceiver(node, source)
			if name != "" {
				structure.Functions = append(structure.Functions, Symbol{
					Name:      name,
					Kind:      SymbolMethod,
					Parent:    receiver,
					StartLine: node.StartPoint().Row + 1,
					EndLine:   node.EndPoint().Row + 1,
					StartByte: node.StartByte(),
					EndByte:   node.EndByte(),
					Exported:  isGoExported(name),
				})
			}

		case "type_declaration":
			typeSpecs := extractGoTypeSpecs(node, source)
			structure.Types = append(structure.Types, typeSpecs...)

		case "var_declaration", "const_declaration":
			vars := extractGoVariables(node, source)
			structure.Variables = append(structure.Variables, vars...)
		}

		return true
	})
}

func extractPythonSymbols(root *sitter.Node, source []byte, structure *CodeStructure) {
	walkTree(root, func(node *sitter.Node) bool {
		nodeType := node.Type()

		switch nodeType {
		case "import_statement", "import_from_statement":
			structure.Imports = append(structure.Imports, Symbol{
				Name:      getNodeText(node, source),
				Kind:      SymbolImport,
				StartLine: node.StartPoint().Row + 1,
				EndLine:   node.EndPoint().Row + 1,
				StartByte: node.StartByte(),
				EndByte:   node.EndByte(),
			})

		case "function_definition":
			name := getPythonFunctionName(node, source)
			if name != "" {
				structure.Functions = append(structure.Functions, Symbol{
					Name:      name,
					Kind:      SymbolFunction,
					StartLine: node.StartPoint().Row + 1,
					EndLine:   node.EndPoint().Row + 1,
					StartByte: node.StartByte(),
					EndByte:   node.EndByte(),
					Exported:  !isPythonPrivate(name),
				})
			}

		case "class_definition":
			name := getPythonClassName(node, source)
			if name != "" {
				structure.Classes = append(structure.Classes, Symbol{
					Name:      name,
					Kind:      SymbolClass,
					StartLine: node.StartPoint().Row + 1,
					EndLine:   node.EndPoint().Row + 1,
					StartByte: node.StartByte(),
					EndByte:   node.EndByte(),
					Exported:  !isPythonPrivate(name),
				})
			}
		}

		return true
	})
}

func extractRustSymbols(root *sitter.Node, source []byte, structure *CodeStructure) {
	// Use generic for now as query-based extraction handles the heavy lifting
	extractGenericSymbols(root, source, structure)
}

func extractPHPSymbols(root *sitter.Node, source []byte, structure *CodeStructure) {
	extractGenericSymbols(root, source, structure)
}

func extractCPPSymbols(root *sitter.Node, source []byte, structure *CodeStructure) {
	extractGenericSymbols(root, source, structure)
}

func extractGenericSymbols(root *sitter.Node, source []byte, structure *CodeStructure) {
	// Generic extraction for unsupported languages - just walk and find common patterns
	walkTree(root, func(node *sitter.Node) bool {
		nodeType := node.Type()

		if containsWord(nodeType, "function", "method", "def") {
			structure.Functions = append(structure.Functions, Symbol{
				Name:      getNodeText(node, source),
				Kind:      SymbolFunction,
				StartLine: node.StartPoint().Row + 1,
				EndLine:   node.EndPoint().Row + 1,
			})
		}

		if containsWord(nodeType, "class", "struct", "type") {
			structure.Classes = append(structure.Classes, Symbol{
				Name:      getNodeText(node, source),
				Kind:      SymbolClass,
				StartLine: node.StartPoint().Row + 1,
				EndLine:   node.EndPoint().Row + 1,
			})
		}

		if containsWord(nodeType, "import", "include", "require") {
			structure.Imports = append(structure.Imports, Symbol{
				Name:      getNodeText(node, source),
				Kind:      SymbolImport,
				StartLine: node.StartPoint().Row + 1,
				EndLine:   node.EndPoint().Row + 1,
			})
		}

		return true
	})
}

// Helper functions

func walkTree(node *sitter.Node, fn func(*sitter.Node) bool) {
	if !fn(node) {
		return
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		walkTree(node.Child(i), fn)
	}
}

func getNodeText(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if int(end) > len(source) {
		end = uint32(len(source))
	}
	return string(source[start:end])
}

func getChildByType(node *sitter.Node, childType string) *sitter.Node {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == childType {
			return child
		}
	}
	return nil
}

func getFunctionName(node *sitter.Node, source []byte) string {
	nameNode := getChildByType(node, "identifier")
	if nameNode != nil {
		return getNodeText(nameNode, source)
	}
	return ""
}

func getFunctionSignature(node *sitter.Node, source []byte) string {
	params := getChildByType(node, "formal_parameters")
	if params != nil {
		return getNodeText(params, source)
	}
	return "()"
}

func getClassName(node *sitter.Node, source []byte) string {
	nameNode := getChildByType(node, "identifier")
	if nameNode != nil {
		return getNodeText(nameNode, source)
	}
	return ""
}

func getInterfaceName(node *sitter.Node, source []byte) string {
	nameNode := getChildByType(node, "type_identifier")
	if nameNode != nil {
		return getNodeText(nameNode, source)
	}
	return ""
}

func getTypeAliasName(node *sitter.Node, source []byte) string {
	nameNode := getChildByType(node, "type_identifier")
	if nameNode != nil {
		return getNodeText(nameNode, source)
	}
	return ""
}

func getEnumName(node *sitter.Node, source []byte) string {
	nameNode := getChildByType(node, "identifier")
	if nameNode != nil {
		return getNodeText(nameNode, source)
	}
	return ""
}

func isExported(node *sitter.Node) bool {
	parent := node.Parent()
	if parent != nil && parent.Type() == "export_statement" {
		return true
	}
	return false
}

func extractClassMethods(node *sitter.Node, source []byte) []Symbol {
	var methods []Symbol
	body := getChildByType(node, "class_body")
	if body == nil {
		return methods
	}

	for i := 0; i < int(body.ChildCount()); i++ {
		child := body.Child(i)
		if child.Type() == "method_definition" {
			name := getFunctionName(child, source)
			if name != "" {
				methods = append(methods, Symbol{
					Name:      name,
					Kind:      SymbolMethod,
					StartLine: child.StartPoint().Row + 1,
					EndLine:   child.EndPoint().Row + 1,
				})
			}
		}
	}
	return methods
}

func extractVariables(node *sitter.Node, source []byte) []Symbol {
	var vars []Symbol
	walkTree(node, func(n *sitter.Node) bool {
		if n.Type() == "variable_declarator" {
			nameNode := getChildByType(n, "identifier")
			if nameNode != nil {
				vars = append(vars, Symbol{
					Name:      getNodeText(nameNode, source),
					Kind:      SymbolVariable,
					StartLine: n.StartPoint().Row + 1,
					EndLine:   n.EndPoint().Row + 1,
				})
			}
		}
		return true
	})
	return vars
}

// Go-specific helpers

func getGoFunctionName(node *sitter.Node, source []byte) string {
	nameNode := getChildByType(node, "identifier")
	if nameNode != nil {
		return getNodeText(nameNode, source)
	}
	return ""
}

func getGoFunctionSignature(node *sitter.Node, source []byte) string {
	params := getChildByType(node, "parameter_list")
	if params != nil {
		return getNodeText(params, source)
	}
	return "()"
}

func getGoMethodName(node *sitter.Node, source []byte) string {
	nameNode := getChildByType(node, "field_identifier")
	if nameNode != nil {
		return getNodeText(nameNode, source)
	}
	return ""
}

func getGoMethodReceiver(node *sitter.Node, source []byte) string {
	params := getChildByType(node, "parameter_list")
	if params != nil {
		typeNode := getChildByType(params, "type_identifier")
		if typeNode != nil {
			return getNodeText(typeNode, source)
		}
		ptrType := getChildByType(params, "pointer_type")
		if ptrType != nil {
			return getNodeText(ptrType, source)
		}
	}
	return ""
}

func isGoExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

func extractGoTypeSpecs(node *sitter.Node, source []byte) []Symbol {
	var types []Symbol
	walkTree(node, func(n *sitter.Node) bool {
		if n.Type() == "type_spec" {
			nameNode := getChildByType(n, "type_identifier")
			if nameNode != nil {
				name := getNodeText(nameNode, source)
				kind := SymbolType
				if getChildByType(n, "struct_type") != nil {
					kind = SymbolStruct
				} else if getChildByType(n, "interface_type") != nil {
					kind = SymbolInterface
				}
				types = append(types, Symbol{
					Name:      name,
					Kind:      kind,
					StartLine: n.StartPoint().Row + 1,
					EndLine:   n.EndPoint().Row + 1,
					Exported:  isGoExported(name),
				})
			}
		}
		return true
	})
	return types
}

func extractGoVariables(node *sitter.Node, source []byte) []Symbol {
	var vars []Symbol
	isConst := node.Type() == "const_declaration"

	walkTree(node, func(n *sitter.Node) bool {
		if n.Type() == "var_spec" || n.Type() == "const_spec" {
			nameNode := getChildByType(n, "identifier")
			if nameNode != nil {
				name := getNodeText(nameNode, source)
				kind := SymbolVariable
				if isConst {
					kind = SymbolConstant
				}
				vars = append(vars, Symbol{
					Name:      name,
					Kind:      kind,
					StartLine: n.StartPoint().Row + 1,
					EndLine:   n.EndPoint().Row + 1,
					Exported:  isGoExported(name),
				})
			}
		}
		return true
	})
	return vars
}

// Python-specific helpers

func getPythonFunctionName(node *sitter.Node, source []byte) string {
	nameNode := getChildByType(node, "identifier")
	if nameNode != nil {
		return getNodeText(nameNode, source)
	}
	return ""
}

func getPythonClassName(node *sitter.Node, source []byte) string {
	nameNode := getChildByType(node, "identifier")
	if nameNode != nil {
		return getNodeText(nameNode, source)
	}
	return ""
}

func isPythonPrivate(name string) bool {
	return len(name) > 0 && name[0] == '_'
}

func containsWord(s string, words ...string) bool {
	for _, word := range words {
		if s == word || len(s) > len(word) && (s[:len(word)] == word || s[len(s)-len(word):] == word) {
			return true
		}
	}
	return false
}
