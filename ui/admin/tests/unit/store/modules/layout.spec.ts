import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useLayoutStore from "@admin/store/modules/layout";

describe("Layout Pinia Store", () => {
  setActivePinia(createPinia());
  const layoutStore = useLayoutStore();

  beforeEach(() => { localStorage.clear(); });

  it("returns default layout and theme", () => {
    expect(layoutStore.layout).toEqual("AppLayout");
    expect(layoutStore.theme).toEqual("dark");
  });

  it("updates layout", () => {
    layoutStore.layout = "SimpleLayout";
    expect(layoutStore.layout).toEqual("SimpleLayout");
  });

  it("updates theme ref and localStorage with setTheme", () => {
    layoutStore.setTheme("light");
    expect(layoutStore.theme).toEqual("light");
    expect(localStorage.getItem("theme")).toEqual("light");
  });
});
