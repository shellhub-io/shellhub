import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import Devices from "@/views/Devices.vue";
import { devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { router } from "@/router";
import useDevicesStore from "@/store/modules/devices";
import { IDevice } from "@/interfaces/IDevice";

type DevicesWrapper = VueWrapper<InstanceType<typeof Devices>>;

describe("Devices View", () => {
  let wrapper: DevicesWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const devicesStore = useDevicesStore();
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

  const mockDevices = [
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
      status: "accepted",
    },
    {
      uid: "a582b47a42e",
      name: "39-5e-2b",
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
      last_seen: "2020-05-20T19:58:53.276Z",
      online: true,
      namespace: "user",
      status: "accepted",
    },
  ];

  beforeEach(async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10").reply(
      200,
      { data: mockDevicesApi, headers: { "x-total-count": 2 } },
    );

    devicesStore.devices = mockDevices as IDevice[];

    wrapper = mount(Devices, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", async () => {
    expect(wrapper.find('[data-test="search-text"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="device-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-header-component-group"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-table-component"]').exists()).toBe(false);
  });

  it("Shows the no items message when there are no public keys", async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10").reply(200, []);
    devicesStore.devices = [];
    expect(wrapper.find('[data-test="no-items-message-component"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="no-items-message-component"]').text()).toContain("Looks like you don't have any Devices");
  });
});
