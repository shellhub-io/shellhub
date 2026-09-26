import { Fragment, useState } from "react";
import { NavLink, useLocation, useNavigate } from "react-router-dom";
import { CheckIcon, ChevronDownIcon } from "@heroicons/react/24/outline";
import { Dropdown } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { adminNavLinks, useAdminNav, type AdminNavSection } from "./adminNav";
import { navDisabled } from "./SidebarShell";

const barClass =
  "absolute inset-x-0 top-0 z-raised h-11 flex items-stretch border-b border-border bg-surface/90 backdrop-blur";

/**
 * The class that starts page content below the bar. It pairs with the bar's h-11 above, and lives
 * here so the two change together.
 */
export const belowAdminNavBar = "top-11";

function PagePicker({
  sections,
  disabled,
}: {
  sections: AdminNavSection[];
  disabled: boolean;
}) {
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const isCurrent = (to: string) =>
    pathname === to || pathname.startsWith(`${to}/`);
  const current = adminNavLinks(sections).find((link) => isCurrent(link.to));
  const CurrentIcon = current?.icon;

  return (
    <nav aria-label="Admin navigation" className={cn(barClass, "px-2")}>
      <Dropdown
        placement="bottom-start"
        portal
        open={open}
        onOpenChange={setOpen}
      >
        <Dropdown.Trigger>
          <button
            type="button"
            disabled={disabled}
            aria-label={`Admin page: ${current?.label ?? "choose a page"}`}
            className="self-center flex items-center gap-2 h-8 px-2.5 rounded-lg text-sm font-medium text-text-primary hover:bg-hover-medium disabled:opacity-dim transition-colors"
          >
            {CurrentIcon && (
              <CurrentIcon
                aria-hidden="true"
                className="w-4 h-4 text-primary"
              />
            )}
            {current?.label ?? "Admin"}
            <ChevronDownIcon
              aria-hidden="true"
              className={cn(
                "w-4 h-4 text-text-muted transition-transform",
                open && "rotate-180",
              )}
            />
          </button>
        </Dropdown.Trigger>
        <Dropdown.Panel aria-label="Admin pages" className="w-64 p-1.5">
          {sections.map((section, index) => (
            <Fragment key={section.title ?? index}>
              {section.title ? (
                <p className="px-2.5 pt-2.5 pb-1 text-2xs font-mono uppercase tracking-label text-text-muted">
                  {section.title}
                </p>
              ) : (
                index > 0 && <div className="my-1 h-px bg-border" />
              )}
              {section.items.map(({ to, label, icon: Icon }) => (
                <Dropdown.Item
                  key={to}
                  label={label}
                  aria-current={isCurrent(to) ? "page" : undefined}
                  onSelect={() => void navigate(to)}
                  className={cn(
                    "gap-2.5 py-2 text-sm",
                    isCurrent(to) && "text-text-primary",
                  )}
                >
                  <Icon aria-hidden="true" className="w-4 h-4 shrink-0" />
                  <span className="flex-1">{label}</span>
                  {isCurrent(to) && (
                    <CheckIcon
                      aria-hidden="true"
                      className="w-4 h-4 text-primary"
                      strokeWidth={2}
                    />
                  )}
                </Dropdown.Item>
              ))}
            </Fragment>
          ))}
        </Dropdown.Panel>
      </Dropdown>
    </nav>
  );
}

const itemBase =
  "relative flex items-center gap-1.5 h-full px-2.5 text-xs whitespace-nowrap transition-colors";

/**
 * The admin navigation as a bar across the top of the page frame, so the admin console reads as
 * a place apart from a namespace, whose navigation runs down the side. Sections sit side by side,
 * split by a rule. compact is for a window too narrow for every link to fit: there it becomes a
 * picker that names the current page and opens onto the rest.
 */
export default function AdminNavBar({ compact }: { compact: boolean }) {
  const { sections, disabled } = useAdminNav();

  if (compact) return <PagePicker sections={sections} disabled={disabled} />;

  return (
    <nav
      aria-label="Admin navigation"
      className={cn(
        barClass,
        "gap-1 px-4 overflow-x-auto [scrollbar-width:none]",
      )}
    >
      {sections.map((section, index) => (
        <Fragment key={section.title ?? index}>
          {index > 0 && (
            <span
              aria-hidden="true"
              className="self-center w-px h-4 mx-1.5 shrink-0 bg-border"
            />
          )}
          {section.items.map(({ to, label, icon: Icon }) =>
            disabled ? (
              <span
                key={to}
                aria-disabled="true"
                className={cn(
                  itemBase,
                  navDisabled,
                )}
              >
                <Icon aria-hidden="true" className="w-4 h-4 shrink-0" />
                {label}
              </span>
            ) : (
              <NavLink
                key={to}
                to={to}
                className={({ isActive }) =>
                  cn(
                    itemBase,
                    "after:absolute after:inset-x-2 after:-bottom-px after:h-0.5 after:rounded-full after:transition-colors",
                    isActive
                      ? "text-text-primary after:bg-primary"
                      : "text-text-muted hover:text-text-secondary after:bg-transparent",
                  )
                }
              >
                <Icon aria-hidden="true" className="w-4 h-4 shrink-0" />
                {label}
              </NavLink>
            ),
          )}
        </Fragment>
      ))}
    </nav>
  );
}
