import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useDevices } from "@/hooks/useDevices";
import type { NormalizedDevice } from "@/hooks/useDevices";
import { useTerminalStore } from "@/stores/terminalStore";
import { useAuthStore } from "@/stores/authStore";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useNamespace } from "@/hooks/useNamespaces";
import { useCopy } from "@/hooks/useCopy";
import {
  buildConnectionItems,
  buildCommandItems,
  buildDeviceActionItems,
  fuzzyMatch,
  NO_CONNECT_PERMISSION,
  type CommandItem,
  type Feedback,
} from "@/components/commandPalette/items";

/** The view-model the palette shell and its presentational parts consume. */
export interface CommandPaletteViewModel {
  open: boolean;
  // Refs the JSX attaches; the hook reads `.current` only inside effects.
  inputRef: React.RefObject<HTMLInputElement>;
  listRef: React.RefObject<HTMLDivElement>;
  // Derived view data (computed during render).
  query: string;
  drillDevice: NormalizedDevice | null;
  commandMode: boolean;
  sections: Map<string, CommandItem[]>;
  hasResults: boolean;
  indexById: Map<string, number>;
  safeIndex: number;
  activeItem: CommandItem | undefined;
  feedback: Feedback | null;
  shakeId: string | null;
  // Handlers.
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
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  /* Drilled-in device whose action menu is showing (null = the normal list). */
  const [drillInUid, setDrillInUid] = useState<string | null>(null);
  /* Inline banner feedback. An "error" (offline/permission) is assertive and
   * shakes the row keyed by `shakeId`; a "success" (copy) is polite. Both clear
   * on close, on new input, and on entering/leaving the drill-in. The shake
   * also self-clears on a timer (effect below) so it resets under
   * prefers-reduced-motion, where the animation — and animationend — never fire. */
  const [feedback, setFeedback] = useState<Feedback | null>(null);
  const [shakeId, setShakeId] = useState<string | null>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  const { devices } = useDevices({ page: 1, perPage: 50, status: "accepted" });
  const terminalSessions = useTerminalStore((s) => s.sessions);
  const restoreTerminal = useTerminalStore((s) => s.restore);
  const logout = useAuthStore((s) => s.logout);
  const tenant = useAuthStore((s) => s.tenant);
  const canConnect = useHasPermission("device:connect");
  const { namespace } = useNamespace(tenant ?? "");
  const nsName = namespace?.name ?? "";
  // The green banner is our copy confirmation, so useCopy's own `copied` flag
  // (its per-button "Copied!" affordance) is intentionally unused here.
  const { copy } = useCopy();

  /* The drilled-in device, resolved from the live list. Deriving drill-in state
   * from the *resolved* device means a drillInUid whose device left the list
   * (e.g. a refetch) transparently falls back to the device list — no
   * self-healing effect needed. */
  const drillDevice = drillInUid
    ? (devices.find((d) => d.uid === drillInUid) ?? null)
    : null;
  const isDrilledIn = drillDevice !== null;

  /* Keep focus on the input across drill-in transitions — a mouse click on the
   * chevron or back button moves focus to that control, which then unmounts.
   * (Initial open focus is owned by BaseDialog's focus trap; this only fires on
   * drillInUid changes.) */
  useEffect(() => {
    inputRef.current?.focus();
  }, [drillInUid]);

  /* Self-clear the shake once its animation would have finished. Driven by a
   * timer rather than onAnimationEnd so it also fires under
   * prefers-reduced-motion (no animation → no animationend). A repeat reject on
   * the same row within this window keeps the existing timer — the banner is
   * the primary signal — and a reject on a different row reschedules it. */
  useEffect(() => {
    if (!shakeId) return undefined;
    const timer = setTimeout(() => setShakeId(null), 450);
    return () => clearTimeout(timer);
  }, [shakeId]);

  /* Single dismissal path: clears local state and flips the shared open-state,
   * so every way of fully closing resets uniformly. */
  const close = useCallback(() => {
    setQuery("");
    setActiveIndex(0);
    setFeedback(null);
    setShakeId(null);
    setDrillInUid(null);
    closePalette();
  }, [closePalette]);

  const go = useCallback(
    (path: string) => {
      close();
      void navigate(path);
    },
    [close, navigate],
  );

  const onLogout = useCallback(() => {
    close();
    logout();
    void navigate("/login");
  }, [close, logout, navigate]);

  /* Reject an action inline: shake the row and show the assertive message; the
   * palette stays open. Backs both the permission and offline guards below. */
  const rejectRow = useCallback((rowId: string, message: string) => {
    setFeedback({ kind: "error", text: message });
    setShakeId(rowId);
  }, []);

  /* Copy to clipboard and confirm inline; the palette stays open. The banner is
   * optimistic — useCopy is fire-and-forget and surfaces failures (insecure
   * context, denied permission) through its own warning dialog rather than a
   * return value, so we can't gate the banner on the outcome here. */
  const copyAction = useCallback(
    (value: string, label: string) => {
      copy(value);
      setFeedback({ kind: "success", text: `Copied ${label} to clipboard` });
    },
    [copy],
  );

  /* Enter/leave a device's action menu. Both reset the filter and highlight so
   * the secondary list starts clean. */
  const enterDrillIn = useCallback((uid: string) => {
    setDrillInUid(uid);
    setQuery("");
    setActiveIndex(0);
    setFeedback(null);
    setShakeId(null);
  }, []);

  const exitDrillIn = useCallback(() => {
    setDrillInUid(null);
    setQuery("");
    setActiveIndex(0);
    setFeedback(null);
    setShakeId(null);
  }, []);

  /* BaseDialog routes Escape (native cancel) and backdrop clicks here. When
   * drilled in, that means "go back one level"; otherwise close. */
  const handleDismiss = useCallback(() => {
    if (isDrilledIn) exitDrillIn();
    else close();
  }, [isDrilledIn, exitDrillIn, close]);

  /* Restore an open terminal for this device, else open the ConnectDrawer for
   * it (TerminalManager owns that drawer and reacts to reconnectTarget). Reads
   * fresh session state via getState(), mirroring the Devices page — including
   * its `device:connect` gate, which covers both connecting and restoring. An
   * offline device with no session can't be connected; an existing session
   * still restores when permitted. */
  const connectOrRestore = useCallback(
    (uid: string, name: string, online: boolean) => {
      if (!canConnect) {
        rejectRow(`device-${uid}`, NO_CONNECT_PERMISSION);
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
        rejectRow(`device-${uid}`, `${name} is offline — start it to connect`);
        return;
      }
      close();
      store.requestConnect(uid, name);
    },
    [canConnect, rejectRow, close],
  );

  /* Cmd/Ctrl+K toggles the palette. The store also backs the visible Sidebar
   * trigger; closing routes through close() so the query resets. */
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

  const connectionItems = useMemo(
    () =>
      buildConnectionItems({
        devices,
        terminalSessions,
        canConnect,
        connectOrRestore,
        restoreTerminal,
        rejectRow,
        enterDrillIn,
        close,
      }),
    [
      devices,
      terminalSessions,
      canConnect,
      connectOrRestore,
      restoreTerminal,
      rejectRow,
      enterDrillIn,
      close,
    ],
  );

  const commandItems = useMemo(
    () => buildCommandItems({ go, onLogout }),
    [go, onLogout],
  );

  /* Whether the drilled-in device has an open session to restore — lets its
   * action-menu Connect stay enabled even when the device is offline. */
  const hasOpenSession = drillDevice
    ? terminalSessions.some((s) => s.deviceUid === drillDevice.uid)
    : false;

  const deviceActionItems = useMemo(
    () =>
      buildDeviceActionItems({
        drillDevice,
        nsName,
        canConnect,
        hasOpenSession,
        connectOrRestore,
        copyAction,
        go,
      }),
    [
      drillDevice,
      nsName,
      canConnect,
      hasOpenSession,
      connectOrRestore,
      copyAction,
      go,
    ],
  );

  /* While drilled in, the device's actions take over and ">" is inert. Else ">"
   * gates page navigation; the default stays connection-first. Derived during
   * render — no extra state. */
  const trimmedQuery = query.trimStart();
  const commandMode = !drillDevice && trimmedQuery.startsWith(">");
  const term = commandMode ? trimmedQuery.slice(1).trim() : query.trim();
  const activeItems = drillDevice
    ? deviceActionItems
    : commandMode
      ? commandItems
      : connectionItems;

  const filtered = useMemo(() => {
    if (!term) return activeItems;
    return activeItems.filter(
      (item) =>
        fuzzyMatch(term, item.label) ||
        (item.sublabel && fuzzyMatch(term, item.sublabel)) ||
        fuzzyMatch(term, item.section),
    );
  }, [activeItems, term]);

  const sections = useMemo(() => {
    const map = new Map<string, CommandItem[]>();
    filtered.forEach((item) => {
      const existing = map.get(item.section);
      if (existing) existing.push(item);
      else map.set(item.section, [item]);
    });
    return map;
  }, [filtered]);

  const flatList = useMemo(() => {
    const flat: CommandItem[] = [];
    sections.forEach((items) => flat.push(...items));
    return flat;
  }, [sections]);

  /* Map item id → flat index (avoids a render-order-coupled counter). */
  const indexById = useMemo(() => {
    const map = new Map<string, number>();
    flatList.forEach((item, i) => map.set(item.id, i));
    return map;
  }, [flatList]);

  /* Derive the in-range active index during render (no setState-in-render). */
  const safeIndex = flatList.length
    ? Math.min(activeIndex, flatList.length - 1)
    : -1;
  const activeItem = safeIndex >= 0 ? flatList[safeIndex] : undefined;

  /* Scroll active into view. */
  useEffect(() => {
    if (!listRef.current) return;
    const active = listRef.current.querySelector("[data-active=true]");
    active?.scrollIntoView({ block: "nearest" });
  }, [safeIndex]);

  /* List navigation + selection. Escape, Tab, and backdrop dismissal are owned
   * by BaseDialog (native <dialog>). Focus stays on the input (combobox), so
   * this handler lives on the input. →/← drill in and out, but only when the
   * caret is at the matching edge so text editing still works. */
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      const len = flatList.length;
      const move = (delta: number) =>
        setActiveIndex((prev) => (Math.min(prev, len - 1) + delta + len) % len);
      const input = e.currentTarget;
      const caretAtEnd =
        input.selectionStart === input.selectionEnd &&
        input.selectionStart === input.value.length;
      const caretAtStart =
        input.selectionStart === input.selectionEnd &&
        input.selectionStart === 0;

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
    },
    [flatList, safeIndex, isDrilledIn, exitDrillIn],
  );

  /* Composite input handler: update the query and reset highlight + feedback. */
  const onQueryChange = useCallback((value: string) => {
    setQuery(value);
    setActiveIndex(0);
    setFeedback(null);
    setShakeId(null);
  }, []);

  const hasResults = flatList.length > 0;

  return {
    open,
    inputRef,
    listRef,
    query,
    drillDevice,
    commandMode,
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
