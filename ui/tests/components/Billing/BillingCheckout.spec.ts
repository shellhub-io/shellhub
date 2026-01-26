import { describe, expect, it, afterEach, beforeEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import useCustomerStore from "@/store/modules/customer";
import BillingCheckout from "@/components/Billing/BillingCheckout.vue";
import { mockCustomer } from "@tests/mocks/customer";

describe("BillingCheckout", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingCheckout>>;
  let customerStore: ReturnType<typeof useCustomerStore>;

  beforeEach(async () => {
    wrapper = mountComponent(BillingCheckout, { piniaOptions: {
      initialState: { customer: { customer: mockCustomer } },
    } });
    customerStore = useCustomerStore();
    await flushPromises();
  });

  afterEach(() => { wrapper?.unmount(); });

  describe("customer data fetching", () => {
    it("fetches customer on mount", () => {
      expect(customerStore.fetchCustomer).toHaveBeenCalled();
    });
  });

  describe("rendering", () => {
    it("displays title and subtitle", () => {
      expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="sub-title"]').exists()).toBe(true);
    });

    it("displays payment card", () => {
      expect(wrapper.find('[data-test="card"]').exists()).toBe(true);
    });

    it("displays billing icon", () => {
      expect(wrapper.findComponent({ name: "BillingIcon" }).exists()).toBe(true);
    });

    it("displays additional information section", () => {
      expect(wrapper.find('[data-test="additional-information"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="additional-information-list"]').exists()).toBe(true);
    });
  });
});
