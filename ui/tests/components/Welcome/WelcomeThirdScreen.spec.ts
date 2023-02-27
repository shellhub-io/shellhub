import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import WelcomeThirdScreen from "../../../src/components/Welcome/WelcomeThirdScreen.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("WelcomeThirdScreen", () => {
  let wrapper: VueWrapper<InstanceType<typeof WelcomeThirdScreen>>;
  const vuetify = createVuetify();

  const device = {
    uid: "a582b47a",
    name: "39-5e-2a",
    identity: {
      mac: "00:00:00",
    },
    info: {
      id: "arch",
      pretty_name: "Linux",
      version: "",
    },
    public_key: "xxxxxxxx",
    tenant_id: "00000000",
    last_seen: "2020-05-20T18:58:53.276Z",
    online: true,
    namespace: "user",
  };

  const store = createStore({
    state: {
      device,
    },
    getters: {
      "devices/getFirstPending": (state) => state.device,
    },
    actions: {
      "devices/setFirstPending": vi.fn(),
      "snackbar/showSnackbarErrorLoading": vi.fn(),
    },
  });

  beforeEach(async () => {
    wrapper = mount(WelcomeThirdScreen, {
      global: {
        plugins: [[store, key], routes, vuetify],
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
    expect(wrapper.vm.getPendingDevice).toEqual(device);
  });
  //////
  // HTML validation
  //////

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="deviceName-field"]').text()).toEqual(device.name);
    expect(wrapper.find('[data-test="devicePrettyName-field"]').text()).toEqual(device.info.pretty_name);
  });
});
