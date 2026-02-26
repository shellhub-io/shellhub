import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useCountdown } from "../useCountdown";

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.restoreAllMocks();
  vi.useRealTimers();
});

describe("useCountdown", () => {
  it("returns empty string when targetTimestamp is null", () => {
    const { result } = renderHook(() => useCountdown(null));

    expect(result.current.timeLeft).toBe("");
    expect(result.current.isExpired).toBe(false);
  });

  it("calculates time left correctly", () => {
    const now = Date.now();
    vi.setSystemTime(now);

    // 5 minutes 30 seconds in the future
    const target = Math.floor(now / 1000) + 330;

    const { result } = renderHook(() => useCountdown(target));

    expect(result.current.timeLeft).toBe("5 minutes 30 seconds");
    expect(result.current.isExpired).toBe(false);
  });

  it("handles singular minute and second", () => {
    const now = Date.now();
    vi.setSystemTime(now);

    // 1 minute 1 second in the future
    const target = Math.floor(now / 1000) + 61;

    const { result } = renderHook(() => useCountdown(target));

    expect(result.current.timeLeft).toBe("1 minute 1 second");
  });

  it("handles plural minutes", () => {
    const now = Date.now();
    vi.setSystemTime(now);

    // 2 minutes 0 seconds in the future
    const target = Math.floor(now / 1000) + 120;

    const { result } = renderHook(() => useCountdown(target));

    expect(result.current.timeLeft).toBe("2 minutes 0 seconds");
  });

  it("updates countdown every second", () => {
    const now = Date.now();
    vi.setSystemTime(now);

    // 10 seconds in the future
    const target = Math.floor(now / 1000) + 10;

    const { result } = renderHook(() => useCountdown(target));

    expect(result.current.timeLeft).toBe("0 minutes 10 seconds");

    // Advance 3 seconds
    act(() => {
      vi.advanceTimersByTime(3000);
    });

    expect(result.current.timeLeft).toBe("0 minutes 7 seconds");
  });

  it("marks as expired when time reaches zero", () => {
    const now = Date.now();
    vi.setSystemTime(now);

    // 2 seconds in the future
    const target = Math.floor(now / 1000) + 2;

    const { result } = renderHook(() => useCountdown(target));

    expect(result.current.isExpired).toBe(false);

    // Advance past the expiry
    act(() => {
      vi.advanceTimersByTime(3000);
    });

    expect(result.current.timeLeft).toBe("0 seconds");
    expect(result.current.isExpired).toBe(true);
  });

  it("stops interval after expiry", () => {
    const now = Date.now();
    vi.setSystemTime(now);

    const target = Math.floor(now / 1000) + 1;

    const { result } = renderHook(() => useCountdown(target));

    // Advance to expiry
    act(() => {
      vi.advanceTimersByTime(1500);
    });

    const expiredTime = result.current.timeLeft;

    // Advance more - should not update
    act(() => {
      vi.advanceTimersByTime(5000);
    });

    expect(result.current.timeLeft).toBe(expiredTime);
  });

  it("cleans up interval on unmount", () => {
    const now = Date.now();
    vi.setSystemTime(now);
    const target = Math.floor(now / 1000) + 60;

    const { unmount } = renderHook(() => useCountdown(target));

    const timerCount = vi.getTimerCount();
    unmount();

    expect(vi.getTimerCount()).toBeLessThan(timerCount);
  });

  it("handles already expired timestamp", () => {
    const now = Date.now();
    vi.setSystemTime(now);

    // 5 seconds in the past
    const target = Math.floor(now / 1000) - 5;

    const { result } = renderHook(() => useCountdown(target));

    expect(result.current.timeLeft).toBe("0 seconds");
    expect(result.current.isExpired).toBe(true);
  });
});
