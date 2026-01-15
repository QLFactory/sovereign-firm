import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import {
  useStreaming,
  StreamEvent,
  isPhaseChangeEvent,
  isCodeChunkEvent,
  isChatMessageEvent,
  isFileStartEvent,
  isFileEndEvent,
  isBuildOutputEvent,
  isErrorEvent,
} from "./useStreaming";

// Access the mock WebSocket
declare global {
  var MockWebSocket: {
    getLastInstance(): {
      simulateOpen(): void;
      simulateMessage(data: unknown): void;
      simulateClose(code?: number, reason?: string): void;
      simulateError(): void;
      readyState: number;
      url: string;
    } | undefined;
    clearInstances(): void;
    OPEN: number;
    CLOSED: number;
  };
}

describe("useStreaming hook", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    MockWebSocket.clearInstances();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe("connection lifecycle", () => {
    it("should not connect when workflowId is null", () => {
      const { result } = renderHook(() =>
        useStreaming({ workflowId: null })
      );

      expect(result.current.isConnected).toBe(false);
      expect(MockWebSocket.getLastInstance()).toBeUndefined();
    });

    it("should connect when workflowId is provided", () => {
      renderHook(() =>
        useStreaming({ workflowId: "test-workflow-123" })
      );

      const ws = MockWebSocket.getLastInstance();
      expect(ws).toBeDefined();
      expect(ws?.url).toContain("test-workflow-123");
    });

    it("should set isConnected to true on open", async () => {
      const { result } = renderHook(() =>
        useStreaming({ workflowId: "test-workflow" })
      );

      const ws = MockWebSocket.getLastInstance();
      expect(ws).toBeDefined();

      act(() => {
        ws?.simulateOpen();
      });

      expect(result.current.isConnected).toBe(true);
    });

    it("should call onConnect callback when connected", () => {
      const onConnect = vi.fn();

      renderHook(() =>
        useStreaming({ workflowId: "test-workflow", onConnect })
      );

      const ws = MockWebSocket.getLastInstance();
      act(() => {
        ws?.simulateOpen();
      });

      expect(onConnect).toHaveBeenCalledTimes(1);
    });

    it("should call onDisconnect callback when connection closes", () => {
      const onDisconnect = vi.fn();

      renderHook(() =>
        useStreaming({ workflowId: "test-workflow", onDisconnect })
      );

      const ws = MockWebSocket.getLastInstance();
      act(() => {
        ws?.simulateOpen();
      });

      act(() => {
        ws?.simulateClose();
      });

      expect(onDisconnect).toHaveBeenCalledTimes(1);
    });

    it("should set isConnected to false on close", () => {
      const { result } = renderHook(() =>
        useStreaming({ workflowId: "test-workflow" })
      );

      const ws = MockWebSocket.getLastInstance();
      act(() => {
        ws?.simulateOpen();
      });

      expect(result.current.isConnected).toBe(true);

      act(() => {
        ws?.simulateClose();
      });

      expect(result.current.isConnected).toBe(false);
    });
  });

  describe("message handling", () => {
    it("should update lastEvent when message received", () => {
      const { result } = renderHook(() =>
        useStreaming({ workflowId: "test-workflow" })
      );

      const ws = MockWebSocket.getLastInstance();
      act(() => {
        ws?.simulateOpen();
      });

      const testEvent: StreamEvent = {
        type: "CHAT_MESSAGE",
        timestamp: new Date().toISOString(),
        seq: 1,
        payload: { role: "agent", content: "Hello" },
      };

      act(() => {
        ws?.simulateMessage(testEvent);
      });

      expect(result.current.lastEvent).toEqual(testEvent);
    });

    it("should accumulate events in events array", () => {
      const { result } = renderHook(() =>
        useStreaming({ workflowId: "test-workflow" })
      );

      const ws = MockWebSocket.getLastInstance();
      act(() => {
        ws?.simulateOpen();
      });

      const event1: StreamEvent = {
        type: "PHASE_CHANGE",
        timestamp: new Date().toISOString(),
        seq: 1,
        payload: { previous_phase: "INIT", current_phase: "DISCOVERY" },
      };

      const event2: StreamEvent = {
        type: "CHAT_MESSAGE",
        timestamp: new Date().toISOString(),
        seq: 2,
        payload: { role: "agent", content: "Working..." },
      };

      act(() => {
        ws?.simulateMessage(event1);
        ws?.simulateMessage(event2);
      });

      expect(result.current.events).toHaveLength(2);
      expect(result.current.events[0]).toEqual(event1);
      expect(result.current.events[1]).toEqual(event2);
    });

    it("should call onEvent callback for each message", () => {
      const onEvent = vi.fn();

      renderHook(() =>
        useStreaming({ workflowId: "test-workflow", onEvent })
      );

      const ws = MockWebSocket.getLastInstance();
      act(() => {
        ws?.simulateOpen();
      });

      const testEvent: StreamEvent = {
        type: "CODE_CHUNK",
        timestamp: new Date().toISOString(),
        seq: 1,
        payload: { file_path: "/src/App.jsx", chunk_index: 0, content: "code" },
      };

      act(() => {
        ws?.simulateMessage(testEvent);
      });

      expect(onEvent).toHaveBeenCalledWith(testEvent);
    });
  });

  describe("reconnection", () => {
    it("should attempt reconnect on disconnect", async () => {
      renderHook(() =>
        useStreaming({
          workflowId: "test-workflow",
          reconnectAttempts: 3,
          reconnectDelay: 1000,
        })
      );

      const ws1 = MockWebSocket.getLastInstance();
      act(() => {
        ws1?.simulateOpen();
      });

      // Close connection
      act(() => {
        ws1?.simulateClose();
      });

      // Fast-forward past reconnect delay
      await act(async () => {
        vi.advanceTimersByTime(1000);
      });

      // Should have created a new WebSocket
      const ws2 = MockWebSocket.getLastInstance();
      expect(ws2).toBeDefined();
      expect(ws2).not.toBe(ws1);
    });

    it("should use exponential backoff for reconnection", async () => {
      const reconnectDelay = 1000;

      renderHook(() =>
        useStreaming({
          workflowId: "test-workflow",
          reconnectAttempts: 5,
          reconnectDelay,
        })
      );

      // First connection
      let ws = MockWebSocket.getLastInstance();
      act(() => {
        ws?.simulateOpen();
        ws?.simulateClose();
      });

      // First reconnect after 1000ms
      await act(async () => {
        vi.advanceTimersByTime(reconnectDelay);
      });

      ws = MockWebSocket.getLastInstance();
      act(() => {
        ws?.simulateOpen();
        ws?.simulateClose();
      });

      // Second reconnect should be after 2000ms (exponential backoff)
      await act(async () => {
        vi.advanceTimersByTime(reconnectDelay * 2);
      });

      // Should have created new WebSocket
      expect(MockWebSocket.getLastInstance()).toBeDefined();
    });

    it("should reset reconnect count on successful connection", async () => {
      const onConnect = vi.fn();

      renderHook(() =>
        useStreaming({
          workflowId: "test-workflow",
          onConnect,
          reconnectAttempts: 3,
          reconnectDelay: 1000,
        })
      );

      const ws1 = MockWebSocket.getLastInstance();
      act(() => {
        ws1?.simulateOpen();
      });

      expect(onConnect).toHaveBeenCalledTimes(1);

      // Disconnect and reconnect
      act(() => {
        ws1?.simulateClose();
      });

      await act(async () => {
        vi.advanceTimersByTime(1000);
      });

      const ws2 = MockWebSocket.getLastInstance();
      act(() => {
        ws2?.simulateOpen();
      });

      // Should be called again
      expect(onConnect).toHaveBeenCalledTimes(2);
    });
  });

  describe("disconnect", () => {
    it("should clean up on disconnect", () => {
      const { result } = renderHook(() =>
        useStreaming({ workflowId: "test-workflow" })
      );

      const ws = MockWebSocket.getLastInstance();
      act(() => {
        ws?.simulateOpen();
      });

      expect(result.current.isConnected).toBe(true);

      act(() => {
        result.current.disconnect();
      });

      expect(result.current.isConnected).toBe(false);
    });

    it("should prevent reconnection after manual disconnect", async () => {
      const { result } = renderHook(() =>
        useStreaming({
          workflowId: "test-workflow",
          reconnectAttempts: 3,
          reconnectDelay: 1000,
        })
      );

      const initialWs = MockWebSocket.getLastInstance();
      act(() => {
        initialWs?.simulateOpen();
      });

      // Manual disconnect
      act(() => {
        result.current.disconnect();
      });

      // Fast-forward past any reconnect delay
      await act(async () => {
        vi.advanceTimersByTime(5000);
      });

      // Should not have reconnected (only 1 instance total)
      expect(result.current.isConnected).toBe(false);
    });
  });

  describe("error handling", () => {
    it("should call onError callback on WebSocket error", () => {
      const onError = vi.fn();

      renderHook(() =>
        useStreaming({ workflowId: "test-workflow", onError })
      );

      const ws = MockWebSocket.getLastInstance();
      act(() => {
        ws?.simulateError();
      });

      expect(onError).toHaveBeenCalledWith(expect.any(Error));
    });
  });
});

// ==================== Type Guard Tests ====================

describe("Type Guards", () => {
  const baseEvent: StreamEvent = {
    type: "CONNECTED",
    timestamp: new Date().toISOString(),
    seq: 1,
  };

  describe("isPhaseChangeEvent", () => {
    it("should return true for PHASE_CHANGE event with payload", () => {
      const event: StreamEvent = {
        ...baseEvent,
        type: "PHASE_CHANGE",
        payload: { previous_phase: "INIT", current_phase: "DISCOVERY" },
      };
      expect(isPhaseChangeEvent(event)).toBe(true);
    });

    it("should return false for PHASE_CHANGE without payload", () => {
      const event: StreamEvent = { ...baseEvent, type: "PHASE_CHANGE" };
      expect(isPhaseChangeEvent(event)).toBe(false);
    });

    it("should return false for other event types", () => {
      const event: StreamEvent = {
        ...baseEvent,
        type: "CHAT_MESSAGE",
        payload: { content: "test" },
      };
      expect(isPhaseChangeEvent(event)).toBe(false);
    });
  });

  describe("isCodeChunkEvent", () => {
    it("should return true for CODE_CHUNK event with payload", () => {
      const event: StreamEvent = {
        ...baseEvent,
        type: "CODE_CHUNK",
        payload: {
          file_path: "/src/App.jsx",
          chunk_index: 0,
          content: "code",
          is_complete: false,
        },
      };
      expect(isCodeChunkEvent(event)).toBe(true);
    });

    it("should return false for CODE_CHUNK without payload", () => {
      const event: StreamEvent = { ...baseEvent, type: "CODE_CHUNK" };
      expect(isCodeChunkEvent(event)).toBe(false);
    });
  });

  describe("isChatMessageEvent", () => {
    it("should return true for CHAT_MESSAGE event with payload", () => {
      const event: StreamEvent = {
        ...baseEvent,
        type: "CHAT_MESSAGE",
        payload: { role: "agent", agent: "PM", content: "Hello" },
      };
      expect(isChatMessageEvent(event)).toBe(true);
    });

    it("should return false for CHAT_MESSAGE without payload", () => {
      const event: StreamEvent = { ...baseEvent, type: "CHAT_MESSAGE" };
      expect(isChatMessageEvent(event)).toBe(false);
    });
  });

  describe("isFileStartEvent", () => {
    it("should return true for FILE_START event with payload", () => {
      const event: StreamEvent = {
        ...baseEvent,
        type: "FILE_START",
        payload: { file_path: "/src/App.jsx", language: "jsx" },
      };
      expect(isFileStartEvent(event)).toBe(true);
    });

    it("should return false for FILE_START without payload", () => {
      const event: StreamEvent = { ...baseEvent, type: "FILE_START" };
      expect(isFileStartEvent(event)).toBe(false);
    });
  });

  describe("isFileEndEvent", () => {
    it("should return true for FILE_END event with payload", () => {
      const event: StreamEvent = {
        ...baseEvent,
        type: "FILE_END",
        payload: { file_path: "/src/App.jsx" },
      };
      expect(isFileEndEvent(event)).toBe(true);
    });

    it("should return false for FILE_END without payload", () => {
      const event: StreamEvent = { ...baseEvent, type: "FILE_END" };
      expect(isFileEndEvent(event)).toBe(false);
    });
  });

  describe("isBuildOutputEvent", () => {
    it("should return true for BUILD_OUTPUT event with payload", () => {
      const event: StreamEvent = {
        ...baseEvent,
        type: "BUILD_OUTPUT",
        payload: { line: "Building...", is_stderr: false },
      };
      expect(isBuildOutputEvent(event)).toBe(true);
    });

    it("should return false for BUILD_OUTPUT without payload", () => {
      const event: StreamEvent = { ...baseEvent, type: "BUILD_OUTPUT" };
      expect(isBuildOutputEvent(event)).toBe(false);
    });
  });

  describe("isErrorEvent", () => {
    it("should return true for ERROR event with payload", () => {
      const event: StreamEvent = {
        ...baseEvent,
        type: "ERROR",
        payload: { code: "E001", message: "Error occurred", details: "Details" },
      };
      expect(isErrorEvent(event)).toBe(true);
    });

    it("should return false for ERROR without payload", () => {
      const event: StreamEvent = { ...baseEvent, type: "ERROR" };
      expect(isErrorEvent(event)).toBe(false);
    });
  });
});
