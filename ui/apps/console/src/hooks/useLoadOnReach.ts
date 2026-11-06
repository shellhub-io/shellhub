import { useEffect, useRef } from "react";

/**
 * Calls onReach whenever sentinel comes within a screen of the bottom of its scroll container, so
 * a list loads its next page before the reader gets there. The container is the nearest ancestor
 * matching root, AppLayout's main by default. Put sentinel on an element below the list; while
 * enabled is false nothing is observed.
 */
export function useLoadOnReach(
  onReach: () => void,
  enabled: boolean,
  root = "main",
) {
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
      { root: el.closest(root), rootMargin: "0px 0px 100% 0px" },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [enabled, root]);

  return sentinel;
}
