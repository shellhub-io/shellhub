import { createPinia, setActivePinia } from "pinia";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { createStore } from "vuex";
import { key } from "@/store";
import DeviceTable from "@/components/Tables/DeviceTable.vue";
import { router } from "@/router";
import { devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { IDevice } from "@/interfaces/IDevice";

type DeviceTableWrapper = VueWrapper<InstanceType<typeof DeviceTable>>;

describe("Device Table", () => {
  let wrapper: DeviceTableWrapper;
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
      status: "accepted",
      tags: ["test"],
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
      tags: ["test"],
    },
  ];

  const mockStore = createStore({
    state: {
      totalCount: 3,
      devices: [
        { name: "Device1", operating_system: "OS1", sshid: "ssh1", tags: "tag1" },
        { name: "Device2", operating_system: "OS2", sshid: "ssh2", tags: "tag2" },
        { name: "Device3", operating_system: "OS3", sshid: "ssh3", tags: "tag3" },
      ],
    },
    getters: {
      totalCount: (state) => state.totalCount,
      devices: (state) => state.devices,
    },
  });

  const mockStoreMethods = {
    fetchDevices: vi.fn(),
    setSort: vi.fn(),
    getFilter: vi.fn(),
    getList: () => devices as IDevice[],
    getSortStatusField: vi.fn(),
    getSortStatusString: vi.fn(),
    getNumber: () => mockStore.state.totalCount,
  };

  beforeEach(async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices?filter=&page=1&per_page=10&status=accepted").reply(200, devices);

    wrapper = shallowMount(DeviceTable, {
      global: {
        plugins: [[mockStore, key], vuetify, router, SnackbarPlugin],
      },
      props: {
        storeMethods: mockStoreMethods,
        status: "accepted",
        header: "primary",
        variant: "device",
      },
    });
  });

  it("Is a Vue instance", async () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component HTML", async () => {
    expect(wrapper.findComponent('[data-test="items-list"]').exists()).toBe(true);
  });
});
