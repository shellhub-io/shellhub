export interface SidebarItem {
  label: string;
  href?: string;
  featured?: boolean;
  items?: SidebarItem[];
}

export interface SidebarSection {
  label: string;
  description: string;
  icon: string;
  items: SidebarItem[];
}

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
    icon: `<path stroke-linecap="round" stroke-linejoin="round" d="M15.59 14.37a6 6 0 0 1-5.84 7.38v-4.8m5.84-2.58a14.98 14.98 0 0 0 6.16-12.12A14.98 14.98 0 0 0 9.631 8.41m5.96 5.96a14.926 14.926 0 0 1-5.841 2.58m-.119-8.54a6 6 0 0 0-7.381 5.84h4.8m2.58-5.84a14.927 14.927 0 0 1-2.58 5.84m2.699 2.7c-.103.021-.207.041-.311.06a15.09 15.09 0 0 1-2.448-2.448 14.9 14.9 0 0 1 .06-.312m-2.24 2.39a4.493 4.493 0 0 0-1.757 4.306 4.493 4.493 0 0 0 4.306-1.758M16.5 9a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0Z" />`,
    items: [
      {
        label: "Introduction",
        href: "/v2/docs/getting-started/introduction",
        featured: true,
      },
      {
        label: "How it Works",
        href: "/v2/docs/getting-started/how-it-works",
        featured: true,
      },
      {
        label: "Features",
        href: "/v2/docs/getting-started/features",
        featured: true,
      },
      { label: "Editions", href: "/v2/docs/getting-started/editions" },
      {
        label: "Quick Start",
        href: "/v2/docs/getting-started/quick-start",
        featured: true,
      },
      { label: "FAQ", href: "/v2/docs/getting-started/faq" },
    ],
  },
  {
    label: "Resources",
    description:
      "Core platform resources — devices, sessions, and web endpoints.",
    icon: `<path stroke-linecap="round" stroke-linejoin="round" d="M6.429 9.75 2.25 12l4.179 2.25m0-4.5 5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0 4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0-5.571 3-5.571-3" />`,
    items: [
      {
        label: "Devices",
        href: "/v2/docs/guides/devices",
        featured: true,
      },
      {
        label: "Sessions",
        href: "/v2/docs/guides/sessions",
        featured: true,
      },
      {
        label: "Web Endpoints",
        href: "/v2/docs/guides/web-endpoints",
        featured: true,
      },
    ],
  },
  {
    label: "Security",
    description:
      "Secure your devices with public keys and firewall rules.",
    icon: `<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" />`,
    items: [
      {
        label: "Public Keys",
        href: "/v2/docs/guides/public-keys",
        featured: true,
      },
      {
        label: "Firewall Rules",
        href: "/v2/docs/guides/firewall-rules",
        featured: true,
      },
    ],
  },
  {
    label: "Management",
    description:
      "Organize your infrastructure with namespaces, teams, and settings.",
    icon: `<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 6h9.75M10.5 6a1.5 1.5 0 1 1-3 0m3 0a1.5 1.5 0 1 0-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-9.75 0h9.75" />`,
    items: [
      {
        label: "Namespaces",
        href: "/v2/docs/guides/namespaces",
        featured: true,
      },
      {
        label: "Team",
        href: "/v2/docs/guides/team",
        featured: true,
      },
      {
        label: "Settings",
        href: "/v2/docs/guides/settings",
        featured: true,
      },
    ],
  },
  {
    label: "User Guides",
    description:
      "Manage devices, connect via SSH, organize with namespaces and tags.",
    icon: `<path stroke-linecap="round" stroke-linejoin="round" d="M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6 2.292m0-14.25v14.25" />`,
    items: [
      {
        label: "Connecting via SSH",
        href: "/v2/docs/guides/connecting",
        featured: true,
      },
      {
        label: "File Transfer",
        href: "/v2/docs/guides/file-transfer",
        featured: true,
      },
      {
        label: "Port Forwarding",
        href: "/v2/docs/guides/port-forwarding",
        featured: true,
      },
      {
        label: "Container Access",
        href: "/v2/docs/guides/container-access",
        featured: true,
      },
      { label: "Session Recording", href: "/v2/docs/guides/session-recording" },
      { label: "Organizing with Tags", href: "/v2/docs/guides/tags" },
    ],
  },
  {
    label: "Agent",
    description:
      "Install the ShellHub agent on Docker, Snap, FreeBSD, or WSL.",
    icon: `<path stroke-linecap="round" stroke-linejoin="round" d="M8.25 3v1.5M4.5 8.25H3m18 0h-1.5M4.5 12H3m18 0h-1.5m-15 3.75H3m18 0h-1.5M8.25 19.5V21M12 3v1.5m0 15V21m3.75-18v1.5m0 15V21m-9-1.5h10.5a2.25 2.25 0 0 0 2.25-2.25V6.75a2.25 2.25 0 0 0-2.25-2.25H6.75A2.25 2.25 0 0 0 4.5 6.75v10.5a2.25 2.25 0 0 0 2.25 2.25Zm.75-12h9v9h-9v-9Z" />`,
    items: [
      {
        label: "Agent Overview",
        href: "/v2/docs/agent/overview",
        featured: true,
      },
      { label: "Install Script", href: "/v2/docs/agent/install-script" },
      {
        label: "Platform Guides",
        items: [
          {
            label: "Docker",
            href: "/v2/docs/agent/docker",
            featured: true,
          },
          {
            label: "Snap",
            href: "/v2/docs/agent/snap",
            featured: true,
          },
          {
            label: "FreeBSD",
            href: "/v2/docs/agent/freebsd",
            featured: true,
          },
          { label: "WSL", href: "/v2/docs/agent/wsl" },
          {
            label: "Building from Source",
            href: "/v2/docs/agent/building-from-source",
          },
        ],
      },
    ],
  },
  {
    label: "Embedded Linux",
    description:
      "Embed the agent on Raspberry Pi, Buildroot, and Yocto images.",
    icon: `<path stroke-linecap="round" stroke-linejoin="round" d="M14.25 9.75 16.5 12l-2.25 2.25m-4.5 0L7.5 12l2.25-2.25M6 20.25h12A2.25 2.25 0 0 0 20.25 18V6A2.25 2.25 0 0 0 18 3.75H6A2.25 2.25 0 0 0 3.75 6v12A2.25 2.25 0 0 0 6 20.25Z" />`,
    items: [
      {
        label: "Embedded Linux Overview",
        href: "/v2/docs/embedded-linux/overview",
        featured: true,
      },
      { label: "Raspberry Pi", href: "/v2/docs/embedded-linux/raspberry-pi", featured: true },
      {
        label: "Buildroot",
        href: "/v2/docs/embedded-linux/buildroot",
        featured: true,
      },
      {
        label: "Yocto Project",
        href: "/v2/docs/embedded-linux/yocto",
        featured: true,
      },
    ],
  },
  {
    label: "Integrations",
    description:
      "Connect ShellHub with Ansible, Terraform, CI/CD, and VS Code.",
    icon: `<path stroke-linecap="round" stroke-linejoin="round" d="M14.25 6.087c0-.355.186-.676.401-.959.221-.29.349-.634.349-1.003 0-1.036-1.007-1.875-2.25-1.875s-2.25.84-2.25 1.875c0 .369.128.713.349 1.003.215.283.401.604.401.959v0a.64.64 0 0 1-.657.643 48.39 48.39 0 0 1-4.163-.3c.186 1.613.293 3.25.315 4.907a.656.656 0 0 1-.658.663v0c-.355 0-.676-.186-.959-.401a1.647 1.647 0 0 0-1.003-.349c-1.036 0-1.875 1.007-1.875 2.25s.84 2.25 1.875 2.25c.369 0 .713-.128 1.003-.349.283-.215.604-.401.959-.401v0c.31 0 .555.26.532.57a48.039 48.039 0 0 1-.642 5.056c1.518.19 3.058.309 4.616.354a.64.64 0 0 0 .657-.643v0c0-.355-.186-.676-.401-.959a1.647 1.647 0 0 1-.349-1.003c0-1.035 1.008-1.875 2.25-1.875 1.243 0 2.25.84 2.25 1.875 0 .369-.128.713-.349 1.003-.215.283-.4.604-.4.959v0c0 .333.277.599.61.58a48.1 48.1 0 0 0 5.427-.63 48.05 48.05 0 0 0 .582-4.717.532.532 0 0 0-.533-.57v0c-.355 0-.676.186-.959.401-.29.221-.634.349-1.003.349-1.035 0-1.875-1.007-1.875-2.25s.84-2.25 1.875-2.25c.37 0 .713.128 1.003.349.283.215.604.401.96.401v0a.656.656 0 0 0 .658-.663 48.422 48.422 0 0 0-.37-5.36c-1.886.342-3.81.574-5.766.689a.578.578 0 0 1-.61-.58v0Z" />`,
    items: [
      { label: "Integrations Overview", href: "/v2/docs/integration/overview" },
      {
        label: "Ansible",
        href: "/v2/docs/integration/ansible",
        featured: true,
      },
      {
        label: "Terraform",
        href: "/v2/docs/integration/terraform",
        featured: true,
      },
      {
        label: "CI/CD Pipelines",
        href: "/v2/docs/integration/ci-cd",
        featured: true,
      },
      {
        label: "VS Code Remote",
        href: "/v2/docs/integration/vscode-remote",
        featured: true,
      },
    ],
  },
  {
    label: "Self-Hosted",
    description:
      "Deploy, configure, and maintain your own ShellHub instance.",
    icon: `<path stroke-linecap="round" stroke-linejoin="round" d="M5.25 14.25h13.5m-13.5 0a3 3 0 0 1-3-3m3 3a3 3 0 1 0 0 6h13.5a3 3 0 1 0 0-6m-16.5-3a3 3 0 0 1 3-3h13.5a3 3 0 0 1 3 3m-19.5 0a4.5 4.5 0 0 1 .9-2.7L5.737 5.1a3.375 3.375 0 0 1 2.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 0 1 .9 2.7m0 0a3 3 0 0 1-3 3m0 3h.008v.008h-.008v-.008Zm0-6h.008v.008h-.008v-.008Zm-3 6h.008v.008h-.008v-.008Zm0-6h.008v.008h-.008v-.008Z" />`,
    items: [
      {
        label: "Setup",
        href: "/v2/docs/self-hosted/deploying",
        featured: true,
      },
      {
        label: "Configuration",
        href: "/v2/docs/self-hosted/configuring",
        featured: true,
      },
      {
        label: "Administration",
        href: "/v2/docs/self-hosted/administration",
        featured: true,
      },
      {
        label: "Upgrade",
        href: "/v2/docs/self-hosted/upgrading",
        featured: true,
      },
      {
        label: "Troubleshooting",
        href: "/v2/docs/self-hosted/troubleshooting",
      },
    ],
  },
  {
    label: "Developers",
    description:
      "Set up a dev environment, use the API, and build the agent from source.",
    icon: `<path stroke-linecap="round" stroke-linejoin="round" d="M17.25 6.75 22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3-4.5 16.5" />`,
    items: [
      {
        label: "Dev Environment",
        href: "/v2/docs/developers/development-environment",
        featured: true,
      },
      {
        label: "API Reference",
        href: "/v2/docs/developers/api-reference",
        featured: true,
      },
      {
        label: "Agent Development",
        href: "/v2/docs/developers/agent-development",
        featured: true,
      },
    ],
  },
];
