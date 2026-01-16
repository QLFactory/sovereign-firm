"use client";

import { useEffect, useRef, useState, useCallback } from "react";

// Stream event types matching backend protocol
export type StreamEventType =
    | "CONNECTED"
    | "DISCONNECTED"
    | "ERROR"
    | "PHASE_CHANGE"
    | "AGENT_SPAWN"
    | "AGENT_DONE"
    | "CODE_CHUNK"
    | "CODE_COMPLETE"
    | "FILE_START"
    | "FILE_END"
    | "CHAT_MESSAGE"
    | "CHAT_TYPING"
    | "BUILD_START"
    | "BUILD_OUTPUT"
    | "BUILD_SUCCESS"
    | "BUILD_ERROR"
    | "TEST_START"
    | "TEST_RESULT"
    | "PREVIEW_READY"
    // Phase 4: Multi-Agent Coordination Events
    | "EXECUTION_STARTED"
    | "EXECUTION_COMPLETED"
    | "EXECUTION_FAILED"
    | "TASK_STARTED"
    | "TASK_COMPLETED"
    | "TASK_FAILED"
    | "TASK_RETRY"
    | "AGENT_ASSIGNED"
    | "BRANCH_MERGED"
    | "MERGE_FAILED"
    | "CI_STAGE_START"
    | "CI_STAGE_COMPLETE"
    | "CI_STAGE_FAILED"
    | "DAG_UPDATE";

export interface StreamEvent {
    type: StreamEventType;
    timestamp: string;
    seq: number;
    payload?: Record<string, unknown>;
}

export interface UseStreamingOptions {
    workflowId: string | null;
    onEvent?: (event: StreamEvent) => void;
    onConnect?: () => void;
    onDisconnect?: () => void;
    onError?: (error: Error) => void;
    reconnectAttempts?: number;
    reconnectDelay?: number;
}

export interface UseStreamingResult {
    isConnected: boolean;
    lastEvent: StreamEvent | null;
    events: StreamEvent[];
    connect: () => void;
    disconnect: () => void;
}

export function useStreaming(options: UseStreamingOptions): UseStreamingResult {
    const {
        workflowId,
        onEvent,
        onConnect,
        onDisconnect,
        onError,
        reconnectAttempts = 5,
        reconnectDelay = 1000,
    } = options;

    const [isConnected, setIsConnected] = useState(false);
    const [lastEvent, setLastEvent] = useState<StreamEvent | null>(null);
    const [events, setEvents] = useState<StreamEvent[]>([]);

    const wsRef = useRef<WebSocket | null>(null);
    const reconnectCountRef = useRef(0);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    const connect = useCallback(() => {
        if (!workflowId) return;
        if (wsRef.current?.readyState === WebSocket.OPEN) return;

        // Determine WebSocket URL
        const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
        const host = window.location.host;
        const wsUrl = `${protocol}//${host}/api/pods/${workflowId}/stream`;

        console.log(`Connecting to WebSocket: ${wsUrl}`);

        try {
            const ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                console.log("WebSocket connected");
                setIsConnected(true);
                reconnectCountRef.current = 0;
                onConnect?.();
            };

            ws.onmessage = (event) => {
                try {
                    const streamEvent: StreamEvent = JSON.parse(event.data);
                    setLastEvent(streamEvent);
                    setEvents((prev) => [...prev, streamEvent]);
                    onEvent?.(streamEvent);
                } catch (err) {
                    console.error("Failed to parse WebSocket message:", err);
                }
            };

            ws.onclose = (event) => {
                console.log(`WebSocket closed: ${event.code} ${event.reason}`);
                setIsConnected(false);
                wsRef.current = null;
                onDisconnect?.();

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
                onError?.(new Error("WebSocket connection error"));
            };

            wsRef.current = ws;
        } catch (err) {
            console.error("Failed to create WebSocket:", err);
            onError?.(err as Error);
        }
    }, [workflowId, onEvent, onConnect, onDisconnect, onError, reconnectAttempts, reconnectDelay]);

    const disconnect = useCallback(() => {
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
        }

        if (wsRef.current) {
            wsRef.current.close();
            wsRef.current = null;
        }

        setIsConnected(false);
        reconnectCountRef.current = reconnectAttempts; // Prevent reconnect
    }, [reconnectAttempts]);

    // Connect when workflowId changes
    useEffect(() => {
        if (workflowId) {
            connect();
        }

        return () => {
            disconnect();
        };
    }, [workflowId, connect, disconnect]);

    return {
        isConnected,
        lastEvent,
        events,
        connect,
        disconnect,
    };
}

// Utility type guards for event payloads
export function isPhaseChangeEvent(event: StreamEvent): event is StreamEvent & {
    payload: { previous_phase: string; current_phase: string; message?: string };
} {
    return event.type === "PHASE_CHANGE" && event.payload !== undefined;
}

export function isCodeChunkEvent(event: StreamEvent): event is StreamEvent & {
    payload: { file_path: string; chunk_index: number; content: string; is_complete: boolean };
} {
    return event.type === "CODE_CHUNK" && event.payload !== undefined;
}

export function isChatMessageEvent(event: StreamEvent): event is StreamEvent & {
    payload: { role: string; agent?: string; content: string };
} {
    return event.type === "CHAT_MESSAGE" && event.payload !== undefined;
}

export function isFileStartEvent(event: StreamEvent): event is StreamEvent & {
    payload: { file_path: string; language: string };
} {
    return event.type === "FILE_START" && event.payload !== undefined;
}

export function isFileEndEvent(event: StreamEvent): event is StreamEvent & {
    payload: { file_path: string };
} {
    return event.type === "FILE_END" && event.payload !== undefined;
}

export function isBuildOutputEvent(event: StreamEvent): event is StreamEvent & {
    payload: { line: string; is_stderr: boolean };
} {
    return event.type === "BUILD_OUTPUT" && event.payload !== undefined;
}

export function isErrorEvent(event: StreamEvent): event is StreamEvent & {
    payload: { code: string; message: string; details?: string };
} {
    return event.type === "ERROR" && event.payload !== undefined;
}

// Phase 4: Multi-Agent Event Type Guards
export interface DAGTask {
    id: string;
    name: string;
    description?: string;
    type: string;
    status: "PENDING" | "READY" | "RUNNING" | "COMPLETED" | "FAILED" | "BLOCKED";
    dependencies: string[];
    assigned_to?: string;
    error?: string;
    retry_count?: number;
}

export interface DAGState {
    id: string;
    project_id: string;
    tasks: DAGTask[];
    total: number;
    completed: number;
    failed: number;
    running: number;
    pending: number;
}

export interface ActiveAgent {
    task_id: string;
    agent_id: string;
    agent_name: string;
    task_name: string;
    started_at: string;
}

export interface CIStageResult {
    stage: "LINT" | "BUILD" | "TEST";
    success: boolean;
    output?: string;
    error?: string;
    duration_ms?: number;
}

export function isTaskStartedEvent(event: StreamEvent): event is StreamEvent & {
    payload: { task_id: string; task_name: string; agent_id?: string };
} {
    return event.type === "TASK_STARTED" && event.payload !== undefined;
}

export function isTaskCompletedEvent(event: StreamEvent): event is StreamEvent & {
    payload: { task_id: string; task_name: string; output?: Record<string, unknown> };
} {
    return event.type === "TASK_COMPLETED" && event.payload !== undefined;
}

export function isTaskFailedEvent(event: StreamEvent): event is StreamEvent & {
    payload: { task_id: string; task_name: string; error: string };
} {
    return event.type === "TASK_FAILED" && event.payload !== undefined;
}

export function isAgentAssignedEvent(event: StreamEvent): event is StreamEvent & {
    payload: { task_id: string; agent_id: string; agent_name: string };
} {
    return event.type === "AGENT_ASSIGNED" && event.payload !== undefined;
}

export function isDAGUpdateEvent(event: StreamEvent): event is StreamEvent & {
    payload: DAGState;
} {
    return event.type === "DAG_UPDATE" && event.payload !== undefined;
}

export function isCIStageEvent(event: StreamEvent): event is StreamEvent & {
    payload: { task_id: string; stage: string; success?: boolean; output?: string; error?: string };
} {
    return (
        (event.type === "CI_STAGE_START" ||
            event.type === "CI_STAGE_COMPLETE" ||
            event.type === "CI_STAGE_FAILED") &&
        event.payload !== undefined
    );
}

export function isExecutionEvent(event: StreamEvent): event is StreamEvent & {
    payload: { message: string };
} {
    return (
        (event.type === "EXECUTION_STARTED" ||
            event.type === "EXECUTION_COMPLETED" ||
            event.type === "EXECUTION_FAILED") &&
        event.payload !== undefined
    );
}
