"use client";

import { useState, useMemo } from "react";

interface FileNode {
  name: string;
  path: string;
  type: "file" | "directory";
  children?: FileNode[] | Record<string, FileNode>;
}

interface FileBrowserProps {
  files: Record<string, string>;
  selectedFile: string | null;
  onSelectFile: (path: string) => void;
  onDeleteFile?: (path: string) => void;
}

// Build tree structure from flat file paths
function buildFileTree(files: Record<string, string>): FileNode[] {
  const root: Record<string, FileNode> = {};

  for (const filePath of Object.keys(files)) {
    const normalizedPath = filePath.startsWith("/") ? filePath.slice(1) : filePath;
    const parts = normalizedPath.split("/");

    let currentLevel = root;
    let currentPath = "";

    for (let i = 0; i < parts.length; i++) {
      const part = parts[i];
      currentPath = currentPath ? `${currentPath}/${part}` : part;
      const isFile = i === parts.length - 1;

      if (!currentLevel[part]) {
        currentLevel[part] = {
          name: part,
          path: `/${currentPath}`,
          type: isFile ? "file" : "directory",
          children: isFile ? undefined : {},
        };
      }

      if (!isFile) {
        currentLevel = currentLevel[part].children as Record<string, FileNode>;
      }
    }
  }

  // Convert object to sorted array
  function toArray(obj: Record<string, FileNode>): FileNode[] {
    return Object.values(obj)
      .map((node) => ({
        ...node,
        children: node.children && !Array.isArray(node.children)
          ? toArray(node.children as Record<string, FileNode>)
          : node.children as FileNode[] | undefined,
      }))
      .sort((a, b) => {
        // Directories first, then alphabetically
        if (a.type !== b.type) return a.type === "directory" ? -1 : 1;
        return a.name.localeCompare(b.name);
      });
  }

  return toArray(root);
}

// Get file icon based on extension
function getFileIcon(name: string): string {
  const ext = name.split(".").pop()?.toLowerCase();
  switch (ext) {
    case "js":
    case "jsx":
      return "📄";
    case "ts":
    case "tsx":
      return "📘";
    case "css":
      return "🎨";
    case "html":
      return "🌐";
    case "json":
      return "📋";
    case "md":
      return "📝";
    case "test.js":
    case "test.jsx":
    case "test.ts":
    case "test.tsx":
      return "🧪";
    default:
      if (name.includes(".test.")) return "🧪";
      return "📄";
  }
}

function FileTreeNode({
  node,
  depth,
  selectedFile,
  onSelectFile,
  expandedDirs,
  toggleDir,
}: {
  node: FileNode;
  depth: number;
  selectedFile: string | null;
  onSelectFile: (path: string) => void;
  expandedDirs: Set<string>;
  toggleDir: (path: string) => void;
}) {
  const isExpanded = expandedDirs.has(node.path);
  const isSelected = selectedFile === node.path;

  if (node.type === "directory") {
    return (
      <div>
        <button
          onClick={() => toggleDir(node.path)}
          className={`w-full flex items-center gap-1 px-2 py-1 text-left text-sm hover:bg-zinc-800 rounded ${
            isExpanded ? "text-zinc-200" : "text-zinc-400"
          }`}
          style={{ paddingLeft: `${depth * 12 + 8}px` }}
        >
          <span className="text-xs">{isExpanded ? "📂" : "📁"}</span>
          <span className="truncate">{node.name}</span>
        </button>
        {isExpanded && node.children && Array.isArray(node.children) && (
          <div>
            {(node.children as FileNode[]).map((child) => (
              <FileTreeNode
                key={child.path}
                node={child}
                depth={depth + 1}
                selectedFile={selectedFile}
                onSelectFile={onSelectFile}
                expandedDirs={expandedDirs}
                toggleDir={toggleDir}
              />
            ))}
          </div>
        )}
      </div>
    );
  }

  return (
    <button
      onClick={() => onSelectFile(node.path)}
      className={`w-full flex items-center gap-1 px-2 py-1 text-left text-sm rounded truncate ${
        isSelected
          ? "bg-blue-600 text-white"
          : "text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
      }`}
      style={{ paddingLeft: `${depth * 12 + 8}px` }}
    >
      <span className="text-xs">{getFileIcon(node.name)}</span>
      <span className="truncate">{node.name}</span>
    </button>
  );
}

export default function FileBrowser({
  files,
  selectedFile,
  onSelectFile,
}: FileBrowserProps) {
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(
    new Set(["/src"]) // Expand src by default
  );

  const fileTree = useMemo(() => buildFileTree(files), [files]);

  const toggleDir = (path: string) => {
    setExpandedDirs((prev) => {
      const next = new Set(prev);
      if (next.has(path)) {
        next.delete(path);
      } else {
        next.add(path);
      }
      return next;
    });
  };

  const fileCount = Object.keys(files).length;

  return (
    <div className="h-full flex flex-col bg-zinc-950">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-zinc-800">
        <span className="text-xs font-medium text-zinc-400 uppercase tracking-wider">
          Files
        </span>
        <span className="text-xs text-zinc-600">{fileCount} files</span>
      </div>

      {/* File tree */}
      <div className="flex-1 overflow-y-auto py-1">
        {fileTree.length === 0 ? (
          <div className="px-3 py-4 text-center text-zinc-600 text-sm">
            No files yet
          </div>
        ) : (
          fileTree.map((node) => (
            <FileTreeNode
              key={node.path}
              node={node}
              depth={0}
              selectedFile={selectedFile}
              onSelectFile={onSelectFile}
              expandedDirs={expandedDirs}
              toggleDir={toggleDir}
            />
          ))
        )}
      </div>
    </div>
  );
}
