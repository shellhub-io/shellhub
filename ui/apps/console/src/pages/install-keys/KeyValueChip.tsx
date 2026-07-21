import { type ReactNode } from "react";

const CHIP =
  "inline-flex items-center gap-1.5 rounded-full border border-border bg-text-muted/[0.08] py-0.5 pl-0.5 pr-2.5";

/**
 * A labelled key/value chip: a tiny uppercase label pill followed by a mono value, on a soft rounded
 * fill. Used for the registration-activity device facts (MAC, source IP, the Device key fingerprint).
 * With `onClick` it renders as a button (the Device key opens its reveal); otherwise a static span.
 */
export default function KeyValueChip({
  label,
  value,
  onClick,
  trailing,
  title,
  ariaLabel,
  labelTone = "muted",
}: {
  /** The prepend label pill. Omit to show just the value (e.g. a bare masked secret). */
  label?: string;
  value: ReactNode;
  onClick?: () => void;
  /** A trailing affordance (e.g. a reveal eye) shown after the value. */
  trailing?: ReactNode;
  title?: string;
  ariaLabel?: string;
  /** Tint of the label pill. "primary" marks the star value (the install key itself). */
  labelTone?: "muted" | "primary";
}) {
  const labelCls =
    labelTone === "primary"
      ? "bg-primary/15 text-primary"
      : "bg-text-muted/[0.16] text-text-muted";

  const inner = (
    <>
      {label && (
        <span
          className={`rounded-full px-1.5 py-px font-mono text-[9px] font-semibold uppercase tracking-wider ${labelCls}`}
        >
          {label}
        </span>
      )}
      <span className="min-w-0 font-mono text-2xs text-text-secondary">
        {value}
      </span>
      {trailing}
    </>
  );

  if (onClick) {
    return (
      <button
        type="button"
        // Stop propagation so a chip inside a clickable row (the install key list) reveals the key
        // without also triggering the row's navigation.
        onClick={(e) => {
          e.stopPropagation();
          onClick();
        }}
        title={title}
        aria-label={ariaLabel}
        className={`${CHIP} transition-colors hover:border-primary/40 hover:bg-primary/5`}
      >
        {inner}
      </button>
    );
  }

  return (
    <span className={CHIP} title={title}>
      {inner}
    </span>
  );
}
