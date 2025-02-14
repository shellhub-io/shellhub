import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import DeviceList from "../../../../../src/components/Device/DeviceList.vue";
import { key } from "../../../../../src/store";
import routes from "../../../../../src/router";

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
      prettyName: "Linux Mint 19.3",
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
    remote_addr: "127.0.0.1",
    status: "accepted",
    tags: ["xxxx", "yyyyy"],
    tenant_id: "00000000",
    uid: "a582b47a42d",
  },
];

const store = createStore({
  state: {
    devices,
  },
  getters: {
    "devices/list": (state) => state.devices,
    "devices/numberDevices": (state) => state.devices.length,
    "auth/isLoggedIn": () => false,
  },
  actions: {
    "modals/showAddDevice": vi.fn(),
    "devices/fetch": vi.fn(),
    "devices/rename": vi.fn(),
    "devices/resetListDevices": vi.fn(),
    "stats/get": vi.fn(),
  },
});

describe("Device List", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(DeviceList, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
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
    const devices = await wrapper.vm.getListDevices;
    expect(devices).toEqual(devices);
  });
});
