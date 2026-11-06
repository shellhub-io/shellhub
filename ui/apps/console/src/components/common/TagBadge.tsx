import { TagIcon } from "@heroicons/react/24/outline";
import { Badge } from "@shellhub/design-system/primitives";

/**
 * A tag's name as a read-only label.
 */
export default function TagBadge({ name }: { name: string }) {
  return (
    <Badge color="primary">
      <TagIcon className="w-2.5 h-2.5 shrink-0" strokeWidth={2} />
      {name}
    </Badge>
  );
}
