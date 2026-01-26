import { describe, expect, it } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import BillingSuccessful from "@/components/Billing/BillingSuccessful.vue";

describe("BillingSuccessful", () => {
  const wrapper = mountComponent(BillingSuccessful);

  describe("rendering", () => {
    it("displays success title", () => {
      expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("Congratulations! Subscription Successful!");
    });

    it("displays success icon", () => {
      expect(wrapper.find('[data-test="green-cloud-icon"]').exists()).toBe(true);
    });

    it("displays thank you message", () => {
      expect(wrapper.find('[data-test="h4-letter"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("Thank you for choosing ShellHub Cloud");
      expect(wrapper.text()).toContain("subscription has been successfully activated");
    });

    it("displays access description", () => {
      expect(wrapper.find('[data-test="p-letter"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("powerful device management solution");
      expect(wrapper.text()).toContain("connect, monitor, and control your devices");
    });
  });
});
