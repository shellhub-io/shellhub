import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import { createStore } from "vuex";
import WelcomeFirstScreen from "../../../src/components/Welcome/WelcomeFirstScreen.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("WelcomeFirstScreen", () => {
  let wrapper: VueWrapper<InstanceType<typeof WelcomeFirstScreen>>;
  const vuetify = createVuetify();

  const user = "ShellHub";
  const name = "ShellHub";

  const store = createStore({
    state: {
      user,
      name,
    },
    getters: {
      "auth/currentUser": (state) => state.user,
      "auth/currentName": (state) => state.name,
    },
    actions: {},
  });

  beforeEach(() => {
    wrapper = mount(WelcomeFirstScreen, {
      global: {
        plugins: [[store, key], routes, vuetify],
      },
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

  ///////
  // Data checking
  //////
  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });
  it("Process data in the computed", () => {
    expect(wrapper.vm.name).toEqual(user);
  });
});
