import { ReactNode, RefObject, useEffect, useEffectEvent, useRef } from "react";
import { cn } from "@shellhub/design-system/cn";
import { useFocusTrap } from "@/hooks/useFocusTrap";
import { useBackdropClose } from "@/hooks/useBackdropClose";
import { desktopWindow } from "@/utils/desktopWindow";

/**
 * The width a dialog takes, from a fixed scale: sm for confirmations, md for short forms, lg for
 * long forms and pickers, xl for wizards and wide lists. full fills the window, and like a phone
 * screen it takes focus itself rather than its first field.
 */
export type DialogSize = "sm" | "md" | "lg" | "xl" | "full";

const SIZE_CLASSES: Record<DialogSize, string> = {
  sm: "max-w-[400px]",
  md: "max-w-[560px]",
  lg: "max-w-[720px]",
  xl: "max-w-[960px]",
  full: "",
};

const FULLSCREEN_CLASSES =
  "mt-0 w-full h-dvh max-h-none max-w-none rounded-none border-0";

const PHONE_FULLSCREEN_CLASSES =
  "max-sm:mt-0 max-sm:w-full max-sm:h-dvh max-sm:max-h-none max-sm:max-w-none max-sm:rounded-none max-sm:border-0 max-sm:pt-[env(safe-area-inset-top)] max-sm:pb-[env(safe-area-inset-bottom)]";

const PHONE = "(max-width: 639px)";

function fillsPhoneScreen() {
  return !desktopWindow() && (window.matchMedia?.(PHONE).matches ?? false);
}

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
 * will escape to the page behind it. The panel never grows past the window: taller content
 * scrolls the panel as a whole. On a phone it fills the screen and takes focus itself rather
 * than its first field, so the keyboard does not cover it; the desktop app keeps it floating,
 * since the app draws its own window controls and a full-screen panel would bury them.
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

  const canClose = () => (canCloseProp ? canCloseProp() : true);

  const backdropHandlers = useBackdropClose(ref, onClose, canClose);
  const isFull = size === "full";
  useFocusTrap(
    ref,
    open,
    focusOnOpen && !isFull && !(open && fillsPhoneScreen()),
  );

  useEffect(() => {
    const dialog = ref.current;
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
  }, [open, ref]);

  const requestClose = useEffectEvent(() => {
    if (canClose()) onClose();
  });

  useEffect(() => {
    const dialog = ref.current;
    if (!dialog) return;

    const handleCancel = (e: Event) => {
      e.preventDefault();
      if (e.cancelable) requestClose();
    };

    const handleClose = () => {
      dialog.showModal();
      requestClose();
    };

    dialog.addEventListener("cancel", handleCancel);
    dialog.addEventListener("close", handleClose);
    return () => {
      dialog.removeEventListener("cancel", handleCancel);
      dialog.removeEventListener("close", handleClose);
    };
  }, [open, ref]);

  if (!open) return null;

  const panelClasses = cn(
    "fixed inset-x-0 top-0 mx-auto mt-[clamp(1.5rem,12vh,8rem)]",
    "w-[calc(100vw-2rem)] h-fit max-h-[calc(100dvh-clamp(1.5rem,12vh,8rem)-2rem)]",
    "overflow-y-auto overscroll-contain",
    "bg-surface border border-border rounded-2xl",
    "shadow-2xl shadow-black/40",
    "animate-slide-up",
    "flex flex-col",
    SIZE_CLASSES[size],
    isFull ? FULLSCREEN_CLASSES : !desktopWindow() && PHONE_FULLSCREEN_CLASSES,
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
