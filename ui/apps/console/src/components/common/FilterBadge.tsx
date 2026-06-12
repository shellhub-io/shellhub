import { Tag } from "@/client";
import { Badge } from "@shellhub/design-system/primitives";
import {
  TagIcon,
  CpuChipIcon,
  GlobeAltIcon,
} from "@heroicons/react/24/outline";

export default function FilterBadge({
  filter,
}: {
  filter: { tags?: Tag[]; hostname?: string };
}) {
  if (filter.tags && filter.tags.length > 0) {
    return (
      <div className="flex flex-wrap gap-1">
        {filter.tags.map((tag) => (
          <Badge key={tag.name} color="primary">
            <TagIcon className="w-2.5 h-2.5" strokeWidth={2} />
            {tag.name}
          </Badge>
        ))}
      </div>
    );
  }

  if (filter.hostname && filter.hostname !== ".*") {
    return (
      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-accent-yellow/10 text-accent-yellow text-2xs rounded font-mono">
        <CpuChipIcon className="w-2.5 h-2.5" strokeWidth={2} />
        {filter.hostname}
      </span>
    );
  }

  return (
    <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-hover-medium text-text-muted text-2xs rounded">
      <GlobeAltIcon className="w-2.5 h-2.5" strokeWidth={2} />
      All devices
    </span>
  );
}
