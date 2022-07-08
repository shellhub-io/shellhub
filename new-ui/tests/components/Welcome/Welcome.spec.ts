import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Welcome from "../../../src/components/Welcome/Welcome.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("Welcome", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const show = true;
  const tenant = "a582b47a42e";

  const stats = {
    registered_devices: 0,
    online_devices: 0,
    active_sessions: 0,
    pending_devices: 1,
    rejected_devices: 0,
  };

  const devicePending = {
    uid: "a582b47a",
    name: "39-5e-2b",
    identity: {
      mac: "00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux",
      version: "",
    },
    public_key: "xxxxxxxx",
    tenant_id: "xxxxxxxx",
    last_seen: "2020-05-20T19:58:53.276Z",
    online: true,
    namespace: "user",
    status: "accepted",
  };

  const store = createStore({
    state: {
      tenant,
      stats,
      devicePending,
    },
    getters: {
      "auth/tenant": (state) => state.tenant,
      "stats/stats": (state) => state.stats,
      "devices/getFirstPending": (state) => state.devicePending,
    },
    actions: {
      "stats/get": vi.fn(),
      "devices/accept": vi.fn(),
      "notifications/fetch": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
      "snackbar/showSnackbarErrorDefault": vi.fn(),
    },
  });

  beforeEach(() => {
    wrapper = mount(Welcome, {
      global: {
        plugins: [[store, key], routes, vuetify],
      },
      props: {
        show,
      },
      // shallow: true,
    });
    window.location.protocol = "http:";
    window.location.hostname = "localhost";
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
  it("Receive data in props", () => {
    expect(wrapper.vm.show).toEqual(show);
  });
  it("Compare data with default value", () => {
    expect(wrapper.vm.el).toEqual(1);
    expect(wrapper.vm.enable).toEqual(false);
    expect(wrapper.vm.polling).toEqual(null);
  });
  it("Process data in the computed", () => {
    expect(wrapper.vm.showWelcome).toEqual(show);
  });
  it("Process data in the methods", () => {
    const command = `curl -sSf "http://localhost/install.sh?tenant_id=${tenant}" | sh`;

    expect(wrapper.vm.command()).toEqual(command);
  });

  it('Compare data with default value', async () => {
    expect(wrapper.vm.el).toEqual(1);
    expect(wrapper.vm.enable).toEqual(false);
    expect(wrapper.vm.curl.hostname).toEqual('localhost');
    expect(wrapper.vm.curl.tenant).toEqual(tenant);

    //////
    // In this case is tested the click event.
    //////
    // todo
  });
});
