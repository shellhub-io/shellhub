import { describe, expect, it, beforeEach } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useSpinnerStore from "@/store/modules/spinner";

describe("Spinner Store", () => {
  let spinnerStore: ReturnType<typeof useSpinnerStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    spinnerStore = useSpinnerStore();
  });

  describe("Initial State", () => {
    it("should have status as false by default", () => {
      expect(spinnerStore.status).toBe(false);
    });
  });

  describe("State Mutations", () => {
    it("should update status to true", () => {
      spinnerStore.status = true;

      expect(spinnerStore.status).toBe(true);
    });

    it("should update status to false", () => {
      spinnerStore.status = true;
      spinnerStore.status = false;

      expect(spinnerStore.status).toBe(false);
    });
  });
});
