"use client";

import { useMemo } from "react";
import { CIStageResult } from "../hooks/useStreaming";

interface CIStatusPanelProps {
  stages: CIStageResult[];
  currentStage?: "LINT" | "BUILD" | "TEST" | null;
  onStageClick?: (stage: CIStageResult) => void;
}

// Stage configuration
const stageConfig: Record<string, { icon: string; label: string; description: string }> = {
  LINT: { icon: "🔍", label: "Lint", description: "Code quality checks" },
  BUILD: { icon: "🔨", label: "Build", description: "Compilation & bundling" },
  TEST: { icon: "🧪", label: "Test", description: "Running test suite" },
};

const stageOrder = ["LINT", "BUILD", "TEST"];

function formatDuration(ms?: number): string {
  if (!ms) return "—";
  if (ms < 1000) return `${ms}ms`;
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${seconds % 60}s`;
}

export default function CIStatusPanel({ stages, currentStage, onStageClick }: CIStatusPanelProps) {
  // Create stage map for easy lookup
  const stageMap = useMemo(() => {
    const map = new Map<string, CIStageResult>();
    for (const stage of stages) {
      map.set(stage.stage, stage);
    }
    return map;
  }, [stages]);

  // Calculate overall status
  const overallStatus = useMemo(() => {
    if (stages.length === 0) return "idle";
    if (currentStage) return "running";
    const hasFailure = stages.some((s) => !s.success);
    if (hasFailure) return "failed";
    if (stages.length === stageOrder.length) return "passed";
    return "partial";
  }, [stages, currentStage]);

  // Status colors
  const statusColors: Record<string, { bg: string; text: string; border: string }> = {
    idle: { bg: "bg-zinc-800", text: "text-zinc-400", border: "border-zinc-600" },
    running: { bg: "bg-yellow-900/50", text: "text-yellow-300", border: "border-yellow-500" },
    passed: { bg: "bg-green-900/50", text: "text-green-300", border: "border-green-500" },
    failed: { bg: "bg-red-900/50", text: "text-red-300", border: "border-red-500" },
    partial: { bg: "bg-blue-900/50", text: "text-blue-300", border: "border-blue-500" },
  };

  const totalDuration = stages.reduce((sum, s) => sum + (s.duration_ms || 0), 0);

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-3 border-b border-zinc-800">
        <div className="flex justify-between items-center">
          <h3 className="font-semibold text-sm">CI Pipeline</h3>
          <span
            className={`text-xs px-2 py-1 rounded-full ${statusColors[overallStatus].bg} ${statusColors[overallStatus].text}`}
          >
            {overallStatus === "idle" && "Idle"}
            {overallStatus === "running" && "Running..."}
            {overallStatus === "passed" && "✓ Passed"}
            {overallStatus === "failed" && "✗ Failed"}
            {overallStatus === "partial" && "In Progress"}
          </span>
        </div>

        {totalDuration > 0 && (
          <div className="text-xs text-zinc-500 mt-1">
            Total: {formatDuration(totalDuration)}
          </div>
        )}
      </div>

      {/* Pipeline Stages */}
      <div className="flex-1 overflow-auto p-3">
        <div className="space-y-3">
          {stageOrder.map((stageName, idx) => {
            const config = stageConfig[stageName];
            const result = stageMap.get(stageName);
            const isCurrentStage = currentStage === stageName;

            // Determine stage status
            let status: "pending" | "running" | "passed" | "failed" = "pending";
            if (isCurrentStage) {
              status = "running";
            } else if (result) {
              status = result.success ? "passed" : "failed";
            }

            const colors = {
              pending: { bg: "bg-zinc-800", text: "text-zinc-500", icon: "○" },
              running: { bg: "bg-yellow-900/50", text: "text-yellow-300", icon: "◉" },
              passed: { bg: "bg-green-900/50", text: "text-green-300", icon: "✓" },
              failed: { bg: "bg-red-900/50", text: "text-red-300", icon: "✗" },
            };

            const c = colors[status];

            return (
              <div key={stageName}>
                {/* Connection Line */}
                {idx > 0 && (
                  <div className="flex justify-center -mt-1 mb-1">
                    <div
                      className={`w-px h-3 ${
                        status === "pending" ? "bg-zinc-700" : "bg-zinc-600"
                      }`}
                    />
                  </div>
                )}

                {/* Stage Card */}
                <button
                  onClick={() => result && onStageClick?.(result)}
                  disabled={!result}
                  className={`
                    w-full p-3 rounded-lg border transition-all text-left
                    ${c.bg} ${status === "pending" ? "border-zinc-700" : "border-zinc-600"}
                    ${result ? "hover:scale-[1.02] hover:shadow-lg cursor-pointer" : "cursor-default"}
                    ${isCurrentStage ? "animate-pulse" : ""}
                  `}
                >
                  <div className="flex items-center gap-3">
                    {/* Stage Icon */}
                    <div className="text-xl">{config.icon}</div>

                    {/* Stage Info */}
                    <div className="flex-1">
                      <div className="flex items-center justify-between">
                        <span className={`font-medium text-sm ${c.text}`}>
                          {config.label}
                        </span>
                        <span className={`text-sm ${c.text}`}>{c.icon}</span>
                      </div>
                      <div className="text-xs text-zinc-500">{config.description}</div>
                    </div>

                    {/* Duration */}
                    {result?.duration_ms && (
                      <div className="text-xs text-zinc-500">
                        {formatDuration(result.duration_ms)}
                      </div>
                    )}
                  </div>

                  {/* Error Message */}
                  {result?.error && (
                    <div className="mt-2 p-2 bg-red-950/50 rounded text-xs text-red-400 truncate">
                      ⚠️ {result.error}
                    </div>
                  )}

                  {/* Output Preview */}
                  {result?.output && status === "failed" && (
                    <div className="mt-2 p-2 bg-zinc-950/50 rounded text-xs text-zinc-400 font-mono truncate">
                      {result.output.slice(0, 100)}...
                    </div>
                  )}
                </button>
              </div>
            );
          })}
        </div>
      </div>

      {/* Footer */}
      <div className="p-3 border-t border-zinc-800">
        <div className="grid grid-cols-3 gap-2 text-center text-xs">
          <div>
            <div className="text-zinc-500">Passed</div>
            <div className="font-medium text-green-400">
              {stages.filter((s) => s.success).length}
            </div>
          </div>
          <div>
            <div className="text-zinc-500">Failed</div>
            <div className="font-medium text-red-400">
              {stages.filter((s) => !s.success).length}
            </div>
          </div>
          <div>
            <div className="text-zinc-500">Pending</div>
            <div className="font-medium text-zinc-400">
              {stageOrder.length - stages.length}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
