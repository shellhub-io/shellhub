import { type ComponentType, type SVGProps } from "react";
import {
  ArrowsRightLeftIcon,
  CheckBadgeIcon,
  HandRaisedIcon,
  ListBulletIcon,
} from "@heroicons/react/24/outline";
import StatusChip from "./StatusChip";

type IconType = ComponentType<SVGProps<SVGSVGElement>>;

/**
 * Single source of truth for the enrollment modes: the icon that carries each mode's identity, its
 * label, a one-liner of what it does, and the longer description used in the create/edit selector.
 */
export const MODE_INFO: Record<
  string,
  { label: string; icon: IconType; summary: string; description: string }
> = {
  automatic: {
    label: "Automatic",
    icon: CheckBadgeIcon,
    summary: "Auto-accepts",
    description: "Accept every device that registers with this key.",
  },
  manual: {
    label: "Manual",
    icon: HandRaisedIcon,
    summary: "Manual review",
    description:
      "Leave registered devices pending for you to review and accept.",
  },
  webhook: {
    label: "Webhook",
    icon: ArrowsRightLeftIcon,
    summary: "Endpoint decides",
    description:
      "Ask your endpoint at registration whether to accept, reject, or leave the device pending.",
  },
  allowlist: {
    label: "MAC allowlist",
    icon: ListBulletIcon,
    summary: "Accepts listed MACs",
    description:
      "Accept a device only when its MAC is on the list below; reject the rest.",
  },
};

export function modeInfo(mode: string) {
  return MODE_INFO[mode] ?? MODE_INFO.automatic;
}

/**
 * Marks the namespace's auto-managed tenant-only registration key: a legacy path kept for
 * compatibility and slated for removal, so it carries a caution (deprecated) badge.
 */
export function DeprecatedBadge() {
  return <StatusChip label="Deprecated" tone="yellow" />;
}
