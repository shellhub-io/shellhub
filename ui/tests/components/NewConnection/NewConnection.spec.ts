import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import NewConnection from "../../../src/components/NewConnection/NewConnection.vue";
import { key } from "../../../src/store";

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

describe("NewConnection", () => {
  let wrapper: VueWrapper<InstanceType<typeof NewConnection>>;
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(NewConnection, {
      global: {
        plugins: [[store, key], vuetify],
      },
      props: {
        size: "default",
      },
    });
  });

  // Component Rendering
  it("is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  // Data and Props checking
  it("has a default size prop of 'default'", () => {
    expect(wrapper.props("size")).toBe("default");
  });

  it("Renders the New Connection Button", async () => {
    expect(wrapper.find('[data-test="new-connection-add-btn"]'));
  });
});
