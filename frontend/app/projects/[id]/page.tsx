"use client";

import { useEffect, useCallback, use } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import dynamic from "next/dynamic";
import { useAppStore } from "../../lib/store";
import { WebSocketProvider } from "../../lib/providers";
import {
  PHASE_COLORS,
  getAllFiles,
  getTotalFileCount,
  categorizeArtifacts,
} from "../../lib/api/types";
import type { LeftPanelTab, RightPanelTab, ConsultancyState } from "../../lib/api/types";

// Import Phase 4 components
import DAGProgressDashboard from "../../components/DAGProgressDashboard";
import AgentActivityPanel from "../../components/AgentActivityPanel";
import CIStatusPanel from "../../components/CIStatusPanel";
import ExecutionEventFeed from "../../components/ExecutionEventFeed";

// Dynamically import heavy components
const WebContainerPreview = dynamic(() => import("../../components/WebContainerPreview"), {
  ssr: false,
  loading: () => <LoadingPanel message="Loading WebContainer..." />,
});

const SandpackPreview = dynamic(() => import("../../components/SandpackPreview"), {
  ssr: false,
  loading: () => <LoadingPanel message="Loading Editor..." />,
});

// Loading Panel
function LoadingPanel({ message }: { message: string }) {
  return (
    <div className="flex items-center justify-center h-full bg-[var(--obsidian)]">
      <div className="text-center">
        <div className="spinner mx-auto mb-4" />
        <p className="text-[var(--silver)]">{message}</p>
      </div>
    </div>
  );
}

// Left Panel Tabs
const LEFT_TABS: { id: LeftPanelTab; label: string; icon: string }[] = [
  { id: "chat", label: "CHAT", icon: "" },
  { id: "spec", label: "SPEC", icon: "" },
  { id: "files", label: "FILES", icon: "" },
  { id: "tasks", label: "TASKS", icon: "📊" },
  { id: "agents", label: "AGENTS", icon: "🤖" },
];

// Right Panel Tabs
const RIGHT_TABS: { id: RightPanelTab; label: string; icon: string }[] = [
  { id: "preview", label: "PREVIEW", icon: "▶" },
  { id: "code", label: "CODE", icon: "📝" },
  { id: "terminal", label: "TERMINAL", icon: "⌨" },
  { id: "ci", label: "CI", icon: "🔧" },
  { id: "events", label: "EVENTS", icon: "📡" },
];

// Chat Message Component
function ChatMessage({ line, index }: { line: string; index: number }) {
  const isUser = line.startsWith("User:") || line.startsWith("You:");
  const isPM = line.startsWith("PM:");
  const isDev = line.startsWith("Dev:");
  const isQA = line.startsWith("QA:");

  let bgColor = "bg-[var(--graphite)]";
  let textColor = "text-[var(--pearl)]";
  let label = "";

  if (isUser) {
    bgColor = "bg-[var(--cyan-glow)]/10";
    textColor = "text-[var(--cyan-glow)]";
    label = "You";
  } else if (isPM) {
    bgColor = "bg-[var(--violet-glow)]/10";
    textColor = "text-[var(--violet-glow)]";
    label = "PM Agent";
  } else if (isDev) {
    bgColor = "bg-[var(--emerald-glow)]/10";
    textColor = "text-[var(--emerald-glow)]";
    label = "Dev Agent";
  } else if (isQA) {
    bgColor = "bg-[var(--amber-glow)]/10";
    textColor = "text-[var(--amber-glow)]";
    label = "QA Agent";
  }

  const content = line.replace(/^(User|You|PM|Dev|QA):/, "").trim();

  return (
    <div key={index} className={`p-3 rounded-lg mb-2 ${bgColor}`}>
      {label && <div className={`text-xs font-semibold mb-1 ${textColor}`}>{label}</div>}
      <div className="text-sm whitespace-pre-wrap text-[var(--pearl)]">{content}</div>
    </div>
  );
}

// File Tree Component
function FileTree({
  files,
  selectedFile,
  onSelectFile,
}: {
  files: Record<string, string>;
  selectedFile: string | null;
  onSelectFile: (file: string) => void;
}) {
  const fileList = Object.keys(files).sort((a, b) => {
    // Sort: src files first, then others
    const aInSrc = a.includes("/src/");
    const bInSrc = b.includes("/src/");
    if (aInSrc && !bInSrc) return -1;
    if (!aInSrc && bInSrc) return 1;
    return a.localeCompare(b);
  });

  if (fileList.length === 0) {
    return (
      <div className="text-center py-8">
        <div className="text-4xl mb-3">📁</div>
        <p className="text-[var(--silver)] text-sm">
          No files generated yet. Approve your spec to start coding.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-1">
      {fileList.map((filepath) => (
        <button
          key={filepath}
          onClick={() => onSelectFile(filepath)}
          className={`w-full text-left px-3 py-2 rounded text-sm font-mono transition-colors ${
            selectedFile === filepath
              ? "bg-[var(--cyan-glow)]/20 text-[var(--cyan-glow)]"
              : "bg-[var(--graphite)] text-[var(--silver)] hover:bg-[var(--slate)] hover:text-[var(--ivory)]"
          }`}
        >
          {filepath}
        </button>
      ))}
    </div>
  );
}

// Main Workspace Content
function WorkspaceContent({ projectId }: { projectId: string }) {
  const router = useRouter();

  // Store state
  const currentProject = useAppStore((state) => state.currentProject);
  const currentState = useAppStore((state) => state.currentState);
  const chatHistory = useAppStore((state) => state.chatHistory);
  const activeLeftTab = useAppStore((state) => state.activeLeftTab);
  const activeRightTab = useAppStore((state) => state.activeRightTab);
  const selectedFile = useAppStore((state) => state.selectedFile);
  const terminalOutput = useAppStore((state) => state.terminalOutput);
  const dagState = useAppStore((state) => state.dagState);
  const activeAgents = useAppStore((state) => state.activeAgents);
  const ciStages = useAppStore((state) => state.ciStages);
  const currentCIStage = useAppStore((state) => state.currentCIStage);
  const streamEvents = useAppStore((state) => state.streamEvents);
  const wsConnected = useAppStore((state) => state.wsConnected);
  const isLoading = useAppStore((state) => state.isLoading);
  const projects = useAppStore((state) => state.projects);

  // Store actions
  const selectProject = useAppStore((state) => state.selectProject);
  const fetchProjectState = useAppStore((state) => state.fetchProjectState);
  const setActiveLeftTab = useAppStore((state) => state.setActiveLeftTab);
  const setActiveRightTab = useAppStore((state) => state.setActiveRightTab);
  const setSelectedFile = useAppStore((state) => state.setSelectedFile);
  const sendMessage = useAppStore((state) => state.sendMessage);
  const addTerminalLine = useAppStore((state) => state.addTerminalLine);
  const setChat = useAppStore((state) => state.setChat);
  const updateCurrentState = useAppStore((state) => state.updateCurrentState);

  // Initialize project
  useEffect(() => {
    // Find project in list or create minimal one
    const project = projects.find((p) => p.id === projectId);
    if (project) {
      selectProject(project);
    } else {
      // Create a minimal project entry if we don't have it
      selectProject({
        id: projectId,
        name: `Project ${projectId.slice(0, 8)}`,
        phase: "INTAKE",
        status: "running",
        createdAt: new Date().toISOString(),
        config: {
          project_name: projectId,
          enable_full_stack: true,
          enable_deployment: true,
          enable_sre: true,
          preferred_frontend: "react",
          preferred_backend: "nodejs",
          preferred_database: "postgresql",
          preferred_cloud: "aws",
        },
      });
    }

    // Fetch initial state
    fetchProjectState(projectId);
  }, [projectId, projects, selectProject, fetchProjectState]);

  // Poll for state updates
  useEffect(() => {
    const pollInterval = setInterval(async () => {
      try {
        const response = await fetch(`/api/pods/${projectId}`);
        if (response.ok) {
          const state: ConsultancyState = await response.json();

          // Update state
          updateCurrentState(state);
          setChat(state.chat_history || "");
        }
      } catch {
        // Silently fail
      }
    }, 3000);

    return () => clearInterval(pollInterval);
  }, [projectId, updateCurrentState, setChat]);

  // Get files from current state
  const files = currentState ? getAllFiles(currentState) : {};
  const fileCount = currentState ? getTotalFileCount(currentState) : 0;
  const categories = currentState ? categorizeArtifacts(currentState) : [];

  // Chat input handler
  const handleSendMessage = useCallback(
    (message: string) => {
      if (message.trim()) {
        sendMessage(message);
      }
    },
    [sendMessage]
  );

  // File selection handler
  const handleSelectFile = useCallback(
    (file: string) => {
      setSelectedFile(file);
      setActiveRightTab("code");
    },
    [setSelectedFile, setActiveRightTab]
  );

  // Current phase
  const phase = currentProject?.phase || "INTAKE";
  const phaseColor = PHASE_COLORS[phase] || "badge";

  // Format chat history for display
  const chatLines = chatHistory.split("\n").filter(Boolean);

  return (
    <WebSocketProvider projectId={projectId}>
      <div className="flex h-screen bg-[var(--obsidian)] text-[var(--ivory)]">
        {/* Left Panel */}
        <div className="w-[400px] flex flex-col border-r border-[var(--steel)] bg-[var(--carbon)]">
          {/* Header */}
          <div className="p-4 border-b border-[var(--steel)]">
            <div className="flex justify-between items-start mb-3">
              <div>
                <h2 className="font-bold text-lg">
                  {currentProject?.name || "Loading..."}
                </h2>
                <div className="flex items-center gap-2 mt-1">
                  <span className={`badge ${phaseColor}`}>{phase}</span>
                  <span
                    className={`text-xs flex items-center gap-1 ${
                      wsConnected ? "text-[var(--emerald-glow)]" : "text-[var(--silver)]"
                    }`}
                  >
                    <span
                      className={`w-2 h-2 rounded-full ${
                        wsConnected
                          ? "bg-[var(--emerald-glow)] animate-pulse"
                          : "bg-[var(--steel)]"
                      }`}
                    />
                    {wsConnected ? "Connected" : "Disconnected"}
                  </span>
                </div>
              </div>
              <div className="flex gap-2">
                <Link
                  href={`/projects/${projectId}/artifacts`}
                  className="btn btn-ghost btn-sm"
                  title="View all artifacts"
                >
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                  >
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                    <polyline points="14 2 14 8 20 8" />
                  </svg>
                  {fileCount > 0 && (
                    <span className="ml-1 text-xs">{fileCount}</span>
                  )}
                </Link>
                <Link href="/projects" className="btn btn-ghost btn-sm" title="Back to projects">
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                  >
                    <path d="M19 12H5M12 19l-7-7 7-7" />
                  </svg>
                </Link>
              </div>
            </div>

            {/* Tabs */}
            <div className="flex gap-1 flex-wrap">
              {LEFT_TABS.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveLeftTab(tab.id)}
                  className={`px-3 py-1.5 rounded text-xs font-medium transition-colors flex items-center gap-1 ${
                    activeLeftTab === tab.id
                      ? "bg-[var(--cyan-glow)] text-[var(--void)]"
                      : "bg-[var(--graphite)] text-[var(--silver)] hover:text-[var(--ivory)]"
                  }`}
                >
                  {tab.icon && <span>{tab.icon}</span>}
                  {tab.label}
                  {tab.id === "files" && fileCount > 0 && (
                    <span className="ml-1 bg-[var(--steel)] px-1.5 rounded">
                      {fileCount}
                    </span>
                  )}
                  {tab.id === "tasks" && dagState && (
                    <span className="ml-1 bg-[var(--steel)] px-1.5 rounded">
                      {dagState.completed}/{dagState.total}
                    </span>
                  )}
                  {tab.id === "agents" && activeAgents.length > 0 && (
                    <span className="ml-1 bg-[var(--emerald-glow)] px-1.5 rounded text-[var(--void)] animate-pulse">
                      {activeAgents.length}
                    </span>
                  )}
                </button>
              ))}
            </div>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto p-4">
            {activeLeftTab === "chat" && (
              <div className="space-y-2">
                {chatLines.length > 0 ? (
                  chatLines.map((line, i) => (
                    <ChatMessage key={i} line={line} index={i} />
                  ))
                ) : (
                  <div className="text-center py-8">
                    <div className="text-4xl mb-3">💬</div>
                    <p className="text-[var(--silver)] text-sm">
                      Describe your app idea to the PM Agent...
                    </p>
                  </div>
                )}
              </div>
            )}

            {activeLeftTab === "spec" && (
              <div className="p-3 bg-[var(--graphite)] rounded-lg border border-[var(--steel)]">
                <h3 className="font-bold mb-2 text-[var(--silver)] text-sm">
                  Specification
                </h3>
                {currentState?.spec ? (
                  <pre className="text-sm whitespace-pre-wrap text-[var(--pearl)]">
                    {currentState.spec}
                  </pre>
                ) : (
                  <p className="text-[var(--steel)] italic text-sm">
                    No specification generated yet. Chat with the PM to define your app.
                  </p>
                )}
              </div>
            )}

            {activeLeftTab === "files" && (
              <FileTree
                files={files}
                selectedFile={selectedFile}
                onSelectFile={handleSelectFile}
              />
            )}

            {activeLeftTab === "tasks" && (
              <DAGProgressDashboard
                dagState={dagState}
                onTaskClick={(task) => {
                  addTerminalLine(`📋 Selected task: ${task.name} (${task.status})`);
                }}
              />
            )}

            {activeLeftTab === "agents" && (
              <AgentActivityPanel
                activeAgents={activeAgents}
                onAgentClick={(agent) => {
                  addTerminalLine(
                    `🤖 Selected agent: ${agent.agent_name} working on ${agent.task_name}`
                  );
                }}
              />
            )}
          </div>

          {/* Input */}
          <div className="p-4 border-t border-[var(--steel)]">
            <ChatInput onSend={handleSendMessage} phase={phase} />
          </div>
        </div>

        {/* Right Panel */}
        <div className="flex-1 flex flex-col min-w-0">
          {/* Tab Bar */}
          <div className="flex text-sm font-medium border-b border-[var(--steel)] bg-[var(--carbon)] overflow-x-auto">
            {RIGHT_TABS.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveRightTab(tab.id)}
                className={`px-4 py-3 transition-colors whitespace-nowrap flex items-center gap-1 ${
                  activeRightTab === tab.id
                    ? "bg-[var(--graphite)] text-[var(--ivory)] border-b-2 border-[var(--cyan-glow)]"
                    : "text-[var(--silver)] hover:text-[var(--pearl)] hover:bg-[var(--graphite)]/50"
                }`}
              >
                <span>{tab.icon}</span>
                {tab.label}
                {tab.id === "ci" && ciStages.length > 0 && (
                  <span
                    className={`ml-1 text-xs px-1.5 rounded ${
                      ciStages.some((s) => !s.success)
                        ? "bg-[var(--rose-glow)]"
                        : ciStages.length === 3
                        ? "bg-[var(--emerald-glow)]"
                        : "bg-[var(--amber-glow)]"
                    } text-[var(--void)]`}
                  >
                    {ciStages.filter((s) => s.success).length}/3
                  </span>
                )}
                {tab.id === "events" && streamEvents.length > 0 && (
                  <span className="ml-1 text-xs bg-[var(--steel)] px-1.5 rounded">
                    {streamEvents.length}
                  </span>
                )}
              </button>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 min-h-0 bg-[var(--obsidian)]">
            {activeRightTab === "preview" && (
              <WebContainerPreview files={files} onTerminalOutput={addTerminalLine} />
            )}

            {activeRightTab === "code" && (
              <SandpackPreview
                files={files}
                activeFile={selectedFile || undefined}
                showFileExplorer={true}
                showEditor={true}
                showConsole={false}
                theme="dark"
              />
            )}

            {activeRightTab === "terminal" && (
              <div className="h-full overflow-auto p-4 font-mono text-sm bg-[var(--obsidian)]">
                <div className="text-[var(--emerald-glow)] mb-2">
                  $ Sovereign Firm Terminal
                </div>
                {terminalOutput.length > 0 ? (
                  terminalOutput.map((line, i) => (
                    <div
                      key={i}
                      className={`whitespace-pre-wrap ${
                        line.includes("❌")
                          ? "text-[var(--rose-glow)]"
                          : line.includes("✅")
                          ? "text-[var(--emerald-glow)]"
                          : line.includes("📍")
                          ? "text-[var(--cyan-glow)]"
                          : line.includes("📝")
                          ? "text-[var(--amber-glow)]"
                          : "text-[var(--silver)]"
                      }`}
                    >
                      {line}
                    </div>
                  ))
                ) : (
                  <div className="text-[var(--steel)]">Waiting for activity...</div>
                )}
              </div>
            )}

            {activeRightTab === "ci" && (
              <CIStatusPanel
                stages={ciStages}
                currentStage={currentCIStage}
                onStageClick={(stage) => {
                  if (stage.output) {
                    addTerminalLine(`\n🔧 ${stage.stage} Output:\n${stage.output}`);
                  }
                }}
              />
            )}

            {activeRightTab === "events" && (
              <ExecutionEventFeed
                events={streamEvents}
                onEventClick={(event) => {
                  addTerminalLine(`📡 Event #${event.seq}: ${event.type}`);
                }}
              />
            )}
          </div>
        </div>
      </div>
    </WebSocketProvider>
  );
}

// Chat Input Component
function ChatInput({
  onSend,
  phase,
}: {
  onSend: (message: string) => void;
  phase: string;
}) {
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const input = form.elements.namedItem("message") as HTMLInputElement;
    if (input.value.trim()) {
      onSend(input.value);
      input.value = "";
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex gap-2">
      <input
        name="message"
        className="flex-1 bg-[var(--obsidian)] border border-[var(--steel)] rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:border-[var(--cyan-glow)] focus:ring-1 focus:ring-[var(--cyan-glow)]"
        placeholder={
          phase === "DISCOVERY" || phase === "INTAKE"
            ? "Describe your app or type /approve..."
            : "Send feedback or type /approve..."
        }
      />
      <button
        type="submit"
        className="btn btn-primary px-5"
      >
        Send
      </button>
    </form>
  );
}

// Page Component
export default function ProjectWorkspacePage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const resolvedParams = use(params);
  return <WorkspaceContent projectId={resolvedParams.id} />;
}
