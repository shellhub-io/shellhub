import { useEffect, useRef } from "react";
import { cn } from "../primitives/cn";

function useReveal() {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      ([e]) => { if (e.isIntersecting) { el.classList.add("landing-visible"); obs.unobserve(el); } },
      { threshold: 0.08, rootMargin: "0px 0px -60px 0px" },
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, []);
  return ref;
}

/**
 * Fades its children in the first time they scroll into view, then stops observing, so the
 * animation plays once per mount rather than on every pass. delay staggers a group of
 * siblings into a sequence.
 */
export function Reveal({ children, className = "", delay = 0 }: { children: React.ReactNode; className?: string; delay?: number }) {
  const ref = useReveal();
  return (
    <div ref={ref} className={cn("landing-reveal", className)} style={delay ? { transitionDelay: `${delay}s` } : undefined}>
      {children}
    </div>
  );
}
