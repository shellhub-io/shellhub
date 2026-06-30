import type { ComponentType, SVGProps } from "react";
import {
  AdjustmentsHorizontalIcon,
  BookOpenIcon,
  CodeBracketIcon,
  CodeBracketSquareIcon,
  CpuChipIcon,
  PuzzlePieceIcon,
  RocketLaunchIcon,
  ServerStackIcon,
  ShieldCheckIcon,
  Square3Stack3DIcon,
} from "@heroicons/react/24/outline";

export interface SidebarItem {
  label: string;
  href?: string;
  featured?: boolean;
  items?: SidebarItem[];
}

export interface SidebarSection {
  label: string;
  description: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  items: SidebarItem[];
}

// pages intentionally not in the sidebar nav
export const PAGES_NOT_IN_NAV: string[] = ["/"];

export function flattenItems(
  items: SidebarItem[],
): { label: string; href: string; featured?: boolean }[] {
  return items.flatMap((item) =>
    item.href
      ? [{ label: item.label, href: item.href, featured: item.featured }]
      : flattenItems(item.items ?? []),
  );
}

export const sidebar: SidebarSection[] = [
  {
    label: "Getting Started",
    description:
      "What ShellHub is, how it works, editions, and your first connection.",
    icon: RocketLaunchIcon,
    items: [
      {
        label: "Introduction",
        href: "/getting-started/introduction",
        featured: true,
      },
      {
        label: "How it Works",
        href: "/getting-started/how-it-works",
        featured: true,
      },
      {
        label: "Features",
        href: "/getting-started/features",
        featured: true,
      },
      { label: "Editions", href: "/getting-started/editions" },
      {
        label: "Quick Start",
        href: "/getting-started/quick-start",
        featured: true,
      },
      { label: "FAQ", href: "/getting-started/faq" },
    ],
  },
  {
    label: "Resources",
    description:
      "Core platform resources — devices, sessions, and web endpoints.",
    icon: Square3Stack3DIcon,
    items: [
      {
        label: "Devices",
        href: "/guides/devices",
        featured: true,
      },
      {
        label: "Sessions",
        href: "/guides/sessions",
        featured: true,
      },
      {
        label: "Web Endpoints",
        href: "/guides/web-endpoints",
        featured: true,
      },
    ],
  },
  {
    label: "Security",
    description: "Secure your devices with public keys and firewall rules.",
    icon: ShieldCheckIcon,
    items: [
      {
        label: "Public Keys",
        href: "/guides/public-keys",
        featured: true,
      },
      {
        label: "Firewall Rules",
        href: "/guides/firewall-rules",
        featured: true,
      },
      { label: "Multi-Factor Authentication", href: "/guides/mfa" },
      { label: "Single Sign-On (SAML)", href: "/guides/sso" },
      { label: "Local Authentication", href: "/guides/local-authentication" },
    ],
  },
  {
    label: "Management",
    description:
      "Organize your infrastructure with namespaces, teams, and settings.",
    icon: AdjustmentsHorizontalIcon,
    items: [
      {
        label: "Namespaces",
        href: "/guides/namespaces",
        featured: true,
      },
      {
        label: "Team",
        href: "/guides/team",
        featured: true,
      },
      {
        label: "Settings",
        href: "/guides/settings",
        featured: true,
      },
    ],
  },
  {
    label: "User Guides",
    description:
      "Manage devices, connect via SSH, organize with namespaces and tags.",
    icon: BookOpenIcon,
    items: [
      {
        label: "Connecting via SSH",
        href: "/guides/connecting",
        featured: true,
      },
      {
        label: "File Transfer",
        href: "/guides/file-transfer",
        featured: true,
      },
      {
        label: "Port Forwarding",
        href: "/guides/port-forwarding",
        featured: true,
      },
      {
        label: "Container Access",
        href: "/guides/container-access",
        featured: true,
      },
      { label: "Session Recording", href: "/guides/session-recording" },
      { label: "Organizing with Tags", href: "/guides/tags" },
    ],
  },
  {
    label: "Agent",
    description: "Install the ShellHub agent on Docker, Snap, FreeBSD, or WSL.",
    icon: CpuChipIcon,
    items: [
      {
        label: "Agent Overview",
        href: "/agent/overview",
        featured: true,
      },
      { label: "Install Script", href: "/agent/install-script" },
      {
        label: "Platform Guides",
        items: [
          {
            label: "Docker",
            href: "/agent/docker",
            featured: true,
          },
          {
            label: "Snap",
            href: "/agent/snap",
            featured: true,
          },
          {
            label: "FreeBSD",
            href: "/agent/freebsd",
            featured: true,
          },
          { label: "WSL", href: "/agent/wsl" },
          {
            label: "Building from Source",
            href: "/agent/building-from-source",
          },
        ],
      },
    ],
  },
  {
    label: "Embedded Linux",
    description:
      "Embed the agent on Raspberry Pi, Buildroot, and Yocto images.",
    icon: CodeBracketSquareIcon,
    items: [
      {
        label: "Embedded Linux Overview",
        href: "/embedded-linux/overview",
        featured: true,
      },
      {
        label: "Raspberry Pi",
        href: "/embedded-linux/raspberry-pi",
        featured: true,
      },
      {
        label: "Buildroot",
        href: "/embedded-linux/buildroot",
        featured: true,
      },
      {
        label: "Yocto Project",
        href: "/embedded-linux/yocto",
        featured: true,
      },
    ],
  },
  {
    label: "Integrations",
    description:
      "Connect ShellHub with Ansible, Terraform, CI/CD, and VS Code.",
    icon: PuzzlePieceIcon,
    items: [
      { label: "Integrations Overview", href: "/integration/overview" },
      {
        label: "Ansible",
        href: "/integration/ansible",
        featured: true,
      },
      {
        label: "Terraform",
        href: "/integration/terraform",
        featured: true,
      },
      {
        label: "CI/CD Pipelines",
        href: "/integration/ci-cd",
        featured: true,
      },
      {
        label: "VS Code Remote",
        href: "/integration/vscode-remote",
        featured: true,
      },
    ],
  },
  {
    label: "Self-Hosted",
    description: "Deploy, configure, and maintain your own ShellHub instance.",
    icon: ServerStackIcon,
    items: [
      {
        label: "Setup",
        href: "/self-hosted/deploying",
        featured: true,
      },
      {
        label: "Configuration",
        href: "/self-hosted/configuring",
        featured: true,
      },
      {
        label: "Administration",
        href: "/self-hosted/administration",
        featured: true,
      },
      {
        label: "Upgrade",
        href: "/self-hosted/upgrading",
        featured: true,
      },
      {
        label: "Troubleshooting",
        href: "/self-hosted/troubleshooting",
      },
    ],
  },
  {
    label: "Developers",
    description:
      "Set up a dev environment, use the API, and build the agent from source.",
    icon: CodeBracketIcon,
    items: [
      {
        label: "Dev Environment",
        href: "/developers/development-environment",
        featured: true,
      },
      {
        label: "API Reference",
        href: "/developers/api-reference",
        featured: true,
      },
      {
        label: "Agent Development",
        href: "/developers/agent-development",
        featured: true,
      },
    ],
  },
];
