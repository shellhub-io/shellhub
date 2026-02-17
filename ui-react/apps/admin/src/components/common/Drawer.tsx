import { ReactNode } from "react";
import { XMarkIcon } from "@heroicons/react/24/outline";

interface DrawerProps {
  open: boolean;
  onClose: () => void;
  title: string;
  subtitle?: ReactNode;
  icon?: ReactNode;
  width?: "sm" | "md";
  children: ReactNode;
  footer?: ReactNode;
  bodyClassName?: string;
}

const WIDTH_MAP = {
  sm: "max-w-sm",
  md: "max-w-md",
};

export default function Drawer({
  open,
  onClose,
  title,
  subtitle,
  icon,
  width = "md",
  children,
  footer,
  bodyClassName,
}: DrawerProps) {
  return (
    <>
      <div
        className={`fixed inset-0 z-[60] bg-black/40 backdrop-blur-[2px] transition-opacity duration-300 ${
          open ? "opacity-100" : "opacity-0 pointer-events-none"
        }`}
        onClick={onClose}
      />
      <div
        className={`fixed inset-y-0 right-0 z-[70] w-full ${WIDTH_MAP[width]} bg-surface border-l border-border shadow-2xl flex flex-col transition-transform duration-300 ease-out ${
          open ? "translate-x-0" : "translate-x-full"
        }`}
      >
        <div className="flex items-center justify-between px-6 py-4 border-b border-border shrink-0">
          <div className="flex items-center gap-2.5">
            {icon && (
              <div className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center">
                {icon}
              </div>
            )}
            <div>
              <h2 className="text-base font-semibold text-text-primary">
                {title}
              </h2>
              {subtitle && (
                <p className="text-2xs text-text-muted mt-0.5">{subtitle}</p>
              )}
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
          >
            <XMarkIcon className="w-5 h-5" />
          </button>
        </div>
        <div className={bodyClassName ?? "flex-1 overflow-y-auto px-6 py-5"}>
          {children}
        </div>
        {footer && (
          <div className="px-6 py-4 border-t border-border shrink-0 flex items-center justify-end gap-2">
            {footer}
          </div>
        )}
      </div>
    </>
  );
}
