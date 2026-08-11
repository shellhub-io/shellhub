import type { ReactNode } from "react";

export function MockDrawer({
  open,
  onClose,
  title,
  children,
  footer,
}: {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  footer?: ReactNode;
}) {
  if (!open) return null;
  return (
    <div role="dialog" aria-label={title}>
      <h2>{title}</h2>
      <button type="button" onClick={onClose}>
        Close Drawer
      </button>
      <div>{children}</div>
      {footer && <div>{footer}</div>}
    </div>
  );
}

export function MockConfirmDialog({
  open,
  onClose,
  onConfirm,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  children,
  errorMessage,
}: {
  open: boolean;
  onClose: () => void;
  onConfirm: () => Promise<void> | void;
  title: string;
  description?: ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  children?: ReactNode;
  errorMessage?: string | null;
}) {
  if (!open) return null;
  return (
    <div role="dialog" aria-label={title}>
      <h2>{title}</h2>
      {description && <div>{description}</div>}
      {errorMessage && <div role="alert">{errorMessage}</div>}
      {children}
      <button type="button" onClick={onClose}>
        {cancelLabel}
      </button>
      <button type="button" onClick={() => void onConfirm()}>
        {confirmLabel}
      </button>
    </div>
  );
}

export function MockCopyButton({ text }: { text: string }) {
  return <button type="button" aria-label={`Copy ${text}`} />;
}

export function MockBaseDialog({
  open,
  onClose,
  canClose,
  children,
  "aria-label": ariaLabel,
  "aria-labelledby": ariaLabelledBy,
  "aria-describedby": ariaDescribedBy,
}: {
  open: boolean;
  onClose: () => void;
  canClose?: () => boolean;
  children: ReactNode;
  "aria-label"?: string;
  "aria-labelledby"?: string;
  "aria-describedby"?: string;
}) {
  if (!open) return null;
  return (
    <div
      role="dialog"
      aria-label={ariaLabel}
      aria-labelledby={ariaLabelledBy}
      aria-describedby={ariaDescribedBy}
      data-can-close={String(canClose ? canClose() : true)}
    >
      <button type="button" onClick={onClose}>
        Close Dialog
      </button>
      {children}
    </div>
  );
}
