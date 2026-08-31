import {
  BookOpenIcon,
  CodeBracketIcon,
  Cog6ToothIcon,
  CpuChipIcon,
  CubeIcon,
  PencilSquareIcon,
  PuzzlePieceIcon,
  RocketLaunchIcon,
  SignalIcon,
  SparklesIcon,
  UsersIcon,
  VideoCameraIcon,
} from "@heroicons/react/24/outline";
import { GithubIcon } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { docsUrl, githubUrl } from "@/links";

const ICON_SIZE = "size-4";

/**
 * One entry of a navigation dropdown. desc is shown beneath the label on desktop and dropped on
 * mobile, where there is no room for it.
 */
export interface DropdownItem {
  label: string;
  href: string;
  desc: string;
  icon: React.ReactNode;
}

/**
 * A navigation entry: either a link that goes straight somewhere, or a dropdown holding several.
 * The kind tag is what both the desktop bar and the mobile sheet switch on.
 */
export type NavEntry =
  | { kind: "link"; label: string; href: string }
  | { kind: "dropdown"; label: string; items: DropdownItem[] };

/**
 * The site navigation, in order. Desktop and mobile both render from this, so an entry added
 * here appears in both and cannot drift between them.
 */
export const navEntries: NavEntry[] = [
  {
    kind: "dropdown",
    label: "Product",
    items: [
      {
        label: "Features",
        href: "/features",
        desc: "Core capabilities and platform overview",
        icon: <SparklesIcon className={cn(ICON_SIZE, "text-accent-green")} />,
      },
      {
        label: "How It Works",
        href: "/how-it-works",
        desc: "Architecture, tunneling and NAT traversal",
        icon: <Cog6ToothIcon className={cn(ICON_SIZE, "text-accent-blue")} />,
      },
      {
        label: "Integrations",
        href: "/integrations",
        desc: "Ansible, Terraform, CI/CD and VS Code",
        icon: <PuzzlePieceIcon className={cn(ICON_SIZE, "text-accent-cyan")} />,
      },
      {
        label: "Getting Started",
        href: "/getting-started",
        desc: "Quickstart guide to your first connection",
        icon: (
          <RocketLaunchIcon className={cn(ICON_SIZE, "text-accent-yellow")} />
        ),
      },
    ],
  },
  {
    kind: "dropdown",
    label: "Solutions",
    items: [
      {
        label: "IoT & Embedded",
        href: "/use-cases/iot-embedded",
        desc: "Manage fleets of IoT and embedded devices",
        icon: <CpuChipIcon className={cn(ICON_SIZE, "text-accent-green")} />,
      },
      {
        label: "Edge Computing",
        href: "/use-cases/edge-computing",
        desc: "Reach distributed edge servers across any topology",
        icon: <SignalIcon className={cn(ICON_SIZE, "text-accent-blue")} />,
      },
      {
        label: "Container Management",
        href: "/use-cases/container-management",
        desc: "SSH directly into Docker containers from anywhere",
        icon: <CubeIcon className={cn(ICON_SIZE, "text-accent-cyan")} />,
      },
      {
        label: "Remote Support",
        href: "/use-cases/remote-support",
        desc: "Session recording, replay and audit trail",
        icon: (
          <VideoCameraIcon className={cn(ICON_SIZE, "text-accent-yellow")} />
        ),
      },
      {
        label: "DevOps & CI/CD",
        href: "/use-cases/devops-ci-cd",
        desc: "Automate with Ansible, Terraform and CI pipelines",
        icon: <CodeBracketIcon className={cn(ICON_SIZE, "text-primary")} />,
      },
    ],
  },
  {
    kind: "dropdown",
    label: "Resources",
    items: [
      {
        label: "Documentation",
        href: docsUrl,
        desc: "Full reference, guides and API docs",
        icon: <BookOpenIcon className={cn(ICON_SIZE, "text-accent-blue")} />,
      },
      {
        label: "GitHub",
        href: githubUrl,
        desc: "Source code, issues and contributions",
        icon: <GithubIcon className={cn(ICON_SIZE, "text-text-secondary")} />,
      },
      {
        label: "Changelog",
        href: `${githubUrl}/releases`,
        desc: "Release notes and version history",
        icon: <PencilSquareIcon className={cn(ICON_SIZE, "text-primary")} />,
      },
      {
        label: "Forum",
        href: `${githubUrl}/discussions`,
        desc: "Ask questions and share knowledge",
        icon: <UsersIcon className={cn(ICON_SIZE, "text-primary")} />,
      },
    ],
  },
  { kind: "link", label: "Enterprise", href: "/enterprise" },
  { kind: "link", label: "Pricing", href: "/pricing" },
];
