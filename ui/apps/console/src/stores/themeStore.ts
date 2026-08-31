import { create } from "zustand";

/**
 * The console's colour scheme. Dark is the default, and the app chrome stays dark in either.
 */
export type AppTheme = "dark" | "light";

const STORAGE_KEY = "appTheme";

function resolveInitialTheme(): AppTheme {
  const saved = localStorage.getItem(STORAGE_KEY);
  if (saved === "light" || saved === "dark") return saved;
  return "dark";
}

function applyTheme(theme: AppTheme) {
  document.documentElement.classList.toggle("light", theme === "light");
}

interface ThemeState {
  theme: AppTheme;
  setTheme: (theme: AppTheme) => void;
  toggleTheme: () => void;
}

/**
 * The active theme. Setting it writes the class onto <html> as well as persisting the choice,
 * because the design-system tokens resolve from that class rather than from React state.
 */
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

applyTheme(resolveInitialTheme());
