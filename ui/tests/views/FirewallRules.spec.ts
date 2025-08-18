import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import FirewallRules from "@/views/FirewallRules.vue";
import { rulesApi } from "@/api/http";
import { store, key } from "@/store";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useFirewallRulesStore from "@/store/modules/firewall_rules";
import { IFirewallRule } from "@/interfaces/IFirewallRule";

type FirewallRulesWrapper = VueWrapper<InstanceType<typeof FirewallRules>>;

describe("Firewall Rules", () => {
  let wrapper: FirewallRulesWrapper;
  setActivePinia(createPinia());
  const firewallRulesStore = useFirewallRulesStore();
  const vuetify = createVuetify();

  const mockRulesApi = new MockAdapter(rulesApi.getAxios());

  const firewallRules = [
    {
      priority: 1,
      action: "allow",
      active: true,
      filter: {
        hostname: ".*",
      },
      source_ip: ".*",
      username: ".*",
    },
  ];

  beforeEach(async () => {
    mockRulesApi.onGet("http://localhost:3000/api/firewall/rules?page=1&per_page=10").reply(200, firewallRules);
    firewallRulesStore.firewallRules = firewallRules as IFirewallRule[];
    firewallRulesStore.firewallRuleCount = 1;
    wrapper = mount(FirewallRules, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="firewall-rules"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="help-icon"]').exists()).toBe(true);
  });

  it("Toggles showHelp when help icon is clicked", async () => {
    const helpIcon = wrapper.find('[data-test="help-icon"]');
    await helpIcon.trigger("click");
    expect(wrapper.vm.showHelp).toBe(true);

    await helpIcon.trigger("click");
    expect(wrapper.vm.showHelp).toBe(false);
  });

  it("Shows the no items message when there are no firewall rules", () => {
    mockRulesApi.onGet("http://localhost:3000/api/firewall/rules?page=1&per_page=10").reply(200, []);
    firewallRulesStore.firewallRules = [];
    firewallRulesStore.firewallRuleCount = 0;
    wrapper = mount(FirewallRules, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
    });
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-items-message-component"]').text()).toContain("Looks like you don't have any Firewall Rules");
  });
});
