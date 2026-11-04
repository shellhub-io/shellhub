import { useState } from "react";
import { Link } from "react-router-dom";
import { ArrowRightIcon, ChevronDownIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { Dropdown } from "@shellhub/design-system/primitives";
import { useProvisioningKeys } from "@/hooks/useProvisioningKeyRequests";
import { useTenantRegistrationQueue } from "@/hooks/useProvisioningKeyEvents";
import { useNamespaceLegacy } from "@/hooks/useNamespaceLegacy";
import { capitalize } from "@/utils/string";
import {
  expiryPhrase,
  keyModeInfo,
  provisioningKeyDisplayName,
  provisioningKeyLink,
} from "@/pages/provisioning-keys/helpers";
import type { ProvisioningKey } from "@/client";
import { CountBadge, TOOLBAR_CONTROL } from "./toolbar";

/**
 * The Devices toolbar's count of devices waiting for approval, for a member who may accept devices.
 * In a new namespace it opens onto the provisioning keys holding them, read from the key list, each
 * linked to its activity, where they are accepted or rejected. In a legacy namespace it links the
 * tenant ID queue, counting its waiting registrations. Renders nothing while the namespace's mode is
 * loading, while nothing waits, and for a member who may not accept devices.
 */
export default function WaitingForApproval() {
  const legacy = useNamespaceLegacy();
  const queue = useTenantRegistrationQueue();
  const { provisioningKeys } = useProvisioningKeys({
    enabled: queue.allowed && legacy === false,
  });

  if (!queue.allowed || legacy === undefined) return null;

  if (legacy) {
    if (queue.waiting === 0) return null;

    return (
      <Link to="/devices/pending" className={TOOLBAR_CONTROL}>
        Waiting for approval <CountBadge count={queue.waiting} tone="waiting" />
        <ArrowRightIcon className="w-3 h-3" strokeWidth={2} />
      </Link>
    );
  }

  const holding = provisioningKeys.filter(
    (key) => (key.pending_devices ?? 0) > 0,
  );
  if (holding.length === 0) return null;

  return <KeysHoldingDevices holding={holding} />;
}

function KeysHoldingDevices({ holding }: { holding: ProvisioningKey[] }) {
  const [open, setOpen] = useState(false);
  const total = holding.reduce(
    (sum, key) => sum + (key.pending_devices ?? 0),
    0,
  );

  return (
    <Dropdown mode="content" portal open={open} onOpenChange={setOpen}>
      <Dropdown.Trigger>
        <button type="button" className={TOOLBAR_CONTROL}>
          Waiting for approval <CountBadge count={total} tone="waiting" />
          <ChevronDownIcon
            className={cn("w-3 h-3 transition-transform", open && "rotate-180")}
            strokeWidth={2}
          />
        </button>
      </Dropdown.Trigger>

      <Dropdown.Panel
        aria-label="Provisioning keys with devices waiting"
        className="w-[440px]"
      >
        <div className="px-3.5 py-2.5 bg-accent-yellow/10">
          <p className="text-sm font-semibold text-accent-yellow">
            {total === 1
              ? "1 device waiting for approval"
              : `${total} devices waiting for approval`}
          </p>
          <p className="text-xs text-text-secondary">
            Accept or reject them on the activity of the key they enrolled with.
          </p>
        </div>
        <ul>
          {holding.map((key, index) => (
            <li
              key={key.id}
              className={cn(index > 0 && "border-t border-border")}
            >
              <Link
                {...provisioningKeyLink(key)}
                className="group flex items-center gap-3 px-3.5 py-2.5 hover:bg-hover-subtle transition-colors"
              >
                <span className="flex-1 min-w-0">
                  <span className="block text-sm font-semibold text-text-primary truncate">
                    {provisioningKeyDisplayName(key)}
                  </span>
                  <span className="block text-xs text-text-muted truncate">
                    {keyModeInfo(key).label} · {capitalize(expiryPhrase(key))}
                  </span>
                </span>
                <span className="text-sm font-medium text-accent-yellow whitespace-nowrap">
                  {key.pending_devices} waiting
                </span>
                <span className="inline-flex items-center gap-1 text-xs whitespace-nowrap text-text-secondary group-hover:text-text-primary transition-colors">
                  Review
                  <ArrowRightIcon className="w-3 h-3" strokeWidth={2} />
                </span>
              </Link>
            </li>
          ))}
        </ul>
      </Dropdown.Panel>
    </Dropdown>
  );
}
