import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useDevicesStore from "@admin/store/modules/devices";
import routes from "@admin/router";
import Device from "@admin/views/Device.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DeviceWrapper = VueWrapper<InstanceType<typeof Device>>;

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
      arch: "x86_64",
      platform: "linux",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T18:58:53.276Z",
    created_at: "2020-05-20T18:00:00.000Z",
    status_updated_at: "2020-05-20T18:58:53.276Z",
    online: false,
    namespace: "user",
    status: "accepted" as const,
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
      arch: "x86_64",
      platform: "linux",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T19:58:53.276Z",
    created_at: "2020-05-20T18:00:00.000Z",
    status_updated_at: "2020-05-20T18:58:53.276Z",
    online: true,
    namespace: "user",
    status: "accepted" as const,
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
  },
];

describe("Device", () => {
  let wrapper: DeviceWrapper;

  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const devicesStore = useDevicesStore();

    devicesStore.fetchDeviceList = vi.fn().mockImplementation(() => {
      devicesStore.devices = devices;
      devicesStore.deviceCount = 1;
    });

    const vuetify = createVuetify();

    wrapper = mount(Device, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with default data", () => {
    expect(wrapper.vm.filter).toBe("");
  });

  it("Must change the filter value when input change", async () => {
    expect(wrapper.vm.filter).toEqual("");
    const input = wrapper.find("input");
    await input.setValue("ShellHub");
    expect(wrapper.vm.filter).toEqual("ShellHub");
  });

  it("Should render all the components in the screen", () => {
    expect(wrapper.find("h1").text()).toContain("Devices");
    expect(wrapper.find("input").exists()).toBe(true);
    expect(wrapper.find("[data-test='devices-list']").exists()).toBe(true);
  });
});
