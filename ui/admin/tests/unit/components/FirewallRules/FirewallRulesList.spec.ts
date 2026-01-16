import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import FirewallRulesList from "@admin/components/FirewallRules/FirewallRulesList.vue";
import { mockFirewallRules } from "../../mocks";
import { Router } from "vue-router";

describe("FirewallRulesList", () => {
  let wrapper: VueWrapper<InstanceType<typeof FirewallRulesList>>;
  let router: Router;
  let firewallRulesStore: ReturnType<typeof useFirewallRulesStore>;

  const mountWrapper = (mockFirewallRulesCount?: number) => {
    router = createCleanAdminRouter();

    wrapper = mountComponent(FirewallRulesList, {
      global: { plugins: [router] },
      piniaOptions: {
        initialState: {
          adminFirewallRules: {
            firewallRules: mockFirewallRules,
            firewallRulesCount: mockFirewallRulesCount ?? mockFirewallRules.length,
          },
        },
      },
    });

    firewallRulesStore = useFirewallRulesStore();
  };

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the data table", () => {
      expect(wrapper.find('[data-test="firewall-rules-list"]').exists()).toBe(true);
    });

    it("displays firewall rule tenant IDs", () => {
      expect(wrapper.text()).toContain(mockFirewallRules[0].tenant_id);
      expect(wrapper.text()).toContain(mockFirewallRules[1].tenant_id);
    });

    it("displays firewall rule priorities", () => {
      expect(wrapper.text()).toContain(mockFirewallRules[0].priority.toString());
      expect(wrapper.text()).toContain(mockFirewallRules[1].priority.toString());
    });

    it("displays firewall rule actions", () => {
      expect(wrapper.text()).toContain(mockFirewallRules[0].action);
      expect(wrapper.text()).toContain(mockFirewallRules[1].action);
    });

    it("displays firewall rule source IPs", () => {
      expect(wrapper.text()).toContain(mockFirewallRules[0].source_ip);
      expect(wrapper.text()).toContain(mockFirewallRules[1].source_ip);
    });

    it("displays firewall rule usernames", () => {
      expect(wrapper.text()).toContain(mockFirewallRules[0].username);
      expect(wrapper.text()).toContain(mockFirewallRules[1].username);
    });

    it("displays info buttons for each firewall rule", () => {
      const infoButtons = wrapper.findAll('[data-test="info-button"]');
      expect(infoButtons).toHaveLength(mockFirewallRules.length);
    });
  });

  describe("fetching firewall rules", () => {
    it("fetches firewall rules on mount", () => {
      mountWrapper();

      expect(firewallRulesStore.fetchFirewallRulesList).toHaveBeenCalledWith(
        expect.objectContaining({
          perPage: 10,
          page: 1,
        }),
      );
    });

    it("refetches firewall rules when page changes", async () => {
      mountWrapper(11); // Mock total count to 11 to enable pagination

      // Click next page button
      const nextPageBtn = wrapper.find('[data-test="pager-next"]');
      await nextPageBtn.trigger("click");
      await flushPromises();

      expect(firewallRulesStore.fetchFirewallRulesList).toHaveBeenCalledWith(
        expect.objectContaining({
          page: 2,
        }),
      );
    });

    it("refetches firewall rules when items per page changes", async () => {
      mountWrapper(20);

      // Change items per page via combobox
      const ippCombo = wrapper.find('[data-test="ipp-combo"] input');
      await ippCombo.setValue(20);
      await flushPromises();

      expect(firewallRulesStore.fetchFirewallRulesList).toHaveBeenCalledWith(
        expect.objectContaining({
          perPage: 20,
        }),
      );
    });
  });

  describe("navigating to firewall rule details", () => {
    it("navigates when clicking info button", async () => {
      mountWrapper();

      const pushSpy = vi.spyOn(router, "push");
      const infoButton = wrapper.findAll('[data-test="info-button"]')[0];

      await infoButton.trigger("click");

      expect(pushSpy).toHaveBeenCalledWith({
        name: "firewallRulesDetails",
        params: { id: mockFirewallRules[0].id },
      });
    });
  });

  describe("error handling", () => {
    it("shows error snackbar when fetching firewall rules fails", async () => {
      mountWrapper(11);
      vi.mocked(firewallRulesStore.fetchFirewallRulesList).mockRejectedValueOnce(
        createAxiosError(500, "Internal Server Error"),
      );

      // Trigger refetch by changing page
      const nextPageBtn = wrapper.find('[data-test="pager-next"]');
      await nextPageBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to fetch firewall rules.");
    });
  });
});
