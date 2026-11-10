import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import type { TerminalThemeColors } from "@/stores/terminalThemeStore";

const THEMES_DIR = join(__dirname, "../../../public/xterm-themes");
const readJson = <T>(file: string) =>
  JSON.parse(readFileSync(join(THEMES_DIR, file), "utf8")) as T;

const STORAGE_KEY = "terminalFontSize";

beforeEach(() => {
  localStorage.clear();
});

afterEach(() => {
  localStorage.clear();
  vi.resetModules();
});

describe("terminalThemeStore font size", () => {
  describe("initialization", () => {
    it("starts at the default size when nothing is persisted", async () => {
      const { useTerminalThemeStore, DEFAULT_FONT_SIZE } = await import("@/stores/terminalThemeStore");

      expect(useTerminalThemeStore.getState().fontSize).toBe(DEFAULT_FONT_SIZE);
    });

    it("restores a persisted size", async () => {
      localStorage.setItem(STORAGE_KEY, "18");
      const { useTerminalThemeStore } = await import("@/stores/terminalThemeStore");

      expect(useTerminalThemeStore.getState().fontSize).toBe(18);
    });

    it("falls back to the default when the persisted value is not a number", async () => {
      localStorage.setItem(STORAGE_KEY, "not-a-number");
      const { useTerminalThemeStore, DEFAULT_FONT_SIZE } = await import("@/stores/terminalThemeStore");

      expect(useTerminalThemeStore.getState().fontSize).toBe(DEFAULT_FONT_SIZE);
    });

    it("clamps a persisted size that is out of range", async () => {
      localStorage.setItem(STORAGE_KEY, "999");
      const { useTerminalThemeStore, MAX_FONT_SIZE } = await import("@/stores/terminalThemeStore");

      expect(useTerminalThemeStore.getState().fontSize).toBe(MAX_FONT_SIZE);
    });
  });

  describe("setFontSize", () => {
    it("applies and persists a size inside the allowed range", async () => {
      const { useTerminalThemeStore } = await import("@/stores/terminalThemeStore");

      useTerminalThemeStore.getState().setFontSize(16);

      expect(useTerminalThemeStore.getState().fontSize).toBe(16);
      expect(localStorage.getItem(STORAGE_KEY)).toBe("16");
    });

    it("clamps a size below the minimum", async () => {
      const { useTerminalThemeStore, MIN_FONT_SIZE } = await import("@/stores/terminalThemeStore");

      useTerminalThemeStore.getState().setFontSize(MIN_FONT_SIZE - 5);

      expect(useTerminalThemeStore.getState().fontSize).toBe(MIN_FONT_SIZE);
      expect(localStorage.getItem(STORAGE_KEY)).toBe(String(MIN_FONT_SIZE));
    });

    it("clamps a size above the maximum", async () => {
      const { useTerminalThemeStore, MAX_FONT_SIZE } = await import("@/stores/terminalThemeStore");

      useTerminalThemeStore.getState().setFontSize(MAX_FONT_SIZE + 5);

      expect(useTerminalThemeStore.getState().fontSize).toBe(MAX_FONT_SIZE);
      expect(localStorage.getItem(STORAGE_KEY)).toBe(String(MAX_FONT_SIZE));
    });
  });
});

const HEX = /^#[0-9a-f]{6}$/i;
const HEX_WITH_ALPHA = /^#[0-9a-f]{6}([0-9a-f]{2})?$/i;
const ANSI = [
  "black",
  "red",
  "green",
  "yellow",
  "blue",
  "magenta",
  "cyan",
  "white",
  "brightBlack",
  "brightRed",
  "brightGreen",
  "brightYellow",
  "brightBlue",
  "brightMagenta",
  "brightCyan",
  "brightWhite",
] as const;

describe("bundled theme files", () => {
  const metadata = readJson<{ name: string; file: string }[]>("metadata.json");

  it("ship ShellHub Dark as the theme the store starts with", async () => {
    const { useTerminalThemeStore } = await import("@/stores/terminalThemeStore");

    expect(useTerminalThemeStore.getState().theme.colors).toEqual(
      readJson<TerminalThemeColors>("shellhub_dark.json"),
    );
  });

  it.each(metadata)("$name sets every colour itself", ({ file }) => {
    const colors = readJson<TerminalThemeColors>(file);

    for (const key of ["background", "foreground", "cursor", "cursorAccent"] as const) {
      expect(colors[key], key).toMatch(HEX);
    }
    expect(colors.selectionBackground, "selectionBackground").toMatch(HEX_WITH_ALPHA);
    for (const key of ANSI) expect(colors[key], key).toMatch(HEX);
  });
});

const night = { background: "#101010", foreground: "#e0e0e0" };
const paper = { background: "#f8f8f8", foreground: "#202020" };

function serveThemes(files: Record<string, TerminalThemeColors | null>) {
  const requested: string[] = [];
  server.use(
    http.get("*/xterm-themes/metadata.json", () =>
      HttpResponse.json(
        Object.keys(files).map((file) => ({
          name: file.replace(".json", ""),
          file,
          dark: true,
        })),
      ),
    ),
    http.get("*/xterm-themes/:file", ({ params }) => {
      requested.push(String(params.file));
      const colors = files[String(params.file)];
      return colors
        ? HttpResponse.json(colors)
        : new HttpResponse(null, { status: 500 });
    }),
  );
  return requested;
}

describe("loadThemes", () => {
  it("adopts the first theme when the saved one left the index", async () => {
    localStorage.setItem("terminalTheme", "Homebrew");
    serveThemes({ "night.json": night, "paper.json": paper });
    const { useTerminalThemeStore } = await import("@/stores/terminalThemeStore");

    await useTerminalThemeStore.getState().loadThemes();

    expect(useTerminalThemeStore.getState().themeName).toBe("night");
    expect(useTerminalThemeStore.getState().theme.colors).toEqual(night);
    expect(localStorage.getItem("terminalTheme")).toBe("night");
  });

  it("keeps the saved theme when only its file fails to load", async () => {
    localStorage.setItem("terminalTheme", "paper");
    serveThemes({ "night.json": night, "paper.json": null });
    const { useTerminalThemeStore } = await import("@/stores/terminalThemeStore");

    await useTerminalThemeStore.getState().loadThemes();

    expect(useTerminalThemeStore.getState().themeName).toBe("paper");
    expect(useTerminalThemeStore.getState().theme.colors).toEqual(night);
    expect(localStorage.getItem("terminalTheme")).toBe("paper");
  });

  it("keeps the restored colours when the saved theme's file fails on the second load", async () => {
    localStorage.setItem("terminalTheme", "paper");
    serveThemes({ "night.json": night, "paper.json": paper });
    const { useTerminalThemeStore } = await import("@/stores/terminalThemeStore");
    await useTerminalThemeStore.getState().restoreTheme();

    serveThemes({ "night.json": night, "paper.json": null });
    await useTerminalThemeStore.getState().loadThemes();

    expect(useTerminalThemeStore.getState().themeName).toBe("paper");
    expect(useTerminalThemeStore.getState().theme.colors).toEqual(paper);
  });
});

describe("restoreTheme", () => {
  it("fetches only the saved theme's file", async () => {
    localStorage.setItem("terminalTheme", "paper");
    const requested = serveThemes({ "night.json": night, "paper.json": paper });
    const { useTerminalThemeStore } = await import("@/stores/terminalThemeStore");

    await useTerminalThemeStore.getState().restoreTheme();

    expect(requested).toEqual(["paper.json"]);
    expect(useTerminalThemeStore.getState().theme.colors).toEqual(paper);
  });

  it("fetches nothing when the saved theme is the bundled default", async () => {
    const requested = serveThemes({ "night.json": night });
    const { useTerminalThemeStore } = await import("@/stores/terminalThemeStore");

    await useTerminalThemeStore.getState().restoreTheme();

    expect(requested).toEqual([]);
    expect(useTerminalThemeStore.getState().themeName).toBe("ShellHub Dark");
  });

  it("adopts the first theme when the saved one left the index", async () => {
    localStorage.setItem("terminalTheme", "Homebrew");
    serveThemes({ "night.json": night, "paper.json": paper });
    const { useTerminalThemeStore } = await import("@/stores/terminalThemeStore");

    await useTerminalThemeStore.getState().restoreTheme();

    expect(useTerminalThemeStore.getState().themeName).toBe("night");
    expect(useTerminalThemeStore.getState().theme.colors).toEqual(night);
    expect(localStorage.getItem("terminalTheme")).toBe("night");
  });

  it("keeps the saved name and the bundled colours when the saved file fails", async () => {
    localStorage.setItem("terminalTheme", "paper");
    serveThemes({ "paper.json": null });
    const { useTerminalThemeStore } = await import("@/stores/terminalThemeStore");

    await useTerminalThemeStore.getState().restoreTheme();

    expect(useTerminalThemeStore.getState().themeName).toBe("paper");
    expect(useTerminalThemeStore.getState().theme.name).toBe("ShellHub Dark");
    expect(useTerminalThemeStore.getState().loaded).toBe(false);
  });

  it("drops its result when the list loaded while it was fetching", async () => {
    localStorage.setItem("terminalTheme", "paper");
    serveThemes({ "night.json": night, "paper.json": paper });
    let release = () => {};
    const held = new Promise<void>((resolve) => (release = resolve));
    let markRequested = () => {};
    const requested = new Promise<void>((resolve) => (markRequested = resolve));
    server.use(
      http.get(
        "*/xterm-themes/paper.json",
        async () => {
          markRequested();
          await held;
          return HttpResponse.json(paper);
        },
        { once: true },
      ),
    );
    const { useTerminalThemeStore } = await import("@/stores/terminalThemeStore");

    const restoring = useTerminalThemeStore.getState().restoreTheme();
    await requested;
    await useTerminalThemeStore.getState().loadThemes();
    useTerminalThemeStore.getState().setTheme("night");
    release();
    await restoring;

    expect(useTerminalThemeStore.getState().themeName).toBe("night");
    expect(useTerminalThemeStore.getState().theme.colors).toEqual(night);
  });
});
