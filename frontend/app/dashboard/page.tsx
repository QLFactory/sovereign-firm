"use client";

import { useState, useEffect, useCallback, Suspense } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import dynamic from "next/dynamic";
import { useAppStore } from "../lib/store";
import { PHASE_COLORS, getPhaseProgress } from "../lib/api/types";
import type { Project, ProjectConfig } from "../lib/api/types";

// Dynamically import heavy components
const PodConsole = dynamic(() => import("../components/PodConsole"), {
  ssr: false,
  loading: () => <DashboardSkeleton />,
});

// Loading skeleton
function DashboardSkeleton() {
  return (
    <div className="flex items-center justify-center h-full">
      <div className="text-center">
        <div className="spinner mx-auto mb-4" />
        <p className="text-[var(--silver)]">Loading workspace...</p>
      </div>
    </div>
  );
}

// Sidebar Navigation
function Sidebar({
  activeView,
  setActiveView,
  projects,
  selectedProject,
  onSelectProject,
  onNewProject,
}: {
  activeView: string;
  setActiveView: (view: string) => void;
  projects: Project[];
  selectedProject: Project | null;
  onSelectProject: (project: Project) => void;
  onNewProject: () => void;
}) {
  return (
    <aside className="w-64 h-screen bg-[var(--carbon)] border-r border-[var(--steel)] flex flex-col">
      {/* Logo */}
      <div className="p-4 border-b border-[var(--steel)]">
        <Link href="/" className="flex items-center gap-3">
          <div className="w-9 h-9 bg-gradient-to-br from-[var(--cyan-glow)] to-[var(--violet-glow)] rounded-lg flex items-center justify-center">
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="white"
              strokeWidth="2.5"
            >
              <path d="M12 2L2 7l10 5 10-5-10-5z" />
              <path d="M2 17l10 5 10-5" />
              <path d="M2 12l10 5 10-5" />
            </svg>
          </div>
          <div>
            <div className="font-semibold text-[var(--ivory)] text-sm">
              Sovereign Firm
            </div>
            <div className="text-xs text-[var(--silver)]">Console</div>
          </div>
        </Link>
      </div>

      {/* New Project Button */}
      <div className="p-4">
        <button onClick={onNewProject} className="btn btn-primary w-full text-sm">
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          New Project
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto px-3">
        <div className="mb-4">
          <div className="text-caption text-[var(--silver)] px-3 mb-2">Workspace</div>
          <button
            onClick={() => setActiveView("projects")}
            className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
              activeView === "projects"
                ? "bg-[var(--slate)] text-[var(--ivory)]"
                : "text-[var(--silver)] hover:bg-[var(--graphite)] hover:text-[var(--pearl)]"
            }`}
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <path d="M3 3h7v7H3zM14 3h7v7h-7zM14 14h7v7h-7zM3 14h7v7H3z" />
            </svg>
            All Projects
          </button>

          <Link
            href="/projects"
            className="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors text-[var(--silver)] hover:bg-[var(--graphite)] hover:text-[var(--pearl)]"
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
              <polyline points="15 3 21 3 21 9" />
              <line x1="10" y1="14" x2="21" y2="3" />
            </svg>
            New Projects View
          </Link>
        </div>

        {/* Recent Projects */}
        {projects.length > 0 && (
          <div className="mb-4">
            <div className="text-caption text-[var(--silver)] px-3 mb-2">Recent</div>
            {projects.slice(0, 5).map((project) => (
              <button
                key={project.id}
                onClick={() => onSelectProject(project)}
                className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                  selectedProject?.id === project.id
                    ? "bg-[var(--slate)] text-[var(--ivory)]"
                    : "text-[var(--silver)] hover:bg-[var(--graphite)] hover:text-[var(--pearl)]"
                }`}
              >
                <div
                  className={`w-2 h-2 rounded-full ${
                    project.phase === "COMPLETE"
                      ? "bg-[var(--emerald-glow)]"
                      : project.phase === "FAILED"
                      ? "bg-[var(--rose-glow)]"
                      : "bg-[var(--amber-glow)] animate-pulse"
                  }`}
                />
                <span className="truncate">{project.name}</span>
              </button>
            ))}
          </div>
        )}
      </nav>

      {/* User section */}
      <div className="p-4 border-t border-[var(--steel)]">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[var(--violet-glow)] to-[var(--cyan-glow)] flex items-center justify-center text-xs font-bold">
            U
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-sm font-medium text-[var(--ivory)] truncate">User</div>
            <div className="text-xs text-[var(--silver)]">Pro Plan</div>
          </div>
        </div>
      </div>
    </aside>
  );
}

// Create Project Modal
function CreateProjectModal({
  isOpen,
  onClose,
  onSubmit,
  isLoading,
}: {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (config: ProjectConfig) => void;
  isLoading: boolean;
}) {
  const [config, setConfig] = useState<ProjectConfig>({
    project_name: "",
    initial_message: "",
    enable_full_stack: true,
    enable_deployment: true,
    enable_sre: true,
    preferred_frontend: "react",
    preferred_backend: "nodejs",
    preferred_database: "postgresql",
    preferred_cloud: "aws",
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!config.project_name.trim()) return;
    onSubmit(config);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-[var(--void)]/80 backdrop-blur-sm"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative w-full max-w-2xl mx-4 glass-card rounded-2xl overflow-hidden animate-scale-in">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-[var(--steel)]">
          <div>
            <h2 className="text-xl font-semibold text-[var(--ivory)]">New Project</h2>
            <p className="text-sm text-[var(--silver)]">
              Configure your AI-powered software project
            </p>
          </div>
          <button
            onClick={onClose}
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
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        {/* Content */}
        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {/* Project Name */}
          <div>
            <label className="input-label">Project Name</label>
            <input
              type="text"
              className="input"
              placeholder="e.g., TaskFlow, InvoiceHub, ChatBot"
              value={config.project_name}
              onChange={(e) => setConfig({ ...config, project_name: e.target.value })}
              autoFocus
            />
          </div>

          {/* Description */}
          <div>
            <label className="input-label">Project Description</label>
            <textarea
              className="input min-h-[100px] resize-none"
              placeholder="Describe what you want to build. Be specific about features, users, and functionality..."
              value={config.initial_message}
              onChange={(e) =>
                setConfig({ ...config, initial_message: e.target.value })
              }
            />
          </div>

          {/* Tech Stack */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="input-label">Frontend</label>
              <select
                className="input select"
                value={config.preferred_frontend}
                onChange={(e) =>
                  setConfig({ ...config, preferred_frontend: e.target.value })
                }
              >
                <option value="react">React</option>
                <option value="vue">Vue.js</option>
                <option value="angular">Angular</option>
                <option value="nextjs">Next.js</option>
                <option value="svelte">Svelte</option>
              </select>
            </div>
            <div>
              <label className="input-label">Backend</label>
              <select
                className="input select"
                value={config.preferred_backend}
                onChange={(e) =>
                  setConfig({ ...config, preferred_backend: e.target.value })
                }
              >
                <option value="nodejs">Node.js</option>
                <option value="python">Python (FastAPI)</option>
                <option value="go">Go</option>
                <option value="rust">Rust</option>
                <option value="java">Java (Spring)</option>
              </select>
            </div>
            <div>
              <label className="input-label">Database</label>
              <select
                className="input select"
                value={config.preferred_database}
                onChange={(e) =>
                  setConfig({ ...config, preferred_database: e.target.value })
                }
              >
                <option value="postgresql">PostgreSQL</option>
                <option value="mysql">MySQL</option>
                <option value="mongodb">MongoDB</option>
                <option value="sqlite">SQLite</option>
                <option value="redis">Redis</option>
              </select>
            </div>
            <div>
              <label className="input-label">Cloud Provider</label>
              <select
                className="input select"
                value={config.preferred_cloud}
                onChange={(e) =>
                  setConfig({ ...config, preferred_cloud: e.target.value })
                }
              >
                <option value="aws">AWS</option>
                <option value="gcp">Google Cloud</option>
                <option value="azure">Microsoft Azure</option>
                <option value="digitalocean">DigitalOcean</option>
              </select>
            </div>
          </div>

          {/* Feature Toggles */}
          <div>
            <label className="input-label mb-3">Generate Artifacts</label>
            <div className="grid grid-cols-3 gap-3">
              <label className="flex items-center gap-3 p-3 rounded-lg bg-[var(--graphite)] cursor-pointer hover:bg-[var(--slate)] transition-colors">
                <input
                  type="checkbox"
                  checked={config.enable_full_stack}
                  onChange={(e) =>
                    setConfig({ ...config, enable_full_stack: e.target.checked })
                  }
                  className="sr-only"
                />
                <div
                  className={`w-5 h-5 rounded border-2 flex items-center justify-center transition-colors ${
                    config.enable_full_stack
                      ? "bg-[var(--cyan-glow)] border-[var(--cyan-glow)]"
                      : "border-[var(--steel)]"
                  }`}
                >
                  {config.enable_full_stack && (
                    <svg
                      width="12"
                      height="12"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="var(--void)"
                      strokeWidth="3"
                    >
                      <polyline points="20 6 9 17 4 12" />
                    </svg>
                  )}
                </div>
                <span className="text-sm text-[var(--pearl)]">Full Stack</span>
              </label>

              <label className="flex items-center gap-3 p-3 rounded-lg bg-[var(--graphite)] cursor-pointer hover:bg-[var(--slate)] transition-colors">
                <input
                  type="checkbox"
                  checked={config.enable_deployment}
                  onChange={(e) =>
                    setConfig({ ...config, enable_deployment: e.target.checked })
                  }
                  className="sr-only"
                />
                <div
                  className={`w-5 h-5 rounded border-2 flex items-center justify-center transition-colors ${
                    config.enable_deployment
                      ? "bg-[var(--cyan-glow)] border-[var(--cyan-glow)]"
                      : "border-[var(--steel)]"
                  }`}
                >
                  {config.enable_deployment && (
                    <svg
                      width="12"
                      height="12"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="var(--void)"
                      strokeWidth="3"
                    >
                      <polyline points="20 6 9 17 4 12" />
                    </svg>
                  )}
                </div>
                <span className="text-sm text-[var(--pearl)]">DevOps</span>
              </label>

              <label className="flex items-center gap-3 p-3 rounded-lg bg-[var(--graphite)] cursor-pointer hover:bg-[var(--slate)] transition-colors">
                <input
                  type="checkbox"
                  checked={config.enable_sre}
                  onChange={(e) =>
                    setConfig({ ...config, enable_sre: e.target.checked })
                  }
                  className="sr-only"
                />
                <div
                  className={`w-5 h-5 rounded border-2 flex items-center justify-center transition-colors ${
                    config.enable_sre
                      ? "bg-[var(--cyan-glow)] border-[var(--cyan-glow)]"
                      : "border-[var(--steel)]"
                  }`}
                >
                  {config.enable_sre && (
                    <svg
                      width="12"
                      height="12"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="var(--void)"
                      strokeWidth="3"
                    >
                      <polyline points="20 6 9 17 4 12" />
                    </svg>
                  )}
                </div>
                <span className="text-sm text-[var(--pearl)]">SRE</span>
              </label>
            </div>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-4 border-t border-[var(--steel)]">
            <button type="button" onClick={onClose} className="btn btn-ghost">
              Cancel
            </button>
            <button
              type="submit"
              className="btn btn-primary"
              disabled={!config.project_name.trim() || isLoading}
            >
              {isLoading ? (
                <>
                  <div className="spinner w-4 h-4" />
                  Creating...
                </>
              ) : (
                <>
                  Create Project
                  <svg
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                  >
                    <path d="M5 12h14M12 5l7 7-7 7" />
                  </svg>
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// Project Card
function ProjectCard({
  project,
  onSelect,
}: {
  project: Project;
  onSelect: () => void;
}) {
  const progress = getPhaseProgress(project.phase);
  const phaseColor = PHASE_COLORS[project.phase] || "badge";

  return (
    <div className="card card-glow group cursor-pointer" onClick={onSelect}>
      <div className="flex items-start justify-between mb-4">
        <div>
          <h3 className="font-semibold text-[var(--ivory)] mb-1">{project.name}</h3>
          <span className={`badge ${phaseColor}`}>{project.phase}</span>
        </div>
        <Link
          href={`/projects/${project.id}`}
          onClick={(e) => e.stopPropagation()}
          className="opacity-0 group-hover:opacity-100 btn btn-ghost btn-icon"
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
            <polyline points="15 3 21 3 21 9" />
            <line x1="10" y1="14" x2="21" y2="3" />
          </svg>
        </Link>
      </div>

      {/* Progress bar */}
      <div className="mb-3">
        <div className="h-1.5 bg-[var(--steel)] rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-[var(--cyan-glow)] to-[var(--violet-glow)] rounded-full transition-all duration-500"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      {/* Stack badges */}
      <div className="flex flex-wrap gap-2 text-xs">
        <span className="px-2 py-1 bg-[var(--graphite)] rounded text-[var(--silver)]">
          {project.config.preferred_frontend}
        </span>
        <span className="px-2 py-1 bg-[var(--graphite)] rounded text-[var(--silver)]">
          {project.config.preferred_backend}
        </span>
        <span className="px-2 py-1 bg-[var(--graphite)] rounded text-[var(--silver)]">
          {project.config.preferred_database}
        </span>
      </div>

      {/* Meta */}
      <div className="mt-4 pt-4 border-t border-[var(--steel)] flex items-center justify-between text-xs text-[var(--silver)]">
        <span>Created {new Date(project.createdAt).toLocaleDateString()}</span>
        <span className="flex items-center gap-1">
          <span
            className={`w-2 h-2 rounded-full ${
              project.phase === "COMPLETE"
                ? "bg-[var(--emerald-glow)]"
                : project.phase === "FAILED"
                ? "bg-[var(--rose-glow)]"
                : "bg-[var(--amber-glow)] animate-pulse"
            }`}
          />
          {project.phase === "COMPLETE"
            ? "Done"
            : project.phase === "FAILED"
            ? "Failed"
            : "In Progress"}
        </span>
      </div>
    </div>
  );
}

// Project Grid View
function ProjectsView({
  projects,
  onSelectProject,
  onNewProject,
}: {
  projects: Project[];
  onSelectProject: (project: Project) => void;
  onNewProject: () => void;
}) {
  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-[var(--ivory)] mb-1">Projects</h1>
          <p className="text-[var(--silver)]">
            Manage your AI-generated software projects
          </p>
        </div>
        <button onClick={onNewProject} className="btn btn-primary">
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          New Project
        </button>
      </div>

      {projects.length === 0 ? (
        <div className="text-center py-20">
          <div className="w-20 h-20 mx-auto mb-6 rounded-2xl bg-[var(--graphite)] flex items-center justify-center">
            <svg
              width="40"
              height="40"
              viewBox="0 0 24 24"
              fill="none"
              stroke="var(--silver)"
              strokeWidth="1.5"
            >
              <path d="M3 3h7v7H3zM14 3h7v7h-7zM14 14h7v7h-7zM3 14h7v7H3z" />
            </svg>
          </div>
          <h2 className="text-xl font-semibold text-[var(--ivory)] mb-2">
            No projects yet
          </h2>
          <p className="text-[var(--silver)] mb-6 max-w-sm mx-auto">
            Create your first project and let our AI agents build your software.
          </p>
          <button onClick={onNewProject} className="btn btn-primary">
            Create Your First Project
          </button>
        </div>
      ) : (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {projects.map((project) => (
            <ProjectCard
              key={project.id}
              project={project}
              onSelect={() => onSelectProject(project)}
            />
          ))}
        </div>
      )}
    </div>
  );
}

// Main Dashboard Component
export default function Dashboard() {
  const router = useRouter();
  const [activeView, setActiveView] = useState("projects");
  const [showCreateModal, setShowCreateModal] = useState(false);

  // Store state and actions
  const projects = useAppStore((state) => state.projects);
  const currentProject = useAppStore((state) => state.currentProject);
  const isCreating = useAppStore((state) => state.isCreating);
  const selectProject = useAppStore((state) => state.selectProject);
  const createProject = useAppStore((state) => state.createProject);

  // Create new project
  const handleCreateProject = useCallback(
    async (config: ProjectConfig) => {
      const workflowId = await createProject(config);
      if (workflowId) {
        setShowCreateModal(false);
        router.push(`/projects/${workflowId}`);
      }
    },
    [createProject, router]
  );

  // Select project and open console
  const handleSelectProject = useCallback(
    (project: Project) => {
      selectProject(project);
      setActiveView("console");
    },
    [selectProject]
  );

  return (
    <div className="flex h-screen bg-[var(--obsidian)]">
      <Sidebar
        activeView={activeView}
        setActiveView={setActiveView}
        projects={projects}
        selectedProject={currentProject}
        onSelectProject={handleSelectProject}
        onNewProject={() => setShowCreateModal(true)}
      />

      <main className="flex-1 overflow-hidden">
        {activeView === "projects" ? (
          <div className="h-full overflow-y-auto">
            <ProjectsView
              projects={projects}
              onSelectProject={handleSelectProject}
              onNewProject={() => setShowCreateModal(true)}
            />
          </div>
        ) : activeView === "console" && currentProject ? (
          <Suspense fallback={<DashboardSkeleton />}>
            <PodConsole key={currentProject.id} workflowId={currentProject.id} />
          </Suspense>
        ) : (
          <div className="h-full flex items-center justify-center">
            <div className="text-center">
              <p className="text-[var(--silver)]">
                Select a project or create a new one
              </p>
            </div>
          </div>
        )}
      </main>

      <CreateProjectModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSubmit={handleCreateProject}
        isLoading={isCreating}
      />
    </div>
  );
}
