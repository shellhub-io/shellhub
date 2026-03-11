import { create } from "zustand";
import axios from "axios";

export interface TerminalThemeColors {
  background: string;
  foreground: string;
  cursor?: string;
  cursorAccent?: string;
  selectionBackground?: string;
  black?: string;
  red?: string;
  green?: string;
  yellow?: string;
  blue?: string;
  magenta?: string;
  cyan?: string;
  white?: string;
  brightBlack?: string;
  brightRed?: string;
  brightGreen?: string;
  brightYellow?: string;
  brightBlue?: string;
  brightMagenta?: string;
  brightCyan?: string;
  brightWhite?: string;
}

export interface TerminalTheme {
  name: string;
  dark: boolean;
  preview: { background: string; foreground: string };
  colors: TerminalThemeColors;
}

interface ThemeMetadata {
  name: string;
  file: string;
  dark: boolean;
  preview: { background: string; foreground: string };
}

export const TERMINAL_FONTS = [
  "IBM Plex Mono",
  "JetBrains Mono",
  "Fira Code",
  "Source Code Pro",
  "Ubuntu Mono",
  "Inconsolata",
  "Monaco",
  "Menlo",
  "monospace",
] as const;

export type TerminalFont = (typeof TERMINAL_FONTS)[number];

const STORAGE_KEYS = {
  theme: "terminalTheme",
  fontFamily: "terminalFontFamily",
  fontSize: "terminalFontSize",
};

const FALLBACK_THEME: TerminalTheme = {
  name: "ShellHub Dark",
  dark: true,
  preview: { background: "#18191B", foreground: "#667ACC" },
  colors: {
    background: "#18191B",
    foreground: "#E1E4EA",
    cursor: "#667ACC",
    cursorAccent: "#18191B",
    selectionBackground: "#667ACC40",
    black: "#1E2127",
    red: "#ca6169",
    green: "#82a568",
    yellow: "#bf8c5d",
    blue: "#56a2e1",
    magenta: "#b07cc8",
    cyan: "#4e9aa3",
    white: "#E1E4EA",
    brightBlack: "#5C6070",
    brightRed: "#d9787f",
    brightGreen: "#99ba82",
    brightYellow: "#d4a676",
    brightBlue: "#72b6ed",
    brightMagenta: "#c495d6",
    brightCyan: "#68b2ba",
    brightWhite: "#f0ede8",
  },
};

function normalizeColors(raw: Record<string, string>): TerminalThemeColors {
  const { selection, ...rest } = raw;
  return {
    ...rest,
    selectionBackground: selection || rest.selectionBackground,
  } as TerminalThemeColors;
}

interface TerminalThemeState {
  themes: TerminalTheme[];
  themeName: string;
  theme: TerminalTheme;
  fontFamily: TerminalFont;
  fontSize: number;
  fontFamilyWithFallback: string;
  loaded: boolean;
  loadThemes: () => Promise<void>;
  setTheme: (name: string) => void;
  setFontFamily: (font: TerminalFont) => void;
  setFontSize: (size: number) => void;
}

export const useTerminalThemeStore = create<TerminalThemeState>((set, get) => {
  const savedTheme = localStorage.getItem(STORAGE_KEYS.theme);
  const savedFont = localStorage.getItem(STORAGE_KEYS.fontFamily);
  const savedSize = localStorage.getItem(STORAGE_KEYS.fontSize);

  const initialThemeName = savedTheme || "ShellHub Dark";
  const initialFont = (savedFont as TerminalFont) || "IBM Plex Mono";
  const initialSize = savedSize ? parseInt(savedSize, 10) : 14;

  return {
    themes: [FALLBACK_THEME],
    themeName: initialThemeName,
    theme: FALLBACK_THEME,
    fontFamily: initialFont,
    fontSize: initialSize,
    fontFamilyWithFallback: `'${initialFont}', monospace`,
    loaded: false,

    loadThemes: async () => {
      if (get().loaded) return;

      try {
        const { data: metadata } = await axios.get<ThemeMetadata[]>("/xterm-themes/metadata.json");

        const results = await Promise.all(
          metadata.map(async (meta) => {
            try {
              const { data: raw } = await axios.get<Record<string, string>>(`/xterm-themes/${meta.file}`);
              return {
                name: meta.name,
                dark: meta.dark,
                preview: meta.preview,
                colors: normalizeColors(raw),
              } as TerminalTheme;
            } catch {
              return null;
            }
          }),
        );

        const themes = results.filter(Boolean) as TerminalTheme[];
        const current = themes.find((t) => t.name === get().themeName) || themes[0] || FALLBACK_THEME;
        set({ themes, theme: current, loaded: true });
      } catch {
        set({ loaded: true });
      }
    },

    setTheme: (name) => {
      const found = get().themes.find((t) => t.name === name);
      if (found) {
        localStorage.setItem(STORAGE_KEYS.theme, name);
        set({ themeName: name, theme: found });
      }
    },

    setFontFamily: (font) => {
      localStorage.setItem(STORAGE_KEYS.fontFamily, font);
      set({ fontFamily: font, fontFamilyWithFallback: `'${font}', monospace` });
    },

    setFontSize: (size) => {
      const clamped = Math.min(Math.max(size, 8), 24);
      localStorage.setItem(STORAGE_KEYS.fontSize, clamped.toString());
      set({ fontSize: clamped });
    },
  };
});
