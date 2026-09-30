import { ReactNode, useId, useRef } from "react";
import { XMarkIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useEscapeKey } from "@/hooks/useEscapeKey";
import { useFocusTrap } from "@/hooks/useFocusTrap";

interface DrawerProps {
  open: boolean;
  onClose: () => void;
  title: string;
  width?: "sm" | "md";
  children: ReactNode;
  bodyClassName?: string;
}

const WIDTH_MAP = {
  sm: "max-w-sm",
  md: "max-w-md",
};

/**
 * A side panel for a task done while watching the page behind it, such as tuning the terminal's
 * look against the terminal itself. Like BaseDialog it traps focus and closes on Escape, but it
 * does not cover the page. A flow that stops the work to finish something is a Modal.
 */
export default function Drawer({
  open,
  onClose,
  title,
  width = "md",
  children,
  bodyClassName,
}: DrawerProps) {
  const panelRef = useRef<HTMLDivElement>(null);
  const headingId = useId();
  useEscapeKey(onClose, open, panelRef);
  useFocusTrap(panelRef, open);

  return (
    <>
      <div
        role="presentation"
        className={cn(
          "fixed inset-0 z-drawer-backdrop bg-black/40 backdrop-blur-[2px] transition-opacity duration-300",
          open ? "opacity-100" : "opacity-0 pointer-events-none",
        )}
        onClick={onClose}
      />
      <div
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={headingId}
        aria-hidden={!open}
        {...(!open ? { inert: true } : {})}
        className={cn(
          "fixed inset-y-0 right-0 z-drawer w-full @container/drawer",
          WIDTH_MAP[width],
          "bg-surface border-l border-border shadow-2xl flex flex-col transition-transform duration-300 ease-out",
          open ? "translate-x-0" : "translate-x-full",
        )}
      >
        <div className="flex items-center justify-between px-6 py-4 border-b border-border shrink-0">
          <h2
            id={headingId}
            className="text-base font-semibold text-text-primary"
          >
            {title}
          </h2>
          <IconButton
            variant="ghost"
            aria-label="Close"
            data-dismiss
            onClick={onClose}
          >
            <XMarkIcon className="w-5 h-5" />
          </IconButton>
        </div>
        <div className={bodyClassName ?? "flex-1 overflow-y-auto px-6 py-5"}>
          {children}
        </div>
      </div>
    </>
  );
}
