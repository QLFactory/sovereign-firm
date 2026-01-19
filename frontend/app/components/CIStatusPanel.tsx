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
    if (!stages) return map;
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
  const statusColors: Record<string, { bg: string; text: string; border: string; glow: string }> = {
    idle: { bg: "bg-[var(--glass-highlight)]", text: "text-[var(--silver)]", border: "border-[var(--glass-border)]", glow: "" },
    running: { bg: "bg-[var(--amber-glow)]/10", text: "text-[var(--amber-glow)]", border: "border-[var(--amber-glow)]/30", glow: "shadow-glow-amber/20" },
    passed: { bg: "bg-[var(--emerald-glow)]/10", text: "text-[var(--emerald-glow)]", border: "border-[var(--emerald-glow)]/30", glow: "shadow-glow-emerald/20" },
    failed: { bg: "bg-[var(--rose-glow)]/10", text: "text-[var(--rose-glow)]", border: "border-[var(--rose-glow)]/30", glow: "shadow-glow-rose/20" },
    partial: { bg: "bg-[var(--cyan-glow)]/10", text: "text-[var(--cyan-glow)]", border: "border-[var(--cyan-glow)]/30", glow: "shadow-glow-cyan/20" },
  };

  const totalDuration = stages.reduce((sum, s) => sum + (s.duration_ms || 0), 0);

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-[var(--glass-border)] bg-[var(--carbon)]/30 backdrop-blur-sm">
        <div className="flex justify-between items-center mb-2">
          <h3 className="font-bold text-[11px] uppercase tracking-widest text-[var(--silver)]">Engine Integrity</h3>
          <span
            className={`text-[10px] font-bold px-2 py-0.5 rounded-full border glass ${statusColors[overallStatus].bg} ${statusColors[overallStatus].text} ${statusColors[overallStatus].border} ${statusColors[overallStatus].glow}`}
          >
            {overallStatus === "idle" && "STANDBY"}
            {overallStatus === "running" && "SYNCHRONIZING..."}
            {overallStatus === "passed" && "INTEGRITY SECURED"}
            {overallStatus === "failed" && "BREACH DETECTED"}
            {overallStatus === "partial" && "PROCESSING"}
          </span>
        </div>

        {totalDuration > 0 && (
          <div className="text-[10px] font-mono text-[var(--silver)] opacity-40">
            METRICS: {formatDuration(totalDuration)} TOTAL EXECUTION
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
              pending: { bg: "bg-transparent", text: "text-[var(--silver)]/40", icon: "○", border: "border-[var(--glass-border)]", glow: "" },
              running: { bg: "bg-[var(--amber-glow)]/10", text: "text-[var(--amber-glow)]", icon: "◉", border: "border-[var(--amber-glow)]/30", glow: "shadow-glow-amber/20" },
              passed: { bg: "bg-[var(--emerald-glow)]/10", text: "text-[var(--emerald-glow)]", icon: "✓", border: "border-[var(--emerald-glow)]/30", glow: "shadow-glow-emerald/20" },
              failed: { bg: "bg-[var(--rose-glow)]/10", text: "text-[var(--rose-glow)]", icon: "✗", border: "border-[var(--rose-glow)]/30", glow: "shadow-glow-rose/20" },
            };

            const c = colors[status];

            return (
              <div key={stageName}>
                {/* Connection Line */}
                {idx > 0 && (
                  <div className="flex justify-center -mt-1 mb-1">
                    <div
                      className={`w-px h-3 ${status === "pending" ? "bg-zinc-700" : "bg-zinc-600"
                        }`}
                    />
                  </div>
                )}

                {/* Stage Card */}
                <button
                  onClick={() => result && onStageClick?.(result)}
                  disabled={!result}
                  className={`
                    w-full p-4 rounded-xl border transition-all duration-300 text-left relative overflow-hidden group glass
                    ${c.bg} ${c.border} ${c.glow}
                    ${result ? "hover:border-[var(--ivory)]/40 hover:-translate-y-0.5 cursor-pointer" : "cursor-default opacity-50"}
                    ${isCurrentStage ? "animate-pulse" : ""}
                  `}
                >
                  <div className="flex items-center gap-4">
                    {/* Stage Icon */}
                    <div className="w-10 h-10 rounded-full glass flex items-center justify-center text-xl border border-[var(--ivory)]/10">
                      {config.icon}
                    </div>

                    {/* Stage Info */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <span className={`font-bold text-xs uppercase tracking-widest ${c.text}`}>
                          {config.label}
                        </span>
                        <span className={`text-xs font-bold ${c.text}`}>{c.icon}</span>
                      </div>
                      <div className="text-[10px] text-[var(--silver)] opacity-60 mt-0.5">{config.description}</div>
                    </div>

                    {/* Duration */}
                    {result?.duration_ms && (
                      <div className="text-[10px] font-mono text-[var(--silver)] opacity-40">
                        {formatDuration(result.duration_ms)}
                      </div>
                    )}
                  </div>

                  {/* Error Message */}
                  {result?.error && (
                    <div className="mt-3 p-3 bg-[var(--rose-glow)]/10 rounded-lg text-[10px] font-medium text-[var(--rose-glow)] border border-[var(--rose-glow)]/20 animate-fade-in">
                      <span className="font-bold mr-1">BREACH:</span> {result.error}
                    </div>
                  )}

                  {/* Output Preview */}
                  {result?.output && status === "failed" && (
                    <div className="mt-2 p-3 bg-[var(--void)]/50 rounded-lg text-[10px] text-[var(--silver)]/60 font-mono border border-[var(--glass-border)] truncate">
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
      <div className="p-4 border-t border-[var(--glass-border)] bg-[var(--carbon)]/20">
        <div className="grid grid-cols-3 gap-3 text-center">
          <div className="glass p-2 rounded-lg border border-[var(--glass-border)]">
            <div className="text-[9px] uppercase font-bold tracking-widest text-[var(--emerald-glow)] opacity-60">Passed</div>
            <div className="text-xs font-bold text-[var(--ivory)]">
              {stages.filter((s) => s.success).length}
            </div>
          </div>
          <div className="glass p-2 rounded-lg border border-[var(--glass-border)]">
            <div className="text-[9px] uppercase font-bold tracking-widest text-[var(--rose-glow)] opacity-60">Failed</div>
            <div className="text-xs font-bold text-[var(--ivory)]">
              {stages.filter((s) => !s.success).length}
            </div>
          </div>
          <div className="glass p-2 rounded-lg border border-[var(--glass-border)]">
            <div className="text-[9px] uppercase font-bold tracking-widest text-[var(--silver)] opacity-40">Pending</div>
            <div className="text-xs font-bold text-[var(--ivory)]">
              {stageOrder.length - stages.length}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
