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
  ChevronRightIcon,
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

/* Shown when a user without `device:connect` tries to connect or restore —
 * mirrors the Devices page, which gates the whole Connect button behind it. */
export const NO_CONNECT_PERMISSION =
  "You don't have permission to connect to devices";

export const icons = {
  search: <MagnifyingGlassIcon className="w-5 h-5" />,
  command: <ChevronRightIcon className="w-5 h-5" />,
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

/* Default (connection-first) view: devices to connect/restore + open sessions. */
export function buildConnectionItems(deps: {
  devices: NormalizedDevice[];
  terminalSessions: TerminalSession[];
  canConnect: boolean;
  connectOrRestore: (uid: string, name: string, online: boolean) => void;
  restoreTerminal: (id: string) => void;
  rejectRow: (rowId: string, message: string) => void;
  close: () => void;
}): CommandItem[] {
  const {
    devices,
    terminalSessions,
    canConnect,
    connectOrRestore,
    restoreTerminal,
    rejectRow,
    close,
  } = deps;
  const list: CommandItem[] = [];

  // useDevices is called with status: "accepted", so the API already scopes
  // this list — no client-side status filter needed.
  devices.forEach((d) => {
    list.push({
      id: `device-${d.uid}`,
      label: d.name,
      sublabel: d.identity?.mac ?? d.uid.slice(0, 12),
      section: "Devices",
      icon: icons.devices,
      badge: d.online
        ? { text: "Online", variant: "green" }
        : { text: "Offline", variant: "muted" },
      onSelect: () => connectOrRestore(d.uid, d.name, d.online),
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
        if (!canConnect) {
          rejectRow(`term-${s.id}`, NO_CONNECT_PERMISSION);
          return;
        }
        close();
        restoreTerminal(s.id);
      },
    });
  });

  return list;
}

/* Command mode (">" prefix): page navigation + account actions. */
export function buildCommandItems(deps: {
  go: (path: string) => void;
  onLogout: () => void;
}): CommandItem[] {
  const { go, onLogout } = deps;
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

  const list: CommandItem[] = nav.map((n) => ({
    id: `nav-${n.path}`,
    label: n.label,
    sublabel: n.path,
    section: "Navigation",
    icon: n.icon,
    onSelect: () => go(n.path),
  }));

  list.push({
    id: "action-logout",
    label: "Logout",
    section: "Actions",
    icon: icons.logout,
    onSelect: onLogout,
  });

  return list;
}
