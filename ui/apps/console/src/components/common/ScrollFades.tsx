import { ChevronDownIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";

/**
 * Fades over the top and bottom edges of a scrolling region that has more content past them, fed
 * by useScrollEdges. The parent must be positioned, and the region's background must be surface,
 * since the fades blend into it.
 */
export default function ScrollFades({
  moreAbove,
  moreBelow,
}: {
  moreAbove: boolean;
  moreBelow: boolean;
}) {
  return (
    <>
      <div
        aria-hidden="true"
        className={cn(
          "pointer-events-none absolute inset-x-0 top-0 h-8 bg-gradient-to-b from-surface to-transparent transition-opacity duration-200",
          moreAbove ? "opacity-100" : "opacity-0",
        )}
      />
      <div
        aria-hidden="true"
        className={cn(
          "pointer-events-none absolute inset-x-0 bottom-0 h-10 flex items-end justify-center pb-1 bg-gradient-to-t from-surface to-transparent transition-opacity duration-200",
          moreBelow ? "opacity-100" : "opacity-0",
        )}
      >
        <ChevronDownIcon
          className="w-4 h-4 text-text-muted/70"
          strokeWidth={2}
        />
      </div>
    </>
  );
}
