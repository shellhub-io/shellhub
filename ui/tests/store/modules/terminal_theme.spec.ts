import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import axios from "axios";
import MockAdapter from "axios-mock-adapter";
import useTerminalThemeStore from "@/store/modules/terminal_theme";
import type { ITerminalTheme, IThemeMetadata } from "@/interfaces/ITerminal";
import handleError from "@/utils/handleError";

vi.mock("@/utils/handleError", () => ({
  default: vi.fn(),
}));

const mockHandleError = vi.mocked(handleError);
const mockConsoleWarn = vi.spyOn(console, "warn").mockImplementation(() => {});
const mockFontsLoad = vi.fn();
Object.defineProperty(document, "fonts", {
  value: { load: mockFontsLoad },
  writable: true,
});

const mockThemeMetadata = [
  { name: "ShellHub Dark", file: "shellhub_dark.json", dark: true },
  { name: "ShellHub Light", file: "shellhub_light.json", dark: false },
  { name: "Dracula", file: "dracula.json", dark: true },
] as IThemeMetadata[];

const mockShellHubDarkColors: ITerminalTheme["colors"] = {
  background: "#0f1526",
  foreground: "#ffffff",
  cursor: "#ffffff",
  black: "#000000",
  red: "#e06c75",
  green: "#98c379",
  yellow: "#d19a66",
  blue: "#61afef",
  magenta: "#c678dd",
  cyan: "#56b6c2",
  white: "#abb2bf",
  brightBlack: "#5c6370",
  brightRed: "#e06c75",
  brightGreen: "#98c379",
  brightYellow: "#d19a66",
  brightBlue: "#61afef",
  brightMagenta: "#c678dd",
  brightCyan: "#56b6c2",
  brightWhite: "#ffffff",
};

const mockShellHubLightColors: ITerminalTheme["colors"] = {
  background: "#ffffff",
  foreground: "#000000",
};

const mockDraculaColors: ITerminalTheme["colors"] = {
  background: "#282a36",
  foreground: "#f8f8f2",
};

describe("TerminalTheme Store", () => {
  let terminalThemeStore: ReturnType<typeof useTerminalThemeStore>;
  let mockAxios: MockAdapter;

  const mockSuccessfulThemeLoads = () => {
    mockAxios.onGet("/xterm-themes/metadata.json").reply(200, mockThemeMetadata);
    mockAxios.onGet("/xterm-themes/shellhub_dark.json").reply(200, mockShellHubDarkColors);
    mockAxios.onGet("/xterm-themes/shellhub_light.json").reply(200, mockShellHubLightColors);
    mockAxios.onGet("/xterm-themes/dracula.json").reply(200, mockDraculaColors);
  };

  beforeEach(() => {
    localStorage.clear();
    mockFontsLoad.mockClear();
    mockFontsLoad.mockResolvedValue(undefined);
    mockHandleError.mockClear();
    mockConsoleWarn.mockClear();
    setActivePinia(createPinia());
    terminalThemeStore = useTerminalThemeStore();
    mockAxios = new MockAdapter(axios);
  });

  afterEach(() => {
    mockAxios.reset();
    localStorage.clear();
  });

  describe("Initial State", () => {
    it("should have default font family as Monospace", () => {
      expect(terminalThemeStore.currentFontFamily).toBe("Monospace");
    });

    it("should have default font size as 15", () => {
      expect(terminalThemeStore.currentFontSize).toBe(15);
    });

    it("should have available fonts list", () => {
      expect(terminalThemeStore.availableFonts).toEqual([
        "Monospace",
        "Source Code Pro",
        "Inconsolata",
        "Ubuntu Mono",
        "Fira Code",
        "Anonymous Pro",
        "JetBrains Mono",
        "Noto Mono",
      ]);
    });

    it("should have default theme name as ShellHub Dark", () => {
      expect(terminalThemeStore.currentThemeName).toBe("ShellHub Dark");
    });

    it("should have undefined terminal themes initially", () => {
      expect(terminalThemeStore.terminalThemes).toBeUndefined();
    });

    it("should have fallback theme as current theme when themes not loaded", () => {
      expect(terminalThemeStore.currentTheme).toEqual({
        name: "Default",
        description: "Fallback theme",
        colors: {
          background: "#0f1526",
          foreground: "#ffffff",
        },
      });
    });

    it("should load font family from localStorage", () => {
      localStorage.setItem("terminalFontFamily", "Fira Code");
      localStorage.setItem("terminalFontSize", "18");

      setActivePinia(createPinia());
      const freshStore = useTerminalThemeStore();

      expect(freshStore.currentFontFamily).toBe("Fira Code");
      expect(freshStore.currentFontSize).toBe(18);
    });

    it("should load theme name from localStorage", () => {
      localStorage.setItem("terminalTheme", "Dracula");

      setActivePinia(createPinia());
      const freshStore = useTerminalThemeStore();

      expect(freshStore.currentThemeName).toBe("Dracula");
    });
  });

  describe("setFontSettings", () => {
    it("should set font family and size successfully", async () => {
      await terminalThemeStore.setFontSettings("Fira Code", 16);

      expect(terminalThemeStore.currentFontFamily).toBe("Fira Code");
      expect(terminalThemeStore.currentFontSize).toBe(16);
      expect(localStorage.getItem("terminalFontFamily")).toBe("Fira Code");
      expect(localStorage.getItem("terminalFontSize")).toBe("16");
      expect(mockFontsLoad).toHaveBeenCalledWith("16px 'Fira Code'");
    });

    it("should set Monospace font without loading", async () => {
      await terminalThemeStore.setFontSettings("Monospace", 14);

      expect(terminalThemeStore.currentFontFamily).toBe("Monospace");
      expect(terminalThemeStore.currentFontSize).toBe(14);
      expect(localStorage.getItem("terminalFontFamily")).toBe("Monospace");
      expect(localStorage.getItem("terminalFontSize")).toBe("14");
      expect(mockFontsLoad).not.toHaveBeenCalled();
    });

    it("should update font settings when called multiple times", async () => {
      await terminalThemeStore.setFontSettings("JetBrains Mono", 14);
      expect(terminalThemeStore.currentFontFamily).toBe("JetBrains Mono");

      await terminalThemeStore.setFontSettings("Source Code Pro", 18);
      expect(terminalThemeStore.currentFontFamily).toBe("Source Code Pro");
      expect(terminalThemeStore.currentFontSize).toBe(18);
    });

    it("should handle font loading failure gracefully", async () => {
      mockFontsLoad.mockRejectedValueOnce(new Error("Font not found"));
      const initialFont = terminalThemeStore.currentFontFamily;

      // @ts-expect-error Testing invalid font
      await terminalThemeStore.setFontSettings("Invalid Font", 16);

      expect(terminalThemeStore.currentFontFamily).toBe(initialFont);
      expect(mockHandleError).toHaveBeenCalledTimes(1);
      expect(mockHandleError).toHaveBeenCalledWith(
        expect.objectContaining({
          message: expect.stringContaining("Failed to load font"),
        }),
      );
    });
  });

  describe("loadInitialFont", () => {
    it("should not load Monospace font", async () => {
      terminalThemeStore.currentFontFamily = "Monospace";

      await terminalThemeStore.loadInitialFont();

      expect(mockFontsLoad).not.toHaveBeenCalled();
    });

    it("should load non-Monospace font successfully", async () => {
      localStorage.setItem("terminalFontFamily", "Fira Code");
      localStorage.setItem("terminalFontSize", "16");

      setActivePinia(createPinia());
      const freshStore = useTerminalThemeStore();

      await freshStore.loadInitialFont();

      expect(mockFontsLoad).toHaveBeenCalledWith("16px 'Fira Code'");
    });

    it("should fallback to Monospace when font loading fails", async () => {
      localStorage.setItem("terminalFontFamily", "Invalid Font");
      localStorage.setItem("terminalFontSize", "16");
      mockFontsLoad.mockRejectedValueOnce(new Error("Font not found"));

      setActivePinia(createPinia());
      const freshStore = useTerminalThemeStore();

      await freshStore.loadInitialFont();

      expect(freshStore.currentFontFamily).toBe("Monospace");
      expect(localStorage.getItem("terminalFontFamily")).toBe("Monospace");
    });
  });

  describe("loadThemes", () => {
    it("should load themes successfully", async () => {
      mockSuccessfulThemeLoads();

      await terminalThemeStore.loadThemes();

      expect(terminalThemeStore.terminalThemes).toHaveLength(3);
      expect(terminalThemeStore.terminalThemes?.[0]).toEqual({
        name: "ShellHub Dark",
        description: "Dark theme",
        colors: mockShellHubDarkColors,
      });
      expect(terminalThemeStore.terminalThemes?.[1]).toEqual({
        name: "ShellHub Light",
        description: "Light theme",
        colors: mockShellHubLightColors,
      });
    });

    it("should not reload themes if already loaded", async () => {
      mockSuccessfulThemeLoads();

      await terminalThemeStore.loadThemes();
      const firstLoad = terminalThemeStore.terminalThemes;

      await terminalThemeStore.loadThemes();

      expect(terminalThemeStore.terminalThemes).toBe(firstLoad);
      expect(mockAxios.history.get.filter((req) => req.url === "/xterm-themes/metadata.json")).toHaveLength(1);
    });

    it("should skip failed theme files and load successful ones", async () => {
      mockSuccessfulThemeLoads();
      mockAxios.onGet("/xterm-themes/shellhub_light.json").reply(404);

      await terminalThemeStore.loadThemes();

      expect(terminalThemeStore.terminalThemes).toHaveLength(2);
      expect(terminalThemeStore.terminalThemes?.map((t) => t.name)).toEqual([
        "ShellHub Dark",
        "Dracula",
      ]);
    });

    it("should set fallback theme when metadata loading fails", async () => {
      mockAxios.onGet("/xterm-themes/metadata.json").reply(500);

      await terminalThemeStore.loadThemes();

      expect(terminalThemeStore.currentThemeName).toBe("Default");
      expect(terminalThemeStore.terminalThemes).toEqual([
        {
          name: "Default",
          description: "Fallback theme",
          colors: {
            background: "#0f1526",
            foreground: "#ffffff",
          },
        },
      ]);
      expect(mockHandleError).toHaveBeenCalledTimes(1);
      expect(mockHandleError).toHaveBeenCalledWith(
        expect.objectContaining({
          message: "Failed to load terminal themes.",
        }),
      );
    });

    it("should handle network error when loading metadata", async () => {
      mockAxios.onGet("/xterm-themes/metadata.json").networkError();

      await terminalThemeStore.loadThemes();

      expect(terminalThemeStore.currentThemeName).toBe("Default");
      expect(terminalThemeStore.terminalThemes).toHaveLength(1);
      expect(mockHandleError).toHaveBeenCalledTimes(1);
    });
  });

  describe("setTheme", () => {
    it("should set theme name and persist to localStorage", () => {
      terminalThemeStore.setTheme("Dracula");

      expect(terminalThemeStore.currentThemeName).toBe("Dracula");
      expect(localStorage.getItem("terminalTheme")).toBe("Dracula");
    });

    it("should update theme multiple times", () => {
      terminalThemeStore.setTheme("Dracula");
      expect(terminalThemeStore.currentThemeName).toBe("Dracula");

      terminalThemeStore.setTheme("ShellHub Light");
      expect(terminalThemeStore.currentThemeName).toBe("ShellHub Light");
      expect(localStorage.getItem("terminalTheme")).toBe("ShellHub Light");
    });
  });

  describe("currentTheme computed", () => {
    it("should return fallback theme when themes not loaded", () => {
      expect(terminalThemeStore.currentTheme).toEqual({
        name: "Default",
        description: "Fallback theme",
        colors: {
          background: "#0f1526",
          foreground: "#ffffff",
        },
      });
    });

    it("should return correct theme after loading", async () => {
      mockSuccessfulThemeLoads();

      await terminalThemeStore.loadThemes();

      expect(terminalThemeStore.currentTheme).toEqual({
        name: "ShellHub Dark",
        description: "Dark theme",
        colors: mockShellHubDarkColors,
      });
    });

    it("should update when theme name changes", async () => {
      mockSuccessfulThemeLoads();

      await terminalThemeStore.loadThemes();

      terminalThemeStore.setTheme("Dracula");

      expect(terminalThemeStore.currentTheme).toEqual({
        name: "Dracula",
        description: "Dark theme",
        colors: mockDraculaColors,
      });
    });

    it("should return fallback when selected theme not found", async () => {
      mockSuccessfulThemeLoads();

      await terminalThemeStore.loadThemes();

      terminalThemeStore.setTheme("Non-Existent Theme");

      expect(terminalThemeStore.currentTheme).toEqual({
        name: "Default",
        description: "Fallback theme",
        colors: {
          background: "#0f1526",
          foreground: "#ffffff",
        },
      });
    });
  });
});
