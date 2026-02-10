import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper, DOMWrapper } from "@vue/test-utils";
import PaywallDialog from "@/components/User/PaywallDialog.vue";
import { mountComponent } from "@tests/utils/mount";
import useUsersStore from "@/store/modules/users";
import { IPremiumFeature } from "@/interfaces/IUser";
import * as usersApi from "@/store/api/users";

vi.mock("@/store/api/users");

describe("PaywallDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof PaywallDialog>>;
  let dialog: DOMWrapper<Element>;
  let usersStore: ReturnType<typeof useUsersStore>;

  const mockPremiumFeatures: IPremiumFeature[] = [
    {
      title: "ShellHub Cloud",
      features: [
        "Protection Against DDoS Attacks",
        "Session record and playback",
        "Managing Firewall Rules",
        "Secure remote communication",
      ],
      button: {
        link: "https://www.shellhub.io/pricing",
        label: "Pricing",
      },
    },
    {
      title: "ShellHub Enterprise",
      features: [
        "Dedicated server for each customer",
        "Supports up to thousands of devices",
        "Reduced maintenance cost",
      ],
      button: {
        link: "https://www.shellhub.io/pricing",
        label: "Get a quote",
      },
    },
  ];

  const mountWrapper = async (premiumFeatures: IPremiumFeature[] = mockPremiumFeatures, modelValue = false) => {
    vi.mocked(usersApi.getPremiumContent).mockResolvedValue(premiumFeatures);
    wrapper = mountComponent(PaywallDialog, {
      props: { modelValue },
      attachTo: document.body,
      piniaOptions: { stubActions: false },
    });

    usersStore = useUsersStore();

    await flushPromises();

    dialog = new DOMWrapper(document.body);
  };

  beforeEach(async () => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component initialization", () => {
    it("fetches premium content on mount", () => {
      expect(usersStore.getPremiumContent).toHaveBeenCalled();
    });

    it("does not show dialog initially", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Component rendering with premium features", () => {
    beforeEach(async () => {
      await wrapper.setProps({ modelValue: true });
      await flushPromises();
    });

    it("renders MessageDialog with correct props", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.exists()).toBe(true);
      expect(messageDialog.props("title")).toBe("Upgrade to have access to all features!");
      expect(messageDialog.props("description")).toContain("ShellHub Community Edition");
      expect(messageDialog.props("icon")).toBe("mdi-crown-circle");
      expect(messageDialog.props("iconColor")).toBe("success");
      expect(messageDialog.props("cancelText")).toBe("Close");
    });

    it("renders items row container", () => {
      expect(dialog.find('[data-test="items-row"]').exists()).toBe(true);
    });

    it("does not render fallback button when features are available", () => {
      expect(dialog.find('[data-test="no-link-available-btn"]').exists()).toBe(false);
    });

    it("renders all premium feature cards", () => {
      expect(dialog.find('[data-test="item-0"]').exists()).toBe(true);
      expect(dialog.find('[data-test="item-1"]').exists()).toBe(true);
      expect(dialog.find('[data-test="item-card-0"]').exists()).toBe(true);
      expect(dialog.find('[data-test="item-card-1"]').exists()).toBe(true);
    });

    it("renders card titles correctly", () => {
      const title0 = dialog.find('[data-test="item-title-0"]');
      const title1 = dialog.find('[data-test="item-title-1"]');

      expect(title0.exists()).toBe(true);
      expect(title1.exists()).toBe(true);
      expect(title0.text()).toContain("ShellHub Cloud");
      expect(title1.text()).toContain("ShellHub Enterprise");
    });

    it("renders all features for each card", () => {
      // Cloud features (4 items)
      expect(dialog.find('[data-test="item-content-row-0-0"]').exists()).toBe(true);
      expect(dialog.find('[data-test="item-content-row-0-1"]').exists()).toBe(true);
      expect(dialog.find('[data-test="item-content-row-0-2"]').exists()).toBe(true);
      expect(dialog.find('[data-test="item-content-row-0-3"]').exists()).toBe(true);

      // Enterprise features (3 items)
      expect(dialog.find('[data-test="item-content-row-1-0"]').exists()).toBe(true);
      expect(dialog.find('[data-test="item-content-row-1-1"]').exists()).toBe(true);
      expect(dialog.find('[data-test="item-content-row-1-2"]').exists()).toBe(true);
    });

    it("renders feature text content correctly", () => {
      const content0 = dialog.find('[data-test="item-content-0"]');
      const content1 = dialog.find('[data-test="item-content-1"]');

      expect(content0.text()).toContain("Protection Against DDoS Attacks");
      expect(content0.text()).toContain("Session record and playback");
      expect(content1.text()).toContain("Dedicated server for each customer");
      expect(content1.text()).toContain("Supports up to thousands of devices");
    });

    it("renders checkmark icons for each feature", () => {
      const content0 = dialog.find('[data-test="item-content-0"]');
      const icons = content0.findAll(".mdi-check-circle");

      expect(icons.length).toBe(4); // 4 features in Cloud card
    });

    it("renders pricing buttons with correct props", () => {
      const pricingBtn0 = dialog.find('[data-test="pricing-btn-0"]');
      const pricingBtn1 = dialog.find('[data-test="pricing-btn-1"]');

      expect(pricingBtn0.exists()).toBe(true);
      expect(pricingBtn1.exists()).toBe(true);
      expect(pricingBtn0.text()).toBe("Pricing");
      expect(pricingBtn1.text()).toBe("Get a quote");
      expect(pricingBtn0.attributes("href")).toBe("https://www.shellhub.io/pricing");
      expect(pricingBtn1.attributes("href")).toBe("https://www.shellhub.io/pricing");
      expect(pricingBtn0.attributes("target")).toBe("_blank");
      expect(pricingBtn0.attributes("rel")).toBe("noreferrer noopener");
    });

    it("renders card actions sections", () => {
      expect(dialog.find('[data-test="item-actions-0"]').exists()).toBe(true);
      expect(dialog.find('[data-test="item-actions-1"]').exists()).toBe(true);
    });
  });

  describe("Component rendering with no premium features", () => {
    beforeEach(async () => {
      wrapper.unmount();
      document.body.innerHTML = "";
      await mountWrapper([], true);
      await flushPromises();
    });

    it("does not render items row when no features available", () => {
      expect(dialog.find('[data-test="items-row"]').exists()).toBe(false);
    });

    it("renders fallback button when no features are available", () => {
      const fallbackBtn = dialog.find('[data-test="no-link-available-btn"]');

      expect(fallbackBtn.exists()).toBe(true);
      expect(fallbackBtn.text()).toBe("Check out our website");
      expect(fallbackBtn.attributes("href")).toBe("https://www.shellhub.io");
      expect(fallbackBtn.attributes("target")).toBe("_blank");
      expect(fallbackBtn.attributes("rel")).toBe("noreferrer noopener");
    });

    it("does not render premium feature cards", () => {
      expect(dialog.find('[data-test="item-0"]').exists()).toBe(false);
      expect(dialog.find('[data-test="item-1"]').exists()).toBe(false);
    });
  });

  describe("Dialog close behavior", () => {
    beforeEach(async () => {
      await wrapper.setProps({ modelValue: true });
      usersStore.showPaywall = true;
      await flushPromises();
    });

    it("closes dialog and resets showPaywall when close is triggered", async () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });

      messageDialog.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
      expect(usersStore.showPaywall).toBe(false);
    });

    it("closes dialog and resets showPaywall when cancel is triggered", async () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });

      messageDialog.vm.$emit("cancel");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
      expect(usersStore.showPaywall).toBe(false);
    });
  });
});
