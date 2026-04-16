import { ReactNode, useId, useState } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import BaseDialog from "./BaseDialog";

interface ConfirmDialogProps {
  /** Controls open/close state. */
  open: boolean;

  /** Called when the dialog should close (ESC, backdrop, or Cancel button).
   *  The parent owns the `open` state. */
  onClose: () => void;

  /** Called when the user clicks the confirm button. If it returns a Promise,
   *  the button shows a spinner until the promise settles. The dialog does NOT
   *  auto-close on success — the caller decides. */
  onConfirm: () => Promise<void> | void;

  /** Dialog title, rendered as an <h2>. */
  title: string;

  /** Description below the title. Accepts ReactNode for inline formatting. */
  description: ReactNode;

  /** Label for the confirm button. Default: "Confirm". */
  confirmLabel?: string;

  /** Label for the cancel button. Default: "Cancel". */
  cancelLabel?: string;

  /** Controls confirm button color scheme.
   *  - "danger"   → bg-accent-red   (default)
   *  - "primary"  → bg-primary
   *  - "success"  → bg-accent-green
   *  - "warning"  → bg-accent-yellow (dark text) */
  variant?: "primary" | "danger" | "success" | "warning";

  /** Disables the confirm button externally (e.g. pending form validation). */
  confirmDisabled?: boolean;

  /** Optional content rendered between the description and the button row.
   *  Used by dialogs that embed extra form fields or warning messages. */
  children?: ReactNode;
}

// Text color is included per-variant so that `warning`'s dark text is not
// overridden by a standalone `text-white` on the button element.
const VARIANT_CLASSES: Record<"primary" | "danger" | "success" | "warning", string> = {
  danger: "bg-accent-red/90 hover:bg-accent-red text-white",
  primary: "bg-primary hover:bg-primary-600 text-white",
  success: "bg-accent-green/90 hover:bg-accent-green text-white",
  warning: "bg-accent-yellow/90 hover:bg-accent-yellow text-background",
};

export default function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  variant = "danger",
  confirmDisabled,
  children,
}: ConfirmDialogProps) {
  const [confirming, setConfirming] = useState(false);
  // useId produces a stable, unique id per component instance, satisfying
  // the aria-labelledby / aria-describedby contracts without manual id management.
  const autoId = useId();
  const titleId = `confirm-dialog-title-${autoId}`;
  const descriptionId = `confirm-dialog-description-${autoId}`;

  useResetOnOpen(open, () => {
    setConfirming(false);
  });

  const handleConfirm = async () => {
    setConfirming(true);
    try {
      await onConfirm();
    } catch {
      // Errors are not surfaced here. Consumers manage their own error state
      // and should pass an already-caught onConfirm handler (e.g. KeyDeleteDialog).
    } finally {
      setConfirming(false);
    }
  };

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="sm"
      aria-labelledby={titleId}
      aria-describedby={description != null ? descriptionId : undefined}
    >
      {/* Header */}
      <div className="p-6 pb-0">
        <h2
          id={titleId}
          className="text-base font-semibold text-text-primary"
        >
          {title}
        </h2>
      </div>

      {/* Body */}
      <div className="px-6 pt-2 pb-6">
        {description != null && (
          <div id={descriptionId} className={`text-sm text-text-muted ${children ? "mb-4" : "mb-6"}`}>
            {description}
          </div>
        )}
        {children}
      </div>

      {/* Footer */}
      <div className="flex justify-end gap-2 px-6 py-4 border-t border-border">
        <button
          onClick={onClose}
          className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
        >
          {cancelLabel}
        </button>
        <button
          onClick={() => void handleConfirm()}
          disabled={confirming || confirmDisabled}
          className={`px-5 py-2.5 ${VARIANT_CLASSES[variant]} rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2`}
        >
          {confirming && (
            <span
              data-testid="confirm-spinner"
              aria-hidden="true"
              className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
            />
          )}
          {confirmLabel}
        </button>
      </div>
    </BaseDialog>
  );
}
