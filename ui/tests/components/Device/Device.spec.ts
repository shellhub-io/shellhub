import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import Devices from "../../../src/components/Devices/Device.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const pendingDevices = 2;

const store = createStore({
  state: {
    stats: {
      registered_devices: 0,
      online_devices: 0,
      active_sessions: 0,
      pending_devices: pendingDevices,
      rejected_devices: 0,
    },
  },
  getters: {
    "stats/stats": (state) => state.stats,
  },
  actions: {
    "stats/get": vi.fn(),
    "devices/setFilter": vi.fn(),
    "devices/refresh": vi.fn(),
  },
});

describe("Device", () => {
  let wrapper: VueWrapper<InstanceType<typeof Devices>>;
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(Devices, {
      global: {
        plugins: [[store, key], vuetify, routes],
        stubs: ["router-link", "router-view"],
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
  // Data and Props checking
  //////

  it("Process data in the computed", () => {
    expect(wrapper.vm.tab).toBeTruthy();
  });
});
