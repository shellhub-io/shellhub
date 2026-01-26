import { describe, expect, it } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import BillingLetter from "@/components/Billing/BillingLetter.vue";

describe("BillingLetter", () => {
  const wrapper = mountComponent(BillingLetter);

  describe("rendering", () => {
    it("displays subscription title", () => {
      expect(wrapper.find('[data-test="title"]').exists()).toBe(true);
      expect(wrapper.text()).toContain("ShellHub Cloud Premium Subscription");
    });

    it("displays subscription description", () => {
      const text = wrapper.find('[data-test="text"]');
      expect(text.exists()).toBe(true);
      expect(text.text()).toContain("hosted service helps you manage your devices");
      expect(text.text()).toContain("costs based on your device count");
      expect(text.text()).toContain("pricing adapts, guaranteeing value");
      expect(text.text()).toContain("Click on next to proceed to the payment checkout");
    });
  });
});
