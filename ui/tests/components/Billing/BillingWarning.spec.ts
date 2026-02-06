import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, DOMWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import BillingWarning from "@/components/Billing/BillingWarning.vue";
import { Router } from "vue-router";
import hasPermission from "@/utils/permission";

vi.mock("@/utils/permission");

describe("BillingWarning", () => {
  let wrapper: VueWrapper<InstanceType<typeof BillingWarning>>;
  let dialog: DOMWrapper<HTMLElement>;
  let router: Router;

  const mountWrapper = (canSubscribeToBilling = true) => {
    router = createCleanRouter();
    vi.mocked(hasPermission).mockReturnValue(canSubscribeToBilling);

    wrapper = mountComponent(BillingWarning, {
      global: { plugins: [router] },
      props: { modelValue: true },
    });

    dialog = new DOMWrapper(document.body).find('[role="dialog"]');
  };

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("when user cannot subscribe", () => {
    it("does not render dialog for users without subscription permission", () => {
      mountWrapper(false);
      expect(dialog.exists()).toBe(false);
    });
  });

  describe("when user can subscribe", () => {
    beforeEach(() => mountWrapper());

    it("renders dialog with correct title", () => {
      expect(dialog.exists()).toBe(true);
      expect(dialog.text()).toContain("Maximum Device Limit Reached");
    });

    it("displays upgrade message", () => {
      const description = "It seems that your current free account has reached the maximum number of devices allowed in this namespace";
      expect(dialog.text()).toContain(description);
    });

    it("closes dialog when close button is clicked", async () => {
      await dialog.find('[data-test="close-btn"]').trigger("click");
      const dialogContent = dialog.find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none");
    });

    it("navigates to billing page when go to billing button is clicked", async () => {
      const navigateSpy = vi.spyOn(router, "push");

      await dialog.find('[data-test="go-to-billing-btn"]').trigger("click");

      expect(navigateSpy).toHaveBeenCalledWith("/settings/billing");
    });
  });
});
