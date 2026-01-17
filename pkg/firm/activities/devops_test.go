package activities

import (
	"encoding/json"
	"strings"
	"testing"
)

// ============================================================================
// Data Structure Tests
// ============================================================================

func TestDevOpsCodeBundleStructure(t *testing.T) {
	bundle := DevOpsCodeBundle{
		Files: map[string]string{
			"Dockerfile":         "FROM node:20-alpine...",
			"docker-compose.yml": "version: '3.8'...",
		},
		BuildCmd:    "docker build -t app .",
		DeployCmd:   "docker-compose up -d",
		TestCmd:     "docker-compose run --rm app npm test",
		Environment: "production",
	}

	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("Failed to marshal DevOpsCodeBundle: %v", err)
	}

	var decoded DevOpsCodeBundle
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DevOpsCodeBundle: %v", err)
	}

	if len(decoded.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(decoded.Files))
	}
	if decoded.Environment != "production" {
		t.Errorf("Expected environment 'production', got '%s'", decoded.Environment)
	}
}

func TestDockerConfigStructure(t *testing.T) {
	config := DockerConfig{
		BaseImage: "node:20-alpine",
		BuildStages: []BuildStage{
			{Name: "builder", From: "node:20-alpine", Commands: []string{"npm ci", "npm run build"}},
			{Name: "production", From: "node:20-alpine", Commands: []string{"npm ci --production"}},
		},
		ExposedPorts: []int{3000, 9229},
		EnvVars: map[string]string{
			"NODE_ENV": "production",
		},
		Volumes: []string{"/app/data"},
		HealthCheck: &HealthCheck{
			Command:  "curl -f http://localhost:3000/health",
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal DockerConfig: %v", err)
	}

	var decoded DockerConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DockerConfig: %v", err)
	}

	if len(decoded.BuildStages) != 2 {
		t.Errorf("Expected 2 build stages, got %d", len(decoded.BuildStages))
	}
	if decoded.HealthCheck == nil {
		t.Error("Expected HealthCheck to be set")
	}
	if decoded.HealthCheck.Retries != 3 {
		t.Errorf("Expected 3 retries, got %d", decoded.HealthCheck.Retries)
	}
}

func TestComposeServiceStructure(t *testing.T) {
	service := ComposeService{
		Name:  "api",
		Build: "./api",
		Ports: []string{"3000:3000"},
		Environment: map[string]string{
			"DATABASE_URL": "postgres://localhost:5432/app",
		},
		DependsOn: []string{"postgres", "redis"},
		Volumes:   []string{"./api:/app"},
		Networks:  []string{"backend"},
	}

	data, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Failed to marshal ComposeService: %v", err)
	}

	var decoded ComposeService
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ComposeService: %v", err)
	}

	if decoded.Name != "api" {
		t.Errorf("Expected name 'api', got '%s'", decoded.Name)
	}
	if len(decoded.DependsOn) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(decoded.DependsOn))
	}
}

func TestCIPipelineStructure(t *testing.T) {
	pipeline := CIPipeline{
		Name:     "CI/CD",
		Triggers: []string{"push", "pull_request"},
		Stages: []CIStage{
			{
				Name:   "build",
				RunsOn: "ubuntu-latest",
				Steps: []CIStep{
					{Name: "Checkout", Uses: "actions/checkout@v4"},
					{Name: "Setup Node", Uses: "actions/setup-node@v4", With: map[string]string{"node-version": "20"}},
					{Name: "Install", Run: "npm ci"},
					{Name: "Build", Run: "npm run build"},
				},
			},
			{
				Name:      "deploy",
				RunsOn:    "ubuntu-latest",
				DependsOn: []string{"build"},
				Steps: []CIStep{
					{Name: "Deploy", Run: "kubectl apply -f k8s/"},
				},
			},
		},
	}

	data, err := json.Marshal(pipeline)
	if err != nil {
		t.Fatalf("Failed to marshal CIPipeline: %v", err)
	}

	var decoded CIPipeline
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CIPipeline: %v", err)
	}

	if len(decoded.Stages) != 2 {
		t.Errorf("Expected 2 stages, got %d", len(decoded.Stages))
	}
	if len(decoded.Stages[0].Steps) != 4 {
		t.Errorf("Expected 4 steps in build stage, got %d", len(decoded.Stages[0].Steps))
	}
}

func TestK8sManifestStructure(t *testing.T) {
	manifest := K8sManifest{
		Kind:      "Deployment",
		Name:      "api",
		Namespace: "production",
		Labels: map[string]string{
			"app":     "api",
			"version": "v1",
		},
		Spec: map[string]interface{}{
			"replicas": 3,
		},
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal K8sManifest: %v", err)
	}

	var decoded K8sManifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal K8sManifest: %v", err)
	}

	if decoded.Kind != "Deployment" {
		t.Errorf("Expected kind 'Deployment', got '%s'", decoded.Kind)
	}
	if decoded.Namespace != "production" {
		t.Errorf("Expected namespace 'production', got '%s'", decoded.Namespace)
	}
}

func TestHelmChartStructure(t *testing.T) {
	chart := HelmChart{
		Name:        "my-app",
		Version:     "1.0.0",
		AppVersion:  "1.0.0",
		Description: "My application Helm chart",
		Values: map[string]interface{}{
			"replicaCount": 2,
			"image": map[string]interface{}{
				"repository": "myapp",
				"tag":        "latest",
			},
		},
	}

	data, err := json.Marshal(chart)
	if err != nil {
		t.Fatalf("Failed to marshal HelmChart: %v", err)
	}

	var decoded HelmChart
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal HelmChart: %v", err)
	}

	if decoded.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", decoded.Version)
	}
}

// ============================================================================
// Input Type Tests
// ============================================================================

func TestDockerfileInputStructure(t *testing.T) {
	input := DockerfileInput{
		ProjectName: "my-api",
		Stack:       "node",
		Framework:   "express",
		HasDatabase: true,
		Port:        3000,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal DockerfileInput: %v", err)
	}

	var decoded DockerfileInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DockerfileInput: %v", err)
	}

	if decoded.Stack != "node" {
		t.Errorf("Expected stack 'node', got '%s'", decoded.Stack)
	}
	if decoded.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", decoded.Port)
	}
}

func TestDockerComposeInputStructure(t *testing.T) {
	input := DockerComposeInput{
		ProjectName: "my-project",
		Services:    []string{"api", "web", "worker"},
		TechStack: &TechStack{
			Backend: BackendStack{
				Framework: "Express",
				Language:  "TypeScript",
				Runtime:   "Node.js",
			},
			Database: DatabaseStack{
				Primary: "PostgreSQL",
				Cache:   "Redis",
				Queue:   "RabbitMQ",
			},
		},
		Environment: "development",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal DockerComposeInput: %v", err)
	}

	var decoded DockerComposeInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal DockerComposeInput: %v", err)
	}

	if len(decoded.Services) != 3 {
		t.Errorf("Expected 3 services, got %d", len(decoded.Services))
	}
}

func TestCIPipelineInputStructure(t *testing.T) {
	input := CIPipelineInput{
		ProjectName:  "my-api",
		Platform:     "github",
		Stack:        "node",
		HasTests:     true,
		HasLinting:   true,
		DeployTarget: "kubernetes",
		Environments: []string{"staging", "production"},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal CIPipelineInput: %v", err)
	}

	var decoded CIPipelineInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CIPipelineInput: %v", err)
	}

	if decoded.Platform != "github" {
		t.Errorf("Expected platform 'github', got '%s'", decoded.Platform)
	}
	if len(decoded.Environments) != 2 {
		t.Errorf("Expected 2 environments, got %d", len(decoded.Environments))
	}
}

func TestK8sManifestsInputStructure(t *testing.T) {
	input := K8sManifestsInput{
		ProjectName: "my-api",
		Namespace:   "production",
		Replicas:    3,
		Environment: "production",
		Ingress: &IngressConfig{
			Host:        "api.example.com",
			TLS:         true,
			CertManager: true,
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal K8sManifestsInput: %v", err)
	}

	var decoded K8sManifestsInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal K8sManifestsInput: %v", err)
	}

	if decoded.Replicas != 3 {
		t.Errorf("Expected 3 replicas, got %d", decoded.Replicas)
	}
	if decoded.Ingress == nil {
		t.Error("Expected Ingress to be set")
	}
	if decoded.Ingress.Host != "api.example.com" {
		t.Errorf("Expected host 'api.example.com', got '%s'", decoded.Ingress.Host)
	}
}

func TestHelmChartInputStructure(t *testing.T) {
	input := HelmChartInput{
		ProjectName:  "my-app",
		TechStack:    &TechStack{Backend: BackendStack{Language: "Go", Framework: "Gin"}},
		Environments: []string{"dev", "staging", "production"},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal HelmChartInput: %v", err)
	}

	var decoded HelmChartInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal HelmChartInput: %v", err)
	}

	if len(decoded.Environments) != 3 {
		t.Errorf("Expected 3 environments, got %d", len(decoded.Environments))
	}
}

// ============================================================================
// Utility Function Tests
// ============================================================================

func TestDetectStackFromTechStack(t *testing.T) {
	tests := []struct {
		name      string
		techStack *TechStack
		expected  string
	}{
		{"nil techstack defaults to node", nil, "node"},
		{"node.js runtime", &TechStack{Backend: BackendStack{Runtime: "Node.js"}}, "node"},
		{"express framework", &TechStack{Backend: BackendStack{Framework: "Express"}}, "node"},
		{"fastify framework", &TechStack{Backend: BackendStack{Framework: "Fastify"}}, "node"},
		{"nestjs framework", &TechStack{Backend: BackendStack{Framework: "NestJS"}}, "node"},
		{"typescript language", &TechStack{Backend: BackendStack{Language: "TypeScript"}}, "node"},
		{"python runtime", &TechStack{Backend: BackendStack{Runtime: "Python"}}, "python"},
		{"fastapi framework", &TechStack{Backend: BackendStack{Framework: "FastAPI"}}, "python"},
		{"django framework", &TechStack{Backend: BackendStack{Framework: "Django"}}, "python"},
		{"flask framework", &TechStack{Backend: BackendStack{Framework: "Flask"}}, "python"},
		{"go language", &TechStack{Backend: BackendStack{Language: "Go"}}, "go"},
		{"gin framework", &TechStack{Backend: BackendStack{Framework: "Gin"}}, "go"},
		{"echo framework", &TechStack{Backend: BackendStack{Framework: "Echo"}}, "go"},
		{"fiber framework", &TechStack{Backend: BackendStack{Framework: "Fiber"}}, "go"},
		{"unknown defaults to node", &TechStack{Backend: BackendStack{Framework: "Spring"}}, "node"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectStackFromTechStack(tt.techStack)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetDefaultPort(t *testing.T) {
	tests := []struct {
		stack    string
		expected int
	}{
		{"node", 3000},
		{"node-express", 3000},
		{"python", 8000},
		{"python-fastapi", 8000},
		{"go", 8080},
		{"go-gin", 8080},
		{"unknown", 3000},
		{"", 3000},
	}

	for _, tt := range tests {
		t.Run(tt.stack, func(t *testing.T) {
			result := getDefaultPort(tt.stack)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestSanitizeK8sName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "MyApp", "myapp"},
		{"spaces to hyphens", "my app", "my-app"},
		{"underscores to hyphens", "my_app", "my-app"},
		{"remove special chars", "my@app#name", "myappname"},
		{"trim hyphens", "-my-app-", "my-app"},
		{"multiple hyphens", "my--app", "my--app"}, // doesn't collapse
		{"numbers allowed", "app123", "app123"},
		{"long name truncated", strings.Repeat("a", 100), strings.Repeat("a", 63)},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeK8sName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetStackCISteps(t *testing.T) {
	tests := []struct {
		stack           string
		expectedContain string
	}{
		{"node", "npm ci"},
		{"node-express", "npm ci"},
		{"python", "pip install"},
		{"python-fastapi", "pytest"},
		{"go", "go mod download"},
		{"go-gin", "golangci-lint"},
		{"unknown", "Install dependencies"},
	}

	for _, tt := range tests {
		t.Run(tt.stack, func(t *testing.T) {
			result := getStackCISteps(tt.stack)
			if !strings.Contains(result, tt.expectedContain) {
				t.Errorf("Expected steps to contain '%s', got '%s'", tt.expectedContain, result)
			}
		})
	}
}

// ============================================================================
// Template Generation Tests
// ============================================================================

func TestGetDockerComposeTemplate(t *testing.T) {
	tests := []struct {
		name     string
		project  string
		services []string
		database string
		checks   []string
	}{
		{
			name:     "with postgres",
			project:  "myapp",
			services: []string{"api"},
			database: "postgresql",
			checks:   []string{"postgres:", "postgres:16-alpine", "postgres_data:"},
		},
		{
			name:     "with mongodb",
			project:  "myapp",
			services: []string{"api"},
			database: "mongodb",
			checks:   []string{"mongo:", "mongo:7", "mongo_data:"},
		},
		{
			name:     "with mysql",
			project:  "myapp",
			services: []string{"api"},
			database: "mysql",
			checks:   []string{"mysql:", "mysql:8", "mysql_data:"},
		},
		{
			name:     "with redis",
			project:  "myapp",
			services: []string{"api"},
			database: "redis",
			checks:   []string{"redis:", "redis:7-alpine"},
		},
		{
			name:     "multiple services",
			project:  "myapp",
			services: []string{"api", "web", "worker"},
			database: "",
			checks:   []string{"api:", "web:", "worker:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDockerComposeTemplate(tt.project, tt.services, tt.database)

			for _, check := range tt.checks {
				if !strings.Contains(result, check) {
					t.Errorf("Expected compose to contain '%s'", check)
				}
			}

			// Verify version
			if !strings.Contains(result, "version: '3.8'") {
				t.Error("Expected compose version 3.8")
			}
		})
	}
}

func TestGenerateDockerfileTemplate(t *testing.T) {
	tests := []struct {
		name   string
		stack  string
		port   int
		checks []string
	}{
		{
			name:  "node dockerfile",
			stack: "node",
			port:  3000,
			checks: []string{
				"FROM node:20-alpine",
				"npm ci",
				"EXPOSE 3000",
				"NODE_ENV=production",
				"USER nodejs",
			},
		},
		{
			name:  "python dockerfile",
			stack: "python",
			port:  8000,
			checks: []string{
				"FROM python:3.11-slim",
				"pip install",
				"EXPOSE 8000",
				"uvicorn",
				"USER appuser",
			},
		},
		{
			name:  "go dockerfile",
			stack: "go",
			port:  8080,
			checks: []string{
				"FROM golang:1.21-alpine",
				"go mod download",
				"CGO_ENABLED=0",
				"EXPOSE 8080",
				"USER appuser",
			},
		},
		{
			name:  "unknown defaults to alpine",
			stack: "unknown",
			port:  9000,
			checks: []string{
				"FROM alpine:3.19",
				"EXPOSE 9000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateDockerfileTemplate(tt.stack, tt.port)

			for _, check := range tt.checks {
				if !strings.Contains(result, check) {
					t.Errorf("Expected Dockerfile to contain '%s'", check)
				}
			}
		})
	}
}

// ============================================================================
// Agent Construction Tests
// ============================================================================

func TestNewDevOpsAgent(t *testing.T) {
	agent := NewDevOpsAgent()

	if agent == nil {
		t.Fatal("NewDevOpsAgent returned nil")
	}

	if agent.llmClient == nil {
		t.Error("DevOpsAgent.llmClient is nil")
	}
}

// ============================================================================
// Default Value Tests
// ============================================================================

func TestEnvironmentDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to development", "", "development"},
		{"development stays", "development", "development"},
		{"staging stays", "staging", "staging"},
		{"production stays", "production", "production"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := tt.input
			if env == "" {
				env = "development"
			}
			if env != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, env)
			}
		})
	}
}

func TestReplicasDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero defaults to 2", 0, 2},
		{"negative defaults to 2", -1, 2},
		{"positive stays", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replicas := tt.input
			if replicas <= 0 {
				replicas = 2
			}
			if replicas != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, replicas)
			}
		})
	}
}

func TestPlatformDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to github", "", "github"},
		{"github stays", "github", "github"},
		{"gitlab stays", "gitlab", "gitlab"},
		{"GITHUB normalizes", "GITHUB", "github"},
		{"GitLab normalizes", "GitLab", "gitlab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform := strings.ToLower(tt.input)
			if platform == "" {
				platform = "github"
			}
			if platform != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, platform)
			}
		})
	}
}

// ============================================================================
// IngressConfig Tests
// ============================================================================

func TestIngressConfigStructure(t *testing.T) {
	config := IngressConfig{
		Host:        "api.example.com",
		TLS:         true,
		CertManager: true,
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal IngressConfig: %v", err)
	}

	var decoded IngressConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal IngressConfig: %v", err)
	}

	if decoded.Host != "api.example.com" {
		t.Errorf("Expected host 'api.example.com', got '%s'", decoded.Host)
	}
	if !decoded.TLS {
		t.Error("Expected TLS to be true")
	}
	if !decoded.CertManager {
		t.Error("Expected CertManager to be true")
	}
}

// ============================================================================
// HealthCheck Tests
// ============================================================================

func TestHealthCheckStructure(t *testing.T) {
	hc := HealthCheck{
		Command:  "curl -f http://localhost:3000/health || exit 1",
		Interval: "30s",
		Timeout:  "10s",
		Retries:  3,
	}

	data, err := json.Marshal(hc)
	if err != nil {
		t.Fatalf("Failed to marshal HealthCheck: %v", err)
	}

	var decoded HealthCheck
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal HealthCheck: %v", err)
	}

	if decoded.Interval != "30s" {
		t.Errorf("Expected interval '30s', got '%s'", decoded.Interval)
	}
	if decoded.Retries != 3 {
		t.Errorf("Expected 3 retries, got %d", decoded.Retries)
	}
}

// ============================================================================
// BuildStage Tests
// ============================================================================

func TestBuildStageStructure(t *testing.T) {
	stage := BuildStage{
		Name: "builder",
		From: "node:20-alpine",
		Commands: []string{
			"WORKDIR /app",
			"COPY package*.json ./",
			"RUN npm ci",
			"COPY . .",
			"RUN npm run build",
		},
	}

	data, err := json.Marshal(stage)
	if err != nil {
		t.Fatalf("Failed to marshal BuildStage: %v", err)
	}

	var decoded BuildStage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal BuildStage: %v", err)
	}

	if decoded.Name != "builder" {
		t.Errorf("Expected name 'builder', got '%s'", decoded.Name)
	}
	if len(decoded.Commands) != 5 {
		t.Errorf("Expected 5 commands, got %d", len(decoded.Commands))
	}
}
