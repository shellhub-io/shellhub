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
  DocumentDuplicateIcon,
  InformationCircleIcon,
} from "@heroicons/react/24/outline";
import { buildSshid } from "@/utils/sshid";
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
  /** When set, the row exposes a drill-in affordance (trailing chevron / "→")
   *  that opens a secondary view instead of selecting. Device rows use it to
   *  open their action menu. */
  onDrillIn?: () => void;
  /** When true, the row is shown but inert (`aria-disabled`): it ignores
   *  click/Enter and can't be selected. The drill-in Connect action uses it
   *  when the device can neither connect nor restore. */
  disabled?: boolean;
}

/** Banner feedback: an assertive error (offline/permission, drives the shake)
 *  or a polite success (e.g. a copy confirmation). */
export type Feedback = { kind: "error" | "success"; text: string };

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
  connect: <ChevronDoubleRightIcon className="w-4 h-4" />,
  copy: <DocumentDuplicateIcon className="w-4 h-4" />,
  details: <InformationCircleIcon className="w-4 h-4" />,
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

/* Default (connection-first) view: open terminal sessions lead so you can jump
 * straight back in (restore), followed by devices to connect/restore. Device
 * rows also expose a drill-in (→) into their action menu. */
export function buildConnectionItems(deps: {
  devices: NormalizedDevice[];
  terminalSessions: TerminalSession[];
  canConnect: boolean;
  connectOrRestore: (uid: string, name: string, online: boolean) => void;
  restoreTerminal: (id: string) => void;
  rejectRow: (rowId: string, message: string) => void;
  enterDrillIn: (uid: string) => void;
  close: () => void;
}): CommandItem[] {
  const {
    devices,
    terminalSessions,
    canConnect,
    connectOrRestore,
    restoreTerminal,
    rejectRow,
    enterDrillIn,
    close,
  } = deps;
  const list: CommandItem[] = [];

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
      onDrillIn: () => enterDrillIn(d.uid),
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

/* Drill-in view: actions for the focused device. The section header doubles
 * as the breadcrumb. Copy/View are ungated (like the Devices page); Connect
 * reuses connectOrRestore, which keeps the permission + offline guards. */
export function buildDeviceActionItems(deps: {
  drillDevice: NormalizedDevice | null;
  nsName: string;
  canConnect: boolean;
  hasOpenSession: boolean;
  connectOrRestore: (uid: string, name: string, online: boolean) => void;
  copyAction: (value: string, label: string) => void;
  go: (path: string) => void;
}): CommandItem[] {
  const {
    drillDevice,
    nsName,
    canConnect,
    hasOpenSession,
    connectOrRestore,
    copyAction,
    go,
  } = deps;
  if (!drillDevice) return [];
  const { uid, name, online } = drillDevice;
  const sshid = nsName ? buildSshid(nsName, name) : uid;
  const sshCommand = `ssh <username>@${sshid}`;
  /* Connect is connect-or-restore, so it's actionable only when the user may
   * connect AND the device is reachable or has an open session to restore.
   * Disable it otherwise — the row stays open for Copy/View details. */
  const connectDisabled = !canConnect || (!online && !hasOpenSession);
  return [
    {
      id: "act-connect",
      label: "Connect",
      sublabel: canConnect ? undefined : "Requires connect permission",
      section: name,
      icon: icons.connect,
      badge: online
        ? { text: "Online", variant: "green" }
        : { text: "Offline", variant: "muted" },
      onSelect: () => connectOrRestore(uid, name, online),
      disabled: connectDisabled,
    },
    {
      id: "act-copy-sshid",
      label: "Copy SSHID",
      sublabel: sshid,
      section: name,
      icon: icons.copy,
      onSelect: () => copyAction(sshid, "SSHID"),
    },
    {
      id: "act-copy-ssh",
      label: "Copy ssh command",
      sublabel: sshCommand,
      section: name,
      icon: icons.copy,
      onSelect: () => copyAction(sshCommand, "ssh command"),
    },
    {
      id: "act-details",
      label: "View details",
      sublabel: `/devices/${uid}`,
      section: name,
      icon: icons.details,
      onSelect: () => go(`/devices/${uid}`),
    },
  ];
}
