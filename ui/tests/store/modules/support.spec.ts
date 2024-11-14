import { describe, expect, it } from "vitest";
import { store } from "@/store";

describe("Support Store", () => {
  it("Returns support with default variables", () => {
    expect(store.getters["support/getIdentifier"]).toEqual("");
  });

  it("Commits support mutations", () => {
    store.commit("support/setIdentifier", "fake-identifier");
    expect(store.getters["support/getIdentifier"]).toEqual("fake-identifier");
  });
});
