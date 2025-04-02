import { createStore } from "vuex";
import { createVuetify } from "vuetify";
import { shallowMount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Login from "../../../../src/views/Login.vue";
import SnackbarComponent from "../../../../src/components/Snackbar/Snackbar.vue";
import { key } from "../../../../src/store";
import routes from "../../../../src/router";

type LoginWrapper = VueWrapper<InstanceType<typeof Login>>;

const store = createStore({
  state: {},
  getters: {
    "layout/getLayout": () => "simpleLayout",
    "auth/isLoggedIn": () => false,
    "license/get": () => undefined,
  },
  actions: {
    "auth/login": vi.fn(),
    "layout/setLayout": vi.fn(),
    "snackbar/showSnackbarErrorDefault": vi.fn(),
    "snackbar/showSnackbarErrorAction": vi.fn(),
  },
});

describe("Login", () => {
  let wrapper: LoginWrapper;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = shallowMount(Login, {
      global: {
        plugins: [[store, key], vuetify, routes],
        components: { SnackbarComponent },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });
  it("Compare data with default value", () => {
    expect(wrapper.vm.username).toEqual("");
    expect(wrapper.vm.password).toEqual("");
    expect(wrapper.vm.usernameError).toEqual(undefined);
    expect(wrapper.vm.passwordError).toEqual(undefined);
  });
});
