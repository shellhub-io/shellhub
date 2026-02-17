import { ReactNode, useState } from "react";

interface ConfirmDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => Promise<void> | void;
  title: string;
  description: ReactNode;
  confirmLabel?: string;
  variant?: "danger" | "primary";
  children?: ReactNode;
  confirmDisabled?: boolean;
}

export default function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  title,
  description,
  confirmLabel = "Confirm",
  variant = "danger",
  children,
  confirmDisabled,
}: ConfirmDialogProps) {
  const [confirming, setConfirming] = useState(false);

  if (!open) return null;

  const handleConfirm = async () => {
    setConfirming(true);
    try {
      await onConfirm();
    } finally {
      setConfirming(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />
      <div className="relative bg-surface border border-border rounded-2xl w-full max-w-sm mx-4 p-6 shadow-2xl animate-slide-up">
        <h2 className="text-base font-semibold text-text-primary mb-2">
          {title}
        </h2>
        <div
          className={`text-sm text-text-muted ${children ? "mb-4" : "mb-6"}`}
        >
          {description}
        </div>
        {children}
        <div className="flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            disabled={confirming || confirmDisabled}
            className={`px-5 py-2.5 ${
              variant === "danger"
                ? "bg-accent-red/90 hover:bg-accent-red"
                : "bg-primary hover:bg-primary-600"
            } text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2`}
          >
            {confirming && (
              <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            )}
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
