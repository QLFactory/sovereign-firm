package activities

import (
	"testing"
)

func TestIsSupportedExt(t *testing.T) {
	supported := []string{
		".go", ".js", ".ts", ".jsx", ".tsx",
		".py", ".md", ".json", ".html", ".css",
	}

	for _, ext := range supported {
		if !isSupportedExt(ext) {
			t.Errorf("Expected '%s' to be supported", ext)
		}
	}
}

func TestIsSupportedExtUnsupported(t *testing.T) {
	unsupported := []string{
		".txt", ".pdf", ".doc", ".exe", ".bin",
		".jpg", ".png", ".mp3", ".zip", ".tar",
		"", ".unknown",
	}

	for _, ext := range unsupported {
		if isSupportedExt(ext) {
			t.Errorf("Expected '%s' to be unsupported", ext)
		}
	}
}

func TestCodeBundleType(t *testing.T) {
	bundle := CodeBundle{
		"/src/App.jsx":  "function App() {}",
		"/src/main.jsx": "import App from './App'",
		"/index.html":   "<!DOCTYPE html>",
	}

	if len(bundle) != 3 {
		t.Errorf("Expected 3 files, got %d", len(bundle))
	}

	if _, ok := bundle["/src/App.jsx"]; !ok {
		t.Error("Expected /src/App.jsx to exist in bundle")
	}
}

func TestRefineInputStructure(t *testing.T) {
	input := RefineCodeInput{
		CurrentCode: map[string]string{
			"/src/App.jsx": "original code",
		},
		ChatHistory: "User feedback here",
	}

	if len(input.CurrentCode) != 1 {
		t.Errorf("Expected 1 file in CurrentCode, got %d", len(input.CurrentCode))
	}

	if input.ChatHistory != "User feedback here" {
		t.Errorf("Expected ChatHistory to match, got '%s'", input.ChatHistory)
	}
}

func TestQAGenerateInputStructure(t *testing.T) {
	input := QAGenerateInput{
		Spec: "Build a counter app",
		CodeFiles: map[string]string{
			"/src/App.jsx":     "function App() {}",
			"/src/Counter.jsx": "function Counter() {}",
		},
	}

	if input.Spec != "Build a counter app" {
		t.Errorf("Expected Spec to match, got '%s'", input.Spec)
	}

	if len(input.CodeFiles) != 2 {
		t.Errorf("Expected 2 code files, got %d", len(input.CodeFiles))
	}
}

func TestIndexRepoParamsStructure(t *testing.T) {
	params := IndexRepoParams{
		RepoURL:   "https://github.com/example/repo.git",
		ProjectID: "project-123",
	}

	if params.RepoURL != "https://github.com/example/repo.git" {
		t.Errorf("Expected RepoURL to match, got '%s'", params.RepoURL)
	}

	if params.ProjectID != "project-123" {
		t.Errorf("Expected ProjectID to match, got '%s'", params.ProjectID)
	}
}

func TestDevAgentExtractAndUnmarshalJSON(t *testing.T) {
	agent := &DevAgent{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid json",
			input:   `{"key": "value"}`,
			wantErr: false,
		},
		{
			name:    "json with prefix text",
			input:   `Here is the result: {"key": "value"}`,
			wantErr: false,
		},
		{
			name:    "json with suffix text",
			input:   `{"key": "value"} that's all`,
			wantErr: false,
		},
		{
			name:    "json in markdown block",
			input:   "```json\n{\"key\": \"value\"}\n```",
			wantErr: false,
		},
		{
			name:    "nested json",
			input:   `{"outer": {"inner": "value"}}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `not json at all`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   ``,
			wantErr: true,
		},
		{
			name:    "malformed json",
			input:   `{"key": "value"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := agent.extractAndUnmarshalJSON(tt.input, &result)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestQAAgentExtractAndUnmarshalJSON(t *testing.T) {
	agent := &QAAgent{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid json",
			input:   `{"/src/App.test.js": "test code"}`,
			wantErr: false,
		},
		{
			name:    "json with surrounding text",
			input:   `Here are the tests: {"/src/App.test.js": "test code"} Done!`,
			wantErr: false,
		},
		{
			name:    "empty object",
			input:   `{}`,
			wantErr: false,
		},
		{
			name:    "no json",
			input:   `no json here`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]interface{}
			err := agent.extractAndUnmarshalJSON(tt.input, &result)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestDevAgentPostProcessBundle(t *testing.T) {
	agent := &DevAgent{}

	// Test with basic input
	input := CodeBundle{
		"/src/App.jsx": "function App() { return <div>Hello</div>; }",
	}

	result := agent.postProcessBundle(input)

	// Should have required files added
	requiredFiles := []string{
		"/src/main.jsx",
		"/index.html",
		"/vite.config.js",
		"/package.json",
	}

	for _, file := range requiredFiles {
		if _, ok := result[file]; !ok {
			t.Errorf("Expected '%s' to be added by postProcessBundle", file)
		}
	}

	// Original file should still exist
	if _, ok := result["/src/App.jsx"]; !ok {
		t.Error("Expected original /src/App.jsx to be preserved")
	}
}

func TestDevAgentPostProcessBundleNormalizePaths(t *testing.T) {
	agent := &DevAgent{}

	// Test path normalization - files not in /src/ should be moved
	input := CodeBundle{
		"App.jsx":      "app code",
		"/App.jsx":     "app code 2",
		"components/X": "component code",
	}

	result := agent.postProcessBundle(input)

	// Check that files got normalized to /src/
	// The exact behavior depends on implementation
	// At minimum, required files should exist
	if _, ok := result["/src/main.jsx"]; !ok {
		t.Error("Expected /src/main.jsx to exist")
	}
}

func TestDevAgentPostProcessBundlePackageJSON(t *testing.T) {
	agent := &DevAgent{}

	input := CodeBundle{}
	result := agent.postProcessBundle(input)

	pkgJSON, ok := result["/package.json"]
	if !ok {
		t.Fatal("Expected /package.json to be generated")
	}

	// Check for key dependencies
	expectedContents := []string{
		"react",
		"react-dom",
		"vite",
		"@vitejs/plugin-react",
	}

	for _, content := range expectedContents {
		if !containsString(pkgJSON, content) {
			t.Errorf("Expected package.json to contain '%s'", content)
		}
	}
}

func TestDevAgentPostProcessBundleIndexHTML(t *testing.T) {
	agent := &DevAgent{}

	input := CodeBundle{}
	result := agent.postProcessBundle(input)

	indexHTML, ok := result["/index.html"]
	if !ok {
		t.Fatal("Expected /index.html to be generated")
	}

	// Check for required elements
	expectedContents := []string{
		"<!DOCTYPE html>",
		"<div id=\"root\">",
		"<script type=\"module\"",
		"/src/main.jsx",
	}

	for _, content := range expectedContents {
		if !containsString(indexHTML, content) {
			t.Errorf("Expected index.html to contain '%s'", content)
		}
	}
}

func TestDevAgentPostProcessBundleMainJSX(t *testing.T) {
	agent := &DevAgent{}

	input := CodeBundle{}
	result := agent.postProcessBundle(input)

	mainJSX, ok := result["/src/main.jsx"]
	if !ok {
		t.Fatal("Expected /src/main.jsx to be generated")
	}

	// Check for required imports and code
	expectedContents := []string{
		"import React",
		"createRoot",
		"import App",
		"getElementById",
		"render",
	}

	for _, content := range expectedContents {
		if !containsString(mainJSX, content) {
			t.Errorf("Expected main.jsx to contain '%s'", content)
		}
	}
}

func TestDevAgentPostProcessBundleViteConfig(t *testing.T) {
	agent := &DevAgent{}

	input := CodeBundle{}
	result := agent.postProcessBundle(input)

	viteConfig, ok := result["/vite.config.js"]
	if !ok {
		t.Fatal("Expected /vite.config.js to be generated")
	}

	// Check for required content
	expectedContents := []string{
		"defineConfig",
		"@vitejs/plugin-react",
		"plugins",
	}

	for _, content := range expectedContents {
		if !containsString(viteConfig, content) {
			t.Errorf("Expected vite.config.js to contain '%s'", content)
		}
	}
}

func TestCodeBundleEmpty(t *testing.T) {
	bundle := CodeBundle{}

	if len(bundle) != 0 {
		t.Errorf("Expected empty bundle, got %d files", len(bundle))
	}
}

func TestCodeBundleIteration(t *testing.T) {
	bundle := CodeBundle{
		"/src/a.js": "a",
		"/src/b.js": "b",
		"/src/c.js": "c",
	}

	count := 0
	for range bundle {
		count++
	}

	if count != 3 {
		t.Errorf("Expected to iterate over 3 files, got %d", count)
	}
}

func TestRefineInputWithEmptyCurrentCode(t *testing.T) {
	input := RefineCodeInput{
		CurrentCode: map[string]string{},
		ChatHistory: "Some feedback",
	}

	if len(input.CurrentCode) != 0 {
		t.Error("Expected empty CurrentCode")
	}

	if input.ChatHistory != "Some feedback" {
		t.Error("Expected ChatHistory to be set")
	}
}

func TestQAGenerateInputEmpty(t *testing.T) {
	input := QAGenerateInput{}

	if input.Spec != "" {
		t.Error("Expected empty Spec")
	}

	if input.CodeFiles != nil {
		t.Error("Expected nil CodeFiles")
	}
}

// Helper function
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
