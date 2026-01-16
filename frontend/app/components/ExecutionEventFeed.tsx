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
  CONNECTED: { icon: "🔗", label: "Connected", color: "text-green-400" },
  DISCONNECTED: { icon: "🔌", label: "Disconnected", color: "text-red-400" },
  ERROR: { icon: "❌", label: "Error", color: "text-red-400" },

  // Phase events
  PHASE_CHANGE: { icon: "📍", label: "Phase", color: "text-blue-400" },

  // Agent events
  AGENT_SPAWN: { icon: "🚀", label: "Agent Spawned", color: "text-purple-400" },
  AGENT_DONE: { icon: "✅", label: "Agent Done", color: "text-green-400" },
  AGENT_ASSIGNED: { icon: "📋", label: "Agent Assigned", color: "text-purple-400" },

  // Code events
  CODE_CHUNK: { icon: "📝", label: "Code", color: "text-zinc-400" },
  CODE_COMPLETE: { icon: "✓", label: "Code Complete", color: "text-green-400" },
  FILE_START: { icon: "📄", label: "File Start", color: "text-blue-400" },
  FILE_END: { icon: "📄", label: "File End", color: "text-blue-400" },

  // Chat events
  CHAT_MESSAGE: { icon: "💬", label: "Message", color: "text-zinc-300" },
  CHAT_TYPING: { icon: "⌨️", label: "Typing", color: "text-zinc-500" },

  // Build events
  BUILD_START: { icon: "🔨", label: "Build Start", color: "text-yellow-400" },
  BUILD_OUTPUT: { icon: "📤", label: "Build Output", color: "text-zinc-400" },
  BUILD_SUCCESS: { icon: "✅", label: "Build Success", color: "text-green-400" },
  BUILD_ERROR: { icon: "❌", label: "Build Error", color: "text-red-400" },

  // Test events
  TEST_START: { icon: "🧪", label: "Test Start", color: "text-yellow-400" },
  TEST_RESULT: { icon: "📊", label: "Test Result", color: "text-blue-400" },

  // Preview events
  PREVIEW_READY: { icon: "👁️", label: "Preview Ready", color: "text-green-400" },

  // Phase 4: Execution events
  EXECUTION_STARTED: { icon: "▶️", label: "Execution Started", color: "text-green-400" },
  EXECUTION_COMPLETED: { icon: "🏁", label: "Execution Completed", color: "text-green-400" },
  EXECUTION_FAILED: { icon: "💥", label: "Execution Failed", color: "text-red-400" },

  // Phase 4: Task events
  TASK_STARTED: { icon: "▶️", label: "Task Started", color: "text-blue-400" },
  TASK_COMPLETED: { icon: "✓", label: "Task Completed", color: "text-green-400" },
  TASK_FAILED: { icon: "✗", label: "Task Failed", color: "text-red-400" },
  TASK_RETRY: { icon: "🔄", label: "Task Retry", color: "text-yellow-400" },

  // Phase 4: Git events
  BRANCH_MERGED: { icon: "🔀", label: "Branch Merged", color: "text-purple-400" },
  MERGE_FAILED: { icon: "⚠️", label: "Merge Failed", color: "text-red-400" },

  // Phase 4: CI events
  CI_STAGE_START: { icon: "🔄", label: "CI Stage Start", color: "text-yellow-400" },
  CI_STAGE_COMPLETE: { icon: "✓", label: "CI Stage Complete", color: "text-green-400" },
  CI_STAGE_FAILED: { icon: "✗", label: "CI Stage Failed", color: "text-red-400" },

  // Phase 4: DAG events
  DAG_UPDATE: { icon: "📊", label: "DAG Update", color: "text-blue-400" },
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
      <div className="p-4 text-center text-zinc-500">
        <div className="text-4xl mb-2">📡</div>
        <p className="text-sm">No events yet</p>
        <p className="text-xs mt-1">Events will appear here as execution progresses</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-3 border-b border-zinc-800">
        <div className="flex justify-between items-center">
          <h3 className="font-semibold text-sm">Event Feed</h3>
          <span className="text-xs text-zinc-500">
            {filteredEvents.length} events
          </span>
        </div>
      </div>

      {/* Event List */}
      <div ref={scrollRef} className="flex-1 overflow-auto p-2 space-y-1">
        {filteredEvents.map((event, idx) => {
          const config = eventConfig[event.type] || {
            icon: "📦",
            label: event.type,
            color: "text-zinc-400",
          };
          const isHighlight = highlightEvents.includes(event.type);
          const message = getEventMessage(event);

          return (
            <button
              key={`${event.seq}-${idx}`}
              onClick={() => onEventClick?.(event)}
              className={`
                w-full p-2 rounded text-left transition-colors
                ${isHighlight ? "bg-zinc-800/50" : "hover:bg-zinc-800/30"}
                ${event.type.includes("FAILED") || event.type === "ERROR" ? "bg-red-950/30" : ""}
              `}
            >
              <div className="flex items-start gap-2">
                {/* Icon */}
                <span className="flex-shrink-0">{config.icon}</span>

                {/* Content */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className={`text-xs font-medium ${config.color}`}>
                      {config.label}
                    </span>
                    <span className="text-xs text-zinc-600">
                      #{event.seq}
                    </span>
                  </div>
                  {message && (
                    <div className="text-xs text-zinc-400 truncate mt-0.5">
                      {message}
                    </div>
                  )}
                </div>

                {/* Timestamp */}
                <span className="text-xs text-zinc-600 flex-shrink-0">
                  {formatTimestamp(event.timestamp)}
                </span>
              </div>
            </button>
          );
        })}
      </div>

      {/* Footer with stats */}
      <div className="p-2 border-t border-zinc-800">
        <div className="flex justify-between text-xs text-zinc-500">
          <span>
            {events.filter((e) => e.type.includes("FAILED") || e.type === "ERROR").length} errors
          </span>
          <span>
            Last: {filteredEvents.length > 0 ? formatTimestamp(filteredEvents[filteredEvents.length - 1].timestamp) : "—"}
          </span>
        </div>
      </div>
    </div>
  );
}
