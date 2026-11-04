import { cn } from "@shellhub/design-system/cn";

/**
 * The look of a neutral control in the Devices toolbar, 36px tall to match the toolbar search.
 */
export const TOOLBAR_CONTROL =
  "h-9 inline-flex items-center gap-1.5 px-3 text-xs font-medium rounded-md border border-border bg-card text-text-secondary hover:text-text-primary transition-colors whitespace-nowrap";

/**
 * A count inside a toolbar control: yellow for devices waiting for approval, primary for active
 * filters.
 */
export function CountBadge({
  count,
  tone,
}: {
  count: number;
  tone: "waiting" | "active";
}) {
  return (
    <span
      className={cn(
        "min-w-4 h-4 px-1 rounded-full text-white text-3xs font-bold inline-flex items-center justify-center leading-none",
        tone === "waiting" ? "bg-accent-yellow" : "bg-primary",
      )}
    >
      {count}
    </span>
  );
}
