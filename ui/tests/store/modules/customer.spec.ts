import { describe, expect, it } from "vitest";
import { store } from "@/store";

describe("Customer Store", () => {
  const customerData = {
    id: "cus_test123",
    name: "testuser",
    email: "test@test.com",
    payment_methods: [
      {
        id: "pm_test123",
        number: "**** **** **** 1234",
        brand: "Visa",
        exp_month: 12,
        exp_year: 2024,
        cvc: "***",
        default: true,
      },
      {
        id: "pm_test456",
        number: "**** **** **** 5678",
        brand: "Mastercard",
        exp_month: 9,
        exp_year: 2023,
        cvc: "***",
        default: false,
      },
    ],
  };

  it("Returns customer with default variables", () => {
    expect(store.getters["customer/getCustomer"]).toEqual({});
    expect(store.getters["customer/hasPaymentMethods"]).toEqual(false);
  });

  it("Verify initial state change for setCustomer mutation", () => {
    store.commit("customer/setCustomer", customerData);
    expect(store.getters["customer/getCustomer"]).toEqual(customerData);
  });
});
