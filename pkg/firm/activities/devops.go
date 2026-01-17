package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// DevOpsAgent handles containerization, CI/CD pipelines, and Kubernetes deployment
// Supports Docker, GitHub Actions, GitLab CI, Kubernetes, and Helm
type DevOpsAgent struct {
	llmClient llm.Client
}

func NewDevOpsAgent() *DevOpsAgent {
	return &DevOpsAgent{
		llmClient: llm.NewClient(),
	}
}

// ============================================================================
// Data Structures
// ============================================================================

// DevOpsCodeBundle represents generated DevOps files
type DevOpsCodeBundle struct {
	Files       map[string]string `json:"files"`        // filename -> content
	BuildCmd    string            `json:"build_cmd"`    // Command to build
	DeployCmd   string            `json:"deploy_cmd"`   // Command to deploy
	TestCmd     string            `json:"test_cmd"`     // Command to run tests
	Environment string            `json:"environment"`  // "development", "staging", "production"
}

// DockerConfig represents Docker-related configuration
type DockerConfig struct {
	BaseImage    string            `json:"base_image"`
	BuildStages  []BuildStage      `json:"build_stages"`
	ExposedPorts []int             `json:"exposed_ports"`
	EnvVars      map[string]string `json:"env_vars"`
	Volumes      []string          `json:"volumes"`
	HealthCheck  *HealthCheck      `json:"health_check,omitempty"`
}

// BuildStage represents a Docker multi-stage build stage
type BuildStage struct {
	Name     string   `json:"name"`
	From     string   `json:"from"`
	Commands []string `json:"commands"`
}

// HealthCheck represents container health check configuration
type HealthCheck struct {
	Command  string `json:"command"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
	Retries  int    `json:"retries"`
}

// ComposeService represents a Docker Compose service
type ComposeService struct {
	Name        string            `json:"name"`
	Image       string            `json:"image,omitempty"`
	Build       string            `json:"build,omitempty"`
	Ports       []string          `json:"ports"`
	Environment map[string]string `json:"environment"`
	DependsOn   []string          `json:"depends_on"`
	Volumes     []string          `json:"volumes"`
	Networks    []string          `json:"networks"`
}

// CIPipeline represents a CI/CD pipeline configuration
type CIPipeline struct {
	Name     string    `json:"name"`
	Triggers []string  `json:"triggers"` // "push", "pull_request", "tag"
	Stages   []CIStage `json:"stages"`
}

// CIStage represents a stage in the CI pipeline
type CIStage struct {
	Name     string   `json:"name"`
	RunsOn   string   `json:"runs_on"` // "ubuntu-latest", "self-hosted"
	Steps    []CIStep `json:"steps"`
	DependsOn []string `json:"depends_on,omitempty"`
}

// CIStep represents a step in a CI stage
type CIStep struct {
	Name    string            `json:"name"`
	Uses    string            `json:"uses,omitempty"`    // GitHub Action
	Run     string            `json:"run,omitempty"`     // Shell command
	With    map[string]string `json:"with,omitempty"`    // Action inputs
	Env     map[string]string `json:"env,omitempty"`
}

// K8sManifest represents Kubernetes manifest configuration
type K8sManifest struct {
	Kind       string            `json:"kind"` // "Deployment", "Service", "Ingress", etc.
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Labels     map[string]string `json:"labels"`
	Spec       interface{}       `json:"spec"`
}

// HelmChart represents a Helm chart structure
type HelmChart struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	AppVersion  string                 `json:"app_version"`
	Description string                 `json:"description"`
	Values      map[string]interface{} `json:"values"`
}

// ============================================================================
// Input Types
// ============================================================================

// DockerfileInput for generating Dockerfiles
type DockerfileInput struct {
	ProjectName string     `json:"project_name"`
	Stack       string     `json:"stack"`       // "node", "python", "go"
	Framework   string     `json:"framework"`   // "express", "fastapi", "gin"
	TechStack   *TechStack `json:"tech_stack,omitempty"`
	HasDatabase bool       `json:"has_database"`
	Port        int        `json:"port"`
}

// DockerComposeInput for generating docker-compose.yml
type DockerComposeInput struct {
	ProjectName string     `json:"project_name"`
	Services    []string   `json:"services"`   // ["api", "web", "worker"]
	TechStack   *TechStack `json:"tech_stack"`
	Environment string     `json:"environment"` // "development", "production"
}

// CIPipelineInput for generating CI/CD pipelines
type CIPipelineInput struct {
	ProjectName  string   `json:"project_name"`
	Platform     string   `json:"platform"`     // "github", "gitlab"
	Stack        string   `json:"stack"`        // "node", "python", "go"
	HasTests     bool     `json:"has_tests"`
	HasLinting   bool     `json:"has_linting"`
	DeployTarget string   `json:"deploy_target"` // "kubernetes", "ecs", "cloud-run"
	Environments []string `json:"environments"`  // ["staging", "production"]
}

// K8sManifestsInput for generating Kubernetes manifests
type K8sManifestsInput struct {
	ProjectName string     `json:"project_name"`
	Namespace   string     `json:"namespace"`
	TechStack   *TechStack `json:"tech_stack"`
	Replicas    int        `json:"replicas"`
	Environment string     `json:"environment"`
	Ingress     *IngressConfig `json:"ingress,omitempty"`
}

// IngressConfig for Kubernetes ingress
type IngressConfig struct {
	Host        string `json:"host"`
	TLS         bool   `json:"tls"`
	CertManager bool   `json:"cert_manager"`
}

// HelmChartInput for generating Helm charts
type HelmChartInput struct {
	ProjectName string     `json:"project_name"`
	TechStack   *TechStack `json:"tech_stack"`
	Environments []string  `json:"environments"`
}

// ============================================================================
// Activity Methods
// ============================================================================

// GenerateDockerfile creates a Dockerfile for the application
func (a *DevOpsAgent) GenerateDockerfile(ctx context.Context, input map[string]interface{}) (*DevOpsCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req DockerfileInput
	json.Unmarshal(inputBytes, &req)

	stack := strings.ToLower(req.Stack)
	if stack == "" {
		stack = detectStackFromTechStack(req.TechStack)
	}

	port := req.Port
	if port == 0 {
		port = getDefaultPort(stack)
	}

	sysPrompt := fmt.Sprintf(`You are a DevOps expert creating production-ready Dockerfiles.
Generate a multi-stage Dockerfile optimized for %s applications.

Output ONLY valid JSON with this structure:
{
  "files": {
    "Dockerfile": "dockerfile content",
    ".dockerignore": "dockerignore content"
  },
  "build_cmd": "docker build -t app .",
  "deploy_cmd": "docker run -p %d:%d app",
  "test_cmd": "docker run app npm test"
}

Dockerfile Best Practices:
1. Use multi-stage builds to minimize image size
2. Use specific version tags, not 'latest'
3. Run as non-root user
4. Use COPY instead of ADD
5. Combine RUN commands to reduce layers
6. Order instructions from least to most frequently changed
7. Include health checks
8. Set proper working directory
9. Use .dockerignore to exclude unnecessary files
10. Handle signals properly (use exec form for CMD)

Stack-Specific Guidelines for %s:
`, stack, port, port, stack)

	// Add stack-specific guidelines
	switch stack {
	case "node", "node-express", "node-fastify", "node-nestjs":
		sysPrompt += `
- Use node:20-alpine as base
- Use npm ci --only=production for production dependencies
- Copy package*.json first for layer caching
- Use node user, not root
- Set NODE_ENV=production`
	case "python", "python-fastapi", "python-django", "python-flask":
		sysPrompt += `
- Use python:3.11-slim as base
- Create virtual environment
- Use pip install --no-cache-dir
- Copy requirements.txt first for layer caching
- Use non-root user`
	case "go", "go-gin", "go-echo", "go-fiber":
		sysPrompt += `
- Use golang:1.21-alpine for build stage
- Use scratch or alpine for final stage
- Use CGO_ENABLED=0 for static binary
- Copy go.mod and go.sum first
- Use go build -ldflags="-s -w" for smaller binary`
	}

	prompt := fmt.Sprintf(`Project: %s
Stack: %s
Framework: %s
Port: %d
Has Database: %v

Generate a production-ready Dockerfile with .dockerignore.`,
		req.ProjectName, stack, req.Framework, port, req.HasDatabase)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Dockerfile generation failed: %w", err)
	}

	var bundle DevOpsCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Dockerfile output: %w", err)
	}

	return &bundle, nil
}

// GenerateDockerCompose creates a docker-compose.yml for the full stack
func (a *DevOpsAgent) GenerateDockerCompose(ctx context.Context, input map[string]interface{}) (*DevOpsCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req DockerComposeInput
	json.Unmarshal(inputBytes, &req)

	env := req.Environment
	if env == "" {
		env = "development"
	}

	sysPrompt := fmt.Sprintf(`You are a DevOps expert creating Docker Compose configurations.
Generate a docker-compose.yml for %s environment.

Output ONLY valid JSON with this structure:
{
  "files": {
    "docker-compose.yml": "compose content",
    "docker-compose.override.yml": "override for development",
    ".env.example": "environment variables template"
  },
  "build_cmd": "docker-compose build",
  "deploy_cmd": "docker-compose up -d",
  "test_cmd": "docker-compose run --rm api npm test",
  "environment": "%s"
}

Docker Compose Best Practices:
1. Use version '3.8' or later
2. Define networks explicitly
3. Use named volumes for persistence
4. Set restart policies
5. Define healthchecks
6. Use environment files (.env)
7. Set resource limits for production
8. Use depends_on with condition: service_healthy
9. Separate override files for dev/prod
10. Include logging configuration`, env, env)

	techStackHint := ""
	if req.TechStack != nil {
		if req.TechStack.Database.Primary != "" {
			techStackHint += fmt.Sprintf("\nDatabase: %s", req.TechStack.Database.Primary)
		}
		if req.TechStack.Database.Cache != "" {
			techStackHint += fmt.Sprintf("\nCache: %s", req.TechStack.Database.Cache)
		}
		if req.TechStack.Database.Queue != "" {
			techStackHint += fmt.Sprintf("\nMessage Queue: %s", req.TechStack.Database.Queue)
		}
	}

	prompt := fmt.Sprintf(`Project: %s
Services: %s
Environment: %s%s

Generate docker-compose.yml with all necessary services, networks, and volumes.`,
		req.ProjectName, strings.Join(req.Services, ", "), env, techStackHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Docker Compose generation failed: %w", err)
	}

	var bundle DevOpsCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Docker Compose output: %w", err)
	}

	bundle.Environment = env
	return &bundle, nil
}

// GenerateCIPipeline creates CI/CD pipeline configuration
func (a *DevOpsAgent) GenerateCIPipeline(ctx context.Context, input map[string]interface{}) (*DevOpsCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req CIPipelineInput
	json.Unmarshal(inputBytes, &req)

	platform := strings.ToLower(req.Platform)
	if platform == "" {
		platform = "github"
	}

	var sysPrompt string
	var pipelineFile string

	switch platform {
	case "github":
		pipelineFile = ".github/workflows/ci.yml"
		sysPrompt = a.getGitHubActionsPrompt(req)
	case "gitlab":
		pipelineFile = ".gitlab-ci.yml"
		sysPrompt = a.getGitLabCIPrompt(req)
	default:
		pipelineFile = ".github/workflows/ci.yml"
		sysPrompt = a.getGitHubActionsPrompt(req)
	}

	prompt := fmt.Sprintf(`Project: %s
Platform: %s
Stack: %s
Has Tests: %v
Has Linting: %v
Deploy Target: %s
Environments: %s

Generate a complete CI/CD pipeline.`,
		req.ProjectName, platform, req.Stack, req.HasTests, req.HasLinting, req.DeployTarget, strings.Join(req.Environments, ", "))

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("CI pipeline generation failed: %w", err)
	}

	var bundle DevOpsCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse CI pipeline output: %w", err)
	}

	// Ensure the pipeline file is in the correct path
	if _, exists := bundle.Files[pipelineFile]; !exists {
		// Move content to correct path if needed
		for k, v := range bundle.Files {
			if strings.Contains(k, "ci") || strings.Contains(k, "workflow") {
				bundle.Files[pipelineFile] = v
				if k != pipelineFile {
					delete(bundle.Files, k)
				}
				break
			}
		}
	}

	return &bundle, nil
}

// GenerateKubernetesManifests creates Kubernetes deployment manifests
func (a *DevOpsAgent) GenerateKubernetesManifests(ctx context.Context, input map[string]interface{}) (*DevOpsCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req K8sManifestsInput
	json.Unmarshal(inputBytes, &req)

	namespace := req.Namespace
	if namespace == "" {
		namespace = sanitizeK8sName(req.ProjectName)
	}

	replicas := req.Replicas
	if replicas <= 0 {
		replicas = 2
	}

	env := req.Environment
	if env == "" {
		env = "production"
	}

	sysPrompt := `You are a Kubernetes expert creating production-ready manifests.
Generate Kubernetes YAML manifests for deployment.

Output ONLY valid JSON with this structure:
{
  "files": {
    "k8s/namespace.yaml": "namespace manifest",
    "k8s/deployment.yaml": "deployment manifest",
    "k8s/service.yaml": "service manifest",
    "k8s/ingress.yaml": "ingress manifest",
    "k8s/configmap.yaml": "configmap manifest",
    "k8s/secret.yaml": "secret manifest (with placeholders)",
    "k8s/hpa.yaml": "horizontal pod autoscaler"
  },
  "deploy_cmd": "kubectl apply -f k8s/",
  "test_cmd": "kubectl get pods -n namespace"
}

Kubernetes Best Practices:
1. Always specify resource requests and limits
2. Use liveness and readiness probes
3. Set pod disruption budgets
4. Use ConfigMaps for configuration
5. Use Secrets for sensitive data (base64 placeholder)
6. Set security contexts (non-root, read-only filesystem)
7. Use proper labels and selectors
8. Configure horizontal pod autoscaling
9. Set proper network policies
10. Use rolling update strategy`

	ingressHint := ""
	if req.Ingress != nil {
		ingressHint = fmt.Sprintf("\nIngress Host: %s, TLS: %v, CertManager: %v",
			req.Ingress.Host, req.Ingress.TLS, req.Ingress.CertManager)
	}

	prompt := fmt.Sprintf(`Project: %s
Namespace: %s
Replicas: %d
Environment: %s%s

Generate Kubernetes manifests for production deployment.`,
		req.ProjectName, namespace, replicas, env, ingressHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Kubernetes manifests generation failed: %w", err)
	}

	var bundle DevOpsCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Kubernetes manifests: %w", err)
	}

	bundle.Environment = env
	return &bundle, nil
}

// GenerateHelmChart creates a Helm chart for the application
func (a *DevOpsAgent) GenerateHelmChart(ctx context.Context, input map[string]interface{}) (*DevOpsCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req HelmChartInput
	json.Unmarshal(inputBytes, &req)

	chartName := sanitizeK8sName(req.ProjectName)

	sysPrompt := `You are a Helm expert creating production-ready charts.
Generate a complete Helm chart structure.

Output ONLY valid JSON with this structure:
{
  "files": {
    "helm/Chart.yaml": "chart metadata",
    "helm/values.yaml": "default values",
    "helm/values-staging.yaml": "staging overrides",
    "helm/values-production.yaml": "production overrides",
    "helm/templates/deployment.yaml": "deployment template",
    "helm/templates/service.yaml": "service template",
    "helm/templates/ingress.yaml": "ingress template",
    "helm/templates/configmap.yaml": "configmap template",
    "helm/templates/secret.yaml": "secret template",
    "helm/templates/hpa.yaml": "hpa template",
    "helm/templates/_helpers.tpl": "template helpers",
    "helm/templates/NOTES.txt": "installation notes"
  },
  "deploy_cmd": "helm upgrade --install release-name ./helm -f helm/values-production.yaml",
  "test_cmd": "helm lint ./helm && helm template ./helm"
}

Helm Best Practices:
1. Use semantic versioning for chart version
2. Define all configurable values in values.yaml
3. Use _helpers.tpl for common labels and selectors
4. Support multiple environments via values files
5. Include NOTES.txt with helpful post-install info
6. Use proper YAML anchors for DRY templates
7. Validate with helm lint
8. Include resource limits as configurable values
9. Support image tag overrides
10. Include proper documentation`

	envHint := ""
	if len(req.Environments) > 0 {
		envHint = fmt.Sprintf("\nEnvironments: %s", strings.Join(req.Environments, ", "))
	}

	prompt := fmt.Sprintf(`Project: %s
Chart Name: %s%s

Generate a complete Helm chart with environment-specific values.`,
		req.ProjectName, chartName, envHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Helm chart generation failed: %w", err)
	}

	var bundle DevOpsCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Helm chart: %w", err)
	}

	return &bundle, nil
}

// RefineDevOps refines DevOps configurations based on feedback
func (a *DevOpsAgent) RefineDevOps(ctx context.Context, input map[string]interface{}) (*DevOpsCodeBundle, error) {
	currentFiles, _ := input["current_files"].(map[string]interface{})
	feedback, _ := input["feedback"].(string)
	fileType, _ := input["file_type"].(string) // "dockerfile", "compose", "ci", "k8s", "helm"

	if fileType == "" {
		fileType = "dockerfile"
	}

	sysPrompt := fmt.Sprintf(`You are a DevOps expert refining %s configurations.
Apply the feedback to improve the configuration.

Output ONLY valid JSON with this structure:
{
  "files": {
    "path/to/file": "updated content"
  },
  "build_cmd": "build command",
  "deploy_cmd": "deploy command",
  "test_cmd": "test command"
}

Apply the feedback while maintaining:
1. Production readiness
2. Security best practices
3. Performance optimization
4. Maintainability`, fileType)

	filesStr, _ := json.MarshalIndent(currentFiles, "", "  ")
	prompt := fmt.Sprintf(`CURRENT FILES:
%s

FEEDBACK:
%s

Apply the feedback and return improved configuration.`, string(filesStr), feedback)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("DevOps refinement failed: %w", err)
	}

	var bundle DevOpsCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse refined DevOps output: %w", err)
	}

	return &bundle, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (a *DevOpsAgent) getGitHubActionsPrompt(req CIPipelineInput) string {
	stackSteps := getStackCISteps(req.Stack)

	return fmt.Sprintf(`You are a GitHub Actions expert creating CI/CD workflows.
Generate a production-ready GitHub Actions workflow.

Output ONLY valid JSON with this structure:
{
  "files": {
    ".github/workflows/ci.yml": "main ci workflow",
    ".github/workflows/deploy.yml": "deployment workflow"
  },
  "build_cmd": "act -j build",
  "deploy_cmd": "gh workflow run deploy.yml",
  "test_cmd": "act -j test"
}

GitHub Actions Best Practices:
1. Use specific action versions (@v3, not @latest)
2. Cache dependencies (actions/cache)
3. Use matrix builds for multiple versions
4. Set proper permissions
5. Use environments for deployments
6. Use OIDC for cloud authentication
7. Run security scanning (CodeQL, Snyk)
8. Use concurrency to cancel outdated runs
9. Set timeout-minutes for jobs
10. Use reusable workflows for DRY

Stack-specific steps for %s:
%s`, req.Stack, stackSteps)
}

func (a *DevOpsAgent) getGitLabCIPrompt(req CIPipelineInput) string {
	stackSteps := getStackCISteps(req.Stack)

	return fmt.Sprintf(`You are a GitLab CI expert creating CI/CD pipelines.
Generate a production-ready GitLab CI configuration.

Output ONLY valid JSON with this structure:
{
  "files": {
    ".gitlab-ci.yml": "gitlab ci configuration"
  },
  "build_cmd": "gitlab-runner exec docker build",
  "deploy_cmd": "gitlab-ci-multi-runner exec",
  "test_cmd": "gitlab-runner exec docker test"
}

GitLab CI Best Practices:
1. Use stages for logical grouping
2. Use extends for DRY configuration
3. Use rules instead of only/except
4. Cache dependencies properly
5. Use artifacts for passing data between jobs
6. Use environments for deployments
7. Set proper resource limits
8. Use parallel for matrix builds
9. Use needs for DAG pipelines
10. Include security scanning templates

Stack-specific steps for %s:
%s`, req.Stack, stackSteps)
}

// ============================================================================
// Utility Functions
// ============================================================================

func detectStackFromTechStack(ts *TechStack) string {
	if ts == nil {
		return "node"
	}

	// Check both framework and language fields
	framework := strings.ToLower(ts.Backend.Framework)
	language := strings.ToLower(ts.Backend.Language)
	runtime := strings.ToLower(ts.Backend.Runtime)

	// Check framework first
	switch {
	case strings.Contains(framework, "express") || strings.Contains(framework, "fastify") ||
		strings.Contains(framework, "nest") || strings.Contains(framework, "koa"):
		return "node"
	case strings.Contains(framework, "fastapi") || strings.Contains(framework, "django") ||
		strings.Contains(framework, "flask"):
		return "python"
	case strings.Contains(framework, "gin") || strings.Contains(framework, "echo") ||
		strings.Contains(framework, "fiber") || strings.Contains(framework, "chi"):
		return "go"
	}

	// Fall back to language/runtime
	switch {
	case strings.Contains(language, "typescript") || strings.Contains(language, "javascript") ||
		strings.Contains(runtime, "node"):
		return "node"
	case strings.Contains(language, "python") || strings.Contains(runtime, "python"):
		return "python"
	case strings.Contains(language, "go") || strings.Contains(runtime, "go"):
		return "go"
	default:
		return "node"
	}
}

func getDefaultPort(stack string) int {
	stack = strings.ToLower(stack)
	switch {
	case strings.Contains(stack, "node"):
		return 3000
	case strings.Contains(stack, "python"):
		return 8000
	case strings.Contains(stack, "go"):
		return 8080
	default:
		return 3000
	}
}

func getStackCISteps(stack string) string {
	stack = strings.ToLower(stack)
	switch {
	case strings.Contains(stack, "node"):
		return `- Setup Node.js (actions/setup-node@v4)
- Install dependencies (npm ci)
- Run linting (npm run lint)
- Run tests (npm test)
- Build (npm run build)
- Docker build and push`
	case strings.Contains(stack, "python"):
		return `- Setup Python (actions/setup-python@v5)
- Install dependencies (pip install -r requirements.txt)
- Run linting (ruff, black --check)
- Run tests (pytest)
- Docker build and push`
	case strings.Contains(stack, "go"):
		return `- Setup Go (actions/setup-go@v5)
- Download dependencies (go mod download)
- Run linting (golangci-lint)
- Run tests (go test ./...)
- Build (go build ./...)
- Docker build and push`
	default:
		return `- Install dependencies
- Run linting
- Run tests
- Build
- Docker build and push`
	}
}

func sanitizeK8sName(name string) string {
	// Kubernetes names must be lowercase, alphanumeric, and can include hyphens
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Remove invalid characters
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// Ensure it doesn't start or end with a hyphen
	s := result.String()
	s = strings.Trim(s, "-")

	// Truncate to 63 characters (K8s limit)
	if len(s) > 63 {
		s = s[:63]
	}

	return s
}

// GetDockerComposeTemplate returns a basic docker-compose template
func GetDockerComposeTemplate(projectName string, services []string, database string) string {
	var sb strings.Builder

	sb.WriteString("version: '3.8'\n\n")
	sb.WriteString("services:\n")

	for _, svc := range services {
		sb.WriteString(fmt.Sprintf("  %s:\n", svc))
		sb.WriteString(fmt.Sprintf("    build: ./%s\n", svc))
		sb.WriteString("    ports:\n")
		sb.WriteString("      - \"3000:3000\"\n")
		sb.WriteString("    environment:\n")
		sb.WriteString("      - NODE_ENV=development\n")
		if database != "" {
			sb.WriteString(fmt.Sprintf("    depends_on:\n      - %s\n", database))
		}
		sb.WriteString("\n")
	}

	if database != "" {
		switch strings.ToLower(database) {
		case "postgresql", "postgres":
			sb.WriteString("  postgres:\n")
			sb.WriteString("    image: postgres:16-alpine\n")
			sb.WriteString("    environment:\n")
			sb.WriteString("      POSTGRES_USER: ${DB_USER:-postgres}\n")
			sb.WriteString("      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}\n")
			sb.WriteString("      POSTGRES_DB: ${DB_NAME:-app}\n")
			sb.WriteString("    volumes:\n")
			sb.WriteString("      - postgres_data:/var/lib/postgresql/data\n")
			sb.WriteString("    ports:\n")
			sb.WriteString("      - \"5432:5432\"\n")
		case "mongodb", "mongo":
			sb.WriteString("  mongo:\n")
			sb.WriteString("    image: mongo:7\n")
			sb.WriteString("    environment:\n")
			sb.WriteString("      MONGO_INITDB_ROOT_USERNAME: ${MONGO_USER:-admin}\n")
			sb.WriteString("      MONGO_INITDB_ROOT_PASSWORD: ${MONGO_PASSWORD:-admin}\n")
			sb.WriteString("    volumes:\n")
			sb.WriteString("      - mongo_data:/data/db\n")
			sb.WriteString("    ports:\n")
			sb.WriteString("      - \"27017:27017\"\n")
		case "mysql":
			sb.WriteString("  mysql:\n")
			sb.WriteString("    image: mysql:8\n")
			sb.WriteString("    environment:\n")
			sb.WriteString("      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD:-root}\n")
			sb.WriteString("      MYSQL_DATABASE: ${DB_NAME:-app}\n")
			sb.WriteString("      MYSQL_USER: ${DB_USER:-app}\n")
			sb.WriteString("      MYSQL_PASSWORD: ${DB_PASSWORD:-app}\n")
			sb.WriteString("    volumes:\n")
			sb.WriteString("      - mysql_data:/var/lib/mysql\n")
			sb.WriteString("    ports:\n")
			sb.WriteString("      - \"3306:3306\"\n")
		case "redis":
			sb.WriteString("  redis:\n")
			sb.WriteString("    image: redis:7-alpine\n")
			sb.WriteString("    ports:\n")
			sb.WriteString("      - \"6379:6379\"\n")
		}
	}

	sb.WriteString("\nvolumes:\n")
	if strings.Contains(strings.ToLower(database), "postgres") {
		sb.WriteString("  postgres_data:\n")
	}
	if strings.Contains(strings.ToLower(database), "mongo") {
		sb.WriteString("  mongo_data:\n")
	}
	if strings.Contains(strings.ToLower(database), "mysql") {
		sb.WriteString("  mysql_data:\n")
	}

	return sb.String()
}

// GenerateDockerfileTemplate returns a basic Dockerfile template
func GenerateDockerfileTemplate(stack string, port int) string {
	stack = strings.ToLower(stack)

	switch {
	case strings.Contains(stack, "node"):
		return fmt.Sprintf(`# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Production stage
FROM node:20-alpine
WORKDIR /app
RUN addgroup -g 1001 -S nodejs && adduser -S nodejs -u 1001
COPY --from=builder --chown=nodejs:nodejs /app/dist ./dist
COPY --from=builder --chown=nodejs:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=nodejs:nodejs /app/package*.json ./
USER nodejs
EXPOSE %d
ENV NODE_ENV=production
CMD ["node", "dist/index.js"]
`, port)

	case strings.Contains(stack, "python"):
		return fmt.Sprintf(`# Build stage
FROM python:3.11-slim AS builder
WORKDIR /app
RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Production stage
FROM python:3.11-slim
WORKDIR /app
RUN useradd -m -u 1001 appuser
COPY --from=builder /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"
COPY --chown=appuser:appuser . .
USER appuser
EXPOSE %d
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "%d"]
`, port, port)

	case strings.Contains(stack, "go"):
		return fmt.Sprintf(`# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/server

# Production stage
FROM alpine:3.19
RUN adduser -D -u 1001 appuser
WORKDIR /app
COPY --from=builder /app/main .
USER appuser
EXPOSE %d
CMD ["./main"]
`, port)

	default:
		return fmt.Sprintf(`FROM alpine:3.19
WORKDIR /app
COPY . .
EXPOSE %d
CMD ["./start.sh"]
`, port)
	}
}
