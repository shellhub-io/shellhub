import { NavLink } from "react-router-dom";
import { ReactNode } from "react";
import { getConfig } from "../../env";
import { useTerminalStore } from "../../stores/terminalStore";
import {
  HomeIcon,
  KeyIcon,
  Cog6ToothIcon,
  UsersIcon,
  CpuChipIcon,
  CommandLineIcon,
  LockClosedIcon,
} from "@heroicons/react/24/outline";

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
        icon: <HomeIcon className="w-[18px] h-[18px]" />,
      },
    ],
  },
  {
    title: "Resources",
    items: [
      {
        to: "/devices",
        label: "Devices",
        icon: <CpuChipIcon className="w-[18px] h-[18px]" />,
      },
      {
        to: "/sessions",
        label: "Sessions",
        icon: <CommandLineIcon className="w-[18px] h-[18px]" />,
      },
    ],
  },
  {
    title: "Security",
    items: [
      {
        to: "/sshkeys/public-keys",
        label: "Public Keys",
        icon: <KeyIcon className="w-[18px] h-[18px]" />,
      },
      {
        to: "/secure-vault",
        label: "Secure Vault",
        icon: <LockClosedIcon className="w-[18px] h-[18px]" />,
      },
    ],
  },
  {
    title: "Management",
    items: [
      {
        to: "/team",
        label: "Team",
        icon: <UsersIcon className="w-[18px] h-[18px]" />,
      },
      {
        to: "/settings",
        label: "Settings",
        icon: <Cog6ToothIcon className="w-[18px] h-[18px]" />,
      },
    ],
  },
];

function NavItemLink({ item }: { item: NavItem }) {
  const minimizeAll = useTerminalStore((s) => s.minimizeAll);
  const showBadge =
    item.premium && !getConfig().cloud && !getConfig().enterprise;

  return (
    <NavLink
      to={item.to}
      onClick={minimizeAll}
      className={({ isActive }) =>
        `flex items-center gap-3 px-3 py-2 rounded-md text-[13px] font-medium transition-all duration-150 ${
          isActive
            ? "bg-primary/10 text-primary border border-primary/20"
            : "text-text-secondary hover:text-text-primary hover:bg-hover-subtle border border-transparent"
        }`
      }
    >
      {item.icon}
      <span className="flex-1">{item.label}</span>
      {showBadge && (
        <span className="text-2xs font-mono font-semibold text-accent-yellow/80 bg-accent-yellow/10 px-1.5 py-0.5 rounded">
          Pro
        </span>
      )}
    </NavLink>
  );
}

export default function Sidebar() {
  const isFullscreen = useTerminalStore((s) =>
    s.sessions.some((s) => s.state === "fullscreen"),
  );

  return (
    <aside
      className={`bg-surface border-r border-border flex flex-col min-h-screen shrink-0 transition-all duration-[150ms] ease-in-out overflow-hidden ${
        isFullscreen ? "w-0 opacity-0" : "w-[220px] opacity-100"
      }`}
    >
      <div className="h-14 flex items-center justify-center border-b border-border">
        <img src="/logo.svg" alt="ShellHub" className="h-8" />
      </div>

      <nav className="flex-1 px-3 pt-4 py-2 overflow-y-auto">
        {sections.map((section, idx) => (
          <div key={section.title} className={idx > 0 ? "mt-5" : ""}>
            <p className="px-3 mb-1.5 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted/60">
              {section.title}
            </p>
            <div className="space-y-0.5">
              {section.items.map((item) => (
                <NavItemLink key={item.to} item={item} />
              ))}
            </div>
          </div>
        ))}
      </nav>

      <div className="h-11 px-4 flex items-center border-t border-border">
        <p className="text-2xs font-mono text-text-muted/60">ShellHub v2</p>
      </div>
    </aside>
  );
}
