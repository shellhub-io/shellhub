import { useEffect, useRef, useState } from "react";

/**
 * Whether a sticky element has left its place and is pinned by the page's scroll. Put sentinel on
 * a zero-height element right above the sticky one: stuck holds while the sentinel is scrolled out
 * of the page's scroll container, AppLayout's main.
 */
export function useStuck() {
  const sentinel = useRef<HTMLDivElement>(null);
  const [stuck, setStuck] = useState(false);

  useEffect(() => {
    const el = sentinel.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      ([entry]) => setStuck(!entry.isIntersecting),
      { root: el.closest("main") },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  return { sentinel, stuck };
}
