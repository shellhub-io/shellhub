import { create } from "zustand";

interface CommandPaletteState {
  open: boolean;
  openPalette: () => void;
  closePalette: () => void;
}

/**
 * Shared open-state for the command palette so both the Cmd/Ctrl+K shortcut
 * (in CommandPalette) and the visible Sidebar trigger can drive it. Ephemeral
 * by design — the palette resets on every open, so there is nothing to persist.
 *
 * Only `open` lives here; the palette's query/highlight stay local to the
 * component and are cleared by its own `close()` so every dismissal resets.
 */
export const useCommandPaletteStore = create<CommandPaletteState>((set) => ({
  open: false,
  openPalette: () => set({ open: true }),
  closePalette: () => set({ open: false }),
}));
