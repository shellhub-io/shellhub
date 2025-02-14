import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import DeviceIcon from "../../../../../src/components/Device/DeviceIcon.vue";

describe("Device Icon", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(DeviceIcon, {
      props: {
        icon: "ubuntu",
      },
      global: {
        plugins: [vuetify],
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
