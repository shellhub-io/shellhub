import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { useIdleControls } from "../useIdleControls";

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("useIdleControls", () => {
  it("brings the controls up on activity and lets them go after two seconds", () => {
    const { result } = renderHook(() => useIdleControls());

    act(() => {
      result.current.wake();
    });
    expect(result.current.shown(false)).toBe(true);

    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(result.current.shown(false)).toBe(false);
  });

  it("restarts the wait on every bit of activity", () => {
    const { result } = renderHook(() => useIdleControls());
    act(() => {
      result.current.wake();
    });

    act(() => {
      vi.advanceTimersByTime(1500);
    });
    act(() => {
      result.current.wake();
    });
    act(() => {
      vi.advanceTimersByTime(1500);
    });

    expect(result.current.shown(false)).toBe(true);
  });

  it("holds them while the pointer rests on them", () => {
    const { result } = renderHook(() => useIdleControls());

    act(() => {
      result.current.barProps.onPointerEnter();
    });
    act(() => {
      vi.advanceTimersByTime(5000);
    });

    expect(result.current.shown(false)).toBe(true);
  });

  it("keeps them up while the caller pins them", () => {
    const { result } = renderHook(() => useIdleControls());

    expect(result.current.shown(true)).toBe(true);
  });

  it("puts them away until the next activity, even while pinned", () => {
    const { result } = renderHook(() => useIdleControls());

    act(() => {
      result.current.stow();
    });
    expect(result.current.shown(true)).toBe(false);

    act(() => {
      result.current.wake();
    });
    expect(result.current.shown(true)).toBe(true);
  });
});
