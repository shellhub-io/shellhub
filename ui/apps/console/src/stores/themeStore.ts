import { create } from "zustand";

export type AppTheme = "dark" | "light";

const STORAGE_KEY = "appTheme";

function resolveInitialTheme(): AppTheme {
  const saved = localStorage.getItem(STORAGE_KEY);
  if (saved === "light" || saved === "dark") return saved;
  return "dark";
}

function applyTheme(theme: AppTheme) {
  // Tokens default to dark in :root; the `light` class swaps the CSS variables.
  document.documentElement.classList.toggle("light", theme === "light");
}

interface ThemeState {
  theme: AppTheme;
  setTheme: (theme: AppTheme) => void;
  toggleTheme: () => void;
}

export const useThemeStore = create<ThemeState>((set, get) => ({
  theme: resolveInitialTheme(),

  setTheme: (theme) => {
    localStorage.setItem(STORAGE_KEY, theme);
    applyTheme(theme);
    set({ theme });
  },

  toggleTheme: () => {
    get().setTheme(get().theme === "dark" ? "light" : "dark");
  },
}));

// Apply the persisted theme as soon as the module loads, before first paint.
applyTheme(resolveInitialTheme());
