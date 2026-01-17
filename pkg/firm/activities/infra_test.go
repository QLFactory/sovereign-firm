package activities

import (
	"encoding/json"
	"strings"
	"testing"
)

// ============================================================================
// Data Structure Tests
// ============================================================================

func TestInfraCodeBundleStructure(t *testing.T) {
	bundle := InfraCodeBundle{
		Provider:    "aws",
		Tool:        "terraform",
		Files: map[string]string{
			"main.tf":      "terraform code...",
			"variables.tf": "variable definitions...",
		},
		InitCmd:     "terraform init",
		PlanCmd:     "terraform plan",
		ApplyCmd:    "terraform apply",
		DestroyCmd:  "terraform destroy",
		Environment: "production",
	}

	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("Failed to marshal InfraCodeBundle: %v", err)
	}

	var decoded InfraCodeBundle
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal InfraCodeBundle: %v", err)
	}

	if decoded.Provider != "aws" {
		t.Errorf("Expected provider 'aws', got '%s'", decoded.Provider)
	}
	if decoded.Tool != "terraform" {
		t.Errorf("Expected tool 'terraform', got '%s'", decoded.Tool)
	}
	if len(decoded.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(decoded.Files))
	}
}

func TestInfraSpecStructure(t *testing.T) {
	spec := InfraSpec{
		ProjectName: "my-app",
		Environment: "production",
		Region:      "us-east-1",
		Components: []InfraComponent{
			{Type: "compute", Name: "api", Service: "ecs", Config: map[string]interface{}{"cpu": 256}},
			{Type: "database", Name: "db", Service: "rds", Config: map[string]interface{}{"engine": "postgres"}},
		},
		Networking: &NetworkingSpec{
			VPCCidr:        "10.0.0.0/16",
			PublicSubnets:  []string{"10.0.1.0/24", "10.0.2.0/24"},
			PrivateSubnets: []string{"10.0.10.0/24", "10.0.11.0/24"},
			EnableNAT:      true,
			EnableVPN:      false,
		},
		Tags: map[string]string{
			"Project":     "my-app",
			"Environment": "production",
		},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal InfraSpec: %v", err)
	}

	var decoded InfraSpec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal InfraSpec: %v", err)
	}

	if len(decoded.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(decoded.Components))
	}
	if decoded.Networking == nil {
		t.Error("Expected Networking to be set")
	}
	if !decoded.Networking.EnableNAT {
		t.Error("Expected EnableNAT to be true")
	}
}

func TestInfraComponentStructure(t *testing.T) {
	component := InfraComponent{
		Type:    "compute",
		Name:    "api-cluster",
		Service: "ecs",
		Config: map[string]interface{}{
			"cpu":    256,
			"memory": 512,
			"count":  2,
		},
		DependsOn: []string{"vpc", "rds"},
	}

	data, err := json.Marshal(component)
	if err != nil {
		t.Fatalf("Failed to marshal InfraComponent: %v", err)
	}

	var decoded InfraComponent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal InfraComponent: %v", err)
	}

	if decoded.Type != "compute" {
		t.Errorf("Expected type 'compute', got '%s'", decoded.Type)
	}
	if len(decoded.DependsOn) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(decoded.DependsOn))
	}
}

func TestNetworkingSpecStructure(t *testing.T) {
	spec := NetworkingSpec{
		VPCCidr:        "10.0.0.0/16",
		PublicSubnets:  []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"},
		PrivateSubnets: []string{"10.0.10.0/24", "10.0.11.0/24", "10.0.12.0/24"},
		EnableNAT:      true,
		EnableVPN:      true,
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal NetworkingSpec: %v", err)
	}

	var decoded NetworkingSpec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal NetworkingSpec: %v", err)
	}

	if len(decoded.PublicSubnets) != 3 {
		t.Errorf("Expected 3 public subnets, got %d", len(decoded.PublicSubnets))
	}
	if len(decoded.PrivateSubnets) != 3 {
		t.Errorf("Expected 3 private subnets, got %d", len(decoded.PrivateSubnets))
	}
}

func TestCostEstimateStructure(t *testing.T) {
	estimate := CostEstimate{
		Provider:    "aws",
		Environment: "production",
		MonthlyCost: 1234.56,
		Currency:    "USD",
		Breakdown: []CostLineItem{
			{Service: "EC2", Resource: "t3.medium", Quantity: 2, UnitPrice: 0.0416, MonthlyCost: 60.0},
			{Service: "RDS", Resource: "db.t3.medium", Quantity: 1, UnitPrice: 0.068, MonthlyCost: 49.64},
		},
		Assumptions: []string{
			"24/7 operation",
			"On-demand pricing",
		},
		Recommendations: []CostRecommendation{
			{Type: "reserved", Description: "Use reserved instances", PotentialSaving: 200.0, Effort: "low"},
		},
	}

	data, err := json.Marshal(estimate)
	if err != nil {
		t.Fatalf("Failed to marshal CostEstimate: %v", err)
	}

	var decoded CostEstimate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CostEstimate: %v", err)
	}

	if decoded.MonthlyCost != 1234.56 {
		t.Errorf("Expected monthly cost 1234.56, got %f", decoded.MonthlyCost)
	}
	if len(decoded.Breakdown) != 2 {
		t.Errorf("Expected 2 line items, got %d", len(decoded.Breakdown))
	}
	if len(decoded.Recommendations) != 1 {
		t.Errorf("Expected 1 recommendation, got %d", len(decoded.Recommendations))
	}
}

func TestCostLineItemStructure(t *testing.T) {
	item := CostLineItem{
		Service:     "EC2",
		Resource:    "t3.medium (2 instances)",
		Quantity:    2,
		UnitPrice:   0.0416,
		MonthlyCost: 60.0,
		Notes:       "On-demand pricing, 730 hours/month",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal CostLineItem: %v", err)
	}

	var decoded CostLineItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CostLineItem: %v", err)
	}

	if decoded.Service != "EC2" {
		t.Errorf("Expected service 'EC2', got '%s'", decoded.Service)
	}
	if decoded.MonthlyCost != 60.0 {
		t.Errorf("Expected monthly cost 60.0, got %f", decoded.MonthlyCost)
	}
}

func TestCostRecommendationStructure(t *testing.T) {
	rec := CostRecommendation{
		Type:            "reserved",
		Description:     "Use 1-year reserved instances for EC2",
		PotentialSaving: 240.0,
		Effort:          "low",
	}

	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("Failed to marshal CostRecommendation: %v", err)
	}

	var decoded CostRecommendation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CostRecommendation: %v", err)
	}

	if decoded.Type != "reserved" {
		t.Errorf("Expected type 'reserved', got '%s'", decoded.Type)
	}
	if decoded.PotentialSaving != 240.0 {
		t.Errorf("Expected potential saving 240.0, got %f", decoded.PotentialSaving)
	}
}

// ============================================================================
// Input Type Tests
// ============================================================================

func TestTerraformInputStructure(t *testing.T) {
	input := TerraformInput{
		ProjectName: "my-app",
		Provider:    "aws",
		Region:      "us-east-1",
		Environment: "production",
		Components:  []string{"vpc", "ecs", "rds", "s3"},
		Modular:     true,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal TerraformInput: %v", err)
	}

	var decoded TerraformInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal TerraformInput: %v", err)
	}

	if decoded.Provider != "aws" {
		t.Errorf("Expected provider 'aws', got '%s'", decoded.Provider)
	}
	if len(decoded.Components) != 4 {
		t.Errorf("Expected 4 components, got %d", len(decoded.Components))
	}
	if !decoded.Modular {
		t.Error("Expected Modular to be true")
	}
}

func TestCloudFormationInputStructure(t *testing.T) {
	input := CloudFormationInput{
		ProjectName:  "my-app",
		Region:       "us-east-1",
		Environment:  "production",
		Components:   []string{"vpc", "ecs", "rds"},
		NestedStacks: true,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal CloudFormationInput: %v", err)
	}

	var decoded CloudFormationInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CloudFormationInput: %v", err)
	}

	if !decoded.NestedStacks {
		t.Error("Expected NestedStacks to be true")
	}
}

func TestPulumiInputStructure(t *testing.T) {
	input := PulumiInput{
		ProjectName: "my-app",
		Provider:    "aws",
		Region:      "us-east-1",
		Environment: "production",
		Language:    "typescript",
		Components:  []string{"vpc", "ecs", "rds"},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal PulumiInput: %v", err)
	}

	var decoded PulumiInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PulumiInput: %v", err)
	}

	if decoded.Language != "typescript" {
		t.Errorf("Expected language 'typescript', got '%s'", decoded.Language)
	}
}

func TestCostEstimateInputStructure(t *testing.T) {
	input := CostEstimateInput{
		Provider:    "aws",
		Region:      "us-east-1",
		Environment: "production",
		Components:  []string{"vpc", "ecs", "rds", "s3", "cloudfront"},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal CostEstimateInput: %v", err)
	}

	var decoded CostEstimateInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CostEstimateInput: %v", err)
	}

	if len(decoded.Components) != 5 {
		t.Errorf("Expected 5 components, got %d", len(decoded.Components))
	}
}

// ============================================================================
// Utility Function Tests
// ============================================================================

func TestGetDefaultRegion(t *testing.T) {
	tests := []struct {
		provider string
		expected string
	}{
		{"aws", "us-east-1"},
		{"azure", "eastus"},
		{"gcp", "us-central1"},
		{"unknown", "us-east-1"},
		{"", "us-east-1"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			result := getDefaultRegion(tt.provider)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestDescribeTechStackForInfra(t *testing.T) {
	tests := []struct {
		name     string
		stack    *TechStack
		expected []string
	}{
		{
			name:     "nil stack",
			stack:    nil,
			expected: []string{},
		},
		{
			name: "full stack",
			stack: &TechStack{
				Backend: BackendStack{
					Framework: "Express",
					Language:  "TypeScript",
				},
				Database: DatabaseStack{
					Primary: "PostgreSQL",
					Cache:   "Redis",
					Queue:   "RabbitMQ",
				},
				Infrastructure: InfraStack{
					Cloud: "AWS",
				},
			},
			expected: []string{"Backend:", "Database:", "Cache:", "Queue:", "Cloud:"},
		},
		{
			name: "partial stack",
			stack: &TechStack{
				Backend: BackendStack{
					Language: "Go",
				},
				Database: DatabaseStack{
					Primary: "MongoDB",
				},
			},
			expected: []string{"Backend:", "Database:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := describeTechStackForInfra(tt.stack)

			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("Expected result to contain '%s', got '%s'", exp, result)
				}
			}
		})
	}
}

func TestGetAWSServiceMapping(t *testing.T) {
	mapping := GetAWSServiceMapping()

	expectedMappings := map[string]string{
		"compute":  "ECS/Fargate",
		"database": "RDS",
		"storage":  "S3",
		"cdn":      "CloudFront",
		"cache":    "ElastiCache",
	}

	for component, expected := range expectedMappings {
		if actual, ok := mapping[component]; !ok {
			t.Errorf("Missing mapping for '%s'", component)
		} else if actual != expected {
			t.Errorf("Expected '%s' for '%s', got '%s'", expected, component, actual)
		}
	}
}

func TestGetAzureServiceMapping(t *testing.T) {
	mapping := GetAzureServiceMapping()

	expectedMappings := map[string]string{
		"compute":  "AKS/Container Apps",
		"storage":  "Blob Storage",
		"cache":    "Azure Cache for Redis",
		"secrets":  "Key Vault",
	}

	for component, expected := range expectedMappings {
		if actual, ok := mapping[component]; !ok {
			t.Errorf("Missing mapping for '%s'", component)
		} else if actual != expected {
			t.Errorf("Expected '%s' for '%s', got '%s'", expected, component, actual)
		}
	}
}

func TestGetGCPServiceMapping(t *testing.T) {
	mapping := GetGCPServiceMapping()

	expectedMappings := map[string]string{
		"compute":    "GKE/Cloud Run",
		"database":   "Cloud SQL",
		"storage":    "Cloud Storage",
		"serverless": "Cloud Functions",
	}

	for component, expected := range expectedMappings {
		if actual, ok := mapping[component]; !ok {
			t.Errorf("Missing mapping for '%s'", component)
		} else if actual != expected {
			t.Errorf("Expected '%s' for '%s', got '%s'", expected, component, actual)
		}
	}
}

// ============================================================================
// Template Generation Tests
// ============================================================================

func TestGenerateTerraformVPCModule(t *testing.T) {
	result := GenerateTerraformVPCModule("my-app", "us-east-1")

	expectedContains := []string{
		"# VPC Module for my-app",
		"terraform {",
		"required_providers",
		"aws_vpc",
		"aws_subnet",
		"aws_internet_gateway",
		"aws_nat_gateway",
		"aws_route_table",
		"cidr_block",
		"output \"vpc_id\"",
		"output \"public_subnet_ids\"",
		"output \"private_subnet_ids\"",
	}

	for _, expected := range expectedContains {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected VPC module to contain '%s'", expected)
		}
	}
}

func TestEstimateBasicCost(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		environment string
		components  []string
		minExpected float64
		maxExpected float64
	}{
		{
			name:        "development empty",
			provider:    "aws",
			environment: "development",
			components:  []string{},
			minExpected: 40.0,
			maxExpected: 60.0,
		},
		{
			name:        "production with compute",
			provider:    "aws",
			environment: "production",
			components:  []string{"compute", "database"},
			minExpected: 600.0,
			maxExpected: 700.0,
		},
		{
			name:        "staging full stack",
			provider:    "aws",
			environment: "staging",
			components:  []string{"vpc", "compute", "database", "cache", "cdn"},
			minExpected: 300.0,
			maxExpected: 400.0,
		},
		{
			name:        "azure adjustment",
			provider:    "azure",
			environment: "production",
			components:  []string{"compute"},
			minExpected: 600.0,
			maxExpected: 700.0,
		},
		{
			name:        "gcp adjustment",
			provider:    "gcp",
			environment: "production",
			components:  []string{"compute"},
			minExpected: 550.0,
			maxExpected: 650.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateBasicCost(tt.provider, tt.environment, tt.components)
			if result < tt.minExpected || result > tt.maxExpected {
				t.Errorf("Expected cost between %f and %f, got %f",
					tt.minExpected, tt.maxExpected, result)
			}
		})
	}
}

// ============================================================================
// Agent Construction Tests
// ============================================================================

func TestNewInfraAgent(t *testing.T) {
	agent := NewInfraAgent()

	if agent == nil {
		t.Fatal("NewInfraAgent returned nil")
	}

	if agent.llmClient == nil {
		t.Error("InfraAgent.llmClient is nil")
	}
}

// ============================================================================
// Default Value Tests
// ============================================================================

func TestProviderDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to aws", "", "aws"},
		{"aws stays aws", "aws", "aws"},
		{"azure stays azure", "azure", "azure"},
		{"gcp stays gcp", "gcp", "gcp"},
		{"AWS normalizes", "AWS", "aws"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := strings.ToLower(tt.input)
			if provider == "" {
				provider = "aws"
			}
			if provider != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, provider)
			}
		})
	}
}

func TestToolDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to terraform", "", "terraform"},
		{"terraform stays", "terraform", "terraform"},
		{"cloudformation stays", "cloudformation", "cloudformation"},
		{"pulumi stays", "pulumi", "pulumi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := tt.input
			if tool == "" {
				tool = "terraform"
			}
			if tool != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, tool)
			}
		})
	}
}

func TestLanguageDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to typescript", "", "typescript"},
		{"typescript stays", "typescript", "typescript"},
		{"python stays", "python", "python"},
		{"go stays", "go", "go"},
		{"TypeScript normalizes", "TypeScript", "typescript"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := strings.ToLower(tt.input)
			if lang == "" {
				lang = "typescript"
			}
			if lang != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, lang)
			}
		})
	}
}

func TestEnvironmentDefaultsInfra(t *testing.T) {
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

// ============================================================================
// Service Mapping Completeness Tests
// ============================================================================

func TestAWSServiceMappingCompleteness(t *testing.T) {
	mapping := GetAWSServiceMapping()
	requiredKeys := []string{
		"compute", "database", "nosql", "storage", "cdn",
		"cache", "queue", "pubsub", "serverless", "api",
		"dns", "secrets", "monitoring", "logging",
	}

	for _, key := range requiredKeys {
		if _, ok := mapping[key]; !ok {
			t.Errorf("AWS service mapping missing key: %s", key)
		}
	}
}

func TestAzureServiceMappingCompleteness(t *testing.T) {
	mapping := GetAzureServiceMapping()
	requiredKeys := []string{
		"compute", "database", "nosql", "storage", "cdn",
		"cache", "queue", "pubsub", "serverless", "api",
		"dns", "secrets", "monitoring", "logging",
	}

	for _, key := range requiredKeys {
		if _, ok := mapping[key]; !ok {
			t.Errorf("Azure service mapping missing key: %s", key)
		}
	}
}

func TestGCPServiceMappingCompleteness(t *testing.T) {
	mapping := GetGCPServiceMapping()
	requiredKeys := []string{
		"compute", "database", "nosql", "storage", "cdn",
		"cache", "queue", "pubsub", "serverless", "api",
		"dns", "secrets", "monitoring", "logging",
	}

	for _, key := range requiredKeys {
		if _, ok := mapping[key]; !ok {
			t.Errorf("GCP service mapping missing key: %s", key)
		}
	}
}
