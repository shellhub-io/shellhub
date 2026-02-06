import { IBilling, IInvoice } from "@/interfaces/IBilling";

/**
 * Mock invoice data for testing.
 * Provides a basic invoice object with all required fields.
 */
export const mockInvoice: IInvoice = {
  id: "inv_123",
  status: "paid",
  currency: "usd",
  amount: 2000,
};

/**
 * Mock invoice data with different statuses for testing.
 */
export const mockInvoiceOpen: IInvoice = {
  ...mockInvoice,
  id: "inv_open",
  status: "open",
};

export const mockInvoiceDraft: IInvoice = {
  ...mockInvoice,
  id: "inv_draft",
  status: "draft",
};

/**
 * Mock billing data for testing.
 * Provides a complete billing object with all required fields.
 */
export const mockBilling: IBilling = {
  id: "billing_123",
  active: true,
  status: "active",
  end_at: 1735689600,
  invoices: [mockInvoice],
};

/**
 * Mock billing data for inactive subscription.
 */
export const mockBillingInactive: IBilling = {
  ...mockBilling,
  active: false,
  status: "",
  invoices: [],
};

/**
 * Mock billing data with various statuses for testing.
 */
export const mockBillingToCancelAtEndOfPeriod: IBilling = {
  ...mockBilling,
  status: "to_cancel_at_end_of_period",
};

export const mockBillingPastDue: IBilling = {
  ...mockBilling,
  status: "past_due",
};

export const mockBillingUnpaid: IBilling = {
  ...mockBilling,
  status: "unpaid",
};

export const mockBillingCanceled: IBilling = {
  ...mockBilling,
  status: "canceled",
};
