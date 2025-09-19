import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useDevicesStore from "@admin/store/modules/devices";
import { SnackbarPlugin } from "@/plugins/snackbar";
import DeviceList from "../../../../../src/components/Device/DeviceList.vue";
import routes from "../../../../../src/router";

type DeviceListWrapper = VueWrapper<InstanceType<typeof DeviceList>>;

const headers = [
  { text: "Online", value: "online", sortable: true },
  { text: "Hostname", value: "name", sortable: true },
  { text: "Info", value: "info", sortable: true },
  { text: "Namespace", value: "namespace", sortable: true },
  { text: "Tags", value: "tags" },
  { text: "Last Seen", value: "last_seen", sortable: true },
  { text: "Status", value: "status", sortable: true },
  { text: "Actions", value: "actions" },
];

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
    online: true,
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
  },
];

describe("Device List", () => {
  let wrapper: DeviceListWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());
    const vuetify = createVuetify();

    const devicesStore = useDevicesStore();
    devicesStore.devices = devices;
    devicesStore.deviceCount = devices.length;
    devicesStore.fetchDeviceList = vi.fn();

    wrapper = mount(DeviceList, {
      global: {
        plugins: [vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    const dt = wrapper.find("[data-test]");
    expect(dt.attributes()["data-test"]).toBe("devices-list");
    expect(wrapper.vm.headers).toEqual(headers);
    expect(wrapper.vm.devices).toEqual(devices);
    expect(wrapper.vm.loading).toEqual(false);
    expect(wrapper.vm.itemsPerPage).toEqual(10);
  });

  it("Renders data in the computed", async () => {
    expect(wrapper.vm.devices).toEqual(devices);
  });
});
