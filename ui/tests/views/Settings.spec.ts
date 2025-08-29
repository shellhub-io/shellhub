import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import Settings from "@/views/Settings.vue";
import { router } from "@/router";

type SettingsWrapper = VueWrapper<InstanceType<typeof Settings>>;

describe("Settings View", () => {
  let wrapper: SettingsWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(Settings, {
      global: {
        plugins: [vuetify, router],
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
});
