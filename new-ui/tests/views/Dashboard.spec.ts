import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Dashboard from "../../src/views/Dashboard.vue";
import { createStore } from "vuex";
import { key } from "../../src/store";
import routes from "../../src/router";

describe("Dashboard", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const numberNamespaces1 = 3;

  const namespace1 = {
    name: "namespace1",
    owner: "user1",
    members: [{ name: "user3" }, { name: "user4" }, { name: "user5" }],
    tenant_id: "a736a52b-5777-4f92-b0b8-e359bf484713",
  };

  const statsDev = {
    registered_devices: 2,
    pending_devices: 1,
    rejected_devices: 1,
  };

  const store = createStore({
    state: {
      stats: statsDev,
      namespace: namespace1,
      numberNamespaces: numberNamespaces1,
    },
    getters: {
      "stats/stats": (state) => state.stats,
      "namespaces/get": (state) => state.namespace,
    },
    actions: {
      "stats/get": vi.fn(),
      "users/setStatusUpdateAccountDialog": vi.fn(),
    },
  });

  beforeEach(async () => {
    wrapper = mount(Dashboard, {
      global: {
        plugins: [[store, key], vuetify, routes],
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

  it("Compare data with default value", () => {
    expect(wrapper.vm.itemsStats.registered_devices).toBe(2);
    expect(wrapper.vm.itemsStats.pending_devices).toBe(1);
    expect(wrapper.vm.itemsStats.rejected_devices).toBe(1);
    expect(wrapper.vm.hasStatus).toBeFalsy();
  });

  //////
  // HTML validation
  //////

  it("Renders the template with data", () => {
    const cards = wrapper.findAll('[data-test="dashboard-card"]');
    expect(wrapper.find('[data-test="dashboard-card"]').exists()).toBe(true);
    expect(cards.length).toBe(3);
  });

  //////
  // In this case is tested the items are empty
  //////

  it("Renders the template with data", () => {
    wrapper = mount(Dashboard, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
      setup() {
        return {
          item: [],
          itemsStats: {},
          hasStatus: true,
        };
      },
    });
    expect(wrapper.find('[data-test="dashboard-card"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="dashboard-failed"]').exists()).toBe(true);
  });
});
