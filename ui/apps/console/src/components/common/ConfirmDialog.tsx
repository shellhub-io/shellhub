import { ReactNode, useId, useState } from "react";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { Button, type ButtonVariant } from "@shellhub/design-system/primitives";
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

  /** Optional error banner rendered above the footer. Use this to surface
   *  mutation failures inline — the dialog does NOT close on error, so the
   *  user can retry or cancel. */
  errorMessage?: string | null;
}

const VARIANT_BUTTON: Record<
  "primary" | "danger" | "success" | "warning",
  ButtonVariant
> = {
  primary: "primary",
  danger: "destructive",
  success: "success",
  warning: "warning",
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
  errorMessage,
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

  const buttonVariant = VARIANT_BUTTON[variant];

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
        <h2 id={titleId} className="text-base font-semibold text-text-primary">
          {title}
        </h2>
      </div>

      {/* Body */}
      <div className="px-6 pt-2 pb-6">
        {description != null && (
          <div
            id={descriptionId}
            className={`text-sm text-text-muted ${children || errorMessage ? "mb-4" : "mb-6"}`}
          >
            {description}
          </div>
        )}
        {children}
        {errorMessage && (
          <div
            role="alert"
            className={`${children ? "mt-4" : ""} flex items-start gap-2 bg-accent-red/[0.06] border border-accent-red/20 rounded-lg px-3 py-2.5 text-xs text-accent-red`}
          >
            <ExclamationCircleIcon
              className="w-4 h-4 shrink-0 mt-px"
              strokeWidth={2}
            />
            <span>{errorMessage}</span>
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="flex justify-end gap-2 px-6 py-4 border-t border-border">
        <Button variant="ghost" onClick={onClose}>
          {cancelLabel}
        </Button>
        <Button
          variant={buttonVariant}
          disabled={confirmDisabled}
          loading={confirming}
          onClick={() => void handleConfirm()}
        >
          {confirmLabel}
        </Button>
      </div>
    </BaseDialog>
  );
}
