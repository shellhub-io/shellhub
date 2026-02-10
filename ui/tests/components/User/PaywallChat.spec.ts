import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper, DOMWrapper } from "@vue/test-utils";
import PaywallChat from "@/components/User/PaywallChat.vue";
import { mountComponent } from "@tests/utils/mount";

describe("PaywallChat", () => {
  let wrapper: VueWrapper<InstanceType<typeof PaywallChat>>;
  let dialog: DOMWrapper<Element>;

  const mountWrapper = async (modelValue = false) => {
    wrapper = mountComponent(PaywallChat, {
      props: { modelValue },
      attachTo: document.body,
    });

    dialog = new DOMWrapper(document.body);
    await flushPromises();
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component rendering when closed", () => {
    it("does not render dialog when modelValue is false", async () => {
      await mountWrapper();
      expect(dialog.find('[data-test="paywall-chat-dialog"]').exists()).toBe(false);
    });
  });

  describe("Component rendering when open", () => {
    beforeEach(() => mountWrapper(true));

    it("renders MessageDialog with correct props", () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.exists()).toBe(true);
      expect(messageDialog.props("title")).toBe("Upgrade to have access to chat support!");
      expect(messageDialog.props("icon")).toBe("mdi-chat-question");
      expect(messageDialog.props("iconColor")).toBe("success");
      expect(messageDialog.props("cancelText")).toBe("Close");
      expect(messageDialog.props("confirmText")).toBe("Upgrade");
    });

    it("renders upgrade description paragraphs", () => {
      expect(dialog.find('[data-test="upgrade-description-1"]').exists()).toBe(true);
      expect(dialog.find('[data-test="upgrade-description-2"]').exists()).toBe(true);

      expect(dialog.text()).toContain("Get real-time assistance from our team");
      expect(dialog.text()).toContain("priority responses");
    });

    it("renders documentation link", () => {
      const link = dialog.find('[data-test="link-anchor"]');
      expect(link.exists()).toBe(true);
      expect(link.attributes("href")).toBe("https://docs.shellhub.io/");
      expect(link.attributes("target")).toBe("_blank");
      expect(link.attributes("rel")).toBe("noopener noreferrer");
      expect(link.text()).toBe("our Documentation");
    });

    it("renders close and upgrade buttons", () => {
      expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="upgrade-btn"]').exists()).toBe(true);
    });
  });

  describe("User interactions", () => {
    beforeEach(async () => mountWrapper(true));

    it("closes dialog when Close button is clicked", async () => {
      const closeBtn = dialog.find('[data-test="close-btn"]');

      await closeBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });

    it("opens pricing page when Upgrade button is clicked", async () => {
      const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);
      const upgradeBtn = dialog.find('[data-test="upgrade-btn"]');

      await upgradeBtn.trigger("click");
      await flushPromises();

      expect(openSpy).toHaveBeenCalledWith(
        "https://www.shellhub.io/pricing",
        "_blank",
        "noopener,noreferrer",
      );

      openSpy.mockRestore();
    });

    it("emits update:modelValue when dialog is closed via MessageDialog", async () => {
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });

      messageDialog.vm.$emit("cancel");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")![0]).toEqual([false]);
    });
  });
});
