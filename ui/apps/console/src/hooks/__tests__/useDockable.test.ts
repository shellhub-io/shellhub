import { afterEach, describe, it, expect, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import type { PointerEvent } from "react";
import { useDockable } from "../useDockable";

const PREVIEW_HEIGHT = 160;

function rect(top: number, left: number, width: number) {
  return { top, left, width, height: PREVIEW_HEIGHT } as DOMRect;
}

function pointer(clientY: number) {
  const target = document.createElement("div");
  target.getBoundingClientRect = () => rect(100, 16, 800);
  target.setPointerCapture = vi.fn();
  Object.defineProperty(target, "offsetHeight", { value: PREVIEW_HEIGHT });
  return {
    clientY,
    pointerId: 1,
    currentTarget: target,
  } as unknown as PointerEvent<HTMLElement>;
}

function resizeWindow(innerHeight: number) {
  vi.stubGlobal("innerHeight", innerHeight);
  window.dispatchEvent(new Event("resize"));
}

function floatAt(top: number) {
  const { result } = renderHook(() => useDockable(true));
  act(() => {
    result.current.handlers.onPointerDown(pointer(100));
  });
  act(() => {
    result.current.handlers.onPointerMove(pointer(top));
  });
  act(() => {
    result.current.handlers.onPointerUp();
  });
  const slot = document.createElement("div");
  slot.getBoundingClientRect = () => rect(0, 16, 800);
  result.current.slot.current = slot;
  return { result, slot };
}

describe("useDockable", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("fits a floating element to its slot when the window resizes", () => {
    const { result, slot } = floatAt(300);

    slot.getBoundingClientRect = () => rect(0, 12, 400);
    act(() => resizeWindow(768));

    expect(result.current.frame).toEqual({ top: 300, left: 12, width: 400 });
  });

  it("keeps a floating element inside a window made shorter", () => {
    const { result } = floatAt(500);

    act(() => resizeWindow(400));

    expect(result.current.frame.top).toBe(232);
  });
});
