import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import FirewallRules from "../../../../src/views/FirewallRules.vue";

describe("Firewall Rules", () => {
  const store = createStore({
    state: {},
    getters: {
      "firewallRules/numberFirewalls": () => 1,
    },
    actions: {
      "firewallRules/fetch": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(FirewallRules, {
      global: {
        plugins: [[store, key], vuetify, routes],
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
    expect(wrapper.find("h1").text()).toContain("Firewall Rules");
    expect(wrapper.find("[data-test='firewallRules-list']").exists()).toBe(true);
  });
});
