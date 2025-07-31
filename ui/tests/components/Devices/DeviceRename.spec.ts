import { createPinia, setActivePinia } from "pinia";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, afterEach, vi } from "vitest";
import { store, key } from "@/store";
import DeviceRename from "@/components/Devices/DeviceRename.vue";
import { devicesApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DeviceRenameWrapper = VueWrapper<InstanceType<typeof DeviceRename>>;

describe("Device Rename", () => {
  let wrapper: DeviceRenameWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

  beforeEach(async () => {
    wrapper = mount(DeviceRename, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
      props: {
        uid: "a582b47a42d",
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("renders the component items", async () => {
    expect(wrapper.findComponent('[data-test="rename-icon"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="rename-title"]').exists()).toBe(true);

    wrapper.vm.showDialog = true;

    await flushPromises();
    expect(wrapper.findComponent('[data-test="device-rename-card"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="text-title"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="rename-field"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="close-btn"]').exists()).toBe(true);
    expect(wrapper.findComponent('[data-test="rename-btn"]').exists()).toBe(true);
  });

  it("renames successfully a device", async () => {
    wrapper.vm.showDialog = true;

    await flushPromises();

    mockDevicesApi.onPut("http://localhost:3000/api/devices/a582b47a42d").reply(200);

    const deviceSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="rename-field"]').setValue("renamed-device");
    await wrapper.findComponent('[data-test="rename-btn"]').trigger("click");

    await flushPromises();

    expect(deviceSpy).toHaveBeenCalledWith("devices/rename", {
      uid: "a582b47a42d",
      name: { name: "renamed-device" },
    });
  });

  it("fails to rename a device", async () => {
    wrapper.vm.showDialog = true;

    await flushPromises();

    mockDevicesApi.onPut("http://localhost:3000/api/devices/a582b47a42d").reply(400);

    const deviceSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="rename-field"]').setValue("badly renamed device");
    await wrapper.findComponent('[data-test="rename-btn"]').trigger("click");

    await flushPromises();

    expect(deviceSpy).toHaveBeenCalledWith("devices/rename", {
      uid: "a582b47a42d",
      name: { name: "badly renamed device" },
    });
  });
});
