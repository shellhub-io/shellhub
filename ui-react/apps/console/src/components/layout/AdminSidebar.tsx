import { useState } from "react";
import { NavLink, useLocation } from "react-router-dom";
import {
  HomeIcon,
  UsersIcon,
  CpuChipIcon,
  CommandLineIcon,
  ShieldCheckIcon,
  ServerStackIcon,
  MegaphoneIcon,
  Cog6ToothIcon,
  KeyIcon,
  DocumentCheckIcon,
  ChevronDownIcon,
  ChevronLeftIcon,
} from "@heroicons/react/24/outline";
import type { ReactNode } from "react";
import { getConfig } from "../../env";
import { useAdminLicense } from "../../hooks/useAdminLicense";

// ---- Types ----

interface NavItem {
  to: string;
  label: string;
  icon: ReactNode;
}

interface NavGroup {
  label: string;
  icon: ReactNode;
  children: NavItem[];
}

type NavEntry = NavItem | NavGroup;

function isNavGroup(entry: NavEntry): entry is NavGroup {
  return "children" in entry;
}

// ---- Nav definition (matches Vue admin's items array exactly) ----

const coreNavEntries: NavEntry[] = [
  {
    to: "/admin/dashboard",
    label: "Dashboard",
    icon: <HomeIcon className="w-[18px] h-[18px]" />,
  },
  {
    to: "/admin/users",
    label: "Users",
    icon: <UsersIcon className="w-[18px] h-[18px]" />,
  },
  {
    to: "/admin/devices",
    label: "Devices",
    icon: <CpuChipIcon className="w-[18px] h-[18px]" />,
  },
  {
    to: "/admin/sessions",
    label: "Sessions",
    icon: <CommandLineIcon className="w-[18px] h-[18px]" />,
  },
  {
    to: "/admin/firewall-rules",
    label: "Firewall Rules",
    icon: <ShieldCheckIcon className="w-[18px] h-[18px]" />,
  },
  {
    to: "/admin/namespaces",
    label: "Namespaces",
    icon: <ServerStackIcon className="w-[18px] h-[18px]" />,
  },
];

const announcementsEntry: NavEntry = {
  to: "/admin/announcements",
  label: "Announcements",
  icon: <MegaphoneIcon className="w-[18px] h-[18px]" />,
};

const settingsGroup: NavGroup = {
  label: "Settings",
  icon: <Cog6ToothIcon className="w-[18px] h-[18px]" />,
  children: [
    {
      to: "/admin/settings/authentication",
      label: "Authentication",
      icon: <KeyIcon className="w-[18px] h-[18px]" />,
    },
    {
      to: "/admin/license",
      label: "License",
      icon: <DocumentCheckIcon className="w-[18px] h-[18px]" />,
    },
  ],
};

const expiredNavEntries: NavEntry[] = [
  {
    label: "Settings",
    icon: <Cog6ToothIcon className="w-[18px] h-[18px]" />,
    children: [
      {
        to: "/admin/license",
        label: "License",
        icon: <DocumentCheckIcon className="w-[18px] h-[18px]" />,
      },
    ],
  },
];

function buildNavEntries(): NavEntry[] {
  const entries: NavEntry[] = [...coreNavEntries];

  if (getConfig().announcements) {
    entries.push(announcementsEntry);
  }

  entries.push(settingsGroup);
  return entries;
}

// ---- Sub-components ----

function NavItemLink({
  item,
  expanded,
}: {
  item: NavItem;
  expanded: boolean;
}) {
  return (
    <NavLink
      to={item.to}
      title={expanded ? undefined : item.label}
      className={({ isActive }) =>
        `flex items-center gap-3 px-3 py-2 rounded-md text-[13px] font-medium transition-all duration-150 ${
          isActive
            ? "bg-primary/10 text-primary border border-primary/20"
            : "text-text-secondary hover:text-text-primary hover:bg-hover-subtle border border-transparent"
        } ${expanded ? "" : "justify-center"}`}
    >
      {item.icon}
      {expanded ? <span className="truncate">{item.label}</span> : null}
    </NavLink>
  );
}

function NavGroupItem({
  group,
  expanded,
  isOpen,
  onToggle,
  currentPath,
}: {
  group: NavGroup;
  expanded: boolean;
  isOpen: boolean;
  onToggle: () => void;
  currentPath: string;
}) {
  const isChildActive = group.children.some((c) =>
    currentPath.startsWith(c.to),
  );

  return (
    <div>
      <button
        type="button"
        onClick={onToggle}
        title={expanded ? undefined : group.label}
        aria-expanded={isOpen}
        className={`w-full flex items-center gap-3 px-3 py-2 rounded-md text-[13px] font-medium transition-all duration-150 ${
          isChildActive
            ? "text-primary"
            : "text-text-secondary hover:text-text-primary hover:bg-hover-subtle"
        } ${expanded ? "" : "justify-center"}`}
      >
        {group.icon}
        {expanded ? (
          <>
            <span className="flex-1 text-left truncate">{group.label}</span>
            <ChevronDownIcon
              className={`w-3.5 h-3.5 transition-transform duration-200 ${
                isOpen ? "rotate-180" : ""
              }`}
              strokeWidth={2}
            />
          </>
        ) : null}
      </button>
      {expanded && isOpen ? (
        <div className="ml-5 mt-0.5 space-y-0.5 border-l border-border pl-3">
          {group.children.map((child) => (
            <NavLink
              key={child.to}
              to={child.to}
              className={({ isActive }) =>
                `flex items-center gap-2 px-2 py-1.5 rounded-md text-[12px] font-medium transition-all duration-150 ${
                  isActive
                    ? "text-primary bg-primary/5"
                    : "text-text-secondary hover:text-text-primary hover:bg-hover-subtle"
                }`}
            >
              <span className="truncate">{child.label}</span>
            </NavLink>
          ))}
        </div>
      ) : null}
    </div>
  );
}

// ---- Sidebar ----

export default function AdminSidebar({
  expanded,
  onToggle,
}: {
  expanded: boolean;
  onToggle: () => void;
}) {
  const { data: license, isLoading } = useAdminLicense();
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>({});
  const { pathname } = useLocation();

  const isExpired = !isLoading && (!license || license.expired);

  const visibleEntries: NavEntry[] = isExpired
    ? expiredNavEntries
    : buildNavEntries();

  const toggleGroup = (label: string) => {
    setOpenGroups((prev) => ({ ...prev, [label]: !prev[label] }));
  };

  return (
    <aside
      className={`bg-surface border-r border-border flex flex-col min-h-screen shrink-0 transition-all duration-200 ease-in-out overflow-hidden ${
        expanded ? "w-[220px]" : "w-[60px]"
      }`}
    >
      {/* Logo */}
      <div className="h-14 flex items-center justify-center border-b border-border px-3">
        <NavLink to="/admin/dashboard">
          {expanded ? (
            <img src="/logo.svg" alt="ShellHub Admin" className="h-8" />
          ) : (
            <img
              src="/logo-icon.svg"
              alt="ShellHub Admin"
              className="h-6 w-6"
              onError={(e) => {
                (e.target as HTMLImageElement).src = "/logo.svg";
                (e.target as HTMLImageElement).className = "h-6";
              }}
            />
          )}
        </NavLink>
      </div>

      {/* Navigation */}
      <nav
        className="flex-1 px-2 pt-4 pb-2 overflow-y-auto"
        aria-label="Admin navigation"
      >
        <div className="space-y-0.5">
          {visibleEntries.map((entry) =>
            isNavGroup(entry) ? (
              <NavGroupItem
                key={entry.label}
                group={entry}
                expanded={expanded}
                isOpen={openGroups[entry.label] ?? false}
                onToggle={() => toggleGroup(entry.label)}
                currentPath={pathname}
              />
            ) : (
              <NavItemLink
                key={entry.to}
                item={entry}
                expanded={expanded}
              />
            ),
          )}
        </div>
      </nav>

      {/* Footer with toggle */}
      <div className="h-11 px-3 flex items-center justify-between border-t border-border">
        {expanded ? (
          <p className="text-2xs font-mono text-text-muted/60">Admin Panel</p>
        ) : null}
        <button
          type="button"
          onClick={onToggle}
          aria-label={expanded ? "Collapse sidebar" : "Expand sidebar"}
          className="p-1 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-subtle transition-colors"
        >
          <ChevronLeftIcon
            className={`w-3.5 h-3.5 transition-transform duration-200 ${
              expanded ? "" : "rotate-180"
            }`}
            strokeWidth={2}
          />
        </button>
      </div>
    </aside>
  );
}
