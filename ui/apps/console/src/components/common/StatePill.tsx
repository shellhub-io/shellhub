import type { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";

/**
 * A small pill that shows whether something is on: tinted with onClassName while on, a muted
 * outline while off. With onToggle it is the switch that turns it on and off, named by label;
 * without it it only states the setting. children are the glyph and the short text inside.
 */
export default function StatePill({
  on,
  onClassName,
  onToggle,
  label,
  describedBy,
  disabled,
  children,
}: {
  on: boolean;
  onClassName: string;
  onToggle?: () => void;
  label?: string;
  describedBy?: string;
  disabled?: boolean;
  children: ReactNode;
}) {
  const className = cn(
    "inline-flex items-center gap-1.5 h-6 pl-1.5 pr-2 shrink-0 rounded-full border transition-all",
    on
      ? onClassName
      : "border-border text-text-muted hover:border-border-light hover:text-text-secondary",
  );

  if (!onToggle) {
    return <span className={className}>{children}</span>;
  }

  return (
    <button
      type="button"
      role="switch"
      aria-checked={on}
      aria-label={label}
      aria-describedby={describedBy}
      disabled={disabled}
      onClick={onToggle}
      className={cn(
        className,
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40 disabled:opacity-dim disabled:cursor-not-allowed",
      )}
    >
      {children}
    </button>
  );
}
