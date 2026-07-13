import { ReactNode } from "react";
import { Link } from "react-router-dom";
import { ChevronRightIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";

export interface BreadcrumbItem {
  label: ReactNode;
  /** Linked when present, unless the item is the last in the list. */
  to?: string;
  title?: string;
}

interface BreadcrumbProps {
  items: BreadcrumbItem[];
  className?: string;
}

export default function Breadcrumb({ items, className = "" }: BreadcrumbProps) {
  return (
    <nav aria-label="Breadcrumb" className={cn("mb-5", className)}>
      <ol className="flex items-center gap-1.5 min-w-0">
        {items.map((item, i) => {
          const isLast = i === items.length - 1;
          const titleAttr =
            item.title ??
            (typeof item.label === "string" ? item.label : undefined);

          return (
            <li
              key={item.to ?? `leaf-${i}`}
              className="flex items-center gap-1.5 min-w-0"
            >
              {i > 0 && (
                <ChevronRightIcon
                  className="w-3 h-3 text-text-muted/40 shrink-0"
                  strokeWidth={2}
                  aria-hidden="true"
                />
              )}
              {item.to && !isLast ? (
                <Link
                  to={item.to}
                  title={titleAttr}
                  className="text-2xs font-mono text-text-muted hover:text-primary transition-colors truncate rounded-sm focus:outline-none focus-visible:ring-1 focus-visible:ring-primary/40 focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                >
                  {item.label}
                </Link>
              ) : (
                <span
                  aria-current={isLast ? "page" : undefined}
                  title={titleAttr}
                  className="text-2xs font-mono text-text-secondary truncate"
                >
                  {item.label}
                </span>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
