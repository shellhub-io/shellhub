import { createStore, useStore as vuexUseStore } from "vuex";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import routes from "../../../../../src/router";
import UserDelete from "../../../../../src/components/User/UserDelete.vue";
import { key } from "../../../../../src/store";

const store = createStore({
  state: {},
  getters: {},
  actions: {
    "users/remove": () => vi.fn(),
    "snackbar/showSnackbarSuccessAction": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("User Delete", () => {
  let wrapper: VueWrapper<any>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(UserDelete, {
      props: {
        id: "6256d9e3ea6f26bc595130fa",
        redirect: false,
      },
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

  it("Receive data in props", () => {
    expect(wrapper.vm.id).toEqual("6256d9e3ea6f26bc595130fa");
    expect(wrapper.vm.redirect).toEqual(false);
  });

  it("Compare data with default value", () => {
    expect(wrapper.vm.dialog).toEqual(false);
  });
});
