import { useCallback, useState } from "react";

/**
 * Whether a scrolling element has content hidden past its top and its bottom edge, for drawing a
 * hint there. Attach the returned ref to the element; it follows scrolling and the size of the
 * element and of each of its children, which grow with whatever loads or expands inside them, and
 * stops when the element unmounts.
 */
export function useScrollEdges<T extends HTMLElement>() {
  const [edges, setEdges] = useState({ moreAbove: false, moreBelow: false });

  const ref = useCallback((el: T | null) => {
    if (!el) return undefined;
    const update = () => {
      const moreAbove = el.scrollTop > 0;
      const moreBelow = el.scrollTop + el.clientHeight < el.scrollHeight - 1;
      setEdges((prev) =>
        prev.moreAbove === moreAbove && prev.moreBelow === moreBelow
          ? prev
          : { moreAbove, moreBelow },
      );
    };
    const sizes = new ResizeObserver(update);
    const watchSizes = () => {
      sizes.disconnect();
      sizes.observe(el);
      for (const child of el.children) sizes.observe(child);
      update();
    };
    const children = new MutationObserver(watchSizes);
    children.observe(el, { childList: true });
    el.addEventListener("scroll", update, { passive: true });
    watchSizes();
    return () => {
      el.removeEventListener("scroll", update);
      children.disconnect();
      sizes.disconnect();
    };
  }, []);

  return { ref, ...edges };
}
