import {
  HomeIcon,
  UsersIcon,
  CpuChipIcon,
  CommandLineIcon,
  ShieldCheckIcon,
  ServerStackIcon,
  MegaphoneIcon,
  KeyIcon,
  LockClosedIcon,
  DocumentCheckIcon,
} from "@heroicons/react/24/outline";
import { getConfig, isCloud } from "@/env";
import { useAdminLicense } from "@/hooks/useAdminLicense";
import { useAuthStore } from "@/stores/authStore";
import { navSectionTitle, type NavSection } from "./navSections";

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
 * A group of admin pages, following the console sidebar's rules for headings and keys.
 */
export type AdminNavSection = NavSection<AdminNavLink>;

const licenseLink: AdminNavLink = {
  to: "/admin/license",
  label: "License",
  icon: DocumentCheckIcon,
};

const licenseOnlySections: AdminNavSection[] = [{ items: [licenseLink] }];

function buildFullSections(): AdminNavSection[] {
  const instance: AdminNavLink[] = [];
  if (getConfig().announcements) {
    instance.push({
      to: "/admin/announcements",
      label: "Announcements",
      icon: MegaphoneIcon,
    });
  }
  instance.push(
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
  );
  if (!isCloud()) instance.push(licenseLink);

  return [
    {
      items: [{ to: "/admin/dashboard", label: "Dashboard", icon: HomeIcon }],
    },
    {
      title: "Accounts",
      items: [
        { to: "/admin/users", label: "Users", icon: UsersIcon },
        { to: "/admin/namespaces", label: "Namespaces", icon: ServerStackIcon },
      ],
    },
    {
      title: "Resources",
      items: [
        { to: "/admin/devices", label: "Devices", icon: CpuChipIcon },
        { to: "/admin/sessions", label: "Sessions", icon: CommandLineIcon },
        {
          to: "/admin/firewall-rules",
          label: "Firewall Rules",
          icon: ShieldCheckIcon,
        },
      ],
    },
    { title: "Instance", items: instance },
  ];
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
    sections: restricted ? licenseOnlySections : buildFullSections(),
    disabled: !isAdmin,
  };
}

/**
 * Every page the sections lead to, in sidebar order, for the command palette.
 */
export function adminNavLinks(sections: AdminNavSection[]): AdminNavLink[] {
  return sections.flatMap((section) => section.items);
}

/**
 * The title of the admin section that links to route, read from the full navigation so the
 * License page keeps its section while the sidebar is narrowed to it.
 */
export function adminNavSectionTitle(route: string): string | undefined {
  return navSectionTitle(buildFullSections(), route);
}
