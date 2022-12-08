import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import DeviceList from "../../../src/components/Devices/DeviceList.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

const numberDevicesGlobal = 2;

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
    text: "Online",
    value: "online",
    sortable: true,
  },
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
    text: "SSHID",
    value: "sshid",
  },
  {
    text: "Tags",
    value: "tags",
  },
  {
    text: "Actions",
    value: "actions",
  },
];

const devicesGlobal = [
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
    tags: ["device1", "device2"],
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
    tags: ["device1", "device2"],
  },
];

const tests = [
  {
    description: "List data when user has owner role",
    role: {
      type: "owner",
      permission: true,
    },
    variables: {
      devices: devicesGlobal,
      numberDevices: numberDevicesGlobal,
    },
    data: {
      headers,
      items: devicesGlobal,
      itemsPerPage: 10,
      numberDevices: numberDevicesGlobal,
      actualPage: 1,
      comboboxOptions: [10,20,50,100],
      nextPage: vi.fn(),
      previousPage: vi.fn(),
      loading: false,
    },
    computed: {
      getListDevices: devicesGlobal,
      getNumberDevices: numberDevicesGlobal,
      hasAuthorizationFormUpdate: true,
    },
  },
  {
    description: "List data when user has observer role",
    role: {
      type: "observer",
      permission: false,
    },
    variables: {
      devices: devicesGlobal,
      numberDevices: numberDevicesGlobal,
    },
    data: {
      headers,
      items: devicesGlobal,
      itemsPerPage: 10,
      numberDevices: numberDevicesGlobal,
      actualPage: 1,
      comboboxOptions: [10,20,50,100],
      nextPage: vi.fn(),
      previousPage: vi.fn(),
      loading: false,
    },
    computed: {
      getListDevices: devicesGlobal,
      getNumberDevices: numberDevicesGlobal,
      hasAuthorizationFormUpdate: false,
    },
  },
];

const store = (devices: any, numberDevices: any, currentRole: any) =>
  createStore({
    state: {
      devices,
      numberDevices,
      currentRole,
    },
    getters: {
      "devices/list": (state) => state.devices,
      "devices/getNumberDevices": (state) => state.numberDevices,
      "auth/role": (state) => state.currentRole,
    },
    actions: {
      "modals/showAddDevice": vi.fn(),
      "devices/fetch": vi.fn(),
      "devices/rename": vi.fn(),
      "tags/clearSelectedTags": vi.fn(),
      "devices/resetListDevices": vi.fn(),
      "tags/setSelected": vi.fn(),
      "devices/setFilter": vi.fn(),
      "stats/get": vi.fn(),
    },
  });

tests.forEach((test) => {
  describe(`${test.description}`, () => {
    let wrapper: VueWrapper<any>;

    beforeEach(() => {
      const vuetify = createVuetify();

      wrapper = mount(DeviceList, {
        global: {
          plugins: [
            [
              store(
                test.variables.devices,
                test.variables.numberDevices,
                test.role
              ),
              key,
            ],
            routes,
            vuetify,
          ],
          stubs: ["router-link", "router-view"],
        },
        shallow: true,
        data() {
          return test.data;
        },
        props: {
          totalCount: 2,
        },
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
    // Data checking
    //////

    it("Compare data with default value", () => {
      expect(wrapper.vm.devices).toEqual(devicesGlobal);
      expect(wrapper.vm.headers).toEqual(headers);
      expect(wrapper.vm.page).toEqual(pagination.page);
      expect(wrapper.vm.itemsPerPage).toEqual(pagination.itemsPerPage);
      expect(wrapper.vm.loading).toEqual(false);
    });
    it('Process data in methods', () => {
      devicesGlobal.forEach((device: any) => {
        const address = `${device.namespace}.${device.name}@`;
        expect(wrapper.vm.sshidAddress(device)).toEqual(address);
      });
    });

    //////
    // HTML validation
    //////

    it("Renders the template with data", () => {
      const dt = wrapper.find('[data-test="devices-list"]');
      const dataTableProps = dt.attributes();
      expect(dataTableProps.items).toContain(test.variables.devices);
    });
  });
});
