import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import DeviceListChooser from "../../../src/components/Devices/DeviceListChooser.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const numberDevices = 2;
const devicesSelected: any = [];
const disableShowSelect = true;
const action = "suggestedDevices";

const pagination = {
  groupBy: [],
  groupDesc: [],
  itemsPerPage: 10,
  multiSort: false,
  mustSort: false,
  page: 1,
  sortBy: [],
  sortDesc: [false],
};

const headers = [
  {
    text: "Hostname",
    value: "hostname",
  },
  {
    text: "Operating System",
    value: "info.pretty_name",
  },
  {
    text: "SSHID",
    value: "namespace",
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
  },
  actions: {
    "devices/fetch": vi.fn(),
    "devices/setDevicesSelected": vi.fn(),
    "devices/getDevicesMostUsed": vi.fn(),
    "devices/setDevicesForUserToChoose": vi.fn(),
    "devices/getNumberForUserToChoose": vi.fn(),
    "devices/resetListDevices": vi.fn(),
    "snackbar/showSnackbarCopy": vi.fn(),
    "snackbar/showSnackbarErrorAssociation": vi.fn(),
    "snackbar/showSnackbarErrorLoading": vi.fn(),
  },
});

describe("Suggested devices", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(DeviceListChooser, {
      global: {
        plugins: [[store, key], routes, vuetify],
        stubs: ["router-link", "router-view"],
      },
      props: {
        action,
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

  ///////
  // Data and Props checking //
  //////

  it("Has the correct data", () => {
    const defaultData = wrapper.vm;
    expect(defaultData.headers).toStrictEqual(headers);
    expect(defaultData.itemsPerPage).toStrictEqual(pagination.itemsPerPage);
    expect(defaultData.page).toStrictEqual(pagination.page);
    expect(defaultData.selected).toStrictEqual(devicesSelected);
    expect(defaultData.action).toStrictEqual(action);
  });

  it("Has the correct props", () => {
    expect(wrapper.vm.action).toStrictEqual(action);
  });

  it("Has the correct computed", () => {
    expect(wrapper.vm.devices).toStrictEqual(devices);
    expect(wrapper.vm.numberDevices).toStrictEqual(numberDevices);
    expect(wrapper.vm.selected).toStrictEqual(devicesSelected);
    expect(wrapper.vm.disableShowSelect).toStrictEqual(disableShowSelect);
  });

  //////
  // HTML validation
  //////

  it("Renders the template with components", async () => {
    expect(wrapper.find('[data-test="devices-list-chooser"]').exists()).toEqual(
      true
    );
  });
});
