import { ComponentType, ReactNode, SVGProps } from "react";
import { Link, NavLink, Navigate, Outlet, useMatch } from "react-router-dom";
import { ChevronLeftIcon, ChevronRightIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import PageHeader from "@/components/common/PageHeader";
import { useIsDesktop } from "@/hooks/useIsDesktop";

/**
 * One entry of a SectionedLayout: the path segment the section lives at, relative to the layout's
 * base, and how its menu item reads.
 */
export interface LayoutSection {
  to: string;
  label: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
}

interface SectionedLayoutProps {
  base: string;
  icon: ReactNode;
  overline?: string;
  title: string;
  description: string;
  sections: LayoutSection[];
}

/**
 * A page split into sections, each at its own URL under base. On a wide window the menu of
 * sections runs down the left beside the one being read, and base opens the first. On a narrow
 * one they take turns: base is the menu alone under the page header, and a section fills the page
 * with a way back instead of the header, so the two never stack.
 */
export default function SectionedLayout({
  base,
  icon,
  overline,
  title,
  description,
  sections,
}: SectionedLayoutProps) {
  const isDesktop = useIsDesktop();
  const atMenu = useMatch(base) !== null;

  if (atMenu && isDesktop) return <Navigate to={sections[0].to} replace />;

  return (
    <div className="min-h-full flex flex-col">
      {(isDesktop || atMenu) && (
        <PageHeader
          icon={icon}
          overline={overline}
          title={title}
          description={description}
        />
      )}

      <div className="flex-1 flex flex-col lg:flex-row">
        {(isDesktop || atMenu) && (
          <nav
            aria-label={`${title} sections`}
            className="flex flex-col gap-1 lg:w-52 lg:shrink-0 lg:pr-6 lg:self-start lg:sticky lg:top-0"
          >
            {sections.map(({ to, label, icon: Icon }) => (
              <NavLink
                key={to}
                to={to}
                className={({ isActive }) =>
                  cn(
                    "flex items-center gap-2.5 px-3 rounded-lg text-sm transition-colors",
                    "py-3 lg:py-2 border border-border lg:border-transparent",
                    isActive
                      ? "bg-hover-medium text-text-primary font-medium"
                      : "text-text-secondary lg:text-text-muted hover:text-text-primary lg:hover:text-text-secondary hover:bg-hover-subtle",
                  )
                }
              >
                <Icon aria-hidden="true" className="w-4 h-4 shrink-0" />
                {label}
                <ChevronRightIcon
                  aria-hidden="true"
                  className="w-4 h-4 ml-auto text-text-muted lg:hidden"
                />
              </NavLink>
            ))}
          </nav>
        )}

        {!atMenu && (
          <div className="min-w-0 flex-1 lg:border-l lg:border-border lg:pl-8 animate-fade-in">
            {!isDesktop && (
              <Link
                to={base}
                className="inline-flex items-center gap-1 mb-5 text-sm text-text-muted hover:text-text-primary transition-colors"
              >
                <ChevronLeftIcon aria-hidden="true" className="w-4 h-4" />
                {title}
              </Link>
            )}
            <Outlet />
          </div>
        )}
      </div>
    </div>
  );
}
