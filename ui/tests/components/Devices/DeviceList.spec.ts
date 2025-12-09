import { createPinia, setActivePinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach } from "vitest";
import DeviceList from "@/components/Devices/DeviceList.vue";
import { router } from "@/router";
import { devicesApi, tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DeviceListWrapper = VueWrapper<InstanceType<typeof DeviceList>>;

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
    tags: [{
      tenant_id: "fake-tenant-data",
      name: "test-tag",
      created_at: "",
      updated_at: "",
    }],
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
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());

  localStorage.setItem("tenant", "fake-tenant-data");
  mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, devices);
  mockTagsApi
    .onGet("http://localhost:3000/api/tags?filter=&page=1&per_page=10")
    .reply(200, []);

  beforeEach(() => {
    wrapper = mount(DeviceList, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component HTML", () => {
    expect(wrapper.findComponent('[data-test="device-table"]').exists()).toBe(true);
  });
});
