import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import FirewallRules from "../../src/views/FirewallRules.vue";
import { createStore } from "vuex";
import { key } from "../../src/store";
import routes from "../../src/router";

describe("FirewallRules", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const numberFirewallsEqualZero = 0;
  const numberFirewallsGreaterThanZero = 1;

  const actionsMock = {
    "box/setStatus": vi.fn(),
    "firewallRules/resetPagePerpage": vi.fn(),
    "firewallRules/refresh": vi.fn(),
    "snackbar/showSnackbarErrorLoading": vi.fn(),
    "tags/fetch": vi.fn(),
    "firewallRules/fetch": vi.fn(),
  };

  const storeWithoutFirewalls = createStore({
    state: {
      numberFirewalls: numberFirewallsEqualZero,
    },
    getters: {
      "firewallRules/getNumberFirewalls": (state) => state.numberFirewalls,
    },
    actions: actionsMock,
  });

  const storeWithFirewalls = createStore({
    state: {
      numberFirewalls: numberFirewallsGreaterThanZero,
    },
    getters: {
      "firewallRules/getNumberFirewalls": (state) => state.numberFirewalls,
    },
    actions: actionsMock,
  });

  ///////
  // In this case, the rendering of the component that shows the
  // message when it does not have access to the device is tested.
  ///////

  describe("Without firewall rules", () => {
    beforeEach(async () => {
      wrapper = mount(FirewallRules, {
        global: {
          plugins: [[storeWithoutFirewalls, key], vuetify, routes],
        },
      });
    });

    ///////
    // Component Rendering
    //////

    it("Is a Vue instance", () => {
      expect(wrapper).toBeTruthy();
    });
    it("Renders the component", () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data checking
    //////
    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });
    it("Process data in the computed", () => {
      expect(wrapper.vm.hasFirewallRule).toEqual(false);
      expect(wrapper.vm.showBoxMessage).toEqual(true);
    });
    it("Compare data with the default", () => {
      expect(wrapper.vm.show).toEqual(true);
      expect(wrapper.vm.showHelp).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it("Renders the template with components", () => {
      expect(wrapper.find('[data-test="device-add-btn"]').exists()).toBe(true);
      expect(
        wrapper.find('[data-test="BoxMessageFirewall-component"]').exists()
      ).toBe(true);
    });
  });

  ///////
  // In this case, it is tested when there is already a registered
  // firewall.
  ///////

  describe("With firewall rules", () => {
    beforeEach(async () => {
      wrapper = mount(FirewallRules, {
        global: {
          plugins: [[storeWithFirewalls, key], vuetify, routes],
        },
      });
    });

    ///////
    // Component Rendering
    //////

    it("Is a Vue instance", () => {
      expect(wrapper).toBeTruthy();
    });
    it("Renders the component", () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////
    // Data checking
    //////
    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });
    it("Process data in the computed", () => {
      expect(wrapper.vm.hasFirewallRule).toEqual(true);
      expect(wrapper.vm.showBoxMessage).toEqual(false);
    });
    it("Compare data with the default", () => {
      expect(wrapper.vm.show).toEqual(true);
      expect(wrapper.vm.showHelp).toEqual(false);
    });

    //////
    // HTML validation
    //////

    it("Renders the template with components", () => {
      expect(wrapper.find('[data-test="device-add-btn"]').exists()).toBe(true);
      expect(
        wrapper.find('[data-test="BoxMessageFirewall-component"]').exists()
      ).toBe(false);
    });
  });
});
