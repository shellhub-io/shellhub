import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

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
