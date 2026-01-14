import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import { formatSourceIP, formatUsername } from "@/utils/string";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import FirewallRulesDetails from "@admin/views/FirewallRulesDetails.vue";
import { mockFirewallRule } from "../mocks";

vi.mock("@admin/store/api/firewall_rules");

describe("FirewallRulesDetails", () => {
  let wrapper: VueWrapper<InstanceType<typeof FirewallRulesDetails>>;

  const mountWrapper = async (mockError?: Error) => {
    const router = createCleanAdminRouter();
    await router.push({ name: "firewallRulesDetails", params: { id: mockFirewallRule.id } });
    await router.isReady();

    wrapper = mountComponent(FirewallRulesDetails, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: { adminFirewallRules: mockError ? {} : { firewallRule: mockFirewallRule } },
        stubActions: !mockError,
      },
    });

    const firewallRulesStore = useFirewallRulesStore();
    if (mockError) vi.mocked(firewallRulesStore.fetchFirewallRuleById).mockRejectedValueOnce(mockError);

    await flushPromises();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("when firewall rule loads successfully", () => {
    beforeEach(() => mountWrapper());

    it("displays the rule priority in the card title", () => {
      expect(wrapper.find(".text-h6").text()).toBe(`Rule #${mockFirewallRule.priority}`);
    });

    it("shows active status icon", () => {
      const icon = wrapper.find('[data-test="active-icon"]');
      expect(icon.exists()).toBe(true);
    });

    it("shows action chip with correct value", () => {
      const actionChip = wrapper.find('[data-test="firewall-action-chip"]');
      expect(actionChip.exists()).toBe(true);
      expect(actionChip.text()).toBe(mockFirewallRule.action);
    });

    it("displays firewall rule id", () => {
      const idField = wrapper.find('[data-test="firewall-id-field"]');
      expect(idField.text()).toContain("ID:");
      expect(idField.text()).toContain(mockFirewallRule.id);
    });

    it("displays priority", () => {
      const priorityField = wrapper.find('[data-test="firewall-priority-field"]');
      expect(priorityField.text()).toContain("Priority:");
      expect(priorityField.text()).toContain(String(mockFirewallRule.priority));
    });

    it("displays namespace with link", () => {
      const tenantField = wrapper.find('[data-test="firewall-tenant-field"]');
      expect(tenantField.text()).toContain("Namespace:");
      const link = tenantField.find("a");
      expect(link.exists()).toBe(true);
      expect(link.text()).toBe(mockFirewallRule.tenant_id);
    });

    it("displays source ip", () => {
      const sourceIpField = wrapper.find('[data-test="firewall-source-ip-field"]');
      expect(sourceIpField.text()).toContain("Source IP:");
      expect(sourceIpField.text()).toContain(formatSourceIP(mockFirewallRule.source_ip));
    });

    it("displays username", () => {
      const usernameField = wrapper.find('[data-test="firewall-username-field"]');
      expect(usernameField.text()).toContain("Username:");
      expect(usernameField.text()).toContain(formatUsername(mockFirewallRule.username));
    });

    it("displays filter information", () => {
      const filterField = wrapper.find('[data-test="firewall-filter-field"]');
      expect(filterField.exists()).toBe(true);
      expect(filterField.text()).toContain("Filter:");
    });
  });

  describe("when firewall rule fails to load", () => {
    it("shows error snackbar", async () => {
      await mountWrapper(createAxiosError(404, "Not Found"));

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to get firewall rule details.");
    });
  });
});
