import { ICustomer } from "@/interfaces/ICustomer";

/**
 * Mock customer data for testing.
 * Provides a basic customer object with all required fields.
 */
export const mockCustomer: ICustomer = {
  id: "cust_123",
  name: "Test Customer",
  email: "test@example.com",
  payment_methods: [
    {
      id: "pm_visa",
      number: "4242424242424242",
      brand: "visa",
      exp_month: 12,
      exp_year: 2029,
      cvc: "424",
      default: true,
    },
    {
      id: "pm_mastercard",
      number: "5555555555555557",
      brand: "mastercard",
      exp_month: 12,
      exp_year: 2029,
      cvc: "555",
      default: false,
    },
  ],
};

/**
 * Mock customer data without payment methods for testing.
 * Provides a customer object with an empty payment methods array.
 */
export const mockCustomerNoPaymentMethods: ICustomer = {
  ...mockCustomer,
  payment_methods: [],
};
