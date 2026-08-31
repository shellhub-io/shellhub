import { useEffect } from "react";
import { useLocation } from "react-router-dom";

/**
 * Scrolls to the top on every route change and renders nothing. Without it the router keeps the
 * previous scroll offset, so a link taken from far down a long page opens the next one midway.
 * The scroll is instant on purpose: a smooth one animates through the page just left behind.
 */
export function ScrollToTop() {
  const { pathname } = useLocation();

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: "instant" });
  }, [pathname]);

  return null;
}
