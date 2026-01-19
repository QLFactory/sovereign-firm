/**
 * Zustand store for global state management.
 * Centralizes projects, current project state, UI state, and WebSocket events.
 */

import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { immer } from "zustand/middleware/immer";
import type {
  Project,
  ProjectConfig,
  ConsultancyState,
  Phase,
  StreamEvent,
  DAGState,
  ActiveAgent,
  CIStageResult,
  LeftPanelTab,
  RightPanelTab,
  User,
  LoginRequest,
  RegisterRequest,
  Invitation,
} from "./api/types";
import { api, loadProjectsFromStorage, saveProjectsToStorage, createProjectFromConfig } from "./api/client";

// =============================================================================
// Store Types
// =============================================================================

interface AppState {
  // Authentication
  user: User | null;
  isAuthenticated: boolean;
  isAuthLoading: boolean;
  authError: string | null;

  // Projects
  projects: Project[];
  currentProject: Project | null;
  currentState: ConsultancyState | null;

  // Team
  team: User[];

  // UI State
  activeLeftTab: LeftPanelTab;
  activeRightTab: RightPanelTab;
  selectedFile: string | null;
  selectedCategory: string | null;
  sidebarCollapsed: boolean;
  isCreateModalOpen: boolean;

  // WebSocket & Streaming
  wsConnected: boolean;
  streamEvents: StreamEvent[];
  terminalOutput: string[];

  // Multi-agent state
  dagState: DAGState | null;
  activeAgents: ActiveAgent[];
  ciStages: CIStageResult[];
  currentCIStage: "LINT" | "BUILD" | "TEST" | null;

  // Chat
  chatHistory: string;

  // Loading states
  isLoading: boolean;
  isCreating: boolean;
  error: string | null;
}

interface AppActions {
  // Auth actions
  login: (data: LoginRequest) => Promise<boolean>;
  register: (data: RegisterRequest) => Promise<boolean>;
  logout: () => Promise<void>;
  checkAuth: () => Promise<void>;
  clearAuthError: () => void;

  // Project actions
  loadProjects: () => void;
  createProject: (config: ProjectConfig) => Promise<string | null>;
  selectProject: (project: Project) => void;
  fetchProjectState: (id: string) => Promise<void>;
  updateProjectPhase: (id: string, phase: Phase) => void;
  sendMessage: (message: string) => Promise<void>;

  // Team actions
  loadTeam: () => Promise<void>;
  updateUserRole: (userId: string, role: string) => Promise<void>;
  inviteUser: (email: string, role?: string) => Promise<Invitation | null>;

  // UI actions
  setActiveLeftTab: (tab: LeftPanelTab) => void;
  setActiveRightTab: (tab: RightPanelTab) => void;
  setSelectedFile: (file: string | null) => void;
  setSelectedCategory: (category: string | null) => void;
  toggleSidebar: () => void;
  setCreateModalOpen: (open: boolean) => void;

  // WebSocket actions
  setWsConnected: (connected: boolean) => void;
  addStreamEvent: (event: StreamEvent) => void;
  clearStreamEvents: () => void;

  // Terminal actions
  addTerminalLine: (line: string) => void;
  clearTerminal: () => void;

  // Multi-agent actions
  setDagState: (state: DAGState | null) => void;
  addActiveAgent: (agent: ActiveAgent) => void;
  removeActiveAgent: (taskId: string) => void;
  updateActiveAgent: (taskId: string, updates: Partial<ActiveAgent>) => void;
  addCIStage: (result: CIStageResult) => void;
  setCurrentCIStage: (stage: "LINT" | "BUILD" | "TEST" | null) => void;

  // Chat actions
  appendChat: (line: string) => void;
  setChat: (history: string) => void;

  // State updates from polling/streaming
  updateCurrentState: (state: Partial<ConsultancyState>) => void;

  // Error handling
  setError: (error: string | null) => void;
  clearError: () => void;

  // Reset
  reset: () => void;
}

type AppStore = AppState & AppActions;

// =============================================================================
// Initial State
// =============================================================================

const initialState: AppState = {
  // Authentication
  user: null,
  isAuthenticated: false,
  isAuthLoading: true, // Start true to check auth on load
  authError: null,

  // Projects
  projects: [],
  currentProject: null,
  currentState: null,

  // Team
  team: [],

  // UI State
  activeLeftTab: "chat",
  activeRightTab: "preview",
  selectedFile: null,
  selectedCategory: null,
  sidebarCollapsed: false,
  isCreateModalOpen: false,

  // WebSocket & Streaming
  wsConnected: false,
  streamEvents: [],
  terminalOutput: [],

  // Multi-agent state
  dagState: null,
  activeAgents: [],
  ciStages: [],
  currentCIStage: null,

  // Chat
  chatHistory: "",

  // Loading states
  isLoading: false,
  isCreating: false,
  error: null,
};

// =============================================================================
// Store Implementation
// =============================================================================

export const useAppStore = create<AppStore>()(
  persist(
    immer((set, get) => ({
      ...initialState,

      // =========================================================================
      // Auth Actions
      // =========================================================================

      login: async (data: LoginRequest) => {
        set((state) => {
          state.isAuthLoading = true;
          state.authError = null;
        });

        try {
          const response = await api.login(data);
          set((state) => {
            state.user = response.user;
            state.isAuthenticated = true;
            state.isAuthLoading = false;
            state.authError = null;
          });
          return true;
        } catch (error) {
          set((state) => {
            state.isAuthLoading = false;
            state.authError = error instanceof Error ? error.message : "Login failed";
          });
          return false;
        }
      },

      register: async (data: RegisterRequest) => {
        set((state) => {
          state.isAuthLoading = true;
          state.authError = null;
        });

        try {
          const response = await api.register(data);
          set((state) => {
            state.user = response.user;
            state.isAuthenticated = true;
            state.isAuthLoading = false;
            state.authError = null;
          });
          return true;
        } catch (error) {
          set((state) => {
            state.isAuthLoading = false;
            state.authError = error instanceof Error ? error.message : "Registration failed";
          });
          return false;
        }
      },

      logout: async () => {
        try {
          await api.logout();
        } finally {
          set((state) => {
            state.user = null;
            state.isAuthenticated = false;
            state.projects = [];
            state.currentProject = null;
            state.currentState = null;
          });
        }
      },

      checkAuth: async () => {
        // Check if we have a token
        if (!api.isAuthenticated()) {
          set((state) => {
            state.isAuthLoading = false;
            state.isAuthenticated = false;
            state.user = null;
          });
          return;
        }

        try {
          const user = await api.getCurrentUser();
          set((state) => {
            state.user = user;
            state.isAuthenticated = true;
            state.isAuthLoading = false;
          });
        } catch {
          // Token invalid or expired
          api.clearTokens();
          set((state) => {
            state.user = null;
            state.isAuthenticated = false;
            state.isAuthLoading = false;
          });
        }
      },

      clearAuthError: () => {
        set((state) => {
          state.authError = null;
        });
      },

      // =========================================================================
      // Project Actions
      // =========================================================================

      loadProjects: () => {
        const projects = loadProjectsFromStorage();
        set((state) => {
          state.projects = projects;
        });
      },

      createProject: async (config: ProjectConfig) => {
        set((state) => {
          state.isCreating = true;
          state.error = null;
        });

        try {
          const response = await api.createProject(config);
          const newProject = createProjectFromConfig(response.id || response.workflow_id, response.workflow_id, config);

          set((state) => {
            state.projects.unshift(newProject);
            state.currentProject = newProject;
            state.isCreating = false;
            state.isCreateModalOpen = false;
            // Reset state for new project
            state.currentState = null;
            state.chatHistory = "";
            state.streamEvents = [];
            state.terminalOutput = [];
            state.dagState = null;
            state.activeAgents = [];
            state.ciStages = [];
          });

          // Save to localStorage
          saveProjectsToStorage(get().projects);

          return response.workflow_id;
        } catch (error) {
          set((state) => {
            state.isCreating = false;
            state.error = error instanceof Error ? error.message : "Failed to create project";
          });
          return null;
        }
      },

      selectProject: (project: Project) => {
        set((state) => {
          state.currentProject = project;
          // Reset state when switching projects
          state.currentState = null;
          state.chatHistory = "";
          state.streamEvents = [];
          state.dagState = null;
          state.activeAgents = [];
          state.ciStages = [];
          state.selectedFile = null;
        });
      },

      fetchProjectState: async (id: string) => {
        set((state) => {
          state.isLoading = true;
          state.error = null;
        });

        try {
          const projectState = await api.getProject(id);

          set((state) => {
            state.isLoading = false;
            // Use updateCurrentState logic via set
            if (state.currentState) {
              Object.assign(state.currentState, projectState);
            } else {
              state.currentState = projectState;
            }

            // Propagation
            state.chatHistory = projectState.chat_history || "";
            state.dagState = projectState.dag || null;
            state.activeAgents = projectState.agents || [];
            state.ciStages = projectState.ci_status?.stages || [];

            // Update project phase in the list
            const projectIndex = state.projects.findIndex((p) => p.id === id);
            if (projectIndex >= 0) {
              state.projects[projectIndex].phase = projectState.phase;
            }
            if (state.currentProject?.id === id) {
              state.currentProject.phase = projectState.phase;
            }
          });

          // Save updated projects to localStorage
          saveProjectsToStorage(get().projects);
        } catch (error) {
          set((state) => {
            state.isLoading = false;
            state.error = error instanceof Error ? error.message : "Failed to fetch project";
          });
        }
      },

      updateProjectPhase: (id: string, phase: Phase) => {
        set((state) => {
          const projectIndex = state.projects.findIndex((p) => p.id === id);
          if (projectIndex >= 0) {
            state.projects[projectIndex].phase = phase;
          }
          if (state.currentProject?.id === id) {
            state.currentProject.phase = phase;
          }
        });
        saveProjectsToStorage(get().projects);
      },

      sendMessage: async (message: string) => {
        const { currentProject, appendChat, addTerminalLine } = get();
        if (!currentProject) return;

        // Optimistically add to chat
        appendChat(`You: ${message}`);
        addTerminalLine(`📤 Sent: ${message}`);

        try {
          await api.sendMessage(currentProject.id, message);
        } catch (error) {
          addTerminalLine(`❌ Failed to send message: ${error}`);
        }
      },

      // =========================================================================
      // Team Actions
      // =========================================================================

      loadTeam: async () => {
        set((state) => {
          state.isLoading = true;
        });
        try {
          const team = await api.listUsers();
          set((state) => {
            state.team = team;
            state.isLoading = false;
          });
        } catch (error) {
          set((state) => {
            state.isLoading = false;
            state.error = error instanceof Error ? error.message : "Failed to load team";
          });
        }
      },

      updateUserRole: async (userId: string, role: string) => {
        try {
          await api.updateUserRole(userId, role);
          set((state) => {
            const index = state.team.findIndex((u) => u.id === userId);
            if (index >= 0) {
              state.team[index].role = role;
            }
          });
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : "Failed to update role";
          });
        }
      },

      inviteUser: async (email: string, role: string = "member") => {
        try {
          const invitation = await api.inviteUser(email, role);
          return invitation;
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : "Failed to invite user";
          });
          return null;
        }
      },

      // =========================================================================
      // UI Actions
      // =========================================================================

      setActiveLeftTab: (tab: LeftPanelTab) => {
        set((state) => {
          state.activeLeftTab = tab;
        });
      },

      setActiveRightTab: (tab: RightPanelTab) => {
        set((state) => {
          state.activeRightTab = tab;
        });
      },

      setSelectedFile: (file: string | null) => {
        set((state) => {
          state.selectedFile = file;
        });
      },

      setSelectedCategory: (category: string | null) => {
        set((state) => {
          state.selectedCategory = category;
        });
      },

      toggleSidebar: () => {
        set((state) => {
          state.sidebarCollapsed = !state.sidebarCollapsed;
        });
      },

      setCreateModalOpen: (open: boolean) => {
        set((state) => {
          state.isCreateModalOpen = open;
        });
      },

      // =========================================================================
      // WebSocket Actions
      // =========================================================================

      setWsConnected: (connected: boolean) => {
        set((state) => {
          state.wsConnected = connected;
        });
      },

      addStreamEvent: (event: StreamEvent) => {
        set((state) => {
          // Keep last 200 events
          if (state.streamEvents.length >= 200) {
            state.streamEvents.shift();
          }
          state.streamEvents.push(event);
        });
      },

      clearStreamEvents: () => {
        set((state) => {
          state.streamEvents = [];
        });
      },

      // =========================================================================
      // Terminal Actions
      // =========================================================================

      addTerminalLine: (line: string) => {
        set((state) => {
          // Keep last 100 lines
          if (state.terminalOutput.length >= 100) {
            state.terminalOutput.shift();
          }
          state.terminalOutput.push(line);
        });
      },

      clearTerminal: () => {
        set((state) => {
          state.terminalOutput = [];
        });
      },

      // =========================================================================
      // Multi-agent Actions
      // =========================================================================

      setDagState: (dagState: DAGState | null) => {
        set((state) => {
          state.dagState = dagState;
        });
      },

      addActiveAgent: (agent: ActiveAgent) => {
        set((state) => {
          const exists = state.activeAgents.some((a) => a.task_id === agent.task_id);
          if (!exists) {
            state.activeAgents.push(agent);
          }
        });
      },

      removeActiveAgent: (taskId: string) => {
        set((state) => {
          state.activeAgents = state.activeAgents.filter((a) => a.task_id !== taskId);
        });
      },

      updateActiveAgent: (taskId: string, updates: Partial<ActiveAgent>) => {
        set((state) => {
          const index = state.activeAgents.findIndex((a) => a.task_id === taskId);
          if (index >= 0) {
            Object.assign(state.activeAgents[index], updates);
          }
        });
      },

      addCIStage: (result: CIStageResult) => {
        set((state) => {
          const index = state.ciStages.findIndex((s) => s.stage === result.stage);
          if (index >= 0) {
            state.ciStages[index] = result;
          } else {
            state.ciStages.push(result);
          }
        });
      },

      setCurrentCIStage: (stage: "LINT" | "BUILD" | "TEST" | null) => {
        set((state) => {
          state.currentCIStage = stage;
        });
      },

      // =========================================================================
      // Chat Actions
      // =========================================================================

      appendChat: (line: string) => {
        set((state) => {
          state.chatHistory = state.chatHistory
            ? `${state.chatHistory}\n${line}`
            : line;
        });
      },

      setChat: (history: string) => {
        set((state) => {
          state.chatHistory = history;
        });
      },

      // =========================================================================
      // State Updates
      // =========================================================================

      updateCurrentState: (updates: Partial<ConsultancyState>) => {
        set((state) => {
          if (state.currentState) {
            Object.assign(state.currentState, updates);
          } else {
            state.currentState = updates as ConsultancyState;
          }

          // Propagate critical fields to top-level store
          if (updates.phase) {
            // Update current project phase
            if (state.currentProject) {
              state.currentProject.phase = updates.phase;
            }
            // Update in projects list
            const idx = state.projects.findIndex((p) => p.id === state.currentProject?.id);
            if (idx >= 0) {
              state.projects[idx].phase = updates.phase;
            }
          }

          if (updates.dag) state.dagState = updates.dag;
          if (updates.agents) state.activeAgents = updates.agents;
          if (updates.ci_status) state.ciStages = updates.ci_status.stages;
          if (updates.chat_history) state.chatHistory = updates.chat_history;
        });
      },

      // =========================================================================
      // Error Handling
      // =========================================================================

      setError: (error: string | null) => {
        set((state) => {
          state.error = error;
        });
      },

      clearError: () => {
        set((state) => {
          state.error = null;
        });
      },

      // =========================================================================
      // Reset
      // =========================================================================

      reset: () => {
        set((state) => {
          Object.assign(state, initialState);
        });
      },
    })),
    {
      name: "sovereign-firm-store",
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        // Only persist these fields
        projects: state.projects,
        sidebarCollapsed: state.sidebarCollapsed,
      }),
    }
  )
);

// =============================================================================
// Selectors
// =============================================================================

// Auth selectors
export const selectUser = (state: AppStore) => state.user;
export const selectIsAuthenticated = (state: AppStore) => state.isAuthenticated;
export const selectIsAuthLoading = (state: AppStore) => state.isAuthLoading;
export const selectAuthError = (state: AppStore) => state.authError;

// Project selectors
export const selectProjects = (state: AppStore) => state.projects;
export const selectCurrentProject = (state: AppStore) => state.currentProject;
export const selectCurrentState = (state: AppStore) => state.currentState;
export const selectActiveLeftTab = (state: AppStore) => state.activeLeftTab;
export const selectActiveRightTab = (state: AppStore) => state.activeRightTab;
export const selectSelectedFile = (state: AppStore) => state.selectedFile;
export const selectIsLoading = (state: AppStore) => state.isLoading;
export const selectIsCreating = (state: AppStore) => state.isCreating;
export const selectError = (state: AppStore) => state.error;
export const selectWsConnected = (state: AppStore) => state.wsConnected;
export const selectStreamEvents = (state: AppStore) => state.streamEvents;
export const selectTerminalOutput = (state: AppStore) => state.terminalOutput;
export const selectDagState = (state: AppStore) => state.dagState;
export const selectActiveAgents = (state: AppStore) => state.activeAgents;
export const selectCiStages = (state: AppStore) => state.ciStages;
export const selectChatHistory = (state: AppStore) => state.chatHistory;
