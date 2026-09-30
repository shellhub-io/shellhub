import { Link } from "react-router-dom";
import { cn } from "@shellhub/design-system/cn";
import { type ProvisioningKey } from "@/client";
import {
  formatCount,
  getKeyBlockers,
  getUsageInfo,
  getWaitingInfo,
  lastUsedPhrase,
  provisioningKeyLink,
  type UsageInfo,
} from "./helpers";

function detail(
  usage: UsageInfo,
  waiting: number,
  beyondLimit: number,
  closed: boolean,
) {
  if (waiting > 0) {
    return beyondLimit > 0
      ? `${formatCount(waiting)} waiting · ${formatCount(beyondLimit)} over`
      : `${formatCount(waiting)} waiting`;
  }
  if (closed) return "not enrolling";
  if (usage.kind === "unlimited") return undefined;
  return usage.used >= usage.limit ? "limit reached" : undefined;
}

/**
 * A key's uses against its allowance, "3 of 10 devices" or "3 of unlimited devices", with the
 * count itself carrying the weight. large sets the count as a page's headline number; muted
 * softens it for a dense list.
 */
export function UsageCount({
  provisioningKey,
  muted,
  large,
}: {
  provisioningKey: ProvisioningKey;
  muted?: boolean;
  large?: boolean;
}) {
  const usage = getUsageInfo(provisioningKey);
  return (
    <div
      className={cn(
        "text-text-muted whitespace-nowrap",
        large ? "text-sm" : "text-xs",
      )}
    >
      <span
        className={cn(
          muted ? "text-text-secondary" : "text-text-primary",
          large
            ? "mr-1.5 text-2xl font-semibold tracking-tight"
            : "font-medium",
        )}
      >
        {formatCount(usage.used)}
      </span>
      {usage.kind === "unlimited"
        ? " of unlimited devices"
        : ` of ${formatCount(usage.limit)} devices`}
    </div>
  );
}

/**
 * What a key's count leaves out, or nothing when the count says it all: devices waiting for a
 * decision (red once accepting them would pass the limit), a spent limit, or a key that stopped
 * enrolling. linkWaiting turns the waiting count into a link to the key's page.
 */
export function UsageDetail({
  provisioningKey,
  linkWaiting,
  className,
}: {
  provisioningKey: ProvisioningKey;
  linkWaiting?: boolean;
  className?: string;
}) {
  const usage = getUsageInfo(provisioningKey);
  const { overused, expired, quiet } = getKeyBlockers(provisioningKey);
  const reached = overused && !quiet;
  const { waiting, beyondLimit, oversubscribed } =
    getWaitingInfo(provisioningKey);
  const text = detail(usage, waiting, beyondLimit, quiet || expired);
  if (!text) return null;

  return (
    <span
      className={cn(
        oversubscribed
          ? "text-accent-red"
          : reached || waiting > 0
            ? "text-accent-yellow"
            : "text-text-muted",
        className,
      )}
      title={reached ? "Limit reached" : undefined}
    >
      {linkWaiting && waiting > 0 ? (
        <Link
          {...provisioningKeyLink(provisioningKey)}
          onClick={(e) => e.stopPropagation()}
          className="underline decoration-dotted underline-offset-2 hover:text-primary"
        >
          {text}
        </Link>
      ) : (
        text
      )}
    </span>
  );
}

/**
 * A key's usage in a list cell: the count, what it leaves out, and when the key was last used.
 */
export default function UsageMeter({
  provisioningKey,
  muted,
  linkWaiting,
}: {
  provisioningKey: ProvisioningKey;
  muted?: boolean;
  linkWaiting?: boolean;
}) {
  return (
    <div className="min-w-[7.5rem] leading-tight">
      <UsageCount provisioningKey={provisioningKey} muted={muted} />
      <UsageDetail
        provisioningKey={provisioningKey}
        linkWaiting={linkWaiting}
        className="mt-0.5 block text-2xs"
      />
      <div className="mt-0.5 text-2xs text-text-muted whitespace-nowrap">
        {lastUsedPhrase(provisioningKey)}
      </div>
    </div>
  );
}
