import { describe, it, expect, vi, beforeEach } from "vitest";
import { act, renderHook } from "@testing-library/react";
import type { PointerEvent } from "react";
import { useTabReorder } from "../useTabReorder";

const ids = ["a", "b", "c"];

function buildStrip() {
  const strip = document.createElement("div");
  const tabs = ids.map((id, i) => {
    const el = document.createElement("div");
    el.dataset.tabId = id;
    el.getBoundingClientRect = () => ({ left: i * 102, width: 100 }) as DOMRect;
    el.setPointerCapture = () => {};
    strip.appendChild(el);
    return el;
  });
  return Object.fromEntries(ids.map((id, i) => [id, tabs[i]]));
}

function pointer(target: HTMLElement, clientX: number) {
  return {
    button: 0,
    pointerId: 1,
    clientX,
    currentTarget: target,
  } as unknown as PointerEvent<HTMLElement>;
}

let tabs: Record<string, HTMLElement>;
const move = vi.fn();

beforeEach(() => {
  tabs = buildStrip();
  move.mockClear();
});

function drag(
  result: { current: ReturnType<typeof useTabReorder> },
  id: string,
  to: number,
) {
  const tab = () => result.current.tab(ids, move, id);
  act(() => tab().onPointerDown(pointer(tabs[id], 0)));
  act(() => tab().onPointerMove(pointer(tabs[id], 10)));
  act(() => tab().onPointerMove(pointer(tabs[id], to)));
}

describe("useTabReorder", () => {
  it("leaves a press that barely moves as a click", () => {
    const { result } = renderHook(() => useTabReorder());
    const tab = () => result.current.tab(ids, move, "a");

    act(() => tab().onPointerDown(pointer(tabs.a, 0)));
    act(() => tab().onPointerMove(pointer(tabs.a, 4)));
    act(() => tab().onPointerUp());

    expect(tab().dragging).toBe(false);
    expect(move).not.toHaveBeenCalled();
  });

  it("previews the tab in its new place while it is dragged", () => {
    const { result } = renderHook(() => useTabReorder());

    drag(result, "a", 170);

    expect(result.current.orderOf(ids)).toEqual(["b", "c", "a"]);
  });

  it("drops the tab where it was dragged to", () => {
    const { result } = renderHook(() => useTabReorder());

    drag(result, "a", 170);
    act(() => result.current.tab(ids, move, "a").onPointerUp());

    expect(move).toHaveBeenCalledWith("a", 2);
  });

  it("puts the tab back when the drag is cancelled", () => {
    const { result } = renderHook(() => useTabReorder());

    drag(result, "a", 170);
    act(() => result.current.cancel());

    expect(move).not.toHaveBeenCalled();
    expect(result.current.tab(ids, move, "a").offset).toBe(0);
    expect(result.current.orderOf(ids)).toEqual(ids);
  });

  it("abandons the drag in progress when another tab is pressed", () => {
    const { result } = renderHook(() => useTabReorder());

    drag(result, "a", 170);
    drag(result, "c", -170);
    act(() => result.current.tab(ids, move, "c").onPointerUp());

    expect(move).toHaveBeenCalledOnce();
    expect(move).toHaveBeenCalledWith("c", 0);
  });

  it.each([
    ["b", 1, 2],
    ["b", -1, 0],
  ] as const)(
    "steps %s by %i to index %i from the keyboard",
    (id, step, index) => {
      const { result } = renderHook(() => useTabReorder());

      act(() => result.current.tab(ids, move, id).onMove(step));

      expect(move).toHaveBeenCalledWith(id, index);
    },
  );
});
