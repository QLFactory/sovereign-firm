"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { WebContainer } from "@webcontainer/api";
import FileBrowser from "./FileBrowser";
import CodeEditor from "./CodeEditor";

type TabType = "preview" | "files" | "console";

interface WebContainerPreviewProps {
  files: Record<string, string>;
  onTerminalOutput?: (line: string) => void;
  onFilesChange?: (files: Record<string, string>) => void;
}

// Singleton WebContainer instance - can only boot once per page
let webcontainerInstance: WebContainer | null = null;
let bootPromise: Promise<WebContainer> | null = null;

async function getWebContainer(): Promise<WebContainer> {
  if (webcontainerInstance) {
    return webcontainerInstance;
  }

  if (bootPromise) {
    return bootPromise;
  }

  bootPromise = WebContainer.boot().then((instance) => {
    webcontainerInstance = instance;
    return instance;
  });

  return bootPromise;
}

// Convert flat file paths to WebContainer's nested file structure
function convertToWebContainerFiles(files: Record<string, string>) {
  const result: Record<string, any> = {};

  for (const [filePath, content] of Object.entries(files)) {
    // Remove leading slash if present
    const normalizedPath = filePath.startsWith("/") ? filePath.slice(1) : filePath;
    const parts = normalizedPath.split("/");

    let current = result;
    for (let i = 0; i < parts.length - 1; i++) {
      const part = parts[i];
      if (!current[part]) {
        current[part] = { directory: {} };
      }
      current = current[part].directory;
    }

    const fileName = parts[parts.length - 1];
    current[fileName] = { file: { contents: content } };
  }

  return result;
}

// Default package.json for React + Vite + Testing
const defaultPackageJson = {
  name: "preview-app",
  private: true,
  version: "0.0.0",
  type: "module",
  scripts: {
    dev: "vite",
    build: "vite build",
    preview: "vite preview",
    test: "vitest run",
  },
  dependencies: {
    react: "^18.2.0",
    "react-dom": "^18.2.0",
  },
  devDependencies: {
    "@testing-library/jest-dom": "^6.4.2",
    "@testing-library/react": "^14.2.1",
    "@vitejs/plugin-react": "^4.2.1",
    jsdom: "^24.0.0",
    vite: "^5.1.0",
    vitest: "^1.3.1",
  },
};

// Default vite.config.js with vitest
const defaultViteConfig = `import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 5173,
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/setupTests.js',
  },
})`;

// Default test setup file
const defaultSetupTests = `import '@testing-library/jest-dom'`;

// Default index.html
const defaultIndexHtml = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Preview</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.jsx"></script>
  </body>
</html>`;

// Default main.jsx
const defaultMainJsx = `import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)`;

// Default App.jsx (waiting state)
const defaultAppJsx = `export default function App() {
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
      <h1 style={{ fontSize: '2rem', marginBottom: '1rem' }}>Waiting for Code...</h1>
      <p style={{ color: '#a1a1aa' }}>
        Describe your app in the chat, then type <code style={{
          background: '#27272a',
          padding: '0.25rem 0.5rem',
          borderRadius: '4px'
        }}>/approve</code> to generate.
      </p>
    </div>
  )
}`;

export default function WebContainerPreview({
  files,
  onTerminalOutput,
  onFilesChange,
}: WebContainerPreviewProps) {
  const [url, setUrl] = useState<string | null>(null);
  const [status, setStatus] = useState<string>("Initializing...");
  const [error, setError] = useState<string | null>(null);
  const [isReady, setIsReady] = useState(false);
  const [isRunningTests, setIsRunningTests] = useState(false);
  const [testResults, setTestResults] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>("preview");
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [localFiles, setLocalFiles] = useState<Record<string, string>>(files);
  const [consoleOutput, setConsoleOutput] = useState<string[]>([]);
  const containerRef = useRef<WebContainer | null>(null);
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const consoleRef = useRef<HTMLDivElement>(null);
  const isBootingRef = useRef(false);
  const lastFilesRef = useRef<string>("");

  // Sync local files with props
  useEffect(() => {
    setLocalFiles(files);
  }, [files]);

  const log = useCallback(
    (message: string) => {
      console.log(`[WebContainer] ${message}`);
      onTerminalOutput?.(message);
      setConsoleOutput((prev) => [...prev.slice(-500), message]); // Keep last 500 lines
    },
    [onTerminalOutput]
  );

  // Auto-scroll console
  useEffect(() => {
    if (consoleRef.current) {
      consoleRef.current.scrollTop = consoleRef.current.scrollHeight;
    }
  }, [consoleOutput]);

  // Handle file selection
  const handleSelectFile = useCallback((path: string) => {
    setSelectedFile(path);
    setActiveTab("files");
  }, []);

  // Handle file content change
  const handleFileChange = useCallback((content: string) => {
    if (!selectedFile) return;
    setLocalFiles((prev) => ({
      ...prev,
      [selectedFile]: content,
    }));
  }, [selectedFile]);

  // Save file to WebContainer
  const handleFileSave = useCallback(async () => {
    if (!selectedFile || !containerRef.current) return;

    const content = localFiles[selectedFile];
    if (!content) return;

    try {
      const normalizedPath = selectedFile.startsWith("/") ? selectedFile.slice(1) : selectedFile;
      await containerRef.current.fs.writeFile(normalizedPath, content);
      log(`Saved: ${selectedFile}`);
      onFilesChange?.(localFiles);
    } catch (err: any) {
      log(`Error saving file: ${err.message}`);
    }
  }, [selectedFile, localFiles, log, onFilesChange]);

  // Refresh the preview iframe
  const refreshPreview = useCallback(() => {
    if (iframeRef.current && url) {
      iframeRef.current.src = url;
      log("Preview refreshed");
    }
  }, [url, log]);

  // Run tests in WebContainer
  const runTests = useCallback(async () => {
    const container = containerRef.current;
    if (!container || isRunningTests) return;

    setIsRunningTests(true);
    setTestResults(null);
    log("\n🧪 Running tests...\n");

    try {
      const testProcess = await container.spawn("npm", ["test"]);

      let output = "";
      testProcess.output.pipeTo(
        new WritableStream({
          write(data) {
            output += data;
            log(data);
          },
        })
      );

      const exitCode = await testProcess.exit;

      if (exitCode === 0) {
        setTestResults("✅ All tests passed!");
        log("\n✅ All tests passed!\n");
      } else {
        setTestResults(`❌ Tests failed (exit code: ${exitCode})`);
        log(`\n❌ Tests failed (exit code: ${exitCode})\n`);
      }
    } catch (err: any) {
      const message = err?.message || "Failed to run tests";
      setTestResults(`❌ Error: ${message}`);
      log(`\n❌ Test error: ${message}\n`);
    } finally {
      setIsRunningTests(false);
    }
  }, [isRunningTests, log]);

  // Boot WebContainer once (uses singleton)
  useEffect(() => {
    if (isBootingRef.current || containerRef.current) return;
    isBootingRef.current = true;

    async function boot() {
      try {
        setStatus("Booting WebContainer...");
        log("Booting WebContainer...");

        const container = await getWebContainer();
        containerRef.current = container;

        // Listen for server-ready event
        container.on("server-ready", (port, serverUrl) => {
          log(`Server ready on port ${port}: ${serverUrl}`);
          setUrl(serverUrl);
          setStatus("Running");
        });

        container.on("error", (err) => {
          log(`Error: ${err.message}`);
          setError(err.message);
        });

        setStatus("WebContainer ready");
        log("WebContainer booted successfully");
        setIsReady(true);
      } catch (err: any) {
        const message = err?.message || "Failed to boot WebContainer";
        log(`Boot error: ${message}`);
        setError(message);
        setStatus("Error");
        isBootingRef.current = false;
      }
    }

    boot();

    return () => {
      // WebContainer persists as singleton for the page session
    };
  }, [log]);

  // Mount files and start dev server when files change or container becomes ready
  useEffect(() => {
    if (!isReady) return;
    const container = containerRef.current;
    if (!container) return;

    const filesJson = JSON.stringify(files);
    if (filesJson === lastFilesRef.current) return;
    lastFilesRef.current = filesJson;

    async function mountAndRun(wc: WebContainer) {
      try {
        // Prepare files with defaults
        const hasAppFile = Object.keys(files).some(
          (f) => f.includes("App.jsx") || f.includes("App.tsx") || f.includes("App.js")
        );

        let filesToMount = { ...files };

        // Add default files if missing
        if (!hasAppFile || Object.keys(files).length === 0) {
          filesToMount = {
            "/src/App.jsx": defaultAppJsx,
            "/src/main.jsx": defaultMainJsx,
          };
        }

        // Always use our package.json to ensure compatible dependencies
        // (LLM-generated package.json may have version conflicts)
        filesToMount["/package.json"] = JSON.stringify(defaultPackageJson, null, 2);

        // Always use our vite.config for consistent setup
        filesToMount["/vite.config.js"] = defaultViteConfig;

        // Ensure we have index.html
        if (!Object.keys(filesToMount).some((f) => f.includes("index.html"))) {
          filesToMount["/index.html"] = defaultIndexHtml;
        }

        // Ensure we have main.jsx if not present
        if (!Object.keys(filesToMount).some((f) => f.includes("main.jsx") || f.includes("main.tsx"))) {
          filesToMount["/src/main.jsx"] = defaultMainJsx;
        }

        // Add test setup file if there are test files
        const hasTestFiles = Object.keys(filesToMount).some((f) => f.includes(".test."));
        if (hasTestFiles && !Object.keys(filesToMount).some((f) => f.includes("setupTests"))) {
          filesToMount["/src/setupTests.js"] = defaultSetupTests;
        }

        setStatus("Mounting files...");
        log(`Mounting ${Object.keys(filesToMount).length} files...`);

        const wcFiles = convertToWebContainerFiles(filesToMount);
        await wc.mount(wcFiles);

        setStatus("Installing dependencies...");
        log("Running npm install...");

        const installProcess = await wc.spawn("npm", ["install"]);

        installProcess.output.pipeTo(
          new WritableStream({
            write(data) {
              log(data);
            },
          })
        );

        const installExitCode = await installProcess.exit;
        if (installExitCode !== 0) {
          throw new Error(`npm install failed with code ${installExitCode}`);
        }

        log("Dependencies installed");
        setStatus("Starting dev server...");
        log("Running npm run dev...");

        const devProcess = await wc.spawn("npm", ["run", "dev"]);

        devProcess.output.pipeTo(
          new WritableStream({
            write(data) {
              log(data);
            },
          })
        );

        // Server URL will be set via the server-ready event
      } catch (err: any) {
        const message = err?.message || "Failed to run project";
        log(`Error: ${message}`);
        setError(message);
        setStatus("Error");
      }
    }

    mountAndRun(container);
  }, [files, log, isReady]);

  if (error) {
    return (
      <div className="h-full flex items-center justify-center bg-zinc-950 text-red-400 p-4">
        <div className="text-center">
          <div className="text-4xl mb-4">⚠️</div>
          <p className="font-semibold mb-2">WebContainer Error</p>
          <p className="text-sm text-zinc-500">{error}</p>
          <p className="text-xs text-zinc-600 mt-4">
            WebContainers require a modern browser with SharedArrayBuffer support.
          </p>
        </div>
      </div>
    );
  }

  // Check if project has tests
  const hasTests = Object.keys(localFiles).some((f) => f.includes(".test."));
  const fileCount = Object.keys(localFiles).length;

  return (
    <div className="h-full flex flex-col bg-zinc-950">
      {/* Tab bar with status */}
      <div className="flex items-center justify-between bg-zinc-900 border-b border-zinc-800">
        {/* Tabs */}
        <div className="flex">
          <button
            onClick={() => setActiveTab("preview")}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "preview"
                ? "text-blue-400 border-blue-400 bg-zinc-800/50"
                : "text-zinc-500 border-transparent hover:text-zinc-300"
            }`}
          >
            Preview
          </button>
          <button
            onClick={() => setActiveTab("files")}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "files"
                ? "text-blue-400 border-blue-400 bg-zinc-800/50"
                : "text-zinc-500 border-transparent hover:text-zinc-300"
            }`}
          >
            Files
            <span className="ml-1.5 text-xs text-zinc-600">({fileCount})</span>
          </button>
          <button
            onClick={() => setActiveTab("console")}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === "console"
                ? "text-blue-400 border-blue-400 bg-zinc-800/50"
                : "text-zinc-500 border-transparent hover:text-zinc-300"
            }`}
          >
            Console
          </button>
        </div>

        {/* Status and actions */}
        <div className="flex items-center gap-2 px-3">
          <div
            className={`w-2 h-2 rounded-full ${
              url ? "bg-green-500" : "bg-yellow-500 animate-pulse"
            }`}
          />
          <span className="text-xs text-zinc-500">{status}</span>

          {/* Test results badge */}
          {testResults && (
            <span className={`px-2 py-0.5 rounded text-xs ${
              testResults.includes("✅") ? "bg-green-900 text-green-300" : "bg-red-900 text-red-300"
            }`}>
              {testResults}
            </span>
          )}

          {/* Refresh button */}
          {url && (
            <button
              onClick={refreshPreview}
              className="px-2 py-1 rounded text-xs font-medium transition-colors bg-zinc-700 hover:bg-zinc-600 text-zinc-300"
              title="Refresh preview"
            >
              🔄
            </button>
          )}

          {/* Run Tests button */}
          {hasTests && url && (
            <button
              onClick={runTests}
              disabled={isRunningTests}
              className={`px-3 py-1 rounded text-xs font-medium transition-colors ${
                isRunningTests
                  ? "bg-zinc-700 text-zinc-400 cursor-wait"
                  : "bg-purple-600 hover:bg-purple-500 text-white"
              }`}
            >
              {isRunningTests ? "Running..." : "🧪 Run Tests"}
            </button>
          )}
        </div>
      </div>

      {/* Tab content */}
      <div className="flex-1 overflow-hidden">
        {/* Preview tab */}
        {activeTab === "preview" && (
          <div className="h-full relative">
            {url ? (
              <iframe
                ref={iframeRef}
                src={url}
                className="w-full h-full border-0 bg-white"
                title="WebContainer Preview"
                allow="cross-origin-isolated"
              />
            ) : (
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="text-center">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4" />
                  <p className="text-zinc-400 text-sm">{status}</p>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Files tab */}
        {activeTab === "files" && (
          <div className="h-full flex">
            {/* File browser sidebar */}
            <div className="w-56 border-r border-zinc-800 flex-shrink-0">
              <FileBrowser
                files={localFiles}
                selectedFile={selectedFile}
                onSelectFile={handleSelectFile}
              />
            </div>
            {/* Code editor */}
            <div className="flex-1">
              <CodeEditor
                filePath={selectedFile}
                content={selectedFile ? localFiles[selectedFile] || "" : ""}
                onChange={handleFileChange}
                onSave={handleFileSave}
              />
            </div>
          </div>
        )}

        {/* Console tab */}
        {activeTab === "console" && (
          <div
            ref={consoleRef}
            className="h-full overflow-y-auto bg-black p-3 font-mono text-xs"
          >
            {consoleOutput.length === 0 ? (
              <div className="text-zinc-600">Console output will appear here...</div>
            ) : (
              consoleOutput.map((line, i) => (
                <div
                  key={i}
                  className={`whitespace-pre-wrap ${
                    line.includes("error") || line.includes("Error") || line.includes("ERR")
                      ? "text-red-400"
                      : line.includes("warning") || line.includes("WARN")
                      ? "text-yellow-400"
                      : line.includes("✅") || line.includes("success")
                      ? "text-green-400"
                      : "text-zinc-400"
                  }`}
                >
                  {line}
                </div>
              ))
            )}
          </div>
        )}
      </div>
    </div>
  );
}
