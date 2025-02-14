import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import FirewallRulesDetails from "../../../../src/views/FirewallRulesDetails.vue";

const firewallRuleDetail = {
  id: "6256b876e5c1d9bbdf954662",
  tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  priority: 1,
  action: "allow",
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
  const store = createStore({
    state: {
      device: firewallRuleDetail,
    },
    getters: {
      "firewallRules/get": () => firewallRuleDetail,
    },
    actions: {
      "firewallRules/get": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(FirewallRulesDetails, {
      global: {
        plugins: [[store, key], vuetify, routes],
        mocks: {
          $route: mockRoute,
        },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Has the correct data", async () => {
    expect(wrapper.vm.firewallRule).toEqual(firewallRuleDetail);
  });

  it("Render the correct title", () => {
    expect(wrapper.find("h1").text()).toEqual("Firewall Details");
  });

  it("Should render the props of the FirewallRule in the Screen", () => {
    expect(wrapper.find(`[data-test='${firewallRuleDetail.id}']`).text()).toContain(firewallRuleDetail.id);
    expect(wrapper.find(`[data-test='${firewallRuleDetail.tenant_id}']`).text()).toContain(firewallRuleDetail.tenant_id);
    expect(wrapper.find(`[data-test='${firewallRuleDetail.priority}']`).text()).toContain(firewallRuleDetail.priority);
    expect(wrapper.find(`[data-test='${firewallRuleDetail.action}']`).text()).toContain(firewallRuleDetail.action);
    expect(wrapper.find(`[data-test='${firewallRuleDetail.source_ip}']`).exists()).toBe(true);
    expect(wrapper.find(`[data-test='${firewallRuleDetail.username}']`).exists()).toBe(true);
    expect(wrapper.find(`[data-test='${firewallRuleDetail.filter}']`).exists()).toBe(true);
  });
});
