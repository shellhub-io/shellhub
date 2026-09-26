import { create } from "zustand";

/**
 * How the sidebar sits on a desktop window: auto pins it open on a wide window and folds it to a
 * rail on a narrower one; pinned and rail hold it open or folded whatever the width.
 */
export type SidebarPin = "auto" | "pinned" | "rail";

const STORAGE_KEY = "sidebarPinned";

function readPin(): SidebarPin {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored === null ? "auto" : stored === "true" ? "pinned" : "rail";
  } catch {
    return "auto";
  }
}

function storePin(pin: SidebarPin) {
  try {
    if (pin === "auto") localStorage.removeItem(STORAGE_KEY);
    else localStorage.setItem(STORAGE_KEY, String(pin === "pinned"));
  } catch {
    return;
  }
}

interface SidebarState {
  pin: SidebarPin;
  setPin: (pin: SidebarPin) => void;
}

/**
 * The sidebar pin preference, shared by the layout's pin button and the Appearance settings, and
 * remembered in this browser under the key the layout has always used.
 */
export const useSidebarStore = create<SidebarState>((set) => ({
  pin: readPin(),
  setPin: (pin) => {
    storePin(pin);
    set({ pin });
  },
}));
