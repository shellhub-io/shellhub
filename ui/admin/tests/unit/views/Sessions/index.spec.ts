import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import Sessions from "../../../../src/views/Sessions.vue";

describe("Sessions", () => {
  const store = createStore({
    state: {},
    getters: {
      "sessions/numberSessions": () => 1,
    },
    actions: {
      "sessions/fetch": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(Sessions, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Should render all the components in the screen", () => {
    expect(wrapper.find("h1").text()).toContain("Sessions");
    expect(wrapper.find("[data-test='session-list']").exists()).toBe(true);
  });
});
