import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, afterEach } from "vitest";
import DeviceListChooser from "@/components/Devices/DeviceListChooser.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

type DeviceListChooserWrapper = VueWrapper<InstanceType<typeof DeviceListChooser>>;

describe("Device Chooser List", () => {
  let wrapper: DeviceListChooserWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(DeviceListChooser, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
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

  it("renders the component data table", async () => {
    expect(wrapper.findComponent('[data-test="devices-dataTable"]').exists()).toBe(true);
  });
});
