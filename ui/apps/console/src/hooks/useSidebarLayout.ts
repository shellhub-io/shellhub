import { useState, useRef, useEffect, useSyncExternalStore } from "react";
import { useIsDesktop, watchWidth } from "@/hooks/useIsDesktop";
import { useSidebarStore } from "@/stores/sidebarStore";

const wideWidth = watchWidth("(min-width: 1280px)");

/**
 * Drives the sidebar across window sizes. Wide windows keep it open, narrower ones fold it to a
 * rail that opens over the content on hover or when the keyboard reaches it (a click leaves focus
 * behind, which must not hold it open once the pointer leaves), and below desktop width it
 * becomes a drawer. Pinning or unpinning it is remembered in useSidebarStore and outranks the
 * width, since it is the user's own choice.
 */
export function useSidebarLayout() {
  const [expanded, setExpanded] = useState(false);
  const pin = useSidebarStore((st) => st.pin);
  const setPin = useSidebarStore((st) => st.setPin);
  const [drawerOpen, setDrawerOpen] = useState(false);

  const isDesktop = useIsDesktop();
  const isWide = useSyncExternalStore(
    wideWidth.subscribe,
    wideWidth.matches,
    () => true,
  );

  const hoverTimer = useRef<ReturnType<typeof setTimeout>>(undefined);

  const pinned = pin === "auto" ? isWide : pin === "pinned";
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
    setPin(pinned ? "rail" : "pinned");
  };

  const handleDrawerKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") closeDrawer();
  };

  return {
    expanded,
    pinned,
    isOpen,
    isDesktop,
    isWide,
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
