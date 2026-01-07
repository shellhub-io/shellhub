import { flushPromises, VueWrapper } from "@vue/test-utils";
import { describe, expect, it, beforeEach, afterEach } from "vitest";
import { mountComponent } from "@tests/utils/mount";
import FirewallRules from "@/views/FirewallRules.vue";
import { mockFirewallRules } from "@tests/views/mocks";

describe("Firewall Rules View", () => {
  let wrapper: VueWrapper<InstanceType<typeof FirewallRules>>;

  const mountWrapper = async (hasRules = true) => {
    const initialState = {
      firewallRules: {
        firewallRules: hasRules ? mockFirewallRules : [],
        firewallRuleCount: hasRules ? 1 : 0,
      },
    };

    wrapper = mountComponent(FirewallRules, { piniaOptions: { initialState } });

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
