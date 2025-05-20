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
  { text: "Last Seen", value: "last_seen", sortable: true, align: "center" },
  { text: "Status", value: "status", sortable: true },
  { text: "Actions", value: "actions" },
];

const devices = [
  {
    created_at: "2020-05-20T19:58:53.276Z",
    identity: { mac: "00:00:00:00:00:00" },
    info: {
      arch: "x86_64",
      id: "linuxmint",
      platform: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "18.4.2",
    },
    last_seen: "2020-05-20T19:58:53.276Z",
    name: "tests",
    namespace: "dev",
    online: true,
    position: {
      latitude: 12,
      longitude: 12,
    },
    public_key: "xxxxxxxxxxxxxxxx",
    remoteAddr: "127.0.0.1",
    status: "accepted",
    tags: ["xxxx", "yyyyy"],
    tenant_id: "00000000",
    uid: "a582b47a42d",
  },
];

describe("Device List", () => {
  let wrapper: DeviceListWrapper;

  beforeEach(() => {
    setActivePinia(createPinia());
    const vuetify = createVuetify();

    const devicesStore = useDevicesStore();
    devicesStore.devices = devices;
    devicesStore.numberDevices = devices.length;

    vi.spyOn(devicesStore, "fetch").mockResolvedValue(false);

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
