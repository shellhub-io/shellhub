import { useEffect, useRef } from "react";

/**
 * Calls onReach whenever sentinel comes within a screen of the bottom of the page's scroll
 * container, AppLayout's main, so a list loads its next page before the reader gets there. Put
 * sentinel on an element below the list; while enabled is false nothing is observed.
 */
export function useLoadOnReach(onReach: () => void, enabled: boolean) {
  const sentinel = useRef<HTMLDivElement>(null);
  const callback = useRef(onReach);

  useEffect(() => {
    callback.current = onReach;
  });

  useEffect(() => {
    const el = sentinel.current;
    if (!el || !enabled) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) callback.current();
      },
      { root: el.closest("main"), rootMargin: "0px 0px 100% 0px" },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [enabled]);

  return sentinel;
}
