import {
  UsersIcon,
  ServerStackIcon,
  ServerIcon,
  KeyIcon,
  LockClosedIcon,
  DocumentCheckIcon,
  ChartBarIcon,
} from "@heroicons/react/24/outline";
import { features } from "@/edition";
import { useAdminLicense } from "@/hooks/useAdminLicense";
import { useAuthStore } from "@/stores/authStore";
import {
  ADMIN_AUTHENTICATION_PATH,
  ADMIN_INSTANCE_PATH,
  ADMIN_TELEMETRY_PATH,
} from "@/utils/adminRoute";
import { navSectionTitle, type NavItem, type NavSection } from "./navSections";

const licenseLink: NavItem = {
  to: "/admin/license",
  label: "License",
  icon: DocumentCheckIcon,
};

const authenticationLink: NavItem = {
  to: ADMIN_AUTHENTICATION_PATH,
  label: "Authentication",
  icon: KeyIcon,
};

const instanceApiKeysLink: NavItem = {
  to: "/admin/instance-api-keys",
  label: "Instance API Keys",
  icon: LockClosedIcon,
};

const telemetryLink: NavItem = {
  to: ADMIN_TELEMETRY_PATH,
  label: "Telemetry",
  icon: ChartBarIcon,
};

const licenseOnlySections: NavSection[] = [{ items: [licenseLink] }];

function buildFullSections(): NavSection[] {
  const instance = features().instanceLicense
    ? [authenticationLink, instanceApiKeysLink, telemetryLink, licenseLink]
    : [instanceApiKeysLink, telemetryLink];

  return [
    {
      items: [{ to: ADMIN_INSTANCE_PATH, label: "Instance", icon: ServerIcon }],
    },
    {
      title: "Accounts",
      items: [
        { to: "/admin/users", label: "Users", icon: UsersIcon },
        { to: "/admin/namespaces", label: "Namespaces", icon: ServerStackIcon },
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
 * The title of the admin section that links to route, read from the full navigation so the
 * License page keeps its section while the sidebar is narrowed to it.
 */
export function adminNavSectionTitle(route: string): string | undefined {
  return navSectionTitle(buildFullSections(), route);
}
