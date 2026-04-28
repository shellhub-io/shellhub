import { useCallback, type ReactNode } from "react";
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
} from "@heroicons/react/24/outline";
import SidebarShell, { NavItemLink, navIcon } from "./SidebarShell";

interface NavItem {
  to: string;
  label: string;
  icon: ReactNode;
  premium?: boolean;
}

interface NavSection {
  title: string;
  items: NavItem[];
}

const sections: NavSection[] = [
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
    items: [
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
    ],
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

function ProBadge() {
  return (
    <span className="text-2xs font-mono font-semibold text-accent-yellow/80 bg-accent-yellow/10 px-1.5 py-0.5 rounded">
      Pro
    </span>
  );
}

export default function Sidebar({
  expanded,
  pinned,
  onToggle,
  onClose,
  toggleLabel,
}: {
  expanded: boolean;
  pinned: boolean;
  onToggle: () => void;
  onClose?: () => void;
  toggleLabel?: string;
}) {
  const minimizeAll = useTerminalStore((s) => s.minimizeAll);
  const isFullscreen = useTerminalStore((state) =>
    state.sessions.some((session) => session.state === "fullscreen"),
  );

  const handleNavClick = useCallback(() => {
    minimizeAll();
    onClose?.();
  }, [minimizeAll, onClose]);

  return (
    <SidebarShell
      expanded={expanded}
      pinned={pinned}
      onToggle={onToggle}
      onClose={onClose}
      toggleLabel={toggleLabel}
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
            {section.items.map((item) => {
              const showBadge =
                item.premium && !getConfig().cloud && !getConfig().enterprise;
              return (
                <NavItemLink
                  key={item.to}
                  item={item}
                  expanded={expanded}
                  onClick={handleNavClick}
                  badge={showBadge ? <ProBadge /> : undefined}
                />
              );
            })}
          </div>
        </div>
      ))}
    </SidebarShell>
  );
}
