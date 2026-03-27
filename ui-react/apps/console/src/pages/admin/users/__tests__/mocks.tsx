import React from "react";

export const MockDrawer = ({
  open,
  onClose,
  title,
  children,
  footer,
}: {
  open: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
}) => {
  if (!open) return null;
  return (
    <div>
      <h2>{title}</h2>
      <button type="button" onClick={onClose}>
        Close Drawer
      </button>
      <div>{children}</div>
      {footer && <div>{footer as React.ReactNode}</div>}
    </div>
  );
};
