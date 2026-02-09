import { defineStore } from "pinia";
import { computed, ref } from "vue";
import axios from "axios";
import { ITerminalTheme, IThemeMetadata } from "@/interfaces/ITerminal";
import handleError from "@/utils/handleError";

const fallbackTheme: ITerminalTheme = {
  name: "Default",
  description: "Fallback theme",
  colors: {
    background: "#0f1526",
    foreground: "#ffffff",
  },
};

const terminalFontFamilies = [
  "Monospace",
  "Source Code Pro",
  "Inconsolata",
  "Ubuntu Mono",
  "Fira Code",
  "Anonymous Pro",
  "JetBrains Mono",
  "Noto Mono",
] as const;

export type TerminalFontFamily = typeof terminalFontFamilies[number];

const useTerminalThemeStore = defineStore("terminalTheme", () => {
  const availableFonts = ref(terminalFontFamilies);
  const currentFontFamily = ref<TerminalFontFamily>(localStorage.getItem("terminalFontFamily") as TerminalFontFamily || "Monospace");
  const currentFontSize = ref<number>(parseInt(localStorage.getItem("terminalFontSize") || "15", 10));

  const setFontSettings = async (fontFamily: TerminalFontFamily, fontSize: number) => {
    try {
      if (fontFamily !== "Monospace") await document.fonts.load(`${fontSize}px '${fontFamily}'`);
      currentFontFamily.value = fontFamily;
      currentFontSize.value = fontSize;
      localStorage.setItem("terminalFontFamily", fontFamily);
      localStorage.setItem("terminalFontSize", fontSize.toString());
    } catch (error) {
      // Font failed to load, keep current settings
      handleError(new Error(`Failed to load font ${fontFamily}:`, { cause: error }));
    }
  };

  const loadInitialFont = async () => {
    // No need to load the browser's default monospace font
    if (currentFontFamily.value === "Monospace") return;

    try {
      await document.fonts.load(`${currentFontSize.value}px '${currentFontFamily.value}'`);
    } catch {
      currentFontFamily.value = "Monospace";
      localStorage.setItem("terminalFontFamily", "Monospace");
    }
  };

  const terminalThemes = ref<ITerminalTheme[]>();
  const currentThemeName = ref<string>(localStorage.getItem("terminalTheme") || "ShellHub Dark");
  const currentTheme = computed(() => terminalThemes.value?.find((theme) => theme.name === currentThemeName.value) || fallbackTheme);

  const loadThemes = async () => {
    if (terminalThemes.value?.length) return; // Themes already loaded

    try {
      const { data: metadata } = await axios.get<IThemeMetadata[]>("/xterm-themes/metadata.json");

      const themePromises = metadata.map(async (meta) => {
        try {
          const { data: themeColors } = await axios.get<ITerminalTheme["colors"]>(`/xterm-themes/${meta.file}`);

          return {
            name: meta.name,
            description: meta.dark ? "Dark theme" : "Light theme",
            colors: themeColors,
          };
        } catch (error) {
          console.warn(`Failed to load theme ${meta.name}:`, error);
          return null;
        }
      });

      const themes = (await Promise.all(themePromises)).filter(Boolean) as ITerminalTheme[];
      terminalThemes.value = themes;
    } catch (error) {
      handleError(new Error("Failed to load terminal themes.", { cause: error }));
      currentThemeName.value = fallbackTheme.name;
      terminalThemes.value = [fallbackTheme];
    }
  };

  const setTheme = (themeName: string) => {
    currentThemeName.value = themeName;
    localStorage.setItem("terminalTheme", themeName);
  };

  return {
    availableFonts,
    currentFontFamily,
    currentFontSize,
    setFontSettings,
    loadInitialFont,

    terminalThemes,
    currentThemeName,
    currentTheme,
    loadThemes,
    setTheme,
  };
});

export default useTerminalThemeStore;
