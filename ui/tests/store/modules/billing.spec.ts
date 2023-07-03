import { describe, expect, it } from "vitest";
import { store } from "../../../src/store";

describe("Billing", () => {
  const customerId = "test_customer123";
  const active = true;

  ///////
  // In this case, the default state of the variables is checked.
  ///////

  it("Return namespace default variables", () => {
    expect(store.getters["billing/get"]).toEqual({});
    expect(store.getters["billing/active"]).toEqual(!active);
    expect(store.getters["billing/status"]).toEqual("inactive");
    expect(store.getters["billing/invoices"]).toEqual([]);
  });

  it("Verify initial state change for setSubscription mutation", () => {
    const data = {
      active: true,
      status: "active",
      customer_id: customerId,
      subscription_id: "sub_test123",
      current_period_end: 12121,
      created_at: 12023,
    };

    store.commit("billing/setSubscription", data);

    ["status", "customer", "id"].map((v) => (
      Reflect.deleteProperty(data, v)
    ));

    expect(store.getters["billing/get"]).toEqual({
      ...data,
    });
  });
});
