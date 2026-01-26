import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useCustomerStore from "@/store/modules/customer";
import BillingDialog from "@/components/Billing/BillingDialog.vue";
import { Router } from "vue-router";

type CustomerStore = ReturnType<typeof useCustomerStore>;

describe("BillingDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingDialog>>;
  let customerStore: CustomerStore;
  let dialog: DOMWrapper<HTMLElement>;
  let router: Router;

  const mountWrapper = async (modelValue = true) => {
    router = createCleanRouter();

    wrapper = mountComponent(BillingDialog, {
      props: { modelValue },
      global: { plugins: [router] },
    });

    customerStore = useCustomerStore();

    dialog = new DOMWrapper(document.body).find('[role="dialog"]');

    await flushPromises();
  };

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("dialog visibility", () => {
    it("does not display dialog when modelValue is false", async () => {
      await mountWrapper(false);

      expect(dialog.exists()).toBe(false);
    });

    it("displays dialog when modelValue is true", async () => {
      await mountWrapper();

      expect(dialog.exists()).toBe(true);
    });
  });

  describe("window navigation - step 1: welcome letter", () => {
    beforeEach(() => mountWrapper());

    it("renders welcome screen initially", () => {
      expect(dialog.text()).toContain("ShellHub Cloud Premium Subscription");
      expect(dialog.find('[data-test="billing-payment-details"]').exists()).toBe(false);
    });

    it("displays next button on welcome screen", () => {
      expect(dialog.find('[data-test="payment-letter-next-button"]').exists()).toBe(true);
    });

    it("moves to payment details when next is clicked", async () => {
      await dialog.find('[data-test="payment-letter-next-button"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="billing-payment-details"]').exists()).toBe(true);
      expect(dialog.find('[data-test="billing-letter"]').exists()).toBe(false);
    });
  });

  describe("window navigation - step 2: payment details", () => {
    beforeEach(() => mountWrapper());

    it("displays back and next buttons", async () => {
      await dialog.find('[data-test="payment-letter-next-button"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="payment-details-back-button"]').exists()).toBe(true);
      expect(dialog.find('[data-test="payment-details-next-button"]').exists()).toBe(true);
    });

    it("returns to welcome when back is clicked", async () => {
      await dialog.find('[data-test="payment-letter-next-button"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="payment-details-back-button"]').trigger("click");
      await flushPromises();

      expect(dialog.text()).toContain("ShellHub Cloud Premium Subscription");
    });

    it("moves to checkout when next is clicked", async () => {
      await dialog.find('[data-test="payment-letter-next-button"]').trigger("click");
      await dialog.find('[data-test="payment-details-next-button"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="billing-checkout"]').exists()).toBe(true);
    });
  });

  describe("window navigation - step 3: checkout", () => {
    beforeEach(() => mountWrapper());

    it("displays back and subscribe buttons", async () => {
      await dialog.find('[data-test="payment-letter-next-button"]').trigger("click");
      await dialog.find('[data-test="payment-details-next-button"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="checkout-back-button"]').exists()).toBe(true);
      expect(dialog.find('[data-test="checkout-button"]').exists()).toBe(true);
    });

    it("returns to payment details when back is clicked", async () => {
      await dialog.find('[data-test="payment-letter-next-button"]').trigger("click");
      await dialog.find('[data-test="payment-details-next-button"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="checkout-back-button"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="billing-payment-details"]').exists()).toBe(true);
    });
  });

  describe("subscription creation", () => {
    beforeEach(() => mountWrapper());
    it("creates subscription when subscribe button is clicked", async () => {
      await dialog.find('[data-test="payment-letter-next-button"]').trigger("click");
      await dialog.find('[data-test="payment-details-next-button"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="checkout-button"]').trigger("click");
      await flushPromises();

      expect(customerStore.createSubscription).toHaveBeenCalled();
    });

    it("shows success screen after successful subscription", async () => {
      await dialog.find('[data-test="payment-letter-next-button"]').trigger("click");
      await dialog.find('[data-test="payment-details-next-button"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="checkout-button"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="card-fourth-page"]').exists()).toBe(true);
    });

    it("handles payment required error (402)", async () => {
      await dialog.find('[data-test="payment-letter-next-button"]').trigger("click");
      await dialog.find('[data-test="payment-details-next-button"]').trigger("click");
      await flushPromises();

      const error = createAxiosError(402, "Payment Required");
      vi.spyOn(customerStore, "createSubscription").mockRejectedValueOnce(error);

      await dialog.find('[data-test="checkout-button"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="checkout-error-alert"]').exists()).toBe(true);
    });
  });

  describe("window navigation - step 4: success", () => {
    beforeEach(() => mountWrapper());
    it("displays close button on success screen", async () => {
      await dialog.find('[data-test="payment-letter-next-button"]').trigger("click");
      await dialog.find('[data-test="payment-details-next-button"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="checkout-button"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="successful-close-button"]').exists()).toBe(true);
    });

    it("closes dialog when close button is clicked", async () => {
      await dialog.find('[data-test="payment-letter-next-button"]').trigger("click");
      await dialog.find('[data-test="payment-details-next-button"]').trigger("click");
      await dialog.find('[data-test="checkout-button"]').trigger("click");
      await flushPromises();

      await dialog.find('[data-test="successful-close-button"]').trigger("click");
      await flushPromises();

      expect(wrapper.emitted("reload")).toBeTruthy();
    });
  });
});
