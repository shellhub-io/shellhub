import { create } from "zustand";

/**
 * The sidebar pins in the order the Appearance settings offer them, with the label each shows.
 */
export const SIDEBAR_PINS = [
  { value: "auto", label: "Automatic" },
  { value: "pinned", label: "Pinned open" },
  { value: "rail", label: "Folded" },
] as const;

/**
 * How the sidebar sits on a desktop window: auto pins it open on a wide window and folds it to a
 * rail on a narrower one; pinned and rail hold it open or folded whatever the width. Derived from
 * SIDEBAR_PINS, so a pin cannot exist without its tile.
 */
export type SidebarPin = (typeof SIDEBAR_PINS)[number]["value"];

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
