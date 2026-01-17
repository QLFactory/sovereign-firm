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

  /**
   * Get current user info
   */
  async getCurrentUser(): Promise<User> {
    return this.request<User>("/auth/me");
  }

  // ==========================================================================
  // Projects
  // ==========================================================================

  /**
   * Create a new project/pod
   */
  async createProject(config: CreateProjectRequest): Promise<CreateProjectResponse> {
    return this.request<CreateProjectResponse>("/pods", {
      method: "POST",
      body: JSON.stringify(config),
    });
  }

  /**
   * Get project state by ID
   */
  async getProject(id: string): Promise<ConsultancyState> {
    return this.request<ConsultancyState>(`/pods/${id}`);
  }

  /**
   * List all projects (if backend supports it)
   */
  async listProjects(): Promise<Project[]> {
    try {
      return this.request<Project[]>("/pods");
    } catch {
      // If backend doesn't support listing, return empty array
      return [];
    }
  }

  /**
   * Send a message to a project (e.g., /approve, feedback)
   */
  async sendMessage(
    projectId: string,
    message: string
  ): Promise<SendMessageResponse> {
    const request: SendMessageRequest = { message };
    return this.request<SendMessageResponse>(`/pods/${projectId}`, {
      method: "POST",
      body: JSON.stringify(request),
    });
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
  workflowId: string,
  config: ProjectConfig
): Project {
  return {
    id: workflowId,
    name: config.project_name,
    phase: "INTAKE",
    status: "running",
    createdAt: new Date().toISOString(),
    config,
  };
}
