import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import FirewallRuleDelete from "@/components/Firewall/FirewallRuleDelete.vue";
import useFirewallRulesStore from "@/store/modules/firewall_rules";
import handleError from "@/utils/handleError";

describe("FirewallRuleDelete", () => {
  let wrapper: VueWrapper<InstanceType<typeof FirewallRuleDelete>>;
  let dialog: DOMWrapper<Element>;
  let firewallRulesStore: ReturnType<typeof useFirewallRulesStore>;

  const mountWrapper = (hasAuthorization = true) => {
    wrapper = mountComponent(FirewallRuleDelete, {
      props: { id: "test-rule-id", hasAuthorization },
    });

    firewallRulesStore = useFirewallRulesStore();
    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("List item", () => {
    it("Renders delete list item", () => {
      expect(wrapper.find('[data-test="firewall-delete-dialog-btn"]').exists()).toBe(true);
    });

    it("Displays remove icon", () => {
      expect(wrapper.find('[data-test="remove-icon"]').exists()).toBe(true);
    });

    it("Displays Remove text", () => {
      expect(wrapper.find('[data-test="remove-title"]').text()).toBe("Remove");
    });

    it("Opens dialog when clicked", async () => {
      await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(true);
    });

    it("Is disabled when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper(false);

      const listItem = wrapper.find('[data-test="firewall-delete-dialog-btn"]');
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });
  });

  describe("Delete dialog", () => {
    it("Shows confirmation dialog with correct title", async () => {
      await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("title")).toBe("Are you sure?");
    });

    it("Shows description about deletion", async () => {
      await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("description")).toBe("You are about to delete this firewall rule");
    });

    it("Displays error icon", async () => {
      await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("icon")).toBe("mdi-alert");
      expect(messageDialog.props("iconColor")).toBe("error");
    });

    it("Shows Delete and Close buttons", async () => {
      await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("confirmText")).toBe("Delete");
      expect(messageDialog.props("confirmColor")).toBe("error");
      expect(messageDialog.props("cancelText")).toBe("Close");
    });
  });

  describe("Firewall rule deletion", () => {
    it("Calls removeFirewallRule when confirmed", async () => {
      await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(firewallRulesStore.removeFirewallRule).toHaveBeenCalledWith("test-rule-id");
    });

    it("Emits update event after successful deletion", async () => {
      await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("Closes dialog after successful deletion", async () => {
      await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });

    it("Handles deletion error and closes dialog", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.mocked(firewallRulesStore.removeFirewallRule).mockRejectedValueOnce(error);

      await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");
      await flushPromises();

      const confirmBtn = dialog.find('[data-test="confirm-btn"]');
      await confirmBtn.trigger("click");
      await flushPromises();

      expect(firewallRulesStore.removeFirewallRule).toHaveBeenCalledWith("test-rule-id");
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to delete firewall rule.");
      expect(handleError).toHaveBeenCalledWith(error);
      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });

    it("Closes dialog when Cancel is clicked", async () => {
      await wrapper.find('[data-test="firewall-delete-dialog-btn"]').trigger("click");
      await flushPromises();

      const closeBtn = dialog.find('[data-test="close-btn"]');
      await closeBtn.trigger("click");
      await flushPromises();

      const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
      expect(messageDialog.props("modelValue")).toBe(false);
    });
  });
});
