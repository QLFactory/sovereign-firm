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
} from "./types";

const API_BASE = "/api";

class ApiClient {
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${API_BASE}${endpoint}`;

    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({
        error: `Request failed with status ${response.status}`,
      }));
      throw new Error(error.error || "Request failed");
    }

    return response.json();
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
