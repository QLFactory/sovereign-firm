"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import dynamic from "next/dynamic";

// localStorage key for persisting workflow ID
const WORKFLOW_STORAGE_KEY = "sovereign-firm-workflow-id";
import {
  useStreaming,
  StreamEvent,
  isPhaseChangeEvent,
  isCodeChunkEvent,
  isChatMessageEvent,
  isFileStartEvent,
  isErrorEvent,
  // Phase 4 event type guards
  isDAGUpdateEvent,
  isTaskStartedEvent,
  isTaskCompletedEvent,
  isTaskFailedEvent,
  isAgentAssignedEvent,
  isCIStageEvent,
  isExecutionEvent,
  // Phase 4 types
  DAGState,
  ActiveAgent,
  CIStageResult,
} from "../hooks/useStreaming";

// Phase 4 components
import DAGProgressDashboard from "./DAGProgressDashboard";
import AgentActivityPanel from "./AgentActivityPanel";
import CIStatusPanel from "./CIStatusPanel";
import ExecutionEventFeed from "./ExecutionEventFeed";

// Dynamically import WebContainerPreview to avoid SSR issues
const WebContainerPreview = dynamic(() => import("./WebContainerPreview"), {
  ssr: false,
  loading: () => (
    <div className="flex items-center justify-center h-full bg-zinc-950">
      <div className="text-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
        <p className="text-zinc-400">Loading WebContainer...</p>
      </div>
    </div>
  ),
});

// Dynamically import SandpackPreview for CODE tab (editor view)
const SandpackPreview = dynamic(() => import("./SandpackPreview"), {
  ssr: false,
  loading: () => (
    <div className="flex items-center justify-center h-full bg-zinc-950">
      <div className="text-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
        <p className="text-zinc-400">Loading Editor...</p>
      </div>
    </div>
  ),
});

// Phase badge colors
const phaseColors: Record<string, string> = {
  DISCOVERY: "bg-blue-600",
  IMPLEMENTATION: "bg-yellow-600",
  REVIEW: "bg-purple-600",
  REFINEMENT: "bg-orange-600",
  DONE: "bg-green-600",
  DELIVERED: "bg-green-600",
};

interface PodConsoleProps {
  workflowId?: string;
}

export default function PodConsole({ workflowId: propWorkflowId }: PodConsoleProps) {
  const router = useRouter();
  const searchParams = useSearchParams();

  // Workflow state - initialize from URL or localStorage
  const [workflowID, setWorkflowID] = useState<string | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [chatHistory, setChatHistory] = useState<string>("");
  const [input, setInput] = useState("");
  const [status, setStatus] = useState("Initializing...");
  const [activeTab, setActiveTab] = useState<"CHAT" | "SPEC" | "FILES" | "TASKS" | "AGENTS">("CHAT");
  const [rightTab, setRightTab] = useState<"PREVIEW" | "CODE" | "TERMINAL" | "CI" | "EVENTS">("PREVIEW");

  // Phase 4: Multi-Agent Coordination State
  const [dagState, setDagState] = useState<DAGState | null>(null);
  const [activeAgents, setActiveAgents] = useState<ActiveAgent[]>([]);
  const [ciStages, setCIStages] = useState<CIStageResult[]>([]);
  const [currentCIStage, setCurrentCIStage] = useState<"LINT" | "BUILD" | "TEST" | null>(null);
  const [streamEvents, setStreamEvents] = useState<StreamEvent[]>([]);
  const [projectSpec, setProjectSpec] = useState("");
  const [currentPhase, setCurrentPhase] = useState("DISCOVERY");

  // Code files state
  const [files, setFiles] = useState<Record<string, string>>({});
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [terminalOutput, setTerminalOutput] = useState<string[]>([]);

  // Streaming state
  const fileBuffersRef = useRef<Record<string, string[]>>({});
  const chatEndRef = useRef<HTMLDivElement>(null);

  // Initialize workflow ID from prop, URL param, or localStorage
  useEffect(() => {
    if (isInitialized) return;

    const urlWorkflowId = searchParams.get("workflow");
    const storedWorkflowId = typeof window !== "undefined"
      ? localStorage.getItem(WORKFLOW_STORAGE_KEY)
      : null;

    // Priority: 1) Prop, 2) URL param, 3) localStorage
    if (propWorkflowId) {
      setWorkflowID(propWorkflowId);
      if (typeof window !== "undefined") {
        localStorage.setItem(WORKFLOW_STORAGE_KEY, propWorkflowId);
      }
      setStatus("Reconnecting to workflow...");
    } else if (urlWorkflowId) {
      setWorkflowID(urlWorkflowId);
      if (typeof window !== "undefined") {
        localStorage.setItem(WORKFLOW_STORAGE_KEY, urlWorkflowId);
      }
      setStatus("Reconnecting to workflow...");
    } else if (storedWorkflowId) {
      setWorkflowID(storedWorkflowId);
      setStatus("Reconnecting to workflow...");
    }

    setIsInitialized(true);
  }, [searchParams, isInitialized, propWorkflowId]);

  // Persist workflow ID to localStorage and URL when it changes
  useEffect(() => {
    if (!workflowID || !isInitialized) return;

    // Save to localStorage
    if (typeof window !== "undefined") {
      localStorage.setItem(WORKFLOW_STORAGE_KEY, workflowID);
    }

    // Update URL without navigation
    const currentUrl = new URL(window.location.href);
    if (currentUrl.searchParams.get("workflow") !== workflowID) {
      currentUrl.searchParams.set("workflow", workflowID);
      router.replace(currentUrl.pathname + currentUrl.search, { scroll: false });
    }
  }, [workflowID, isInitialized, router]);

  // Terminal output helper
  const addTerminalLine = useCallback((line: string) => {
    setTerminalOutput((prev) => [...prev.slice(-100), line]); // Keep last 100 lines
  }, []);

  // Handle streaming events
  const handleStreamEvent = useCallback((event: StreamEvent) => {
    console.log("Stream event:", event.type, event.payload);

    if (isPhaseChangeEvent(event)) {
      setCurrentPhase(event.payload.current_phase);
      setStatus(event.payload.message || `Phase: ${event.payload.current_phase}`);
      addTerminalLine(`\n📍 Phase changed: ${event.payload.previous_phase} → ${event.payload.current_phase}`);
    }

    if (isChatMessageEvent(event)) {
      const { role, agent, content } = event.payload;
      const prefix = role === "user" ? "You" : agent || "Agent";
      setChatHistory((prev) => prev + `\n${prefix}: ${content}`);
    }

    if (isFileStartEvent(event)) {
      const { file_path, language } = event.payload;
      addTerminalLine(`\n📝 Generating: ${file_path} (${language})`);
      fileBuffersRef.current[file_path] = [];
    }

    if (isCodeChunkEvent(event)) {
      const { file_path, content, is_complete } = event.payload;

      // Accumulate chunks
      if (!fileBuffersRef.current[file_path]) {
        fileBuffersRef.current[file_path] = [];
      }
      fileBuffersRef.current[file_path].push(content);

      // When complete, add to actual files
      if (is_complete) {
        const fullContent = fileBuffersRef.current[file_path].join("");
        setFiles((prev) => ({
          ...prev,
          [file_path]: fullContent,
        }));
        addTerminalLine(`✅ Complete: ${file_path}`);

        // Auto-select first generated file
        if (!selectedFile) {
          setSelectedFile(file_path);
        }
      }
    }

    if (isErrorEvent(event)) {
      const { code, message } = event.payload;
      addTerminalLine(`\n❌ Error [${code}]: ${message}`);
      setStatus(`Error: ${message}`);
    }

    // Phase 4: Multi-Agent Coordination Events
    if (isDAGUpdateEvent(event)) {
      setDagState(event.payload);
      addTerminalLine(`📊 DAG Update: ${event.payload.completed}/${event.payload.total} tasks`);
    }

    if (isTaskStartedEvent(event)) {
      const { task_id, task_name, agent_id } = event.payload;
      addTerminalLine(`▶️ Task Started: ${task_name}`);
      if (agent_id) {
        // Add to active agents
        setActiveAgents((prev) => {
          const existing = prev.find((a) => a.task_id === task_id);
          if (existing) return prev;
          return [
            ...prev,
            {
              task_id,
              agent_id,
              agent_name: agent_id, // Will be updated by AGENT_ASSIGNED
              task_name,
              started_at: event.timestamp,
            },
          ];
        });
      }
    }

    if (isTaskCompletedEvent(event)) {
      const { task_id, task_name } = event.payload;
      addTerminalLine(`✅ Task Completed: ${task_name}`);
      // Remove from active agents
      setActiveAgents((prev) => prev.filter((a) => a.task_id !== task_id));
    }

    if (isTaskFailedEvent(event)) {
      const { task_id, task_name, error } = event.payload;
      addTerminalLine(`❌ Task Failed: ${task_name} - ${error}`);
      // Remove from active agents
      setActiveAgents((prev) => prev.filter((a) => a.task_id !== task_id));
    }

    if (isAgentAssignedEvent(event)) {
      const { task_id, agent_id, agent_name } = event.payload;
      addTerminalLine(`🤖 Agent Assigned: ${agent_name} → ${task_id.slice(0, 8)}`);
      // Update active agent info
      setActiveAgents((prev) => {
        const idx = prev.findIndex((a) => a.task_id === task_id);
        if (idx >= 0) {
          const updated = [...prev];
          updated[idx] = { ...updated[idx], agent_name, agent_id };
          return updated;
        }
        return prev;
      });
    }

    if (isCIStageEvent(event)) {
      const { stage, success, output, error } = event.payload;
      const stageName = stage as "LINT" | "BUILD" | "TEST";

      if (event.type === "CI_STAGE_START") {
        setCurrentCIStage(stageName);
        addTerminalLine(`🔄 CI Stage: ${stage} starting...`);
      } else if (event.type === "CI_STAGE_COMPLETE" || event.type === "CI_STAGE_FAILED") {
        setCurrentCIStage(null);
        setCIStages((prev) => {
          // Update or add stage result
          const idx = prev.findIndex((s) => s.stage === stageName);
          const result: CIStageResult = {
            stage: stageName,
            success: success || false,
            output,
            error,
          };
          if (idx >= 0) {
            const updated = [...prev];
            updated[idx] = result;
            return updated;
          }
          return [...prev, result];
        });
        if (success) {
          addTerminalLine(`✅ CI Stage: ${stage} passed`);
        } else {
          addTerminalLine(`❌ CI Stage: ${stage} failed - ${error}`);
        }
      }
    }

    if (isExecutionEvent(event)) {
      const { message } = event.payload;
      if (event.type === "EXECUTION_STARTED") {
        addTerminalLine(`🚀 Execution Started: ${message}`);
        setStatus("Multi-Agent Executing...");
      } else if (event.type === "EXECUTION_COMPLETED") {
        addTerminalLine(`🏁 Execution Completed: ${message}`);
        setStatus("Execution Complete");
      } else if (event.type === "EXECUTION_FAILED") {
        addTerminalLine(`💥 Execution Failed: ${message}`);
        setStatus("Execution Failed");
      }
    }

    // Track all events for the event feed
    setStreamEvents((prev) => [...prev.slice(-200), event]); // Keep last 200 events
  }, [addTerminalLine, selectedFile]);

  // Streaming hook
  useStreaming({
    workflowId: workflowID,
    onEvent: handleStreamEvent,
    onConnect: () => {
      console.log("Streaming connected");
      addTerminalLine("🔌 Connected to workflow stream");
    },
    onDisconnect: () => {
      console.log("Streaming disconnected");
    },
    onError: (err) => {
      console.error("Streaming error:", err);
    },
  });

  // Auto-scroll chat
  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [chatHistory]);

  // Start or reconnect to workflow after initialization
  useEffect(() => {
    if (!isInitialized) return;

    async function verifyOrStartPod() {
      // If we have a workflow ID, verify it's still valid
      if (workflowID) {
        try {
          addTerminalLine(`🔄 Reconnecting to workflow: ${workflowID}...`);
          const res = await fetch(`/api/pods/${workflowID}`);
          if (res.ok) {
            const state = await res.json();
            setStatus("PM Agent Active");
            addTerminalLine(`✅ Reconnected to workflow: ${workflowID}`);
            // State will be populated by the polling useEffect
            return;
          } else {
            // Workflow not found, clear stored ID and create new
            addTerminalLine(`⚠️ Workflow not found, starting new pod...`);
            if (typeof window !== "undefined") {
              localStorage.removeItem(WORKFLOW_STORAGE_KEY);
            }
            setWorkflowID(null);
          }
        } catch (err) {
          console.error("Failed to verify workflow:", err);
          addTerminalLine(`⚠️ Could not verify workflow, starting new pod...`);
          if (typeof window !== "undefined") {
            localStorage.removeItem(WORKFLOW_STORAGE_KEY);
          }
          setWorkflowID(null);
        }
      }

      // No valid workflow ID, create new pod
      try {
        setStatus("Starting Pod...");
        addTerminalLine("🚀 Starting new project pod...");

        // Generate unique project ID
        const projectId = `project-${Date.now().toString(36)}`;

        const res = await fetch("/api/pods", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ project_name: projectId }),
        });

        if (!res.ok) {
          throw new Error(`Failed to start pod: ${res.status}`);
        }

        const data = await res.json();
        setWorkflowID(data.workflow_id);
        setStatus("PM Agent Active");
        addTerminalLine(`✅ Pod started: ${data.workflow_id}`);
      } catch (err) {
        console.error(err);
        setStatus("Error Starting Pod");
        addTerminalLine(`❌ Failed to start pod: ${err}`);
      }
    }

    verifyOrStartPod();
  }, [isInitialized]); // Only run once after initialization

  // Poll for workflow updates
  useEffect(() => {
    if (!workflowID) return;

    const interval = setInterval(async () => {
      try {
        const res = await fetch(`/api/pods/${workflowID}`);
        if (res.ok) {
          const state = await res.json();
          setChatHistory(state.chat_history || "");
          setProjectSpec(state.spec || "");
          setCurrentPhase(state.phase || "DISCOVERY");

          // Update status based on phase
          switch (state.phase) {
            case "DISCOVERY":
              setStatus("PM Agent Active");
              break;
            case "IMPLEMENTATION":
              setStatus("Dev Agent Coding...");
              break;
            case "REVIEW":
              setStatus("QA Agent Testing...");
              break;
            case "REFINEMENT":
              setStatus("Refining Code...");
              break;
            case "DONE":
            case "DELIVERED":
              setStatus("Ready!");
              break;
          }

          // Update files if available - check all possible field names
          const codeFiles = state.all_code_files || state.code_files || {};

          // Also merge frontend_code and backend_code if available
          const mergedFiles = {
            ...codeFiles,
            ...(state.frontend_code || {}),
            ...(state.backend_code || {}),
            ...(state.database_code || {}),
          };

          if (Object.keys(mergedFiles).length > 0) {
            setFiles((prev) => ({
              ...prev,
              ...mergedFiles,
            }));

            // Auto-select App.jsx if not already selected
            if (!selectedFile) {
              const appFile = Object.keys(mergedFiles).find(
                (f) => f.includes("App.jsx") || f.includes("App.tsx")
              );
              if (appFile) setSelectedFile(appFile);
            }
          }
        }
      } catch (err) {
        console.error("Poll failed", err);
      }
    }, 2000);

    return () => clearInterval(interval);
  }, [workflowID, selectedFile]);

  // Send message to workflow
  const sendMessage = async () => {
    if (!input.trim() || !workflowID) return;
    const msg = input;
    setInput("");

    // Optimistically add to chat
    setChatHistory((prev) => prev + `\nYou: ${msg}`);
    addTerminalLine(`📤 Sent: ${msg}`);

    try {
      await fetch(`/api/pods/${workflowID}/message`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message: msg }),
      });
    } catch (err) {
      console.error("Send failed", err);
      addTerminalLine(`❌ Failed to send message`);
    }
  };

  // Start a new pod (clears current workflow)
  const startNewPod = async () => {
    // Clear localStorage
    if (typeof window !== "undefined") {
      localStorage.removeItem(WORKFLOW_STORAGE_KEY);
    }

    // Clear URL param
    const currentUrl = new URL(window.location.href);
    currentUrl.searchParams.delete("workflow");
    router.replace(currentUrl.pathname + currentUrl.search, { scroll: false });

    // Reset state
    setWorkflowID(null);
    setFiles({});
    setChatHistory("");
    setProjectSpec("");
    setCurrentPhase("DISCOVERY");
    setSelectedFile(null);
    setTerminalOutput([]);
    setStatus("Starting new pod...");
    fileBuffersRef.current = {};
    // Reset Phase 4 state
    setDagState(null);
    setActiveAgents([]);
    setCIStages([]);
    setCurrentCIStage(null);
    setStreamEvents([]);

    // Create new pod
    try {
      addTerminalLine("🚀 Starting new project pod...");
      const projectId = `project-${Date.now().toString(36)}`;

      const res = await fetch("/api/pods", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ project_name: projectId }),
      });

      if (!res.ok) {
        throw new Error(`Failed to start pod: ${res.status}`);
      }

      const data = await res.json();
      setWorkflowID(data.workflow_id);
      setStatus("PM Agent Active");
      addTerminalLine(`✅ Pod started: ${data.workflow_id}`);
    } catch (err) {
      console.error(err);
      setStatus("Error Starting Pod");
      addTerminalLine(`❌ Failed to start pod: ${err}`);
    }
  };

  // Retry workflow start (alias for startNewPod)
  const retryStart = startNewPod;

  // Get file list for sidebar
  const fileList = Object.keys(files).sort((a, b) => {
    // Sort: src files first, then others
    const aInSrc = a.includes("/src/");
    const bInSrc = b.includes("/src/");
    if (aInSrc && !bInSrc) return -1;
    if (!aInSrc && bInSrc) return 1;
    return a.localeCompare(b);
  });

  // Format chat history for display
  const formatChatHistory = (history: string) => {
    if (!history.trim()) return null;

    return history.split("\n").filter(Boolean).map((line, i) => {
      const isUser = line.startsWith("User:") || line.startsWith("You:");
      const isPM = line.startsWith("PM:");
      const isDev = line.startsWith("Dev:");
      const isQA = line.startsWith("QA:");

      let bgColor = "bg-[var(--glass-highlight)]";
      let textColor = "text-[var(--pearl)]";
      let borderColor = "border-[var(--glass-border)]";
      let label = "";

      if (isUser) {
        bgColor = "bg-[var(--cyan-glow)]/10";
        textColor = "text-[var(--cyan-glow)]";
        borderColor = "border-[var(--cyan-glow)]/20";
        label = "You";
      } else if (isPM) {
        bgColor = "bg-[var(--violet-glow)]/10";
        textColor = "text-[var(--violet-glow)]";
        borderColor = "border-[var(--violet-glow)]/20";
        label = "PM Agent";
      } else if (isDev) {
        bgColor = "bg-[var(--emerald-glow)]/10";
        textColor = "text-[var(--emerald-glow)]";
        borderColor = "border-[var(--emerald-glow)]/20";
        label = "Dev Agent";
      } else if (isQA) {
        bgColor = "bg-[var(--amber-glow)]/10";
        textColor = "text-[var(--amber-glow)]";
        borderColor = "border-[var(--amber-glow)]/20";
        label = "QA Agent";
      }

      const content = line.replace(/^(User|You|PM|Dev|QA):/, "").trim();

      return (
        <div key={i} className={`p-4 rounded-xl mb-3 border glass ${bgColor} ${borderColor} animate-fade-in`}>
          {label && (
            <div className={`text-[10px] uppercase font-bold tracking-widest mb-1 opacity-70 ${textColor}`}>{label}</div>
          )}
          <div className={`text-sm leading-relaxed ${textColor}`}>{content}</div>
        </div>
      );
    });
  };

  return (
    <div className="flex h-screen w-full bg-[var(--obsidian)] text-[var(--ivory)] font-[var(--font-body)]">
      {/* Left Panel - Chat/Spec/Files */}
      <div className="w-1/3 flex flex-col border-r border-[var(--glass-border)] min-w-[350px] glass">
        {/* Header */}
        <div className="p-4 border-b border-[var(--glass-border)] bg-[var(--carbon)]/50">
          <div className="flex justify-between items-start mb-4">
            <div>
              <h2 className="font-bold text-xl text-gradient-white tracking-tight">Project Pod</h2>
              <div className="flex items-center gap-2 mt-2">
                <span className={`px-2.5 py-0.5 rounded-full text-[10px] uppercase tracking-wider font-bold ${phaseColors[currentPhase] || "badge-cyan"}`}>
                  {currentPhase}
                </span>
                <span className={`text-[11px] font-medium tracking-wide ${status.includes("Error") ? "text-[var(--rose-glow)]" : "text-[var(--silver)]"}`}>
                  {status}
                </span>
              </div>
            </div>
            <div className="flex gap-2">
              {status.includes("Error") && (
                <button
                  onClick={retryStart}
                  className="btn btn-secondary px-3 py-1.5 text-xs bg-red-900/20 text-red-500 border-red-500/50"
                >
                  Retry
                </button>
              )}
              <button
                onClick={startNewPod}
                className="btn btn-secondary px-3 py-1.5 text-xs hover:border-[var(--cyan-glow)] hover:text-[var(--cyan-glow)]"
                title="Start a new project pod"
              >
                + New Pod
              </button>
            </div>
          </div>

          {/* Tabs */}
          <div className="flex gap-1 flex-wrap">
            {(["CHAT", "SPEC", "FILES", "TASKS", "AGENTS"] as const).map((tab) => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`px-3 py-1.5 rounded-lg text-[11px] font-bold tracking-wider transition-all duration-300 ${activeTab === tab
                  ? "bg-[var(--cyan-glow)] text-[var(--void)] shadow-glow-cyan"
                  : "bg-[var(--glass-highlight)] text-[var(--silver)] hover:text-[var(--ivory)] hover:bg-[var(--glass-border)]"
                  }`}
              >
                {tab === "TASKS" && "📊 "}
                {tab === "AGENTS" && "🤖 "}
                {tab}
                {tab === "FILES" && fileList.length > 0 && (
                  <span className={`ml-1 px-1.5 rounded shadow-inner ${activeTab === tab ? "bg-[var(--void)]/20" : "bg-[var(--carbon)]"}`}>
                    {fileList.length}
                  </span>
                )}
                {tab === "TASKS" && dagState && (
                  <span className={`ml-1 px-1.5 rounded shadow-inner ${activeTab === tab ? "bg-[var(--void)]/20" : "bg-[var(--carbon)]"}`}>
                    {dagState.completed}/{dagState.total}
                  </span>
                )}
                {tab === "AGENTS" && activeAgents.length > 0 && (
                  <span className={`ml-1 px-1.5 rounded animate-pulse ${activeTab === tab ? "bg-[var(--void)]/20" : "bg-[var(--emerald-glow)]/30 text-[var(--emerald-glow)]"}`}>
                    {activeAgents.length}
                  </span>
                )}
              </button>
            ))}
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4">
          {activeTab === "CHAT" && (
            <div className="space-y-2">
              {formatChatHistory(chatHistory) || (
                <div className="text-center py-8">
                  <div className="text-4xl mb-3">💬</div>
                  <p className="text-zinc-500 text-sm">
                    Describe your app idea to the PM Agent...
                  </p>
                </div>
              )}
              <div ref={chatEndRef} />
            </div>
          )}

          {activeTab === "SPEC" && (
            <div className="p-3 bg-zinc-900 rounded-lg border border-zinc-800">
              <h3 className="font-bold mb-2 text-zinc-400 text-sm">Specification</h3>
              {projectSpec ? (
                <pre className="text-sm whitespace-pre-wrap text-zinc-300">{projectSpec}</pre>
              ) : (
                <p className="text-zinc-600 italic text-sm">
                  No specification generated yet. Chat with the PM to define your app.
                </p>
              )}
            </div>
          )}

          {activeTab === "FILES" && (
            <div className="space-y-1.5">
              {fileList.length > 0 ? (
                fileList.map((filepath) => (
                  <button
                    key={filepath}
                    onClick={() => {
                      setSelectedFile(filepath);
                      setRightTab("CODE");
                    }}
                    className={`w-full text-left px-3 py-2.5 rounded-lg text-xs font-mono transition-all duration-200 border ${selectedFile === filepath
                      ? "bg-[var(--cyan-glow)]/10 border-[var(--cyan-glow)]/40 text-[var(--cyan-glow)] shadow-glow-cyan/20"
                      : "bg-transparent border-transparent text-[var(--silver)] hover:bg-[var(--glass-highlight)] hover:text-[var(--pearl)]"
                      }`}
                  >
                    <span className="opacity-50 mr-2 text-[10px]">📄</span>
                    {filepath}
                  </button>
                ))
              ) : (
                <div className="text-center py-12 glass rounded-2xl">
                  <div className="text-4xl mb-4 opacity-50">📁</div>
                  <p className="text-[var(--silver)] text-sm px-4">
                    No files generated yet. Approve your spec to start coding.
                  </p>
                </div>
              )}
            </div>
          )}

          {/* Phase 4: Tasks (DAG) Tab */}
          {activeTab === "TASKS" && (
            <DAGProgressDashboard
              dagState={dagState}
              onTaskClick={(task) => {
                console.log("Task clicked:", task);
                addTerminalLine(`📋 Selected task: ${task.name} (${task.status})`);
              }}
            />
          )}

          {/* Phase 4: Agents Tab */}
          {activeTab === "AGENTS" && (
            <AgentActivityPanel
              activeAgents={activeAgents}
              onAgentClick={(agent) => {
                console.log("Agent clicked:", agent);
                addTerminalLine(`🤖 Selected agent: ${agent.agent_name} working on ${agent.task_name}`);
              }}
            />
          )}
        </div>

        {/* Input */}
        <div className="p-4 border-t border-[var(--glass-border)] bg-[var(--carbon)]/50">
          <div className="flex gap-2">
            <input
              className="flex-1 bg-[var(--obsidian)] border border-[var(--steel)] rounded-xl px-4 py-3 text-sm focus:outline-none focus:border-[var(--cyan-glow)] focus:ring-1 focus:ring-[var(--cyan-glow)]/50 transition-all placeholder:text-[var(--silver)]/50"
              placeholder={
                currentPhase === "DISCOVERY"
                  ? "Describe your app or type /approve..."
                  : "Send feedback or type /approve..."
              }
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && sendMessage()}
            />
            <button
              onClick={sendMessage}
              disabled={!input.trim()}
              className="btn btn-primary px-6 rounded-xl text-sm font-bold"
            >
              Send
            </button>
          </div>
          <div className="mt-2 text-xs text-zinc-600">
            Type <code className="bg-zinc-800 px-1.5 py-0.5 rounded">/approve</code> to proceed to next phase
          </div>
        </div>
      </div>

      {/* Right Panel - Preview/Code/Terminal */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Tab Bar */}
        <div className="flex text-[11px] font-bold tracking-widest uppercase border-b border-[var(--glass-border)] bg-[var(--carbon)]/30 backdrop-blur-md overflow-x-auto">
          {(["PREVIEW", "CODE", "TERMINAL", "CI", "EVENTS"] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setRightTab(tab)}
              className={`px-6 py-4 transition-all duration-300 relative whitespace-nowrap flex items-center gap-2 ${rightTab === tab
                  ? "text-[var(--cyan-glow)]"
                  : "text-[var(--silver)] hover:text-[var(--ivory)] hover:bg-[var(--glass-highlight)]"
                }`}
            >
              {rightTab === tab && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-[var(--cyan-glow)] shadow-glow-cyan" />
              )}
              {tab === "PREVIEW" && "▶ "}
              {tab === "CODE" && "📝 "}
              {tab === "TERMINAL" && "⌨ "}
              {tab === "CI" && "🔧 "}
              {tab === "EVENTS" && "📡 "}
              {tab}
              {tab === "CI" && ciStages.length > 0 && (
                <span className={`ml-1 text-[10px] px-1.5 py-0.5 rounded shadow-glow-rose/20 ${ciStages.some(s => !s.success) ? "bg-[var(--rose-glow)]/20 text-[var(--rose-glow)]" :
                    ciStages.length === 3 ? "bg-[var(--emerald-glow)]/20 text-[var(--emerald-glow)]" : "bg-[var(--amber-glow)]/20 text-[var(--amber-glow)]"
                  }`}>
                  {ciStages.filter(s => s.success).length}/3
                </span>
              )}
              {tab === "EVENTS" && streamEvents.length > 0 && (
                <span className="ml-1 text-[10px] bg-[var(--carbon)] px-1.5 py-0.5 rounded text-[var(--silver)] border border-[var(--glass-border)]">
                  {streamEvents.length}
                </span>
              )}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="flex-1 min-h-0 bg-[var(--void)] relative">
          <div className="absolute inset-0 grid-overlay opacity-30 pointer-events-none" />

          {/* Preview Tab - WebContainer */}
          {rightTab === "PREVIEW" && (
            <div className="h-full relative z-10 animate-fade-in">
              <WebContainerPreview
                files={files}
                onTerminalOutput={addTerminalLine}
              />
            </div>
          )}

          {/* Code Tab - Sandpack with editor */}
          {rightTab === "CODE" && (
            <div className="h-full relative z-10 animate-fade-in">
              <SandpackPreview
                files={files}
                activeFile={selectedFile || undefined}
                showFileExplorer={true}
                showEditor={true}
                showConsole={false}
                theme="dark"
              />
            </div>
          )}

          {/* Terminal Tab */}
          {rightTab === "TERMINAL" && (
            <div className="h-full overflow-auto p-6 font-mono text-sm bg-[var(--void)] relative z-10 animate-fade-in">
              <div className="text-[var(--emerald-glow)] mb-4 font-bold tracking-tight opacity-80 flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-[var(--emerald-glow)] animate-pulse" />
                Sovereign Firm Console — v3.0
              </div>
              {terminalOutput.length > 0 ? (
                <div className="space-y-1">
                  {terminalOutput.map((line, i) => (
                    <div
                      key={i}
                      className={`whitespace-pre-wrap leading-relaxed ${line.includes("❌") ? "text-[var(--rose-glow)]" :
                          line.includes("✅") ? "text-[var(--emerald-glow)]" :
                            line.includes("📍") ? "text-[var(--cyan-glow)]" :
                              line.includes("📝") ? "text-[var(--amber-glow)]" :
                                "text-[var(--silver)]"
                        }`}
                    >
                      <span className="opacity-30 mr-2">›</span>
                      {line}
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-[var(--silver)]/30 italic">Awaiting neural orchestration events...</div>
              )}
            </div>
          )}

          {/* CI Tab - Phase 4 */}
          {rightTab === "CI" && (
            <div className="h-full relative z-10 animate-fade-in">
              <CIStatusPanel
                stages={ciStages}
                currentStage={currentCIStage}
                onStageClick={(stage) => {
                  console.log("CI Stage clicked:", stage);
                  if (stage.output) {
                    addTerminalLine(`\n🔧 ${stage.stage} Output:\n${stage.output}`);
                  }
                }}
              />
            </div>
          )}

          {/* Events Tab - Phase 4 */}
          {rightTab === "EVENTS" && (
            <div className="h-full relative z-10 animate-fade-in">
              <ExecutionEventFeed
                events={streamEvents}
                onEventClick={(event) => {
                  console.log("Event clicked:", event);
                  addTerminalLine(`📡 Event #${event.seq}: ${event.type}`);
                }}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
