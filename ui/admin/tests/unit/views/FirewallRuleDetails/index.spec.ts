import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useFirewallRulesStore from "@admin/store/modules/firewall_rules";
import routes from "@admin/router";
import FirewallRulesDetails from "@admin/views/FirewallRulesDetails.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type FirewallRulesDetailsWrapper = VueWrapper<InstanceType<typeof FirewallRulesDetails>>;

const firewallRuleDetail = {
  id: "6256b876e5c1d9bbdf954662",
  tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  priority: 1,
  action: "allow" as const,
  active: true,
  source_ip: ".*",
  username: "^[A-a]",
  filter: {
    hostname: ".*",
  },
};

const mockRoute = {
  params: {
    id: firewallRuleDetail.id,
  },
};

describe("Firewall Rule Details", () => {
  let wrapper: FirewallRulesDetailsWrapper;

  beforeEach(async () => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const firewallStore = useFirewallRulesStore();
    vi.spyOn(firewallStore, "getFirewall", "get").mockReturnValue(firewallRuleDetail);
    firewallStore.get = vi.fn().mockResolvedValue(firewallRuleDetail);

    const vuetify = createVuetify();

    wrapper = mount(FirewallRulesDetails, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
        mocks: {
          $route: mockRoute,
        },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Has the correct data", async () => {
    expect(wrapper.vm.firewallRule).toEqual(firewallRuleDetail);
  });

  it("Render the correct title", () => {
    expect(wrapper.find("h1").text()).toBe("Firewall Details");
  });

  it("Should render the props of the FirewallRule in the Screen", () => {
    expect(wrapper.find(`[data-test='${firewallRuleDetail.id}']`).text()).toContain(firewallRuleDetail.id);
    expect(wrapper.find(`[data-test='${firewallRuleDetail.tenant_id}']`).text()).toContain(firewallRuleDetail.tenant_id);
    expect(wrapper.find(`[data-test='${firewallRuleDetail.priority}']`).text()).toContain(String(firewallRuleDetail.priority));
    expect(wrapper.find(`[data-test='${firewallRuleDetail.action}']`).text()).toContain(firewallRuleDetail.action);
    expect(wrapper.find(`[data-test='${firewallRuleDetail.source_ip}']`).exists()).toBe(true);
    expect(wrapper.find(`[data-test='${firewallRuleDetail.username}']`).exists()).toBe(true);
    expect(wrapper.find(`[data-test='${firewallRuleDetail.filter}']`).exists()).toBe(true);
  });
});
