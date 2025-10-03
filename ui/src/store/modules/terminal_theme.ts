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

const useTerminalThemeStore = defineStore("terminal-theme", () => {
  const terminalThemes = ref<ITerminalTheme[]>();
  const currentThemeName = ref<string>(localStorage.getItem("terminalTheme") || "ShellHub Dark");
  const currentTheme = computed(() => terminalThemes.value?.find((theme) => theme.name === currentThemeName.value) || fallbackTheme);

  const loadThemes = async () => {
    if (terminalThemes.value?.length) return; // Themes already loaded

    try {
      const { data: metadata } = await axios.get<IThemeMetadata[]>("/xterm-themes/metadata.json");

      const themePromises = metadata.map(async (meta) => {
        try {
          const { data: themeColors } = await axios.get(`/xterm-themes/${meta.file}`);

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

  return { terminalThemes, currentThemeName, currentTheme, loadThemes, setTheme };
});

export default useTerminalThemeStore;
