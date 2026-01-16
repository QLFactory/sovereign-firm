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

export default function PodConsole() {
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

  // Initialize workflow ID from URL param or localStorage
  useEffect(() => {
    if (isInitialized) return;

    const urlWorkflowId = searchParams.get("workflow");
    const storedWorkflowId = typeof window !== "undefined"
      ? localStorage.getItem(WORKFLOW_STORAGE_KEY)
      : null;

    if (urlWorkflowId) {
      // URL param takes priority
      setWorkflowID(urlWorkflowId);
      if (typeof window !== "undefined") {
        localStorage.setItem(WORKFLOW_STORAGE_KEY, urlWorkflowId);
      }
      setStatus("Reconnecting to workflow...");
    } else if (storedWorkflowId) {
      // Fall back to localStorage
      setWorkflowID(storedWorkflowId);
      setStatus("Reconnecting to workflow...");
    }

    setIsInitialized(true);
  }, [searchParams, isInitialized]);

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

          // Update files if available
          if (state.code_files && Object.keys(state.code_files).length > 0) {
            setFiles((prev) => ({
              ...prev,
              ...state.code_files,
            }));

            // Auto-select App.jsx if not already selected
            if (!selectedFile) {
              const appFile = Object.keys(state.code_files).find(
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

      let bgColor = "bg-zinc-800";
      let textColor = "text-zinc-300";
      let label = "";

      if (isUser) {
        bgColor = "bg-blue-900/50";
        textColor = "text-blue-200";
        label = "You";
      } else if (isPM) {
        bgColor = "bg-purple-900/50";
        textColor = "text-purple-200";
        label = "PM Agent";
      } else if (isDev) {
        bgColor = "bg-green-900/50";
        textColor = "text-green-200";
        label = "Dev Agent";
      } else if (isQA) {
        bgColor = "bg-yellow-900/50";
        textColor = "text-yellow-200";
        label = "QA Agent";
      }

      const content = line.replace(/^(User|You|PM|Dev|QA):/, "").trim();

      return (
        <div key={i} className={`p-3 rounded-lg mb-2 ${bgColor}`}>
          {label && (
            <div className={`text-xs font-semibold mb-1 ${textColor}`}>{label}</div>
          )}
          <div className={`text-sm whitespace-pre-wrap ${textColor}`}>{content}</div>
        </div>
      );
    });
  };

  return (
    <div className="flex h-screen w-full bg-zinc-950 text-white">
      {/* Left Panel - Chat/Spec/Files */}
      <div className="w-1/3 flex flex-col border-r border-zinc-800 min-w-[350px]">
        {/* Header */}
        <div className="p-4 border-b border-zinc-800 bg-zinc-900">
          <div className="flex justify-between items-start mb-3">
            <div>
              <h2 className="font-bold text-lg">Project Pod</h2>
              <div className="flex items-center gap-2 mt-1">
                <span className={`px-2 py-0.5 rounded text-xs font-medium ${phaseColors[currentPhase] || "bg-zinc-700"}`}>
                  {currentPhase}
                </span>
                <span className={`text-xs ${status.includes("Error") ? "text-red-400" : "text-zinc-400"}`}>
                  {status}
                </span>
              </div>
            </div>
            <div className="flex gap-2">
              {status.includes("Error") && (
                <button
                  onClick={retryStart}
                  className="text-xs bg-red-900 px-3 py-1 rounded hover:bg-red-800"
                >
                  Retry
                </button>
              )}
              <button
                onClick={startNewPod}
                className="text-xs bg-zinc-700 px-3 py-1 rounded hover:bg-zinc-600"
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
                className={`px-3 py-1.5 rounded text-xs font-medium transition-colors ${
                  activeTab === tab
                    ? "bg-blue-600 text-white"
                    : "bg-zinc-800 text-zinc-400 hover:text-white"
                }`}
              >
                {tab === "TASKS" && "📊 "}
                {tab === "AGENTS" && "🤖 "}
                {tab}
                {tab === "FILES" && fileList.length > 0 && (
                  <span className="ml-1 bg-zinc-700 px-1.5 rounded">{fileList.length}</span>
                )}
                {tab === "TASKS" && dagState && (
                  <span className="ml-1 bg-zinc-700 px-1.5 rounded">
                    {dagState.completed}/{dagState.total}
                  </span>
                )}
                {tab === "AGENTS" && activeAgents.length > 0 && (
                  <span className="ml-1 bg-green-700 px-1.5 rounded animate-pulse">
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
            <div className="space-y-1">
              {fileList.length > 0 ? (
                fileList.map((filepath) => (
                  <button
                    key={filepath}
                    onClick={() => {
                      setSelectedFile(filepath);
                      setRightTab("CODE");
                    }}
                    className={`w-full text-left px-3 py-2 rounded text-sm font-mono transition-colors ${
                      selectedFile === filepath
                        ? "bg-blue-600 text-white"
                        : "bg-zinc-800 text-zinc-400 hover:bg-zinc-700 hover:text-white"
                    }`}
                  >
                    {filepath}
                  </button>
                ))
              ) : (
                <div className="text-center py-8">
                  <div className="text-4xl mb-3">📁</div>
                  <p className="text-zinc-500 text-sm">
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
        <div className="p-4 border-t border-zinc-800 bg-zinc-900">
          <div className="flex gap-2">
            <input
              className="flex-1 bg-zinc-950 border border-zinc-700 rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
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
              className="bg-blue-600 hover:bg-blue-500 disabled:bg-zinc-700 disabled:cursor-not-allowed px-5 rounded-lg font-semibold text-sm transition-colors"
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
        <div className="flex text-sm font-medium border-b border-zinc-800 bg-zinc-900 overflow-x-auto">
          {(["PREVIEW", "CODE", "TERMINAL", "CI", "EVENTS"] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setRightTab(tab)}
              className={`px-4 py-3 transition-colors whitespace-nowrap flex items-center gap-1 ${
                rightTab === tab
                  ? "bg-zinc-800 text-white border-b-2 border-blue-500"
                  : "text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800/50"
              }`}
            >
              {tab === "PREVIEW" && "▶ "}
              {tab === "CODE" && "📝 "}
              {tab === "TERMINAL" && "⌨ "}
              {tab === "CI" && "🔧 "}
              {tab === "EVENTS" && "📡 "}
              {tab}
              {tab === "CI" && ciStages.length > 0 && (
                <span className={`ml-1 text-xs px-1.5 rounded ${
                  ciStages.some(s => !s.success) ? "bg-red-700" :
                  ciStages.length === 3 ? "bg-green-700" : "bg-yellow-700"
                }`}>
                  {ciStages.filter(s => s.success).length}/3
                </span>
              )}
              {tab === "EVENTS" && streamEvents.length > 0 && (
                <span className="ml-1 text-xs bg-zinc-700 px-1.5 rounded">
                  {streamEvents.length}
                </span>
              )}
            </button>
          ))}
        </div>

        {/* Content */}
        <div className="flex-1 min-h-0 bg-zinc-950">
          {/* Preview Tab - WebContainer */}
          {rightTab === "PREVIEW" && (
            <div className="h-full">
              <WebContainerPreview
                files={files}
                onTerminalOutput={addTerminalLine}
              />
            </div>
          )}

          {/* Code Tab - Sandpack with editor */}
          {rightTab === "CODE" && (
            <div className="h-full">
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
            <div className="h-full overflow-auto p-4 font-mono text-sm bg-zinc-950">
              <div className="text-green-400 mb-2">$ Sovereign Firm Terminal</div>
              {terminalOutput.length > 0 ? (
                terminalOutput.map((line, i) => (
                  <div
                    key={i}
                    className={`whitespace-pre-wrap ${
                      line.includes("❌") ? "text-red-400" :
                      line.includes("✅") ? "text-green-400" :
                      line.includes("📍") ? "text-blue-400" :
                      line.includes("📝") ? "text-yellow-400" :
                      "text-zinc-400"
                    }`}
                  >
                    {line}
                  </div>
                ))
              ) : (
                <div className="text-zinc-600">Waiting for activity...</div>
              )}
            </div>
          )}

          {/* CI Tab - Phase 4 */}
          {rightTab === "CI" && (
            <div className="h-full">
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
            <div className="h-full">
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
