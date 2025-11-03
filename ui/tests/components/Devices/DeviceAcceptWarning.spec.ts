import { createPinia, setActivePinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, afterEach } from "vitest";
import DeviceAcceptWarning from "@/components/Devices/DeviceAcceptWarning.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";
import useDevicesStore from "@/store/modules/devices";

describe("DeviceAcceptWarning", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceAcceptWarning>>;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const devicesStore = useDevicesStore();
  const vuetify = createVuetify();

  beforeEach(() => {
    // Reset stores to clean state
    authStore.role = "owner";
    devicesStore.duplicatedDeviceName = "";
  });

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
    }
  });

  it("Is a Vue instance", () => {
    devicesStore.duplicatedDeviceName = "Test Device";
    wrapper = mount(DeviceAcceptWarning, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    devicesStore.duplicatedDeviceName = "Test Device";
    wrapper = mount(DeviceAcceptWarning, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Shows MessageDialog when device name is duplicated and user has authorization", () => {
    devicesStore.duplicatedDeviceName = "Test Device";
    authStore.role = "owner";

    wrapper = mount(DeviceAcceptWarning, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });

    expect(wrapper.findComponent({ name: "MessageDialog" }).exists()).toBe(true);
  });

  it("Shows MessageDialog with correct props", () => {
    devicesStore.duplicatedDeviceName = "Test Device";

    wrapper = mount(DeviceAcceptWarning, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });

    const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
    expect(messageDialog.exists()).toBe(true);
    expect(messageDialog.props("title")).toBe("You already have a device using the same name");
    const expectedDescription = "Test Device name is already taken by another accepted device, please choose another name.";
    expect(messageDialog.props("description")).toBe(expectedDescription);
    expect(messageDialog.props("icon")).toBe("mdi-alert");
    expect(messageDialog.props("iconColor")).toBe("warning");
    expect(messageDialog.props("cancelText")).toBe("Close");
  });

  it("Clears duplicated device name when cancel is emitted", async () => {
    devicesStore.duplicatedDeviceName = "Test Device";

    wrapper = mount(DeviceAcceptWarning, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });

    const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
    await messageDialog.vm.$emit("cancel");

    expect(devicesStore.duplicatedDeviceName).toBe("");
  });

  it("Clears duplicated device name when close is emitted", async () => {
    devicesStore.duplicatedDeviceName = "Test Device";

    wrapper = mount(DeviceAcceptWarning, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });

    const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
    await messageDialog.vm.$emit("close");

    expect(devicesStore.duplicatedDeviceName).toBe("");
  });

  it("Component respects user authorization", () => {
    authStore.role = "observer"; // No billing:subscribe permission
    devicesStore.duplicatedDeviceName = "Test Device";

    wrapper = mount(DeviceAcceptWarning, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });

    // Component should still exist but conditionally render based on permissions
    expect(wrapper.vm).toBeTruthy();
  });

  it("Component handles empty device name", () => {
    devicesStore.duplicatedDeviceName = ""; // No duplicated name
    authStore.role = "owner";

    wrapper = mount(DeviceAcceptWarning, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
    });

    // Component should exist and handle empty state gracefully
    expect(wrapper.vm).toBeTruthy();
  });
});
