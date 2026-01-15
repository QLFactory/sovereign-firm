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
function convertToSandpackFiles(files: Record<string, string>): Record<string, string> {
  const sandpackFiles: Record<string, string> = {};

  for (const [path, content] of Object.entries(files)) {
    // Sandpack uses paths without leading slash for some files
    const normalizedPath = path.startsWith("/") ? path : `/${path}`;
    sandpackFiles[normalizedPath] = content;
  }

  return sandpackFiles;
}

// Default files for empty state
const defaultFiles: Record<string, string> = {
  "/src/App.jsx": `export default function App() {
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
}`,
  "/src/main.jsx": `import React from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';

const root = createRoot(document.getElementById('root'));
root.render(<App />);`,
  "/index.html": `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Sovereign Preview</title>
</head>
<body style="margin: 0; padding: 0;">
  <div id="root"></div>
</body>
</html>`,
  "/package.json": JSON.stringify({
    dependencies: {
      react: "^18.2.0",
      "react-dom": "^18.2.0",
    },
  }, null, 2),
};

export default function SandpackPreview({
  files,
  activeFile,
  showFileExplorer = true,
  showConsole = false,
  showEditor = true,
  theme = "dark",
}: SandpackPreviewProps) {
  // Merge default files with provided files
  const sandpackFiles = useMemo(() => {
    const hasAppFile = Object.keys(files).some(
      (f) => f.includes("App.jsx") || f.includes("App.tsx") || f.includes("App.js")
    );

    if (!hasAppFile || Object.keys(files).length === 0) {
      return convertToSandpackFiles(defaultFiles);
    }

    // Ensure we have required files
    const merged = { ...defaultFiles, ...files };
    return convertToSandpackFiles(merged);
  }, [files]);

  // Determine active file
  const visibleFile = useMemo(() => {
    if (activeFile) return activeFile;
    // Find the main App file
    const appFile = Object.keys(sandpackFiles).find(
      (f) => f.includes("App.jsx") || f.includes("App.tsx") || f.includes("App.js")
    );
    return appFile || "/src/App.jsx";
  }, [activeFile, sandpackFiles]);

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
        recompileDelay: 500,
      }}
      customSetup={{
        dependencies: {
          react: "^18.2.0",
          "react-dom": "^18.2.0",
        },
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
