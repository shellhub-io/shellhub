import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useLayoutStore from "@/store/modules/layout";

describe("Layout Store", () => {
  let layoutStore: ReturnType<typeof useLayoutStore>;

  beforeEach(() => {
    localStorage.clear();
    setActivePinia(createPinia());
    layoutStore = useLayoutStore();
  });

  afterEach(() => { localStorage.clear(); });

  describe("Initial State", () => {
    it("should have undefined layout", () => {
      expect(layoutStore.layout).toBeUndefined();
    });

    it("should have dark theme as default", () => {
      expect(layoutStore.theme).toBe("dark");
    });

    it("should initialize with theme from localStorage when available", () => {
      localStorage.setItem("theme", "light");

      setActivePinia(createPinia());
      const freshLayoutStore = useLayoutStore();

      expect(freshLayoutStore.theme).toBe("light");
    });
  });

  describe("setTheme", () => {
    it("should update theme state to light", () => {
      layoutStore.setTheme("light");

      expect(layoutStore.theme).toBe("light");
      expect(localStorage.getItem("theme")).toBe("light");
    });

    it("should update theme state to dark", () => {
      layoutStore.setTheme("dark");

      expect(layoutStore.theme).toBe("dark");
      expect(localStorage.getItem("theme")).toBe("dark");
    });
  });
});
