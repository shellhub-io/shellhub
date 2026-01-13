import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import routes from "@admin/router";
import FirewallRules from "@admin/views/FirewallRules.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type FirewallRulesWrapper = VueWrapper<InstanceType<typeof FirewallRules>>;

const firewallRules = [
  {
    action: "allow" as const,
    active: true,
    filter: {
      tags: [
        {
          tenant_id: "fake-tenant-data",
          name: "test-tag",
          created_at: "",
          updated_at: "",
        },
      ],
    },
    id: "5f1996c84d2190a22d5857bb",
    tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    priority: 4,
    source_ip: "127.0.0.1",
    username: "shellhub",
  },
];

describe("Firewall Rules", () => {
  let wrapper: FirewallRulesWrapper;

  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const firewallStore = useFirewallRulesStore();
    firewallStore.fetchFirewallRulesList = vi.fn().mockImplementation(() => {
      firewallStore.firewallRules = firewallRules;
      firewallStore.firewallRulesCount = 1;
    });

    const vuetify = createVuetify();

    wrapper = mount(FirewallRules, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Should render all the components in the screen", () => {
    expect(wrapper.find("[data-test='firewall-rules-list']").exists()).toBe(true);
  });
});
