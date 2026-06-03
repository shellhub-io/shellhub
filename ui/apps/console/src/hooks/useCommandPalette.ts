import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useDevices } from "@/hooks/useDevices";
import { useTerminalStore } from "@/stores/terminalStore";
import { useAuthStore } from "@/stores/authStore";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";
import {
  buildItems,
  fuzzyMatch,
  type CommandItem,
} from "@/components/commandPalette/items";

/** The view-model the palette shell and its presentational parts consume. */
export interface CommandPaletteViewModel {
  open: boolean;
  // Ref the JSX attaches; the hook reads `.current` only inside effects.
  listRef: React.RefObject<HTMLDivElement>;
  // Derived view data (computed during render).
  query: string;
  sections: Map<string, CommandItem[]>;
  hasResults: boolean;
  indexById: Map<string, number>;
  safeIndex: number;
  activeItem: CommandItem | undefined;
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
  const listRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  const { devices } = useDevices({ page: 1, perPage: 50, status: "accepted" });
  const terminalSessions = useTerminalStore((s) => s.sessions);
  const restoreTerminal = useTerminalStore((s) => s.restore);
  const logout = useAuthStore((s) => s.logout);

  /* Single dismissal path: clears the local query/highlight and flips the
   * shared open-state, so every way of closing (Escape via BaseDialog, the
   * shortcut) resets uniformly. */
  const close = useCallback(() => {
    setQuery("");
    setActiveIndex(0);
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

  const items = useMemo(
    () =>
      buildItems({
        devices,
        terminalSessions,
        go,
        close,
        restoreTerminal,
        onLogout,
      }),
    [devices, terminalSessions, go, close, restoreTerminal, onLogout],
  );

  const filtered = useMemo(() => {
    if (!query.trim()) return items;
    return items.filter(
      (item) =>
        fuzzyMatch(query, item.label) ||
        (item.sublabel && fuzzyMatch(query, item.sublabel)) ||
        fuzzyMatch(query, item.section),
    );
  }, [items, query]);

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

  /* Composite input handler: update the query and reset the highlight. */
  const onQueryChange = useCallback((value: string) => {
    setQuery(value);
    setActiveIndex(0);
  }, []);

  const hasResults = flatList.length > 0;

  return {
    open,
    listRef,
    query,
    sections,
    hasResults,
    indexById,
    safeIndex,
    activeItem,
    onQueryChange,
    setActiveIndex,
    handleKeyDown,
    close,
  };
}
