"use client";

import { useMemo } from "react";
import { ActiveAgent } from "../hooks/useStreaming";

interface AgentActivityPanelProps {
  activeAgents: ActiveAgent[];
  onAgentClick?: (agent: ActiveAgent) => void;
}

// Agent role colors
const agentColors: Record<string, { bg: string; border: string; text: string }> = {
  architect: { bg: "bg-purple-900/50", border: "border-purple-500", text: "text-purple-300" },
  developer: { bg: "bg-blue-900/50", border: "border-blue-500", text: "text-blue-300" },
  devops: { bg: "bg-orange-900/50", border: "border-orange-500", text: "text-orange-300" },
  qa: { bg: "bg-green-900/50", border: "border-green-500", text: "text-green-300" },
  default: { bg: "bg-zinc-800", border: "border-zinc-600", text: "text-zinc-300" },
};

// Agent role icons
const agentIcons: Record<string, string> = {
  architect: "🏛️",
  developer: "👨‍💻",
  devops: "🔧",
  qa: "🧪",
  default: "🤖",
};

function formatDuration(startedAt: string): string {
  const start = new Date(startedAt);
  const now = new Date();
  const diffMs = now.getTime() - start.getTime();

  const seconds = Math.floor(diffMs / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  }
  return `${seconds}s`;
}

function getAgentRole(agentName: string): string {
  const nameLower = agentName.toLowerCase();
  if (nameLower.includes("architect")) return "architect";
  if (nameLower.includes("developer") || nameLower.includes("dev")) return "developer";
  if (nameLower.includes("devops") || nameLower.includes("ops")) return "devops";
  if (nameLower.includes("qa") || nameLower.includes("test")) return "qa";
  return "default";
}

export default function AgentActivityPanel({ activeAgents, onAgentClick }: AgentActivityPanelProps) {
  // Sort agents by start time (newest first)
  const sortedAgents = useMemo(() => {
    return [...activeAgents].sort((a, b) => {
      return new Date(b.started_at).getTime() - new Date(a.started_at).getTime();
    });
  }, [activeAgents]);

  // Group agents by role
  const agentsByRole = useMemo(() => {
    const groups: Record<string, ActiveAgent[]> = {};
    for (const agent of activeAgents) {
      const role = getAgentRole(agent.agent_name);
      if (!groups[role]) {
        groups[role] = [];
      }
      groups[role].push(agent);
    }
    return groups;
  }, [activeAgents]);

  if (activeAgents.length === 0) {
    return (
      <div className="p-4 text-center text-zinc-500">
        <div className="text-4xl mb-2">🤖</div>
        <p className="text-sm">No active agents</p>
        <p className="text-xs mt-1">Agents will appear here when tasks start</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-3 border-b border-zinc-800">
        <div className="flex justify-between items-center">
          <h3 className="font-semibold text-sm">Active Agents</h3>
          <span className="text-xs px-2 py-1 bg-green-900/50 text-green-300 rounded-full">
            {activeAgents.length} running
          </span>
        </div>

        {/* Role Summary */}
        <div className="flex gap-2 mt-2 flex-wrap">
          {Object.entries(agentsByRole).map(([role, agents]) => {
            const colors = agentColors[role] || agentColors.default;
            const icon = agentIcons[role] || agentIcons.default;
            return (
              <span
                key={role}
                className={`text-xs px-2 py-0.5 rounded-full ${colors.bg} ${colors.text}`}
              >
                {icon} {agents.length}
              </span>
            );
          })}
        </div>
      </div>

      {/* Agent List */}
      <div className="flex-1 overflow-auto p-3 space-y-2">
        {sortedAgents.map((agent) => {
          const role = getAgentRole(agent.agent_name);
          const colors = agentColors[role] || agentColors.default;
          const icon = agentIcons[role] || agentIcons.default;

          return (
            <button
              key={`${agent.agent_id}-${agent.task_id}`}
              onClick={() => onAgentClick?.(agent)}
              className={`
                w-full p-3 rounded-lg border transition-all text-left
                ${colors.bg} ${colors.border}
                hover:scale-[1.02] hover:shadow-lg
              `}
            >
              <div className="flex items-start gap-3">
                {/* Agent Icon */}
                <div className="text-2xl flex-shrink-0">{icon}</div>

                {/* Agent Info */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <span className={`font-medium text-sm truncate ${colors.text}`}>
                      {agent.agent_name}
                    </span>
                    <span className="text-xs text-zinc-500 flex-shrink-0">
                      {formatDuration(agent.started_at)}
                    </span>
                  </div>

                  {/* Task Info */}
                  <div className="text-xs text-zinc-400 mt-1 truncate">
                    📋 {agent.task_name}
                  </div>

                  {/* Agent ID */}
                  <div className="text-xs text-zinc-600 mt-1 font-mono truncate">
                    ID: {agent.agent_id.slice(0, 8)}...
                  </div>
                </div>

                {/* Activity Indicator */}
                <div className="flex-shrink-0">
                  <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                </div>
              </div>
            </button>
          );
        })}
      </div>

      {/* Footer Stats */}
      <div className="p-3 border-t border-zinc-800">
        <div className="grid grid-cols-3 gap-2 text-center text-xs">
          <div>
            <div className="text-zinc-500">Total</div>
            <div className="font-medium text-zinc-300">{activeAgents.length}</div>
          </div>
          <div>
            <div className="text-zinc-500">Roles</div>
            <div className="font-medium text-zinc-300">{Object.keys(agentsByRole).length}</div>
          </div>
          <div>
            <div className="text-zinc-500">Tasks</div>
            <div className="font-medium text-zinc-300">
              {new Set(activeAgents.map(a => a.task_id)).size}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
