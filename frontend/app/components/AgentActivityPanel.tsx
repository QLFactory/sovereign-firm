"use client";

import { useMemo } from "react";
import { ActiveAgent } from "../hooks/useStreaming";

interface AgentActivityPanelProps {
  activeAgents: ActiveAgent[];
  onAgentClick?: (agent: ActiveAgent) => void;
}

const agentColors: Record<string, { bg: string; border: string; text: string; glow: string }> = {
  architect: { bg: "bg-[var(--violet-glow)]/10", border: "border-[var(--violet-glow)]/30", text: "text-[var(--violet-glow)]", glow: "shadow-glow-violet/20" },
  developer: { bg: "bg-[var(--cyan-glow)]/10", border: "border-[var(--cyan-glow)]/30", text: "text-[var(--cyan-glow)]", glow: "shadow-glow-cyan/20" },
  devops: { bg: "bg-[var(--amber-glow)]/10", border: "border-[var(--amber-glow)]/30", text: "text-[var(--amber-glow)]", glow: "shadow-glow-amber/20" },
  qa: { bg: "bg-[var(--emerald-glow)]/10", border: "border-[var(--emerald-glow)]/30", text: "text-[var(--emerald-glow)]", glow: "shadow-glow-emerald/20" },
  default: { bg: "bg-[var(--glass-highlight)]", border: "border-[var(--glass-border)]", text: "text-[var(--silver)]", glow: "" },
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
    if (!activeAgents) return [];
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
      <div className="p-8 text-center glass rounded-2xl border-[var(--glass-border)] mx-4 my-8">
        <div className="text-5xl mb-4 opacity-40 animate-pulse">🤖</div>
        <p className="text-[var(--ivory)] font-bold text-sm tracking-tight">System Idle</p>
        <p className="text-[var(--silver)] text-xs mt-2 opacity-60">Awaiting agent orchestration events...</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b border-[var(--glass-border)] bg-[var(--carbon)]/30">
        <div className="flex justify-between items-center mb-3">
          <h3 className="font-bold text-[11px] uppercase tracking-widest text-[var(--silver)]">Active Agents</h3>
          <span className="text-[10px] font-bold px-2 py-0.5 bg-[var(--emerald-glow)]/20 text-[var(--emerald-glow)] rounded-full border border-[var(--emerald-glow)]/30 shadow-glow-emerald/10">
            {activeAgents.length} LIVE
          </span>
        </div>

        {/* Role Summary */}
        <div className="flex gap-1.5 mt-2 flex-wrap">
          {Object.entries(agentsByRole).map(([role, agents]) => {
            const colors = agentColors[role] || agentColors.default;
            const icon = agentIcons[role] || agentIcons.default;
            return (
              <span
                key={role}
                className={`text-[9px] font-bold uppercase tracking-tighter px-2 py-0.5 rounded-full glass border ${colors.border} ${colors.text} flex items-center gap-1`}
              >
                <span>{icon}</span> {agents.length} {role}
              </span>
            );
          })}
        </div>
      </div>

      {/* Agent List */}
      <div className="flex-1 overflow-auto p-4 space-y-3">
        {sortedAgents.map((agent) => {
          const role = getAgentRole(agent.agent_name);
          const colors = agentColors[role] || agentColors.default;
          const icon = agentIcons[role] || agentIcons.default;

          return (
            <button
              key={`${agent.agent_id}-${agent.task_id}`}
              onClick={() => onAgentClick?.(agent)}
              className={`
                w-full p-4 rounded-xl border transition-all duration-300 text-left relative overflow-hidden group
                ${colors.bg} ${colors.border} ${colors.glow}
                hover:border-[var(--ivory)]/40 hover:-translate-y-1
              `}
            >
              <div className="absolute top-0 right-0 p-2 opacity-5 group-hover:opacity-10 transition-opacity">
                <span className="text-4xl">{icon}</span>
              </div>

              <div className="flex items-start gap-4 h-full">
                {/* Agent Icon */}
                <div className="w-10 h-10 rounded-full glass flex items-center justify-center text-xl shadow-inner border border-[var(--ivory)]/10">
                  {icon}
                </div>

                {/* Agent Info */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <span className={`font-bold text-xs uppercase tracking-wide tracking-tight truncate ${colors.text}`}>
                      {agent.agent_name}
                    </span>
                    <span className="text-[10px] font-mono text-[var(--silver)] flex-shrink-0 opacity-60">
                      {formatDuration(agent.started_at)}
                    </span>
                  </div>

                  {/* Task Info */}
                  <div className="text-[11px] text-[var(--pearl)] mt-1.5 font-medium flex items-center gap-1.5">
                    <span className="opacity-40 text-[9px]">EXEC:</span>
                    <span className="truncate">{agent.task_name}</span>
                  </div>

                  {/* Agent ID */}
                  <div className="text-[9px] text-[var(--silver)] mt-1.5 font-mono opacity-30 truncate uppercase tracking-widest">
                    NODE_{agent.agent_id.slice(0, 8)}
                  </div>
                </div>

                {/* Activity Indicator */}
                <div className="flex-shrink-0 pt-1">
                  <div className={`w-1.5 h-1.5 rounded-full ${colors.text.replace('text-', 'bg-')} animate-pulse shadow-[0_0_8px_currentColor]`} />
                </div>
              </div>
            </button>
          );
        })}
      </div>

      {/* Footer Stats */}
      <div className="p-4 border-t border-[var(--glass-border)] bg-[var(--carbon)]/20">
        <div className="grid grid-cols-3 gap-4 text-center">
          <div className="glass p-2 rounded-lg border border-[var(--glass-border)]">
            <div className="text-[9px] uppercase font-bold tracking-widest text-[var(--silver)] opacity-50">Total</div>
            <div className="text-xs font-bold text-[var(--ivory)]">{activeAgents.length}</div>
          </div>
          <div className="glass p-2 rounded-lg border border-[var(--glass-border)]">
            <div className="text-[9px] uppercase font-bold tracking-widest text-[var(--silver)] opacity-50">Roles</div>
            <div className="text-xs font-bold text-[var(--ivory)]">{Object.keys(agentsByRole).length}</div>
          </div>
          <div className="glass p-2 rounded-lg border border-[var(--glass-border)]">
            <div className="text-[9px] uppercase font-bold tracking-widest text-[var(--silver)] opacity-50">Tasks</div>
            <div className="text-xs font-bold text-[var(--ivory)]">
              {new Set(activeAgents.map(a => a.task_id)).size}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
