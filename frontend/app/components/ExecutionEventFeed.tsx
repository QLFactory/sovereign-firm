"use client";

import { useEffect, useRef, useMemo } from "react";
import { StreamEvent, StreamEventType } from "../hooks/useStreaming";

interface ExecutionEventFeedProps {
  events: StreamEvent[];
  maxEvents?: number;
  autoScroll?: boolean;
  onEventClick?: (event: StreamEvent) => void;
  filter?: StreamEventType[];
}

// Event type configuration
const eventConfig: Record<StreamEventType, { icon: string; label: string; color: string }> = {
  // Connection events
  CONNECTED: { icon: "📡", label: "LINK_ESTABLISHED", color: "text-[var(--emerald-glow)]" },
  DISCONNECTED: { icon: "🔌", label: "LINK_SEVERED", color: "text-[var(--rose-glow)]" },
  ERROR: { icon: "⚠️", label: "SYSTEM_ERROR", color: "text-[var(--rose-glow)]" },

  // Phase events
  PHASE_CHANGE: { icon: "📍", label: "PHASE_ALIGN", color: "text-[var(--cyan-glow)]" },

  // Agent events
  AGENT_SPAWN: { icon: "🚀", label: "NEURAL_SPAWN", color: "text-[var(--violet-glow)]" },
  AGENT_DONE: { icon: "✅", label: "CORE_STABLE", color: "text-[var(--emerald-glow)]" },
  AGENT_ASSIGNED: { icon: "📋", label: "TASK_BIND", color: "text-[var(--violet-glow)]" },

  // Code events
  CODE_CHUNK: { icon: "📝", label: "STREAM_DATA", color: "text-[var(--silver)]" },
  CODE_COMPLETE: { icon: "✓", label: "HEX_STABLE", color: "text-[var(--emerald-glow)]" },
  FILE_START: { icon: "📄", label: "NODE_OPEN", color: "text-[var(--cyan-glow)]" },
  FILE_END: { icon: "📄", label: "NODE_CLOSE", color: "text-[var(--cyan-glow)]" },

  // Chat events
  CHAT_MESSAGE: { icon: "💬", label: "COMMS_INT", color: "text-[var(--pearl)]" },
  CHAT_TYPING: { icon: "⌨️", label: "INPUT_REQ", color: "text-[var(--silver)]" },

  // Build events
  BUILD_START: { icon: "🔨", label: "ASSEMBLY_INIT", color: "text-[var(--amber-glow)]" },
  BUILD_OUTPUT: { icon: "📤", label: "COMP_LOG", color: "text-[var(--silver)]" },
  BUILD_SUCCESS: { icon: "✅", label: "BUILD_VALID", color: "text-[var(--emerald-glow)]" },
  BUILD_ERROR: { icon: "❌", label: "BUILD_FAULT", color: "text-[var(--rose-glow)]" },

  // Test events
  TEST_START: { icon: "🧪", label: "LOGIC_VERIFY", color: "text-[var(--amber-glow)]" },
  TEST_RESULT: { icon: "📊", label: "TEST_METRIC", color: "text-[var(--cyan-glow)]" },

  // Preview events
  PREVIEW_READY: { icon: "👁️", label: "OPTIC_SYNC", color: "text-[var(--emerald-glow)]" },

  // Phase 4: Execution events
  EXECUTION_STARTED: { icon: "▶️", label: "EXEC_INIT", color: "text-[var(--emerald-glow)]" },
  EXECUTION_COMPLETED: { icon: "🏁", label: "EXEC_TERM", color: "text-[var(--emerald-glow)]" },
  EXECUTION_FAILED: { icon: "💥", label: "EXEC_CRASH", color: "text-[var(--rose-glow)]" },

  // Phase 4: Task events
  TASK_STARTED: { icon: "▶️", label: "TASK_INIT", color: "text-[var(--cyan-glow)]" },
  TASK_COMPLETED: { icon: "✓", label: "TASK_RESOLVED", color: "text-[var(--emerald-glow)]" },
  TASK_FAILED: { icon: "✗", label: "TASK_FAULT", color: "text-[var(--rose-glow)]" },
  TASK_RETRY: { icon: "🔄", label: "TASK_REBOUND", color: "text-[var(--amber-glow)]" },

  // Phase 4: Git events
  BRANCH_MERGED: { icon: "🔀", label: "CELL_MERGE", color: "text-[var(--violet-glow)]" },
  MERGE_FAILED: { icon: "⚠️", label: "MERGE_CONFLICT", color: "text-[var(--rose-glow)]" },

  // Phase 4: CI events
  CI_STAGE_START: { icon: "🔄", label: "PIPE_START", color: "text-[var(--amber-glow)]" },
  CI_STAGE_COMPLETE: { icon: "✓", label: "PIPE_SUCCESS", color: "text-[var(--emerald-glow)]" },
  CI_STAGE_FAILED: { icon: "✗", label: "PIPE_FAULT", color: "text-[var(--rose-glow)]" },

  // Phase 4: DAG events
  DAG_UPDATE: { icon: "📊", label: "GRAPH_SYNC", color: "text-[var(--cyan-glow)]" },
};

// High-priority events to highlight
const highlightEvents: StreamEventType[] = [
  "EXECUTION_STARTED",
  "EXECUTION_COMPLETED",
  "EXECUTION_FAILED",
  "TASK_FAILED",
  "TASK_RETRY",
  "BUILD_ERROR",
  "MERGE_FAILED",
  "CI_STAGE_FAILED",
  "ERROR",
];

// Events to show by default (filter out noisy events)
const defaultFilter: StreamEventType[] = [
  "CONNECTED",
  "DISCONNECTED",
  "ERROR",
  "PHASE_CHANGE",
  "AGENT_SPAWN",
  "AGENT_DONE",
  "AGENT_ASSIGNED",
  "CODE_COMPLETE",
  "FILE_START",
  "FILE_END",
  "CHAT_MESSAGE",
  "BUILD_START",
  "BUILD_SUCCESS",
  "BUILD_ERROR",
  "TEST_START",
  "TEST_RESULT",
  "PREVIEW_READY",
  "EXECUTION_STARTED",
  "EXECUTION_COMPLETED",
  "EXECUTION_FAILED",
  "TASK_STARTED",
  "TASK_COMPLETED",
  "TASK_FAILED",
  "TASK_RETRY",
  "BRANCH_MERGED",
  "MERGE_FAILED",
  "CI_STAGE_START",
  "CI_STAGE_COMPLETE",
  "CI_STAGE_FAILED",
  "DAG_UPDATE",
];

function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });
}

function getEventMessage(event: StreamEvent): string {
  const payload = event.payload || {};

  switch (event.type) {
    case "PHASE_CHANGE":
      return `${payload.previous_phase} → ${payload.current_phase}`;
    case "AGENT_SPAWN":
    case "AGENT_DONE":
      return `${payload.agent_name || payload.agent_id || "Agent"}`;
    case "AGENT_ASSIGNED":
      return `${payload.agent_name} → ${payload.task_id}`;
    case "TASK_STARTED":
    case "TASK_COMPLETED":
      return payload.task_name as string || payload.task_id as string || "Task";
    case "TASK_FAILED":
      return `${payload.task_name}: ${payload.error}`;
    case "TASK_RETRY":
      return `${payload.task_name} (attempt ${payload.retry_count})`;
    case "FILE_START":
    case "FILE_END":
      return payload.file_path as string || "File";
    case "CHAT_MESSAGE":
      return `${payload.agent || payload.role}: ${(payload.content as string || "").slice(0, 50)}...`;
    case "BUILD_ERROR":
      return payload.error as string || "Build failed";
    case "TEST_RESULT":
      return `${payload.passed || 0} passed, ${payload.failed || 0} failed`;
    case "EXECUTION_STARTED":
    case "EXECUTION_COMPLETED":
    case "EXECUTION_FAILED":
      return payload.message as string || "";
    case "BRANCH_MERGED":
      return `${payload.branch_name} → ${payload.target_branch || "main"}`;
    case "MERGE_FAILED":
      return `${payload.branch_name}: ${payload.error}`;
    case "CI_STAGE_START":
    case "CI_STAGE_COMPLETE":
    case "CI_STAGE_FAILED":
      return `${payload.stage}${payload.error ? `: ${payload.error}` : ""}`;
    case "DAG_UPDATE":
      return `${payload.completed}/${payload.total} tasks`;
    case "ERROR":
      return payload.message as string || "Unknown error";
    default:
      return JSON.stringify(payload).slice(0, 50);
  }
}

export default function ExecutionEventFeed({
  events,
  maxEvents = 100,
  autoScroll = true,
  onEventClick,
  filter,
}: ExecutionEventFeedProps) {
  const scrollRef = useRef<HTMLDivElement>(null);

  // Filter and limit events
  const filteredEvents = useMemo(() => {
    const activeFilter = filter || defaultFilter;
    return events
      .filter((e) => activeFilter.includes(e.type))
      .slice(-maxEvents);
  }, [events, filter, maxEvents]);

  // Auto-scroll to bottom when new events arrive
  useEffect(() => {
    if (autoScroll && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [filteredEvents, autoScroll]);

  if (filteredEvents.length === 0) {
    return (
      <div className="p-8 text-center glass rounded-2xl border-[var(--glass-border)] mx-4 my-8">
        <div className="text-5xl mb-4 opacity-40 animate-pulse">📡</div>
        <p className="text-[var(--ivory)] font-bold text-sm tracking-tight">Signal Loss</p>
        <p className="text-[var(--silver)] text-xs mt-2 opacity-60">Awaiting stream telemetry...</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-[var(--glass-border)] bg-[var(--carbon)]/30">
        <div className="flex justify-between items-center">
          <h3 className="font-bold text-[11px] uppercase tracking-widest text-[var(--silver)]">Telemetry Stream</h3>
          <span className="text-[10px] font-mono text-[var(--silver)] opacity-40">
            {filteredEvents.length} PKTS
          </span>
        </div>
      </div>

      {/* Event List */}
      <div ref={scrollRef} className="flex-1 overflow-auto p-4 space-y-2">
        {filteredEvents.map((event, idx) => {
          const config = eventConfig[event.type] || {
            icon: "📦",
            label: "UNKNOWN_PKT",
            color: "text-[var(--silver)]",
          };
          const isHighlight = highlightEvents.includes(event.type);
          const message = getEventMessage(event);
          const isError = event.type.includes("FAILED") || event.type === "ERROR";

          return (
            <button
              key={`${event.seq}-${idx}`}
              onClick={() => onEventClick?.(event)}
              className={`
                w-full p-3 rounded-xl border transition-all duration-300 text-left relative overflow-hidden group glass
                ${isHighlight ? "bg-[var(--glass-highlight)]" : "bg-transparent"}
                ${isError ? "border-[var(--rose-glow)]/30 bg-[var(--rose-glow)]/5" : "border-[var(--glass-border)]"}
                hover:border-[var(--ivory)]/40 hover:-translate-y-0.5
              `}
            >
              <div className="flex items-start gap-4 h-full">
                {/* Icon */}
                <div className={`w-8 h-8 rounded-full glass flex items-center justify-center text-sm border ${isError ? "border-[var(--rose-glow)]/30" : "border-[var(--ivory)]/10"}`}>
                  {config.icon}
                </div>

                {/* Content */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className={`text-[10px] font-bold uppercase tracking-widest ${config.color}`}>
                      {config.label}
                    </span>
                    <span className="text-[9px] font-mono text-[var(--silver)] opacity-30">
                      #{event.seq.toString().padStart(4, '0')}
                    </span>
                  </div>
                  {message && (
                    <div className="text-[11px] text-[var(--pearl)] mt-1 font-medium tracking-tight truncate leading-tight">
                      {message}
                    </div>
                  )}
                </div>

                {/* Timestamp */}
                <span className="text-[9px] font-mono text-[var(--silver)] opacity-30 pt-0.5">
                  {formatTimestamp(event.timestamp)}
                </span>
              </div>
            </button>
          );
        })}
      </div>

      {/* Footer with stats */}
      <div className="p-4 border-t border-[var(--glass-border)] bg-[var(--carbon)]/20">
        <div className="flex justify-between items-center text-[10px] font-bold tracking-widest">
          <span className="text-[var(--rose-glow)] flex items-center gap-1.5 uppercase">
            <span className="w-1.5 h-1.5 rounded-full bg-[var(--rose-glow)] shadow-glow-rose" />
            {events.filter((e) => e.type.includes("FAILED") || e.type === "ERROR").length} Anomalies
          </span>
          <span className="text-[var(--silver)] opacity-40 uppercase">
            Last Sync: {filteredEvents.length > 0 ? formatTimestamp(filteredEvents[filteredEvents.length - 1].timestamp) : "—"}
          </span>
        </div>
      </div>
    </div>
  );
}
