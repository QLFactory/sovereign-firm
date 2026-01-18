import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useAppStore } from "../store";
import type { Project, Phase } from "../api/types";

// Mock the API client
vi.mock("../api/client", () => ({
  api: {
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
    isAuthenticated: vi.fn(() => false),
    getCurrentUser: vi.fn(),
    clearTokens: vi.fn(),
    createProject: vi.fn(),
    getProject: vi.fn(),
    sendMessage: vi.fn(),
    listProjects: vi.fn(),
  },
  loadProjectsFromStorage: vi.fn(() => []),
  saveProjectsToStorage: vi.fn(),
  createProjectFromConfig: vi.fn((workflowId, config) => ({
    id: workflowId,
    name: config.project_name,
    phase: "INTAKE" as Phase,
    status: "running",
    createdAt: new Date().toISOString(),
    config,
  })),
}));

describe("useAppStore", () => {
  beforeEach(() => {
    // Reset store state before each test
    const store = useAppStore.getState();
    store.reset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("initial state", () => {
    it("has correct initial values", () => {
      const { result } = renderHook(() => useAppStore());

      expect(result.current.user).toBeNull();
      expect(result.current.isAuthenticated).toBe(false);
      expect(result.current.projects).toEqual([]);
      expect(result.current.currentProject).toBeNull();
      expect(result.current.currentState).toBeNull();
      expect(result.current.activeLeftTab).toBe("chat");
      expect(result.current.activeRightTab).toBe("preview");
      expect(result.current.wsConnected).toBe(false);
      expect(result.current.error).toBeNull();
    });
  });

  describe("UI actions", () => {
    it("setActiveLeftTab changes left tab", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.setActiveLeftTab("artifacts");
      });

      expect(result.current.activeLeftTab).toBe("artifacts");
    });

    it("setActiveRightTab changes right tab", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.setActiveRightTab("code");
      });

      expect(result.current.activeRightTab).toBe("code");
    });

    it("setSelectedFile updates selected file", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.setSelectedFile("/src/App.jsx");
      });

      expect(result.current.selectedFile).toBe("/src/App.jsx");
    });

    it("toggleSidebar toggles sidebar state", () => {
      const { result } = renderHook(() => useAppStore());

      expect(result.current.sidebarCollapsed).toBe(false);

      act(() => {
        result.current.toggleSidebar();
      });

      expect(result.current.sidebarCollapsed).toBe(true);

      act(() => {
        result.current.toggleSidebar();
      });

      expect(result.current.sidebarCollapsed).toBe(false);
    });

    it("setCreateModalOpen controls modal state", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.setCreateModalOpen(true);
      });

      expect(result.current.isCreateModalOpen).toBe(true);

      act(() => {
        result.current.setCreateModalOpen(false);
      });

      expect(result.current.isCreateModalOpen).toBe(false);
    });
  });

  describe("WebSocket actions", () => {
    it("setWsConnected updates connection status", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.setWsConnected(true);
      });

      expect(result.current.wsConnected).toBe(true);
    });

    it("addStreamEvent adds event to stream", () => {
      const { result } = renderHook(() => useAppStore());

      const event = {
        type: "phase_change" as const,
        phase: "DEVELOPMENT" as Phase,
        timestamp: new Date().toISOString(),
      };

      act(() => {
        result.current.addStreamEvent(event);
      });

      expect(result.current.streamEvents).toContainEqual(event);
    });

    it("addStreamEvent limits events to 200", () => {
      const { result } = renderHook(() => useAppStore());

      // Add 210 events
      act(() => {
        for (let i = 0; i < 210; i++) {
          result.current.addStreamEvent({
            type: "chat" as const,
            agent: "test",
            message: `Event ${i}`,
            timestamp: new Date().toISOString(),
          });
        }
      });

      expect(result.current.streamEvents.length).toBeLessThanOrEqual(200);
    });

    it("clearStreamEvents removes all events", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.addStreamEvent({
          type: "chat" as const,
          agent: "test",
          message: "test",
          timestamp: new Date().toISOString(),
        });
        result.current.clearStreamEvents();
      });

      expect(result.current.streamEvents).toHaveLength(0);
    });
  });

  describe("terminal actions", () => {
    it("addTerminalLine adds line to output", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.addTerminalLine("Test output");
      });

      expect(result.current.terminalOutput).toContain("Test output");
    });

    it("addTerminalLine limits to 100 lines", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        for (let i = 0; i < 110; i++) {
          result.current.addTerminalLine(`Line ${i}`);
        }
      });

      expect(result.current.terminalOutput.length).toBeLessThanOrEqual(100);
    });

    it("clearTerminal removes all output", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.addTerminalLine("Test");
        result.current.clearTerminal();
      });

      expect(result.current.terminalOutput).toHaveLength(0);
    });
  });

  describe("multi-agent actions", () => {
    it("addActiveAgent adds agent to list", () => {
      const { result } = renderHook(() => useAppStore());

      const agent = {
        task_id: "task-1",
        agent_type: "DEV",
        status: "running" as const,
        started_at: new Date().toISOString(),
      };

      act(() => {
        result.current.addActiveAgent(agent);
      });

      expect(result.current.activeAgents).toContainEqual(agent);
    });

    it("addActiveAgent prevents duplicates", () => {
      const { result } = renderHook(() => useAppStore());

      const agent = {
        task_id: "task-1",
        agent_type: "DEV",
        status: "running" as const,
        started_at: new Date().toISOString(),
      };

      act(() => {
        result.current.addActiveAgent(agent);
        result.current.addActiveAgent(agent);
      });

      expect(result.current.activeAgents).toHaveLength(1);
    });

    it("removeActiveAgent removes agent by taskId", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.addActiveAgent({
          task_id: "task-1",
          agent_type: "DEV",
          status: "running" as const,
          started_at: new Date().toISOString(),
        });
        result.current.removeActiveAgent("task-1");
      });

      expect(result.current.activeAgents).toHaveLength(0);
    });

    it("updateActiveAgent updates agent properties", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.addActiveAgent({
          task_id: "task-1",
          agent_type: "DEV",
          status: "running" as const,
          started_at: new Date().toISOString(),
        });
        result.current.updateActiveAgent("task-1", { status: "completed" as const });
      });

      expect(result.current.activeAgents[0].status).toBe("completed");
    });

    it("setDagState updates DAG state", () => {
      const { result } = renderHook(() => useAppStore());

      const dagState = {
        nodes: [{ id: "node-1", status: "running" }],
        edges: [],
      };

      act(() => {
        result.current.setDagState(dagState as any);
      });

      expect(result.current.dagState).toEqual(dagState);
    });
  });

  describe("CI stage actions", () => {
    it("addCIStage adds new stage result", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.addCIStage({
          stage: "LINT",
          status: "passed",
          duration: 1000,
        });
      });

      expect(result.current.ciStages).toContainEqual({
        stage: "LINT",
        status: "passed",
        duration: 1000,
      });
    });

    it("addCIStage updates existing stage", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.addCIStage({
          stage: "LINT",
          status: "running",
          duration: 0,
        });
        result.current.addCIStage({
          stage: "LINT",
          status: "passed",
          duration: 1000,
        });
      });

      expect(result.current.ciStages).toHaveLength(1);
      expect(result.current.ciStages[0].status).toBe("passed");
    });

    it("setCurrentCIStage updates current stage", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.setCurrentCIStage("BUILD");
      });

      expect(result.current.currentCIStage).toBe("BUILD");
    });
  });

  describe("chat actions", () => {
    it("appendChat adds line to chat history", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.appendChat("User: Hello");
      });

      expect(result.current.chatHistory).toBe("User: Hello");
    });

    it("appendChat appends with newline", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.appendChat("User: Hello");
        result.current.appendChat("PM: Hi there");
      });

      expect(result.current.chatHistory).toBe("User: Hello\nPM: Hi there");
    });

    it("setChat replaces entire history", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.appendChat("Old message");
        result.current.setChat("New history");
      });

      expect(result.current.chatHistory).toBe("New history");
    });
  });

  describe("error handling", () => {
    it("setError sets error message", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.setError("Something went wrong");
      });

      expect(result.current.error).toBe("Something went wrong");
    });

    it("clearError removes error", () => {
      const { result } = renderHook(() => useAppStore());

      act(() => {
        result.current.setError("Error");
        result.current.clearError();
      });

      expect(result.current.error).toBeNull();
    });
  });

  describe("project actions", () => {
    it("selectProject updates current project", () => {
      const { result } = renderHook(() => useAppStore());

      const project: Project = {
        id: "project-1",
        name: "Test Project",
        phase: "DEVELOPMENT" as Phase,
        status: "running",
        createdAt: new Date().toISOString(),
      };

      act(() => {
        result.current.selectProject(project);
      });

      expect(result.current.currentProject).toEqual(project);
    });

    it("selectProject resets related state", () => {
      const { result } = renderHook(() => useAppStore());

      // Set some state first
      act(() => {
        result.current.appendChat("Old chat");
        result.current.setSelectedFile("/old/file.js");
        result.current.addStreamEvent({
          type: "chat" as const,
          agent: "test",
          message: "test",
          timestamp: new Date().toISOString(),
        });
      });

      // Select new project
      act(() => {
        result.current.selectProject({
          id: "project-2",
          name: "New Project",
          phase: "INTAKE" as Phase,
          status: "running",
          createdAt: new Date().toISOString(),
        });
      });

      expect(result.current.chatHistory).toBe("");
      expect(result.current.selectedFile).toBeNull();
      expect(result.current.currentState).toBeNull();
      expect(result.current.streamEvents).toHaveLength(0);
    });

    it("updateProjectPhase updates project in list", () => {
      const { result } = renderHook(() => useAppStore());

      const project: Project = {
        id: "project-1",
        name: "Test Project",
        phase: "INTAKE" as Phase,
        status: "running",
        createdAt: new Date().toISOString(),
      };

      act(() => {
        result.current.selectProject(project);
        // Manually add to projects list
        useAppStore.setState({
          projects: [project],
        });
      });

      act(() => {
        result.current.updateProjectPhase("project-1", "DEVELOPMENT" as Phase);
      });

      expect(result.current.projects[0].phase).toBe("DEVELOPMENT");
      expect(result.current.currentProject?.phase).toBe("DEVELOPMENT");
    });
  });

  describe("reset", () => {
    it("resets all state to initial values", () => {
      const { result } = renderHook(() => useAppStore());

      // Modify state
      act(() => {
        result.current.setActiveLeftTab("artifacts");
        result.current.appendChat("Chat history");
        result.current.setWsConnected(true);
        result.current.setError("An error");
      });

      // Reset
      act(() => {
        result.current.reset();
      });

      expect(result.current.activeLeftTab).toBe("chat");
      expect(result.current.chatHistory).toBe("");
      expect(result.current.wsConnected).toBe(false);
      expect(result.current.error).toBeNull();
    });
  });

  describe("selectors", () => {
    it("exports selector functions", async () => {
      const { selectUser, selectIsAuthenticated, selectProjects } = await import("../store");

      expect(typeof selectUser).toBe("function");
      expect(typeof selectIsAuthenticated).toBe("function");
      expect(typeof selectProjects).toBe("function");
    });
  });
});
