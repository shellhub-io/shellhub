import { describe, expect, it, afterEach, beforeEach, vi } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import FirewallRuleList from "@/components/Firewall/FirewallRuleList.vue";
import useFirewallRulesStore from "@/store/modules/firewall_rules";
import { IFirewallRule } from "@/interfaces/IFirewallRule";
import { createAxiosError } from "@tests/utils/axiosError";
import handleError from "@/utils/handleError";
import { mockFirewallRules } from "@tests/mocks";

describe("FirewallRuleList", () => {
  let wrapper: VueWrapper<InstanceType<typeof FirewallRuleList>>;
  let firewallRulesStore: ReturnType<typeof useFirewallRulesStore>;

  const mountWrapper = (rules: IFirewallRule[] = mockFirewallRules) => {
    wrapper = mountComponent(FirewallRuleList, {
      piniaOptions: {
        initialState: {
          firewallRules: {
            firewallRules: rules,
            firewallRuleCount: rules.length,
          },
        },
      },
    });

    firewallRulesStore = useFirewallRulesStore();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  describe("Table data rendering", () => {
    it("Displays all firewall rules", () => {
      const rows = wrapper.findAll("tbody tr");
      expect(rows).toHaveLength(3);
    });

    it("Displays priority values correctly", () => {
      const priorityCells = wrapper.findAll('[data-test="firewall-rules-priority"]');
      expect(priorityCells[0].text()).toBe("1");
      expect(priorityCells[1].text()).toBe("2");
      expect(priorityCells[2].text()).toBe("3");
    });

    it("Displays action values with capitalization", () => {
      const actionCells = wrapper.findAll('[data-test="firewall-rules-action"]');
      expect(actionCells[0].text()).toBe("Allow");
      expect(actionCells[1].text()).toBe("Deny");
      expect(actionCells[2].text()).toBe("Allow");
    });

    it("Displays source IP values correctly", () => {
      const sourceIpCells = wrapper.findAll('[data-test="firewall-rules-source-ip"]');
      expect(sourceIpCells[0].text()).toContain("Any IP");
      expect(sourceIpCells[1].text()).toBe("192.168.1.1");
      expect(sourceIpCells[2].text()).toBe("10.0.0.0/8");
    });

    it("Displays username values correctly", () => {
      const usernameCells = wrapper.findAll('[data-test="firewall-rules-username"]');
      expect(usernameCells[0].text()).toContain("All");
      expect(usernameCells[1].text()).toBe("testuser");
      expect(usernameCells[2].text()).toBe("admin");
    });

    it("Displays hostname filter correctly", () => {
      const filterCells = wrapper.findAll('[data-test="firewall-rules-filter"]');
      expect(filterCells[0].text()).toContain("All");
      expect(filterCells[2].text()).toBe("server-.*");
    });

    it("Displays tags in filter column", () => {
      const filterCells = wrapper.findAll('[data-test="firewall-rules-filter"]');
      expect(filterCells[1].text()).toContain("tag-1");
      expect(filterCells[1].text()).toContain("tag-2");
    });
  });

  describe("Active status icons", () => {
    it("Shows check icons for active column", () => {
      const activeIcons = wrapper.findAll('[data-test="firewall-rules-active"]');
      expect(activeIcons).toHaveLength(3);
    });

    it("Applies success color to active rules", () => {
      const activeIcons = wrapper.findAll('[data-test="firewall-rules-active"]');
      expect(activeIcons[0].classes()).toContain("text-success");
      expect(activeIcons[2].classes()).toContain("text-success");
    });

    it("Does not apply success color to inactive rules", () => {
      const activeIcons = wrapper.findAll('[data-test="firewall-rules-active"]');
      expect(activeIcons[1].classes()).not.toContain("text-success");
    });
  });

  describe("Actions menu", () => {
    it("Renders action button for each rule", () => {
      const actionButtons = wrapper.findAll('[data-test="firewall-rules-actions"]');
      expect(actionButtons).toHaveLength(3);
    });

    it("Opens actions menu when button is clicked", async () => {
      const actionButton = wrapper.find('[data-test="firewall-rules-actions"]');
      await actionButton.trigger("click");
      await flushPromises();

      expect(wrapper.findComponent({ name: "FirewallRuleEdit" }).exists()).toBe(true);
      expect(wrapper.findComponent({ name: "FirewallRuleDelete" }).exists()).toBe(true);
    });
  });

  describe("Empty state", () => {
    it("Shows no data when firewall rules list is empty", () => {
      wrapper.unmount();
      mountWrapper([]);

      const priorityCells = wrapper.findAll('[data-test="firewall-rules-priority"]');
      expect(priorityCells).toHaveLength(0);
    });
  });

  describe("Error handling", () => {
    it("Displays error message when user lacks permission", async () => {
      const error = createAxiosError(403, "Forbidden");

      mountWrapper();
      vi.mocked(firewallRulesStore.fetchFirewallRuleList).mockRejectedValueOnce(error);

      await wrapper.find('[data-test="ipp-combo"] input').setValue(20);

      expect(mockSnackbar.showError).toHaveBeenCalledWith("You don't have permission to access this resource.");
      expect(handleError).toHaveBeenCalledWith(error);
    });

    it("Displays error message for other errors", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      mountWrapper();
      vi.mocked(firewallRulesStore.fetchFirewallRuleList).mockRejectedValueOnce(error);

      await wrapper.find('[data-test="ipp-combo"] input').setValue(20);

      expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while loading the firewall rules.");
      expect(handleError).toHaveBeenCalledWith(error);
    });
  });
});
