import type { ReactNode } from "react";
import {
  HomeIcon,
  KeyIcon,
  Cog6ToothIcon,
  UsersIcon,
  CpuChipIcon,
  CommandLineIcon,
  LockClosedIcon,
  CubeIcon,
  GlobeAltIcon,
  ShieldExclamationIcon,
  ShieldCheckIcon,
  FingerPrintIcon,
} from "@heroicons/react/24/outline";
import { getConfig, isEnterpriseOrCloud } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace } from "@/hooks/useNamespaces";
import { navIcon } from "./SidebarShell";

/**
 * One sidebar link. premium marks a page that needs a paid edition, beta one still in preview.
 */
export interface NavItem {
  to: string;
  label: string;
  icon: ReactNode;
  premium?: boolean;
  beta?: boolean;
}

/**
 * A group of sidebar links, under a heading when it has a title. A group of one link goes without,
 * since a heading over a single link names nothing the link does not. The sidebar keys each group
 * by its first link, so no link may open two groups.
 */
export interface NavSection {
  title?: string;
  items: NavItem[];
}

function buildSections(isIdentityMode: boolean): NavSection[] {
  const config = getConfig();

  const resources: NavItem[] = [
    {
      to: "/devices",
      label: "Devices",
      icon: <CpuChipIcon className={navIcon} />,
    },
    {
      to: "/containers",
      label: "Containers",
      icon: <CubeIcon className={navIcon} />,
    },
  ];

  if (!isIdentityMode) {
    resources.push({
      to: "/sessions",
      label: "Sessions",
      icon: <CommandLineIcon className={navIcon} />,
    });
  }

  if (config.webEndpoints && isEnterpriseOrCloud()) {
    resources.push({
      to: "/web-endpoints",
      label: "Web Endpoints",
      icon: <GlobeAltIcon className={navIcon} />,
      beta: true,
    });
  }

  const security: NavItem[] = isIdentityMode
    ? [
        {
          to: "/sessions",
          label: "Sessions",
          icon: <CommandLineIcon className={navIcon} />,
        },
        {
          to: "/access-policies",
          label: "Access Policies",
          icon: <ShieldCheckIcon className={navIcon} />,
        },
        {
          to: "/ssh-identities",
          label: "SSH Identities",
          icon: <FingerPrintIcon className={navIcon} />,
        },
      ]
    : [
        {
          to: "/sshkeys/public-keys",
          label: "Public Keys",
          icon: <KeyIcon className={navIcon} />,
        },
        {
          to: "/firewall-rules",
          label: "Firewall Rules",
          icon: <ShieldExclamationIcon className={navIcon} />,
          premium: true,
        },
        {
          to: "/secure-vault",
          label: "Secure Vault",
          icon: <LockClosedIcon className={navIcon} />,
        },
      ];

  return [
    {
      items: [
        {
          to: "/dashboard",
          label: "Dashboard",
          icon: <HomeIcon className={navIcon} />,
        },
      ],
    },
    {
      title: "Resources",
      items: resources,
    },
    {
      title: isIdentityMode ? "SSH" : "Security",
      items: security,
    },
    {
      title: "Management",
      items: [
        {
          to: "/team",
          label: "Team",
          icon: <UsersIcon className={navIcon} />,
        },
        {
          to: "/settings",
          label: "Settings",
          icon: <Cog6ToothIcon className={navIcon} />,
        },
      ],
    },
  ];
}

/**
 * The navigation sections for the current namespace. The SSH access mode decides which security
 * pages exist, so the result depends on the namespace settings.
 */
export function useNavSections(): NavSection[] {
  const { tenant } = useAuthStore();
  const { namespace } = useNamespace(tenant ?? "");
  return buildSections(namespace?.settings?.ssh_access_mode === "identity");
}
