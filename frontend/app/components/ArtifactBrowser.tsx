"use client";

import { useState, useMemo, ReactNode } from "react";

interface ArtifactBrowserProps {
  state: ConsultancyState | null;
  onFileSelect?: (path: string, content: string, category: string) => void;
}

interface ConsultancyState {
  project_name: string;
  phase: string;
  frontend_code: Record<string, string>;
  backend_code: Record<string, string>;
  database_code: Record<string, string>;
  unit_tests: Record<string, string>;
  integration_tests: Record<string, string>;
  e2e_tests: Record<string, string>;
  dockerfile: string;
  docker_compose: string;
  ci_pipeline: string;
  kube_manifests: Record<string, string>;
  helm_chart: Record<string, string>;
  infra_code: Record<string, string>;
  monitoring_config: Record<string, unknown>;
  alert_rules: Record<string, unknown>;
  dashboards: Record<string, unknown>;
  runbooks: Record<string, string>;
}

interface ArtifactCategory {
  id: string;
  name: string;
  icon: ReactNode;
  color: string;
  files: { path: string; content: string }[];
}

export default function ArtifactBrowser({ state, onFileSelect }: ArtifactBrowserProps) {
  const [expandedCategory, setExpandedCategory] = useState<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<{ path: string; content: string; category: string } | null>(null);
  const [searchQuery, setSearchQuery] = useState("");

  // Organize artifacts into categories
  const categories = useMemo<ArtifactCategory[]>(() => {
    if (!state) return [];

    const toFiles = (obj: Record<string, string> | undefined | null) =>
      Object.entries(obj || {}).map(([path, content]) => ({ path, content }));

    return [
      {
        id: "frontend",
        name: "Frontend",
        icon: (
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="3" y="3" width="18" height="18" rx="2" />
            <path d="M3 9h18" />
            <path d="M9 21V9" />
          </svg>
        ),
        color: "cyan",
        files: toFiles(state.frontend_code),
      },
      {
        id: "backend",
        name: "Backend",
        icon: (
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="2" y="2" width="20" height="8" rx="2" />
            <rect x="2" y="14" width="20" height="8" rx="2" />
            <path d="M6 6h.01M6 18h.01" />
          </svg>
        ),
        color: "violet",
        files: toFiles(state.backend_code),
      },
      {
        id: "database",
        name: "Database",
        icon: (
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <ellipse cx="12" cy="5" rx="9" ry="3" />
            <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
            <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
          </svg>
        ),
        color: "emerald",
        files: toFiles(state.database_code),
      },
      {
        id: "tests",
        name: "Tests",
        icon: (
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
            <polyline points="14 2 14 8 20 8" />
            <path d="m9 15 2 2 4-4" />
          </svg>
        ),
        color: "amber",
        files: [
          ...toFiles(state.unit_tests),
          ...toFiles(state.integration_tests),
          ...toFiles(state.e2e_tests),
        ],
      },
      {
        id: "docker",
        name: "Docker",
        icon: (
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M22 12.5c-1.5 0-2-.5-2-2 0 2-1.5 2.5-3.5 2.5H2V9h4V5h4v4h4V5h4v4h2.5c1 0 1.5.5 1.5 1.5v2" />
          </svg>
        ),
        color: "cyan",
        files: [
          ...(state.dockerfile ? [{ path: "Dockerfile", content: state.dockerfile }] : []),
          ...(state.docker_compose ? [{ path: "docker-compose.yaml", content: state.docker_compose }] : []),
        ],
      },
      {
        id: "kubernetes",
        name: "Kubernetes",
        icon: (
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 2L2 7l10 5 10-5-10-5z" />
            <path d="M2 17l10 5 10-5" />
            <path d="M2 12l10 5 10-5" />
          </svg>
        ),
        color: "violet",
        files: toFiles(state.kube_manifests),
      },
      {
        id: "helm",
        name: "Helm Charts",
        icon: (
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="10" />
            <path d="M12 2a7 7 0 1 0 7 7" />
            <path d="M12 8v4l2 2" />
          </svg>
        ),
        color: "emerald",
        files: toFiles(state.helm_chart),
      },
      {
        id: "ci",
        name: "CI/CD",
        icon: (
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="3" />
            <path d="M12 1v6M12 17v6M4.22 4.22l4.24 4.24M15.54 15.54l4.24 4.24M1 12h6M17 12h6M4.22 19.78l4.24-4.24M15.54 8.46l4.24-4.24" />
          </svg>
        ),
        color: "amber",
        files: state.ci_pipeline ? [{ path: ".github/workflows/ci.yaml", content: state.ci_pipeline }] : [],
      },
      {
        id: "infrastructure",
        name: "Infrastructure",
        icon: (
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
          </svg>
        ),
        color: "rose",
        files: toFiles(state.infra_code),
      },
      {
        id: "monitoring",
        name: "Monitoring & SRE",
        icon: (
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M3 3v18h18" />
            <path d="m19 9-5 5-4-4-3 3" />
          </svg>
        ),
        color: "cyan",
        files: [
          ...toFiles(state.runbooks),
          ...(state.monitoring_config?.files
            ? Object.entries(state.monitoring_config.files as Record<string, string>).map(([path, content]) => ({ path, content }))
            : []),
        ],
      },
    ].filter((cat) => cat.files.length > 0);
  }, [state]);

  // Filter files by search
  const filteredCategories = useMemo(() => {
    if (!searchQuery.trim()) return categories;

    const query = searchQuery.toLowerCase();
    return categories
      .map((cat) => ({
        ...cat,
        files: cat.files.filter(
          (f) =>
            f.path.toLowerCase().includes(query) ||
            f.content.toLowerCase().includes(query)
        ),
      }))
      .filter((cat) => cat.files.length > 0);
  }, [categories, searchQuery]);

  // Calculate total files
  const totalFiles = useMemo(
    () => categories.reduce((sum, cat) => sum + cat.files.length, 0),
    [categories]
  );

  const colorClasses: Record<string, { bg: string; text: string; border: string }> = {
    cyan: {
      bg: "bg-[rgba(0,240,255,0.1)]",
      text: "text-[var(--cyan-glow)]",
      border: "border-[var(--cyan-glow)]",
    },
    violet: {
      bg: "bg-[rgba(168,85,247,0.1)]",
      text: "text-[var(--violet-glow)]",
      border: "border-[var(--violet-glow)]",
    },
    emerald: {
      bg: "bg-[rgba(16,185,129,0.1)]",
      text: "text-[var(--emerald-glow)]",
      border: "border-[var(--emerald-glow)]",
    },
    amber: {
      bg: "bg-[rgba(255,184,0,0.1)]",
      text: "text-[var(--amber-glow)]",
      border: "border-[var(--amber-glow)]",
    },
    rose: {
      bg: "bg-[rgba(244,63,94,0.1)]",
      text: "text-[var(--rose-glow)]",
      border: "border-[var(--rose-glow)]",
    },
  };

  const handleFileClick = (file: { path: string; content: string }, categoryId: string) => {
    setSelectedFile({ ...file, category: categoryId });
    onFileSelect?.(file.path, file.content, categoryId);
  };

  const handleDownloadAll = () => {
    // Create a ZIP-like structure (for demo, just downloads as JSON)
    const allFiles: Record<string, string> = {};
    categories.forEach((cat) => {
      cat.files.forEach((f) => {
        allFiles[f.path] = f.content;
      });
    });

    const blob = new Blob([JSON.stringify(allFiles, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${state?.project_name || "project"}-artifacts.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  if (!state) {
    return (
      <div className="p-6 text-center">
        <div className="text-[var(--silver)]">No artifacts available yet</div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col bg-[var(--carbon)]">
      {/* Header */}
      <div className="p-4 border-b border-[var(--steel)]">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-lg font-semibold text-[var(--ivory)]">Generated Artifacts</h2>
            <p className="text-sm text-[var(--silver)]">
              {totalFiles} files across {categories.length} categories
            </p>
          </div>
          <button onClick={handleDownloadAll} className="btn btn-secondary text-sm">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            Download All
          </button>
        </div>

        {/* Search */}
        <div className="relative">
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="var(--silver)"
            strokeWidth="2"
            className="absolute left-3 top-1/2 -translate-y-1/2"
          >
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            type="text"
            className="input pl-10"
            placeholder="Search files..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      {/* Categories */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {filteredCategories.map((category) => (
          <div key={category.id} className="rounded-xl overflow-hidden border border-[var(--steel)]">
            {/* Category Header */}
            <button
              onClick={() =>
                setExpandedCategory(expandedCategory === category.id ? null : category.id)
              }
              className={`w-full flex items-center justify-between p-4 transition-colors hover:bg-[var(--graphite)] ${
                expandedCategory === category.id ? "bg-[var(--graphite)]" : ""
              }`}
            >
              <div className="flex items-center gap-3">
                <div
                  className={`w-10 h-10 rounded-lg ${colorClasses[category.color].bg} ${colorClasses[category.color].text} flex items-center justify-center`}
                >
                  {category.icon}
                </div>
                <div className="text-left">
                  <div className="font-medium text-[var(--ivory)]">{category.name}</div>
                  <div className="text-xs text-[var(--silver)]">{category.files.length} files</div>
                </div>
              </div>
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="var(--silver)"
                strokeWidth="2"
                className={`transition-transform ${expandedCategory === category.id ? "rotate-180" : ""}`}
              >
                <path d="m6 9 6 6 6-6" />
              </svg>
            </button>

            {/* Files List */}
            {expandedCategory === category.id && (
              <div className="border-t border-[var(--steel)] bg-[var(--void)]/50">
                {category.files.map((file) => (
                  <button
                    key={file.path}
                    onClick={() => handleFileClick(file, category.id)}
                    className={`w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-[var(--graphite)] transition-colors ${
                      selectedFile?.path === file.path ? "bg-[var(--graphite)]" : ""
                    }`}
                  >
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="var(--silver)"
                      strokeWidth="2"
                    >
                      <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
                      <polyline points="14 2 14 8 20 8" />
                    </svg>
                    <span className="flex-1 font-mono text-sm text-[var(--pearl)] truncate">
                      {file.path}
                    </span>
                    <span className="text-xs text-[var(--silver)]">
                      {(file.content.length / 1024).toFixed(1)}KB
                    </span>
                  </button>
                ))}
              </div>
            )}
          </div>
        ))}

        {filteredCategories.length === 0 && (
          <div className="text-center py-12">
            <div className="text-[var(--silver)]">No files match your search</div>
          </div>
        )}
      </div>

      {/* File Preview */}
      {selectedFile && (
        <div className="border-t border-[var(--steel)] bg-[var(--void)]">
          <div className="flex items-center justify-between px-4 py-2 border-b border-[var(--steel)]">
            <span className="font-mono text-sm text-[var(--pearl)]">{selectedFile.path}</span>
            <button
              onClick={() => setSelectedFile(null)}
              className="text-[var(--silver)] hover:text-[var(--ivory)]"
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>
          <pre className="p-4 max-h-64 overflow-auto text-xs font-mono text-[var(--pearl)] whitespace-pre-wrap">
            {selectedFile.content.slice(0, 2000)}
            {selectedFile.content.length > 2000 && "\n\n... (truncated)"}
          </pre>
        </div>
      )}
    </div>
  );
}
