"use client";

import { useState, useEffect, useCallback, useRef } from "react";

interface CodeEditorProps {
  filePath: string | null;
  content: string;
  onChange: (content: string) => void;
  onSave: () => void;
  readOnly?: boolean;
}

// Get language from file extension
function getLanguage(filePath: string): string {
  const ext = filePath.split(".").pop()?.toLowerCase();
  switch (ext) {
    case "js":
    case "jsx":
      return "javascript";
    case "ts":
    case "tsx":
      return "typescript";
    case "css":
      return "css";
    case "html":
      return "html";
    case "json":
      return "json";
    case "md":
      return "markdown";
    default:
      return "plaintext";
  }
}

// Simple syntax highlighting colors
function getLanguageColor(lang: string): string {
  switch (lang) {
    case "javascript":
    case "typescript":
      return "text-yellow-400";
    case "css":
      return "text-pink-400";
    case "html":
      return "text-orange-400";
    case "json":
      return "text-green-400";
    default:
      return "text-zinc-300";
  }
}

export default function CodeEditor({
  filePath,
  content,
  onChange,
  onSave,
  readOnly = false,
}: CodeEditorProps) {
  const [localContent, setLocalContent] = useState(content);
  const [isDirty, setIsDirty] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Sync with external content
  useEffect(() => {
    setLocalContent(content);
    setIsDirty(false);
  }, [content, filePath]);

  // Handle keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Cmd/Ctrl + S to save
      if ((e.metaKey || e.ctrlKey) && e.key === "s") {
        e.preventDefault();
        if (isDirty) {
          onChange(localContent);
          onSave();
          setIsDirty(false);
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isDirty, localContent, onChange, onSave]);

  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      const newContent = e.target.value;
      setLocalContent(newContent);
      setIsDirty(newContent !== content);
    },
    [content]
  );

  const handleSave = useCallback(() => {
    onChange(localContent);
    onSave();
    setIsDirty(false);
  }, [localContent, onChange, onSave]);

  // Handle tab key for indentation
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === "Tab") {
        e.preventDefault();
        const textarea = textareaRef.current;
        if (!textarea) return;

        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const newContent =
          localContent.substring(0, start) +
          "  " +
          localContent.substring(end);

        setLocalContent(newContent);
        setIsDirty(true);

        // Set cursor position after the inserted spaces
        requestAnimationFrame(() => {
          textarea.selectionStart = textarea.selectionEnd = start + 2;
        });
      }
    },
    [localContent]
  );

  if (!filePath) {
    return (
      <div className="h-full flex items-center justify-center bg-zinc-950 text-zinc-600">
        <div className="text-center">
          <div className="text-4xl mb-2">📄</div>
          <p className="text-sm">Select a file to edit</p>
        </div>
      </div>
    );
  }

  const language = getLanguage(filePath);
  const languageColor = getLanguageColor(language);
  const lineCount = localContent.split("\n").length;

  return (
    <div className="h-full flex flex-col bg-zinc-950">
      {/* Editor header */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-zinc-800 bg-zinc-900">
        <div className="flex items-center gap-2">
          <span className="text-sm text-zinc-300 font-mono">
            {filePath.split("/").pop()}
          </span>
          {isDirty && (
            <span className="w-2 h-2 rounded-full bg-yellow-500" title="Unsaved changes" />
          )}
          <span className={`text-xs px-1.5 py-0.5 rounded ${languageColor} bg-zinc-800`}>
            {language}
          </span>
        </div>
        <div className="flex items-center gap-2">
          {isDirty && !readOnly && (
            <button
              onClick={handleSave}
              className="px-2 py-1 text-xs bg-blue-600 hover:bg-blue-500 text-white rounded"
            >
              Save
            </button>
          )}
          <span className="text-xs text-zinc-600">
            {lineCount} lines
          </span>
        </div>
      </div>

      {/* Editor content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Line numbers */}
        <div className="w-12 bg-zinc-900 text-zinc-600 text-xs font-mono text-right py-3 pr-2 overflow-hidden select-none border-r border-zinc-800">
          {Array.from({ length: lineCount }, (_, i) => (
            <div key={i} className="leading-5">
              {i + 1}
            </div>
          ))}
        </div>

        {/* Code textarea */}
        <textarea
          ref={textareaRef}
          value={localContent}
          onChange={handleChange}
          onKeyDown={handleKeyDown}
          readOnly={readOnly}
          spellCheck={false}
          className={`flex-1 bg-zinc-950 text-zinc-300 font-mono text-sm p-3 resize-none outline-none leading-5 ${
            readOnly ? "cursor-default" : ""
          }`}
          style={{ tabSize: 2 }}
        />
      </div>

      {/* Status bar */}
      <div className="flex items-center justify-between px-3 py-1 border-t border-zinc-800 bg-zinc-900 text-xs text-zinc-500">
        <span>{filePath}</span>
        <div className="flex items-center gap-3">
          {!readOnly && <span>Cmd+S to save</span>}
          <span>UTF-8</span>
        </div>
      </div>
    </div>
  );
}
