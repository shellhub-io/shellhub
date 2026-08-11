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
