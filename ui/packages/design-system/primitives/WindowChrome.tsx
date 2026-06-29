import type { ReactNode } from "react";
import { LockClosedIcon } from "@heroicons/react/24/outline";
import { cn } from "./cn";

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

export type WindowChromeVariant = "terminal" | "browser";
export type WindowChromeSize = "sm" | "md";
export type WindowChromeAccent =
  | "green"
  | "cyan"
  | "primary"
  | "red"
  | "yellow";

export interface WindowChromeProps {
  variant: WindowChromeVariant;
  size?: WindowChromeSize;
  title?: string;
  path?: string;
  /** Optional 4th animated pulse dot, by accent token */
  accent?: WindowChromeAccent;
  /** Slot rendered on the right side of the title bar */
  titleBarSlot?: ReactNode;
  className?: string;
  bodyClassName?: string;
  children?: ReactNode;
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

const DOT_SIZE: Record<WindowChromeSize, string> = {
  sm: "w-2.5 h-2.5",
  md: "w-3 h-3",
};

const ACCENT_BG: Record<WindowChromeAccent, string> = {
  green: "bg-accent-green",
  cyan: "bg-accent-cyan",
  primary: "bg-primary",
  red: "bg-accent-red",
  yellow: "bg-accent-yellow",
};

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function WindowChrome({
  variant,
  size = "md",
  title,
  path,
  accent,
  titleBarSlot,
  className,
  bodyClassName,
  children,
}: WindowChromeProps) {
  const dotSize = DOT_SIZE[size];

  return (
    <div
      className={cn(
        "rounded-xl border border-border bg-surface overflow-hidden",
        className,
      )}
    >
      {/* Title bar */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border">
        {/* Traffic-light dots */}
        <div className="flex items-center gap-1.5 shrink-0">
          <span className={cn("rounded-full", dotSize, "bg-accent-red/60")} />
          <span
            className={cn("rounded-full", dotSize, "bg-accent-yellow/60")}
          />
          <span className={cn("rounded-full", dotSize, "bg-accent-green/60")} />
          {accent && (
            <span
              className={cn(
                "rounded-full animate-pulse",
                dotSize,
                ACCENT_BG[accent],
              )}
            />
          )}
        </div>

        {/* Center content */}
        <div className="flex flex-1 items-center min-w-0">
          {variant === "terminal" && title && (
            <span className="text-2xs text-text-muted font-mono truncate">
              {title}
            </span>
          )}
          {variant === "browser" && (
            <div className="flex items-center gap-1 text-xs text-text-secondary min-w-0 max-w-xs w-full bg-background rounded px-2 py-0.5">
              <LockClosedIcon className="w-3 h-3 shrink-0" strokeWidth={2} />
              {path && (
                <span className="truncate font-mono">shellhub.io{path}</span>
              )}
            </div>
          )}
        </div>

        {/* Actions slot */}
        {titleBarSlot && (
          <div className="flex items-center gap-1 shrink-0">{titleBarSlot}</div>
        )}
      </div>

      {/* Body */}
      <div
        className={cn(
          "relative p-5 font-mono text-xs leading-relaxed",
          bodyClassName,
        )}
      >
        {children}
      </div>
    </div>
  );
}
