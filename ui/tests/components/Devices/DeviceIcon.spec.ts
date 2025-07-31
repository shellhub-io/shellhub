import { setActivePinia, createPinia } from "pinia";
import { mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import { store, key } from "@/store";
import DeviceIcon from "@/components/Devices/DeviceIcon.vue";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("Device Icon", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceIcon>>;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(DeviceIcon, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      props: {
        icon: "",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the component", () => {
    expect(wrapper.find('[data-test="type-icon"]').exists()).toBe(true);
  });

  it("Renders the default icon when no icon prop is provided", () => {
    expect(wrapper.find('[data-test="type-icon"]').classes()).toContain("fl-tux");
  });

  it("Renders a specific icon based on the icon prop", async () => {
    const iconProp = "ubuntu";
    const expectedIconClass = "fl-ubuntu";

    await wrapper.setProps({ icon: iconProp });

    expect(wrapper.find('[data-test="type-icon"]').classes()).toContain(expectedIconClass);
  });

  it("Renders the default icon when an unknown icon prop is provided", async () => {
    const unknownIconProp = "unknown-icon";
    const defaultIconClass = "fl-tux";

    await wrapper.setProps({ icon: unknownIconProp });

    expect(wrapper.find('[data-test="type-icon"]').classes()).toContain(defaultIconClass);
  });
});
