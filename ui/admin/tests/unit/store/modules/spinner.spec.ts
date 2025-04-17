import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useSpinnerStore from "@admin/store/modules/spinner";

describe("Spinner Pinia Store", () => {
  let spinnerStore: ReturnType<typeof useSpinnerStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    spinnerStore = useSpinnerStore();
  });

  it("returns spinner default status", () => {
    expect(spinnerStore.getStatus).toBe(false);
  });

  it("sets spinner status to true", () => {
    spinnerStore.setStatus(true);
    expect(spinnerStore.getStatus).toBe(true);
  });

  it("sets spinner status to false", () => {
    spinnerStore.setStatus(false);
    expect(spinnerStore.getStatus).toBe(false);
  });
});
