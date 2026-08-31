import { cn } from "@shellhub/design-system/cn";
import { type InstallKey } from "@/client";
import { getKeyBlockers, getUsageInfo, type UsageInfo } from "./helpers";

function formatLabel(usage: UsageInfo): string {
  const cap = usage.kind === "unlimited" ? "∞" : usage.limit;
  return `${usage.used} / ${cap}`;
}

function Bar({
  usage,
  dimmed,
  reached,
}: {
  usage: UsageInfo;
  dimmed: boolean;
  reached: boolean;
}) {
  if (usage.kind === "unlimited") {
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

  const width = Math.max(usage.ratio * 100, usage.used > 0 ? 6 : 0) + "%";

  return (
    <div className="h-1.5 rounded-full bg-border/60 overflow-hidden">
      <div
        className={cn("h-full rounded-full transition-all duration-500", fill)}
        style={{ width }}
      />
    </div>
  );
}

/**
 * How much of an install key's allowance is spent. An unlimited key shows a count rather than a
 * bar, since there is nothing to fill.
 */
export default function UsageMeter({
  installKey,
  muted,
}: {
  installKey: InstallKey;
  muted?: boolean;
}) {
  const usage = getUsageInfo(installKey);
  const { inert, overused, revoked, disabled } = getKeyBlockers(installKey);
  const reached = overused && !revoked && !disabled;

  return (
    <div title={reached ? "Limit reached" : undefined}>
      <div
        className={cn(
          "mb-1.5 text-2xs font-mono",
          muted && "text-text-secondary",
        )}
      >
        {formatLabel(usage)}
      </div>
      <Bar usage={usage} dimmed={inert} reached={reached} />
    </div>
  );
}
