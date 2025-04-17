import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useLayoutStore from "@admin/store/modules/layout";

describe("Layout Pinia Store", () => {
  let layoutStore: ReturnType<typeof useLayoutStore>;

  const initialLayout = "appLayout";
  const initialDarkMode = "dark";
  const newLayout = "defaultLayout";
  const darkModeStatus = false;

  beforeEach(() => {
    setActivePinia(createPinia());
    localStorage.clear();
    layoutStore = useLayoutStore();
  });

  it("returns default layout and dark mode", () => {
    expect(layoutStore.getLayout).toEqual(initialLayout);
    expect(layoutStore.getStatusDarkMode).toEqual(initialDarkMode);
  });

  it("updates layout when setLayout is called", () => {
    layoutStore.setLayout(newLayout);
    expect(layoutStore.getLayout).toEqual(newLayout);
  });

  it("updates dark mode status when setStatusDarkMode is called", () => {
    layoutStore.setStatusDarkMode(darkModeStatus);
    expect(layoutStore.getStatusDarkMode).toEqual("light");
    expect(localStorage.getItem("statusDarkMode")).toEqual("light");
  });
});
