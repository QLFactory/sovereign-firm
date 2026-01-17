/**
 * TypeScript types matching the backend API exactly.
 * This is the single source of truth for API contracts.
 */

// =============================================================================
// Enums & Constants
// =============================================================================

export type Phase =
  | "INTAKE"
  | "SIZING"
  | "PLANNING"
  | "ARCHITECTURE"
  | "DEVELOPMENT"
  | "TESTING"
  | "DEPLOYMENT"
  | "OPERATIONS"
  | "HANDOFF"
  | "COMPLETE"
  | "REVIEW"
  | "FAILED";

export const PHASE_ORDER: Phase[] = [
  "INTAKE",
  "SIZING",
  "PLANNING",
  "ARCHITECTURE",
  "DEVELOPMENT",
  "TESTING",
  "DEPLOYMENT",
  "OPERATIONS",
  "HANDOFF",
  "COMPLETE",
];

export const PHASE_COLORS: Record<Phase, string> = {
  INTAKE: "badge-cyan",
  SIZING: "badge-cyan",
  PLANNING: "badge-violet",
  ARCHITECTURE: "badge-violet",
  DEVELOPMENT: "badge-amber",
  TESTING: "badge-amber",
  DEPLOYMENT: "badge-emerald",
  OPERATIONS: "badge-emerald",
  HANDOFF: "badge-emerald",
  COMPLETE: "badge-emerald",
  REVIEW: "badge-violet",
  FAILED: "badge-rose",
};

// =============================================================================
// Project Configuration
// =============================================================================

export interface ProjectConfig {
  project_name: string;
  client_id?: string;
  initial_message?: string;
  enable_full_stack: boolean;
  enable_deployment: boolean;
  enable_sre: boolean;
  preferred_frontend: string;
  preferred_backend: string;
  preferred_database: string;
  preferred_cloud: string;
}

// =============================================================================
// ConsultancyState - The main project state from the backend
// =============================================================================

export interface ConsultancyState {
  // Core identifiers
  project_name: string;
  workflow_id?: string;

  // Phase tracking
  phase: Phase;
  phase_history: string[];

  // Chat and spec
  chat_history?: string;
  spec?: string;

  // Frontend code (17 files typically)
  frontend_code: Record<string, string>;

  // Backend code (29 files typically)
  backend_code: Record<string, string>;

  // Database code (1 file typically)
  database_code: Record<string, string>;

  // ALL code files merged (78 files) - use this for display
  all_code_files: Record<string, string>;

  // Tests (10 files typically)
  unit_tests: Record<string, string>;
  integration_tests: Record<string, string>;
  e2e_tests: Record<string, string>;

  // DevOps - Single files
  dockerfile: string;
  docker_compose: string;
  ci_pipeline: string;

  // Kubernetes (7 files typically)
  kube_manifests: Record<string, string>;

  // Helm Chart (12 files typically)
  helm_chart: Record<string, string>;

  // Terraform/Infrastructure (5 files typically)
  infra_code: Record<string, string>;

  // SRE - Monitoring & Observability
  monitoring_config: Record<string, unknown>;
  alert_rules: Record<string, unknown>;
  dashboards: Record<string, unknown>;
  runbooks: Record<string, string>;

  // Errors and warnings
  errors: string[];
  warnings: string[];
}

// =============================================================================
// Artifact Categories - For organizing files in the UI
// =============================================================================

export interface ArtifactCategory {
  id: string;
  name: string;
  icon: string;
  description: string;
  fileCount: number;
  files: Record<string, string>;
}

export function categorizeArtifacts(state: ConsultancyState): ArtifactCategory[] {
  const categories: ArtifactCategory[] = [];

  // Frontend
  if (state.frontend_code && Object.keys(state.frontend_code).length > 0) {
    categories.push({
      id: "frontend",
      name: "Frontend",
      icon: "⚛️",
      description: "React components, pages, and styles",
      fileCount: Object.keys(state.frontend_code).length,
      files: state.frontend_code,
    });
  }

  // Backend
  if (state.backend_code && Object.keys(state.backend_code).length > 0) {
    categories.push({
      id: "backend",
      name: "Backend",
      icon: "🔧",
      description: "API routes, controllers, and models",
      fileCount: Object.keys(state.backend_code).length,
      files: state.backend_code,
    });
  }

  // Database
  if (state.database_code && Object.keys(state.database_code).length > 0) {
    categories.push({
      id: "database",
      name: "Database",
      icon: "🗄️",
      description: "Schema and migrations",
      fileCount: Object.keys(state.database_code).length,
      files: state.database_code,
    });
  }

  // Tests
  const allTests = {
    ...state.unit_tests,
    ...state.integration_tests,
    ...state.e2e_tests,
  };
  if (Object.keys(allTests).length > 0) {
    categories.push({
      id: "tests",
      name: "Tests",
      icon: "🧪",
      description: "Unit, integration, and e2e tests",
      fileCount: Object.keys(allTests).length,
      files: allTests,
    });
  }

  // DevOps (single files)
  const devopsFiles: Record<string, string> = {};
  if (state.dockerfile) devopsFiles["Dockerfile"] = state.dockerfile;
  if (state.docker_compose) devopsFiles["docker-compose.yml"] = state.docker_compose;
  if (state.ci_pipeline) devopsFiles[".github/workflows/ci.yml"] = state.ci_pipeline;

  if (Object.keys(devopsFiles).length > 0) {
    categories.push({
      id: "devops",
      name: "DevOps",
      icon: "🐳",
      description: "Dockerfile, docker-compose, CI pipeline",
      fileCount: Object.keys(devopsFiles).length,
      files: devopsFiles,
    });
  }

  // Kubernetes
  if (state.kube_manifests && Object.keys(state.kube_manifests).length > 0) {
    categories.push({
      id: "kubernetes",
      name: "Kubernetes",
      icon: "☸️",
      description: "Deployment, service, ingress, HPA",
      fileCount: Object.keys(state.kube_manifests).length,
      files: state.kube_manifests,
    });
  }

  // Helm
  if (state.helm_chart && Object.keys(state.helm_chart).length > 0) {
    categories.push({
      id: "helm",
      name: "Helm Chart",
      icon: "⎈",
      description: "Chart, values, and templates",
      fileCount: Object.keys(state.helm_chart).length,
      files: state.helm_chart,
    });
  }

  // Terraform
  if (state.infra_code && Object.keys(state.infra_code).length > 0) {
    categories.push({
      id: "terraform",
      name: "Terraform",
      icon: "🏗️",
      description: "VPC, ECS, RDS, Redis",
      fileCount: Object.keys(state.infra_code).length,
      files: state.infra_code,
    });
  }

  // SRE - Runbooks
  if (state.runbooks && Object.keys(state.runbooks).length > 0) {
    categories.push({
      id: "sre",
      name: "SRE",
      icon: "📟",
      description: "Runbooks, alerts, dashboards",
      fileCount: Object.keys(state.runbooks).length,
      files: state.runbooks,
    });
  }

  return categories;
}

// =============================================================================
// Project - Client-side project representation
// =============================================================================

export interface Project {
  id: string;
  name: string;
  phase: Phase;
  status: "running" | "complete" | "failed" | "pending";
  createdAt: string;
  config: ProjectConfig;
  state?: ConsultancyState;
}

// =============================================================================
// API Request/Response Types
// =============================================================================

export interface CreateProjectRequest {
  project_name: string;
  initial_message?: string;
  enable_full_stack?: boolean;
  enable_deployment?: boolean;
  enable_sre?: boolean;
  preferred_frontend?: string;
  preferred_backend?: string;
  preferred_database?: string;
  preferred_cloud?: string;
}

export interface CreateProjectResponse {
  workflow_id: string;
  message?: string;
}

export interface SendMessageRequest {
  message: string;
}

export interface SendMessageResponse {
  status: string;
  message?: string;
}

export interface ApiError {
  error: string;
  details?: string;
}

// =============================================================================
// Stream Event Types (re-exported from useStreaming for convenience)
// =============================================================================

export type StreamEventType =
  | "CONNECTED"
  | "DISCONNECTED"
  | "ERROR"
  | "PHASE_CHANGE"
  | "AGENT_SPAWN"
  | "AGENT_DONE"
  | "CODE_CHUNK"
  | "CODE_COMPLETE"
  | "FILE_START"
  | "FILE_END"
  | "CHAT_MESSAGE"
  | "CHAT_TYPING"
  | "BUILD_START"
  | "BUILD_OUTPUT"
  | "BUILD_SUCCESS"
  | "BUILD_ERROR"
  | "TEST_START"
  | "TEST_RESULT"
  | "PREVIEW_READY"
  | "EXECUTION_STARTED"
  | "EXECUTION_COMPLETED"
  | "EXECUTION_FAILED"
  | "TASK_STARTED"
  | "TASK_COMPLETED"
  | "TASK_FAILED"
  | "TASK_RETRY"
  | "AGENT_ASSIGNED"
  | "BRANCH_MERGED"
  | "MERGE_FAILED"
  | "CI_STAGE_START"
  | "CI_STAGE_COMPLETE"
  | "CI_STAGE_FAILED"
  | "DAG_UPDATE";

export interface StreamEvent {
  type: StreamEventType;
  timestamp: string;
  seq: number;
  payload?: Record<string, unknown>;
}

// =============================================================================
// DAG & Multi-Agent Types
// =============================================================================

export type TaskStatus = "PENDING" | "READY" | "RUNNING" | "COMPLETED" | "FAILED" | "BLOCKED";

export interface DAGTask {
  id: string;
  name: string;
  description?: string;
  type: string;
  status: TaskStatus;
  dependencies: string[];
  assigned_to?: string;
  error?: string;
  retry_count?: number;
}

export interface DAGState {
  id: string;
  project_id: string;
  tasks: DAGTask[];
  total: number;
  completed: number;
  failed: number;
  running: number;
  pending: number;
}

export interface ActiveAgent {
  task_id: string;
  agent_id: string;
  agent_name: string;
  task_name: string;
  started_at: string;
}

export type CIStage = "LINT" | "BUILD" | "TEST";

export interface CIStageResult {
  stage: CIStage;
  success: boolean;
  output?: string;
  error?: string;
  duration_ms?: number;
}

// =============================================================================
// UI State Types
// =============================================================================

export type LeftPanelTab = "chat" | "spec" | "files" | "tasks" | "agents";
export type RightPanelTab = "preview" | "code" | "terminal" | "ci" | "events";

export interface UIState {
  activeLeftTab: LeftPanelTab;
  activeRightTab: RightPanelTab;
  selectedFile: string | null;
  selectedCategory: string | null;
  sidebarCollapsed: boolean;
}

// =============================================================================
// Helper Functions
// =============================================================================

export function getPhaseIndex(phase: Phase): number {
  return PHASE_ORDER.indexOf(phase);
}

export function getPhaseProgress(phase: Phase): number {
  if (phase === "COMPLETE") return 100;
  if (phase === "FAILED") return 0;
  const index = getPhaseIndex(phase);
  if (index < 0) return 0;
  return Math.round((index / (PHASE_ORDER.length - 1)) * 100);
}

export function getTotalFileCount(state: ConsultancyState): number {
  // Prefer all_code_files if available
  if (state.all_code_files && Object.keys(state.all_code_files).length > 0) {
    return Object.keys(state.all_code_files).length;
  }

  // Otherwise count from individual categories
  let count = 0;
  count += Object.keys(state.frontend_code || {}).length;
  count += Object.keys(state.backend_code || {}).length;
  count += Object.keys(state.database_code || {}).length;
  count += Object.keys(state.unit_tests || {}).length;
  count += Object.keys(state.integration_tests || {}).length;
  count += Object.keys(state.e2e_tests || {}).length;
  count += Object.keys(state.kube_manifests || {}).length;
  count += Object.keys(state.helm_chart || {}).length;
  count += Object.keys(state.infra_code || {}).length;
  count += Object.keys(state.runbooks || {}).length;
  if (state.dockerfile) count++;
  if (state.docker_compose) count++;
  if (state.ci_pipeline) count++;

  return count;
}

export function getAllFiles(state: ConsultancyState): Record<string, string> {
  // Prefer all_code_files if available
  if (state.all_code_files && Object.keys(state.all_code_files).length > 0) {
    return state.all_code_files;
  }

  // Otherwise merge all categories
  const files: Record<string, string> = {};

  Object.assign(files, state.frontend_code || {});
  Object.assign(files, state.backend_code || {});
  Object.assign(files, state.database_code || {});
  Object.assign(files, state.unit_tests || {});
  Object.assign(files, state.integration_tests || {});
  Object.assign(files, state.e2e_tests || {});
  Object.assign(files, state.kube_manifests || {});
  Object.assign(files, state.helm_chart || {});
  Object.assign(files, state.infra_code || {});
  Object.assign(files, state.runbooks || {});

  if (state.dockerfile) files["Dockerfile"] = state.dockerfile;
  if (state.docker_compose) files["docker-compose.yml"] = state.docker_compose;
  if (state.ci_pipeline) files[".github/workflows/ci.yml"] = state.ci_pipeline;

  return files;
}
