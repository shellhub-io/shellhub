import { Link } from "react-router-dom";
import { CpuChipIcon } from "@heroicons/react/24/outline";
import DistroIcon from "./DistroIcon";

interface DeviceChipProps {
  uid: string;
  name: string;
  online?: boolean;
  osId?: string;
  onClick?: (e: React.MouseEvent) => void;
}

export default function DeviceChip({
  uid,
  name,
  online,
  osId,
  onClick,
}: DeviceChipProps) {
  return (
    <Link
      to={`/devices/${uid}`}
      onClick={onClick}
      className="
        inline-flex items-center gap-1.5
        px-2 py-1
        bg-surface border border-border
        rounded-md
        text-xs font-mono font-medium text-text-secondary
        hover:text-text-primary hover:border-primary/40 hover:bg-primary/5
        transition-all duration-150
        group/chip
      "
    >
      {osId ? (
        <DistroIcon
          id={osId}
          className="text-[0.8rem] leading-none text-text-muted group-hover/chip:text-text-secondary shrink-0 transition-colors"
        />
      ) : (
        <CpuChipIcon
          className="w-3 h-3 text-text-muted group-hover/chip:text-primary shrink-0 transition-colors"
          strokeWidth={2}
        />
      )}

      <span className="truncate max-w-[16ch]">{name}</span>

      {/* online dot */}
      {online !== undefined && (
        <span
          className={`w-1.5 h-1.5 rounded-full shrink-0 ${
            online
              ? "bg-accent-green shadow-[0_0_4px_rgba(130,165,104,0.6)]"
              : "bg-text-muted/40"
          }`}
        />
      )}
    </Link>
  );
}
