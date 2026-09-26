import type { ComponentProps, ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";

/**
 * The session player's floating control bar, docked at the bottom centre of its positioned parent.
 * className carries whether it is shown; every other prop reaches the bar's element.
 */
export function PlayerBarShell({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      {...props}
      className={cn(
        "absolute bottom-[18px] left-1/2 -translate-x-1/2 w-[min(calc(100%-32px),660px)] h-[54px] flex items-center gap-2.5 pl-2.5 pr-2 rounded-xl border border-border bg-surface/90 backdrop-blur-sm shadow-lg transition-opacity duration-150",
        className,
      )}
    />
  );
}

/**
 * The player bar's position and length label, sized so the bar does not shift as the time ticks.
 */
export function PlayerTime({ children }: { children: ReactNode }) {
  return (
    <span className="shrink-0 min-w-[84px] text-center text-xs font-mono tabular-nums text-text-secondary">
      {children}
    </span>
  );
}
