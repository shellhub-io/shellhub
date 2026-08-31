import { ReactNode, useId, useState } from "react";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { Button, type ButtonVariant } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import BaseDialog from "./BaseDialog";

interface ConfirmDialogProps {
  open: boolean;

  onClose: () => void;

  onConfirm: () => Promise<void> | void;

  title: string;

  description: ReactNode;

  confirmLabel?: string;

  cancelLabel?: string;

  variant?: "primary" | "danger" | "success" | "warning";

  confirmDisabled?: boolean;

  children?: ReactNode;

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
    // eslint-disable-next-line no-empty -- the consumer of the dialog owns its own error state and surfaces it there
    } catch {
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
            className={cn("text-sm text-text-muted", children || errorMessage ? "mb-4" : "mb-6")}
          >
            {description}
          </div>
        )}
        {children}
        {errorMessage && (
          <div
            role="alert"
            className={cn("flex items-start gap-2 bg-accent-red/[0.06] border border-accent-red/20 rounded-lg px-3 py-2.5 text-xs text-accent-red", children && "mt-4")}
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
