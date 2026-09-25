import { useState } from "react";
import { NavLink, useLocation } from "react-router-dom";
import { ChevronDownIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import SessionMenu from "./SessionMenu";
import SidebarShell, {
  NavItemLink,
  navBase,
  navDisabled,
  navIcon,
} from "./SidebarShell";
import { isAdminNavGroup, useAdminNav, type AdminNavGroup } from "./adminNav";

function NavGroupItem({
  group,
  expanded,
  isOpen,
  disabled,
  onToggle,
  currentPath,
  onNavClick,
}: {
  group: AdminNavGroup;
  expanded: boolean;
  isOpen: boolean;
  disabled?: boolean;
  onToggle: () => void;
  currentPath: string;
  onNavClick?: () => void;
}) {
  const isChildActive =
    !disabled && group.children.some((c) => currentPath.startsWith(c.to));
  const isGroupActive = isChildActive || (expanded && isOpen);

  return (
    <div>
      <button
        type="button"
        onClick={disabled ? undefined : onToggle}
        title={expanded ? undefined : group.label}
        aria-expanded={disabled ? undefined : isGroupActive}
        aria-disabled={disabled || undefined}
        className={cn(
          "w-full",
          navBase,
          "transition-all duration-150",
          disabled
            ? navDisabled
            : isChildActive
              ? "text-primary"
              : "text-text-secondary hover:text-text-primary hover:bg-hover-subtle",
        )}
      >
        <group.icon className={navIcon} />
        <span
          className={cn(
            "flex-1 min-w-0 flex items-center gap-3 transition-opacity duration-200",
            expanded ? "opacity-100" : "opacity-0",
          )}
        >
          <span className="flex-1 text-left truncate">{group.label}</span>
          {!disabled && (
            <ChevronDownIcon
              className={cn(
                "w-3.5 h-3.5 shrink-0 transition-transform duration-200",
                isGroupActive && "rotate-180",
              )}
              strokeWidth={2}
            />
          )}
        </span>
      </button>
      {!disabled && expanded && isGroupActive ? (
        <div className="ml-5 mt-0.5 space-y-0.5 border-l border-border pl-3">
          {group.children.map((child) => (
            <NavLink
              key={child.to}
              to={child.to}
              onClick={onNavClick}
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-2 px-2 py-1.5 rounded-md text-[12px] font-medium transition-all duration-150",
                  isActive
                    ? "text-primary bg-primary/5"
                    : "text-text-secondary hover:text-text-primary hover:bg-hover-subtle",
                )
              }
            >
              <span className="truncate">{child.label}</span>
            </NavLink>
          ))}
        </div>
      ) : null}
    </div>
  );
}

/**
 * The admin navigation. Its links are instance-wide, so nothing here is filtered by namespace
 * role, only by whether the user is an instance admin at all.
 */
export default function AdminSidebar({
  expanded,
  onClose,
}: {
  expanded: boolean;
  onClose?: () => void;
}) {
  const { entries, disabled } = useAdminNav();
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>({});
  const { pathname } = useLocation();

  const toggleGroup = (label: string) => {
    setOpenGroups((prev) => ({ ...prev, [label]: !prev[label] }));
  };

  return (
    <SidebarShell
      expanded={expanded}
      onClose={onClose}
      ariaLabel="Admin navigation"
      logoHref="/admin/dashboard"
      account={<SessionMenu expanded={expanded} />}
    >
      <div className="space-y-0.5">
        {entries.map((entry) =>
          isAdminNavGroup(entry) ? (
            <NavGroupItem
              key={entry.label}
              group={entry}
              expanded={expanded}
              isOpen={openGroups[entry.label] ?? false}
              disabled={disabled}
              onToggle={() => toggleGroup(entry.label)}
              currentPath={pathname}
              onNavClick={onClose}
            />
          ) : (
            <NavItemLink
              key={entry.to}
              item={{ ...entry, icon: <entry.icon className={navIcon} /> }}
              expanded={expanded}
              disabled={disabled}
              onClick={onClose}
            />
          ),
        )}
      </div>
    </SidebarShell>
  );
}
