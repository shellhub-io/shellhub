import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useDevices } from "@/hooks/useDevices";
import { useTerminalStore } from "@/stores/terminalStore";
import { useAuthStore } from "@/stores/authStore";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";
import { useHasPermission } from "@/hooks/useHasPermission";
import {
  buildConnectionItems,
  buildCommandItems,
  fuzzyMatch,
  NO_CONNECT_PERMISSION,
  type CommandItem,
} from "@/components/commandPalette/items";

/** The view-model the palette shell and its presentational parts consume. */
export interface CommandPaletteViewModel {
  open: boolean;
  // Ref the JSX attaches; the hook reads `.current` only inside effects.
  listRef: React.RefObject<HTMLDivElement>;
  // Derived view data (computed during render).
  query: string;
  commandMode: boolean;
  sections: Map<string, CommandItem[]>;
  hasResults: boolean;
  indexById: Map<string, number>;
  safeIndex: number;
  activeItem: CommandItem | undefined;
  // Inline rejection feedback: an assertive message + the id of the row to shake.
  rejectMessage: string | null;
  shakeId: string | null;
  // Handlers.
  onQueryChange: (value: string) => void;
  setActiveIndex: (index: number) => void;
  handleKeyDown: (e: React.KeyboardEvent) => void;
  close: () => void;
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
  /* Inline rejection feedback (e.g. connecting to an offline device): an
   * assertive (role="alert") message plus the id of the row to shake. Both are
   * cleared on close and on new input; the shake also self-clears on a timer
   * (see the effect below) so it resets even under prefers-reduced-motion,
   * where the animation — and thus animationend — never fires. */
  const [rejectMessage, setRejectMessage] = useState<string | null>(null);
  const [shakeId, setShakeId] = useState<string | null>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  const { devices } = useDevices({ page: 1, perPage: 50, status: "accepted" });
  const terminalSessions = useTerminalStore((s) => s.sessions);
  const restoreTerminal = useTerminalStore((s) => s.restore);
  const logout = useAuthStore((s) => s.logout);
  const canConnect = useHasPermission("device:connect");

  /* Self-clear the shake once its animation would have finished. Driven by a
   * timer rather than onAnimationEnd so it also fires under
   * prefers-reduced-motion (no animation → no animationend). A repeat reject on
   * the same row within this window keeps the existing timer — the alert above
   * is the primary signal — and a reject on a different row reschedules it. */
  useEffect(() => {
    if (!shakeId) return undefined;
    const timer = setTimeout(() => setShakeId(null), 450);
    return () => clearTimeout(timer);
  }, [shakeId]);

  /* Single dismissal path: clears the local query/highlight and flips the
   * shared open-state, so every way of closing (Escape via BaseDialog, the
   * shortcut) resets uniformly. */
  const close = useCallback(() => {
    setQuery("");
    setActiveIndex(0);
    setRejectMessage(null);
    setShakeId(null);
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
    setRejectMessage(message);
    setShakeId(rowId);
  }, []);

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

  /* Default (connection-first) view: devices to connect/restore + open sessions. */
  const connectionItems = useMemo(
    () =>
      buildConnectionItems({
        devices,
        terminalSessions,
        canConnect,
        connectOrRestore,
        restoreTerminal,
        rejectRow,
        close,
      }),
    [
      devices,
      terminalSessions,
      canConnect,
      connectOrRestore,
      restoreTerminal,
      rejectRow,
      close,
    ],
  );

  /* Command mode (">" prefix): page navigation + account actions. */
  const commandItems = useMemo(
    () => buildCommandItems({ go, onLogout }),
    [go, onLogout],
  );

  /* ">" gates the page navigation behind command mode; the default view stays
   * connection-first. Derived during render — no extra state. */
  const trimmedQuery = query.trimStart();
  const commandMode = trimmedQuery.startsWith(">");
  const term = commandMode ? trimmedQuery.slice(1).trim() : query.trim();
  const activeItems = commandMode ? commandItems : connectionItems;

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
   * this handler lives on the input. */
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      const len = flatList.length;
      const move = (delta: number) =>
        setActiveIndex((prev) => (Math.min(prev, len - 1) + delta + len) % len);

      switch (e.key) {
        case "ArrowDown":
          e.preventDefault();
          if (len) move(1);
          break;
        case "ArrowUp":
          e.preventDefault();
          if (len) move(-1);
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
        case "Enter":
          e.preventDefault();
          if (safeIndex >= 0) flatList[safeIndex]?.onSelect();
          break;
      }
    },
    [flatList, safeIndex],
  );

  /* Composite input handler: update the query, reset the highlight, and clear
   * any standing rejection feedback. */
  const onQueryChange = useCallback((value: string) => {
    setQuery(value);
    setActiveIndex(0);
    setRejectMessage(null);
    setShakeId(null);
  }, []);

  const hasResults = flatList.length > 0;

  return {
    open,
    listRef,
    query,
    commandMode,
    sections,
    hasResults,
    indexById,
    safeIndex,
    activeItem,
    rejectMessage,
    shakeId,
    onQueryChange,
    setActiveIndex,
    handleKeyDown,
    close,
  };
}
