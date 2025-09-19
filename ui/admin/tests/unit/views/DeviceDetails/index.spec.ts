import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useDevicesStore from "@admin/store/modules/devices";
import routes from "@admin/router";
import DeviceDetails from "@admin/views/DeviceDetails.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DeviceDetailsWrapper = VueWrapper<InstanceType<typeof DeviceDetails>>;

const deviceDetail = {
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
  online: true,
  namespace: "user",
  status: "accepted",
  created_at: "2020-05-01T00:00:00.000Z",
  remote_addr: "127.0.0.1",
  position: {
    longitude: 0,
    latitude: 0,
  },
  tags: [
    {
      tenant_id: "fake-tenant-data",
      name: "test-tag",
      created_at: "",
      updated_at: "",
    },
  ],
};

const mockRoute = {
  params: {
    id: deviceDetail.uid,
  },
};

describe("Device Details", () => {
  let wrapper: DeviceDetailsWrapper;

  beforeEach(async () => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const devicesStore = useDevicesStore();
    devicesStore.fetchDeviceById = vi.fn().mockResolvedValue(deviceDetail);

    const vuetify = createVuetify();

    wrapper = mount(DeviceDetails, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
        mocks: {
          $route: mockRoute,
        },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Has the correct data", async () => {
    expect(wrapper.vm.device).toEqual(deviceDetail);
  });

  it("Render the correct title", () => {
    expect(wrapper.find("h1").text()).toBe("Device Details");
  });

  it("Should render the props of the device on the screen", () => {
    expect(wrapper.find(`[data-test='${deviceDetail.uid}']`).text()).toContain(deviceDetail.uid);
    expect(wrapper.find(`[data-test='${deviceDetail.name}']`).text()).toContain(deviceDetail.name);
    expect(wrapper.find(`[data-test='${deviceDetail.identity.mac}']`).text()).toContain(deviceDetail.identity.mac);
    expect(wrapper.find(`[data-test='${deviceDetail.info.id}']`).text()).toContain(deviceDetail.info.id);
    expect(wrapper.find(`[data-test='${deviceDetail.tenant_id}']`).text()).toContain(deviceDetail.tenant_id);
    expect(wrapper.find(`[data-test='${deviceDetail.online}']`).text()).toContain(String(deviceDetail.online));
    expect(wrapper.find(`[data-test='${deviceDetail.tags}']`).text()).toContain(deviceDetail.tags[0].name);
    expect(wrapper.find(`[data-test='${deviceDetail.namespace}']`).text()).toContain(deviceDetail.namespace);
    expect(wrapper.find(`[data-test='${deviceDetail.status}']`).text()).toContain(deviceDetail.status);
  });
});
