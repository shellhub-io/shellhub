import { useSyncExternalStore } from "react";

/**
 * A media query as a store useSyncExternalStore can read: the current match, and a subscription
 * to its changes. Outside a browser it reads as a match.
 */
export function watchWidth(query: string) {
  const mql =
    typeof window !== "undefined" ? window.matchMedia(query) : undefined;
  return {
    subscribe: (callback: () => void) => {
      mql?.addEventListener("change", callback);
      return () => mql?.removeEventListener("change", callback);
    },
    matches: () => mql?.matches ?? true,
  };
}

const desktopWidth = watchWidth("(min-width: 1024px)");

/**
 * Whether the window is at least desktop width (Tailwind's lg), following it as it is resized.
 */
export function useIsDesktop() {
  return useSyncExternalStore(
    desktopWidth.subscribe,
    desktopWidth.matches,
    () => true,
  );
}
