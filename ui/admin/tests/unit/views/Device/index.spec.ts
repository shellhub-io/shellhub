import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useDevicesStore from "@admin/store/modules/devices";
import routes from "@admin/router";
import Device from "@admin/views/Device.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DeviceWrapper = VueWrapper<InstanceType<typeof Device>>;

const mockDevices = [
  {
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
    tenant_id: "00000000-0000-4000-0000-000000000000",
    last_seen: "2022-06-06T18:51:53.813Z",
    online: true,
    namespace: "dev",
    status: "accepted",
    created_at: "2022-04-13T11:43:25.218Z",
    remote_addr: "172.22.0.1",
    position: { latitude: 0, longitude: 0 },
    tags: ["tag1"],
  },
];

describe("Device", () => {
  let wrapper: DeviceWrapper;

  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const devicesStore = useDevicesStore();

    devicesStore.fetchDeviceList = vi.fn().mockImplementation(() => {
      devicesStore.devices = mockDevices;
      devicesStore.deviceCount = 1;
    });

    const vuetify = createVuetify();

    wrapper = mount(Device, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with default data", () => {
    expect(wrapper.vm.filter).toBe("");
  });

  it("Must change the filter value when input change", async () => {
    expect(wrapper.vm.filter).toEqual("");
    const input = wrapper.find("input");
    await input.setValue("ShellHub");
    expect(wrapper.vm.filter).toEqual("ShellHub");
  });

  it("Should render all the components in the screen", () => {
    expect(wrapper.find("h1").text()).toContain("Devices");
    expect(wrapper.find("input").exists()).toBe(true);
    expect(wrapper.find("[data-test='devices-list']").exists()).toBe(true);
  });
});
