import { formatRelative } from "@/utils/date";

interface LastSeenCellProps {
  value?: string;
}

export default function LastSeenCell({ value }: LastSeenCellProps) {
  return (
    <span className="text-xs text-text-secondary">
      {formatRelative(value ?? "")}
    </span>
  );
}
