import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import Devices from "../../src/views/Devices.vue";
import { key } from "../../src/store";
import routes from "../../src/router";

describe("Devices", () => {
  let wrapper: VueWrapper<InstanceType<typeof Devices>>;
  const vuetify = createVuetify();

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

  beforeEach(async () => {
    wrapper = mount(Devices, {
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
  // Data and Props checking
  //////

  it("Process data in the computed", () => {
    expect(wrapper.vm.hasDevice).toEqual(true);
    expect(wrapper.vm.showMessageBox).toEqual(false);
    expect(wrapper.vm.isDeviceList).toEqual(false);
  });

  it("Compare data with the default and defined value", () => {
    expect(wrapper.vm.filter).toEqual("");

    wrapper.vm.filter = "ShellHub";

    expect(wrapper.vm.filter).toEqual("ShellHub");
  });

  //////
  // HTML validation
  //////

  it("Renders the template with components", () => {
    expect(wrapper.find('[data-test="tagSelector-component"]').exists()).toBe(
      false,
    );
    expect(
      wrapper.find('[data-test="boxMessageDevice-component"]').exists(),
    ).toBe(false);
    expect(wrapper.find('[data-test="search-text"]').exists()).toBe(true);
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="search-text"]').exists()).toBe(true);
    const textInputSearch = wrapper.find('[data-test="search-text"]');
    textInputSearch.element.textContent = "ShellHub";
    expect(textInputSearch.text()).toBe("ShellHub");
  });
});
