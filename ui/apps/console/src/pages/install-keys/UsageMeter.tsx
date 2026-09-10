import { ClockIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { type InstallKey } from "@/client";
import StatusChip from "./StatusChip";
import {
  getKeyBlockers,
  getUsageInfo,
  getWaitingInfo,
  type UsageInfo,
} from "./helpers";

function formatLabel(usage: UsageInfo): string {
  const cap = usage.kind === "unlimited" ? "∞" : usage.limit;
  return `${usage.used} / ${cap}`;
}

function Bar({
  usage,
  dimmed,
  reached,
  waiting,
  oversubscribed,
}: {
  usage: UsageInfo;
  dimmed: boolean;
  reached: boolean;
  waiting: number;
  oversubscribed: boolean;
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

  const usedWidth = Math.max(usage.ratio * 100, usage.used > 0 ? 6 : 0);

  if (waiting === 0) {
    return (
      <div className="h-1.5 rounded-full bg-border/60 overflow-hidden">
        <div
          className={cn(
            "h-full rounded-full transition-all duration-500",
            fill,
          )}
          style={{ width: usedWidth + "%" }}
        />
      </div>
    );
  }

  const waitingWidth = Math.min(100 - usedWidth, (waiting / usage.limit) * 100);

  return (
    <div className="h-1.5 rounded-full bg-border/60 overflow-hidden flex">
      <div
        className={cn("h-full transition-all duration-500", fill)}
        style={{ width: usedWidth + "%" }}
      />
      <div
        className={cn(
          "h-full",
          oversubscribed
            ? "usage-oversubscribed text-accent-red"
            : "bg-primary/35",
        )}
        style={{ width: waitingWidth + "%" }}
      />
    </div>
  );
}

/**
 * How much of an install key's allowance is spent, and how much of it is claimed by devices still
 * awaiting a decision. An unlimited key shows a count rather than a bar, since there is nothing
 * to fill.
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
  const { waiting, beyondLimit, oversubscribed } = getWaitingInfo(installKey);

  return (
    <div
      className="min-w-[7.5rem]"
      title={reached ? "Limit reached" : undefined}
    >
      {waiting > 0 ? (
        <div className="mb-2 flex whitespace-nowrap">
          <StatusChip
            icon={ClockIcon}
            label={`${waiting} waiting`}
            tone="primary"
          />
        </div>
      ) : (
        <div
          className={cn(
            "mb-1.5 text-2xs font-mono",
            muted && "text-text-secondary",
          )}
        >
          {formatLabel(usage)}
        </div>
      )}
      <Bar
        usage={usage}
        dimmed={inert}
        reached={reached}
        waiting={waiting}
        oversubscribed={oversubscribed}
      />
      {waiting > 0 && (
        <div
          className={cn(
            "mt-1.5 text-2xs font-mono",
            oversubscribed ? "text-accent-red" : "text-text-muted",
          )}
        >
          {formatLabel(usage)} used
          {oversubscribed && ` · ${beyondLimit} over`}
        </div>
      )}
    </div>
  );
}
