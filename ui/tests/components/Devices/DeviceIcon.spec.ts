import { setActivePinia, createPinia } from "pinia";
import { mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it } from "vitest";
import DeviceIcon from "@/components/Devices/DeviceIcon.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Device Icon", () => {
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  const wrapper = mount(DeviceIcon, {
    global: { plugins: [vuetify, router, SnackbarPlugin] },
    props: { icon: "" },
  });

  it("Renders the default icon when no icon prop is provided", () => {
    expect(wrapper.find('[data-test="device-icon"]').classes()).toContain("fl-tux");
  });

  it("Renders a specific icon based on the icon prop", async () => {
    const iconProp = "ubuntu";
    const expectedIconClass = "fl-ubuntu";

    await wrapper.setProps({ icon: iconProp });

    expect(wrapper.find('[data-test="device-icon"]').classes()).toContain(expectedIconClass);
  });

  it("Renders the default icon when an unknown icon prop is provided", async () => {
    const unknownIconProp = "unknown-icon";
    const defaultIconClass = "fl-tux";

    await wrapper.setProps({ icon: unknownIconProp });

    expect(wrapper.find('[data-test="device-icon"]').classes()).toContain(defaultIconClass);
  });
});
