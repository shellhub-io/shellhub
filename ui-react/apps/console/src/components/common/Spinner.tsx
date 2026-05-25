export type SpinnerSize = "xs" | "sm" | "md" | "lg" | "xl" | "2xl";
export type SpinnerTone = "onPrimary" | "onSurface" | "subtle";

interface SpinnerProps {
  size?: SpinnerSize;
  tone?: SpinnerTone;
  /** Layout overrides only (margins, block/inline-block). Don't override color or size. */
  className?: string;
  /**
   * If provided, the spinner is announced to assistive tech as a live status
   * region. Omit for purely decorative spinners (the default).
   */
  "aria-label"?: string;
}

const SIZE: Record<SpinnerSize, string> = {
  xs: "w-3 h-3",
  sm: "w-3.5 h-3.5",
  md: "w-4 h-4",
  lg: "w-5 h-5",
  xl: "w-6 h-6",
  "2xl": "w-10 h-10",
};

const TONE: Record<SpinnerTone, string> = {
  onPrimary: "border-white/30 border-t-white",
  onSurface: "border-primary/30 border-t-primary",
  subtle: "border-text-muted/30 border-t-text-muted",
};

export default function Spinner({
  size = "md",
  tone = "onSurface",
  className,
  "aria-label": ariaLabel,
}: SpinnerProps) {
  const classes = [
    SIZE[size],
    TONE[tone],
    "border-2 rounded-full animate-spin",
    className,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <span
      role={ariaLabel ? "status" : undefined}
      aria-label={ariaLabel}
      aria-hidden={ariaLabel ? undefined : true}
      className={classes}
    />
  );
}
