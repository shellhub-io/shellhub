import { createPinia, setActivePinia } from "pinia";
import { describe, expect, it } from "vitest";
import useBillingStore from "@/store/modules/billing";

describe("Billing", () => {
  setActivePinia(createPinia());
  const billingStore = useBillingStore();

  it("Return namespace default variables", () => {
    const { billing, isActive, status } = billingStore;
    expect(billing).toEqual({});
    expect(isActive).toEqual(false);
    expect(status).toEqual("inactive");
    expect(billingStore.invoices).toEqual([]);
  });
});
