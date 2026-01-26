import { describe, expect, it, afterEach } from "vitest";
import { VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import BillingIcon from "@/components/Billing/BillingIcon.vue";

describe("BillingIcon", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingIcon>>;

  const mountWrapper = (iconName = "") => {
    wrapper = mountComponent(BillingIcon, { props: { iconName } });
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("default icon", () => {
    it("renders default credit card icon when no brand specified", () => {
      mountWrapper();

      expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(false);
    });

    it("renders default icon for unknown card brand", () => {
      mountWrapper("unknown-brand");

      expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(false);
    });
  });

  describe("brand-specific icons", () => {
    it("renders visa icon", () => {
      mountWrapper("visa");

      expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(false);
    });

    it("renders mastercard icon", () => {
      mountWrapper("mastercard");

      expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(false);
    });

    it("renders amex icon", () => {
      mountWrapper("amex");

      expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(false);
    });

    it("renders diners-club icon", () => {
      mountWrapper("diners-club");

      expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(false);
    });

    it("renders discover icon", () => {
      mountWrapper("discover");

      expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(false);
    });

    it("renders jcb icon", () => {
      mountWrapper("jcb");

      expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="default-icon"]').exists()).toBe(false);
    });
  });
});
