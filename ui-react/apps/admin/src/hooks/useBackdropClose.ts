import { useRef, type RefObject } from "react";

/**
 * Returns `onMouseDown` and `onClick` handlers to close a native `<dialog>`
 * when the user clicks the backdrop.
 *
 * ## Why not `useClickOutside`?
 * `useClickOutside` checks whether the click target is outside the ref element
 * in the DOM. With `showModal()`, the backdrop is the `<dialog>` element itself
 * (a `::backdrop` pseudo-element), so the target is never "outside" — that hook
 * won't fire.
 *
 * ## Why track `mousedown`?
 * Checking only `onClick`'s `e.target` has two bugs:
 * - Dragging from inside the dialog to the backdrop fires `click` on the dialog,
 *   incorrectly closing it.
 * - Starting a text selection inside also triggers the same false positive.
 *
 * Tracking the `mousedown` origin fixes both: the dialog only closes when the
 * click both *started* and *ended* on the `<dialog>` element (the backdrop area).
 *
 * ## Usage
 * ```tsx
 * const backdropHandlers = useBackdropClose(dialogRef, onClose);
 * <dialog ref={dialogRef} {...backdropHandlers}>
 * ```
 *
 * Optionally guard the close with a condition (e.g. block on the final step):
 * ```tsx
 * const backdropHandlers = useBackdropClose(dialogRef, onClose, () => step < TOTAL_STEPS);
 * ```
 */
export function useBackdropClose(
  dialogRef: RefObject<HTMLDialogElement | null>,
  onClose: () => void,
  enabled: () => boolean = () => true,
): {
  onMouseDown: React.MouseEventHandler<HTMLDialogElement>;
  onClick: React.MouseEventHandler<HTMLDialogElement>;
} {
  const mouseDownTarget = useRef<EventTarget | null>(null);

  return {
    onMouseDown(e) {
      mouseDownTarget.current = e.target;
    },
    onClick(e) {
      if (
        enabled()
        && e.target === dialogRef.current
        && mouseDownTarget.current === dialogRef.current
      ) {
        onClose();
      }
    },
  };
}
