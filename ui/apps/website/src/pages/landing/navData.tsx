import {
  BookOpenIcon,
  CodeBracketIcon,
  CommandLineIcon,
  ComputerDesktopIcon,
  CpuChipIcon,
  CubeIcon,
  DocumentTextIcon,
  FireIcon,
  HomeModernIcon,
  KeyIcon,
  PencilIcon,
  PencilSquareIcon,
  PlayCircleIcon,
  RectangleGroupIcon,
  RocketLaunchIcon,
  ServerStackIcon,
  ShieldCheckIcon,
  SignalIcon,
  UsersIcon,
  VideoCameraIcon,
} from "@heroicons/react/24/outline";
import { DiscordIcon, GithubIcon } from "@shellhub/design-system/primitives";
import { C } from "./constants";
import { docsUrl, githubUrl } from "@/links";

const ICON_CLASS = "w-4 h-4";

export interface MenuItem {
  label: string;
  href: string;
  desc: string;
  icon: React.ReactNode;
  badge?: string;
}

export interface MenuSection {
  title: string;
  items: MenuItem[];
}

export const productCols: MenuSection[] = [
  {
    title: "Access & Security",
    items: [
      {
        label: "SSH Gateway",
        href: "/features",
        desc: "Centralized SSH access via reverse tunnel — no VPN",
        icon: (
          <ComputerDesktopIcon
            className={ICON_CLASS}
            style={{ color: C.primary }}
          />
        ),
      },
      {
        label: "MFA & RBAC",
        href: "/features",
        desc: "Multi-factor auth and role-based access control",
        icon: <ShieldCheckIcon className={ICON_CLASS} style={{ color: C.blue }} />,
      },
      {
        label: "Firewall Rules",
        href: "/features",
        desc: "IP filtering, port restrictions and access policies",
        icon: <FireIcon className={ICON_CLASS} style={{ color: C.red }} />,
      },
      {
        label: "Public Keys",
        href: "/features",
        desc: "Centralized SSH key management across your fleet",
        icon: <KeyIcon className={ICON_CLASS} style={{ color: C.cyan }} />,
      },
    ],
  },
  {
    title: "Device Management",
    items: [
      {
        label: "Fleet Manager",
        href: "/features",
        desc: "Monitor and manage thousands of devices at once",
        icon: (
          <ServerStackIcon className={ICON_CLASS} style={{ color: C.green }} />
        ),
      },
      {
        label: "Namespaces",
        href: "/features",
        desc: "Isolate teams, projects and devices logically",
        icon: (
          <RectangleGroupIcon
            className={ICON_CLASS}
            style={{ color: C.primary }}
          />
        ),
      },
      {
        label: "Remote Exec",
        href: "/features",
        desc: "Run commands across device groups simultaneously",
        badge: "New",
        icon: (
          <CommandLineIcon className={ICON_CLASS} style={{ color: C.yellow }} />
        ),
      },
      {
        label: "Audit Logs",
        href: "/features",
        desc: "Immutable history of sessions, commands and events",
        icon: (
          <DocumentTextIcon className={ICON_CLASS} style={{ color: C.cyan }} />
        ),
      },
    ],
  },
  {
    title: "Observability",
    items: [
      {
        label: "Session Replay",
        href: "/features",
        desc: "Watch and replay recorded SSH sessions",
        icon: <PlayCircleIcon className={ICON_CLASS} style={{ color: C.blue }} />,
      },
      {
        label: "Getting Started",
        href: "/getting-started",
        desc: "Deploy ShellHub in minutes, self-hosted or cloud",
        badge: "Free",
        icon: (
          <RocketLaunchIcon className={ICON_CLASS} style={{ color: C.green }} />
        ),
      },
      {
        label: "Changelog",
        href: "#",
        desc: "Release notes and version history",
        icon: (
          <PencilSquareIcon className={ICON_CLASS} style={{ color: C.primary }} />
        ),
      },
    ],
  },
];

export const solutionsCols: MenuSection[] = [
  {
    title: "Infrastructure",
    items: [
      {
        label: "IoT & Embedded",
        href: "/use-cases/iot-embedded",
        desc: "Manage fleets of IoT and embedded Linux devices",
        icon: <CpuChipIcon className={ICON_CLASS} style={{ color: C.green }} />,
      },
      {
        label: "Edge Computing",
        href: "/use-cases/edge-computing",
        desc: "Reach distributed edge servers across any topology",
        icon: <SignalIcon className={ICON_CLASS} style={{ color: C.blue }} />,
      },
      {
        label: "Container Management",
        href: "/use-cases/container-management",
        desc: "SSH directly into Docker containers from anywhere",
        icon: <CubeIcon className={ICON_CLASS} style={{ color: C.cyan }} />,
      },
    ],
  },
  {
    title: "Teams",
    items: [
      {
        label: "Remote Support",
        href: "/use-cases/remote-support",
        desc: "Session recording, replay and audit trail",
        icon: (
          <VideoCameraIcon className={ICON_CLASS} style={{ color: C.yellow }} />
        ),
      },
      {
        label: "DevOps & CI/CD",
        href: "/use-cases/devops-ci-cd",
        desc: "Automate with Ansible, Terraform and CI pipelines",
        icon: (
          <CodeBracketIcon className={ICON_CLASS} style={{ color: C.primary }} />
        ),
      },
      {
        label: "Getting Started",
        href: "/getting-started",
        desc: "Deploy ShellHub in minutes — self-hosted or cloud",
        badge: "Free",
        icon: (
          <RocketLaunchIcon className={ICON_CLASS} style={{ color: C.green }} />
        ),
      },
    ],
  },
];

export const resourcesCols: MenuSection[] = [
  {
    title: "Spotlight",
    items: [
      {
        label: "Documentation",
        href: docsUrl,
        desc: "Full reference, guides and API docs",
        icon: <BookOpenIcon className={ICON_CLASS} style={{ color: C.blue }} />,
      },
      {
        label: "Quick Start",
        href: "/getting-started",
        desc: "Get a device connected in under 5 minutes",
        icon: (
          <RocketLaunchIcon className={ICON_CLASS} style={{ color: C.green }} />
        ),
      },
      {
        label: "Blog",
        href: "#",
        desc: "Tutorials, use cases and product updates",
        icon: <PencilIcon className={ICON_CLASS} style={{ color: C.yellow }} />,
      },
      {
        label: "Changelog",
        href: "#",
        desc: "See what shipped in every release",
        icon: (
          <PencilSquareIcon className={ICON_CLASS} style={{ color: C.primary }} />
        ),
      },
    ],
  },
  {
    title: "Docs & Community",
    items: [
      {
        label: "API Reference",
        href: docsUrl,
        desc: "REST API endpoints and authentication",
        icon: <CodeBracketIcon className={ICON_CLASS} style={{ color: C.cyan }} />,
      },
      {
        label: "GitHub",
        href: githubUrl,
        desc: "Source code, issues and contributions",
        icon: <GithubIcon className={ICON_CLASS} style={{ color: C.textSec }} />,
      },
      {
        label: "Discord",
        href: "#",
        desc: "Chat with the ShellHub community",
        icon: <DiscordIcon className={ICON_CLASS} style={{ color: C.blue }} />,
      },
      {
        label: "Forum",
        href: "#",
        desc: "Ask questions and share knowledge",
        icon: <UsersIcon className={ICON_CLASS} style={{ color: C.primary }} />,
      },
    ],
  },
  {
    title: "Self-Hosted",
    items: [
      {
        label: "Deploy Guide",
        href: docsUrl,
        desc: "Docker Compose, Kubernetes, bare metal",
        icon: <CubeIcon className={ICON_CLASS} style={{ color: C.green }} />,
      },
      {
        label: "Enterprise",
        href: "/enterprise",
        desc: "SAML, LDAP, audit trails and SLA support",
        icon: (
          <HomeModernIcon className={ICON_CLASS} style={{ color: C.yellow }} />
        ),
      },
      {
        label: "Security",
        href: "#",
        desc: "CVEs, responsible disclosure and trust",
        icon: <ShieldCheckIcon className={ICON_CLASS} style={{ color: C.red }} />,
      },
    ],
  },
];

export const simpleLinks: { label: string; href: string }[] = [
  { label: "Enterprise", href: "/enterprise" },
  { label: "Pricing", href: "/pricing" },
];

// Flat de-duplicated list of every href that appears in the nav.
export const navHrefs: string[] = Array.from(
  new Set([
    ...productCols.flatMap((s) => s.items.map((i) => i.href)),
    ...solutionsCols.flatMap((s) => s.items.map((i) => i.href)),
    ...resourcesCols.flatMap((s) => s.items.map((i) => i.href)),
    ...simpleLinks.map((l) => l.href),
  ]),
);
