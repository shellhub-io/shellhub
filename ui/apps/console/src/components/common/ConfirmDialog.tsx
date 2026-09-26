import { ReactNode, useId, useState } from "react";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import {
  Button,
  type ButtonVariant,
  type Palette,
} from "@shellhub/design-system/primitives";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import BaseDialog from "./BaseDialog";
import DialogHeader from "./DialogHeader";
import { ignoreFailure } from "@/utils/failure";

interface ConfirmDialogProps {
  open: boolean;

  onClose: () => void;

  onConfirm: () => Promise<void> | void;

  icon: ReactNode;

  title: string;

  description: ReactNode;

  confirmLabel?: string;

  cancelLabel?: string;

  variant?: Variant;

  confirmDisabled?: boolean;

  children?: ReactNode;

  errorMessage?: string | null;
}

type Variant = "primary" | "danger" | "success" | "warning";

const VARIANT_STYLE: Record<Variant, { button: ButtonVariant; icon: Palette }> =
  {
    primary: { button: "primary", icon: "primary" },
    danger: { button: "destructive", icon: "red" },
    success: { button: "success", icon: "green" },
    warning: { button: "warning", icon: "yellow" },
  };

/**
 * A yes/no confirmation. The variant colours both the icon and the confirm button, so the gravity
 * reads before the text does. onConfirm may be async: the dialog stays open and busy until it
 * settles, so a slow action cannot be triggered twice.
 */
export default function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  icon,
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

  const handleConfirm = () => {
    setConfirming(true);
    void Promise.resolve(onConfirm())
      .catch(ignoreFailure)
      .finally(() => setConfirming(false));
  };

  const style = VARIANT_STYLE[variant];

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="sm"
      aria-labelledby={titleId}
      aria-describedby={descriptionId}
    >
      <DialogHeader
        icon={icon}
        iconColor={style.icon}
        title={title}
        description={description}
        titleId={titleId}
        descriptionId={descriptionId}
        onClose={onClose}
      />
      {(children || errorMessage) && (
        <div className="px-6 pb-6 space-y-4">
          {children}
          {errorMessage && (
            <div
              role="alert"
              className="flex items-start gap-2 bg-accent-red/[0.06] border border-accent-red/20 rounded-lg px-3 py-2.5 text-xs text-accent-red"
            >
              <ExclamationCircleIcon
                className="w-4 h-4 shrink-0 mt-px"
                strokeWidth={2}
              />
              <span>{errorMessage}</span>
            </div>
          )}
        </div>
      )}
      <div className="flex justify-end gap-2 px-6 py-4 border-t border-border">
        <Button variant="ghost" onClick={onClose}>
          {cancelLabel}
        </Button>
        <Button
          variant={style.button}
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
