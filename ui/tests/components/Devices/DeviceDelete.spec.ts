import { setActivePinia, createPinia } from "pinia";
import { DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
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
        plugins: [vuetify, router, SnackbarPlugin],
      },
      props: {
        uid: "a582b47a42d",
        variant: "device",
        hasAuthorization: true,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the list item with correct elements", () => {
    expect(wrapper.find('[data-test="device-delete-item"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="remove-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="remove-title"]').exists()).toBe(true);
  });

  it("Opens MessageDialog when item is clicked", async () => {
    await wrapper.find('[data-test="device-delete-item"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="delete-device-dialog"]').exists()).toBe(true);
  });

  it("Shows MessageDialog with correct props", async () => {
    await wrapper.find('[data-test="device-delete-item"]').trigger("click");
    const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
    expect(messageDialog.exists()).toBe(true);
    expect(messageDialog.props("title")).toBe("Are you sure?");
    expect(messageDialog.props("description")).toBe("You are about to remove this device. After confirming this action cannot be redone.");
    expect(messageDialog.props("icon")).toBe("mdi-alert");
    expect(messageDialog.props("iconColor")).toBe("error");
    expect(messageDialog.props("confirmText")).toBe("Remove");
    expect(messageDialog.props("confirmColor")).toBe("error");
    expect(messageDialog.props("cancelText")).toBe("Close");
  });

  it("Shows dialog buttons with correct data-test attributes", async () => {
    await wrapper.find('[data-test="device-delete-item"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="confirm-btn"]').exists()).toBe(true);
  });

  it("Closes dialog when cancel is emitted", async () => {
    await wrapper.find('[data-test="device-delete-item"]').trigger("click");
    const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
    expect(messageDialog.props("modelValue")).toBe(true);

    await messageDialog.vm.$emit("cancel");
    expect(messageDialog.props("modelValue")).toBe(false);
  });

  it("Calls removeDevice when confirm is emitted", async () => {
    const storeSpy = vi.spyOn(devicesStore, "removeDevice").mockResolvedValue();
    mockDevicesApi.onDelete("http://localhost:3000/api/devices/a582b47a42d").reply(200);

    await wrapper.find('[data-test="device-delete-item"]').trigger("click");
    const messageDialog = wrapper.findComponent({ name: "MessageDialog" });

    await messageDialog.vm.$emit("confirm");
    expect(storeSpy).toHaveBeenCalledWith("a582b47a42d");
  });

  it("Disables list item when hasAuthorization is false", async () => {
    await wrapper.setProps({ hasAuthorization: false });
    const listItem = wrapper.find('[data-test="device-delete-item"]');
    expect(listItem.classes()).toContain("v-list-item--disabled");
  });

  it("Exposes removeDevice method", () => {
    expect(wrapper.vm.removeDevice).toBeDefined();
    expect(typeof wrapper.vm.removeDevice).toBe("function");
  });
});
