"use client";

import { useState, useEffect } from "react";
import {
  SandpackProvider,
  SandpackLayout,
  SandpackCodeEditor,
  SandpackPreview,
  SandpackFileExplorer,
} from "@codesandbox/sandpack-react";

export default function PodConsole() {
  const [workflowID, setWorkflowID] = useState<string | null>(null);
  const [chatHistory, setChatHistory] = useState<string>("");
  const [input, setInput] = useState("");
  const [status, setStatus] = useState("Idle");

  const [activeTab, setActiveTab] = useState<"CHAT" | "SPEC">("CHAT");
  const [projectSpec, setProjectSpec] = useState("");

  // Default files
  const [files, setFiles] = useState<any>({
    "/App.js": `import React from 'react';

export default function App() {
  return (
    <div style={{ fontFamily: 'system-ui, sans-serif', padding: '2rem', background: '#18181b', color: '#fff', height: '100vh' }}>
      <h1>Waiting for Spec Approval...</h1>
      <p style={{ color: '#a1a1aa' }}>Type "/approve" to generate app.</p>
    </div>
  );
}`,
    "/index.js": `import React from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';

const root = createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);`,
  });

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
  }, [workflowID]); // Add dep to allow retry if ID reset? No.

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
            setFiles(state.code_files);
          }
        }
      } catch (err) {
        console.error("Poll failed", err);
      }
    }, 2000);
    return () => clearInterval(interval);
  }, [workflowID]);

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

  return (
    <div className="flex h-screen w-full bg-zinc-950 text-white">
      <div className="w-1/3 flex flex-col border-r border-zinc-800">
        <div className="p-4 border-b border-zinc-800 bg-zinc-900 flex justify-between items-center">
          <div>
            <h2 className="font-bold">Project Pod</h2>
            <div className="flex items-center gap-2 mt-1">
              <div className={`text-xs ${status.includes("Error") ? "text-red-500" : "text-green-400"}`}>● {status}</div>
              {status.includes("Error") && (
                <button onClick={retryStart} className="text-xs bg-red-900 px-2 py-0.5 rounded hover:bg-red-800">Retry</button>
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

      <div className="flex-1 flex flex-col">
        <SandpackProvider
          template="react"
          theme="dark"
          files={files || {}}
        >
          <SandpackLayout className="!h-screen !border-none">
            <SandpackFileExplorer />
            <SandpackCodeEditor showLineNumbers />
            <SandpackPreview />
          </SandpackLayout>
        </SandpackProvider>
      </div>
    </div>
  );
}
