import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useSettingSave } from "../useSettingSave";

const FAILURE = "Couldn't save. Try again.";

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("useSettingSave", () => {
  it("reports saved after a save succeeds, then settles", async () => {
    const { result } = renderHook(() => useSettingSave(FAILURE));

    await act(() => result.current.run(() => Promise.resolve()));
    expect(result.current.saved).toBe(true);
    expect(result.current.error).toBe("");

    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(result.current.saved).toBe(false);
  });

  it("reports the failure message when the save throws", async () => {
    const { result } = renderHook(() => useSettingSave(FAILURE));

    await act(() => result.current.run(() => Promise.reject(new Error("x"))));

    expect(result.current.error).toBe(FAILURE);
    expect(result.current.saved).toBe(false);
  });

  it("clears a failure when the setting is saved again", async () => {
    const { result } = renderHook(() => useSettingSave(FAILURE));

    await act(() => result.current.run(() => Promise.reject(new Error("x"))));
    await act(() => result.current.run(() => Promise.resolve()));

    expect(result.current.error).toBe("");
    expect(result.current.saved).toBe(true);
  });
});
