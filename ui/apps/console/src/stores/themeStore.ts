import { create } from "zustand";

/**
 * The console's colour scheme, applied to the whole page, chrome included.
 */
export type AppTheme = "dark" | "light";

/**
 * What the user asked for: a fixed scheme, or whatever the operating system is set to.
 */
export type ThemePreference = AppTheme | "system";

const STORAGE_KEY = "appTheme";

const systemLight =
  typeof window !== "undefined" && typeof window.matchMedia === "function"
    ? window.matchMedia("(prefers-color-scheme: light)")
    : undefined;

function systemTheme(): AppTheme {
  return systemLight?.matches ? "light" : "dark";
}

function resolve(preference: ThemePreference): AppTheme {
  return preference === "system" ? systemTheme() : preference;
}

function readPreference(): ThemePreference {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    return saved === "light" || saved === "dark" ? saved : "system";
  } catch {
    return "system";
  }
}

function storePreference(preference: ThemePreference) {
  try {
    if (preference === "system") localStorage.removeItem(STORAGE_KEY);
    else localStorage.setItem(STORAGE_KEY, preference);
  } catch {
    return;
  }
}

function applyTheme(theme: AppTheme) {
  document.documentElement.classList.toggle("light", theme === "light");
}

interface ThemeState {
  preference: ThemePreference;
  theme: AppTheme;
  setPreference: (preference: ThemePreference) => void;
}

/**
 * The active theme and the preference it comes from. Following the system is the default, and
 * the theme then changes with it while the console is open. Setting a preference writes the class
 * onto <html> as well as persisting the choice, because the design-system tokens resolve from
 * that class rather than from React state.
 */
export const useThemeStore = create<ThemeState>((set) => ({
  preference: readPreference(),
  theme: resolve(readPreference()),

  setPreference: (preference) => {
    storePreference(preference);
    const theme = resolve(preference);
    applyTheme(theme);
    set({ preference, theme });
  },
}));

systemLight?.addEventListener("change", () => {
  const { preference } = useThemeStore.getState();
  if (preference !== "system") return;
  const theme = systemTheme();
  applyTheme(theme);
  useThemeStore.setState({ theme });
});

applyTheme(useThemeStore.getState().theme);
