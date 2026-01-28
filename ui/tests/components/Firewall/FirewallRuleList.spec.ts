import { createPinia, setActivePinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import FirewallRuleList from "@/components/Firewall/FirewallRuleList.vue";
import { rulesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useFirewallRulesStore from "@/store/modules/firewall_rules";
import { IFirewallRule } from "@/interfaces/IFirewallRule";

type FirewallRuleListWrapper = VueWrapper<InstanceType<typeof FirewallRuleList>>;

describe("Firewall Rule List", () => {
  let wrapper: FirewallRuleListWrapper;
  setActivePinia(createPinia());
  const firewallRulesStore = useFirewallRulesStore();
  const vuetify = createVuetify();
  const mockRulesApi = new MockAdapter(rulesApi.getAxios());

  const firewallRules = [
    {
      id: "1000",
      tenant_id: "00000000-0000-4000-0000-000000000000",
      priority: 1,
      action: "allow",
      active: true,
      source_ip: ".*",
      username: ".*",
      filter: {
        hostname: ".*",
      },
    },
    {
      id: "1001",
      tenant_id: "00000000-0000-4000-0000-000000000000",
      priority: 2,
      action: "allow",
      active: false,
      source_ip: ".*",
      username: ".*",
      filter: {
        hostname: ".*",
      },
    },
  ];

  const res = {
    data: firewallRules,
    headers: {
      "x-total-count": 2,
    },
  };

  beforeEach(() => {
    mockRulesApi.onGet("http://localhost:3000/api/firewall/rules?page=1&per_page=10").reply(200, res);
    firewallRulesStore.firewallRules = firewallRules as IFirewallRule[];

    wrapper = mount(FirewallRuleList, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
  });

  it("Renders the Firewall List", () => {
    expect(wrapper.find('[data-test="firewall-rules-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-active"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-priority"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-action"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-source-ip"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-username"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-filter"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-actions"]').exists()).toBe(true);
  });
});
