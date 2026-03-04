import {
  describe,
  it,
  expect,
  vi,
  beforeEach,
  afterEach,
} from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useDevicePolling } from "../useDevicePolling";

vi.mock("@/api/stats", () => ({
  getStats: vi.fn(),
}));

import { getStats } from "@/api/stats";

const mockGetStats = vi.mocked(getStats);

const defaultStats = {
  registered_devices: 0,
  online_devices: 0,
  active_sessions: 0,
  pending_devices: 0,
  rejected_devices: 0,
};

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
      mockGetStats.mockResolvedValue(defaultStats);

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
      mockGetStats.mockResolvedValue(defaultStats);
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
    it("calls getStats after the initial interval", async () => {
      mockGetStats.mockResolvedValue(defaultStats);

      const { result } = renderHook(() =>
        useDevicePolling({ onPoll: () => false, initialInterval: 2000 }),
      );

      act(() => {
        result.current.start();
      });

      expect(mockGetStats).not.toHaveBeenCalled();

      await act(async () => {
        vi.advanceTimersByTime(2000);
        await Promise.resolve();
      });

      expect(mockGetStats).toHaveBeenCalledTimes(1);
    });

    it("passes stats to the onPoll callback", async () => {
      const stats = { ...defaultStats, pending_devices: 3 };
      mockGetStats.mockResolvedValue(stats);

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

      expect(onPoll).toHaveBeenCalledWith(stats);
    });

    it("stops polling when onPoll returns true", async () => {
      mockGetStats.mockResolvedValue(defaultStats);
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
      mockGetStats.mockResolvedValue(defaultStats);
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
      mockGetStats.mockResolvedValue(defaultStats);
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
      mockGetStats.mockResolvedValue(defaultStats);
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
    it("continues polling after getStats throws", async () => {
      mockGetStats
        .mockRejectedValueOnce(new Error("network error"))
        .mockResolvedValue(defaultStats);

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
      let resolveStats!: (v: typeof defaultStats) => void;
      mockGetStats.mockReturnValue(
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
        resolveStats(defaultStats);
        await Promise.resolve();
      });

      expect(onPoll).not.toHaveBeenCalled();
    });
  });

  describe("cleanup on unmount", () => {
    it("clears the pending timeout when the hook unmounts", async () => {
      mockGetStats.mockResolvedValue(defaultStats);
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
