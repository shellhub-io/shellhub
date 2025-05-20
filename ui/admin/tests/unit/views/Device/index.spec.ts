import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useDevicesStore from "@admin/store/modules/devices";
import { SnackbarPlugin } from "@/plugins/snackbar";
import routes from "../../../../src/router";
import Device from "../../../../src/views/Device.vue";

type DeviceWrapper = VueWrapper<InstanceType<typeof Device>>;

describe("Device", () => {
  let wrapper: DeviceWrapper;

  beforeEach(() => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const devicesStore = useDevicesStore();

    vi.spyOn(devicesStore, "getPerPage", "get").mockReturnValue(10);
    vi.spyOn(devicesStore, "getPage", "get").mockReturnValue(1);
    vi.spyOn(devicesStore, "getNumberDevices", "get").mockReturnValue(1);

    devicesStore.search = vi.fn();
    devicesStore.fetch = vi.fn();

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
