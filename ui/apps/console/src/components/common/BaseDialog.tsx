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

/**
 * Props of BaseDialog. canClose is what a step mid-flow uses to refuse dismissal — a dialog that
 * has already charged a card must not be closeable by Escape.
 */
export interface BaseDialogProps {
  open: boolean;

  onClose: () => void;

  canClose?: () => boolean;

  focusOnOpen?: boolean;

  size?: DialogSize;

  "aria-labelledby"?: string;

  "aria-describedby"?: string;

  "aria-label"?: string;

  className?: string;

  dialogRef?: RefObject<HTMLDialogElement | null>;

  children: ReactNode;
}

/**
 * The dialog every other dialog is built on: the overlay, the focus trap, Escape, and the
 * scroll lock. Anything that needs a modal should use this rather than a fixed div, or focus
 * will escape to the page behind it.
 */
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
