import { describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useCustomerStore from "@/store/modules/customer";
import { ICustomer } from "@/interfaces/ICustomer";

describe("Customer Pinia Store", () => {
  setActivePinia(createPinia());
  const customerStore = useCustomerStore();

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

  describe("initial state", () => {
    it("should have initial state values", () => {
      expect(customerStore.customer).toEqual({});

      const hasPaymentMethods = !!(customerStore.customer.payment_methods && customerStore.customer.payment_methods.length > 0);
      expect(hasPaymentMethods).toBe(false);
    });
  });

  describe("customer data management", () => {
    it("should update customer data directly", () => {
      customerStore.customer = customerData;

      expect(customerStore.customer).toEqual(customerData);

      const hasPaymentMethods = !!(customerStore.customer.payment_methods && customerStore.customer.payment_methods.length > 0);
      expect(hasPaymentMethods).toBe(true);
    });

    it("should detect payment methods correctly", () => {
      customerStore.customer = {} as ICustomer;

      const hasPaymentMethods = () => !!(customerStore.customer.payment_methods && customerStore.customer.payment_methods.length > 0);
      expect(hasPaymentMethods()).toBe(false);

      customerStore.customer = customerData;
      expect(hasPaymentMethods()).toBe(true);

      const customerWithoutPayments = { ...customerData, payment_methods: [] };
      customerStore.customer = customerWithoutPayments;
      expect(hasPaymentMethods()).toBe(false);
    });
  });
});
