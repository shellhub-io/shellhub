import { describe, expect, it, beforeEach } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useLayoutStore from "@/store/modules/layout";

describe("Layout Store", () => {
  setActivePinia(createPinia());
  const layoutStore = useLayoutStore();

  beforeEach(() => {
    localStorage.clear();
  });

  it("should have initial state values", () => {
    expect(layoutStore.layout).toEqual(undefined);
    expect(layoutStore.theme).toEqual("dark");
  });

  it("should initialize with theme from localStorage if available", () => {
    localStorage.setItem("theme", "light");

    setActivePinia(createPinia());
    const freshLayoutStore = useLayoutStore();

    expect(freshLayoutStore.theme).toEqual("light");
  });

  it("should persist theme changes to localStorage", () => {
    layoutStore.setTheme("light");
    expect(localStorage.getItem("theme")).toEqual("light");

    layoutStore.setTheme("dark");
    expect(localStorage.getItem("theme")).toEqual("dark");
  });
});
