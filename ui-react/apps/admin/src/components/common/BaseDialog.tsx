import { Ref, ReactNode, RefObject, useCallback, useEffect, useRef } from "react";
import { useFocusTrap } from "@/hooks/useFocusTrap";
import { useBackdropClose } from "@/hooks/useBackdropClose";

type DialogSize = "sm" | "md" | "lg" | "xl" | "full";

const SIZE_CLASSES: Record<DialogSize, string> = {
  sm: "sm:max-w-sm",
  md: "sm:max-w-md",
  lg: "sm:max-w-lg",
  xl: "sm:max-w-xl",
  full: "",
};

export interface BaseDialogProps {
  /** Controls open/close. The component returns null when false. */
  open: boolean;

  /** Called when the user requests closing (ESC or backdrop click).
   *  The parent owns the `open` state and must set it to `false`. */
  onClose: () => void;

  /** Optional predicate. When provided and returning `false`, ESC and
   *  backdrop clicks are blocked. Used by WelcomeWizard's final step.
   *  Returning `true` (or omitting this prop) allows closing. */
  canClose?: () => boolean;

  /** Panel max-width. Defaults to "sm" (max-w-sm = 384px).
   *  Below the sm breakpoint all sizes go full-screen. */
  size?: DialogSize;

  /** Wired to <dialog aria-labelledby>. The consuming component must
   *  render an element with this id. */
  "aria-labelledby"?: string;

  /** Wired to <dialog aria-describedby>. Optional. */
  "aria-describedby"?: string;

  /** Wired to <dialog aria-label>. Use when a visible title element is
   *  absent (e.g. WelcomeWizard). Mutually exclusive with aria-labelledby. */
  "aria-label"?: string;

  /** Additional classes appended to the dialog panel. Use sparingly —
   *  for one-off overrides like max-height. */
  className?: string;

  /** Ref forwarded to the underlying <dialog> element. When provided,
   *  the consuming component shares the same ref used by BaseDialog's
   *  internal hooks. If omitted, BaseDialog manages its own ref. */
  dialogRef?: RefObject<HTMLDialogElement | null>;

  /** Dialog content. BaseDialog imposes no internal structure. */
  children: ReactNode;
}

export default function BaseDialog({
  open,
  onClose,
  canClose: canCloseProp,
  size = "sm",
  "aria-labelledby": ariaLabelledBy,
  "aria-describedby": ariaDescribedBy,
  "aria-label": ariaLabel,
  className,
  dialogRef: externalRef,
  children,
}: BaseDialogProps) {
  const internalRef = useRef<HTMLDialogElement>(null);
  // Use the externally provided ref if given, otherwise fall back to the
  // internal one. Both point to the same <dialog> element so all hooks
  // and the consumer share a single DOM node reference.
  const ref = externalRef ?? internalRef;

  // Wrap the canClose prop in a stable callback so useBackdropClose's
  // internal ref comparison stays consistent across renders.
  const canClose = useCallback(
    () => (canCloseProp ? canCloseProp() : true),
    [canCloseProp],
  );

  const backdropHandlers = useBackdropClose(ref, onClose, canClose);
  useFocusTrap(ref, open);

  // Drive showModal() / close() from the open prop.
  useEffect(() => {
    const dialog = ref.current;
    if (!dialog) return;
    // Guard each branch explicitly:
    // - showModal() throws InvalidStateError if already open (Strict Mode double-mount).
    // - close() should only fire when transitioning to closed.
    if (open && !dialog.open) dialog.showModal();
  }, [open, ref]);

  // Handle ESC via the native cancel event fired by showModal() dialogs.
  //
  // Why not useEscapeKey?
  // useEscapeKey attaches a global document keydown listener that fires
  // regardless of dialog stacking order. The cancel event is scoped to the
  // topmost dialog in the top layer, so stacking works correctly.
  useEffect(() => {
    const dialog = ref.current;
    if (!dialog) return;

    const handleCancel = (e: Event) => {
      // Prevent the browser from closing the dialog itself — React owns the state.
      e.preventDefault();
      // Honor the close guard before delegating to onClose.
      if (!canClose()) return;
      onClose();
    };

    dialog.addEventListener("cancel", handleCancel);
    return () => dialog.removeEventListener("cancel", handleCancel);
  }, [ref, onClose, canClose]);

  if (!open) return null;

  // Panel classes. Notes:
  // - No z-index needed. showModal() places the dialog in the browser's top
  //   layer, which stacks above all other content regardless of z-index.
  //   Multiple showModal() dialogs stack in document order (last = on top).
  // - Below sm breakpoint: full-screen (w-full h-full), no border or radius.
  // - At sm and above: auto-height, border, rounded corners, max-width per size.
  // - "full" size omits the max-width class, staying full-screen at all sizes.
  const isFull = size === "full";
  const panelClasses = [
    "fixed inset-0 m-auto",
    "w-full h-full",
    isFull ? "" : "sm:h-fit",
    "bg-surface",
    isFull ? "" : "sm:border sm:border-border sm:rounded-2xl",
    "shadow-2xl shadow-black/40",
    "animate-slide-up",
    "flex flex-col",
    SIZE_CLASSES[size],
    className ?? "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <dialog
      ref={ref as Ref<HTMLDialogElement>}
      data-custom-backdrop
      aria-labelledby={ariaLabelledBy}
      aria-describedby={ariaDescribedBy}
      aria-label={ariaLabel}
      {...backdropHandlers}
      className={panelClasses}
    >
      {children}
    </dialog>
  );
}
