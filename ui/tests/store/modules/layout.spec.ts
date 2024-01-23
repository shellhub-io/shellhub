import { describe, expect, it } from "vitest";
import { store } from "../../../src/store";

describe("Layout Store", () => {
  it("Return layout with default variables", () => {
    expect(store.getters["layout/getLayout"]).toEqual("appLayout");
    expect(store.getters["layout/getStatusDarkMode"]).toEqual("dark");
  });

  it("Mutates layout state", async () => {
    await store.dispatch("layout/setLayout", "newLayout");
    expect(store.getters["layout/getLayout"]).toEqual("newLayout");
  });

  it("Mutates statusDarkMode state", () => {
    store.commit("layout/setStatusDarkMode", "light");
    expect(store.getters["layout/getStatusDarkMode"]).toEqual("light");
  });

  it("Commits layout mutation", async () => {
    await store.dispatch("layout/setLayout", "newLayout");
    expect(store.getters["layout/getLayout"]).toEqual("newLayout");
  });

  it("Commits statusDarkMode mutation", async () => {
    await store.dispatch("layout/setStatusDarkMode", true);
    expect(store.getters["layout/getStatusDarkMode"]).toEqual("dark");
  });
});
