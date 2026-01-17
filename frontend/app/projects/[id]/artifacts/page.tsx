"use client";

import { useEffect, useState, use } from "react";
import Link from "next/link";
import { useAppStore } from "../../../lib/store";
import { api } from "../../../lib/api/client";
import {
  categorizeArtifacts,
  getAllFiles,
  getTotalFileCount,
  PHASE_COLORS,
} from "../../../lib/api/types";
import type { ArtifactCategory, ConsultancyState } from "../../../lib/api/types";

// Syntax highlighting helper (basic)
function getLanguageFromPath(path: string): string {
  const ext = path.split(".").pop()?.toLowerCase();
  switch (ext) {
    case "ts":
    case "tsx":
      return "typescript";
    case "js":
    case "jsx":
      return "javascript";
    case "json":
      return "json";
    case "yaml":
    case "yml":
      return "yaml";
    case "md":
      return "markdown";
    case "css":
    case "scss":
      return "css";
    case "html":
      return "html";
    case "sql":
      return "sql";
    case "go":
      return "go";
    case "py":
      return "python";
    case "rs":
      return "rust";
    case "tf":
      return "hcl";
    default:
      return "plaintext";
  }
}

// Category Card
function CategoryCard({
  category,
  isSelected,
  onClick,
}: {
  category: ArtifactCategory;
  isSelected: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={`w-full text-left p-4 rounded-xl transition-all ${
        isSelected
          ? "bg-[var(--cyan-glow)]/10 border-2 border-[var(--cyan-glow)]"
          : "bg-[var(--graphite)] border-2 border-transparent hover:bg-[var(--slate)]"
      }`}
    >
      <div className="flex items-center gap-3 mb-2">
        <span className="text-2xl">{category.icon}</span>
        <div className="flex-1">
          <div className="font-semibold text-[var(--ivory)]">{category.name}</div>
          <div className="text-xs text-[var(--silver)]">{category.fileCount} files</div>
        </div>
      </div>
      <p className="text-xs text-[var(--silver)]">{category.description}</p>
    </button>
  );
}

// File List Item
function FileListItem({
  filepath,
  isSelected,
  onClick,
}: {
  filepath: string;
  isSelected: boolean;
  onClick: () => void;
}) {
  const filename = filepath.split("/").pop() || filepath;
  const directory = filepath.includes("/")
    ? filepath.substring(0, filepath.lastIndexOf("/"))
    : "";

  return (
    <button
      onClick={onClick}
      className={`w-full text-left px-4 py-3 transition-colors flex items-center gap-3 ${
        isSelected
          ? "bg-[var(--cyan-glow)]/10 text-[var(--cyan-glow)]"
          : "text-[var(--pearl)] hover:bg-[var(--graphite)]"
      }`}
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
      <div className="flex-1 min-w-0">
        <div className="font-mono text-sm truncate">{filename}</div>
        {directory && (
          <div className="text-xs text-[var(--silver)] truncate">{directory}</div>
        )}
      </div>
    </button>
  );
}

// Code Viewer
function CodeViewer({
  filepath,
  content,
  onDownload,
}: {
  filepath: string;
  content: string;
  onDownload: () => void;
}) {
  const language = getLanguageFromPath(filepath);
  const lineCount = content.split("\n").length;

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--steel)] bg-[var(--graphite)]">
        <div className="flex items-center gap-3">
          <span className="font-mono text-sm text-[var(--ivory)]">{filepath}</span>
          <span className="text-xs text-[var(--silver)] px-2 py-0.5 bg-[var(--slate)] rounded">
            {language}
          </span>
          <span className="text-xs text-[var(--silver)]">{lineCount} lines</span>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => {
              navigator.clipboard.writeText(content);
            }}
            className="btn btn-ghost btn-sm"
            title="Copy to clipboard"
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
            </svg>
          </button>
          <button onClick={onDownload} className="btn btn-ghost btn-sm" title="Download file">
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
          </button>
        </div>
      </div>

      {/* Code */}
      <div className="flex-1 overflow-auto bg-[var(--obsidian)]">
        <pre className="p-4 font-mono text-sm">
          <code className="text-[var(--pearl)]">
            {content.split("\n").map((line, i) => (
              <div key={i} className="flex">
                <span className="w-12 pr-4 text-right text-[var(--steel)] select-none">
                  {i + 1}
                </span>
                <span className="flex-1 whitespace-pre-wrap break-all">{line}</span>
              </div>
            ))}
          </code>
        </pre>
      </div>
    </div>
  );
}

// Main Artifacts Content
function ArtifactsContent({ projectId }: { projectId: string }) {
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [isDownloading, setIsDownloading] = useState(false);

  // Store state
  const currentState = useAppStore((state) => state.currentState);
  const currentProject = useAppStore((state) => state.currentProject);
  const projects = useAppStore((state) => state.projects);
  const selectProject = useAppStore((state) => state.selectProject);
  const fetchProjectState = useAppStore((state) => state.fetchProjectState);
  const updateCurrentState = useAppStore((state) => state.updateCurrentState);

  // Initialize project
  useEffect(() => {
    const project = projects.find((p) => p.id === projectId);
    if (project) {
      selectProject(project);
    }
    fetchProjectState(projectId);
  }, [projectId, projects, selectProject, fetchProjectState]);

  // Get categories and files
  const categories = currentState ? categorizeArtifacts(currentState) : [];
  const allFiles = currentState ? getAllFiles(currentState) : {};
  const totalFileCount = currentState ? getTotalFileCount(currentState) : 0;

  // Get files for selected category
  const categoryFiles = selectedCategory
    ? categories.find((c) => c.id === selectedCategory)?.files || {}
    : allFiles;

  // Get selected file content
  const fileContent = selectedFile ? categoryFiles[selectedFile] || allFiles[selectedFile] : null;

  // Download handlers
  const handleDownloadFile = () => {
    if (selectedFile && fileContent) {
      const filename = selectedFile.split("/").pop() || selectedFile;
      api.downloadFile(filename, fileContent);
    }
  };

  const handleDownloadCategory = async () => {
    if (!selectedCategory) return;
    const category = categories.find((c) => c.id === selectedCategory);
    if (!category) return;

    setIsDownloading(true);
    try {
      await api.downloadAsZip(category.files, `${category.id}.zip`);
    } finally {
      setIsDownloading(false);
    }
  };

  const handleDownloadAll = async () => {
    setIsDownloading(true);
    try {
      await api.downloadAsZip(allFiles, `${currentProject?.name || "project"}.zip`);
    } finally {
      setIsDownloading(false);
    }
  };

  const phase = currentProject?.phase || "INTAKE";
  const phaseColor = PHASE_COLORS[phase] || "badge";

  return (
    <div className="min-h-screen bg-[var(--obsidian)]">
      {/* Header */}
      <header className="border-b border-[var(--steel)] bg-[var(--carbon)]">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link
              href={`/projects/${projectId}`}
              className="w-8 h-8 rounded-lg bg-[var(--graphite)] flex items-center justify-center text-[var(--silver)] hover:text-[var(--ivory)] transition-colors"
            >
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M19 12H5M12 19l-7-7 7-7" />
              </svg>
            </Link>
            <div>
              <div className="font-semibold text-[var(--ivory)]">
                {currentProject?.name || "Project"} - Artifacts
              </div>
              <div className="flex items-center gap-2 text-xs">
                <span className={`badge ${phaseColor}`}>{phase}</span>
                <span className="text-[var(--silver)]">
                  {totalFileCount} files generated
                </span>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-3">
            {selectedCategory && (
              <button
                onClick={handleDownloadCategory}
                className="btn btn-ghost"
                disabled={isDownloading}
              >
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                  <polyline points="7 10 12 15 17 10" />
                  <line x1="12" y1="15" x2="12" y2="3" />
                </svg>
                Download Category
              </button>
            )}
            <button
              onClick={handleDownloadAll}
              className="btn btn-primary"
              disabled={isDownloading || totalFileCount === 0}
            >
              {isDownloading ? (
                <>
                  <div className="spinner w-4 h-4" />
                  Downloading...
                </>
              ) : (
                <>
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                  >
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                    <polyline points="7 10 12 15 17 10" />
                    <line x1="12" y1="15" x2="12" y2="3" />
                  </svg>
                  Download All ({totalFileCount})
                </>
              )}
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex h-[calc(100vh-73px)]">
        {/* Left Sidebar - Categories */}
        <div className="w-64 border-r border-[var(--steel)] bg-[var(--carbon)] overflow-y-auto">
          <div className="p-4">
            <h3 className="text-xs font-semibold text-[var(--silver)] mb-3 uppercase tracking-wider">
              Categories
            </h3>

            {/* All Files Option */}
            <button
              onClick={() => {
                setSelectedCategory(null);
                setSelectedFile(null);
              }}
              className={`w-full text-left p-3 rounded-lg mb-2 transition-colors ${
                selectedCategory === null
                  ? "bg-[var(--cyan-glow)]/10 text-[var(--cyan-glow)]"
                  : "text-[var(--pearl)] hover:bg-[var(--graphite)]"
              }`}
            >
              <div className="flex items-center gap-2">
                <span className="text-lg">📦</span>
                <div>
                  <div className="font-medium">All Files</div>
                  <div className="text-xs text-[var(--silver)]">{totalFileCount} files</div>
                </div>
              </div>
            </button>

            <div className="border-t border-[var(--steel)] my-3" />

            {/* Category List */}
            <div className="space-y-2">
              {categories.map((category) => (
                <button
                  key={category.id}
                  onClick={() => {
                    setSelectedCategory(category.id);
                    setSelectedFile(null);
                  }}
                  className={`w-full text-left p-3 rounded-lg transition-colors ${
                    selectedCategory === category.id
                      ? "bg-[var(--cyan-glow)]/10 text-[var(--cyan-glow)]"
                      : "text-[var(--pearl)] hover:bg-[var(--graphite)]"
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <span className="text-lg">{category.icon}</span>
                    <div>
                      <div className="font-medium">{category.name}</div>
                      <div className="text-xs text-[var(--silver)]">
                        {category.fileCount} files
                      </div>
                    </div>
                  </div>
                </button>
              ))}
            </div>

            {categories.length === 0 && (
              <div className="text-center py-8 text-[var(--silver)]">
                <div className="text-3xl mb-2">📭</div>
                <p className="text-sm">No artifacts yet</p>
              </div>
            )}
          </div>
        </div>

        {/* Middle - File List */}
        <div className="w-80 border-r border-[var(--steel)] bg-[var(--carbon)] overflow-y-auto">
          <div className="sticky top-0 bg-[var(--carbon)] border-b border-[var(--steel)] px-4 py-3">
            <h3 className="text-xs font-semibold text-[var(--silver)] uppercase tracking-wider">
              {selectedCategory
                ? categories.find((c) => c.id === selectedCategory)?.name || "Files"
                : "All Files"}
            </h3>
          </div>
          <div className="divide-y divide-[var(--steel)]">
            {Object.keys(categoryFiles)
              .sort()
              .map((filepath) => (
                <FileListItem
                  key={filepath}
                  filepath={filepath}
                  isSelected={selectedFile === filepath}
                  onClick={() => setSelectedFile(filepath)}
                />
              ))}
          </div>
          {Object.keys(categoryFiles).length === 0 && (
            <div className="text-center py-8 text-[var(--silver)]">
              <p className="text-sm">No files in this category</p>
            </div>
          )}
        </div>

        {/* Right - Code Viewer */}
        <div className="flex-1 bg-[var(--obsidian)]">
          {selectedFile && fileContent ? (
            <CodeViewer
              filepath={selectedFile}
              content={fileContent}
              onDownload={handleDownloadFile}
            />
          ) : (
            <div className="h-full flex items-center justify-center text-[var(--silver)]">
              <div className="text-center">
                <div className="text-5xl mb-4">📄</div>
                <p className="text-lg mb-2">Select a file to view</p>
                <p className="text-sm">
                  {totalFileCount > 0
                    ? `${totalFileCount} files available`
                    : "No files generated yet"}
                </p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// Page Component
export default function ArtifactsPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const resolvedParams = use(params);
  return <ArtifactsContent projectId={resolvedParams.id} />;
}
