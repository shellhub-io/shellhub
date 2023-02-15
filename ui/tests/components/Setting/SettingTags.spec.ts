import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SettingTags from "../../../src/components/Setting/SettingTags.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("SettingTags", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(SettingTags, {
      global: {
        plugins: [routes, vuetify],
      },
      shallow: true,
    });
  });

  ///////
  // Component Rendering
  //////

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  ///////s
  // Data checking
  //////
  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });
});
