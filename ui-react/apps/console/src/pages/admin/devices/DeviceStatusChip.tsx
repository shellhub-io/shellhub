import {
  CheckCircleIcon,
  ClockIcon,
  XCircleIcon,
  MinusCircleIcon,
} from "@heroicons/react/24/outline";
import type { DeviceStatus } from "../../../client";

const STATUS_CONFIG: Record<
  DeviceStatus,
  {
    Icon: typeof CheckCircleIcon;
    label: string;
    className: string;
  }
> = {
  accepted: {
    Icon: CheckCircleIcon,
    label: "Accepted",
    className:
      "bg-accent-green/10 text-accent-green border border-accent-green/20",
  },
  pending: {
    Icon: ClockIcon,
    label: "Pending",
    className:
      "bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20",
  },
  rejected: {
    Icon: XCircleIcon,
    label: "Rejected",
    className: "bg-accent-red/10 text-accent-red border border-accent-red/20",
  },
  removed: {
    Icon: MinusCircleIcon,
    label: "Removed",
    className: "bg-text-muted/10 text-text-muted border border-text-muted/20",
  },
  unused: {
    Icon: MinusCircleIcon,
    label: "Unused",
    className: "bg-text-muted/10 text-text-muted border border-text-muted/20",
  },
};

interface DeviceStatusChipProps {
  status: DeviceStatus;
}

export default function DeviceStatusChip({ status }: DeviceStatusChipProps) {
  const config = STATUS_CONFIG[status] ?? STATUS_CONFIG.pending;
  const { Icon, label, className } = config;

  return (
    <span
      className={`inline-flex items-center gap-1 px-2 py-0.5 text-2xs font-semibold rounded-md ${className}`}
    >
      <Icon className="w-3 h-3" />
      {label}
    </span>
  );
}
