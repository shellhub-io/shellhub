import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import FirewallRuleEdit from "@/components/Firewall/FirewallRuleEdit.vue";
import useFirewallRulesStore from "@/store/modules/firewall_rules";
import { IFirewallRule } from "@/interfaces/IFirewallRule";
import { mockFirewallRule, mockTags } from "@tests/mocks";
import handleError from "@/utils/handleError";

describe("FirewallRuleEdit", () => {
  let wrapper: VueWrapper<InstanceType<typeof FirewallRuleEdit>>;
  let dialog: DOMWrapper<Element>;
  let firewallRulesStore: ReturnType<typeof useFirewallRulesStore>;

  const mountWrapper = (rule: IFirewallRule = mockFirewallRule, hasAuthorization = true) => {
    wrapper = mountComponent(FirewallRuleEdit, {
      props: {
        firewallRule: rule,
        hasAuthorization,
      },
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
    it("Renders edit list item", () => {
      expect(wrapper.find('[data-test="firewall-edit-rule-btn"]').exists()).toBe(true);
    });

    it("Displays Edit text", () => {
      expect(wrapper.find('[data-test="mdi-information-list-item"]').text()).toBe("Edit");
    });

    it("Opens dialog when clicked", async () => {
      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(true);
    });

    it("Is disabled when hasAuthorization is false", () => {
      wrapper.unmount();
      mountWrapper(mockFirewallRule, false);

      const listItem = wrapper.find('[data-test="firewall-edit-rule-btn"]');
      expect(listItem.classes()).toContain("v-list-item--disabled");
    });
  });

  describe("Edit dialog", () => {
    beforeEach(async () => {
      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();
    });

    it("Shows FormDialog with correct props", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("title")).toBe("Edit Firewall Rule");
      expect(formDialog.props("icon")).toBe("mdi-shield-check");
      expect(formDialog.props("confirmText")).toBe("Edit");
      expect(formDialog.props("cancelText")).toBe("Cancel");
    });

    it("Renders all form fields", () => {
      expect(dialog.find('[data-test="firewall-rule-status"]').exists()).toBe(true);
      expect(dialog.find('[data-test="firewall-rule-priority"]').exists()).toBe(true);
      expect(dialog.find('[data-test="firewall-rule-policy"]').exists()).toBe(true);
      expect(dialog.find('[data-test="firewall-rule-source-ip-select"]').exists()).toBe(true);
      expect(dialog.find('[data-test="username-field"]').exists()).toBe(true);
      expect(dialog.find('[data-test="filter-select"]').exists()).toBe(true);
    });

    it("Renders action buttons", () => {
      expect(dialog.find('[data-test="firewall-rule-edit-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="firewall-rule-cancel"]').exists()).toBe(true);
    });
  });

  describe("Form fields - populated values", () => {
    it("Populates priority from firewall rule", async () => {
      wrapper.unmount();
      mountWrapper({ ...mockFirewallRule, priority: 5 });

      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      const input = dialog.find('[data-test="firewall-rule-priority"] input').element as HTMLInputElement;
      expect(input.value).toBe("5");
    });

    it("Populates action from firewall rule", async () => {
      wrapper.unmount();
      mountWrapper({ ...mockFirewallRule, action: "deny" });

      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      const select = dialog.find('[data-test="firewall-rule-policy"]');
      expect(select.text()).toContain("Deny");
    });

    it("Populates status from firewall rule", async () => {
      wrapper.unmount();
      mountWrapper({ ...mockFirewallRule, active: false });

      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      const select = dialog.find('[data-test="firewall-rule-status"]');
      expect(select.text()).toContain("Inactive");
    });
  });

  describe("Conditional field rendering - Source IP", () => {
    it("Shows source IP field when rule has restricted source IP", async () => {
      wrapper.unmount();
      mountWrapper({ ...mockFirewallRule, source_ip: "192.168.1.0/24" });

      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="firewall-rule-source-ip"]').exists()).toBe(true);
      const input = dialog.find('[data-test="firewall-rule-source-ip"] input').element as HTMLInputElement;
      expect(input.value).toBe("192.168.1.0/24");
    });

    it("Does not show source IP field when rule has .* source IP", async () => {
      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="firewall-rule-source-ip"]').exists()).toBe(false);
    });
  });

  describe("Conditional field rendering - Username", () => {
    it("Shows username field when rule has restricted username", async () => {
      wrapper.unmount();
      mountWrapper({ ...mockFirewallRule, username: "testuser" });

      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="firewall-rule-username-restriction"]').exists()).toBe(true);
      const input = dialog.find('[data-test="firewall-rule-username-restriction"] input').element as HTMLInputElement;
      expect(input.value).toBe("testuser");
    });

    it("Does not show username field when rule has .* username", async () => {
      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="firewall-rule-username-restriction"]').exists()).toBe(false);
    });
  });

  describe("Conditional field rendering - Device filter", () => {
    it("Shows hostname field when filter is hostname", async () => {
      wrapper.unmount();
      mountWrapper({ ...mockFirewallRule, filter: { hostname: "test-host" } });

      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="firewall-rule-hostname-restriction"]').exists()).toBe(true);
      const input = dialog.find('[data-test="firewall-rule-hostname-restriction"] input').element as HTMLInputElement;
      expect(input.value).toBe("test-host");
    });

    it("Shows tags selector when filter is tags", async () => {
      wrapper.unmount();
      mountWrapper({
        ...mockFirewallRule,
        filter: { tags: mockTags },
      });

      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(true);
    });
  });

  describe("Form validation", () => {
    beforeEach(async () => {
      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();
    });

    it("Shows error when priority is empty", async () => {
      const input = dialog.find('[data-test="firewall-rule-priority"] input');
      await input.setValue("");
      await flushPromises();

      expect(dialog.find('[data-test="firewall-rule-priority"]').text()).toContain("This must be a valid integer");
    });

    it("Shows error when priority is zero", async () => {
      const input = dialog.find('[data-test="firewall-rule-priority"] input');
      await input.setValue("0");
      await flushPromises();

      expect(dialog.find('[data-test="firewall-rule-priority"]').text()).toContain("cannot be zero");
    });
  });

  describe("Firewall rule update", () => {
    it("Calls updateFirewallRule with current values", async () => {
      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      const editBtn = dialog.find('[data-test="firewall-rule-edit-btn"]');
      await editBtn.trigger("click");
      await flushPromises();

      expect(firewallRulesStore.updateFirewallRule).toHaveBeenCalledWith({
        id: "rule-1",
        action: "allow",
        priority: 1,
        active: true,
        source_ip: ".*",
        username: ".*",
        filter: { hostname: ".*" },
      });
    });

    it("Calls updateFirewallRule with modified priority", async () => {
      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      const priorityInput = dialog.find('[data-test="firewall-rule-priority"] input');
      await priorityInput.setValue("10");
      await flushPromises();

      const editBtn = dialog.find('[data-test="firewall-rule-edit-btn"]');
      await editBtn.trigger("click");
      await flushPromises();

      expect(firewallRulesStore.updateFirewallRule).toHaveBeenCalledWith(
        expect.objectContaining({
          priority: 10,
        }),
      );
    });

    it("Emits update event after successful edit", async () => {
      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      const editBtn = dialog.find('[data-test="firewall-rule-edit-btn"]');
      await editBtn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("Closes dialog after successful edit", async () => {
      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      const editBtn = dialog.find('[data-test="firewall-rule-edit-btn"]');
      await editBtn.trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Dialog actions", () => {
    it("Closes dialog when Cancel is clicked", async () => {
      await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
      await flushPromises();

      const cancelBtn = dialog.find('[data-test="firewall-rule-cancel"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(false);
    });

    describe("Error handling", () => {
      it("Handles generic server error", async () => {
        const error = createAxiosError(500, "Internal Server Error");
        vi.mocked(firewallRulesStore.updateFirewallRule).mockRejectedValueOnce(error);

        await wrapper.find('[data-test="firewall-edit-rule-btn"]').trigger("click");
        await flushPromises();

        const editBtn = dialog.find('[data-test="firewall-rule-edit-btn"]');
        await editBtn.trigger("click");
        await flushPromises();

        expect(firewallRulesStore.updateFirewallRule).toHaveBeenCalled();
        expect(mockSnackbar.showError).toHaveBeenCalledWith("Error while updating firewall rule.");
        expect(handleError).toHaveBeenCalledWith(error);
      });
    });
  });
});
