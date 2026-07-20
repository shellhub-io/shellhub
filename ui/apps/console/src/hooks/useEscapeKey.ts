import { RefObject, useEffect } from "react";

export function useEscapeKey(
  handler: () => void,
  enabled = true,
  containerRef?: RefObject<HTMLElement | null>,
): void {
  useEffect(() => {
    if (!enabled) return;
    const target = containerRef?.current ?? document;

    const listener = (e: Event) => {
      if ("key" in e && e.key === "Escape") handler();
    };

    target.addEventListener("keydown", listener);
    return () => target.removeEventListener("keydown", listener);
  }, [handler, enabled, containerRef]);
}
