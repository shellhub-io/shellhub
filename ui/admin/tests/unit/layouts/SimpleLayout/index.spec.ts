import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import SimpleLayout from "../../../../src/layouts/SimpleLayout.vue";

const layout = "simpleLayout";

const store = createStore({
  state: {
    layout,
  },
  getters: {
    "layout/getLayout": (state) => state.layout,
  },
  actions: {
    "layout/setLayout": vi.fn(),
    "auth/logout": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("SimpleLayout", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = shallowMount(SimpleLayout, {
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
});
