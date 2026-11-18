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
import { useNamespaceLegacy } from "@/hooks/useNamespaceLegacy";
import { useHasPermission } from "@/hooks/useHasPermission";

/**
 * A navigation icon as a component, so the sidebar and the command palette can each draw it at
 * their own size.
 */
export type NavIcon = typeof HomeIcon;

/**
 * One navigation link, as the sidebar, the admin page picker and the command palette list it.
 * premium marks a page that needs a paid edition, beta one still in preview.
 */
export interface NavItem {
  to: string;
  label: string;
  icon: NavIcon;
  premium?: boolean;
  beta?: boolean;
}

/**
 * A group of sidebar links, under a heading when it has a title. A group of one link goes without,
 * since a heading over a single link names nothing the link does not. The sidebar keys each group
 * by its first link, so no link may open two groups.
 */
export interface NavSection<T extends { to: string } = NavItem> {
  title?: string;
  items: T[];
}

/**
 * The title of the section that links to route, or undefined when that section has no title or no
 * section links there.
 */
export function navSectionTitle(
  sections: NavSection<{ to: string }>[],
  route: string,
): string | undefined {
  return sections.find((section) =>
    section.items.some((item) => item.to === route),
  )?.title;
}

/**
 * Every page the sections lead to, in sidebar order.
 */
export function sectionLinks(sections: NavSection[]): NavItem[] {
  return sections.flatMap((section) => section.items);
}

function buildSections(
  isNewNamespace: boolean,
  canListWebEndpoints: boolean,
): NavSection[] {
  const config = getConfig();

  const resources: NavItem[] = [
    {
      to: "/devices",
      label: "Devices",
      icon: CpuChipIcon,
    },
    {
      to: "/containers",
      label: "Containers",
      icon: CubeIcon,
    },
  ];

  if (!isNewNamespace) {
    resources.push({
      to: "/sessions",
      label: "Sessions",
      icon: CommandLineIcon,
    });
  }

  if (config.webEndpoints && isEnterpriseOrCloud() && canListWebEndpoints) {
    resources.push({
      to: "/web-endpoints",
      label: "Web Endpoints",
      icon: GlobeAltIcon,
      beta: true,
    });
  }

  const security: NavItem[] = isNewNamespace
    ? [
        {
          to: "/sessions",
          label: "Sessions",
          icon: CommandLineIcon,
        },
        {
          to: "/access-policies",
          label: "Access Policies",
          icon: ShieldCheckIcon,
        },
        {
          to: "/ssh-identities",
          label: "SSH Identities",
          icon: FingerPrintIcon,
        },
      ]
    : [
        {
          to: "/sshkeys/public-keys",
          label: "Public Keys",
          icon: KeyIcon,
        },
        {
          to: "/firewall-rules",
          label: "Firewall Rules",
          icon: ShieldExclamationIcon,
          premium: true,
        },
        {
          to: "/secure-vault",
          label: "Secure Vault",
          icon: LockClosedIcon,
        },
      ];

  return [
    {
      items: [
        {
          to: "/dashboard",
          label: "Dashboard",
          icon: HomeIcon,
        },
      ],
    },
    {
      title: "Resources",
      items: resources,
    },
    {
      title: "SSH",
      items: security,
    },
    {
      title: "Management",
      items: [
        {
          to: "/team",
          label: "Team",
          icon: UsersIcon,
        },
        {
          to: "/settings",
          label: "Settings",
          icon: Cog6ToothIcon,
        },
      ],
    },
  ];
}

/**
 * The navigation sections for the current namespace. Whether it is legacy decides which security
 * pages exist, and the current role decides whether web endpoints are offered, so the result
 * depends on the namespace settings and the role.
 */
export function useNavSections(): NavSection[] {
  const isNewNamespace = useNamespaceLegacy() !== true;
  const canListWebEndpoints = useHasPermission("DeviceConnect");
  return buildSections(isNewNamespace, canListWebEndpoints);
}

/**
 * navSectionTitle over the current namespace's sidebar, whose sections follow whether the namespace is legacy.
 */
export function useNavSectionTitle(route: string): string | undefined {
  return navSectionTitle(useNavSections(), route);
}
