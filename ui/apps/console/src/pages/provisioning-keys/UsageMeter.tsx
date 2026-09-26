import { cn } from "@shellhub/design-system/cn";
import { type ProvisioningKey } from "@/client";
import {
  getKeyBlockers,
  getUsageInfo,
  getWaitingInfo,
  type UsageInfo,
} from "./helpers";

const SIZE = 22;
const STROKE = 3;
const RADIUS = (SIZE - STROKE) / 2;
const CIRCUMFERENCE = 2 * Math.PI * RADIUS;

function Arc({
  from,
  to,
  className,
}: {
  from: number;
  to: number;
  className: string;
}) {
  const length = Math.max(0, to - from) * CIRCUMFERENCE;
  if (length === 0) return null;
  return (
    <circle
      cx={SIZE / 2}
      cy={SIZE / 2}
      r={RADIUS}
      fill="none"
      strokeWidth={STROKE}
      strokeLinecap="round"
      strokeDasharray={`${length} ${CIRCUMFERENCE}`}
      strokeDashoffset={-from * CIRCUMFERENCE}
      className={cn("transition-all duration-500", className)}
    />
  );
}

function Ring({
  usage,
  waiting,
  dimmed,
  reached,
  oversubscribed,
}: {
  usage: UsageInfo;
  waiting: number;
  dimmed: boolean;
  reached: boolean;
  oversubscribed: boolean;
}) {
  if (usage.kind === "unlimited") {
    return (
      <span
        aria-hidden="true"
        style={{ width: SIZE, height: SIZE }}
        className="grid place-items-center shrink-0 rounded-full border-2 border-dashed border-border text-[11px] leading-none text-text-muted"
      >
        ∞
      </span>
    );
  }

  const used = Math.min(1, usage.ratio);
  const claimed = Math.min(1, used + waiting / usage.limit);

  return (
    <svg
      aria-hidden="true"
      width={SIZE}
      height={SIZE}
      viewBox={`0 0 ${SIZE} ${SIZE}`}
      className="shrink-0 -rotate-90"
    >
      <circle
        cx={SIZE / 2}
        cy={SIZE / 2}
        r={RADIUS}
        fill="none"
        strokeWidth={STROKE}
        className="stroke-border"
      />
      <Arc
        from={used}
        to={claimed}
        className={oversubscribed ? "stroke-accent-red" : "stroke-primary/35"}
      />
      <Arc
        from={0}
        to={used}
        className={
          reached
            ? "stroke-accent-yellow"
            : dimmed
              ? "stroke-text-muted/40"
              : "stroke-primary"
        }
      />
    </svg>
  );
}

function detail(
  usage: UsageInfo,
  waiting: number,
  beyondLimit: number,
  closed: boolean,
) {
  if (waiting > 0) {
    return beyondLimit > 0
      ? `${waiting} waiting · ${beyondLimit} over`
      : `${waiting} waiting`;
  }
  if (closed) return "not enrolling";
  if (usage.kind === "unlimited") return "no limit";
  const left = Math.max(0, usage.limit - usage.used);
  return left === 0 ? "limit reached" : `${left} left`;
}

/**
 * How much of a provisioning key's quota is spent, drawn as a ring the way a quota usually is,
 * with the count beside it. Devices still awaiting a decision claim a lighter part of the ring,
 * red once they would push the key past its limit. An unlimited key has nothing to fill, so it
 * shows its count and says there is no limit.
 */
export default function UsageMeter({
  provisioningKey,
  muted,
}: {
  provisioningKey: ProvisioningKey;
  muted?: boolean;
}) {
  const usage = getUsageInfo(provisioningKey);
  const { inert, overused, expired, quiet } = getKeyBlockers(provisioningKey);
  const reached = overused && !quiet;
  const { waiting, beyondLimit, oversubscribed } =
    getWaitingInfo(provisioningKey);
  const text = detail(usage, waiting, beyondLimit, quiet || expired);

  return (
    <div
      className="flex items-center gap-2.5 min-w-[7.5rem]"
      title={reached ? "Limit reached" : undefined}
    >
      <Ring
        usage={usage}
        waiting={waiting}
        dimmed={inert}
        reached={reached}
        oversubscribed={oversubscribed}
      />
      <div className="min-w-0 leading-tight">
        <div
          className={cn(
            "text-xs font-mono",
            muted ? "text-text-secondary" : "text-text-primary",
          )}
        >
          {usage.kind === "unlimited"
            ? `${usage.used} used`
            : `${usage.used} / ${usage.limit}`}
        </div>
        <div
          className={cn(
            "mt-0.5 text-2xs",
            oversubscribed
              ? "text-accent-red"
              : reached
                ? "text-accent-yellow"
                : "text-text-muted",
          )}
        >
          {text}
        </div>
      </div>
    </div>
  );
}
