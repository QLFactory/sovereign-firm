"use client";

import { useMemo } from "react";
import { DAGTask, DAGState } from "../hooks/useStreaming";

interface DAGProgressDashboardProps {
  dagState: DAGState | null;
  onTaskClick?: (task: DAGTask) => void;
}

// Status colors and icons
const statusConfig: Record<DAGTask["status"], { bg: string; text: string; icon: string; border: string }> = {
  PENDING: { bg: "bg-zinc-800", text: "text-zinc-400", icon: "○", border: "border-zinc-600" },
  READY: { bg: "bg-blue-900/50", text: "text-blue-300", icon: "◎", border: "border-blue-500" },
  RUNNING: { bg: "bg-yellow-900/50", text: "text-yellow-300", icon: "◉", border: "border-yellow-500" },
  COMPLETED: { bg: "bg-green-900/50", text: "text-green-300", icon: "✓", border: "border-green-500" },
  FAILED: { bg: "bg-red-900/50", text: "text-red-300", icon: "✗", border: "border-red-500" },
  BLOCKED: { bg: "bg-zinc-900", text: "text-zinc-500", icon: "⊘", border: "border-zinc-700" },
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
        const depsPlaced = task.dependencies.every((dep) => placed.has(dep));
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
      <div className="p-3 border-b border-zinc-800">
        <div className="flex justify-between items-center mb-2">
          <h3 className="font-semibold text-sm">Task Progress</h3>
          <span className="text-xs text-zinc-400">
            {dagState.completed}/{dagState.total} tasks
          </span>
        </div>

        {/* Progress Bar */}
        <div className="h-2 bg-zinc-800 rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-blue-600 to-green-500 transition-all duration-500"
            style={{ width: `${progress}%` }}
          />
        </div>

        {/* Status Summary */}
        <div className="flex gap-3 mt-2 text-xs">
          <span className="flex items-center gap-1">
            <span className="w-2 h-2 rounded-full bg-green-500" />
            {dagState.completed} done
          </span>
          <span className="flex items-center gap-1">
            <span className="w-2 h-2 rounded-full bg-yellow-500 animate-pulse" />
            {dagState.running} running
          </span>
          <span className="flex items-center gap-1">
            <span className="w-2 h-2 rounded-full bg-zinc-500" />
            {dagState.pending} pending
          </span>
          {dagState.failed > 0 && (
            <span className="flex items-center gap-1 text-red-400">
              <span className="w-2 h-2 rounded-full bg-red-500" />
              {dagState.failed} failed
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
              <div className="text-xs text-zinc-600 mb-2 flex items-center gap-2">
                <span className="font-medium">Layer {layerIdx + 1}</span>
                <span className="flex-1 h-px bg-zinc-800" />
                <span>{layer.length} task{layer.length !== 1 ? "s" : ""}</span>
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
                        p-3 rounded-lg border transition-all text-left
                        ${config.bg} ${config.border}
                        hover:scale-[1.02] hover:shadow-lg
                        ${task.status === "RUNNING" ? "animate-pulse" : ""}
                      `}
                    >
                      <div className="flex items-start gap-2">
                        <span className="text-lg">{typeIcon}</span>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className={`font-medium text-sm truncate ${config.text}`}>
                              {task.name}
                            </span>
                            <span className={`text-xs ${config.text}`}>
                              {config.icon}
                            </span>
                          </div>

                          {task.description && (
                            <p className="text-xs text-zinc-500 truncate mt-0.5">
                              {task.description}
                            </p>
                          )}

                          {/* Agent Assignment */}
                          {task.assigned_to && (
                            <div className="text-xs text-zinc-500 mt-1 flex items-center gap-1">
                              <span>🤖</span>
                              <span className="truncate">{task.assigned_to}</span>
                            </div>
                          )}

                          {/* Error Message */}
                          {task.error && (
                            <div className="text-xs text-red-400 mt-1 truncate">
                              ⚠️ {task.error}
                            </div>
                          )}

                          {/* Retry Count */}
                          {task.retry_count && task.retry_count > 0 && (
                            <div className="text-xs text-yellow-500 mt-1">
                              🔄 Retry {task.retry_count}
                            </div>
                          )}
                        </div>
                      </div>

                      {/* Dependencies Indicator */}
                      {task.dependencies.length > 0 && (
                        <div className="mt-2 pt-2 border-t border-zinc-700/50 text-xs text-zinc-600">
                          ← {task.dependencies.length} dep{task.dependencies.length !== 1 ? "s" : ""}
                        </div>
                      )}
                    </button>
                  );
                })}
              </div>

              {/* Connection Lines to Next Layer */}
              {layerIdx < layers.length - 1 && (
                <div className="flex justify-center mt-2">
                  <div className="w-px h-4 bg-zinc-700" />
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
