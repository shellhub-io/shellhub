import { type ComponentType, type SVGProps } from "react";
import {
  ArrowsRightLeftIcon,
  CheckBadgeIcon,
  HandRaisedIcon,
  ListBulletIcon,
} from "@heroicons/react/24/outline";

type IconType = ComponentType<SVGProps<SVGSVGElement>>;

/**
 * What each enrolment mode is called and what it does, in one place, so the selector and the
 * list's Enrollment cell describe a mode identically.
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

/**
 * The description for a mode, falling back to automatic for one this build does not know — a
 * key made by a newer server still renders rather than showing a blank cell.
 */
export function modeInfo(mode: string) {
  return MODE_INFO[mode] ?? MODE_INFO.automatic;
}
