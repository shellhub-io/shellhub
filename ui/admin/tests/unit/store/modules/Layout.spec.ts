import { describe, expect, it } from "vitest";
import { store } from "../../../../src/store";

describe("Layout", () => {
  const layout = "defaultLayout";
  const darkModeStatus = false;

  it("Return firewall default variables", () => {
    expect(store.getters["layout/getLayout"]).toEqual("appLayout");
    expect(store.getters["layout/getStatusDarkMode"]).toEqual("dark");
  });
  it("Verify initial state change for setLayout mutation", () => {
    store.dispatch("layout/setLayout", layout);
    expect(store.getters["layout/getLayout"]).toEqual(layout);
  });
  it("Verify initial state change for setLayout mutation", () => {
    store.dispatch("layout/setStatusDarkMode", darkModeStatus);
    expect(store.getters["layout/getStatusDarkMode"]).toEqual("light");
  });
});
