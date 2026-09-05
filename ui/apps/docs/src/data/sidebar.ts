import type { ComponentType, SVGProps } from "react";
import {
  AdjustmentsHorizontalIcon,
  BookOpenIcon,
  CommandLineIcon,
  LifebuoyIcon,
  PuzzlePieceIcon,
  RocketLaunchIcon,
  ServerStackIcon,
} from "@heroicons/react/24/outline";

/**
 * One entry in the docs navigation. An item with no href is a group heading and carries its
 * children in items; featured promotes it onto the section landing page.
 */
export interface SidebarItem {
  label: string;
  href?: string;
  featured?: boolean;
  items?: SidebarItem[];
}

/**
 * A top-level section of the docs navigation. description and icon are rendered on the section
 * index, not in the sidebar itself.
 */
export interface SidebarSection {
  label: string;
  description: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  items: SidebarItem[];
}

/**
 * Routes that exist but are deliberately absent from the navigation, so a link check can tell an
 * unreachable page from an intentionally unlisted one.
 *
 * Everything under the second block is the pre-rewrite tree. This app is a throwaway spike
 * comparing documentation structures, and the sandbox running it cannot delete files, so the
 * superseded pages stay on disk and are retired from navigation here instead. Their content lives
 * on, reorganized, under the journey-shaped tree the sidebar actually exposes.
 */
export const PAGES_NOT_IN_NAV: string[] = [
  "/",
  // Served when nothing matches, so it is reached by mistake rather than by navigation.
  "/404",
  // The API reference is the specification rendered full-window, outside this layout. It is
  // reached from the REST API page rather than listed beside pages that read like pages.
  "/api/reference",
];

/**
 * Flattens a nested item tree to the entries that actually resolve to a page, dropping the group
 * headings on the way down. Used wherever the tree has to be walked as a list — search, the
 * previous/next links, the link check.
 */
export function flattenItems(
  items: SidebarItem[],
): { label: string; href: string; featured?: boolean }[] {
  return items.flatMap((item) =>
    item.href
      ? [{ label: item.label, href: item.href, featured: item.featured }]
      : flattenItems(item.items ?? []),
  );
}

/**
 * The docs navigation, in render order. This is the single source of it: the sidebar, the section
 * index pages and the previous/next links are all derived from this array, so a page added here
 * appears in all three and a page missing from it appears in none.
 */
export const sidebar: SidebarSection[] = [
  {
    label: "About ShellHub",
    description:
      "What ShellHub is, how it reaches your devices, and which deployment fits you.",
    icon: BookOpenIcon,
    items: [
      {
        label: "Introduction",
        href: "/introduction",
        featured: true,
      },
      {
        label: "How ShellHub Works",
        href: "/about/how-shellhub-works",
        featured: true,
      },
      {
        label: "ShellHub vs. Jump Hosts and VPNs",
        href: "/about/shellhub-vs-jump-hosts",
      },
      {
        label: "Self-Hosted vs. Cloud",
        href: "/about/self-hosted-vs-cloud",
        featured: true,
      },
      { label: "Editions", href: "/about/editions" },
      {
        label: "Ports and Connectivity",
        href: "/about/ports-and-connectivity",
      },
      { label: "FAQ", href: "/about/faq" },
    ],
  },
  {
    label: "Get Started",
    description:
      "Install the agent on a device and see it appear in your namespace.",
    icon: RocketLaunchIcon,
    items: [
      {
        label: "Quickstart",
        href: "/get-started",
        featured: true,
      },
      {
        label: "Install the Agent",
        items: [
          {
            label: "Choosing a Method",
            href: "/get-started/install",
            featured: true,
          },
          { label: "Docker", href: "/get-started/install/docker" },
          { label: "Podman", href: "/get-started/install/podman" },
          { label: "Snap", href: "/get-started/install/snap" },
          {
            label: "Standalone Binary",
            href: "/get-started/install/standalone",
          },
          { label: "WSL", href: "/get-started/install/wsl" },
          { label: "FreeBSD", href: "/get-started/install/freebsd" },
          {
            label: "Raspberry Pi",
            href: "/get-started/install/raspberry-pi",
          },
          { label: "Buildroot", href: "/get-started/install/buildroot" },
          { label: "Yocto Project", href: "/get-started/install/yocto" },
          {
            label: "Building from Source",
            href: "/get-started/install/build-from-source",
          },
        ],
      },
      {
        label: "Enrolling Devices",
        href: "/get-started/enrolling-devices",
        featured: true,
      },
      {
        label: "Agent Configuration",
        href: "/get-started/agent-configuration",
      },
    ],
  },
  {
    label: "Connect",
    description:
      "Open a terminal, move files, forward a port, or reach a service on the device.",
    icon: CommandLineIcon,
    items: [
      {
        label: "Connecting to a Device",
        href: "/connect",
        featured: true,
      },
      { label: "SSH Client Configuration", href: "/connect/ssh-clients" },
      {
        label: "File Transfer",
        href: "/connect/file-transfer",
        featured: true,
      },
      {
        label: "Port Forwarding",
        href: "/connect/port-forwarding",
        featured: true,
      },
      {
        label: "Container Access",
        href: "/connect/container-access",
        featured: true,
      },
      { label: "Web Endpoints", href: "/connect/web-endpoints" },
    ],
  },
  {
    label: "Manage",
    description:
      "Run a fleet: devices, who may reach them, your team, and what was done.",
    icon: AdjustmentsHorizontalIcon,
    items: [
      {
        label: "Devices",
        items: [
          {
            label: "Managing Devices",
            href: "/manage/devices",
            featured: true,
          },
          { label: "Install Keys", href: "/manage/devices/install-keys" },
          { label: "Tags", href: "/manage/devices/tags" },
          {
            label: "Device Not Appearing",
            href: "/manage/devices/device-not-appearing",
          },
        ],
      },
      {
        label: "Access Control",
        items: [
          {
            label: "How Authorization Works",
            href: "/manage/access-control",
            featured: true,
          },
          {
            label: "SSH Identities",
            href: "/manage/access-control/ssh-identities",
            featured: true,
          },
          {
            label: "Access Policies",
            href: "/manage/access-control/access-policies",
            featured: true,
          },
          {
            label: "Policy Reference",
            href: "/manage/access-control/policy-reference",
          },
          {
            label: "Service Accounts",
            href: "/manage/access-control/service-accounts",
          },
          {
            label: "Login Approvals",
            href: "/manage/access-control/login-approvals",
          },
          {
            label: "Connection Denied",
            href: "/manage/access-control/connection-denied",
          },
          {
            label: "Migrating from the Legacy Model",
            href: "/manage/access-control/migrating-from-legacy",
          },
          {
            label: "Public Keys",
            href: "/manage/access-control/public-keys",
          },
          {
            label: "Firewall Rules",
            href: "/manage/access-control/firewall-rules",
          },
        ],
      },
      {
        label: "Team",
        items: [
          {
            label: "Members",
            href: "/manage/team",
            featured: true,
          },
          { label: "User Roles", href: "/manage/team/user-roles" },
          {
            label: "Single Sign-On (SAML)",
            href: "/manage/team/single-sign-on",
          },
          {
            label: "SAML Sign-In Fails",
            href: "/manage/team/sso-not-working",
          },
          {
            label: "Multi-Factor Authentication",
            href: "/manage/team/multi-factor-authentication",
          },
        ],
      },
      {
        label: "Namespaces",
        href: "/manage/namespaces",
        featured: true,
      },
      {
        label: "Sessions",
        href: "/manage/sessions",
        featured: true,
      },
      { label: "Session Recording", href: "/manage/session-recording" },
      { label: "Settings", href: "/manage/settings" },
    ],
  },
  {
    label: "Self-Hosted",
    description: "Run your own ShellHub server.",
    icon: ServerStackIcon,
    items: [
      {
        label: "Self-Hosting Quickstart",
        href: "/selfhosted/quickstart",
        featured: true,
      },
      {
        label: "Environment Variables",
        href: "/selfhosted/environment-variables",
        featured: true,
      },
      { label: "Administration", href: "/selfhosted/administration" },
      { label: "Backup", href: "/selfhosted/backup" },
      { label: "Upgrade", href: "/selfhosted/upgrade" },
      { label: "Troubleshooting", href: "/selfhosted/troubleshooting" },
    ],
  },
  {
    label: "API and Automation",
    description:
      "Drive ShellHub from a script, a pipeline, a config tool, or an AI agent.",
    icon: PuzzlePieceIcon,
    items: [
      {
        label: "REST API",
        href: "/api",
        featured: true,
      },
      { label: "Install Key Webhook", href: "/api/install-key-webhook" },
      {
        label: "MCP Server",
        href: "/api/mcp",
        featured: true,
      },
      { label: "Ansible", href: "/api/ansible" },
      { label: "Terraform", href: "/api/terraform" },
      { label: "CI/CD Pipelines", href: "/api/ci-cd" },
      { label: "Debugging a CI Run", href: "/api/ci-debug" },
    ],
  },
  {
    label: "Help",
    description: "Find the guide for what is not working, or ask.",
    icon: LifebuoyIcon,
    items: [
      { label: "Troubleshooting", href: "/help/troubleshooting", featured: true },
      { label: "Getting Support", href: "/help/support", featured: true },
    ],
  },
];
