import { describe, expect, it } from "vitest";
import { store } from "../../../src/store";

describe("Mobile", () => {
  const mobile = false;

  it("Returns mobile default variable", () => {
    expect(store.getters["mobile/isMobile"]).toEqual(mobile);
  });
  it("Verify initial state change for setmobile mutation", () => {
    store.commit("mobile/setIsMobileStatus", !mobile);
    expect(store.getters["mobile/isMobile"]).toEqual(!mobile);
  });
});
