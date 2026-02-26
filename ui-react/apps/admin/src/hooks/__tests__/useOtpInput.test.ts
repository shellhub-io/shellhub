import { describe, it, expect } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useOtpInput } from "../useOtpInput";

describe("useOtpInput", () => {
  it("initializes with empty code array", () => {
    const { result } = renderHook(() => useOtpInput(6));

    expect(result.current.code).toEqual(["", "", "", "", "", ""]);
    expect(result.current.getValue()).toBe("");
    expect(result.current.isComplete).toBe(false);
  });

  it("handles digit input and auto-advances", () => {
    const { result } = renderHook(() => useOtpInput(6));

    act(() => {
      result.current.handleChange(0, "5");
    });

    expect(result.current.code[0]).toBe("5");
    expect(result.current.getValue()).toBe("5");
  });

  it("rejects non-digit input", () => {
    const { result } = renderHook(() => useOtpInput(6));

    act(() => {
      result.current.handleChange(0, "a");
    });

    expect(result.current.code[0]).toBe("");
  });

  it("detects completion when all digits entered", () => {
    const { result } = renderHook(() => useOtpInput(6));

    act(() => {
      ["1", "2", "3", "4", "5", "6"].forEach((digit, i) => {
        result.current.handleChange(i, digit);
      });
    });

    expect(result.current.isComplete).toBe(true);
    expect(result.current.getValue()).toBe("123456");
  });

  it("resets to empty state", () => {
    const { result } = renderHook(() => useOtpInput(6));

    act(() => {
      ["1", "2", "3", "4", "5", "6"].forEach((digit, i) => {
        result.current.handleChange(i, digit);
      });
    });

    expect(result.current.isComplete).toBe(true);

    act(() => {
      result.current.reset();
    });

    expect(result.current.code).toEqual(["", "", "", "", "", ""]);
    expect(result.current.isComplete).toBe(false);
  });

  it("supports custom length", () => {
    const { result } = renderHook(() => useOtpInput(4));

    expect(result.current.code).toEqual(["", "", "", ""]);
  });

  it("handles backspace on empty field", () => {
    const { result } = renderHook(() => useOtpInput(6));

    // Fill first two digits
    act(() => {
      result.current.handleChange(0, "1");
      result.current.handleChange(1, "2");
    });

    expect(result.current.code).toEqual(["1", "2", "", "", "", ""]);

    // Backspace on empty third field should clear second field
    act(() => {
      result.current.handleKeyDown(2, { key: "Backspace" } as React.KeyboardEvent);
    });

    expect(result.current.code[1]).toBe("");
  });

  it("handles backspace on filled field", () => {
    const { result } = renderHook(() => useOtpInput(6));

    act(() => {
      result.current.handleChange(0, "1");
    });

    expect(result.current.code[0]).toBe("1");

    // Backspace on filled field should clear it
    act(() => {
      result.current.handleKeyDown(0, { key: "Backspace" } as React.KeyboardEvent);
    });

    expect(result.current.code[0]).toBe("");
  });
});
