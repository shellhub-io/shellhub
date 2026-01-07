import { flushPromises, VueWrapper } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import createCleanRouter from "@tests/utils/router";
import FirewallRules from "@/views/FirewallRules.vue";
import { mockFirewallRules } from "@tests/views/mocks";

type FirewallRulesWrapper = VueWrapper<InstanceType<typeof FirewallRules>>;

describe("Firewall Rules View", () => {
  let wrapper: FirewallRulesWrapper;
  const router = createCleanRouter();

  const mountWrapper = async (hasRules = true) => {
    const initialState = hasRules
      ? {
        firewallRules: {
          firewallRules: mockFirewallRules,
          firewallRuleCount: 1,
        },
      }
      : {
        firewallRules: {
          firewallRules: [],
          firewallRuleCount: 0,
        },
      };

    wrapper = mountComponent(FirewallRules, {
      global: { plugins: [router] },
      piniaOptions: { initialState },
    }) as FirewallRulesWrapper;

    await flushPromises();
  };

  afterEach(() => { wrapper?.unmount(); });

  describe("when firewall rules exist", () => {
    beforeEach(async () => { await mountWrapper(); });

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="firewall-rules-header"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Firewall Rules");
      expect(pageHeader.find('[data-test="add-firewall-rule-btn"]').exists()).toBe(true);
    });

    it("displays the firewall rules list", () => {
      expect(wrapper.find('[data-test="firewall-rules-list"]').exists()).toBe(true);
    });

    it("does not show the no items message", () => {
      expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(false);
    });
  });

  describe("when no firewall rules exist", () => {
    beforeEach(async () => { await mountWrapper(false); });

    it("renders the page header", () => {
      const pageHeader = wrapper.find('[data-test="firewall-rules-header"]');
      expect(pageHeader.exists()).toBe(true);
      expect(pageHeader.text()).toContain("Firewall Rules");
      expect(pageHeader.find('[data-test="add-firewall-rule-btn"]').exists()).toBe(true);
    });

    it("does not display the firewall rules list", () => {
      expect(wrapper.find('[data-test="firewall-rules-list"]').exists()).toBe(false);
    });

    it("shows the no items message", () => {
      const noItemsMessage = wrapper.find('[data-test="no-items-message-component"]');
      expect(noItemsMessage.exists()).toBe(true);
      expect(noItemsMessage.text()).toContain("Firewall Rules");
      expect(noItemsMessage.find('[data-test="add-firewall-rule-btn"]').exists()).toBe(true);
    });
  });
});
