import { describe, it, expect } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { useScrollEdges } from "../useScrollEdges";

function scroller(scrollTop: number, clientHeight = 100, scrollHeight = 300) {
  const el = document.createElement("div");
  Object.defineProperties(el, {
    scrollTop: { value: scrollTop, writable: true },
    clientHeight: { value: clientHeight },
    scrollHeight: { value: scrollHeight },
  });
  return el;
}

function attach(el: HTMLDivElement) {
  const { result } = renderHook(() => useScrollEdges<HTMLDivElement>());
  act(() => {
    result.current.ref(el);
  });
  return result;
}

describe("useScrollEdges", () => {
  it.each([
    ["at the top", 0, false, true],
    ["in the middle", 100, true, true],
    ["at the bottom", 200, true, false],
    ["within a pixel of the bottom", 199.5, true, false],
  ])(
    "reports the hidden edges of content scrolled %s",
    (_, scrollTop, moreAbove, moreBelow) => {
      const result = attach(scroller(scrollTop));

      expect(result.current.moreAbove).toBe(moreAbove);
      expect(result.current.moreBelow).toBe(moreBelow);
    },
  );

  it("reports no hidden edges when the content fits", () => {
    const result = attach(scroller(0, 100, 100));

    expect(result.current.moreAbove).toBe(false);
    expect(result.current.moreBelow).toBe(false);
  });

  it("follows the element as it scrolls", () => {
    const el = scroller(0);
    const result = attach(el);

    act(() => {
      el.scrollTop = 200;
      el.dispatchEvent(new Event("scroll"));
    });

    expect(result.current.moreAbove).toBe(true);
    expect(result.current.moreBelow).toBe(false);
  });
});
