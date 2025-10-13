import { setActivePinia, createPinia } from "pinia";
import { mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it } from "vitest";
import WelcomeThirdScreen from "@/components/Welcome/WelcomeThirdScreen.vue";
import { devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { IDevice } from "@/interfaces/IDevice";

describe("Welcome Third Screen", () => {
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

  mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=1&status=pending").reply(200, devices);

  const wrapper = mount(WelcomeThirdScreen, {
    global: { plugins: [vuetify, SnackbarPlugin] },
    props: { firstPendingDevice: devices[0] as IDevice },
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Shows fallback message when no device is detected", () => {
    const wrapperNoDevice = mount(WelcomeThirdScreen, {
      global: { plugins: [vuetify, SnackbarPlugin] },
      props: { firstPendingDevice: undefined },
    });

    expect(wrapperNoDevice.find("[data-test='no-device-heading']").text()).toBe("No Device Detected Yet");
    expect(wrapperNoDevice.find("[data-test='no-device-text']").text()).toContain("Please run the installation command");
  });
});
