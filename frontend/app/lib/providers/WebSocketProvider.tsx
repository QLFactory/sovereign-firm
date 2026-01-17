"use client";

import { useEffect, useRef, useCallback, type ReactNode } from "react";
import { useAppStore } from "../store";
import type { StreamEvent, Phase, CIStageResult } from "../api/types";

interface WebSocketProviderProps {
  children: ReactNode;
  projectId: string | null;
  reconnectAttempts?: number;
  reconnectDelay?: number;
}

/**
 * WebSocketProvider manages the WebSocket connection for streaming events.
 * It automatically connects when a projectId is provided and handles
 * reconnection with exponential backoff.
 */
export function WebSocketProvider({
  children,
  projectId,
  reconnectAttempts = 5,
  reconnectDelay = 1000,
}: WebSocketProviderProps) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectCountRef = useRef(0);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Store actions
  const setWsConnected = useAppStore((state) => state.setWsConnected);
  const addStreamEvent = useAppStore((state) => state.addStreamEvent);
  const addTerminalLine = useAppStore((state) => state.addTerminalLine);
  const updateProjectPhase = useAppStore((state) => state.updateProjectPhase);
  const appendChat = useAppStore((state) => state.appendChat);
  const setDagState = useAppStore((state) => state.setDagState);
  const addActiveAgent = useAppStore((state) => state.addActiveAgent);
  const removeActiveAgent = useAppStore((state) => state.removeActiveAgent);
  const updateActiveAgent = useAppStore((state) => state.updateActiveAgent);
  const addCIStage = useAppStore((state) => state.addCIStage);
  const setCurrentCIStage = useAppStore((state) => state.setCurrentCIStage);
  const updateCurrentState = useAppStore((state) => state.updateCurrentState);

  const handleEvent = useCallback(
    (event: StreamEvent) => {
      // Add to stream events
      addStreamEvent(event);

      // Handle specific event types
      switch (event.type) {
        case "PHASE_CHANGE": {
          const payload = event.payload as {
            previous_phase: string;
            current_phase: string;
            message?: string;
          };
          if (projectId) {
            updateProjectPhase(projectId, payload.current_phase as Phase);
          }
          addTerminalLine(
            `📍 Phase changed: ${payload.previous_phase} → ${payload.current_phase}`
          );
          break;
        }

        case "CHAT_MESSAGE": {
          const payload = event.payload as {
            role: string;
            agent?: string;
            content: string;
          };
          const prefix = payload.role === "user" ? "You" : payload.agent || "Agent";
          appendChat(`${prefix}: ${payload.content}`);
          break;
        }

        case "FILE_START": {
          const payload = event.payload as { file_path: string; language: string };
          addTerminalLine(`📝 Generating: ${payload.file_path} (${payload.language})`);
          break;
        }

        case "CODE_CHUNK": {
          const payload = event.payload as {
            file_path: string;
            chunk_index: number;
            content: string;
            is_complete: boolean;
          };
          if (payload.is_complete) {
            addTerminalLine(`✅ Complete: ${payload.file_path}`);
          }
          break;
        }

        case "CODE_COMPLETE": {
          const payload = event.payload as { file_path: string };
          addTerminalLine(`✅ Generated: ${payload.file_path}`);
          break;
        }

        case "ERROR": {
          const payload = event.payload as { code: string; message: string };
          addTerminalLine(`❌ Error [${payload.code}]: ${payload.message}`);
          break;
        }

        // Multi-agent events
        case "DAG_UPDATE": {
          const payload = event.payload as {
            id: string;
            project_id: string;
            tasks: unknown[];
            total: number;
            completed: number;
            failed: number;
            running: number;
            pending: number;
          };
          setDagState(payload as Parameters<typeof setDagState>[0]);
          addTerminalLine(`📊 DAG Update: ${payload.completed}/${payload.total} tasks`);
          break;
        }

        case "TASK_STARTED": {
          const payload = event.payload as {
            task_id: string;
            task_name: string;
            agent_id?: string;
          };
          addTerminalLine(`▶️ Task Started: ${payload.task_name}`);
          if (payload.agent_id) {
            addActiveAgent({
              task_id: payload.task_id,
              agent_id: payload.agent_id,
              agent_name: payload.agent_id,
              task_name: payload.task_name,
              started_at: event.timestamp,
            });
          }
          break;
        }

        case "TASK_COMPLETED": {
          const payload = event.payload as { task_id: string; task_name: string };
          addTerminalLine(`✅ Task Completed: ${payload.task_name}`);
          removeActiveAgent(payload.task_id);
          break;
        }

        case "TASK_FAILED": {
          const payload = event.payload as {
            task_id: string;
            task_name: string;
            error: string;
          };
          addTerminalLine(`❌ Task Failed: ${payload.task_name} - ${payload.error}`);
          removeActiveAgent(payload.task_id);
          break;
        }

        case "AGENT_ASSIGNED": {
          const payload = event.payload as {
            task_id: string;
            agent_id: string;
            agent_name: string;
          };
          addTerminalLine(
            `🤖 Agent Assigned: ${payload.agent_name} → ${payload.task_id.slice(0, 8)}`
          );
          updateActiveAgent(payload.task_id, {
            agent_id: payload.agent_id,
            agent_name: payload.agent_name,
          });
          break;
        }

        case "CI_STAGE_START": {
          const payload = event.payload as { stage: string };
          setCurrentCIStage(payload.stage as "LINT" | "BUILD" | "TEST");
          addTerminalLine(`🔄 CI Stage: ${payload.stage} starting...`);
          break;
        }

        case "CI_STAGE_COMPLETE":
        case "CI_STAGE_FAILED": {
          const payload = event.payload as {
            stage: string;
            success?: boolean;
            output?: string;
            error?: string;
          };
          setCurrentCIStage(null);
          const result: CIStageResult = {
            stage: payload.stage as "LINT" | "BUILD" | "TEST",
            success: event.type === "CI_STAGE_COMPLETE",
            output: payload.output,
            error: payload.error,
          };
          addCIStage(result);
          if (event.type === "CI_STAGE_COMPLETE") {
            addTerminalLine(`✅ CI Stage: ${payload.stage} passed`);
          } else {
            addTerminalLine(`❌ CI Stage: ${payload.stage} failed - ${payload.error}`);
          }
          break;
        }

        case "EXECUTION_STARTED": {
          const payload = event.payload as { message: string };
          addTerminalLine(`🚀 Execution Started: ${payload.message}`);
          break;
        }

        case "EXECUTION_COMPLETED": {
          const payload = event.payload as { message: string };
          addTerminalLine(`🏁 Execution Completed: ${payload.message}`);
          break;
        }

        case "EXECUTION_FAILED": {
          const payload = event.payload as { message: string };
          addTerminalLine(`💥 Execution Failed: ${payload.message}`);
          break;
        }

        case "BUILD_START":
          addTerminalLine("🔨 Build starting...");
          break;

        case "BUILD_OUTPUT": {
          const payload = event.payload as { line: string; is_stderr: boolean };
          addTerminalLine(payload.is_stderr ? `⚠️ ${payload.line}` : payload.line);
          break;
        }

        case "BUILD_SUCCESS":
          addTerminalLine("✅ Build successful");
          break;

        case "BUILD_ERROR": {
          const payload = event.payload as { message: string };
          addTerminalLine(`❌ Build failed: ${payload.message}`);
          break;
        }

        case "PREVIEW_READY":
          addTerminalLine("🌐 Preview is ready");
          break;
      }
    },
    [
      projectId,
      addStreamEvent,
      addTerminalLine,
      updateProjectPhase,
      appendChat,
      setDagState,
      addActiveAgent,
      removeActiveAgent,
      updateActiveAgent,
      addCIStage,
      setCurrentCIStage,
    ]
  );

  const connect = useCallback(() => {
    if (!projectId) return;
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/pods/${projectId}/stream`;

    console.log(`Connecting to WebSocket: ${wsUrl}`);

    try {
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        console.log("WebSocket connected");
        setWsConnected(true);
        reconnectCountRef.current = 0;
        addTerminalLine("🔌 Connected to workflow stream");
      };

      ws.onmessage = (event) => {
        try {
          const streamEvent: StreamEvent = JSON.parse(event.data);
          handleEvent(streamEvent);
        } catch (err) {
          console.error("Failed to parse WebSocket message:", err);
        }
      };

      ws.onclose = (event) => {
        console.log(`WebSocket closed: ${event.code} ${event.reason}`);
        setWsConnected(false);
        wsRef.current = null;

        // Attempt reconnect
        if (reconnectCountRef.current < reconnectAttempts) {
          reconnectCountRef.current++;
          const delay = reconnectDelay * Math.pow(2, reconnectCountRef.current - 1);
          console.log(`Reconnecting in ${delay}ms (attempt ${reconnectCountRef.current})`);

          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, delay);
        }
      };

      ws.onerror = (error) => {
        console.error("WebSocket error:", error);
      };

      wsRef.current = ws;
    } catch (err) {
      console.error("Failed to create WebSocket:", err);
    }
  }, [
    projectId,
    handleEvent,
    setWsConnected,
    addTerminalLine,
    reconnectAttempts,
    reconnectDelay,
  ]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    setWsConnected(false);
    reconnectCountRef.current = reconnectAttempts; // Prevent reconnect
  }, [setWsConnected, reconnectAttempts]);

  // Connect when projectId changes
  useEffect(() => {
    if (projectId) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [projectId, connect, disconnect]);

  return <>{children}</>;
}
