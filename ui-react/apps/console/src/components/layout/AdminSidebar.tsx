import { useState, useMemo } from "react";
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
import { useAuthStore } from "../../stores/authStore";

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

const navIcon = "w-[18px] h-[18px]";

const coreNavEntries: NavEntry[] = [
  { to: "/admin/dashboard", label: "Dashboard", icon: <HomeIcon className={navIcon} /> },
  { to: "/admin/users", label: "Users", icon: <UsersIcon className={navIcon} /> },
  { to: "/admin/devices", label: "Devices", icon: <CpuChipIcon className={navIcon} /> },
  { to: "/admin/sessions", label: "Sessions", icon: <CommandLineIcon className={navIcon} /> },
  { to: "/admin/firewall-rules", label: "Firewall Rules", icon: <ShieldCheckIcon className={navIcon} /> },
  { to: "/admin/namespaces", label: "Namespaces", icon: <ServerStackIcon className={navIcon} /> },
];

const announcementsEntry: NavEntry = {
  to: "/admin/announcements",
  label: "Announcements",
  icon: <MegaphoneIcon className={navIcon} />,
};

const settingsGroup: NavGroup = {
  label: "Settings",
  icon: <Cog6ToothIcon className={navIcon} />,
  children: [
    { to: "/admin/settings/authentication", label: "Authentication", icon: <KeyIcon className={navIcon} /> },
    { to: "/admin/license", label: "License", icon: <DocumentCheckIcon className={navIcon} /> },
  ],
};

const expiredNavEntries: NavEntry[] = [
  {
    label: "Settings",
    icon: <Cog6ToothIcon className={navIcon} />,
    children: [
      { to: "/admin/license", label: "License", icon: <DocumentCheckIcon className={navIcon} /> },
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

// ---- Shared styles ----

const navBase = "flex items-center gap-3 px-3 py-2 rounded-md text-[13px] font-medium";
const navActive = "bg-primary/10 text-primary border border-primary/20";
const navIdle = "text-text-secondary hover:text-text-primary hover:bg-hover-subtle border border-transparent";
const navDisabled = "text-text-muted/50 cursor-not-allowed";

// ---- Sub-components ----

function NavItemLink({
  item,
  expanded,
  disabled,
  onClick,
}: {
  item: NavItem;
  expanded: boolean;
  disabled?: boolean;
  onClick?: () => void;
}) {
  const align = expanded ? "" : "justify-center";
  const label = expanded ? <span className="truncate">{item.label}</span> : null;

  if (disabled) {
    return (
      <span aria-disabled="true" className={`${navBase} ${navDisabled} ${align}`}>
        {item.icon}
        {label}
      </span>
    );
  }

  return (
    <NavLink
      to={item.to}
      title={expanded ? undefined : item.label}
      onClick={onClick}
      className={({ isActive }) =>
        `${navBase} transition-all duration-150 ${isActive ? navActive : navIdle} ${align}`}
    >
      {item.icon}
      {label}
    </NavLink>
  );
}

function NavGroupItem({
  group,
  expanded,
  isOpen,
  disabled,
  onToggle,
  currentPath,
  onNavClick,
}: {
  group: NavGroup;
  expanded: boolean;
  isOpen: boolean;
  disabled?: boolean;
  onToggle: () => void;
  currentPath: string;
  onNavClick?: () => void;
}) {
  const isChildActive = !disabled && group.children.some((c) =>
    currentPath.startsWith(c.to),
  );
  const align = expanded ? "" : "justify-center";

  return (
    <div>
      <button
        type="button"
        onClick={disabled ? undefined : onToggle}
        title={expanded ? undefined : group.label}
        aria-expanded={disabled ? undefined : isOpen}
        aria-disabled={disabled || undefined}
        className={`w-full ${navBase} transition-all duration-150 ${
          disabled
            ? navDisabled
            : isChildActive
              ? "text-primary"
              : "text-text-secondary hover:text-text-primary hover:bg-hover-subtle"
        } ${align}`}
      >
        {group.icon}
        {expanded ? (
          <>
            <span className="flex-1 text-left truncate">{group.label}</span>
            {!disabled && (
              <ChevronDownIcon
                className={`w-3.5 h-3.5 transition-transform duration-200 ${
                  isOpen ? "rotate-180" : ""
                }`}
                strokeWidth={2}
              />
            )}
          </>
        ) : null}
      </button>
      {!disabled && expanded && isOpen ? (
        <div className="ml-5 mt-0.5 space-y-0.5 border-l border-border pl-3">
          {group.children.map((child) => (
            <NavLink
              key={child.to}
              to={child.to}
              onClick={onNavClick}
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
  pinned,
  onToggle,
  onClose,
}: {
  expanded: boolean;
  pinned: boolean;
  onToggle: () => void;
  onClose?: () => void;
}) {
  const { data: license, isLoading } = useAdminLicense();
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>({});
  const { pathname } = useLocation();

  const isExpired = !isLoading && (!license || license.expired);
  const showRestrictedNav = !isAdmin || isLoading || isExpired;
  const isDisabled = !isAdmin;

  const fullNavEntries = useMemo(() => buildNavEntries(), []);

  const visibleEntries: NavEntry[] = showRestrictedNav
    ? expiredNavEntries
    : fullNavEntries;

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
        <NavLink to="/admin/dashboard" onClick={onClose} className="relative flex items-center justify-center">
          <img
            src="/logo.svg"
            alt="ShellHub Admin"
            className={`h-8 transition-opacity duration-200 ${expanded ? "opacity-100" : "opacity-0 absolute"}`}
          />
          <img
            src="/cloud-icon.svg"
            alt="ShellHub Admin"
            className={`h-6 w-6 transition-opacity duration-200 ${expanded ? "opacity-0 absolute" : "opacity-100"}`}
          />
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
                disabled={isDisabled}
                onToggle={() => toggleGroup(entry.label)}
                currentPath={pathname}
                onNavClick={onClose}
              />
            ) : (
              <NavItemLink
                key={entry.to}
                item={entry}
                expanded={expanded}
                disabled={isDisabled}
                onClick={onClose}
              />
            ),
          )}
        </div>
      </nav>

      {/* Footer with pin toggle */}
      <div className={`h-11 px-3 flex items-center justify-between transition-colors duration-200 ${expanded ? "border-t border-border" : "border-t border-transparent"}`}>
        <p className={`text-2xs font-mono text-text-muted/60 whitespace-nowrap transition-opacity duration-200 ${expanded ? "opacity-100" : "opacity-0"}`}>
          Admin Panel
        </p>
        <button
          type="button"
          onClick={onToggle}
          tabIndex={expanded ? 0 : -1}
          aria-label={pinned ? "Unpin sidebar" : "Pin sidebar"}
          title={pinned ? "Unpin sidebar" : "Pin sidebar open"}
          className={`p-1 rounded-md transition-all duration-200 ${
            expanded ? "opacity-100" : "opacity-0"
          } ${
            pinned
              ? "text-primary bg-primary/10"
              : "text-text-muted hover:text-text-primary hover:bg-hover-subtle"
          }`}
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
