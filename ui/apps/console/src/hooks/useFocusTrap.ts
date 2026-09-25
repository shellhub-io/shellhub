import { RefObject, useEffect } from "react";

const FOCUSABLE =
  ':is(a[href],button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),[tabindex]):not([tabindex="-1"])';

/**
 * Traps keyboard focus within `containerRef` while `active` is true.
 * Restores focus to the previously focused element when deactivated.
 *
 * On activation focus goes to the first focusable child that is not marked
 * `data-dismiss`, so a close button in a title bar does not take focus ahead of
 * the first field; it is the fallback when nothing else can take focus.
 *
 * With `autoFocus` false the container itself takes focus instead of its first
 * focusable child, so no control shows a focus ring on open (the browser would
 * otherwise land on the close button). Tab still works, moving into the content
 * from the container.
 */
export function useFocusTrap(
  containerRef: RefObject<HTMLElement | null>,
  active: boolean,
  autoFocus: boolean = true,
): void {
  useEffect(() => {
    if (!active || !containerRef.current) return;

    const container = containerRef.current;
    const previousFocus = document.activeElement as HTMLElement | null;

    const raf = requestAnimationFrame(() => {
      if (!autoFocus) {
        container.focus();
        return;
      }
      const first =
        container.querySelector<HTMLElement>(
          `${FOCUSABLE}:not([data-dismiss])`,
        ) ?? container.querySelector<HTMLElement>(FOCUSABLE);
      first?.focus();
    });

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key !== "Tab") return;

      const focusable = Array.from(
        container.querySelectorAll<HTMLElement>(FOCUSABLE),
      );
      if (focusable.length === 0) {
        e.preventDefault();
        return;
      }

      const first = focusable[0];
      const last = focusable[focusable.length - 1];

      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    };

    container.addEventListener("keydown", handleKeyDown);

    return () => {
      cancelAnimationFrame(raf);
      container.removeEventListener("keydown", handleKeyDown);
      previousFocus?.focus();
    };
  }, [active, containerRef, autoFocus]);
}
