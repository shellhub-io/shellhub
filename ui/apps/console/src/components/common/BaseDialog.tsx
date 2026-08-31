import { ReactNode, RefObject, useCallback, useEffect, useRef } from "react";
import { cn } from "@shellhub/design-system/cn";
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

  /** When false, no control is focused on open (the dialog itself takes focus)
   *  instead of the first focusable child. Defaults to true. */
  focusOnOpen?: boolean;

  /** Panel max-width. Defaults to "sm" (max-w-sm = 384px).
   *  Below the sm breakpoint all sizes go full-screen. */
  size?: DialogSize;

  "aria-labelledby"?: string;

  "aria-describedby"?: string;

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
  focusOnOpen = true,
  size = "sm",
  "aria-labelledby": ariaLabelledBy,
  "aria-describedby": ariaDescribedBy,
  "aria-label": ariaLabel,
  className,
  dialogRef: externalRef,
  children,
}: BaseDialogProps) {
  const internalRef = useRef<HTMLDialogElement>(null);
  const ref = externalRef ?? internalRef;

  const canClose = useCallback(
    () => (canCloseProp ? canCloseProp() : true),
    [canCloseProp],
  );

  const backdropHandlers = useBackdropClose(ref, onClose, canClose);
  useFocusTrap(ref, open, focusOnOpen);

  useEffect(() => {
    const dialog = ref.current;
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
  }, [open, ref]);

  useEffect(() => {
    const dialog = ref.current;
    if (!dialog) return;

    const handleCancel = (e: Event) => {
      e.preventDefault();
      if (!canClose()) return;
      onClose();
    };

    dialog.addEventListener("cancel", handleCancel);
    return () => dialog.removeEventListener("cancel", handleCancel);
  }, [ref, onClose, canClose]);

  if (!open) return null;

  const isFull = size === "full";
  const panelClasses = cn(
    "fixed inset-0 m-auto",
    "w-full h-full",
    !isFull && "sm:h-fit",
    "bg-surface",
    !isFull && "sm:border sm:border-border sm:rounded-2xl",
    "shadow-2xl shadow-black/40",
    "animate-slide-up",
    "flex flex-col",
    SIZE_CLASSES[size],
    className,
  );

  return (
    <dialog
      ref={ref}
      tabIndex={-1}
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
