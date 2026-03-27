import {
  CheckCircleIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";

export type UserStatus = "confirmed" | "not-confirmed";

const STATUS_CONFIG: Record<
  UserStatus,
  {
    Icon: typeof CheckCircleIcon;
    label: string;
    className: string;
  }
> = {
  confirmed: {
    Icon: CheckCircleIcon,
    label: "Confirmed",
    className:
      "bg-accent-green/10 text-accent-green border border-accent-green/20",
  },
  "not-confirmed": {
    Icon: ExclamationCircleIcon,
    label: "Not Confirmed",
    className: "bg-accent-red/10 text-accent-red border border-accent-red/20",
  },
};

interface UserStatusChipProps {
  status: UserStatus;
}

export default function UserStatusChip({ status }: UserStatusChipProps) {
  const config = STATUS_CONFIG[status] ?? STATUS_CONFIG["not-confirmed"];
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
