import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Login from "../../src/views/Login.vue";
import { createStore } from "vuex";
import { key } from "../../src/store";
import routes from "../../src/router";
import { envVariables } from "../../src/envVariables";

describe("Login is not cloud", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const statusDarkMode = true;

  const store = createStore({
    state: {
      statusDarkMode,
    },
    getters: {
      "layout/getStatusDarkMode": (state) => state.statusDarkMode,
    },
    actions: {
      "auth/logout": vi.fn(),
      "auth/login": vi.fn(),
      "auth/loginToken": vi.fn(),
      "notifications/fetch": vi.fn(),
      "layout/setLayout": vi.fn(),
      "snackbar/showSnackbarErrorIncorrect": vi.fn(),
      "snackbar/showSnackbarErrorDefault": vi.fn(),
    },
  });

  describe("Login screen", () => {
    beforeEach(async () => {
      wrapper = mount(Login, {
        global: {
          plugins: [[store, key], vuetify, routes],
        },
      });

      envVariables.isCloud = false;
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
    it("Compare data with default value", () => {
      expect(wrapper.vm.username).toEqual("");
      expect(wrapper.vm.password).toEqual("");
      expect(wrapper.vm.usernameError).toEqual(undefined);
      expect(wrapper.vm.passwordError).toEqual(undefined);
      expect(wrapper.vm.showPassword).toEqual(false);
      expect(wrapper.vm.showMessage).toEqual(false);
    });

    //////
    // HTML validation
    //////
    it("Renders the template with components", () => {
      // expect(wrapper.find('[data-test="accountCreated-component"]').exists()).toEqual(false);
    });

    it("Renders the template with components", () => {
      expect(
        wrapper.find('[data-test="accountCreated-component"]').exists()
      ).toEqual(false);
    });
    it("Renders the template with data", () => {
      expect(wrapper.find('[data-test="username-text"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="password-text"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="unknownReason-card"]').exists()).toBe(
        false
      );
      expect(wrapper.find('[data-test="forgotPassword-card"]').exists()).toBe(
        false
      );
      expect(wrapper.find('[data-test="isCloud-card"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="loadingToken-alert"]').exists()).toBe(
        false
      );
    });
  });
});
