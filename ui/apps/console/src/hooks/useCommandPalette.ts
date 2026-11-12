import { useEffect, useEffectEvent, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useDevices } from "@/hooks/useDevices";
import type { NormalizedDevice } from "@/hooks/useDevices";
import { useOrderedWindows, useTerminalStore } from "@/stores/terminalStore";
import { useAuthStore } from "@/stores/authStore";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";
import { adminNavLinks, useAdminNav } from "@/components/layout/adminNav";
import {
  useRecentDevicesStore,
  type RecentDevice,
} from "@/stores/recentDevicesStore";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useNamespace } from "@/hooks/useNamespaceRequests";
import { useNamespaceLegacy } from "@/hooks/useNamespaceLegacy";
import { useWorkspaceTabs } from "@/hooks/useWorkspaceTabs";
import { useCopy } from "@/hooks/useCopy";
import { isAdminPath } from "@/utils/adminRoute";
import {
  buildConnectionItems,
  buildAdminItems,
  buildCommandItems,
  buildDeviceActionItems,
  fuzzyMatch,
  handleListboxKey,
  useListboxHighlight,
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
  connect: boolean;
  nsName: string;
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
  handleBackdropClose: () => void;
  exitDrillIn: () => void;
  backToPicker: () => void;
}

/**
 * Headless controller for the device palette: owns all state, derivation,
 * effects, and handlers, and returns a view-model. Keeps the rendering shell
 * and its parts free of logic. It serves both the "palette" view Ctrl/Cmd+K
 * opens and the "connect" view the new-tab picker opens; connect leaves out
 * the terminal sessions and ">" commands. In either view Escape, or ArrowLeft
 * at the start of the input, first leaves a device's actions; after that, in
 * connect, Escape, or ArrowLeft on an empty query, returns to the picker
 * rather than closing. A click outside closes connect outright, while in the palette
 * view it acts as Escape does.
 */
export function useCommandPalette(): CommandPaletteViewModel {
  const view = useCommandPaletteStore((s) => s.view);
  const closeView = useCommandPaletteStore((s) => s.close);
  const showView = useCommandPaletteStore((s) => s.show);
  const open = view === "palette" || view === "connect";
  const connect = view === "connect";
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
  const openSessions = useTerminalStore((s) => s.sessions);
  const windowIds = useOrderedWindows();
  const terminalSessions = windowIds.flatMap(
    (id) => openSessions.find((s) => s.id === id) ?? [],
  );
  const logout = useAuthStore((s) => s.logout);
  const tenant = useAuthStore((s) => s.tenant);
  const canConnect = useHasPermission("device:connect");
  const { namespace } = useNamespace(tenant ?? "");
  const legacy = useNamespaceLegacy();
  const workspace = useWorkspaceTabs();
  const adminNav = useAdminNav({ active: adminContext });
  const nsName = namespace?.name ?? "";
  const copy = useCopy();
  const recentEntries = useRecentDevicesStore(
    (s) => s.byTenant[tenant ?? ""] ?? EMPTY_RECENTS,
  );

  const drillDevice = drillInUid
    ? (devices.find((d) => d.uid === drillInUid) ?? null)
    : null;
  const isDrilledIn = drillDevice !== null;

  useEffect(() => {
    inputRef.current?.focus();
  }, [drillInUid, view]);

  useEffect(() => {
    if (!shakeId) return undefined;
    const timer = setTimeout(() => setShakeId(null), 450);
    return () => clearTimeout(timer);
  }, [shakeId]);

  const close = () => {
    exitDrillIn();
    closeView();
  };

  const backToPicker = () => {
    exitDrillIn();
    showView("picker");
  };

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

  const resetSearch = () => {
    setQuery("");
    setActiveIndex(0);
    setFeedback(null);
    setShakeId(null);
  };

  const enterDrillIn = (uid: string) => {
    resetSearch();
    setDrillInUid(uid);
  };

  const exitDrillIn = () => {
    resetSearch();
    setDrillInUid(null);
  };

  const handleDismiss = () => {
    if (isDrilledIn) exitDrillIn();
    else if (connect) backToPicker();
    else close();
  };

  const handleBackdropClose = connect ? close : handleDismiss;

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
      void workspace.showTerminal(existing);
      return;
    }
    if (!online) {
      rejectRow(rowId, `${name} is offline — start it to connect`);
      return;
    }
    close();
    store.requestConnect(uid, name);
  };

  const toggleFromShortcut = useEffectEvent(() => {
    if (view) close();
    else showView("palette");
  });

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        toggleFromShortcut();
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  const openUids = new Set(terminalSessions.map((s) => s.deviceUid));
  const recentDevices: { device: NormalizedDevice; connectedAt: string }[] = [];
  for (const entry of recentEntries) {
    if (openUids.has(entry.uid)) continue;
    const device = devices.find((d) => d.uid === entry.uid);
    if (!device) continue;
    recentDevices.push({ device, connectedAt: entry.connectedAt });
    if (recentDevices.length >= RECENT_LIMIT) break;
  }

  const connectionItems = buildConnectionItems({
    devices: adminContext ? [] : devices,
    terminalSessions: connect ? [] : terminalSessions,
    recentDevices: adminContext ? [] : recentDevices,
    canConnect,
    connectOrRestore,
    restoreTerminal: (session) => void workspace.showTerminal(session),
    rejectRow,
    enterDrillIn,
    close,
  });

  const commandItems = buildCommandItems({
    go,
    onLogout,
    isNewNamespace: legacy !== true,
    pages: adminContext
      ? buildAdminItems({ links: adminNavLinks(adminNav.sections), go })
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
  const commandMode =
    !drillDevice && !connect && trimmedQuery.startsWith(">");
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

  const { safeIndex, activeItem } = useListboxHighlight(
    flatList,
    activeIndex,
    listRef,
  );

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (handleListboxKey(e, flatList, safeIndex, setActiveIndex)) return;
    const input = e.currentTarget;
    const caretAtEnd =
      input.selectionStart === input.selectionEnd &&
      input.selectionStart === input.value.length;
    const caretAtStart =
      input.selectionStart === input.selectionEnd && input.selectionStart === 0;

    if (e.key === "ArrowRight") {
      const item = flatList[safeIndex];
      if (caretAtEnd && !isDrilledIn && item?.onDrillIn) {
        e.preventDefault();
        item.onDrillIn();
      }
    } else if (e.key === "ArrowLeft" && caretAtStart) {
      if (isDrilledIn) {
        e.preventDefault();
        exitDrillIn();
      } else if (connect && !query) {
        e.preventDefault();
        backToPicker();
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
    connect,
    nsName,
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
    handleBackdropClose,
    exitDrillIn,
    backToPicker,
  };
}
