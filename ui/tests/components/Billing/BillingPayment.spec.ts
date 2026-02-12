import { describe, expect, it, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import useCustomerStore from "@/store/modules/customer";
import BillingPayment from "@/components/Billing/BillingPayment.vue";
import { mockNamespace, mockNamespaceWithBilling } from "@tests/mocks";
import { mockCustomer, mockCustomerNoPaymentMethods } from "@tests/mocks/customer";
import handleError from "@/utils/handleError";
import { envVariables } from "@/envVariables";

vi.mock("@stripe/stripe-js", () => {
  return {
    loadStripe: vi.fn().mockResolvedValue({
      elements: vi.fn(() => ({
        create: vi.fn(() => ({
          mount: vi.fn(),
          unmount: vi.fn(),
          clear: vi.fn(),
        })),
      })),
      createPaymentMethod: vi.fn(),
    }),
  };
});

describe("BillingPayment", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingPayment>>;
  let customerStore: ReturnType<typeof useCustomerStore>;

  envVariables.stripeKey = "pk_test_12345";

  const mountWrapper = async ({ hasBilling = true, customer = mockCustomer } = {}) => {
    wrapper = mountComponent(BillingPayment, {
      global: { stubs: ["StripeElements"] },
      piniaOptions: {
        initialState: {
          customer: { customer },
          namespaces: { currentNamespace: hasBilling ? mockNamespaceWithBilling : mockNamespace },
        },
      },
    });
    customerStore = useCustomerStore();

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("customer initialization", () => {
    it("fetches existing customer on mount", async () => {
      await mountWrapper();

      expect(customerStore.fetchCustomer).toHaveBeenCalled();
    });

    it("creates new customer if none exists", async () => {
      await mountWrapper({ hasBilling: false });

      expect(customerStore.createCustomer).toHaveBeenCalled();
    });

    it("emits customer-id-created after creating customer", async () => {
      await mountWrapper({ hasBilling: false });
      vi.mocked(customerStore.fetchCustomer).mockRejectedValueOnce(createAxiosError(404, "Not Found"));

      expect(wrapper.emitted("customer-id-created")).toBeTruthy();
    });
  });

  describe("customer information display", () => {
    it("displays customer name", async () => {
      await mountWrapper();

      const nameField = wrapper.find('[data-test="customer-name"]');
      expect(nameField.exists()).toBe(true);
    });

    it("displays customer email", async () => {
      await mountWrapper();

      const emailField = wrapper.find('[data-test="customer-email"]');
      expect(emailField.exists()).toBe(true);
    });

    it("displays credit card text", async () => {
      await mountWrapper();

      expect(wrapper.find('[data-test="credit-card-text"]').exists()).toBe(true);
    });
  });

  describe("payment methods list", () => {
    it("renders payment methods list", async () => {
      await mountWrapper();

      expect(wrapper.find('[data-test="payment-methods-list"]').exists()).toBe(true);
    });

    it("displays payment method items", async () => {
      await mountWrapper();

      const items = wrapper.findAll('[data-test="payment-methods-item"]');
      expect(items.length).toBeGreaterThan(0);
    });

    it("shows delete button for each payment method", async () => {
      await mountWrapper();

      expect(wrapper.find('[data-test="payment-methods-delete-btn"]').exists()).toBe(true);
    });

    it("emits has-default-payment when default method exists", async () => {
      await mountWrapper();

      expect(wrapper.emitted("has-default-payment")).toBeTruthy();
    });

    it("emits no-payment-methods when list is empty", async () => {
      await mountWrapper({ customer: mockCustomerNoPaymentMethods });

      expect(wrapper.emitted("no-payment-methods")).toBeTruthy();
    });
  });

  describe("payment method deletion", () => {
    it("detaches payment method when delete button is clicked", async () => {
      await mountWrapper();
      const deleteBtn = wrapper.find('[data-test="payment-methods-delete-btn"]');
      await deleteBtn.trigger("click");

      expect(customerStore.detachPaymentMethod).toHaveBeenCalledWith("pm_mastercard");
    });

    it("handles detach error", async () => {
      await mountWrapper();
      const error = createAxiosError(400, "Bad Request");
      vi.mocked(customerStore.detachPaymentMethod).mockRejectedValueOnce(error);
      await flushPromises();

      const deleteBtn = wrapper.find('[data-test="payment-methods-delete-btn"]');
      await deleteBtn.trigger("click");
      await flushPromises();

      // Error should be handled gracefully
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("default payment method", () => {
    it("sets default payment method when item is clicked", async () => {
      await mountWrapper();

      const item = wrapper.find('[data-test="payment-methods-item"]');
      await item.trigger("click");

      expect(customerStore.setDefaultPaymentMethod).toHaveBeenCalledWith("pm_visa");
    });

    it("handles set default error", async () => {
      await mountWrapper();

      const error = createAxiosError(400, "Bad Request");
      vi.mocked(customerStore.setDefaultPaymentMethod).mockRejectedValueOnce(error);

      const item = wrapper.find('[data-test="payment-methods-item"]');
      await item.trigger("click");
      await flushPromises();

      // Error should be handled gracefully
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("add new card", () => {
    it("displays add card button", async () => {
      await mountWrapper();

      expect(wrapper.find('[data-test="add-card-btn"]').exists()).toBe(true);
    });

    it("shows Stripe card element when add card is clicked", async () => {
      await mountWrapper();

      await wrapper.find('[data-test="add-card-btn"]').trigger("click");
      await flushPromises();

      expect(wrapper.findComponent({ name: "StripeElements" }).exists()).toBe(true);
    });
  });
});
