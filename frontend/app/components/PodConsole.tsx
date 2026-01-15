"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { WebContainer, FileSystemTree } from "@webcontainer/api";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import {
  useStreaming,
  StreamEvent,
  isPhaseChangeEvent,
  isCodeChunkEvent,
  isChatMessageEvent,
  isFileStartEvent,
  isErrorEvent
} from "../hooks/useStreaming";

// Convert flat file map to WebContainer file tree format
function convertToFileTree(files: Record<string, string>): FileSystemTree {
  const tree: FileSystemTree = {};

  for (const [path, contents] of Object.entries(files)) {
    const parts = path.replace(/^\//, "").split("/");
    let current: FileSystemTree = tree;

    for (let i = 0; i < parts.length - 1; i++) {
      const dir = parts[i];
      if (!current[dir]) {
        current[dir] = { directory: {} };
      }
      const node = current[dir];
      if ("directory" in node && node.directory) {
        current = node.directory;
      }
    }

    const fileName = parts[parts.length - 1];
    current[fileName] = { file: { contents } };
  }

  return tree;
}

// Default project files
const defaultFiles: Record<string, string> = {
  "/package.json": JSON.stringify({
    name: "sovereign-preview",
    type: "module",
    scripts: {
      dev: "vite",
      build: "vite build",
      test: "vitest run"
    },
    dependencies: {
      react: "^18.2.0",
      "react-dom": "^18.2.0"
    },
    devDependencies: {
      vite: "^5.0.0",
      "@vitejs/plugin-react": "^4.2.0",
      vitest: "^1.0.0",
      "@testing-library/react": "^14.0.0",
      jsdom: "^23.0.0"
    }
  }, null, 2),

  "/vite.config.js": `import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
});`,

  "/vitest.config.js": `import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
  },
});`,

  "/index.html": `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Sovereign Preview</title>
</head>
<body>
  <div id="root"></div>
  <script type="module" src="/src/main.jsx"></script>
</body>
</html>`,

  "/src/main.jsx": `import React from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';

const root = createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);`,

  "/src/App.jsx": `import React from 'react';

export default function App() {
  return (
    <div style={{ 
      fontFamily: 'system-ui, sans-serif', 
      padding: '2rem', 
      background: '#18181b', 
      color: '#fff', 
      minHeight: '100vh' 
    }}>
      <h1>👋 Waiting for Spec Approval...</h1>
      <p style={{ color: '#a1a1aa' }}>
        Type "/approve" in the chat to generate your app.
      </p>
    </div>
  );
}`,
};

export default function PodConsole() {
  // Workflow state
  const [workflowID, setWorkflowID] = useState<string | null>(null);
  const [chatHistory, setChatHistory] = useState<string>("");
  const [input, setInput] = useState("");
  const [status, setStatus] = useState("Initializing...");
  const [activeTab, setActiveTab] = useState<"CHAT" | "SPEC">("CHAT");
  const [rightTab, setRightTab] = useState<"PREVIEW" | "TERMINAL" | "TESTS">("PREVIEW");
  const [projectSpec, setProjectSpec] = useState("");
  const [currentPhase, setCurrentPhase] = useState("DISCOVERY");

  // WebContainer state
  const [files, setFiles] = useState<Record<string, string>>(defaultFiles);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [isBooting, setIsBooting] = useState(true);
  const [terminalOutput, setTerminalOutput] = useState<string[]>([]);

  // Streaming state
  const [streamingFiles, setStreamingFiles] = useState<Record<string, string>>({});
  const fileBuffersRef = useRef<Record<string, string[]>>({});

  // Refs
  const webcontainerRef = useRef<WebContainer | null>(null);
  const terminalRef = useRef<HTMLDivElement>(null);
  const terminalInstanceRef = useRef<Terminal | null>(null);
  const iframeRef = useRef<HTMLIFrameElement>(null);

  // Handle streaming events
  const handleStreamEvent = useCallback((event: StreamEvent) => {
    console.log("Stream event:", event.type, event.payload);

    if (isPhaseChangeEvent(event)) {
      setCurrentPhase(event.payload.current_phase);
      setStatus(`Phase: ${event.payload.current_phase}`);
      addTerminalLine(`\n📍 Phase: ${event.payload.current_phase}`);
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

      // Update streaming preview
      setStreamingFiles((prev) => ({
        ...prev,
        [file_path]: fileBuffersRef.current[file_path].join(""),
      }));

      // When complete, add to actual files
      if (is_complete) {
        const fullContent = fileBuffersRef.current[file_path].join("");
        setFiles((prev) => ({
          ...prev,
          [file_path]: fullContent,
        }));
        addTerminalLine(`✅ Complete: ${file_path}`);
      }
    }

    if (isErrorEvent(event)) {
      const { code, message } = event.payload;
      addTerminalLine(`\n❌ Error [${code}]: ${message}`);
      setStatus(`Error: ${message}`);
    }
  }, []);

  // Streaming hook
  const { isConnected: isStreaming } = useStreaming({
    workflowId: workflowID,
    onEvent: handleStreamEvent,
    onConnect: () => {
      console.log("Streaming connected");
      addTerminalLine("\n🔌 Streaming connected");
    },
    onDisconnect: () => {
      console.log("Streaming disconnected");
    },
    onError: (err) => {
      console.error("Streaming error:", err);
    },
  });

  // Boot WebContainer
  useEffect(() => {
    let mounted = true;

    async function bootWebContainer() {
      try {
        setStatus("Booting WebContainer...");

        // Boot only once
        if (webcontainerRef.current) return;

        const instance = await WebContainer.boot();
        if (!mounted) return;

        webcontainerRef.current = instance;

        // Mount initial files
        const fileTree = convertToFileTree(files);
        await instance.mount(fileTree);

        // Listen for server ready
        instance.on("server-ready", (port, url) => {
          console.log(`Server ready on port ${port}: ${url}`);
          setPreviewUrl(url);
          setStatus("Preview Ready");
        });

        // Install dependencies
        setStatus("Installing dependencies...");
        addTerminalLine("$ npm install");

        const installProcess = await instance.spawn("npm", ["install"]);

        installProcess.output.pipeTo(
          new WritableStream({
            write(data) {
              addTerminalLine(data);
            },
          })
        );

        const installExitCode = await installProcess.exit;

        if (installExitCode !== 0) {
          setStatus("Install failed");
          addTerminalLine(`\n❌ npm install failed with code ${installExitCode}`);
          return;
        }

        addTerminalLine("\n✅ Dependencies installed");

        // Start dev server
        setStatus("Starting dev server...");
        addTerminalLine("\n$ npm run dev");

        const devProcess = await instance.spawn("npm", ["run", "dev"]);

        devProcess.output.pipeTo(
          new WritableStream({
            write(data) {
              addTerminalLine(data);
            },
          })
        );

        setIsBooting(false);

      } catch (error) {
        console.error("WebContainer boot failed:", error);
        setStatus("WebContainer Error");
        addTerminalLine(`\n❌ Error: ${error}`);
        setIsBooting(false);
      }
    }

    bootWebContainer();

    return () => {
      mounted = false;
    };
  }, []);

  // Update files in WebContainer when they change
  useEffect(() => {
    async function updateFiles() {
      if (!webcontainerRef.current || isBooting) return;

      try {
        const fileTree = convertToFileTree(files);
        await webcontainerRef.current.mount(fileTree);
        addTerminalLine("\n📝 Files updated");
      } catch (error) {
        console.error("Failed to update files:", error);
      }
    }

    updateFiles();
  }, [files, isBooting]);

  // Terminal output helper
  const addTerminalLine = useCallback((line: string) => {
    setTerminalOutput((prev) => [...prev, line]);
  }, []);

  // Initialize terminal
  useEffect(() => {
    if (!terminalRef.current || terminalInstanceRef.current) return;

    const terminal = new Terminal({
      theme: {
        background: "#18181b",
        foreground: "#fafafa",
        cursor: "#fafafa",
      },
      fontFamily: "JetBrains Mono, Menlo, Monaco, monospace",
      fontSize: 13,
      lineHeight: 1.4,
    });

    const fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.open(terminalRef.current);
    fitAddon.fit();

    terminalInstanceRef.current = terminal;

    // Handle resize
    const resizeObserver = new ResizeObserver(() => {
      fitAddon.fit();
    });
    resizeObserver.observe(terminalRef.current);

    return () => {
      resizeObserver.disconnect();
      terminal.dispose();
    };
  }, []);

  // Write terminal output
  useEffect(() => {
    if (!terminalInstanceRef.current) return;
    terminalOutput.forEach((line) => {
      terminalInstanceRef.current?.write(line);
    });
  }, [terminalOutput]);

  // Start workflow on mount
  useEffect(() => {
    async function startPod() {
      try {
        setStatus("Starting Pod...");
        const res = await fetch("/api/pods", {
          method: "POST",
          body: JSON.stringify({ project_name: "demo-project" }),
        });
        const data = await res.json();
        setWorkflowID(data.workflow_id);
        setStatus("PM Agent Active");
      } catch (err) {
        console.error(err);
        setStatus("Error Starting Pod");
      }
    }
    if (!workflowID) startPod();
  }, [workflowID]);

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

          if (state.phase === "IMPLEMENTATION" || state.phase === "REFINEMENT") {
            setStatus(state.phase === "REFINEMENT" ? "Refining Code..." : "Architecting Solution...");
          }

          if ((state.phase === "DELIVERED" || state.phase === "REVIEW" || state.phase === "DONE") && state.code_files) {
            setStatus("App Deployed / Ready for Review");

            // Merge generated files with defaults (keep package.json, vite config, etc.)
            setFiles((prev) => ({
              ...prev,
              ...state.code_files,
            }));
          }
        }
      } catch (err) {
        console.error("Poll failed", err);
      }
    }, 2000);

    return () => clearInterval(interval);
  }, [workflowID]);

  // Send message to workflow
  const sendMessage = async () => {
    if (!input.trim() || !workflowID) return;
    const msg = input;
    setInput("");
    try {
      await fetch(`/api/pods/${workflowID}/message`, {
        method: "POST",
        body: JSON.stringify({ message: msg }),
      });
    } catch (err) {
      console.error("Send failed", err);
    }
  };

  // Run tests
  const runTests = async () => {
    if (!webcontainerRef.current) return;

    addTerminalLine("\n$ npm run test");
    setRightTab("TERMINAL");

    const testProcess = await webcontainerRef.current.spawn("npm", ["run", "test"]);

    testProcess.output.pipeTo(
      new WritableStream({
        write(data) {
          addTerminalLine(data);
        },
      })
    );

    const exitCode = await testProcess.exit;
    addTerminalLine(exitCode === 0 ? "\n✅ Tests passed" : "\n❌ Tests failed");
  };

  // Retry workflow start
  const retryStart = async () => {
    setStatus("Retrying...");
    try {
      const res = await fetch("/api/pods", {
        method: "POST",
        body: JSON.stringify({ project_name: "demo-project" }),
      });
      if (!res.ok) throw new Error("Failed");
      const data = await res.json();
      setWorkflowID(data.workflow_id);
      setStatus("PM Agent Active");
    } catch (err) {
      setStatus("Error Starting Pod");
    }
  };

  return (
    <div className="flex h-screen w-full bg-zinc-950 text-white">
      {/* Left Panel - Chat */}
      <div className="w-1/3 flex flex-col border-r border-zinc-800">
        <div className="p-4 border-b border-zinc-800 bg-zinc-900 flex justify-between items-center">
          <div>
            <h2 className="font-bold">Project Pod</h2>
            <div className="flex items-center gap-2 mt-1">
              <div className={`text-xs ${status.includes("Error") ? "text-red-500" : "text-green-400"}`}>
                ● {status}
              </div>
              {status.includes("Error") && (
                <button onClick={retryStart} className="text-xs bg-red-900 px-2 py-0.5 rounded hover:bg-red-800">
                  Retry
                </button>
              )}
            </div>
          </div>
          <div className="flex gap-2 text-xs">
            <button
              onClick={() => setActiveTab("CHAT")}
              className={`px-2 py-1 rounded ${activeTab === "CHAT" ? "bg-blue-600" : "bg-zinc-800"}`}
            >
              Chat
            </button>
            <button
              onClick={() => setActiveTab("SPEC")}
              className={`px-2 py-1 rounded ${activeTab === "SPEC" ? "bg-blue-600" : "bg-zinc-800"}`}
            >
              Spec
            </button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-4 font-mono text-sm whitespace-pre-wrap">
          {activeTab === "CHAT" ? (
            chatHistory || <span className="text-zinc-600 italic">Waiting for PM...</span>
          ) : (
            <div className="p-2 bg-zinc-900 rounded border border-zinc-800">
              <h3 className="font-bold mb-2 text-zinc-400">Current Specification</h3>
              {projectSpec || <span className="text-zinc-600 italic">No spec generated yet...</span>}
            </div>
          )}
        </div>

        <div className="p-4 border-t border-zinc-800 bg-zinc-900">
          <div className="flex gap-2">
            <input
              className="flex-1 bg-zinc-950 border border-zinc-700 rounded px-3 py-2 focus:outline-none focus:border-blue-500"
              placeholder="Type message or /approve..."
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && sendMessage()}
            />
            <button
              onClick={sendMessage}
              className="bg-blue-600 hover:bg-blue-500 px-4 rounded font-bold"
            >
              Send
            </button>
          </div>
        </div>
      </div>

      {/* Right Panel - Preview/Terminal/Tests */}
      <div className="flex-1 flex flex-col">
        {/* Tab Bar */}
        <div className="flex text-xs font-bold border-b border-zinc-800 bg-zinc-900">
          <button
            onClick={() => setRightTab("PREVIEW")}
            className={`flex-1 p-2 ${rightTab === "PREVIEW" ? "bg-zinc-800 text-white" : "text-zinc-500 hover:text-zinc-300"}`}
          >
            Preview {previewUrl && "🟢"}
          </button>
          <button
            onClick={() => setRightTab("TERMINAL")}
            className={`flex-1 p-2 ${rightTab === "TERMINAL" ? "bg-zinc-800 text-white" : "text-zinc-500 hover:text-zinc-300"}`}
          >
            Terminal
          </button>
          <button
            onClick={runTests}
            className="flex-1 p-2 text-zinc-500 hover:text-zinc-300 border-l border-zinc-800"
          >
            ▶ Run Tests
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 relative bg-zinc-950">
          {/* Preview iframe */}
          <iframe
            ref={iframeRef}
            src={previewUrl || "about:blank"}
            className={`w-full h-full border-none ${rightTab === "PREVIEW" ? "block" : "hidden"}`}
            title="Preview"
          />

          {/* Terminal */}
          <div
            ref={terminalRef}
            className={`w-full h-full p-2 ${rightTab === "TERMINAL" ? "block" : "hidden"}`}
          />

          {/* Loading state */}
          {rightTab === "PREVIEW" && !previewUrl && (
            <div className="absolute inset-0 flex items-center justify-center bg-zinc-950">
              <div className="text-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
                <p className="text-zinc-400">{status}</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
