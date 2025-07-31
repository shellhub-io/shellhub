import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import { nextTick } from "vue";
import { store, key } from "@/store";
import DeviceAcceptWarning from "@/components/Devices/DeviceAcceptWarning.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";

describe("Device Accept Warning", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceAcceptWarning>>;
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const vuetify = createVuetify();

  beforeEach(async () => {
    store.commit("users/updateDeviceDuplicationError", true);
    authStore.role = "owner";
    wrapper = mount(DeviceAcceptWarning, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component", async () => {
    const dialog = new DOMWrapper(document.body);
    expect(dialog.find('[data-test="device-accept-warning-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="card-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="card-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="card-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });

  it("Closes the dialog when the close button is clicked", async () => {
    await nextTick();

    const closeButton = wrapper.findComponent('[data-test="close-btn"]');
    await closeButton.trigger("click");

    expect(wrapper.find('[data-test="device-accept-warning-dialog"]').exists()).toBe(false);
  });

  it("Dialog is shown based on user authorization", async () => {
    authStore.role = "administrator";
    await nextTick();

    expect(wrapper.find('[data-test="device-accept-warning-dialog"]').exists()).toBe(false);

    authStore.role = "observer";
    await nextTick();

    expect(wrapper.find('[data-test="device-accept-warning-dialog"]').exists()).toBe(false);
  });
});
