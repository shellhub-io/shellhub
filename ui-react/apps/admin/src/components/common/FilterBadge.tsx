import {
  TagIcon,
  ClipboardDocumentListIcon,
  GlobeAltIcon,
} from "@heroicons/react/24/outline";

export default function FilterBadge({
  filter,
}: {
  filter: { tags?: string[]; hostname?: string };
}) {
  if (filter.tags && filter.tags.length > 0) {
    return (
      <div className="flex flex-wrap gap-1">
        {filter.tags.map((tag) => (
          <span
            key={tag}
            className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-primary/10 text-primary text-2xs rounded font-medium"
          >
            <TagIcon className="w-2.5 h-2.5" strokeWidth={2} />
            {tag}
          </span>
        ))}
      </div>
    );
  }

  if (filter.hostname && filter.hostname !== ".*") {
    return (
      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-accent-yellow/10 text-accent-yellow text-2xs rounded font-mono">
        <ClipboardDocumentListIcon className="w-2.5 h-2.5" strokeWidth={2} />
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
