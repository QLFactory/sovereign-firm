package treesitter

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Framework represents a detected framework or library
type Framework string

const (
	FrameworkReact    Framework = "react"
	FrameworkNextJS   Framework = "nextjs"
	FrameworkVue      Framework = "vue"
	FrameworkAngular  Framework = "angular"
	FrameworkExpress  Framework = "express"
	FrameworkFastify  Framework = "fastify"
	FrameworkGin      Framework = "gin"
	FrameworkEcho     Framework = "echo"
	FrameworkFiber    Framework = "fiber"
	FrameworkDjango   Framework = "django"
	FrameworkFlask    Framework = "flask"
	FrameworkFastAPI  Framework = "fastapi"
	FrameworkRust     Framework = "rust"
	FrameworkTailwind Framework = "tailwindcss"
	FrameworkVite     Framework = "vite"
	FrameworkWebpack  Framework = "webpack"
	FrameworkVitest   Framework = "vitest"
	FrameworkJest     Framework = "jest"
	FrameworkUnknown  Framework = "unknown"
)

// ProjectStack represents the detected technology stack of a project
type ProjectStack struct {
	PrimaryLanguage Language    `json:"primary_language"`
	Languages       []Language  `json:"languages"`
	Frameworks      []Framework `json:"frameworks"`
	TestFrameworks  []Framework `json:"test_frameworks"`
	BuildTools      []Framework `json:"build_tools"`
	PackageManager  string      `json:"package_manager"`
	HasTypeScript   bool        `json:"has_typescript"`
	HasTests        bool        `json:"has_tests"`
	IsMonorepo      bool        `json:"is_monorepo"`
	EntryPoints     []string    `json:"entry_points"`
	ConfigFiles     []string    `json:"config_files"`
}

// ProjectAnalyzer analyzes a project to detect its technology stack
type ProjectAnalyzer struct {
	parser *Parser
}

// NewProjectAnalyzer creates a new project analyzer
func NewProjectAnalyzer() *ProjectAnalyzer {
	return &ProjectAnalyzer{
		parser: NewParser(),
	}
}

// AnalyzeProject analyzes a project directory and returns its stack
func (a *ProjectAnalyzer) AnalyzeProject(ctx context.Context, projectDir string) (*ProjectStack, error) {
	stack := &ProjectStack{
		Languages:      []Language{},
		Frameworks:     []Framework{},
		TestFrameworks: []Framework{},
		BuildTools:     []Framework{},
		ConfigFiles:    []string{},
		EntryPoints:    []string{},
	}

	// Check for package managers and config files
	a.detectPackageManager(projectDir, stack)
	a.detectConfigFiles(projectDir, stack)

	// Detect frameworks from package.json
	a.detectJSFrameworks(projectDir, stack)

	// Detect Go frameworks from go.mod
	a.detectGoFrameworks(projectDir, stack)

	// Detect Python frameworks from requirements.txt or pyproject.toml
	a.detectPythonFrameworks(projectDir, stack)

	// Scan source files for language distribution
	a.scanSourceFiles(ctx, projectDir, stack)

	// Determine primary language
	stack.PrimaryLanguage = a.determinePrimaryLanguage(stack)

	// Detect entry points
	a.detectEntryPoints(projectDir, stack)

	return stack, nil
}

func (a *ProjectAnalyzer) detectPackageManager(projectDir string, stack *ProjectStack) {
	// Node.js package managers
	if fileExists(filepath.Join(projectDir, "package-lock.json")) {
		stack.PackageManager = "npm"
	} else if fileExists(filepath.Join(projectDir, "yarn.lock")) {
		stack.PackageManager = "yarn"
	} else if fileExists(filepath.Join(projectDir, "pnpm-lock.yaml")) {
		stack.PackageManager = "pnpm"
	} else if fileExists(filepath.Join(projectDir, "bun.lockb")) {
		stack.PackageManager = "bun"
	}

	// Go
	if fileExists(filepath.Join(projectDir, "go.mod")) {
		if stack.PackageManager == "" {
			stack.PackageManager = "go"
		}
	}

	// Python
	if fileExists(filepath.Join(projectDir, "poetry.lock")) {
		stack.PackageManager = "poetry"
	} else if fileExists(filepath.Join(projectDir, "Pipfile.lock")) {
		stack.PackageManager = "pipenv"
	} else if fileExists(filepath.Join(projectDir, "requirements.txt")) {
		if stack.PackageManager == "" {
			stack.PackageManager = "pip"
		}
	}

	// Rust
	if fileExists(filepath.Join(projectDir, "Cargo.lock")) {
		stack.PackageManager = "cargo"
	}
}

func (a *ProjectAnalyzer) detectConfigFiles(projectDir string, stack *ProjectStack) {
	configPatterns := []string{
		"package.json",
		"tsconfig.json",
		"vite.config.*",
		"next.config.*",
		"tailwind.config.*",
		"vitest.config.*",
		"jest.config.*",
		"webpack.config.*",
		"go.mod",
		"requirements.txt",
		"pyproject.toml",
		"Cargo.toml",
		"Dockerfile",
		"docker-compose.yaml",
		"docker-compose.yml",
		".env",
		".env.example",
	}

	for _, pattern := range configPatterns {
		matches, _ := filepath.Glob(filepath.Join(projectDir, pattern))
		for _, match := range matches {
			relPath, _ := filepath.Rel(projectDir, match)
			stack.ConfigFiles = append(stack.ConfigFiles, relPath)
		}
	}

	// Check for TypeScript
	if fileExists(filepath.Join(projectDir, "tsconfig.json")) {
		stack.HasTypeScript = true
	}
}

func (a *ProjectAnalyzer) detectJSFrameworks(projectDir string, stack *ProjectStack) {
	pkgPath := filepath.Join(projectDir, "package.json")
	if !fileExists(pkgPath) {
		return
	}

	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	allDeps := make(map[string]bool)
	for dep := range pkg.Dependencies {
		allDeps[dep] = true
	}
	for dep := range pkg.DevDependencies {
		allDeps[dep] = true
	}

	// Frontend frameworks
	if allDeps["react"] || allDeps["react-dom"] {
		stack.Frameworks = append(stack.Frameworks, FrameworkReact)
	}
	if allDeps["next"] {
		stack.Frameworks = append(stack.Frameworks, FrameworkNextJS)
	}
	if allDeps["vue"] {
		stack.Frameworks = append(stack.Frameworks, FrameworkVue)
	}
	if allDeps["@angular/core"] {
		stack.Frameworks = append(stack.Frameworks, FrameworkAngular)
	}

	// Backend frameworks
	if allDeps["express"] {
		stack.Frameworks = append(stack.Frameworks, FrameworkExpress)
	}
	if allDeps["fastify"] {
		stack.Frameworks = append(stack.Frameworks, FrameworkFastify)
	}

	// Build tools
	if allDeps["vite"] {
		stack.BuildTools = append(stack.BuildTools, FrameworkVite)
	}
	if allDeps["webpack"] {
		stack.BuildTools = append(stack.BuildTools, FrameworkWebpack)
	}

	// CSS frameworks
	if allDeps["tailwindcss"] {
		stack.Frameworks = append(stack.Frameworks, FrameworkTailwind)
	}

	// Test frameworks
	if allDeps["vitest"] {
		stack.TestFrameworks = append(stack.TestFrameworks, FrameworkVitest)
		stack.HasTests = true
	}
	if allDeps["jest"] {
		stack.TestFrameworks = append(stack.TestFrameworks, FrameworkJest)
		stack.HasTests = true
	}

	// Add JavaScript/TypeScript to languages
	if stack.HasTypeScript {
		stack.Languages = append(stack.Languages, LangTypeScript)
	}
	stack.Languages = append(stack.Languages, LangJavaScript)
}

func (a *ProjectAnalyzer) detectGoFrameworks(projectDir string, stack *ProjectStack) {
	modPath := filepath.Join(projectDir, "go.mod")
	if !fileExists(modPath) {
		return
	}

	data, err := os.ReadFile(modPath)
	if err != nil {
		return
	}

	content := string(data)

	// Check for Go frameworks
	if strings.Contains(content, "github.com/gin-gonic/gin") {
		stack.Frameworks = append(stack.Frameworks, FrameworkGin)
	}
	if strings.Contains(content, "github.com/labstack/echo") {
		stack.Frameworks = append(stack.Frameworks, FrameworkEcho)
	}
	if strings.Contains(content, "github.com/gofiber/fiber") {
		stack.Frameworks = append(stack.Frameworks, FrameworkFiber)
	}

	stack.Languages = append(stack.Languages, LangGo)
}

func (a *ProjectAnalyzer) detectPythonFrameworks(projectDir string, stack *ProjectStack) {
	// Check requirements.txt
	reqPath := filepath.Join(projectDir, "requirements.txt")
	if fileExists(reqPath) {
		data, err := os.ReadFile(reqPath)
		if err == nil {
			content := strings.ToLower(string(data))
			a.detectPythonDeps(content, stack)
		}
	}

	// Check pyproject.toml
	pyprojectPath := filepath.Join(projectDir, "pyproject.toml")
	if fileExists(pyprojectPath) {
		data, err := os.ReadFile(pyprojectPath)
		if err == nil {
			content := strings.ToLower(string(data))
			a.detectPythonDeps(content, stack)
		}
	}
}

func (a *ProjectAnalyzer) detectPythonDeps(content string, stack *ProjectStack) {
	hasPython := false

	if strings.Contains(content, "django") {
		stack.Frameworks = append(stack.Frameworks, FrameworkDjango)
		hasPython = true
	}
	if strings.Contains(content, "flask") {
		stack.Frameworks = append(stack.Frameworks, FrameworkFlask)
		hasPython = true
	}
	if strings.Contains(content, "fastapi") {
		stack.Frameworks = append(stack.Frameworks, FrameworkFastAPI)
		hasPython = true
	}
	if strings.Contains(content, "pytest") {
		stack.HasTests = true
		hasPython = true
	}

	if hasPython {
		stack.Languages = append(stack.Languages, LangPython)
	}
}

func (a *ProjectAnalyzer) scanSourceFiles(ctx context.Context, projectDir string, stack *ProjectStack) {
	langCount := make(map[Language]int)

	filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip common non-source directories
		if info.IsDir() {
			base := info.Name()
			if base == "node_modules" || base == ".git" || base == "vendor" ||
				base == "__pycache__" || base == ".next" || base == "dist" ||
				base == "build" || base == "target" {
				return filepath.SkipDir
			}
			return nil
		}

		lang := DetectLanguage(path, nil)
		if lang != LangUnknown && lang != LangMarkdown {
			langCount[lang]++

			// Check for test files
			base := filepath.Base(path)
			if strings.Contains(base, ".test.") || strings.Contains(base, "_test.") ||
				strings.Contains(base, ".spec.") || strings.HasPrefix(base, "test_") {
				stack.HasTests = true
			}
		}

		return nil
	})

	// Add unique languages to stack
	for lang := range langCount {
		found := false
		for _, existing := range stack.Languages {
			if existing == lang {
				found = true
				break
			}
		}
		if !found {
			stack.Languages = append(stack.Languages, lang)
		}
	}
}

func (a *ProjectAnalyzer) determinePrimaryLanguage(stack *ProjectStack) Language {
	// Priority order for primary language detection
	priority := []Language{
		LangGo, LangRust, LangPython,
		LangTypeScript, LangJavaScript,
	}

	for _, lang := range priority {
		for _, l := range stack.Languages {
			if l == lang {
				return lang
			}
		}
	}

	if len(stack.Languages) > 0 {
		return stack.Languages[0]
	}

	return LangUnknown
}

func (a *ProjectAnalyzer) detectEntryPoints(projectDir string, stack *ProjectStack) {
	// Common entry points
	entryPatterns := []string{
		"main.go",
		"cmd/*/main.go",
		"src/main.ts",
		"src/main.tsx",
		"src/index.ts",
		"src/index.tsx",
		"src/index.js",
		"src/index.jsx",
		"src/App.tsx",
		"src/App.jsx",
		"app/page.tsx",    // Next.js
		"pages/index.tsx", // Next.js pages router
		"main.py",
		"app.py",
		"manage.py", // Django
		"src/main.rs",
		"src/lib.rs",
	}

	for _, pattern := range entryPatterns {
		matches, _ := filepath.Glob(filepath.Join(projectDir, pattern))
		for _, match := range matches {
			relPath, _ := filepath.Rel(projectDir, match)
			stack.EntryPoints = append(stack.EntryPoints, relPath)
		}
	}

	// Check for monorepo
	if fileExists(filepath.Join(projectDir, "packages")) ||
		fileExists(filepath.Join(projectDir, "apps")) ||
		fileExists(filepath.Join(projectDir, "lerna.json")) ||
		fileExists(filepath.Join(projectDir, "pnpm-workspace.yaml")) {
		stack.IsMonorepo = true
	}
}

// AnalyzeFile analyzes a single file and returns its code structure
func (a *ProjectAnalyzer) AnalyzeFile(ctx context.Context, filePath string, source []byte) (*CodeStructure, error) {
	result, err := a.parser.Parse(ctx, filePath, source)
	if err != nil {
		return nil, err
	}

	return ExtractSymbols(a.parser, result)
}

// AnalyzeFiles analyzes multiple files and returns their combined structure
func (a *ProjectAnalyzer) AnalyzeFiles(ctx context.Context, files map[string]string) (*ProjectAnalysis, error) {
	analysis := &ProjectAnalysis{
		Files:     make(map[string]*CodeStructure),
		Languages: make(map[Language]int),
		Summary:   &AnalysisSummary{},
	}

	for path, content := range files {
		result, err := a.parser.Parse(ctx, path, []byte(content))
		if err != nil {
			continue // Skip unparseable files
		}

		structure, err := ExtractSymbols(a.parser, result)
		if err != nil {
			continue
		}

		analysis.Files[path] = structure
		analysis.Languages[result.Language]++
		analysis.Summary.TotalFiles++
		analysis.Summary.TotalLOC += structure.LOC
		analysis.Summary.TotalFunctions += len(structure.Functions)
		analysis.Summary.TotalClasses += len(structure.Classes)
	}

	return analysis, nil
}

// ProjectAnalysis contains the full analysis of a project
type ProjectAnalysis struct {
	Files     map[string]*CodeStructure `json:"files"`
	Languages map[Language]int          `json:"languages"`
	Summary   *AnalysisSummary          `json:"summary"`
}

// AnalysisSummary provides high-level metrics
type AnalysisSummary struct {
	TotalFiles     int `json:"total_files"`
	TotalLOC       int `json:"total_loc"`
	TotalFunctions int `json:"total_functions"`
	TotalClasses   int `json:"total_classes"`
}

// Close releases analyzer resources
func (a *ProjectAnalyzer) Close() {
	a.parser.Close()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
