import { Spinner, type SpinnerSize } from "@shellhub/design-system/primitives";

type Padding = "none" | "sm" | "md" | "lg" | "fill";

interface PageLoaderProps {
  /** Announced to assistive tech. Also rendered as visible text when `showLabel` is true. */
  label: string;
  /** Spinner size. Defaults to `"lg"` when the label is hidden, `"md"` when `showLabel` is true. */
  size?: SpinnerSize;
  /** Render the label as visible text next to the spinner. Default `false`
   *  (label is delivered only to screen readers via the Spinner's aria-label). */
  showLabel?: boolean;
  /** Vertical breathing room inside the centered wrapper.
   *  - `"none"` — no padding (use when the parent already controls spacing)
   *  - `"sm"`   — `py-12` (drawers, popovers)
   *  - `"md"`   — `py-24` (default, most detail pages)
   *  - `"lg"`   — `py-32` (long-form settings pages)
   *  - `"fill"` — `flex-1` (fills the remaining height of a flex parent) */
  padding?: Padding;
}

const PADDING: Record<Padding, string> = {
  none: "",
  sm: "py-12",
  md: "py-24",
  lg: "py-32",
  fill: "flex-1",
};

export default function PageLoader({
  label,
  size,
  showLabel = false,
  padding = "md",
}: PageLoaderProps) {
  const resolvedSize = size ?? (showLabel ? "md" : "lg");
  const wrapper = ["flex h-full items-center justify-center", PADDING[padding]]
    .filter(Boolean)
    .join(" ");

  if (showLabel) {
    // role="status" is `nameFrom: author` — the visible text doesn't become
    // the accessible name, so set it explicitly via aria-label.
    return (
      <div role="status" aria-label={label} className={`${wrapper} gap-3`}>
        <Spinner size={resolvedSize} />
        <span className="text-xs font-mono text-text-muted">{label}</span>
      </div>
    );
  }

  return (
    <div className={wrapper}>
      <Spinner size={resolvedSize} aria-label={label} />
    </div>
  );
}
