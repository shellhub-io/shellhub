import { describe, expect, it } from "vitest";
import { store } from "../../../../src/store";

describe("Spinner", () => {
  const status = "loading";

  it("Return firewall default variables", () => {
    expect(store.getters["spinner/status"]).toEqual(undefined);
  });
  it("Verify initial state change for setStatus mutation", () => {
    store.dispatch("spinner/setStatus", status);
    expect(store.getters["spinner/status"]).toEqual(status);
  });
});
