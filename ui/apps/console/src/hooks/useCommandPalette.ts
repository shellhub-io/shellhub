import { useCallback, useEffect, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useDevices } from "@/hooks/useDevices";
import type { NormalizedDevice } from "@/hooks/useDevices";
import { useTerminalStore } from "@/stores/terminalStore";
import { useAuthStore } from "@/stores/authStore";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";
import { useCreateNamespaceStore } from "@/stores/createNamespaceStore";
import { adminNavLinks, useAdminNav } from "@/components/layout/adminNav";
import {
  useRecentDevicesStore,
  type RecentDevice,
} from "@/stores/recentDevicesStore";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useNamespace, useNamespaces } from "@/hooks/useNamespaces";
import { useWorkspaceTabs } from "@/hooks/useWorkspaceTabs";
import { isEnterpriseOrCloud } from "@/env";
import { useCopy } from "@/hooks/useCopy";
import { isAdminPath } from "@/utils/adminRoute";
import {
  buildConnectionItems,
  buildWorkspaceItems,
  buildAdminItems,
  buildCommandItems,
  buildDeviceActionItems,
  fuzzyMatch,
  NO_CONNECT_PERMISSION,
  type CommandItem,
  type Feedback,
} from "@/components/commandPalette/items";

/* Stable empty reference for tenants with no recents — keeps the selector from
 * returning a fresh array each render. */
const EMPTY_RECENTS: RecentDevice[] = [];
/* How many recent devices the palette shows. The store keeps more (see
 * `STORE_CAP` in recentDevicesStore) so hiding open-session ones still fills it. */
const RECENT_LIMIT = 5;

/** The view-model the palette shell and its presentational parts consume. */
export interface CommandPaletteViewModel {
  open: boolean;
  inputRef: React.RefObject<HTMLInputElement | null>;
  listRef: React.RefObject<HTMLDivElement | null>;
  query: string;
  drillDevice: NormalizedDevice | null;
  commandMode: boolean;
  adminContext: boolean;
  sections: Map<string, CommandItem[]>;
  hasResults: boolean;
  indexById: Map<string, number>;
  safeIndex: number;
  activeItem: CommandItem | undefined;
  feedback: Feedback | null;
  shakeId: string | null;
  onQueryChange: (value: string) => void;
  setActiveIndex: (index: number) => void;
  handleKeyDown: (e: React.KeyboardEvent<HTMLInputElement>) => void;
  handleDismiss: () => void;
  exitDrillIn: () => void;
}

/**
 * Headless controller for the command palette: owns all state, derivation,
 * effects, and handlers, and returns a view-model. Keeps the rendering shell
 * and its parts free of logic.
 */
export function useCommandPalette(): CommandPaletteViewModel {
  const open = useCommandPaletteStore((s) => s.open);
  const closePalette = useCommandPaletteStore((s) => s.closePalette);
  const openCreateNamespace = useCreateNamespaceStore((s) => s.openDialog);
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const [drillInUid, setDrillInUid] = useState<string | null>(null);
  const [feedback, setFeedback] = useState<Feedback | null>(null);
  const [shakeId, setShakeId] = useState<string | null>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();
  const adminContext = isAdminPath(useLocation().pathname);

  const { devices } = useDevices({ page: 1, perPage: 50, status: "accepted" });
  const terminalSessions = useTerminalStore((s) => s.sessions);
  const restoreTerminal = useTerminalStore((s) => s.restore);
  const logout = useAuthStore((s) => s.logout);
  const tenant = useAuthStore((s) => s.tenant);
  const canConnect = useHasPermission("device:connect");
  const { namespace } = useNamespace(tenant ?? "");
  const { namespaces } = useNamespaces();
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const workspace = useWorkspaceTabs();
  const adminNav = useAdminNav({ active: adminContext });
  const nsName = namespace?.name ?? "";
  const { copy } = useCopy();
  const recentEntries = useRecentDevicesStore(
    (s) => s.byTenant[tenant ?? ""] ?? EMPTY_RECENTS,
  );

  const drillDevice = drillInUid
    ? (devices.find((d) => d.uid === drillInUid) ?? null)
    : null;
  const isDrilledIn = drillDevice !== null;

  useEffect(() => {
    inputRef.current?.focus();
  }, [drillInUid]);

  useEffect(() => {
    if (!shakeId) return undefined;
    const timer = setTimeout(() => setShakeId(null), 450);
    return () => clearTimeout(timer);
  }, [shakeId]);

  const close = useCallback(() => {
    setQuery("");
    setActiveIndex(0);
    setFeedback(null);
    setShakeId(null);
    setDrillInUid(null);
    closePalette();
  }, [closePalette]);

  const go = (path: string) => {
    close();
    void navigate(path);
  };

  const onLogout = () => {
    close();
    logout();
    void navigate("/login");
  };

  const rejectRow = (rowId: string, message: string) => {
    setFeedback({ kind: "error", text: message });
    setShakeId(rowId);
  };

  const copyAction = (value: string, label: string) => {
    copy(value);
    setFeedback({ kind: "success", text: `Copied ${label} to clipboard` });
  };

  const enterDrillIn = (uid: string) => {
    setDrillInUid(uid);
    setQuery("");
    setActiveIndex(0);
    setFeedback(null);
    setShakeId(null);
  };

  const exitDrillIn = () => {
    setDrillInUid(null);
    setQuery("");
    setActiveIndex(0);
    setFeedback(null);
    setShakeId(null);
  };

  const handleDismiss = () => {
    if (isDrilledIn) exitDrillIn();
    else close();
  };

  const connectOrRestore = (
    uid: string,
    name: string,
    online: boolean,
    rowId: string,
  ) => {
    if (!canConnect) {
      rejectRow(rowId, NO_CONNECT_PERMISSION);
      return;
    }
    const store = useTerminalStore.getState();
    const existing = store.sessions.find((s) => s.deviceUid === uid);
    if (existing) {
      close();
      store.restore(existing.id);
      return;
    }
    if (!online) {
      rejectRow(rowId, `${name} is offline — start it to connect`);
      return;
    }
    close();
    store.requestConnect(uid, name);
  };

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        const store = useCommandPaletteStore.getState();
        if (store.open) close();
        else store.openPalette();
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [close]);

  const openUids = new Set(terminalSessions.map((s) => s.deviceUid));
  const recentDevices: { device: NormalizedDevice; connectedAt: string }[] = [];
  for (const entry of recentEntries) {
    if (openUids.has(entry.uid)) continue;
    const device = devices.find((d) => d.uid === entry.uid);
    if (!device) continue;
    recentDevices.push({ device, connectedAt: entry.connectedAt });
    if (recentDevices.length >= RECENT_LIMIT) break;
  }

  const connectionItems = [
    ...buildConnectionItems({
      devices: adminContext ? [] : devices,
      terminalSessions,
      recentDevices: adminContext ? [] : recentDevices,
      canConnect,
      connectOrRestore,
      restoreTerminal,
      rejectRow,
      enterDrillIn,
      close,
    }),
    ...buildWorkspaceItems({
      namespaces,
      currentTenant: tenant,
      showAdmin: isEnterpriseOrCloud() && isAdmin,
      openNamespace: (namespaceTenant, name) =>
        void workspace.openNamespace(namespaceTenant, name),
      openAdmin: () => void workspace.openAdmin(),
      createNamespace: openCreateNamespace,
      close,
    }),
  ];

  const commandItems = buildCommandItems({
    go,
    onLogout,
    isIdentityMode: namespace?.settings?.ssh_access_mode === "identity",
    pages: adminContext
      ? buildAdminItems({ links: adminNavLinks(adminNav.entries), go })
      : undefined,
  });

  const hasOpenSession = drillDevice
    ? terminalSessions.some((s) => s.deviceUid === drillDevice.uid)
    : false;

  const deviceActionItems = buildDeviceActionItems({
    drillDevice,
    nsName,
    canConnect,
    hasOpenSession,
    connectOrRestore,
    copyAction,
    go,
  });

  const trimmedQuery = query.trimStart();
  const commandMode = !drillDevice && trimmedQuery.startsWith(">");
  const term = commandMode ? trimmedQuery.slice(1).trim() : query.trim();
  const activeItems = drillDevice
    ? deviceActionItems
    : commandMode
      ? commandItems
      : connectionItems;

  const filtered = term
    ? activeItems.filter(
        (item) =>
          fuzzyMatch(term, item.label) ||
          (item.sublabel && fuzzyMatch(term, item.sublabel)) ||
          fuzzyMatch(term, item.section),
      )
    : activeItems;

  const sections = new Map<string, CommandItem[]>();
  filtered.forEach((item) => {
    const existing = sections.get(item.section);
    if (existing) existing.push(item);
    else sections.set(item.section, [item]);
  });

  const flatList = [...sections.values()].flat();
  const indexById = new Map(flatList.map((item, i) => [item.id, i]));

  const safeIndex = flatList.length
    ? Math.min(activeIndex, flatList.length - 1)
    : -1;
  const activeItem = safeIndex >= 0 ? flatList[safeIndex] : undefined;

  useEffect(() => {
    if (!listRef.current) return;
    const active = listRef.current.querySelector("[data-active=true]");
    active?.scrollIntoView({ block: "nearest" });
  }, [safeIndex]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    const len = flatList.length;
    const move = (delta: number) =>
      setActiveIndex((prev) => (Math.min(prev, len - 1) + delta + len) % len);
    const input = e.currentTarget;
    const caretAtEnd =
      input.selectionStart === input.selectionEnd &&
      input.selectionStart === input.value.length;
    const caretAtStart =
      input.selectionStart === input.selectionEnd && input.selectionStart === 0;

    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        if (len) move(1);
        break;
      case "ArrowUp":
        e.preventDefault();
        if (len) move(-1);
        break;
      case "ArrowRight": {
        const item = flatList[safeIndex];
        if (caretAtEnd && !isDrilledIn && item?.onDrillIn) {
          e.preventDefault();
          item.onDrillIn();
        }
        break;
      }
      case "ArrowLeft":
        if (caretAtStart && isDrilledIn) {
          e.preventDefault();
          exitDrillIn();
        }
        break;
      case "Home":
        if (!len) break;
        e.preventDefault();
        setActiveIndex(0);
        break;
      case "End":
        if (!len) break;
        e.preventDefault();
        setActiveIndex(len - 1);
        break;
      case "Enter": {
        e.preventDefault();
        const active = safeIndex >= 0 ? flatList[safeIndex] : undefined;
        if (active && !active.disabled) active.onSelect();
        break;
      }
    }
  };

  const onQueryChange = (value: string) => {
    setQuery(value);
    setActiveIndex(0);
    setFeedback(null);
    setShakeId(null);
  };

  const hasResults = flatList.length > 0;

  return {
    open,
    inputRef,
    listRef,
    query,
    drillDevice,
    commandMode,
    adminContext,
    sections,
    hasResults,
    indexById,
    safeIndex,
    activeItem,
    feedback,
    shakeId,
    onQueryChange,
    setActiveIndex,
    handleKeyDown,
    handleDismiss,
    exitDrillIn,
  };
}
