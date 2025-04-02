import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";
import SimpleLayout from "../../../../src/layouts/SimpleLayout.vue";

type SimpleLayoutWrapper = VueWrapper<InstanceType<typeof SimpleLayout>>;

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
  let wrapper: SimpleLayoutWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = shallowMount(SimpleLayout, {
      global: {
        plugins: [[store, key], vuetify, routes],
      },
    }) as unknown as SimpleLayoutWrapper;
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
});
