import { createPinia, setActivePinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import { store, key } from "@/store";
import FirewallRuleList from "@/components/firewall/FirewallRuleList.vue";
import { rulesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type FirewallRuleListWrapper = VueWrapper<InstanceType<typeof FirewallRuleList>>;

describe("Firewall Rule List", () => {
  let wrapper: FirewallRuleListWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockFirewallApi = new MockAdapter(rulesApi.getAxios());

  const firewallRule = [
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
    data: firewallRule,
    headers: {
      "x-total-count": 2,
    },
  };

  beforeEach(async () => {
    mockFirewallApi.onGet("http://localhost:3000/api/firewall/rules?page=1&per_page=10").reply(200, res);
    store.commit("firewallRules/setFirewalls", res);

    wrapper = mount(FirewallRuleList, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the Firewall List", async () => {
    expect(wrapper.find('[data-test="firewallRules-list"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-active"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-priority"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-action"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-source-ip"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-username"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-filter"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="firewall-rules-actions"]').exists()).toBe(true);
  });
});
