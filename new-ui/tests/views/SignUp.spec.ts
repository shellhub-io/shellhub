import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SignUp from "../../src/views/SignUp.vue";
import { createStore } from "vuex";
import { key } from "../../src/store";
import routes from "../../src/router";

describe("SignUp", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const newUser = {
    name: "",
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
  };

  const statusDarkMode = true;

  const store = createStore({
    state: {
      statusDarkMode,
    },
    getters: {
      "layout/getStatusDarkMode": (state) => state.statusDarkMode,
    },
    actions: {
      "users/signUp": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });

  beforeEach(async () => {
    wrapper = mount(SignUp, {
      global: {
        plugins: [[store, key], vuetify, routes],
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
  it("Compare data with default value", () => {
    expect(wrapper.vm.name).toEqual(newUser.name);
    expect(wrapper.vm.username).toEqual(newUser.username);
    expect(wrapper.vm.email).toEqual(newUser.email);
    expect(wrapper.vm.password).toEqual(newUser.password);
    // expect(wrapper.vm.confirmPassword).toEqual(newUser.confirmPassword);
    expect(wrapper.vm.overlay).toEqual(false);
    expect(wrapper.vm.delay).toEqual(500);
  });

  //////
  // HTML validation
  //////

  it("Show empty fields required in validation", async () => {
    wrapper.vm.name = undefined;
    wrapper.vm.username = undefined;
    wrapper.vm.email = undefined;
    wrapper.vm.password = undefined;

    await flushPromises();

    expect(wrapper.vm.nameError).toEqual("this is a required field");
    expect(wrapper.vm.usernameError).toEqual("this is a required field");
    expect(wrapper.vm.emailError).toEqual("this is a required field");
    expect(wrapper.vm.passwordError).toEqual("this is a required field");

    wrapper.vm.name = "shellhub";
    wrapper.vm.username = "shellhub";
    wrapper.vm.password = "12";

    await flushPromises();

    expect(wrapper.vm.passwordError).toBe(
      "Your password should be 5-30 characters long"
    );

    wrapper.vm.password = "123456789123456789123456789123456789123456789";

    await flushPromises();

    expect(wrapper.vm.passwordError).toBe(
      "Your password should be 5-30 characters long"
    );
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="name-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="username-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="email-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="password-text"]').exists()).toBeTruthy();
    // todo expect(wrapper.find('[data-test="confirmPassword-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="login-btn"]').exists()).toBeTruthy();
  });
});
