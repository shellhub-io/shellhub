import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useDevicesStore from "@admin/store/modules/devices";
import useSnackbarStore from "@admin/store/modules/snackbar";
import routes from "../../../../src/router";
import DeviceDetails from "../../../../src/views/DeviceDetails.vue";

type DeviceDetailsWrapper = VueWrapper<InstanceType<typeof DeviceDetails>>;

const deviceDetail = {
  uid: "cb1533e2e683aec21aee89b24ac4604b1a1955930362d33fb22e4e03fac52c75",
  name: "08-97-98-68-7a-97",
  identity: { mac: "08:97:98:68:7a:97" },
  info: {
    id: "ubuntu",
    pretty_name: "Ubuntu 20.04.4 LTS",
    version: "latest",
    arch: "amd64",
    platform: "docker",
  },
  public_key: "---BEGIN RSA KEY---",
  tenant_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  last_seen: "2022-06-06T18:51:53.813Z",
  online: true,
  namespace: "dev",
  status: "accepted",
  created_at: "2022-04-13T11:43:25.218Z",
  remoteAddr: "172.22.0.1",
  position: { latitude: 0, longitude: 0 },
  tags: ["dev"],
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

    const deviceStore = useDevicesStore();
    vi.spyOn(deviceStore, "getDevice", "get").mockReturnValue(deviceDetail);
    deviceStore.get = vi.fn().mockResolvedValue(deviceDetail);

    const snackbarStore = useSnackbarStore();
    snackbarStore.showSnackbarErrorAction = vi.fn();

    const vuetify = createVuetify();

    wrapper = mount(DeviceDetails, {
      global: {
        plugins: [pinia, vuetify, routes],
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
    expect(wrapper.find(`[data-test='${deviceDetail.tags}']`).text()).toContain(deviceDetail.tags[0]);
    expect(wrapper.find(`[data-test='${deviceDetail.namespace}']`).text()).toContain(deviceDetail.namespace);
    expect(wrapper.find(`[data-test='${deviceDetail.status}']`).text()).toContain(deviceDetail.status);
  });
});
