import { create } from "zustand";

type PaletteView = "picker" | "palette" | "connect";

interface CommandPaletteState {
  view: PaletteView | null;
  show: (view: PaletteView) => void;
  close: () => void;
}

/**
 * Shared open-state for the palette surfaces, so the Cmd/Ctrl+K shortcut, the tab strip's "+"
 * and the picker's actions can move between them. Only one surface is open at a time. Ephemeral
 * by design: each surface resets on every open, so there is nothing to persist.
 */
export const useCommandPaletteStore = create<CommandPaletteState>((set) => ({
  view: null,
  show: (view) => set({ view }),
  close: () => set({ view: null }),
}));
