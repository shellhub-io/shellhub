import { formatRelative } from "@/utils/date";

interface LastSeenCellProps {
  value?: string;
}

/**
 * A last-seen timestamp as a relative phrase. An absent value renders as an em dash rather than
 * as the epoch.
 */
export default function LastSeenCell({ value }: LastSeenCellProps) {
  return (
    <span className="text-xs text-text-secondary">
      {formatRelative(value ?? "")}
    </span>
  );
}
