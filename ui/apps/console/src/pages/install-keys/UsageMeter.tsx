import { type UsageInfo } from "./helpers";

/**
 * The usage bar: how much of a key's allowance is spent. The bar itself carries the state — it turns
 * amber once the key reaches its limit (so "limit reached" reads off the bar, no separate chip); a key
 * that is inert for another reason greys out. Unlimited keys render an infinity track.
 */
export default function UsageMeter({
  usage,
  dimmed,
  reached,
}: {
  usage: UsageInfo;
  dimmed: boolean;
  reached: boolean;
}) {
  if (usage.kind === "unlimited") {
    // The KeyValueChip shell (soft fill + hairline border) with an animated barber-pole hatch, so an
    // unlimited key reads as a defined, "always-on" bar rather than floating stripes.
    return (
      <div className="h-1.5 overflow-hidden rounded-full border border-border bg-text-muted/[0.08]">
        <div className="usage-infinity h-full" />
      </div>
    );
  }

  const fill = reached
    ? "bg-accent-yellow"
    : dimmed
      ? "bg-text-muted/40"
      : "bg-primary";

  return (
    <div className="h-1.5 rounded-full bg-border/60 overflow-hidden">
      <div
        className={`h-full rounded-full transition-all duration-500 ${fill}`}
        style={{
          width: `${Math.max(usage.ratio * 100, usage.used > 0 ? 6 : 0)}%`,
        }}
      />
    </div>
  );
}

// Uniform `used / cap`, cap being a number or ∞ for unlimited. State (spent, over limit) is carried by
// the meter's colour, not a text suffix, so the label stays even.
export function usageLabel(usage: UsageInfo): string {
  const cap = usage.kind === "unlimited" ? "∞" : usage.limit;

  return `${usage.used} / ${cap}`;
}
