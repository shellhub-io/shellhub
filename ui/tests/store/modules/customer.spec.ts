import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { billingApi } from "@/api/http";
import useCustomerStore from "@/store/modules/customer";
import { ICustomer } from "@/interfaces/ICustomer";
import * as handleErrorModule from "@/utils/handleError";

const mockCustomerBase: ICustomer = {
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
      exp_year: 2025,
      cvc: "***",
      default: false,
    },
  ],
};

describe("Customer Store", () => {
  let mockBillingApi: MockAdapter;
  let store: ReturnType<typeof useCustomerStore>;
  let handleErrorSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    setActivePinia(createPinia());
    mockBillingApi = new MockAdapter(billingApi.getAxios());
    store = useCustomerStore();
    handleErrorSpy = vi.spyOn(handleErrorModule, "default").mockImplementation((error) => {
      throw error;
    });
  });

  afterEach(() => {
    mockBillingApi.reset();
    handleErrorSpy.mockRestore();
  });

  describe("Initial State", () => {
    it("should have empty customer object", () => {
      expect(store.customer).toEqual({});
    });
  });

  describe("fetchCustomer", () => {
    it("should fetch customer data successfully", async () => {
      mockBillingApi
        .onGet("http://localhost:3000/api/billing/customer")
        .reply(200, mockCustomerBase);

      await store.fetchCustomer();

      expect(store.customer).toEqual(mockCustomerBase);
    });

    it("should fetch customer with multiple payment methods", async () => {
      const customerWithMultiplePayments = {
        ...mockCustomerBase,
        payment_methods: [
          ...mockCustomerBase.payment_methods,
          {
            id: "pm_test789",
            number: "**** **** **** 9012",
            brand: "Amex",
            exp_month: 3,
            exp_year: 2026,
            cvc: "***",
            default: false,
          },
        ],
      };

      mockBillingApi
        .onGet("http://localhost:3000/api/billing/customer")
        .reply(200, customerWithMultiplePayments);

      await store.fetchCustomer();

      expect(store.customer.payment_methods).toHaveLength(3);
    });

    it("should fetch customer with no payment methods", async () => {
      const customerWithoutPayments = {
        ...mockCustomerBase,
        payment_methods: [],
      };

      mockBillingApi
        .onGet("http://localhost:3000/api/billing/customer")
        .reply(200, customerWithoutPayments);

      await store.fetchCustomer();

      expect(store.customer.payment_methods).toEqual([]);
    });

    it("should handle not found error when fetching customer", async () => {
      mockBillingApi
        .onGet("http://localhost:3000/api/billing/customer")
        .reply(404, { message: "Customer not found" });

      await expect(
        store.fetchCustomer(),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when fetching customer", async () => {
      mockBillingApi
        .onGet("http://localhost:3000/api/billing/customer")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.fetchCustomer(),
      ).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when fetching customer", async () => {
      mockBillingApi
        .onGet("http://localhost:3000/api/billing/customer")
        .networkError();

      await expect(store.fetchCustomer()).rejects.toThrow();
    });
  });

  describe("createCustomer", () => {
    it("should create customer successfully", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/customer")
        .reply(201);

      await expect(store.createCustomer()).resolves.not.toThrow();
    });

    it("should handle validation error when creating customer", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/customer")
        .reply(400, { message: "Invalid customer data" });

      await expect(
        store.createCustomer(),
      ).rejects.toBeAxiosErrorWithStatus(400);
    });

    it("should handle server error when creating customer", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/customer")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.createCustomer(),
      ).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when creating customer", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/customer")
        .networkError();

      await expect(store.createCustomer()).rejects.toThrow();
    });
  });

  describe("attachPaymentMethod", () => {
    it("should attach payment method successfully", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/attach")
        .reply(200);

      await expect(store.attachPaymentMethod("pm_new123")).resolves.not.toThrow();
    });

    it("should throw error data when request fails with validation error", async () => {
      const errorData = { message: "Invalid payment method" };

      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/attach")
        .reply(400, errorData);

      await expect(
        store.attachPaymentMethod("pm_invalid"),
      ).rejects.toEqual(errorData);
    });

    it("should throw error data when payment method already exists", async () => {
      const errorData = { message: "Payment method already attached" };

      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/attach")
        .reply(409, errorData);

      await expect(
        store.attachPaymentMethod("pm_duplicate"),
      ).rejects.toEqual(errorData);
    });

    it("should handle network error when attaching payment method", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/attach")
        .networkError();

      await expect(store.attachPaymentMethod("pm_test")).rejects.toThrow();
    });
  });

  describe("detachPaymentMethod", () => {
    it("should detach payment method successfully", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/detach")
        .reply(200);

      await expect(store.detachPaymentMethod("pm_test123")).resolves.not.toThrow();
    });

    it("should handle not found error when detaching payment method", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/detach")
        .reply(404, { message: "Payment method not found" });

      await expect(
        store.detachPaymentMethod("pm_nonexistent"),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when detaching payment method", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/detach")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.detachPaymentMethod("pm_test123"),
      ).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when detaching payment method", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/detach")
        .networkError();

      await expect(store.detachPaymentMethod("pm_test123")).rejects.toThrow();
    });
  });

  describe("createSubscription", () => {
    it("should create subscription successfully", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/subscription")
        .reply(201);

      await expect(store.createSubscription()).resolves.not.toThrow();
    });

    it("should throw error when no payment method available", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/subscription")
        .reply(400, { message: "No payment method available" });

      await expect(
        store.createSubscription(),
      ).rejects.toThrow();
    });

    it("should handle conflict error when subscription already exists", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/subscription")
        .reply(409, { message: "Subscription already exists" });

      await expect(
        store.createSubscription(),
      ).rejects.toBeAxiosErrorWithStatus(409);
    });

    it("should handle network error when creating subscription", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/subscription")
        .networkError();

      await expect(store.createSubscription()).rejects.toThrow();
    });
  });

  describe("setDefaultPaymentMethod", () => {
    it("should set default payment method successfully", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/default")
        .reply(200);

      await expect(store.setDefaultPaymentMethod("pm_test456")).resolves.not.toThrow();
    });

    it("should handle not found error when setting default payment method", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/default")
        .reply(404, { message: "Payment method not found" });

      await expect(
        store.setDefaultPaymentMethod("pm_nonexistent"),
      ).rejects.toBeAxiosErrorWithStatus(404);
    });

    it("should handle server error when setting default payment method", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/default")
        .reply(500, { message: "Internal Server Error" });

      await expect(
        store.setDefaultPaymentMethod("pm_test123"),
      ).rejects.toBeAxiosErrorWithStatus(500);
    });

    it("should handle network error when setting default payment method", async () => {
      mockBillingApi
        .onPost("http://localhost:3000/api/billing/paymentmethod/default")
        .networkError();

      await expect(store.setDefaultPaymentMethod("pm_test123")).rejects.toThrow();
    });
  });
});
