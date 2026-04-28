import { useCallback, useMemo, type ReactNode } from "react";
import { getConfig } from "@/env";
import { useTerminalStore } from "@/stores/terminalStore";
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
} from "@heroicons/react/24/outline";
import SidebarShell, { NavItemLink, navIcon } from "./SidebarShell";

interface NavItem {
  to: string;
  label: string;
  icon: ReactNode;
  premium?: boolean;
  beta?: boolean;
}

interface NavSection {
  title: string;
  items: NavItem[];
}

function buildSections(): NavSection[] {
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
    {
      to: "/sessions",
      label: "Sessions",
      icon: <CommandLineIcon className={navIcon} />,
    },
  ];

  if (config.webEndpoints && (config.cloud || config.enterprise)) {
    resources.push({
      to: "/web-endpoints",
      label: "Web Endpoints",
      icon: <GlobeAltIcon className={navIcon} />,
      beta: true,
    });
  }

  return [
    {
      title: "Overview",
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
      title: "Security",
      items: [
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
      ],
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

function ProBadge() {
  return (
    <span className="text-2xs font-mono font-semibold text-accent-yellow/80 bg-accent-yellow/10 px-1.5 py-0.5 rounded">
      Pro
    </span>
  );
}

function BetaBadge() {
  return (
    <span className="text-2xs font-mono font-semibold text-accent-cyan/90 bg-accent-cyan/10 px-1.5 py-0.5 rounded">
      Beta
    </span>
  );
}

function pickBadge(item: NavItem): ReactNode | undefined {
  const config = getConfig();
  const isPaidEdition = config.cloud || config.enterprise;
  if (item.premium && !isPaidEdition) return <ProBadge />;
  if (item.beta) return <BetaBadge />;
  return undefined;
}

export default function Sidebar({
  expanded,
  onToggle,
  onClose,
}: {
  expanded: boolean;
  onToggle?: () => void;
  onClose?: () => void;
}) {
  const minimizeAll = useTerminalStore((s) => s.minimizeAll);
  const isFullscreen = useTerminalStore((state) =>
    state.sessions.some((session) => session.state === "fullscreen"),
  );

  const sections = useMemo(() => buildSections(), []);

  const handleNavClick = useCallback(() => {
    minimizeAll();
    onClose?.();
  }, [minimizeAll, onClose]);

  return (
    <SidebarShell
      expanded={expanded}
      onToggle={onToggle}
      onClose={onClose}
      hidden={isFullscreen}
      ariaLabel="Main navigation"
      footerLabel="Console"
      logoHref="/dashboard"
    >
      {sections.map((section, idx) => (
        <div
          key={section.title}
          className={idx > 0 ? (expanded ? "mt-5" : "mt-1") : ""}
        >
          <p
            className={`px-3 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted/60 transition-all duration-200 ${
              expanded
                ? "opacity-100 mb-1.5"
                : "opacity-0 h-0 overflow-hidden mb-0"
            }`}
          >
            {section.title}
          </p>
          <div className="space-y-0.5">
            {section.items.map((item) => (
              <NavItemLink
                key={item.to}
                item={item}
                expanded={expanded}
                onClick={handleNavClick}
                badge={pickBadge(item)}
              />
            ))}
          </div>
        </div>
      ))}
    </SidebarShell>
  );
}
