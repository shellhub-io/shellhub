import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import { SnackbarPlugin } from "@/plugins/snackbar";
import DeviceIcon from "../../../../../src/components/Device/DeviceIcon.vue";

type DeviceIconWrapper = VueWrapper<InstanceType<typeof DeviceIcon>>;

describe("Device Icon", () => {
  let wrapper: DeviceIconWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(DeviceIcon, {
      props: {
        icon: "ubuntu",
      },
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the icon", () => {
    expect(wrapper.find("i").classes()).toContain("fl-ubuntu");
    expect(wrapper.find("i").exists()).toBeTruthy();
  });
});
