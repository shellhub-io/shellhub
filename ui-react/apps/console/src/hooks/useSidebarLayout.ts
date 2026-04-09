import {
  useState,
  useCallback,
  useRef,
  useEffect,
  useSyncExternalStore,
} from "react";
import { useLocation } from "react-router-dom";

const lgQuery = "(min-width: 1024px)";

const lgMql =
  typeof window !== "undefined" ? window.matchMedia(lgQuery) : undefined;

function subscribeToMediaQuery(callback: () => void) {
  lgMql?.addEventListener("change", callback);
  return () => lgMql?.removeEventListener("change", callback);
}

function getIsDesktop() {
  return lgMql?.matches ?? true;
}

function getIsDesktopServer() {
  return true;
}

export function useSidebarLayout() {
  const [expanded, setExpanded] = useState(false);
  const [pinned, setPinned] = useState(false);
  const [drawerPathname, setDrawerPathname] = useState<string | null>(null);
  const { pathname } = useLocation();

  const isDesktop = useSyncExternalStore(
    subscribeToMediaQuery,
    getIsDesktop,
    getIsDesktopServer,
  );

  const hoverTimer = useRef<ReturnType<typeof setTimeout>>(undefined);
  const [prevIsDesktop, setPrevIsDesktop] = useState(isDesktop);

  // Reset transient state when crossing the desktop/mobile breakpoint
  if (prevIsDesktop !== isDesktop) {
    setPrevIsDesktop(isDesktop);
    setExpanded(false);
    setDrawerPathname(null);
  }

  const isOpen = expanded || pinned;
  const drawerOpen = drawerPathname === pathname;

  // Clean up hover timer on unmount
  useEffect(() => {
    return () => clearTimeout(hoverTimer.current);
  }, []);

  const handleExpand = useCallback(() => {
    clearTimeout(hoverTimer.current);
    hoverTimer.current = setTimeout(() => setExpanded(true), 75);
  }, []);

  const handleCollapse = useCallback(() => {
    clearTimeout(hoverTimer.current);
    hoverTimer.current = setTimeout(() => setExpanded(false), 150);
  }, []);

  const handleToggle = useCallback(() => {
    setPinned((prev) => !prev);
  }, []);

  const openDrawer = useCallback(() => setDrawerPathname(pathname), [pathname]);
  const closeDrawer = useCallback(() => setDrawerPathname(null), []);

  const handleDrawerKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Escape") closeDrawer();
    },
    [closeDrawer],
  );

  return {
    expanded,
    pinned,
    isOpen,
    isDesktop,
    drawerOpen,
    handlers: {
      onMouseEnter: handleExpand,
      onMouseLeave: handleCollapse,
      onFocus: handleExpand,
      onBlur: handleCollapse,
      onToggle: handleToggle,
      openDrawer,
      closeDrawer,
      onDrawerKeyDown: handleDrawerKeyDown,
    },
  };
}
