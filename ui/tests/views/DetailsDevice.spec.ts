import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import DetailsDevice from "@/views/DetailsDevice.vue";
import { devicesApi } from "@/api/http";
import { routes } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useDevicesStore from "@/store/modules/devices";
import { IDevice } from "@/interfaces/IDevice";

type DetailsDeviceWrapper = VueWrapper<InstanceType<typeof DetailsDevice>>;

describe("Details Device", () => {
  let wrapper: DetailsDeviceWrapper;
  setActivePinia(createPinia());
  const devicesStore = useDevicesStore();
  const vuetify = createVuetify();

  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

  const device = {
    uid: "a582b47a42d",
    name: "39-5e-2a",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
      arch: "x86_64",
      platform: "linux",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T18:58:53.276Z",
    created_at: "2020-05-20T18:00:00.000Z",
    online: false,
    namespace: "user",
    status: "accepted",
    remote_addr: "127.0.0.1",
    position: { latitude: 0, longitude: 0 },
    tags: [
      {
        tenant_id: "fake-tenant-data",
        name: "test-tag",
        created_at: "",
        updated_at: "",
      },
    ],
  };

  beforeEach(async () => {
    const router = createRouter({
      history: createWebHistory(),
      routes,
    });
    router.push("/devices/123456");
    await router.isReady();

    mockDevicesApi.onGet("http://localhost:3000/api/devices/resolve?uid=123456")
      .reply(200, device);
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted")
      .reply(200, [device]);

    devicesStore.device = device;

    wrapper = mount(DetailsDevice, {
      global: {
        plugins: [vuetify, [router], SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="device-uid-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-mac-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-pretty-name-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-version-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-tags-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="device-last-seen-field"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="connect-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="tunnel-list"]').exists()).toBe(false);
  });

  it("Renders the component when deviceIsEmpty is true", async () => {
    devicesStore.device = {} as IDevice;
    await nextTick();
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component when device status is not accepted", async () => {
    // Set device status to 'pending'
    devicesStore.device = { ...device, status: "pending" };
    await nextTick();
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component when device is offline", async () => {
    // Set device online status to false
    devicesStore.device = { ...device, online: false };
    await nextTick();
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component when device has no tags", async () => {
    // Set device tags to empty array
    devicesStore.device = { ...device, tags: [] };
    await nextTick();
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component when device has no last seen date", async () => {
    // Set device last_seen to empty string
    devicesStore.device = { ...device, last_seen: "" };
    await nextTick();
    expect(wrapper.html()).toMatchSnapshot();
  });
});
