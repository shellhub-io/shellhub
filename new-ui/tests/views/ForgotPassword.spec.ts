import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import ForgotPassword from "../../src/views/ForgotPassword.vue";
import { createStore } from "vuex";
import { key } from "../../src/store";
import routes from "../../src/router";

describe("ForgotPassword", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const statusDarkMode = true;

  const invalidEmails = [
    "notemail",
    "missing@dot",
    "with.only.dots",
    "r4ndomCH@r5",
  ];
  const validEmails = ["new@email.com", "another@email.org"];

  const store = createStore({
    state: {
      statusDarkMode,
    },
    getters: {
      "layout/getStatusDarkMode": (state) => state.statusDarkMode,
    },
    actions: {
      "users/recoverPassword": vi.fn(),
      "snackbar/showSnackbarSuccessAction": vi.fn(),
      "snackbar/showSnackbarErrorAction": vi.fn(),
    },
  });

  beforeEach(async () => {
    wrapper = mount(ForgotPassword, {
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
    expect(wrapper.vm.email).toEqual("");
  });
  //////
  // In this case, the empty fields are validated.
  //////

  it("Show validation messages", async () => {
    wrapper.vm.email = undefined;

    await flushPromises();

    expect(wrapper.vm.emailError).toBe("this is a required field");
  });

  //////
  // In this case, invalid email error are validated.
  //////

  it("Show validation messages", async () => {
    wrapper.vm.email = invalidEmails[0];
    await flushPromises();
    expect(wrapper.vm.emailError).toBe("this must be a valid email");

    wrapper.vm.email = invalidEmails[1];
    await flushPromises();
    expect(wrapper.vm.emailError).toBe("this must be a valid email");

    wrapper.vm.email = invalidEmails[2];
    await flushPromises();
    expect(wrapper.vm.emailError).toBe("this must be a valid email");

    wrapper.vm.email = invalidEmails[3];
    await flushPromises();
    expect(wrapper.vm.emailError).toBe("this must be a valid email");
    wrapper.vm.email = "";
  });

  //////
  // In this case, valid email are validated.
  //////

  it('Show validation messages', async () => {
    wrapper.vm.email = validEmails[0];
    await flushPromises();
    expect(wrapper.vm.emailError).toBe(undefined);

    wrapper.vm.email = validEmails[1];
    await flushPromises();
    expect(wrapper.vm.emailError).toBe(undefined);
  });

  it('Renders the template with data', () => {
    expect(wrapper.find('[data-test="email-text"]').exists()).toBeTruthy();
  });
});
