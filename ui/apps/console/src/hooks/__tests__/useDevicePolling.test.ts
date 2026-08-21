import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useDevicePolling } from "../useDevicePolling";
import { mockSdkResponse } from "@/tests/sdk";

const mockGetStatusDevices = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return { ...actual, getStatusDevices: mockGetStatusDevices };
});

const defaultStats = {
  registered_devices: 0,
  online_devices: 0,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

function mockPollResponse(overrides: Partial<typeof defaultStats> = {}) {
  mockGetStatusDevices.mockResolvedValue(
    mockSdkResponse({ ...defaultStats, ...overrides }),
  );
}

beforeEach(() => {
  vi.useFakeTimers();
  vi.clearAllMocks();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("useDevicePolling", () => {
  describe("initial state", () => {
    it("starts with isPolling=false", () => {
      const { result } = renderHook(() =>
        useDevicePolling({ onPoll: () => false }),
      );
      expect(result.current.isPolling).toBe(false);
    });
  });

  describe("start()", () => {
    it("sets isPolling=true when started", () => {
      const { result } = renderHook(() =>
        useDevicePolling({ onPoll: () => false }),
      );

      act(() => {
        result.current.start();
      });

      expect(result.current.isPolling).toBe(true);
    });

    it("is idempotent — calling start() twice does not double-start", async () => {
      mockPollResponse();

      const onPoll = vi.fn().mockReturnValue(false);
      const { result } = renderHook(() =>
        useDevicePolling({ onPoll, initialInterval: 100 }),
      );

      act(() => {
        result.current.start();
        result.current.start();
      });

      await act(async () => {
        vi.advanceTimersByTime(100);
        await Promise.resolve();
      });

      // onPoll should be called exactly once, not twice
      expect(onPoll).toHaveBeenCalledTimes(1);
    });
  });

  describe("stop()", () => {
    it("sets isPolling=false when stopped", () => {
      const { result } = renderHook(() =>
        useDevicePolling({ onPoll: () => false }),
      );

      act(() => {
        result.current.start();
      });
      act(() => {
        result.current.stop();
      });

      expect(result.current.isPolling).toBe(false);
    });

    it("cancels pending timeout when stopped", async () => {
      mockPollResponse();
      const onPoll = vi.fn().mockReturnValue(false);

      const { result } = renderHook(() =>
        useDevicePolling({ onPoll, initialInterval: 1000 }),
      );

      act(() => {
        result.current.start();
      });

      // Trigger first poll
      await act(async () => {
        vi.advanceTimersByTime(1000);
        await Promise.resolve();
      });

      expect(onPoll).toHaveBeenCalledTimes(1);

      // Stop before second poll fires
      act(() => {
        result.current.stop();
      });

      await act(async () => {
        vi.advanceTimersByTime(5000);
        await Promise.resolve();
      });

      // Should still be 1 — stopped
      expect(onPoll).toHaveBeenCalledTimes(1);
    });
  });

  describe("polling behavior", () => {
    it("calls getStatusDevices after the initial interval", async () => {
      mockPollResponse();

      const { result } = renderHook(() =>
        useDevicePolling({ onPoll: () => false, initialInterval: 2000 }),
      );

      act(() => {
        result.current.start();
      });

      expect(mockGetStatusDevices).not.toHaveBeenCalled();

      await act(async () => {
        vi.advanceTimersByTime(2000);
        await Promise.resolve();
      });

      expect(mockGetStatusDevices).toHaveBeenCalledTimes(1);
    });

    it("passes stats to the onPoll callback", async () => {
      mockPollResponse({ pending_devices: 3 });

      const onPoll = vi.fn().mockReturnValue(false);

      const { result } = renderHook(() =>
        useDevicePolling({ onPoll, initialInterval: 100 }),
      );

      act(() => {
        result.current.start();
      });

      await act(async () => {
        vi.advanceTimersByTime(100);
        await Promise.resolve();
      });

      expect(onPoll).toHaveBeenCalledWith({
        ...defaultStats,
        pending_devices: 3,
      });
    });

    it("stops polling when onPoll returns true", async () => {
      mockPollResponse();
      const onPoll = vi.fn().mockReturnValue(true);

      const { result } = renderHook(() =>
        useDevicePolling({ onPoll, initialInterval: 100 }),
      );

      act(() => {
        result.current.start();
      });

      await act(async () => {
        vi.advanceTimersByTime(100);
        await Promise.resolve();
      });

      expect(result.current.isPolling).toBe(false);
    });

    it("continues polling when onPoll returns false", async () => {
      mockPollResponse();
      const onPoll = vi.fn().mockReturnValue(false);

      const { result } = renderHook(() =>
        useDevicePolling({ onPoll, initialInterval: 100, backoffFactor: 1 }),
      );

      act(() => {
        result.current.start();
      });

      // First poll
      await act(async () => {
        vi.advanceTimersByTime(100);
        await Promise.resolve();
      });

      // Second poll
      await act(async () => {
        vi.advanceTimersByTime(100);
        await Promise.resolve();
      });

      expect(onPoll).toHaveBeenCalledTimes(2);
    });
  });

  describe("exponential backoff", () => {
    it("applies backoff factor to subsequent intervals", async () => {
      mockPollResponse();
      const onPoll = vi.fn().mockReturnValue(false);

      const { result } = renderHook(() =>
        useDevicePolling({
          onPoll,
          initialInterval: 1000,
          backoffFactor: 2,
          maxInterval: 10000,
        }),
      );

      act(() => {
        result.current.start();
      });

      // First poll fires at t=1000
      await act(async () => {
        vi.advanceTimersByTime(1000);
        await Promise.resolve();
      });
      expect(onPoll).toHaveBeenCalledTimes(1);

      // Second poll fires at t=1000+2000=3000 (backoff: 1000 * 2 = 2000)
      await act(async () => {
        vi.advanceTimersByTime(2000);
        await Promise.resolve();
      });
      expect(onPoll).toHaveBeenCalledTimes(2);
    });

    it("caps backoff at maxInterval", async () => {
      mockPollResponse();
      const onPoll = vi.fn().mockReturnValue(false);

      const { result } = renderHook(() =>
        useDevicePolling({
          onPoll,
          initialInterval: 1000,
          backoffFactor: 100,
          maxInterval: 2000,
        }),
      );

      act(() => {
        result.current.start();
      });

      // First poll
      await act(async () => {
        vi.advanceTimersByTime(1000);
        await Promise.resolve();
      });
      expect(onPoll).toHaveBeenCalledTimes(1);

      // Second poll should fire at 2000ms (capped), not 100000ms
      await act(async () => {
        vi.advanceTimersByTime(2000);
        await Promise.resolve();
      });
      expect(onPoll).toHaveBeenCalledTimes(2);
    });
  });

  describe("error handling", () => {
    it("continues polling after getStatusDevices throws", async () => {
      mockPollResponse();
      mockGetStatusDevices.mockRejectedValueOnce(new Error("network error"));

      const onPoll = vi.fn().mockReturnValue(false);

      const { result } = renderHook(() =>
        useDevicePolling({
          onPoll,
          initialInterval: 100,
          backoffFactor: 1,
          maxInterval: 100,
        }),
      );

      act(() => {
        result.current.start();
      });

      // First poll fails (no onPoll call)
      await act(async () => {
        vi.advanceTimersByTime(100);
        await Promise.resolve();
      });
      expect(onPoll).toHaveBeenCalledTimes(0);

      // Second poll succeeds
      await act(async () => {
        vi.advanceTimersByTime(100);
        await Promise.resolve();
      });
      expect(onPoll).toHaveBeenCalledTimes(1);
    });

    it("does not crash when stop is called mid-flight", async () => {
      let resolveStats!: (v: unknown) => void;
      mockGetStatusDevices.mockReturnValue(
        new Promise((r) => {
          resolveStats = r;
        }),
      );

      const onPoll = vi.fn().mockReturnValue(false);

      const { result } = renderHook(() =>
        useDevicePolling({ onPoll, initialInterval: 100 }),
      );

      act(() => {
        result.current.start();
      });

      act(() => {
        vi.advanceTimersByTime(100);
      });

      // Stop while request is in flight
      act(() => {
        result.current.stop();
      });

      // Resolve the promise — should not call onPoll or throw
      await act(async () => {
        resolveStats(mockSdkResponse(defaultStats));
        await Promise.resolve();
      });

      expect(onPoll).not.toHaveBeenCalled();
    });
  });

  describe("cleanup on unmount", () => {
    it("clears the pending timeout when the hook unmounts", async () => {
      mockPollResponse();
      const onPoll = vi.fn().mockReturnValue(false);

      const { result, unmount } = renderHook(() =>
        useDevicePolling({ onPoll, initialInterval: 1000 }),
      );

      act(() => {
        result.current.start();
      });

      unmount();

      // Advancing timers after unmount should not trigger onPoll
      await act(async () => {
        vi.advanceTimersByTime(5000);
        await Promise.resolve();
      });

      expect(onPoll).not.toHaveBeenCalled();
    });
  });
});
