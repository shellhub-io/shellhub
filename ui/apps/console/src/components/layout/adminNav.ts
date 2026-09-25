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
  LockClosedIcon,
  DocumentCheckIcon,
} from "@heroicons/react/24/outline";
import { getConfig, isCloud } from "@/env";
import { useAdminLicense } from "@/hooks/useAdminLicense";
import { useAuthStore } from "@/stores/authStore";

type Icon = typeof HomeIcon;

/**
 * One admin page. The icon is a component rather than an element, so the sidebar and the command
 * palette can each draw it at their own size.
 */
export interface AdminNavLink {
  to: string;
  label: string;
  icon: Icon;
}

/**
 * Admin pages under a heading of their own, which has no route: the sidebar folds it open, and
 * adminNavLinks lays its pages out flat for the palette.
 */
export interface AdminNavGroup {
  label: string;
  icon: Icon;
  children: AdminNavLink[];
}

/**
 * A top-level entry of the admin navigation, told apart by isAdminNavGroup.
 */
export type AdminNavEntry = AdminNavLink | AdminNavGroup;

/**
 * Whether an entry is a group, the only kind with children and the only kind without a route.
 */
export function isAdminNavGroup(entry: AdminNavEntry): entry is AdminNavGroup {
  return "children" in entry;
}

const licenseLink: AdminNavLink = {
  to: "/admin/license",
  label: "License",
  icon: DocumentCheckIcon,
};

const licenseOnlyEntries: AdminNavEntry[] = [
  { label: "Settings", icon: Cog6ToothIcon, children: [licenseLink] },
];

function buildFullEntries(): AdminNavEntry[] {
  const entries: AdminNavEntry[] = [
    { to: "/admin/dashboard", label: "Dashboard", icon: HomeIcon },
    { to: "/admin/users", label: "Users", icon: UsersIcon },
    { to: "/admin/devices", label: "Devices", icon: CpuChipIcon },
    { to: "/admin/sessions", label: "Sessions", icon: CommandLineIcon },
    {
      to: "/admin/firewall-rules",
      label: "Firewall Rules",
      icon: ShieldCheckIcon,
    },
    { to: "/admin/namespaces", label: "Namespaces", icon: ServerStackIcon },
  ];
  if (getConfig().announcements) {
    entries.push({
      to: "/admin/announcements",
      label: "Announcements",
      icon: MegaphoneIcon,
    });
  }
  entries.push({
    label: "Settings",
    icon: Cog6ToothIcon,
    children: [
      {
        to: "/admin/settings/authentication",
        label: "Authentication",
        icon: KeyIcon,
      },
      {
        to: "/admin/instance-api-keys",
        label: "Instance API Keys",
        icon: LockClosedIcon,
      },
      ...(isCloud() ? [] : [licenseLink]),
    ],
  });
  return entries;
}

/**
 * The admin navigation as the current user may use it. Until the licence is known to be valid,
 * and for anyone who is not an instance admin, it narrows to the licence page; disabled then says
 * whether even that is off limits. active false skips the licence lookup, for a caller that
 * renders outside the admin console, where the licence endpoint may not exist.
 */
export function useAdminNav({ active = true }: { active?: boolean } = {}) {
  const { isLoading, isExpired } = useAdminLicense({ active });
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const restricted = !isAdmin || isLoading || isExpired;
  return {
    entries: restricted ? licenseOnlyEntries : buildFullEntries(),
    disabled: !isAdmin,
  };
}

/**
 * Every page the entries lead to, each group replaced by its pages, in sidebar order. Headings
 * drop out, since they lead nowhere.
 */
export function adminNavLinks(entries: AdminNavEntry[]): AdminNavLink[] {
  return entries.flatMap((entry) =>
    isAdminNavGroup(entry) ? entry.children : [entry],
  );
}
