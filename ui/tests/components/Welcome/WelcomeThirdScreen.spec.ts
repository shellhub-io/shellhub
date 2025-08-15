import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import WelcomeThirdScreen from "@/components/Welcome/WelcomeThirdScreen.vue";
import { router } from "@/router";
import { devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { IDevice } from "@/interfaces/IDevice";

type WelcomeThirdScreenWrapper = VueWrapper<InstanceType<typeof WelcomeThirdScreen>>;

describe("Welcome Third Screen", () => {
  let wrapper: WelcomeThirdScreenWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

  const devices = [
    {
      uid: "a582b47a42d",
      name: "39-5e-2a",
      identity: {
        mac: "00:00:00:00:00:00",
      },
      info: {
        id: "linuxmint",
        pretty_name: "Linux Mint 19.3",
        version: "",
      },
      public_key: "----- PUBLIC KEY -----",
      tenant_id: "fake-tenant-data",
      last_seen: "2020-05-20T18:58:53.276Z",
      online: false,
      namespace: "user",
      status: "pending",
    },
  ];

  beforeEach(async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=1&status=pending").reply(200, devices);

    wrapper = mount(WelcomeThirdScreen, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
      props: {
        firstPendingDevice: devices[0] as IDevice,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the components", async () => {
    expect(wrapper.find('[data-test="welcome-third-screen-name"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-third-screen-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-third-screen-hostname"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="welcome-third-screen-os"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-pretty-name-field"]').exists()).toBe(true);
  });
});
