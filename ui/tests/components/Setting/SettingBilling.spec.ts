import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import SettingBilling from "@/components/Setting/SettingBilling.vue";
import useBillingStore from "@/store/modules/billing";
import useNamespacesStore from "@/store/modules/namespaces";
import handleError from "@/utils/handleError";
import {
  mockBilling, mockBillingInactive, mockBillingToCancelAtEndOfPeriod, mockBillingPastDue,
  mockBillingUnpaid, mockBillingCanceled, mockNamespace,
} from "@tests/mocks";
import { INamespace } from "@/interfaces/INamespace";
import * as hasPermissionModule from "@/utils/permission";

describe("SettingBilling", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingBilling>>;
  let billingStore: ReturnType<typeof useBillingStore>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;

  const namespaceWithBilling = {
    ...mockNamespace,
    billing: { customer_id: "cus_123" },
  };

  const mountWrapper = ({
    hasPermission = true,
    billing = mockBilling,
    namespace = namespaceWithBilling as INamespace,
  } = {}) => {
    vi.spyOn(hasPermissionModule, "default").mockReturnValue(hasPermission);

    localStorage.setItem("tenant", namespace.tenant_id);

    wrapper = mountComponent(SettingBilling, {
      piniaOptions: {
        initialState: {
          billing: {
            billing,
            invoices: billing.invoices,
            isActive: billing.active,
            status: billing.status,
          },
          namespaces: {
            namespaces: [namespace],
            currentNamespace: namespace,
          },
        },
      },
    });

    billingStore = useBillingStore();
    namespacesStore = useNamespacesStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Permission check", () => {
    it("Shows SettingOwnerInfo when user cannot subscribe", () => {
      wrapper.unmount();
      mountWrapper({ hasPermission: false });

      const ownerInfo = wrapper.find('[data-test="settings-owner-info-component"]');
      expect(ownerInfo.exists()).toBe(true);
    });

    it("Shows billing content when user can subscribe", () => {
      const billingCard = wrapper.find('[data-test="billing-card"]');
      expect(billingCard.exists()).toBe(true);
    });
  });

  describe("Page header", () => {
    it("Renders page header with correct props", () => {
      const header = wrapper.findComponent({ name: "PageHeader" });
      expect(header.exists()).toBe(true);
      expect(header.props("icon")).toBe("mdi-credit-card");
      expect(header.props("title")).toBe("Billing");
      expect(header.props("overline")).toBe("Settings");
    });

    it("Renders subscribe button", () => {
      const subscribeBtn = wrapper.find('[data-test="subscribe-button"]');
      expect(subscribeBtn.exists()).toBe(true);
      expect(subscribeBtn.text()).toContain("Subscribe");
    });

    it("Opens billing dialog when subscribe button is clicked", async () => {
      const subscribeBtn = wrapper.find('[data-test="subscribe-button"]');
      await subscribeBtn.trigger("click");

      const dialog = wrapper.findComponent({ name: "BillingDialog" });
      expect(dialog.props("modelValue")).toBe(true);
    });
  });

  describe("Billing portal section", () => {
    it("Renders billing portal section", () => {
      const portalSection = wrapper.find('[data-test="billing-portal-section"]');
      expect(portalSection.exists()).toBe(true);
    });

    it("Renders billing portal icon", () => {
      const icon = wrapper.find('[data-test="billing-portal-icon"]');
      expect(icon.exists()).toBe(true);
      expect(icon.classes().join(" ")).toContain("mdi-account");
    });

    it("Renders billing portal title", () => {
      const title = wrapper.find('[data-test="billing-portal-title"]');
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Billing Portal");
    });

    it("Renders billing portal description", () => {
      const description = wrapper.find('[data-test="billing-portal-description"]');
      expect(description.exists()).toBe(true);
      expect(description.text()).toContain("Update your ShellHub payment method");
    });

    it("Renders billing portal button", () => {
      const button = wrapper.find('[data-test="billing-portal-button"]');
      expect(button.exists()).toBe(true);
      expect(button.text()).toContain("Open Billing Portal");
    });

    it("Disables portal button when no customer", async () => {
      wrapper.unmount();
      mountWrapper({ namespace: mockNamespace });
      await flushPromises();

      const button = wrapper.find('[data-test="billing-portal-button"]');
      expect(button.attributes("disabled")).toBeDefined();
    });

    it("Opens billing portal when button is clicked", async () => {
      const button = wrapper.find('[data-test="billing-portal-button"]');
      await button.trigger("click");
      await flushPromises();

      expect(billingStore.openBillingPortal).toHaveBeenCalled();
    });

    it("Shows error when opening billing portal fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(billingStore.openBillingPortal).mockRejectedValueOnce(error);

      const button = wrapper.find('[data-test="billing-portal-button"]');
      await button.trigger("click");
      await flushPromises();

      expect(handleError).toHaveBeenCalledWith(error);
    });
  });

  describe("Billing plan section", () => {
    it("Renders billing plan section", () => {
      const planSection = wrapper.find('[data-test="billing-plan-section"]');
      expect(planSection.exists()).toBe(true);
    });

    it("Renders plan icon", () => {
      const icon = wrapper.find('[data-test="billing-plan-icon"]');
      expect(icon.exists()).toBe(true);
      expect(icon.classes().join(" ")).toContain("mdi-credit-card");
    });

    it("Renders plan title", () => {
      const title = wrapper.find('[data-test="billing-plan-title"]');
      expect(title.exists()).toBe(true);
      expect(title.text()).toBe("Plan");
    });

    it("Shows free plan when billing is not active", async () => {
      wrapper.unmount();
      mountWrapper({ billing: mockBillingInactive });
      await flushPromises();

      const description = wrapper.find('[data-test="billing-plan-description-free"]');
      expect(description.exists()).toBe(true);
      expect(description.text()).toContain("You can add up to 3 devices");

      const plan = wrapper.find('[data-test="billing-plan-free"]');
      expect(plan.exists()).toBe(true);
      expect(plan.text()).toBe("Free");
    });

    it("Shows premium plan when billing is active", () => {
      const description = wrapper.find('[data-test="billing-plan-description-premium"]');
      expect(description.exists()).toBe(true);
      expect(description.text()).toContain("the amount is charged according to the number of devices");

      const plan = wrapper.find('[data-test="billing-plan-premium"]');
      expect(plan.exists()).toBe(true);
      expect(plan.text()).toBe("Premium usage");
    });
  });

  describe("Billing active section", () => {
    it("Shows billing active section when billing is active", () => {
      const activeSection = wrapper.find('[data-test="billing-active-section"]');
      expect(activeSection.exists()).toBe(true);
    });

    it("Does not show billing active section when billing is not active", async () => {
      wrapper.unmount();
      mountWrapper({ billing: mockBillingInactive });
      await flushPromises();

      const activeSection = wrapper.find('[data-test="billing-active-section"]');
      expect(activeSection.exists()).toBe(false);
    });

    it("Shows billing total section", () => {
      const totalSection = wrapper.find('[data-test="billing-total-section"]');
      expect(totalSection.exists()).toBe(true);

      const totalIcon = wrapper.find('[data-test="billing-total-icon"]');
      expect(totalIcon.exists()).toBe(true);

      const totalTitle = wrapper.find('[data-test="billing-total-title"]');
      expect(totalTitle.exists()).toBe(true);
      expect(totalTitle.text()).toBe("Billing estimated total");

      const totalAmount = wrapper.find('[data-test="billing-total-amount"]');
      expect(totalAmount.exists()).toBe(true);
    });

    it("Shows billing end date section", () => {
      const endDateSection = wrapper.find('[data-test="billing-end-date-section"]');
      expect(endDateSection.exists()).toBe(true);

      const endDateIcon = wrapper.find('[data-test="billing-end-date-icon"]');
      expect(endDateIcon.exists()).toBe(true);

      const endDateTitle = wrapper.find('[data-test="billing-end-date-title"]');
      expect(endDateTitle.exists()).toBe(true);
      expect(endDateTitle.text()).toBe("Current billing ends at");

      const endDate = wrapper.find('[data-test="billing-end-date"]');
      expect(endDate.exists()).toBe(true);
    });
  });

  describe("Billing status messages", () => {
    it("Shows warning message for to_cancel_at_end_of_period status", async () => {
      wrapper.unmount();
      mountWrapper({ billing: mockBillingToCancelAtEndOfPeriod });
      await flushPromises();

      const statusSection = wrapper.find('[data-test="billing-status-section"]');
      expect(statusSection.exists()).toBe(true);

      const message = wrapper.find('[data-test="billing-status-message"]');
      expect(message.exists()).toBe(true);
      expect(message.classes()).toContain("text-warning");
      expect(message.text()).toContain("will be canceled");
    });

    it("Shows warning message for past_due status", async () => {
      wrapper.unmount();
      mountWrapper({ billing: mockBillingPastDue });
      await flushPromises();

      const message = wrapper.find('[data-test="billing-status-message"]');
      expect(message.classes()).toContain("text-warning");
      expect(message.text()).toContain("payment method has failed");
    });

    it("Shows error message for unpaid status", async () => {
      wrapper.unmount();
      mountWrapper({ billing: mockBillingUnpaid });
      await flushPromises();

      const message = wrapper.find('[data-test="billing-status-message"]');
      expect(message.classes()).toContain("text-error");
      expect(message.text()).toContain("unpaid invoices");
    });

    it("Shows error message for canceled status", async () => {
      wrapper.unmount();
      mountWrapper({ billing: mockBillingCanceled });
      await flushPromises();

      const message = wrapper.find('[data-test="billing-status-message"]');
      expect(message.classes()).toContain("text-error");
      expect(message.text()).toContain("subscription was canceled");
    });
  });

  describe("Subscription info fetching", () => {
    it("Fetches namespace on mount", async () => {
      await flushPromises();
      expect(namespacesStore.fetchNamespace).toHaveBeenCalledWith("fake-tenant-data");
    });

    it("Fetches subscription info on mount", async () => {
      await flushPromises();
      expect(billingStore.getSubscriptionInfo).toHaveBeenCalled();
    });

    it("Shows error when fetching subscription info fails", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      wrapper.unmount();
      mountWrapper();
      vi.mocked(billingStore.getSubscriptionInfo).mockRejectedValueOnce(error);
      await flushPromises();

      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
