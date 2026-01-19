"use client";

import { useMemo } from "react";
import { DAGTask, DAGState } from "../hooks/useStreaming";

interface DAGProgressDashboardProps {
  dagState: DAGState | null;
  onTaskClick?: (task: DAGTask) => void;
}

// Status colors and icons
const statusConfig: Record<DAGTask["status"], { bg: string; text: string; icon: string; border: string; glow: string }> = {
  PENDING: { bg: "bg-transparent", text: "text-[var(--silver)]/40", icon: "○", border: "border-[var(--glass-border)]", glow: "" },
  READY: { bg: "bg-[var(--cyan-glow)]/10", text: "text-[var(--cyan-glow)]", icon: "◎", border: "border-[var(--cyan-glow)]/30", glow: "shadow-glow-cyan/10" },
  RUNNING: { bg: "bg-[var(--amber-glow)]/10", text: "text-[var(--amber-glow)]", icon: "◉", border: "border-[var(--amber-glow)]/30", glow: "shadow-glow-amber/20" },
  COMPLETED: { bg: "bg-[var(--emerald-glow)]/10", text: "text-[var(--emerald-glow)]", icon: "✓", border: "border-[var(--emerald-glow)]/30", glow: "shadow-glow-emerald/20" },
  FAILED: { bg: "bg-[var(--rose-glow)]/10", text: "text-[var(--rose-glow)]", icon: "✗", border: "border-[var(--rose-glow)]/30", glow: "shadow-glow-rose/20" },
  BLOCKED: { bg: "bg-[var(--carbon)]/50", text: "text-[var(--silver)]/30", icon: "⊘", border: "border-[var(--glass-border)]", glow: "" },
};

// Task type icons
const typeIcons: Record<string, string> = {
  frontend: "🎨",
  backend: "⚙️",
  database: "🗄️",
  api: "🔌",
  test: "🧪",
  deploy: "🚀",
  code: "📝",
  default: "📦",
};

export default function DAGProgressDashboard({ dagState, onTaskClick }: DAGProgressDashboardProps) {
  // Organize tasks into layers based on dependencies
  const layers = useMemo(() => {
    if (!dagState?.tasks?.length) return [];

    const tasks = dagState.tasks;
    const taskMap = new Map(tasks.map((t) => [t.id, t]));
    const layers: DAGTask[][] = [];
    const placed = new Set<string>();

    // Build layers using topological sort
    while (placed.size < tasks.length) {
      const layer: DAGTask[] = [];

      for (const task of tasks) {
        if (placed.has(task.id)) continue;

        // Check if all dependencies are placed
        const depsPlaced = !task.dependencies || task.dependencies.every((dep) => placed.has(dep));
        if (depsPlaced) {
          layer.push(task);
        }
      }

      if (layer.length === 0) {
        // Circular dependency or error - add remaining tasks
        for (const task of tasks) {
          if (!placed.has(task.id)) {
            layer.push(task);
          }
        }
      }

      layer.forEach((t) => placed.add(t.id));
      if (layer.length > 0) {
        layers.push(layer);
      }
    }

    return layers;
  }, [dagState]);

  // Calculate progress percentage
  const progress = dagState
    ? Math.round((dagState.completed / Math.max(dagState.total, 1)) * 100)
    : 0;

  if (!dagState) {
    return (
      <div className="p-4 text-center text-zinc-500">
        <div className="text-4xl mb-2">📊</div>
        <p className="text-sm">No task graph available</p>
        <p className="text-xs mt-1">Start a multi-agent project to see the DAG</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      {/* Progress Header */}
      <div className="p-4 border-b border-[var(--glass-border)] bg-[var(--carbon)]/30">
        <div className="flex justify-between items-center mb-3">
          <h3 className="font-bold text-[11px] uppercase tracking-widest text-[var(--silver)]">Task Orchestration</h3>
          <span className="text-[10px] font-mono text-[var(--silver)] opacity-40">
            {progress}% COHERENCE
          </span>
        </div>

        {/* Progress Bar */}
        <div className="h-1.5 bg-[var(--void)] rounded-full overflow-hidden border border-[var(--glass-border)] shadow-inner">
          <div
            className="h-full bg-gradient-to-r from-[var(--cyan-glow)] via-[var(--cyan-bright)] to-[var(--emerald-glow)] transition-all duration-700 ease-out shadow-glow-cyan"
            style={{ width: `${progress}%` }}
          />
        </div>

        {/* Status Summary */}
        <div className="flex gap-4 mt-3 text-[10px] font-bold tracking-tight uppercase">
          <span className="flex items-center gap-1.5 text-[var(--emerald-glow)]">
            <span className="w-1.5 h-1.5 rounded-full bg-[var(--emerald-glow)] shadow-glow-emerald" />
            {dagState.completed} DONE
          </span>
          <span className="flex items-center gap-1.5 text-[var(--amber-glow)]">
            <span className="w-1.5 h-1.5 rounded-full bg-[var(--amber-glow)] animate-pulse shadow-glow-amber" />
            {dagState.running} RUNNING
          </span>
          <span className="flex items-center gap-1.5 text-[var(--silver)] opacity-50">
            <span className="w-1.5 h-1.5 rounded-full bg-[var(--silver)] opacity-30" />
            {dagState.pending} PENDING
          </span>
          {dagState.failed > 0 && (
            <span className="flex items-center gap-1.5 text-[var(--rose-glow)]">
              <span className="w-1.5 h-1.5 rounded-full bg-[var(--rose-glow)] shadow-glow-rose" />
              {dagState.failed} FAULT
            </span>
          )}
        </div>
      </div>

      {/* DAG Visualization */}
      <div className="flex-1 overflow-auto p-3">
        <div className="flex flex-col gap-4">
          {layers.map((layer, layerIdx) => (
            <div key={layerIdx} className="relative">
              {/* Layer Label */}
              <div className="text-[10px] text-[var(--silver)]/40 mb-3 flex items-center gap-3 uppercase font-bold tracking-[0.2em]">
                <span className="whitespace-nowrap">LAYER_{layerIdx.toString().padStart(2, '0')}</span>
                <span className="flex-1 h-px bg-gradient-to-r from-[var(--glass-border)] to-transparent" />
                <span className="whitespace-nowrap italic">{layer.length} STACKED_NODES</span>
              </div>

              {/* Tasks in Layer */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
                {layer.map((task) => {
                  const config = statusConfig[task.status];
                  const typeIcon = typeIcons[task.type] || typeIcons.default;

                  return (
                    <button
                      key={task.id}
                      onClick={() => onTaskClick?.(task)}
                      className={`
                        p-4 rounded-xl border transition-all duration-300 text-left relative overflow-hidden group glass
                        ${config.bg} ${config.border} ${config.glow}
                        hover:border-[var(--ivory)]/40 hover:-translate-y-0.5
                        ${task.status === "RUNNING" ? "animate-pulse" : ""}
                      `}
                    >
                      <div className="flex items-start gap-4">
                        <div className="w-10 h-10 rounded-full glass flex items-center justify-center text-xl border border-[var(--ivory)]/10">
                          {typeIcon}
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center justify-between gap-2">
                            <span className={`font-bold text-xs uppercase tracking-widest truncate ${config.text}`}>
                              {task.name}
                            </span>
                            <span className={`text-[10px] font-bold ${config.text}`}>
                              {config.icon}
                            </span>
                          </div>

                          {task.description && (
                            <p className="text-[10px] text-[var(--silver)]/60 truncate mt-1 leading-tight">
                              {task.description}
                            </p>
                          )}

                          {/* Agent Assignment */}
                          {task.assigned_to && (
                            <div className="text-[9px] font-bold text-[var(--violet-glow)] mt-2 flex items-center gap-1.5 uppercase tracking-wider">
                              <span className="w-1 h-1 rounded-full bg-[var(--violet-glow)]" />
                              <span className="truncate">{task.assigned_to}</span>
                            </div>
                          )}

                          {/* Error Message */}
                          {task.error && (
                            <div className="text-[9px] text-[var(--rose-glow)] mt-1.5 font-medium border-t border-[var(--rose-glow)]/10 pt-1">
                              FAULT: {task.error}
                            </div>
                          )}

                          {/* Retry Count */}
                          {task.retry_count && task.retry_count > 0 && (
                            <div className="text-[9px] text-[var(--amber-glow)] mt-1 font-bold">
                              REBOUND_{task.retry_count}
                            </div>
                          )}
                        </div>
                      </div>

                      {/* Dependencies Indicator */}
                      {task.dependencies && task.dependencies.length > 0 && (
                        <div className="mt-3 pt-2 border-t border-[var(--glass-border)] text-[9px] font-mono text-[var(--silver)] opacity-30 flex justify-between items-center">
                          <span>DEPENDENCIES</span>
                          <span className="px-1.5 py-0.5 rounded bg-[var(--carbon)] border border-[var(--glass-border)]">0x{task.dependencies.length.toString(16)}</span>
                        </div>
                      )}
                    </button>
                  );
                })}
              </div>

              {/* Connection Lines to Next Layer */}
              {layerIdx < layers.length - 1 && (
                <div className="flex justify-center mt-3">
                  <div className="w-px h-6 bg-gradient-to-b from-[var(--glass-border)] to-transparent" />
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
