import { cn } from "@shellhub/design-system/cn";

/**
 * The REC mark: a pill whose dot is filled and pulsing while recording and an empty ring while
 * not. With onToggle it is the switch that turns recording on and off, named by label; without it
 * it only states that recording is on or off.
 */
export default function RecBadge({
  on,
  onToggle,
  label,
  describedBy,
  disabled,
}: {
  on: boolean;
  onToggle?: () => void;
  label?: string;
  describedBy?: string;
  disabled?: boolean;
}) {
  const className = cn(
    "inline-flex items-center gap-1.5 h-6 pl-1.5 pr-2 shrink-0 rounded-full border transition-all",
    on
      ? "border-accent-red/40 bg-accent-red/10 text-accent-red"
      : "border-border text-text-muted hover:border-border-light hover:text-text-secondary",
  );
  const content = (
    <>
      <span
        aria-hidden="true"
        className={cn(
          "w-3 h-3 rounded-full border-2 transition-all",
          on
            ? "border-accent-red bg-accent-red animate-pulse-subtle"
            : "border-current bg-transparent",
        )}
      />
      <span className="font-mono text-[10px] font-bold tracking-wider">
        REC
      </span>
    </>
  );

  if (!onToggle) {
    return <span className={className}>{content}</span>;
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
      {content}
    </button>
  );
}
