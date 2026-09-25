import { useState, useRef, useEffect, useSyncExternalStore } from "react";

const PINNED_KEY = "sidebarPinned";

function watchWidth(query: string) {
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
const wideWidth = watchWidth("(min-width: 1280px)");

function readPinned(): boolean | null {
  try {
    const stored = localStorage.getItem(PINNED_KEY);
    return stored === null ? null : stored === "true";
  } catch {
    return null;
  }
}

function writePinned(pinned: boolean) {
  try {
    localStorage.setItem(PINNED_KEY, String(pinned));
  } catch {
    return;
  }
}

/**
 * Drives the sidebar across window sizes. Wide windows keep it open, narrower ones fold it to a
 * rail that opens over the content on hover or when the keyboard reaches it (a click leaves focus
 * behind, which must not hold it open once the pointer leaves), and below desktop width it
 * becomes a drawer. Pinning
 * or unpinning it is remembered and outranks the width, since it is the user's own choice.
 */
export function useSidebarLayout() {
  const [expanded, setExpanded] = useState(false);
  const [pinnedChoice, setPinnedChoice] = useState(readPinned);
  const [drawerOpen, setDrawerOpen] = useState(false);

  const isDesktop = useSyncExternalStore(
    desktopWidth.subscribe,
    desktopWidth.matches,
    () => true,
  );
  const isWide = useSyncExternalStore(
    wideWidth.subscribe,
    wideWidth.matches,
    () => true,
  );

  const hoverTimer = useRef<ReturnType<typeof setTimeout>>(undefined);

  const pinned = pinnedChoice ?? isWide;
  const isOpen = expanded || pinned;

  const openDrawer = () => setDrawerOpen(true);
  const closeDrawer = () => setDrawerOpen(false);
  const toggleDrawer = () => setDrawerOpen((prev) => !prev);

  useEffect(() => () => clearTimeout(hoverTimer.current), []);

  const handleKeyboardFocus = (e: React.FocusEvent) => {
    if (e.target instanceof Element && e.target.matches(":focus-visible")) {
      handleExpand();
    }
  };

  const handleExpand = () => {
    clearTimeout(hoverTimer.current);
    hoverTimer.current = setTimeout(() => setExpanded(true), 75);
  };

  const handleCollapse = () => {
    clearTimeout(hoverTimer.current);
    hoverTimer.current = setTimeout(() => setExpanded(false), 150);
  };

  const handleToggle = () => {
    clearTimeout(hoverTimer.current);
    setExpanded(false);
    setPinnedChoice(!pinned);
    writePinned(!pinned);
  };

  const handleDrawerKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") closeDrawer();
  };

  return {
    expanded,
    pinned,
    isOpen,
    isDesktop,
    drawerOpen,
    handlers: {
      onMouseEnter: handleExpand,
      onMouseLeave: handleCollapse,
      onFocus: handleKeyboardFocus,
      onBlur: handleCollapse,
      onToggle: handleToggle,
      openDrawer,
      closeDrawer,
      toggleDrawer,
      onDrawerKeyDown: handleDrawerKeyDown,
    },
  };
}
