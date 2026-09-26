import { ClockIcon } from "@heroicons/react/24/outline";
import { ExclamationCircleIcon } from "@heroicons/react/24/solid";
import { type ProvisioningKey } from "@/client";
import { getExpiryInfo, getKeyBlockers } from "./helpers";
import { cn } from "@shellhub/design-system/cn";

const ICON_CLASS = "w-3.5 h-3.5 shrink-0";

/**
 * When a provisioning key expires, as a phrase. A key with no expiry reads as "Never" rather than
 * as a date.
 */
export default function ExpiryLabel({
  provisioningKey,
  className,
}: {
  provisioningKey: ProvisioningKey;
  className?: string;
}) {
  const { expired, quiet } = getKeyBlockers(provisioningKey);

  return (
    <span
      className={cn("flex items-center gap-1 font-mono", className)}
      title={expired ? "Expired" : undefined}
    >
      {expired ? (
        <ExclamationCircleIcon
          className={cn(ICON_CLASS, !quiet && "text-accent-red")}
        />
      ) : (
        <ClockIcon className={ICON_CLASS} />
      )}
      {getExpiryInfo(provisioningKey.expires_at).label}
    </span>
  );
}
