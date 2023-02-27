import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import ForgotPassword from "../../src/views/ForgotPassword.vue";
import { key } from "../../src/store";
import routes from "../../src/router";

describe("ForgotPassword", () => {
  let wrapper: VueWrapper<InstanceType<typeof ForgotPassword>>;
  const vuetify = createVuetify();

  const statusDarkMode = true;

  const invalidInputs = [
    "an",
    "invalidChar(/",
    "with space",
    `moreThan255charactersmoreThan255charactersmoreThan255charactersmoreThan255characters
    moreThan255charactersmoreThan255charactersmoreThan255charactersmoreThan255characters
    moreThan255charactersmoreThan255charactersmoreThan255charactersmoreThan255charactermoreT`,
  ];

  const referentError = [
    "this must be at least 3 characters",
    "The field only accepts the special characters _, ., - and @.",
    "The field cannot contain white spaces.",
    "this must be at most 255 characters",
  ];

  const validInputs = ["new@email.com", "another@email.org", "shellhub", "myUser"];

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
    expect(wrapper.vm.account).toEqual("");
    expect(wrapper.vm.accountError).toEqual(undefined);
  });
  //////
  // In this case, the empty field are validated.
  //////

  it("Show validation messages", async () => {
    wrapper.vm.account = undefined;

    await flushPromises();

    expect(wrapper.vm.accountError).toBe("this is a required field");
  });

  //////
  // In this case, test invalid inputs.
  //////

  invalidInputs.forEach((input, index) => {
    it("Show validation messages", async () => {
      wrapper.vm.account = input;
      await flushPromises();
      expect(wrapper.vm.accountError).toBe(referentError[index]);
    });
  });

  //////
  // In this case, test valid inputs.
  //////

  validInputs.forEach((input) => {
    it("Show validation messages", async () => {
      wrapper.vm.account = input;
      await flushPromises();
      expect(wrapper.vm.accountError).toBe(undefined);
    });
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="account-text"]').exists()).toBeTruthy();
  });
});
