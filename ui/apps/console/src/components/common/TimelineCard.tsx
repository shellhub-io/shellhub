import { ClockIcon } from "@heroicons/react/24/outline";
import { Card } from "@shellhub/design-system/primitives";
import InfoItem from "@/components/common/InfoItem";
import { formatDateFull, formatRelative } from "@/utils/date";

interface TimelineCardProps {
  createdAt: string;
  lastSeen: string;
  statusUpdatedAt: string;
}

export default function TimelineCard({
  createdAt,
  lastSeen,
  statusUpdatedAt,
}: TimelineCardProps) {
  return (
    <Card className="p-5 space-y-4">
      <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
        <ClockIcon className="w-4 h-4 text-primary" />
        Timeline
      </h3>
      <dl className="space-y-3">
        <InfoItem label="Created" value={formatDateFull(createdAt)} />
        <InfoItem label="Last Seen">
          <span className="text-sm text-text-primary font-medium">
            {formatRelative(lastSeen)}
          </span>
          <span className="text-2xs text-text-muted">
            {formatDateFull(lastSeen)}
          </span>
        </InfoItem>
        <InfoItem
          label="Status Updated"
          value={formatDateFull(statusUpdatedAt)}
        />
      </dl>
    </Card>
  );
}
