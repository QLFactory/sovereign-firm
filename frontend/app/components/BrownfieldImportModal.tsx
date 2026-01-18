"use client";

import { useState, useEffect, useCallback } from "react";
import { api } from "../lib/api/client";

interface BrownfieldImportModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: (projectId: string) => void;
}

type ImportSource = "git" | "local";
type ImportStatus = "idle" | "analyzing" | "success" | "error";

interface ImportProgress {
  files: number;
  chunks: number;
  symbols: number;
}

export default function BrownfieldImportModal({
  isOpen,
  onClose,
  onSuccess,
}: BrownfieldImportModalProps) {
  const [source, setSource] = useState<ImportSource>("git");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [repoUrl, setRepoUrl] = useState("");
  const [localPath, setLocalPath] = useState("");
  const [status, setStatus] = useState<ImportStatus>("idle");
  const [progress, setProgress] = useState<ImportProgress>({
    files: 0,
    chunks: 0,
    symbols: 0,
  });
  const [error, setError] = useState("");
  const [workflowId, setWorkflowId] = useState("");
  const [projectId, setProjectId] = useState("");

  // Reset form when modal closes
  useEffect(() => {
    if (!isOpen) {
      setSource("git");
      setName("");
      setDescription("");
      setRepoUrl("");
      setLocalPath("");
      setStatus("idle");
      setProgress({ files: 0, chunks: 0, symbols: 0 });
      setError("");
      setWorkflowId("");
      setProjectId("");
    }
  }, [isOpen]);

  // Validate Git URL format
  const isValidGitUrl = useCallback((url: string): boolean => {
    if (!url) return false;
    return (
      url.startsWith("http://") ||
      url.startsWith("https://") ||
      url.startsWith("git@")
    );
  }, []);

  // Check if form is valid
  const isFormValid = useCallback((): boolean => {
    if (!name.trim()) return false;
    if (source === "git") {
      return isValidGitUrl(repoUrl);
    }
    return localPath.trim().length > 0;
  }, [name, source, repoUrl, localPath, isValidGitUrl]);

  // Poll for import status
  const pollStatus = useCallback(
    async (projId: string) => {
      try {
        const statusData = await api.getImportStatus(projId);

        setProgress({
          files: statusData.files_indexed || 0,
          chunks: statusData.chunks_created || 0,
          symbols: statusData.symbols_found || 0,
        });

        if (statusData.status === "complete") {
          setStatus("success");
          setTimeout(() => {
            onClose();
            if (onSuccess) {
              onSuccess(projId);
            } else {
              window.location.href = `/projects/${projId}`;
            }
          }, 1500);
        } else if (statusData.status === "error") {
          setStatus("error");
          setError(statusData.error || "Import failed");
        } else {
          // Continue polling
          setTimeout(() => pollStatus(projId), 2000);
        }
      } catch (err) {
        console.error("Poll error:", err);
        // Retry on network errors
        setTimeout(() => pollStatus(projId), 3000);
      }
    },
    [onClose, onSuccess]
  );

  // Handle import submission
  const handleImport = async () => {
    setStatus("analyzing");
    setError("");

    try {
      const data = await api.importBrownfield({
        name,
        description,
        repo_url: source === "git" ? repoUrl : undefined,
        local_path: source === "local" ? localPath : undefined,
      });

      setWorkflowId(data.workflow_id);
      setProjectId(data.project_id);

      // Start polling for status
      pollStatus(data.project_id);
    } catch (err) {
      setStatus("error");
      setError(err instanceof Error ? err.message : "Import failed");
    }
  };

  // Handle escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen && status === "idle") {
        onClose();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose, status]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/70 backdrop-blur-sm"
        onClick={status === "idle" ? onClose : undefined}
      />

      {/* Modal */}
      <div className="relative glass-card w-full max-w-lg p-8 m-4 rounded-2xl animate-scale-in">
        {/* Close button */}
        {status === "idle" && (
          <button
            onClick={onClose}
            className="absolute top-4 right-4 text-[var(--silver)] hover:text-[var(--ivory)] transition-colors"
            aria-label="Close modal"
          >
            <svg
              className="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        )}

        <h2 className="text-display-sm text-gradient-cyan mb-6">
          Import Existing Project
        </h2>

        {/* Idle State - Form */}
        {status === "idle" && (
          <>
            {/* Source Selection */}
            <div className="flex gap-3 mb-6">
              <button
                onClick={() => setSource("git")}
                className={`flex-1 p-4 rounded-xl border transition-all ${
                  source === "git"
                    ? "border-[var(--cyan-glow)] bg-[rgba(0,240,255,0.1)]"
                    : "border-[var(--steel)] hover:border-[var(--silver)]"
                }`}
              >
                <svg
                  className="w-6 h-6 mx-auto mb-2 text-[var(--cyan-glow)]"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M13 10V3L4 14h7v7l9-11h-7z"
                  />
                </svg>
                <span className="text-body-sm text-[var(--pearl)]">
                  Git Repository
                </span>
              </button>

              <button
                onClick={() => setSource("local")}
                className={`flex-1 p-4 rounded-xl border transition-all ${
                  source === "local"
                    ? "border-[var(--cyan-glow)] bg-[rgba(0,240,255,0.1)]"
                    : "border-[var(--steel)] hover:border-[var(--silver)]"
                }`}
              >
                <svg
                  className="w-6 h-6 mx-auto mb-2 text-[var(--cyan-glow)]"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                  />
                </svg>
                <span className="text-body-sm text-[var(--pearl)]">
                  Local Directory
                </span>
              </button>
            </div>

            {/* Form Fields */}
            <div className="space-y-4">
              <div>
                <label className="input-label">Project Name *</label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="My Legacy App"
                  className="input"
                  name="name"
                />
              </div>

              <div>
                <label className="input-label">Description</label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Describe what this project does..."
                  className="input h-20 resize-none"
                  name="description"
                />
              </div>

              {source === "git" ? (
                <div>
                  <label className="input-label">Repository URL *</label>
                  <input
                    type="text"
                    value={repoUrl}
                    onChange={(e) => setRepoUrl(e.target.value)}
                    placeholder="https://github.com/org/repo.git"
                    className={`input ${
                      repoUrl && !isValidGitUrl(repoUrl)
                        ? "border-[var(--rose-glow)]"
                        : ""
                    }`}
                    name="repo_url"
                  />
                  {repoUrl && !isValidGitUrl(repoUrl) && (
                    <p className="text-xs text-[var(--rose-glow)] mt-1">
                      Please enter a valid Git URL (https:// or git@)
                    </p>
                  )}
                </div>
              ) : (
                <div>
                  <label className="input-label">Local Path *</label>
                  <input
                    type="text"
                    value={localPath}
                    onChange={(e) => setLocalPath(e.target.value)}
                    placeholder="/path/to/project"
                    className="input"
                    name="local_path"
                  />
                </div>
              )}
            </div>

            <button
              onClick={handleImport}
              disabled={!isFormValid()}
              className="btn btn-primary w-full mt-6 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
            >
              <svg
                className="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
                />
              </svg>
              Start Import
            </button>
          </>
        )}

        {/* Analyzing State */}
        {status === "analyzing" && (
          <div className="text-center py-8">
            <div className="w-12 h-12 mx-auto mb-4 spinner" />
            <h3 className="text-lg font-semibold text-[var(--ivory)] mb-2">
              Analyzing Codebase
            </h3>
            <p className="text-body-sm text-[var(--silver)] mb-6">
              Parsing files, extracting patterns, and building knowledge graph...
            </p>

            <div className="grid grid-cols-3 gap-4 text-center">
              <div className="p-3 glass rounded-xl">
                <div className="text-2xl font-bold text-[var(--cyan-glow)]">
                  {progress.files}
                </div>
                <div className="text-caption text-[var(--silver)]">
                  Files Indexed
                </div>
              </div>
              <div className="p-3 glass rounded-xl">
                <div className="text-2xl font-bold text-[var(--violet-glow)]">
                  {progress.chunks}
                </div>
                <div className="text-caption text-[var(--silver)]">
                  Code Chunks
                </div>
              </div>
              <div className="p-3 glass rounded-xl">
                <div className="text-2xl font-bold text-[var(--emerald-glow)]">
                  {progress.symbols}
                </div>
                <div className="text-caption text-[var(--silver)]">
                  Symbols Found
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Success State */}
        {status === "success" && (
          <div className="text-center py-8">
            <div className="w-12 h-12 mx-auto mb-4 rounded-full bg-[rgba(16,185,129,0.2)] flex items-center justify-center">
              <svg
                className="w-6 h-6 text-[var(--emerald-glow)]"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-[var(--ivory)] mb-2">
              Import Complete!
            </h3>
            <p className="text-body-sm text-[var(--silver)]">
              Redirecting to workspace...
            </p>

            <div className="grid grid-cols-3 gap-4 text-center mt-6">
              <div className="p-3 glass rounded-xl">
                <div className="text-xl font-bold text-[var(--cyan-glow)]">
                  {progress.files}
                </div>
                <div className="text-caption text-[var(--silver)]">Files</div>
              </div>
              <div className="p-3 glass rounded-xl">
                <div className="text-xl font-bold text-[var(--violet-glow)]">
                  {progress.chunks}
                </div>
                <div className="text-caption text-[var(--silver)]">Chunks</div>
              </div>
              <div className="p-3 glass rounded-xl">
                <div className="text-xl font-bold text-[var(--emerald-glow)]">
                  {progress.symbols}
                </div>
                <div className="text-caption text-[var(--silver)]">Symbols</div>
              </div>
            </div>
          </div>
        )}

        {/* Error State */}
        {status === "error" && (
          <div className="text-center py-8">
            <div className="w-12 h-12 mx-auto mb-4 rounded-full bg-[rgba(244,63,94,0.2)] flex items-center justify-center">
              <svg
                className="w-6 h-6 text-[var(--rose-glow)]"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-[var(--ivory)] mb-2">
              Import Failed
            </h3>
            <p className="text-body-sm text-[var(--rose-glow)] mb-4">{error}</p>
            <button onClick={() => setStatus("idle")} className="btn btn-secondary">
              Try Again
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
