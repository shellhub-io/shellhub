import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import DevicePendingList from "../../../src/components/Devices/DevicePendingList.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const numberDevices = 2;
const devicesSelected: any = [];
const action = "suggestedDevices";

const headers = [
  {
    text: "Hostname",
    value: "name",
    sortable: true,
  },
  {
    text: "Operating System",
    value: "operating_system",
  },
  {
    text: "Request Time",
    value: "request_time",
  },
  {
    text: "Actions",
    value: "actions",
  },
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
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "00000000",
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
    tenant_id: "00000001",
    last_seen: "2020-05-20T19:58:53.276Z",
    online: true,
    namespace: "user",
    status: "accepted",
  },
];

const store = createStore({
  state: {
    devices,
    numberDevices,
    devicesSelected,
  },
  getters: {
    "devices/getDevicesForUserToChoose": (state) => state.devices,
    "devices/getNumberForUserToChoose": (state) => state.numberDevices,
    "devices/getDevicesSelected": (state) => state.devicesSelected,
    "devices/getNumberDevices": (state) => state.numberDevices,
    "devices/list": (state) => state.devices,
  },
  actions: {
    "devices/fetch": () => {},
    "devices/setDevicesSelected": () => {},
    "devices/getDevicesMostUsed": () => {},
    "devices/setDevicesForUserToChoose": () => {},
    "devices/getNumberForUserToChoose": () => {},
    "devices/resetListDevices": () => {},
    "snackbar/showSnackbarCopy": () => {},
    "snackbar/showSnackbarErrorAssociation": () => {},
    "snackbar/showSnackbarErrorLoading": () => {},
  },
});

describe("Devices Pending List", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();
    wrapper = mount(DevicePendingList, {
      global: {
        plugins: [[store, key], vuetify, routes],
        stubs: ["router-link", "router-view"],
      },
      props: {
        action,
      },
      data() {
        return {
          headers,
          devices,
        };
      },
      shallow: true,
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

  it("Compare data with default value", () => {
    expect(wrapper.vm.headers).toEqual(headers);
    expect(wrapper.vm.devices).toEqual(devices);
    expect(wrapper.vm.numberDevices).toEqual(numberDevices);
  });

  it("Process data in methods", () => {
    const address = `${devices[0].namespace}.${devices[0].name}@`;
    expect(wrapper.vm.sshidAddress(devices[0])).toEqual(address);
  });

  //////
  // HTML validation
  //////

  it("Renders the template with components", async () => {
    expect(wrapper.find('[data-test="devices-list"]').exists()).toEqual(
      true
    );
  });
  it('Renders the template with data', () => {
    const dataTable = wrapper.find('[data-test="devices-list"]');
    const dataTableAttr = dataTable.attributes();
    expect(dataTable.exists()).toEqual(true);
    expect(+dataTableAttr.totalcount).toBe(numberDevices);
    expect(+dataTableAttr.actualpage).toBe(1);

  });
});
