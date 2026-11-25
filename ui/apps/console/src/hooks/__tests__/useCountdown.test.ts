import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useCountdown } from "../useCountdown";

const NOW = 1_700_000_000;

function renderRecordingCountdown(end: number | null) {
  const seen: (number | null)[] = [];
  const { rerender } = renderHook<number | null, { end: number | null }>(
    ({ end }) => {
      const left = useCountdown(end);
      seen.push(left);
      return left;
    },
    { initialProps: { end } },
  );
  return { seen, rerender };
}

beforeEach(() => {
  vi.useFakeTimers();
  vi.setSystemTime(NOW * 1000);
});

afterEach(() => {
  vi.useRealTimers();
});

describe("useCountdown", () => {
  it.each([null, NaN])("is null and runs no timer for an end of %p", (end) => {
    const { result } = renderHook(() => useCountdown(end));
    expect(result.current).toBeNull();
    expect(vi.getTimerCount()).toBe(0);
  });

  it("is 0 at once for an end already past", () => {
    const { result } = renderHook(() => useCountdown(NOW - 5));
    expect(result.current).toBe(0);
  });

  it("counts the seconds left down each second, stopping at 0", () => {
    const { result } = renderHook(() => useCountdown(NOW + 3));
    expect(result.current).toBe(3);

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(result.current).toBe(2);

    act(() => {
      vi.advanceTimersByTime(5000);
    });
    expect(result.current).toBe(0);
    expect(vi.getTimerCount()).toBe(0);
  });

  it("stops its timer on unmount", () => {
    const { unmount } = renderHook(() => useCountdown(NOW + 60));
    unmount();
    expect(vi.getTimerCount()).toBe(0);
  });

  it("never shows the old end's count for a new end", () => {
    const { seen, rerender } = renderRecordingCountdown(NOW + 1);
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    seen.length = 0;

    rerender({ end: NOW + 60 });

    expect(seen).not.toContain(0);
    expect(seen.at(-1)).toBe(58);
  });

  it("never shows a count from before the end was cleared and set again", () => {
    const { seen, rerender } = renderRecordingCountdown(NOW + 60);
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    rerender({ end: null });
    act(() => {
      vi.advanceTimersByTime(3000);
    });
    seen.length = 0;

    rerender({ end: NOW + 60 });

    expect(seen).not.toContain(58);
    expect(seen.at(-1)).toBe(55);
  });
});
