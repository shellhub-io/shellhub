import { ReactNode, useId } from "react";
import { XMarkIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import BaseDialog, { type DialogSize } from "@/components/common/BaseDialog";

/**
 * Props of Modal. canClose refuses Escape and a backdrop click while it returns false; the close
 * button and onClose itself are the caller's to guard.
 */
export interface ModalProps {
  open: boolean;
  onClose: () => void;
  canClose?: () => boolean;
  title: string;
  subtitle?: ReactNode;
  icon?: ReactNode;
  size?: Exclude<DialogSize, "full">;
  children: ReactNode;
  footer?: ReactNode;
  bodyClassName?: string;
}

/**
 * The dialog for detail and edit flows: a title bar, a body and an optional footer of actions.
 * Only the body scrolls, so the title and the actions stay in view however long the content
 * runs. The panel is the `modal` container, which the fields laid out with container queries
 * read. The title bar is also the window's drag region inside the desktop app, since an open
 * modal leaves the app's own chrome inert.
 */
export default function Modal({
  open,
  onClose,
  canClose,
  title,
  subtitle,
  icon,
  size = "md",
  children,
  footer,
  bodyClassName,
}: ModalProps) {
  const headingId = useId();

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      canClose={canClose}
      size={size}
      aria-labelledby={headingId}
      className="overflow-hidden @container/modal"
    >
      <div
        data-tauri-drag-region
        className="flex items-center justify-between px-6 py-4 border-b border-border shrink-0"
      >
        <div
          data-tauri-drag-region
          className="flex items-center gap-2.5 min-w-0"
        >
          {icon && (
            <div
              data-tauri-drag-region
              className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center shrink-0"
            >
              {icon}
            </div>
          )}
          <div data-tauri-drag-region className="min-w-0">
            <h2
              data-tauri-drag-region
              id={headingId}
              className="text-base font-semibold text-text-primary"
            >
              {title}
            </h2>
            {subtitle && (
              <p
                data-tauri-drag-region
                className="text-2xs text-text-muted mt-0.5"
              >
                {subtitle}
              </p>
            )}
          </div>
        </div>
        <IconButton variant="ghost" aria-label="Close" onClick={onClose}>
          <XMarkIcon className="w-5 h-5" />
        </IconButton>
      </div>
      <div
        className={cn(
          "flex-1 min-h-0 overflow-y-auto overscroll-contain",
          bodyClassName ?? "px-6 py-5",
        )}
      >
        {children}
      </div>
      {footer && (
        <div className="px-6 py-4 border-t border-border shrink-0 flex flex-col-reverse gap-2 sm:flex-row sm:items-center sm:justify-end">
          {footer}
        </div>
      )}
    </BaseDialog>
  );
}
