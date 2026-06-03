import {
  MagnifyingGlassIcon,
  HomeIcon,
  KeyIcon,
  Cog6ToothIcon,
  ArrowRightStartOnRectangleIcon,
  UsersIcon,
  CpuChipIcon,
  CommandLineIcon,
  PlusIcon,
  ChevronDoubleRightIcon,
  LockClosedIcon,
} from "@heroicons/react/24/outline";
import type { NormalizedDevice } from "@/hooks/useDevices";
import type { TerminalSession } from "@/stores/terminalStore";

export type BadgeVariant = "green" | "yellow" | "red" | "muted";

export interface CommandItem {
  id: string;
  label: string;
  sublabel?: string;
  section: string;
  icon: JSX.Element;
  badge?: { text: string; variant: BadgeVariant };
  onSelect: () => void;
}

export const LISTBOX_ID = "cmdk-listbox";
export const optionId = (itemId: string) => `cmdk-opt-${itemId}`;

export const icons = {
  search: <MagnifyingGlassIcon className="w-5 h-5" />,
  dashboard: <HomeIcon className="w-4 h-4" />,
  devices: <CpuChipIcon className="w-4 h-4" />,
  sessions: <CommandLineIcon className="w-4 h-4" />,
  keys: <KeyIcon className="w-4 h-4" />,
  settings: <Cog6ToothIcon className="w-4 h-4" />,
  add: <PlusIcon className="w-4 h-4" />,
  terminal: <ChevronDoubleRightIcon className="w-4 h-4" />,
  logout: <ArrowRightStartOnRectangleIcon className="w-4 h-4" />,
  team: <UsersIcon className="w-4 h-4" />,
  vault: <LockClosedIcon className="w-4 h-4" />,
};

export function fuzzyMatch(query: string, text: string): boolean {
  const q = query.toLowerCase();
  const t = text.toLowerCase();
  if (t.includes(q)) return true;
  let qi = 0;
  for (let ti = 0; ti < t.length && qi < q.length; ti++) {
    if (t[ti] === q[qi]) qi++;
  }
  return qi === q.length;
}

/* All palette entries in one flat list: page navigation, accepted devices,
 * open terminal sessions, and account actions. */
export function buildItems(deps: {
  devices: NormalizedDevice[];
  terminalSessions: TerminalSession[];
  go: (path: string) => void;
  close: () => void;
  restoreTerminal: (id: string) => void;
  onLogout: () => void;
}): CommandItem[] {
  const { devices, terminalSessions, go, close, restoreTerminal, onLogout } =
    deps;
  const list: CommandItem[] = [];

  const nav: Array<{ label: string; path: string; icon: JSX.Element }> = [
    { label: "Dashboard", path: "/dashboard", icon: icons.dashboard },
    { label: "Devices", path: "/devices", icon: icons.devices },
    { label: "Sessions", path: "/sessions", icon: icons.sessions },
    { label: "Public Keys", path: "/sshkeys/public-keys", icon: icons.keys },
    { label: "Secure Vault", path: "/secure-vault", icon: icons.vault },
    { label: "Team", path: "/team", icon: icons.team },
    { label: "Settings", path: "/settings", icon: icons.settings },
    { label: "Add Device", path: "/devices/add", icon: icons.add },
  ];

  nav.forEach((n) => {
    list.push({
      id: `nav-${n.path}`,
      label: n.label,
      sublabel: n.path,
      section: "Navigation",
      icon: n.icon,
      onSelect: () => go(n.path),
    });
  });

  devices
    .filter((d) => d.status === "accepted")
    .forEach((d) => {
      list.push({
        id: `device-${d.uid}`,
        label: d.name,
        sublabel: d.identity?.mac ?? d.uid.slice(0, 12),
        section: "Devices",
        icon: icons.devices,
        badge: d.online
          ? { text: "Online", variant: "green" }
          : { text: "Offline", variant: "muted" },
        onSelect: () => go(`/devices/${d.uid}`),
      });
    });

  terminalSessions.forEach((s) => {
    const statusLabel =
      s.connectionStatus === "connected"
        ? "Connected"
        : s.connectionStatus === "connecting"
          ? "Connecting"
          : "Disconnected";
    const statusVariant: BadgeVariant =
      s.connectionStatus === "connected"
        ? "green"
        : s.connectionStatus === "connecting"
          ? "yellow"
          : "red";
    list.push({
      id: `term-${s.id}`,
      label: `${s.username}@${s.deviceName}`,
      sublabel:
        s.state === "minimized"
          ? "Minimized"
          : s.state === "fullscreen"
            ? "Fullscreen"
            : "Docked",
      section: "Terminal Sessions",
      icon: icons.terminal,
      badge: { text: statusLabel, variant: statusVariant },
      onSelect: () => {
        close();
        restoreTerminal(s.id);
      },
    });
  });

  list.push({
    id: "action-logout",
    label: "Logout",
    section: "Actions",
    icon: icons.logout,
    onSelect: onLogout,
  });

  return list;
}
