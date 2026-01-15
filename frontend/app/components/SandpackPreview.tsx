"use client";

import { useMemo } from "react";
import {
  SandpackProvider,
  SandpackLayout,
  SandpackCodeEditor,
  SandpackPreview as SandpackPreviewPane,
  SandpackFileExplorer,
  SandpackConsole,
} from "@codesandbox/sandpack-react";

interface SandpackPreviewProps {
  files: Record<string, string>;
  activeFile?: string;
  showFileExplorer?: boolean;
  showConsole?: boolean;
  showEditor?: boolean;
  theme?: "dark" | "light";
}

// Convert our file format to Sandpack format
function convertToSandpackFiles(files: Record<string, string>): Record<string, { code: string; active?: boolean }> {
  const sandpackFiles: Record<string, { code: string; active?: boolean }> = {};

  for (const [path, content] of Object.entries(files)) {
    // Sandpack requires paths with leading slash
    const normalizedPath = path.startsWith("/") ? path : `/${path}`;
    sandpackFiles[normalizedPath] = { code: content };
  }

  return sandpackFiles;
}

// Default App.jsx for empty state
const DEFAULT_APP = `export default function App() {
  return (
    <div style={{
      fontFamily: 'system-ui, sans-serif',
      padding: '2rem',
      background: '#18181b',
      color: '#fff',
      minHeight: '100vh',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center'
    }}>
      <h1 style={{ fontSize: '2rem', marginBottom: '1rem' }}>👋 Waiting for Code...</h1>
      <p style={{ color: '#a1a1aa' }}>
        Describe your app in the chat, then type <code style={{
          background: '#27272a',
          padding: '0.25rem 0.5rem',
          borderRadius: '4px'
        }}>/approve</code> to generate.
      </p>
    </div>
  );
}`;

export default function SandpackPreview({
  files,
  activeFile,
  showFileExplorer = true,
  showConsole = false,
  showEditor = true,
  theme = "dark",
}: SandpackPreviewProps) {
  // Build Sandpack files with proper format
  const sandpackFiles = useMemo(() => {
    // Check if we have any App file
    const hasAppFile = Object.keys(files).some(
      (f) => f.includes("App.jsx") || f.includes("App.tsx") || f.includes("App.js")
    );

    // If no files or no App file, use default
    if (!hasAppFile || Object.keys(files).length === 0) {
      return {
        "/App.js": { code: DEFAULT_APP, active: true },
      };
    }

    // Convert provided files to Sandpack format
    const converted = convertToSandpackFiles(files);

    // Find the main entry file and mark it active
    const appFileKey = Object.keys(converted).find(
      (f) => f.includes("App.jsx") || f.includes("App.tsx") || f.includes("App.js")
    );

    if (appFileKey) {
      converted[appFileKey] = { ...converted[appFileKey], active: true };
    }

    return converted;
  }, [files]);

  // Determine active file for editor
  const visibleFile = useMemo(() => {
    if (activeFile) {
      const normalized = activeFile.startsWith("/") ? activeFile : `/${activeFile}`;
      return normalized;
    }
    // Find the main App file
    const appFile = Object.keys(sandpackFiles).find(
      (f) => f.includes("App.jsx") || f.includes("App.tsx") || f.includes("App.js")
    );
    return appFile || "/App.js";
  }, [activeFile, sandpackFiles]);

  // For preview-only mode, render just the preview without layout wrapper
  if (!showEditor && !showFileExplorer) {
    return (
      <SandpackProvider
        template="react"
        theme={theme === "dark" ? "dark" : "light"}
        files={sandpackFiles}
        options={{
          activeFile: visibleFile,
          recompileMode: "delayed",
          recompileDelay: 300,
        }}
      >
        <div style={{ height: "100%", width: "100%", display: "flex", flexDirection: "column" }}>
          <SandpackPreviewPane
            style={{
              height: "100%",
              width: "100%",
              flex: 1,
            }}
            showNavigator
            showRefreshButton
            showOpenInCodeSandbox={false}
          />
          {showConsole && (
            <SandpackConsole
              style={{
                height: "150px",
              }}
            />
          )}
        </div>
      </SandpackProvider>
    );
  }

  return (
    <SandpackProvider
      template="react"
      theme={theme === "dark" ? "dark" : "light"}
      files={sandpackFiles}
      options={{
        activeFile: visibleFile,
        visibleFiles: Object.keys(sandpackFiles).filter(
          (f) => !f.includes("package.json") && !f.includes("vite.config")
        ),
        recompileMode: "delayed",
        recompileDelay: 300,
      }}
    >
      <SandpackLayout
        style={{
          height: "100%",
          borderRadius: 0,
          border: "none",
        }}
      >
        {showFileExplorer && (
          <SandpackFileExplorer
            style={{
              height: "100%",
              minWidth: "150px",
              maxWidth: "200px",
            }}
          />
        )}
        {showEditor && (
          <SandpackCodeEditor
            style={{
              height: "100%",
              minWidth: "300px",
            }}
            showLineNumbers
            showTabs
            closableTabs
            wrapContent
          />
        )}
        <SandpackPreviewPane
          style={{
            height: "100%",
            flex: 1,
          }}
          showNavigator
          showRefreshButton
          showOpenInCodeSandbox={false}
        />
        {showConsole && (
          <SandpackConsole
            style={{
              height: "150px",
            }}
          />
        )}
      </SandpackLayout>
    </SandpackProvider>
  );
}
