import type { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";
import { NavItemLink } from "./SidebarShell";
import type { NavSection } from "./navSections";

interface NavSectionLink {
  to: string;
  label: string;
  icon: ReactNode;
}

/**
 * The sidebar's link groups, each under its title when it has one and named by it for assistive
 * technology. Collapsed, a title turns into a divider. disabled greys out every link; badge adds a marker after a link's label.
 */
export default function NavSectionList<T extends NavSectionLink>({
  sections,
  expanded,
  disabled,
  onNavClick,
  badge,
}: {
  sections: NavSection<T>[];
  expanded: boolean;
  disabled?: boolean;
  onNavClick?: () => void;
  badge?: (item: T) => ReactNode | undefined;
}) {
  return sections.map((section, idx) => (
    <div
      key={section.items[0]?.to}
      role={section.title ? "group" : undefined}
      aria-label={section.title}
      className={idx > 0 ? "mt-2.5" : ""}
    >
      {section.title && (
        <div className="relative mb-1.5 h-[14px]">
          <p
            className={cn(
              "px-3 text-2xs leading-[14px] font-mono font-semibold uppercase tracking-label text-text-muted/60 whitespace-nowrap overflow-hidden transition-opacity duration-200",
              expanded ? "opacity-100" : "opacity-0",
            )}
          >
            {section.title}
          </p>
          <span
            aria-hidden="true"
            className={cn(
              "absolute left-3 right-3 top-[3px] h-px bg-border transition-opacity duration-200",
              expanded ? "opacity-0" : "opacity-100",
            )}
          />
        </div>
      )}
      <div className="space-y-0.5">
        {section.items.map((item) => (
          <NavItemLink
            key={item.to}
            item={item}
            expanded={expanded}
            disabled={disabled}
            onClick={onNavClick}
            badge={badge?.(item)}
          />
        ))}
      </div>
    </div>
  ));
}
