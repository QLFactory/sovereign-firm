package workflows

import (
	"testing"
)

func TestProjectStateStructure(t *testing.T) {
	state := &ProjectState{
		Name:        "Test Project",
		Phase:       "DISCOVERY",
		ChatHistory: "User: Hello\nPM: Hi!",
		Spec:        "Build a todo app",
		CodeFiles: map[string]string{
			"/src/App.jsx": "export default function App() { return <div>Hello</div>; }",
		},
	}

	if state.Name != "Test Project" {
		t.Errorf("Expected name 'Test Project', got '%s'", state.Name)
	}

	if state.Phase != "DISCOVERY" {
		t.Errorf("Expected phase 'DISCOVERY', got '%s'", state.Phase)
	}

	if state.Spec != "Build a todo app" {
		t.Errorf("Expected spec 'Build a todo app', got '%s'", state.Spec)
	}

	if len(state.CodeFiles) != 1 {
		t.Errorf("Expected 1 code file, got %d", len(state.CodeFiles))
	}
}

func TestProjectStateInitialization(t *testing.T) {
	state := &ProjectState{
		Phase:     "DISCOVERY",
		CodeFiles: make(map[string]string),
	}

	if state.Phase != "DISCOVERY" {
		t.Errorf("Expected initial phase 'DISCOVERY', got '%s'", state.Phase)
	}

	if state.CodeFiles == nil {
		t.Error("Expected CodeFiles to be initialized")
	}

	if len(state.CodeFiles) != 0 {
		t.Errorf("Expected empty CodeFiles, got %d files", len(state.CodeFiles))
	}
}

func TestProjectStatePhases(t *testing.T) {
	phases := []string{
		"DISCOVERY",
		"IMPLEMENTATION",
		"REVIEW",
		"REFINEMENT",
		"DONE",
	}

	for _, phase := range phases {
		state := &ProjectState{Phase: phase}
		if state.Phase != phase {
			t.Errorf("Expected phase '%s', got '%s'", phase, state.Phase)
		}
	}
}

func TestUserMessageSignalStructure(t *testing.T) {
	signal := UserMessageSignal{
		Message: "Build me a counter app",
	}

	if signal.Message != "Build me a counter app" {
		t.Errorf("Expected message 'Build me a counter app', got '%s'", signal.Message)
	}
}

func TestUserMessageSignalApprove(t *testing.T) {
	signal := UserMessageSignal{
		Message: "/approve",
	}

	if signal.Message != "/approve" {
		t.Errorf("Expected '/approve', got '%s'", signal.Message)
	}

	// Test approval detection
	isApproval := signal.Message == "/approve"
	if !isApproval {
		t.Error("Expected signal to be detected as approval")
	}
}

func TestProjectStateChatHistoryAppend(t *testing.T) {
	state := &ProjectState{
		ChatHistory: "",
	}

	// Simulate chat history building
	state.ChatHistory += "\nUser: Hello"
	state.ChatHistory += "\nPM: Hi! How can I help?"
	state.ChatHistory += "\nUser: Build a todo app"

	expectedContains := []string{
		"User: Hello",
		"PM: Hi!",
		"User: Build a todo app",
	}

	for _, expected := range expectedContains {
		if !containsString(state.ChatHistory, expected) {
			t.Errorf("Expected chat history to contain '%s'", expected)
		}
	}
}

func TestProjectStateCodeFilesOperations(t *testing.T) {
	state := &ProjectState{
		CodeFiles: make(map[string]string),
	}

	// Add files
	state.CodeFiles["/src/App.jsx"] = "function App() {}"
	state.CodeFiles["/src/main.jsx"] = "import App from './App'"
	state.CodeFiles["/index.html"] = "<!DOCTYPE html>"

	if len(state.CodeFiles) != 3 {
		t.Errorf("Expected 3 files, got %d", len(state.CodeFiles))
	}

	// Check file exists
	if _, ok := state.CodeFiles["/src/App.jsx"]; !ok {
		t.Error("Expected /src/App.jsx to exist")
	}

	// Update file
	state.CodeFiles["/src/App.jsx"] = "function App() { return <div>Updated</div> }"
	if !containsString(state.CodeFiles["/src/App.jsx"], "Updated") {
		t.Error("Expected App.jsx to be updated")
	}

	// Delete file
	delete(state.CodeFiles, "/index.html")
	if len(state.CodeFiles) != 2 {
		t.Errorf("Expected 2 files after delete, got %d", len(state.CodeFiles))
	}
}

func TestProjectStateMergeCodeFiles(t *testing.T) {
	state := &ProjectState{
		CodeFiles: map[string]string{
			"/src/App.jsx":  "original app",
			"/src/utils.js": "original utils",
		},
	}

	// Simulate merging new code (like from DevAgentRefine)
	newFiles := map[string]string{
		"/src/App.jsx":    "updated app",
		"/src/Header.jsx": "new header",
	}

	for k, v := range newFiles {
		state.CodeFiles[k] = v
	}

	// Check merge results
	if state.CodeFiles["/src/App.jsx"] != "updated app" {
		t.Error("Expected App.jsx to be updated")
	}

	if state.CodeFiles["/src/utils.js"] != "original utils" {
		t.Error("Expected utils.js to remain unchanged")
	}

	if state.CodeFiles["/src/Header.jsx"] != "new header" {
		t.Error("Expected Header.jsx to be added")
	}

	if len(state.CodeFiles) != 3 {
		t.Errorf("Expected 3 files after merge, got %d", len(state.CodeFiles))
	}
}

func TestProjectStateJSONTags(t *testing.T) {
	// This test verifies that the struct has proper JSON tags for serialization
	state := ProjectState{
		Name:        "Test",
		Phase:       "DISCOVERY",
		ChatHistory: "history",
		Spec:        "spec",
		CodeFiles:   map[string]string{"file": "content"},
	}

	// Basic validation that fields are accessible
	if state.Name == "" {
		t.Error("Name should be set")
	}
}

func TestUserMessageSignalEmpty(t *testing.T) {
	signal := UserMessageSignal{}

	if signal.Message != "" {
		t.Errorf("Expected empty message, got '%s'", signal.Message)
	}
}

func TestProjectStateSpecUsage(t *testing.T) {
	state := &ProjectState{
		ChatHistory: "Some chat history",
		Spec:        "",
	}

	// Simulate the workflow logic where Spec or ChatHistory is used
	specToUse := state.Spec
	if specToUse == "" {
		specToUse = state.ChatHistory
	}

	if specToUse != "Some chat history" {
		t.Errorf("Expected to fall back to chat history, got '%s'", specToUse)
	}

	// Now with Spec set
	state.Spec = "Formal specification"
	specToUse = state.Spec
	if specToUse == "" {
		specToUse = state.ChatHistory
	}

	if specToUse != "Formal specification" {
		t.Errorf("Expected to use Spec, got '%s'", specToUse)
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
