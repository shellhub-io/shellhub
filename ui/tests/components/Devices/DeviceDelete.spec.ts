import { setActivePinia, createPinia } from "pinia";
import { DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import { store, key } from "@/store";
import DeviceDelete from "@/components/Devices/DeviceDelete.vue";
import { router } from "@/router";
import { devicesApi, tagsApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useDevicesStore from "@/store/modules/devices";

describe("Device Delete", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceDelete>>;
  setActivePinia(createPinia());
  const devicesStore = useDevicesStore();
  const vuetify = createVuetify();

  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());

  beforeEach(async () => {
    mockTagsApi.onGet("http://localhost:3000/api/tags").reply(200, []);

    wrapper = mount(DeviceDelete, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      props: {
        uid: "a582b47a42d",
        variant: "device",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component data table", async () => {
    expect(wrapper.find('[data-test="device-delete-item"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="remove-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="remove-title"]').exists()).toBe(true);
    await wrapper.find('[data-test="device-delete-item"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="delete-device-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="device-delete-card"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="confirm-btn"]').exists()).toBe(true);
  });

  it("Closes the delete dialog when close button is clicked", async () => {
    const dialog = new DOMWrapper(document.body);
    await wrapper.findComponent('[data-test="device-delete-item"]').trigger("click");
    expect(dialog.find('[data-test="delete-device-dialog"]').exists()).toBe(true);

    await wrapper.findComponent('[data-test="close-btn"]').trigger("click");
    expect(wrapper.find('[data-test="delete-dialog"]').exists()).toBe(false);
  });

  it("Calls remove method when confirm button is clicked", async () => {
    await wrapper.setProps({ uid: "test_device" });
    const storeSpy = vi.spyOn(devicesStore, "removeDevice");

    mockDevicesApi.onDelete("http://localhost:3000/api/devices/test_device").reply(200);

    await wrapper.findComponent('[data-test="device-delete-item"]').trigger("click");
    await wrapper.findComponent('[data-test="confirm-btn"]').trigger("click");

    expect(storeSpy).toHaveBeenCalledWith("test_device");
  });
});
