import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, DOMWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import FirewallRuleAdd from "@/components/Firewall/FirewallRuleAdd.vue";
import useFirewallRulesStore from "@/store/modules/firewall_rules";
import { FormFilterOptions } from "@/interfaces/IFilter";
import handleError from "@/utils/handleError";

vi.mock("@/utils/permission", () => ({
  default: vi.fn().mockReturnValue(true),
}));

describe("FirewallRuleAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof FirewallRuleAdd>>;
  let dialog: DOMWrapper<Element>;
  let firewallRulesStore: ReturnType<typeof useFirewallRulesStore>;

  const openDialog = async () => {
    await wrapper.find('[data-test="add-firewall-rule-btn"]').trigger("click");
    await flushPromises();
  };

  const addFirewallRule = async () => {
    const addBtn = dialog.find('[data-test="firewall-rule-add-btn"]');
    await addBtn.trigger("click");
    await flushPromises();
  };

  beforeEach(() => {
    wrapper = mountComponent(FirewallRuleAdd, {});
    firewallRulesStore = useFirewallRulesStore();
    dialog = new DOMWrapper(document.body);
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Add button", () => {
    it("Renders add firewall rule button", () => {
      expect(wrapper.find('[data-test="add-firewall-rule-btn"]').exists()).toBe(true);
    });

    it("Displays Add Rule text", () => {
      const btn = wrapper.find('[data-test="add-firewall-rule-btn"]');
      expect(btn.text()).toBe("Add Rule");
    });

    it("Opens dialog when clicked", async () => {
      await openDialog();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(true);
    });
  });

  describe("Add dialog", () => {
    beforeEach(() => openDialog());

    it("Shows FormDialog with correct props", () => {
      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("title")).toBe("New Firewall Rule");
      expect(formDialog.props("icon")).toBe("mdi-shield-check");
      expect(formDialog.props("confirmText")).toBe("Add");
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
      expect(dialog.find('[data-test="firewall-rule-add-btn"]').exists()).toBe(true);
      expect(dialog.find('[data-test="firewall-rule-cancel"]').exists()).toBe(true);
    });
  });

  describe("Form fields - default values", () => {
    beforeEach(() => openDialog());

    it("Has default priority value", () => {
      const input = dialog.find('[data-test="firewall-rule-priority"] input').element as HTMLInputElement;
      expect(input.value).toBe("1");
    });

    it("Has default status as Active", () => {
      const select = dialog.find('[data-test="firewall-rule-status"]');
      expect(select.text()).toContain("Active");
    });

    it("Has default policy as Allow", () => {
      const select = dialog.find('[data-test="firewall-rule-policy"]');
      expect(select.text()).toContain("Allow");
    });
  });

  describe("Conditional field rendering - Source IP", () => {
    beforeEach(() => openDialog());

    it("Does not show source IP field by default", () => {
      expect(dialog.find('[data-test="firewall-rule-source-ip"]').exists()).toBe(false);
    });

    it("Shows source IP field when selectedIPOption is 'restrict'", async () => {
      wrapper.vm.selectedIPOption = "restrict";
      await wrapper.vm.$nextTick();

      expect(dialog.find('[data-test="firewall-rule-source-ip"]').exists()).toBe(true);
    });

    it("Calls handleSourceIpUpdate when source IP select changes", async () => {
      wrapper.vm.selectedIPOption = "restrict";
      await wrapper.vm.$nextTick();
      wrapper.vm.handleSourceIpUpdate();
      await wrapper.vm.$nextTick();

      const sourceIpField = dialog.find('[data-test="firewall-rule-source-ip"]');
      expect(sourceIpField.find("input").element.value).toBe("");
      expect(sourceIpField.text()).toContain("required");
    });
  });

  describe("Conditional field rendering - Username", () => {
    beforeEach(() => openDialog());

    it("Does not show username field by default", () => {
      expect(dialog.find('[data-test="firewall-rule-username-restriction"]').exists()).toBe(false);
    });

    it("Shows username field when selectedUsernameOption is 'username'", async () => {
      wrapper.vm.selectedUsernameOption = "username";
      await wrapper.vm.$nextTick();

      expect(dialog.find('[data-test="firewall-rule-username-restriction"]').exists()).toBe(true);
    });

    it("Calls handleUsernameUpdate when username select changes", async () => {
      wrapper.vm.selectedUsernameOption = "username";
      await wrapper.vm.$nextTick();
      wrapper.vm.handleUsernameUpdate();
      await wrapper.vm.$nextTick();

      const usernameField = dialog.find('[data-test="firewall-rule-username-restriction"]');
      expect(usernameField.find("input").element.value).toBe("");
      expect(usernameField.text()).toContain("required");
    });
  });

  describe("Conditional field rendering - Device filter", () => {
    beforeEach(() => openDialog());

    it("Does not show hostname field by default", () => {
      expect(dialog.find('[data-test="firewall-rule-hostname-restriction"]').exists()).toBe(false);
    });

    it("Shows hostname field when selectedFilterOption is 'hostname'", async () => {
      wrapper.vm.selectedFilterOption = FormFilterOptions.Hostname;
      await wrapper.vm.$nextTick();

      expect(dialog.find('[data-test="firewall-rule-hostname-restriction"]').exists()).toBe(true);
    });

    it("Calls handleFilterUpdate when filter is set to hostname", async () => {
      wrapper.vm.selectedFilterOption = FormFilterOptions.Hostname;
      await wrapper.vm.$nextTick();
      wrapper.vm.handleFilterUpdate();
      await wrapper.vm.$nextTick();

      const hostnameField = dialog.find('[data-test="firewall-rule-hostname-restriction"]');
      expect(hostnameField.find("input").element.value).toBe("");
      expect(hostnameField.text()).toContain("required");
    });

    it("Shows tags selector when selectedFilterOption is 'tags'", async () => {
      wrapper.vm.selectedFilterOption = FormFilterOptions.Tags;
      await wrapper.vm.$nextTick();

      expect(dialog.find('[data-test="tags-selector"]').exists()).toBe(true);
    });

    it("Calls handleFilterUpdate when filter is set to tags", async () => {
      wrapper.vm.selectedFilterOption = FormFilterOptions.Tags;
      await wrapper.vm.$nextTick();
      wrapper.vm.handleFilterUpdate();
      await wrapper.vm.$nextTick();

      const tagsComponent = dialog.find('[data-test="tags-selector"]');
      expect(tagsComponent.exists()).toBe(true);
      expect(wrapper.vm.selectedTags).toEqual([]);
      expect(wrapper.vm.selectedTagsError).toBe("");
    });
  });

  describe("Form validation", () => {
    beforeEach(() => openDialog());

    it("Shows error when priority is empty", async () => {
      const input = dialog.find('[data-test="firewall-rule-priority"] input');
      await input.setValue("");
      await flushPromises();

      expect(dialog.find('[data-test="firewall-rule-priority"]').text()).toContain("valid integer");
    });

    it("Shows error when priority is zero", async () => {
      const input = dialog.find('[data-test="firewall-rule-priority"] input');
      await input.setValue("0");
      await flushPromises();

      expect(dialog.find('[data-test="firewall-rule-priority"]').text()).toContain("cannot be zero");
    });
  });

  describe("Firewall rule creation", () => {
    beforeEach(() => openDialog());

    it("Calls createFirewallRule with default values", async () => {
      await addFirewallRule();

      expect(firewallRulesStore.createFirewallRule).toHaveBeenCalledWith({
        active: true,
        action: "allow",
        priority: 1,
        source_ip: ".*",
        username: ".*",
        filter: {
          hostname: ".*",
        },
      });
    });

    it("Calls createFirewallRule with custom priority and deny action", async () => {
      const priorityInput = dialog.find('[data-test="firewall-rule-priority"] input');
      await priorityInput.setValue("5");
      await flushPromises();

      wrapper.vm.action = "deny";

      await addFirewallRule();

      expect(firewallRulesStore.createFirewallRule).toHaveBeenCalledWith({
        active: true,
        action: "deny",
        priority: 5,
        source_ip: ".*",
        username: ".*",
        filter: {
          hostname: ".*",
        },
      });
    });

    it("Calls createFirewallRule with restricted source IP", async () => {
      const sourceIpSelect = dialog.find('[data-test="firewall-rule-source-ip-select"]');
      await sourceIpSelect.trigger("click");
      await flushPromises();

      wrapper.vm.selectedIPOption = "restrict";
      await flushPromises();

      const sourceIpInput = dialog.find('[data-test="firewall-rule-source-ip"] input');
      await sourceIpInput.setValue("192.168.1.0/24");
      await flushPromises();

      await addFirewallRule();

      expect(firewallRulesStore.createFirewallRule).toHaveBeenCalledWith(
        expect.objectContaining({
          source_ip: "192.168.1.0/24",
        }),
      );
    });

    it("Emits update event after successful creation", async () => {
      await addFirewallRule();

      expect(wrapper.emitted("update")).toBeTruthy();
    });

    it("Closes dialog after successful creation", async () => {
      await addFirewallRule();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(false);
    });

    it("Shows success snackbar after successful creation", async () => {
      await addFirewallRule();

      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Successfully created a new firewall rule.");
    });
  });

  describe("Dialog actions", () => {
    it("Closes dialog when Cancel is clicked", async () => {
      await openDialog();

      const cancelBtn = dialog.find('[data-test="firewall-rule-cancel"]');
      await cancelBtn.trigger("click");
      await flushPromises();

      const formDialog = wrapper.findComponent({ name: "FormDialog" });
      expect(formDialog.props("modelValue")).toBe(false);
    });
  });

  describe("Error handling", () => {
    it("Handles generic server error", async () => {
      const error = createAxiosError(500, "Internal server error");
      vi.mocked(firewallRulesStore.createFirewallRule).mockRejectedValueOnce(error);

      await openDialog();

      await addFirewallRule();

      expect(firewallRulesStore.createFirewallRule).toHaveBeenCalled();
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to create a new firewall rule.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
