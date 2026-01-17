package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qlfactory/sovereign-firm/pkg/sovereign/llm"
)

// InfraAgent handles Infrastructure as Code generation for cloud deployments
// Supports Terraform (AWS, Azure, GCP), CloudFormation, and Pulumi
type InfraAgent struct {
	llmClient llm.Client
}

func NewInfraAgent() *InfraAgent {
	return &InfraAgent{
		llmClient: llm.NewClient(),
	}
}

// ============================================================================
// Data Structures
// ============================================================================

// InfraCodeBundle represents generated infrastructure code
type InfraCodeBundle struct {
	Provider    string            `json:"provider"`     // "aws", "azure", "gcp"
	Tool        string            `json:"tool"`         // "terraform", "cloudformation", "pulumi"
	Files       map[string]string `json:"files"`        // filename -> content
	InitCmd     string            `json:"init_cmd"`     // Command to initialize
	PlanCmd     string            `json:"plan_cmd"`     // Command to plan changes
	ApplyCmd    string            `json:"apply_cmd"`    // Command to apply changes
	DestroyCmd  string            `json:"destroy_cmd"`  // Command to destroy resources
	Environment string            `json:"environment"`  // "development", "staging", "production"
}

// InfraSpec represents infrastructure requirements
type InfraSpec struct {
	ProjectName  string            `json:"project_name"`
	Environment  string            `json:"environment"`
	Region       string            `json:"region"`
	Components   []InfraComponent  `json:"components"`
	Networking   *NetworkingSpec   `json:"networking,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// InfraComponent represents a single infrastructure component
type InfraComponent struct {
	Type     string                 `json:"type"`     // "compute", "database", "storage", "cache", "queue", "cdn"
	Name     string                 `json:"name"`
	Service  string                 `json:"service"`  // e.g., "ecs", "rds", "s3"
	Config   map[string]interface{} `json:"config"`
	DependsOn []string              `json:"depends_on,omitempty"`
}

// NetworkingSpec represents networking configuration
type NetworkingSpec struct {
	VPCCidr         string   `json:"vpc_cidr"`
	PublicSubnets   []string `json:"public_subnets"`
	PrivateSubnets  []string `json:"private_subnets"`
	EnableNAT       bool     `json:"enable_nat"`
	EnableVPN       bool     `json:"enable_vpn"`
}

// CostEstimate represents cloud cost estimation
type CostEstimate struct {
	Provider       string              `json:"provider"`
	Environment    string              `json:"environment"`
	MonthlyCost    float64             `json:"monthly_cost"`
	Currency       string              `json:"currency"`
	Breakdown      []CostLineItem      `json:"breakdown"`
	Assumptions    []string            `json:"assumptions"`
	Recommendations []CostRecommendation `json:"recommendations,omitempty"`
}

// CostLineItem represents a single cost item
type CostLineItem struct {
	Service     string  `json:"service"`
	Resource    string  `json:"resource"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	MonthlyCost float64 `json:"monthly_cost"`
	Notes       string  `json:"notes,omitempty"`
}

// CostRecommendation represents a cost optimization suggestion
type CostRecommendation struct {
	Type           string  `json:"type"`     // "rightsizing", "reserved", "spot", "architecture"
	Description    string  `json:"description"`
	PotentialSaving float64 `json:"potential_saving"`
	Effort         string  `json:"effort"`   // "low", "medium", "high"
}

// ============================================================================
// Input Types
// ============================================================================

// TerraformInput for generating Terraform code
type TerraformInput struct {
	ProjectName string     `json:"project_name"`
	Provider    string     `json:"provider"`    // "aws", "azure", "gcp"
	Region      string     `json:"region"`
	Environment string     `json:"environment"` // "development", "staging", "production"
	TechStack   *TechStack `json:"tech_stack,omitempty"`
	Components  []string   `json:"components"`  // ["vpc", "ecs", "rds", "s3", "cloudfront"]
	Modular     bool       `json:"modular"`     // Generate modular structure
}

// CloudFormationInput for generating CloudFormation templates
type CloudFormationInput struct {
	ProjectName string     `json:"project_name"`
	Region      string     `json:"region"`
	Environment string     `json:"environment"`
	TechStack   *TechStack `json:"tech_stack,omitempty"`
	Components  []string   `json:"components"`
	NestedStacks bool      `json:"nested_stacks"` // Use nested stacks
}

// PulumiInput for generating Pulumi code
type PulumiInput struct {
	ProjectName string     `json:"project_name"`
	Provider    string     `json:"provider"`
	Region      string     `json:"region"`
	Environment string     `json:"environment"`
	Language    string     `json:"language"`    // "typescript", "python", "go"
	TechStack   *TechStack `json:"tech_stack,omitempty"`
	Components  []string   `json:"components"`
}

// CostEstimateInput for estimating cloud costs
type CostEstimateInput struct {
	Provider    string        `json:"provider"`
	Region      string        `json:"region"`
	Environment string        `json:"environment"`
	InfraSpec   *InfraSpec    `json:"infra_spec,omitempty"`
	TechStack   *TechStack    `json:"tech_stack,omitempty"`
	Components  []string      `json:"components,omitempty"`
}

// ============================================================================
// Activity Methods
// ============================================================================

// GenerateTerraform creates Terraform code for cloud infrastructure
func (a *InfraAgent) GenerateTerraform(ctx context.Context, input map[string]interface{}) (*InfraCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req TerraformInput
	json.Unmarshal(inputBytes, &req)

	provider := strings.ToLower(req.Provider)
	if provider == "" {
		provider = "aws"
	}

	region := req.Region
	if region == "" {
		region = getDefaultRegion(provider)
	}

	env := req.Environment
	if env == "" {
		env = "development"
	}

	sysPrompt := a.getTerraformPrompt(provider, req.Modular)

	componentsHint := ""
	if len(req.Components) > 0 {
		componentsHint = fmt.Sprintf("\nComponents to provision: %s", strings.Join(req.Components, ", "))
	}

	techStackHint := ""
	if req.TechStack != nil {
		techStackHint = describeTechStackForInfra(req.TechStack)
	}

	prompt := fmt.Sprintf(`Project: %s
Provider: %s
Region: %s
Environment: %s%s%s

Generate production-ready Terraform code.`,
		req.ProjectName, provider, region, env, componentsHint, techStackHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Terraform generation failed: %w", err)
	}

	var bundle InfraCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Terraform output: %w", err)
	}

	bundle.Provider = provider
	bundle.Tool = "terraform"
	bundle.Environment = env
	return &bundle, nil
}

// GenerateCloudFormation creates AWS CloudFormation templates
func (a *InfraAgent) GenerateCloudFormation(ctx context.Context, input map[string]interface{}) (*InfraCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req CloudFormationInput
	json.Unmarshal(inputBytes, &req)

	region := req.Region
	if region == "" {
		region = "us-east-1"
	}

	env := req.Environment
	if env == "" {
		env = "development"
	}

	sysPrompt := a.getCloudFormationPrompt(req.NestedStacks)

	componentsHint := ""
	if len(req.Components) > 0 {
		componentsHint = fmt.Sprintf("\nComponents to provision: %s", strings.Join(req.Components, ", "))
	}

	prompt := fmt.Sprintf(`Project: %s
Region: %s
Environment: %s
Use Nested Stacks: %v%s

Generate production-ready CloudFormation templates.`,
		req.ProjectName, region, env, req.NestedStacks, componentsHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("CloudFormation generation failed: %w", err)
	}

	var bundle InfraCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse CloudFormation output: %w", err)
	}

	bundle.Provider = "aws"
	bundle.Tool = "cloudformation"
	bundle.Environment = env
	return &bundle, nil
}

// GeneratePulumi creates Pulumi infrastructure code
func (a *InfraAgent) GeneratePulumi(ctx context.Context, input map[string]interface{}) (*InfraCodeBundle, error) {
	inputBytes, _ := json.Marshal(input)
	var req PulumiInput
	json.Unmarshal(inputBytes, &req)

	provider := strings.ToLower(req.Provider)
	if provider == "" {
		provider = "aws"
	}

	language := strings.ToLower(req.Language)
	if language == "" {
		language = "typescript"
	}

	region := req.Region
	if region == "" {
		region = getDefaultRegion(provider)
	}

	env := req.Environment
	if env == "" {
		env = "development"
	}

	sysPrompt := a.getPulumiPrompt(provider, language)

	componentsHint := ""
	if len(req.Components) > 0 {
		componentsHint = fmt.Sprintf("\nComponents to provision: %s", strings.Join(req.Components, ", "))
	}

	prompt := fmt.Sprintf(`Project: %s
Provider: %s
Language: %s
Region: %s
Environment: %s%s

Generate production-ready Pulumi code.`,
		req.ProjectName, provider, language, region, env, componentsHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("Pulumi generation failed: %w", err)
	}

	var bundle InfraCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse Pulumi output: %w", err)
	}

	bundle.Provider = provider
	bundle.Tool = "pulumi"
	bundle.Environment = env
	return &bundle, nil
}

// EstimateCost estimates cloud infrastructure costs
func (a *InfraAgent) EstimateCost(ctx context.Context, input map[string]interface{}) (*CostEstimate, error) {
	inputBytes, _ := json.Marshal(input)
	var req CostEstimateInput
	json.Unmarshal(inputBytes, &req)

	provider := strings.ToLower(req.Provider)
	if provider == "" {
		provider = "aws"
	}

	region := req.Region
	if region == "" {
		region = getDefaultRegion(provider)
	}

	env := req.Environment
	if env == "" {
		env = "production"
	}

	sysPrompt := fmt.Sprintf(`You are a Cloud Cost Analyst estimating infrastructure costs for %s.
Provide realistic cost estimates based on current %s pricing.

Output ONLY valid JSON with this structure:
{
  "provider": "%s",
  "environment": "%s",
  "monthly_cost": 1234.56,
  "currency": "USD",
  "breakdown": [
    {
      "service": "EC2",
      "resource": "t3.medium (2 instances)",
      "quantity": 2,
      "unit_price": 0.0416,
      "monthly_cost": 60.0,
      "notes": "On-demand pricing, 730 hours/month"
    }
  ],
  "assumptions": [
    "24/7 operation",
    "On-demand pricing (no reserved instances)",
    "Standard data transfer estimates"
  ],
  "recommendations": [
    {
      "type": "reserved",
      "description": "Use 1-year reserved instances for EC2",
      "potential_saving": 240.0,
      "effort": "low"
    }
  ]
}

Cost Estimation Guidelines:
1. Use current public pricing for %s %s region
2. Include all related costs (compute, storage, network, data transfer)
3. Estimate based on environment (%s workload patterns)
4. Include realistic data transfer costs
5. Consider multi-AZ costs if production
6. Include support costs if enterprise
7. Add CloudWatch/monitoring costs
8. Include NAT Gateway costs if VPC
9. Provide cost-saving recommendations
10. Note all assumptions clearly`, provider, provider, provider, env, provider, region, env)

	infraHint := ""
	if req.InfraSpec != nil {
		specBytes, _ := json.MarshalIndent(req.InfraSpec, "", "  ")
		infraHint = fmt.Sprintf("\n\nInfrastructure Specification:\n%s", string(specBytes))
	}

	componentsHint := ""
	if len(req.Components) > 0 {
		componentsHint = fmt.Sprintf("\n\nComponents: %s", strings.Join(req.Components, ", "))
	}

	techStackHint := ""
	if req.TechStack != nil {
		techStackHint = describeTechStackForInfra(req.TechStack)
	}

	prompt := fmt.Sprintf(`Provider: %s
Region: %s
Environment: %s%s%s%s

Estimate monthly infrastructure costs.`,
		provider, region, env, infraHint, componentsHint, techStackHint)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("cost estimation failed: %w", err)
	}

	var estimate CostEstimate
	if err := extractJSON(resp.Response, &estimate); err != nil {
		return nil, fmt.Errorf("failed to parse cost estimate: %w", err)
	}

	return &estimate, nil
}

// RefineInfra refines infrastructure code based on feedback
func (a *InfraAgent) RefineInfra(ctx context.Context, input map[string]interface{}) (*InfraCodeBundle, error) {
	currentFiles, _ := input["current_files"].(map[string]interface{})
	feedback, _ := input["feedback"].(string)
	tool, _ := input["tool"].(string)
	provider, _ := input["provider"].(string)

	if tool == "" {
		tool = "terraform"
	}
	if provider == "" {
		provider = "aws"
	}

	sysPrompt := fmt.Sprintf(`You are an Infrastructure expert refining %s code for %s.
Apply the feedback to improve the infrastructure code.

Output ONLY valid JSON with this structure:
{
  "provider": "%s",
  "tool": "%s",
  "files": {
    "path/to/file": "updated content"
  },
  "init_cmd": "initialization command",
  "plan_cmd": "plan command",
  "apply_cmd": "apply command",
  "destroy_cmd": "destroy command"
}

Apply the feedback while maintaining:
1. Security best practices
2. Cost optimization
3. High availability
4. Proper resource tagging
5. Least privilege IAM policies`, tool, provider, provider, tool)

	filesStr, _ := json.MarshalIndent(currentFiles, "", "  ")
	prompt := fmt.Sprintf(`CURRENT FILES:
%s

FEEDBACK:
%s

Apply the feedback and return improved infrastructure code.`, string(filesStr), feedback)

	resp, err := a.llmClient.Generate(ctx, llm.GenerateRequest{
		Prompt: prompt,
		System: sysPrompt,
		Format: "json",
	})
	if err != nil {
		return nil, fmt.Errorf("infrastructure refinement failed: %w", err)
	}

	var bundle InfraCodeBundle
	if err := extractJSON(resp.Response, &bundle); err != nil {
		return nil, fmt.Errorf("failed to parse refined infrastructure output: %w", err)
	}

	return &bundle, nil
}

// ============================================================================
// Prompt Generators
// ============================================================================

func (a *InfraAgent) getTerraformPrompt(provider string, modular bool) string {
	basePrompt := `You are a Terraform expert creating production-ready infrastructure code.

Output ONLY valid JSON with this structure:
{
  "files": {
    "main.tf": "main terraform config",
    "variables.tf": "variable definitions",
    "outputs.tf": "output definitions",
    "providers.tf": "provider configuration",
    "terraform.tfvars.example": "example variable values"
  },
  "init_cmd": "terraform init",
  "plan_cmd": "terraform plan -var-file=terraform.tfvars",
  "apply_cmd": "terraform apply -var-file=terraform.tfvars -auto-approve",
  "destroy_cmd": "terraform destroy -var-file=terraform.tfvars -auto-approve"
}

Terraform Best Practices:
1. Use modules for reusable components
2. Use variables for all configurable values
3. Use locals for computed values
4. Use data sources for existing resources
5. Tag all resources with standard tags
6. Use count/for_each for multiple resources
7. Enable encryption at rest for all services
8. Use private subnets for backend services
9. Implement least-privilege IAM policies
10. Use remote state with state locking
`

	if modular {
		basePrompt += `
Modular Structure:
- modules/vpc/main.tf, variables.tf, outputs.tf
- modules/ecs/main.tf, variables.tf, outputs.tf
- modules/rds/main.tf, variables.tf, outputs.tf
- environments/dev/main.tf, terraform.tfvars
- environments/prod/main.tf, terraform.tfvars
`
	}

	switch provider {
	case "aws":
		basePrompt += `
AWS-Specific Guidelines:
- Use aws provider with proper version constraints
- Use aws_region data source
- Enable VPC flow logs
- Use security groups with specific rules
- Enable CloudWatch logging
- Use KMS for encryption keys
- Configure S3 bucket policies
- Use ALB for load balancing
- Configure auto-scaling policies
`
	case "azure":
		basePrompt += `
Azure-Specific Guidelines:
- Use azurerm provider with proper version constraints
- Use resource groups for organization
- Enable Azure Monitor
- Use NSGs with specific rules
- Enable Azure Key Vault for secrets
- Use Azure Front Door or Application Gateway
- Configure auto-scaling with VMSS
- Use managed identities
`
	case "gcp":
		basePrompt += `
GCP-Specific Guidelines:
- Use google provider with proper version constraints
- Use projects for organization
- Enable Cloud Audit Logs
- Use VPC firewall rules
- Enable Cloud KMS for encryption
- Use Cloud Load Balancing
- Configure instance groups with autoscaler
- Use service accounts with minimal permissions
`
	}

	return basePrompt
}

func (a *InfraAgent) getCloudFormationPrompt(nestedStacks bool) string {
	basePrompt := `You are a CloudFormation expert creating production-ready AWS templates.

Output ONLY valid JSON with this structure:
{
  "files": {
    "template.yaml": "main cloudformation template",
    "parameters-dev.json": "development parameters",
    "parameters-prod.json": "production parameters"
  },
  "init_cmd": "aws cloudformation validate-template --template-body file://template.yaml",
  "plan_cmd": "aws cloudformation deploy --template-file template.yaml --stack-name mystack --parameter-overrides file://parameters-dev.json --no-execute-changeset",
  "apply_cmd": "aws cloudformation deploy --template-file template.yaml --stack-name mystack --parameter-overrides file://parameters-dev.json --capabilities CAPABILITY_IAM",
  "destroy_cmd": "aws cloudformation delete-stack --stack-name mystack"
}

CloudFormation Best Practices:
1. Use Parameters for configurable values
2. Use Mappings for environment-specific values
3. Use Conditions for optional resources
4. Use Outputs for important values
5. Use !Ref and !GetAtt for references
6. Use DependsOn for explicit dependencies
7. Enable termination protection for production
8. Use stack policies to prevent updates
9. Enable encryption for all services
10. Use IAM roles with least privilege
`

	if nestedStacks {
		basePrompt += `
Nested Stack Structure:
- master.yaml (parent stack)
- networking.yaml (VPC, subnets, etc.)
- compute.yaml (ECS, EC2, etc.)
- database.yaml (RDS, DynamoDB, etc.)
- storage.yaml (S3, EFS, etc.)
`
	}

	return basePrompt
}

func (a *InfraAgent) getPulumiPrompt(provider, language string) string {
	basePrompt := fmt.Sprintf(`You are a Pulumi expert creating production-ready infrastructure code in %s.

Output ONLY valid JSON with this structure:
{
  "files": {
    "Pulumi.yaml": "pulumi project config",
    "Pulumi.dev.yaml": "development stack config",
    "Pulumi.prod.yaml": "production stack config",
    "__main__.py": "main pulumi code (for python)",
    "index.ts": "main pulumi code (for typescript)",
    "main.go": "main pulumi code (for go)"
  },
  "init_cmd": "pulumi stack init dev",
  "plan_cmd": "pulumi preview",
  "apply_cmd": "pulumi up --yes",
  "destroy_cmd": "pulumi destroy --yes"
}

Pulumi Best Practices:
1. Use stack configuration for environment-specific values
2. Use ComponentResources for reusable modules
3. Use Pulumi secrets for sensitive data
4. Tag all resources with standard tags
5. Use proper resource naming conventions
6. Handle dependencies correctly
7. Export important outputs
8. Use strongly-typed configurations
9. Enable encryption at rest
10. Implement least-privilege IAM
`, language)

	switch language {
	case "typescript":
		basePrompt += `
TypeScript Guidelines:
- Use @pulumi/aws, @pulumi/azure, or @pulumi/gcp
- Use async/await for resource creation
- Use pulumi.output for computed values
- Use interfaces for configuration types
`
	case "python":
		basePrompt += `
Python Guidelines:
- Use pulumi_aws, pulumi_azure, or pulumi_gcp
- Use type hints for configuration
- Use dataclasses for configuration objects
- Follow PEP 8 style guidelines
`
	case "go":
		basePrompt += `
Go Guidelines:
- Use github.com/pulumi/pulumi-aws/sdk/v5/go
- Use proper error handling
- Use structs for configuration
- Follow Go idioms
`
	}

	return basePrompt
}

// ============================================================================
// Utility Functions
// ============================================================================

func getDefaultRegion(provider string) string {
	switch provider {
	case "aws":
		return "us-east-1"
	case "azure":
		return "eastus"
	case "gcp":
		return "us-central1"
	default:
		return "us-east-1"
	}
}

func describeTechStackForInfra(ts *TechStack) string {
	if ts == nil {
		return ""
	}

	var parts []string

	if ts.Backend.Framework != "" || ts.Backend.Language != "" {
		parts = append(parts, fmt.Sprintf("Backend: %s/%s", ts.Backend.Language, ts.Backend.Framework))
	}

	if ts.Database.Primary != "" {
		parts = append(parts, fmt.Sprintf("Database: %s", ts.Database.Primary))
	}

	if ts.Database.Cache != "" {
		parts = append(parts, fmt.Sprintf("Cache: %s", ts.Database.Cache))
	}

	if ts.Database.Queue != "" {
		parts = append(parts, fmt.Sprintf("Queue: %s", ts.Database.Queue))
	}

	if ts.Infrastructure.Cloud != "" {
		parts = append(parts, fmt.Sprintf("Cloud: %s", ts.Infrastructure.Cloud))
	}

	if len(parts) == 0 {
		return ""
	}

	return "\n\nTech Stack:\n- " + strings.Join(parts, "\n- ")
}

// GetAWSServiceMapping returns AWS service names for common components
func GetAWSServiceMapping() map[string]string {
	return map[string]string{
		"compute":    "ECS/Fargate",
		"database":   "RDS",
		"nosql":      "DynamoDB",
		"storage":    "S3",
		"cdn":        "CloudFront",
		"cache":      "ElastiCache",
		"queue":      "SQS",
		"pubsub":     "SNS",
		"serverless": "Lambda",
		"api":        "API Gateway",
		"dns":        "Route 53",
		"secrets":    "Secrets Manager",
		"monitoring": "CloudWatch",
		"logging":    "CloudWatch Logs",
	}
}

// GetAzureServiceMapping returns Azure service names for common components
func GetAzureServiceMapping() map[string]string {
	return map[string]string{
		"compute":    "AKS/Container Apps",
		"database":   "Azure SQL/PostgreSQL",
		"nosql":      "Cosmos DB",
		"storage":    "Blob Storage",
		"cdn":        "Azure CDN",
		"cache":      "Azure Cache for Redis",
		"queue":      "Service Bus",
		"pubsub":     "Event Grid",
		"serverless": "Azure Functions",
		"api":        "API Management",
		"dns":        "Azure DNS",
		"secrets":    "Key Vault",
		"monitoring": "Azure Monitor",
		"logging":    "Log Analytics",
	}
}

// GetGCPServiceMapping returns GCP service names for common components
func GetGCPServiceMapping() map[string]string {
	return map[string]string{
		"compute":    "GKE/Cloud Run",
		"database":   "Cloud SQL",
		"nosql":      "Firestore/Bigtable",
		"storage":    "Cloud Storage",
		"cdn":        "Cloud CDN",
		"cache":      "Memorystore",
		"queue":      "Cloud Tasks",
		"pubsub":     "Pub/Sub",
		"serverless": "Cloud Functions",
		"api":        "API Gateway",
		"dns":        "Cloud DNS",
		"secrets":    "Secret Manager",
		"monitoring": "Cloud Monitoring",
		"logging":    "Cloud Logging",
	}
}

// GenerateTerraformVPCModule generates a basic VPC module for AWS
func GenerateTerraformVPCModule(projectName, region string) string {
	return fmt.Sprintf(`# VPC Module for %s
# Region: %s

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "%s"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones"
  type        = list(string)
  default     = ["%s-1a", "%s-1b", "%s-1c"]
}

locals {
  name_prefix = "${var.project_name}-${var.environment}"

  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-vpc"
  })
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-igw"
  })
}

# Public Subnets
resource "aws_subnet" "public" {
  count = length(var.availability_zones)

  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 4, count.index)
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-public-${count.index + 1}"
    Type = "public"
  })
}

# Private Subnets
resource "aws_subnet" "private" {
  count = length(var.availability_zones)

  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 4, count.index + length(var.availability_zones))
  availability_zone = var.availability_zones[count.index]

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-private-${count.index + 1}"
    Type = "private"
  })
}

# NAT Gateway (one per AZ for HA)
resource "aws_eip" "nat" {
  count  = length(var.availability_zones)
  domain = "vpc"

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-nat-eip-${count.index + 1}"
  })
}

resource "aws_nat_gateway" "main" {
  count = length(var.availability_zones)

  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-nat-${count.index + 1}"
  })

  depends_on = [aws_internet_gateway.main]
}

# Route Tables
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-public-rt"
  })
}

resource "aws_route_table" "private" {
  count  = length(var.availability_zones)
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main[count.index].id
  }

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-private-rt-${count.index + 1}"
  })
}

# Route Table Associations
resource "aws_route_table_association" "public" {
  count = length(var.availability_zones)

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "private" {
  count = length(var.availability_zones)

  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}

# Outputs
output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = aws_subnet.private[*].id
}
`, projectName, region, projectName, region, region, region)
}

// EstimateBasicCost provides a quick cost estimate for common configurations
func EstimateBasicCost(provider, environment string, components []string) float64 {
	// Base costs per environment
	baseCosts := map[string]float64{
		"development": 50.0,
		"staging":     150.0,
		"production":  500.0,
	}

	// Component costs (rough estimates)
	componentCosts := map[string]float64{
		"vpc":         10.0,  // NAT Gateway primarily
		"compute":     100.0, // 2 small instances
		"database":    50.0,  // Small RDS
		"cache":       30.0,  // Small ElastiCache
		"storage":     5.0,   // S3 bucket
		"cdn":         20.0,  // CloudFront
		"queue":       5.0,   // SQS
		"serverless":  10.0,  // Lambda
		"monitoring":  15.0,  // CloudWatch
		"loadbalancer": 25.0, // ALB
	}

	total := baseCosts[environment]
	if total == 0 {
		total = baseCosts["development"]
	}

	for _, comp := range components {
		if cost, ok := componentCosts[strings.ToLower(comp)]; ok {
			total += cost
		}
	}

	// Provider adjustments (rough)
	switch provider {
	case "azure":
		total *= 1.05 // Azure slightly more expensive
	case "gcp":
		total *= 0.95 // GCP slightly cheaper
	}

	return total
}
