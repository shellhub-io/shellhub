import { describe, it, expect } from "vitest";
import { readNamespaceBilling, toCustomer, toSubscription } from "../billing";

describe("toCustomer", () => {
  it("returns null when input is undefined", () => {
    expect(toCustomer(undefined)).toBeNull();
  });

  it("returns null when input is null", () => {
    expect(toCustomer(null)).toBeNull();
  });

  it("returns null when input has no id", () => {
    expect(toCustomer({})).toBeNull();
  });

  it("normalizes a full customer object", () => {
    const customer = toCustomer({
      id: "cus_1",
      name: "Alice",
      email: "alice@example.com",
      payment_methods: [],
    });

    expect(customer).toEqual({
      id: "cus_1",
      name: "Alice",
      email: "alice@example.com",
      payment_methods: [],
    });
  });

  it("does not include cvc on the normalized payment method", () => {
    const customer = toCustomer({
      id: "cus_1",
      payment_methods: [
        {
          id: "pm_1",
          number: "xxxxxxxxxxxx4242",
          brand: "visa",
          exp_month: 12,
          exp_year: 2030,
          cvc: "xxx",
          default: true,
        },
      ],
    });

    expect(customer?.payment_methods[0]).not.toHaveProperty("cvc");
  });

  it("maps all payment method fields correctly", () => {
    const customer = toCustomer({
      id: "cus_1",
      payment_methods: [
        {
          id: "pm_1",
          number: "xxxxxxxxxxxx4242",
          brand: "visa",
          exp_month: 12,
          exp_year: 2030,
          cvc: "xxx",
          default: true,
        },
        { id: "pm_2" },
      ],
    });

    expect(customer?.payment_methods).toHaveLength(2);
    expect(customer?.payment_methods[0]).toMatchObject({
      id: "pm_1",
      number: "xxxxxxxxxxxx4242",
      brand: "visa",
      exp_month: 12,
      exp_year: 2030,
      default: true,
    });
    expect(customer?.payment_methods[1]).toMatchObject({
      id: "pm_2",
      number: "",
      brand: "",
      exp_month: 0,
      exp_year: 0,
      default: false,
    });
  });

  it("fills missing optional fields with empty defaults", () => {
    const customer = toCustomer({ id: "cus_1" });

    expect(customer?.name).toBe("");
    expect(customer?.email).toBe("");
    expect(customer?.payment_methods).toEqual([]);
  });
});

describe("toSubscription", () => {
  it("returns null when input is undefined", () => {
    expect(toSubscription(undefined)).toBeNull();
  });

  it("returns null when input is null", () => {
    expect(toSubscription(null)).toBeNull();
  });

  it("defaults status to 'inactive' when missing", () => {
    const s = toSubscription({});
    expect(s?.status).toBe("inactive");
    expect(s?.active).toBe(false);
    expect(s?.invoices).toEqual([]);
  });

  it("returns the correct shape for a full subscription", () => {
    const s = toSubscription({
      id: "sub_1",
      active: true,
      status: "active",
      end_at: 1735689600,
      invoices: [{ id: "in_1", status: "paid", currency: "usd", amount: 2999 }],
    });

    expect(s).toEqual({
      id: "sub_1",
      active: true,
      status: "active",
      end_at: 1735689600,
      invoices: [{ id: "in_1", status: "paid", currency: "usd", amount: 2999 }],
    });
  });

  it("normalizes invoices with missing fields", () => {
    const s = toSubscription({
      id: "sub_1",
      active: true,
      status: "active",
      end_at: 1735689600,
      invoices: [
        { id: "in_1", status: "open", currency: "usd", amount: 1000 },
        {},
      ],
    });

    expect(s?.invoices).toHaveLength(2);
    expect(s?.invoices[0]).toEqual({
      id: "in_1",
      status: "open",
      currency: "usd",
      amount: 1000,
    });
    expect(s?.invoices[1]).toEqual({
      id: "",
      status: "draft",
      currency: "usd",
      amount: 0,
    });
  });
});

describe("readNamespaceBilling", () => {
  it("returns null for null", () => {
    expect(readNamespaceBilling(null)).toBeNull();
  });

  it("returns null for undefined", () => {
    expect(readNamespaceBilling(undefined)).toBeNull();
  });

  it("returns an object with all undefined fields for an empty object", () => {
    const b = readNamespaceBilling({});

    expect(b).not.toBeNull();
    expect(b?.active).toBeUndefined();
    expect(b?.status).toBeUndefined();
    expect(b?.customer_id).toBeUndefined();
    expect(b?.subscription_id).toBeUndefined();
    expect(b?.current_period_end).toBeUndefined();
    expect(b?.created_at).toBeUndefined();
    expect(b?.updated_at).toBeUndefined();
  });

  it("narrows status and customer_id correctly", () => {
    const b = readNamespaceBilling({
      status: "active",
      customer_id: "cus_123",
    });

    expect(b?.status).toBe("active");
    expect(b?.customer_id).toBe("cus_123");
  });

  it("narrows a fully populated billing object", () => {
    const b = readNamespaceBilling({
      active: true,
      status: "active",
      customer_id: "cus_1",
      subscription_id: "sub_1",
      current_period_end: 1735689600,
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-06-01T00:00:00Z",
    });

    expect(b).toEqual({
      active: true,
      status: "active",
      customer_id: "cus_1",
      subscription_id: "sub_1",
      current_period_end: 1735689600,
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-06-01T00:00:00Z",
    });
  });

  it("ignores fields with wrong types", () => {
    const b = readNamespaceBilling({
      active: "yes",
      customer_id: 42,
      current_period_end: "not-a-number",
    });

    expect(b?.active).toBeUndefined();
    expect(b?.customer_id).toBeUndefined();
    expect(b?.current_period_end).toBeUndefined();
  });
});
