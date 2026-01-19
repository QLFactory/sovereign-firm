/**
 * Type-safe API client for the Sovereign Firm backend.
 */

import type {
  ConsultancyState,
  CreateProjectRequest,
  CreateProjectResponse,
  SendMessageRequest,
  SendMessageResponse,
  ApiError,
  Project,
  ProjectConfig,
  User,
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  RefreshResponse,
  Invitation,
} from "./types";

const API_BASE = "/api";
const AUTH_TOKEN_KEY = "sovereign-firm-access-token";
const REFRESH_TOKEN_KEY = "sovereign-firm-refresh-token";

class ApiClient {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;

  constructor() {
    // Load tokens from localStorage on init
    if (typeof window !== "undefined") {
      this.accessToken = localStorage.getItem(AUTH_TOKEN_KEY);
      this.refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY);
    }
  }

  // Token management
  setTokens(accessToken: string, refreshToken: string): void {
    this.accessToken = accessToken;
    this.refreshToken = refreshToken;
    if (typeof window !== "undefined") {
      localStorage.setItem(AUTH_TOKEN_KEY, accessToken);
      localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
    }
  }

  clearTokens(): void {
    this.accessToken = null;
    this.refreshToken = null;
    if (typeof window !== "undefined") {
      localStorage.removeItem(AUTH_TOKEN_KEY);
      localStorage.removeItem(REFRESH_TOKEN_KEY);
    }
  }

  getAccessToken(): string | null {
    return this.accessToken;
  }

  isAuthenticated(): boolean {
    return !!this.accessToken;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    includeAuth: boolean = true
  ): Promise<T> {
    const url = `${API_BASE}${endpoint}`;

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };

    // Add auth header if we have a token and auth is needed
    if (includeAuth && this.accessToken) {
      headers["Authorization"] = `Bearer ${this.accessToken}`;
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    // Handle 401 - try to refresh token
    if (response.status === 401 && this.refreshToken && includeAuth) {
      const refreshed = await this.tryRefreshToken();
      if (refreshed) {
        // Retry the request with new token
        headers["Authorization"] = `Bearer ${this.accessToken}`;
        const retryResponse = await fetch(url, { ...options, headers });
        if (!retryResponse.ok) {
          const error: ApiError = await retryResponse.json().catch(() => ({
            error: `Request failed with status ${retryResponse.status}`,
          }));
          throw new Error(error.error || "Request failed");
        }
        return retryResponse.json();
      } else {
        // Refresh failed, clear tokens
        this.clearTokens();
        throw new Error("Session expired. Please log in again.");
      }
    }

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({
        error: `Request failed with status ${response.status}`,
      }));
      throw new Error(error.error || "Request failed");
    }

    return response.json();
  }

  private async tryRefreshToken(): Promise<boolean> {
    if (!this.refreshToken) return false;

    try {
      const response = await fetch(`${API_BASE}/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: this.refreshToken }),
      });

      if (!response.ok) return false;

      const data: RefreshResponse = await response.json();
      this.setTokens(data.access_token, data.refresh_token);
      return true;
    } catch {
      return false;
    }
  }

  // ==========================================================================
  // Authentication
  // ==========================================================================

  /**
   * Register a new user and tenant
   */
  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>(
      "/auth/register",
      {
        method: "POST",
        body: JSON.stringify(data),
      },
      false
    );
    this.setTokens(response.access_token, response.refresh_token);
    return response;
  }

  /**
   * Login with email and password
   */
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>(
      "/auth/login",
      {
        method: "POST",
        body: JSON.stringify(data),
      },
      false
    );
    this.setTokens(response.access_token, response.refresh_token);
    return response;
  }

  /**
   * Logout - clear tokens and invalidate refresh token
   */
  async logout(): Promise<void> {
    try {
      if (this.refreshToken) {
        await this.request(
          "/auth/logout",
          {
            method: "POST",
            body: JSON.stringify({ refresh_token: this.refreshToken }),
          },
          true
        );
      }
    } catch {
      // Ignore errors, we're logging out anyway
    } finally {
      this.clearTokens();
    }
  }

  async getCurrentUser(): Promise<User> {
    return this.request<User>("/auth/me");
  }

  /**
   * List all users in the current tenant
   */
  async listUsers(): Promise<User[]> {
    return this.request<User[]>("/users");
  }

  /**
   * Update a user's role
   */
  async updateUserRole(userId: string, role: string): Promise<void> {
    await this.request(`/users/${userId}/role`, {
      method: "PUT",
      body: JSON.stringify({ role }),
    });
  }

  /**
   * Invite a new user
   */
  async inviteUser(email: string, role: string = "member"): Promise<Invitation> {
    return this.request<Invitation>("/users/invite", {
      method: "POST",
      body: JSON.stringify({ email, role }),
    });
  }

  // ==========================================================================
  // Projects
  // ==========================================================================

  /**
   * Create a new project (uses authenticated /projects endpoint)
   */
  async createProject(config: CreateProjectRequest): Promise<CreateProjectResponse> {
    // Use authenticated endpoint if we have a token, otherwise fallback to legacy
    if (this.accessToken) {
      const response = await this.request<{
        id: string;
        workflow_id: string;
        name: string;
        phase: string;
      }>("/projects", {
        method: "POST",
        body: JSON.stringify({
          name: config.project_name,
          description: config.initial_message,
          initial_message: config.initial_message,
          enable_full_stack: config.enable_full_stack,
          enable_deployment: config.enable_deployment,
          enable_sre: config.enable_sre,
          preferred_frontend: config.preferred_frontend,
          preferred_backend: config.preferred_backend,
          preferred_database: config.preferred_database,
          preferred_cloud: config.preferred_cloud,
        }),
      });
      return {
        id: response.id,
        workflow_id: response.workflow_id,
        status: "started",
      };
    }

    // Fallback to legacy endpoint for unauthenticated access
    return this.request<CreateProjectResponse>("/pods", {
      method: "POST",
      body: JSON.stringify(config),
    }, false);
  }

  /**
   * Get project state by workflow ID or project ID
   */
  async getProject(id: string): Promise<ConsultancyState> {
    // Try authenticated endpoint first
    if (this.accessToken) {
      try {
        // First try to get state from /projects/{id}/state
        return await this.request<ConsultancyState>(`/projects/${id}/state`);
      } catch {
        // Fallback to legacy endpoint
      }
    }
    return this.request<ConsultancyState>(`/pods/${id}`, {}, false);
  }

  /**
   * List all projects for the current user
   */
  async listProjects(): Promise<Project[]> {
    if (this.accessToken) {
      try {
        const response = await this.request<{
          projects: Array<{
            id: string;
            workflow_id: string;
            name: string;
            description?: string;
            phase: string;
            status: string;
            created_at: string;
          }>;
        }>("/projects");

        // Transform to Project format
        return response.projects.map((p) => ({
          id: p.id,
          workflow_id: p.workflow_id,
          name: p.name,
          description: p.description,
          status: p.status === "active" ? "running" : p.status,
          phase: p.phase as Project["phase"],
          created_at: p.created_at,
        }));
      } catch {
        return [];
      }
    }
    return [];
  }

  /**
   * Send a message to a project (e.g., /approve, feedback)
   */
  async sendMessage(
    projectId: string,
    message: string
  ): Promise<SendMessageResponse> {
    const request: SendMessageRequest = { message };

    if (this.accessToken) {
      try {
        await this.request(`/projects/${projectId}/message`, {
          method: "POST",
          body: JSON.stringify(request),
        });
        return { success: true };
      } catch {
        // Fallback to legacy
      }
    }

    return this.request<SendMessageResponse>(`/pods/${projectId}`, {
      method: "POST",
      body: JSON.stringify(request),
    }, false);
  }

  // ==========================================================================
  // Brownfield Import
  // ==========================================================================

  /**
   * Import an existing project (brownfield import)
   */
  async importBrownfield(params: {
    name: string;
    description?: string;
    repo_url?: string;
    local_path?: string;
  }): Promise<{ workflow_id: string; project_id: string; status: string }> {
    return this.request<{ workflow_id: string; project_id: string; status: string }>(
      "/projects/import",
      {
        method: "POST",
        body: JSON.stringify(params),
      }
    );
  }

  /**
   * Get the status of a brownfield import
   */
  async getImportStatus(projectId: string): Promise<{
    status: "analyzing" | "complete" | "error";
    files_indexed?: number;
    chunks_created?: number;
    symbols_found?: number;
    primary_lang?: string;
    detected_stack?: string[];
    error?: string;
    phase?: string;
  }> {
    return this.request<{
      status: "analyzing" | "complete" | "error";
      files_indexed?: number;
      chunks_created?: number;
      symbols_found?: number;
      primary_lang?: string;
      detected_stack?: string[];
      error?: string;
      phase?: string;
    }>(`/projects/${projectId}/import/status`);
  }

  // ==========================================================================
  // Utilities
  // ==========================================================================

  /**
   * Create a WebSocket connection for streaming events
   */
  createWebSocket(projectId: string): WebSocket {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    const url = `${protocol}//${host}/api/pods/${projectId}/stream`;
    return new WebSocket(url);
  }

  /**
   * Download files as a ZIP (client-side implementation)
   */
  async downloadAsZip(
    files: Record<string, string>,
    filename: string = "project.zip"
  ): Promise<void> {
    // Dynamically import JSZip only when needed
    const JSZip = (await import("jszip")).default;
    const zip = new JSZip();

    for (const [path, content] of Object.entries(files)) {
      zip.file(path, content);
    }

    const blob = await zip.generateAsync({ type: "blob" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  /**
   * Download a single file
   */
  downloadFile(filename: string, content: string): void {
    const blob = new Blob([content], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }
}

// Export a singleton instance
export const api = new ApiClient();

// Also export the class for testing
export { ApiClient };

// ==========================================================================
// Project helpers for localStorage management
// ==========================================================================

const PROJECTS_STORAGE_KEY = "sovereign-firm-projects";

export function loadProjectsFromStorage(): Project[] {
  if (typeof window === "undefined") return [];

  try {
    const stored = localStorage.getItem(PROJECTS_STORAGE_KEY);
    if (stored) {
      return JSON.parse(stored);
    }
  } catch (e) {
    console.error("Failed to load projects from storage:", e);
  }
  return [];
}

export function saveProjectsToStorage(projects: Project[]): void {
  if (typeof window === "undefined") return;

  try {
    localStorage.setItem(PROJECTS_STORAGE_KEY, JSON.stringify(projects));
  } catch (e) {
    console.error("Failed to save projects to storage:", e);
  }
}

export function createProjectFromConfig(
  projectId: string,
  workflowId: string,
  config: ProjectConfig
): Project {
  return {
    id: projectId,
    workflow_id: workflowId,
    name: config.project_name,
    phase: "INTAKE",
    status: "running",
    createdAt: new Date().toISOString(),
    config,
  };
}
