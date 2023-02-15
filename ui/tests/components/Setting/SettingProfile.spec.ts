import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SettingProfile from "../../../src/components/Setting/SettingProfile.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("SettingProfile", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const username = "ShellHub";
  const emailUser = "shellhub@shellhub.com";
  const tenant = "xxxxxxxx";

  // vee-validate variables bellow
  const invalidEmails = [
    "notemail",
    "missing@dot",
    "with.only.dots",
    "r4ndomCH@r5",
  ];
  const validEmails = ["new@email.com", "another@email.org"];
  const invalidPasswords = [
    "aPasswordBiggerThanExpectedBecauseHasMoreThan30chars",
    "shor",
  ];
  const validPasswords = ["newPassword", "password123"];
  const confirmPasswordsMatchError = [
    { new: "newpass", confirmNew: "newpas" },
    { new: "Newpass", confirmNew: "newpass" },
  ];
  const confirmPasswordsMatchSuccess = [
    { new: "newpass", confirmNew: "newpass" },
    { new: "changedpassword", confirmNew: "changedpassword" },
  ];
  const compareOldNewError = [
    { old: "oldpass", new: "oldpass" },
    { old: "currentPass", new: "currentPass" },
  ];
  const compareOldNewSuccess = [
    { old: "oldpass", new: "newpass" },
    { old: "currentPass", new: "newPassword" },
  ];

  const store = createStore({
    state: {
      username,
      email: emailUser,
      tenant,
    },
    getters: {
      "auth/currentUser": (state) => state.username,
      "auth/email": (state) => state.email,
      "auth/tenant": (state) => state.tenant,
    },
    actions: {
      "users/put": vi.fn(),
    },
  });

  beforeEach(() => {
    wrapper = mount(SettingProfile, {
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

  ///////s
  // Data checking
  //////
  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });
  it("Compare data with default value", () => {
    expect(wrapper.vm.username).toEqual(username);
    expect(wrapper.vm.email).toEqual(emailUser);
    expect(wrapper.vm.currentPassword).toEqual("");
    expect(wrapper.vm.newPassword).toEqual("");
    expect(wrapper.vm.newPasswordConfirm).toEqual("");
    expect(wrapper.vm.editDataStatus).toEqual(false);
    expect(wrapper.vm.editPasswordStatus).toEqual(false);
    expect(wrapper.vm.show).toEqual(false);
    expect(wrapper.vm.showCurrentPassword).toEqual(false);
    expect(wrapper.vm.showNewPassword).toEqual(false);
    expect(wrapper.vm.showConfirmPassword).toEqual(false);
  });

  //////
  // In this case, the empty fields are validated.
  //////

  it("Show validation messages", async () => {
    wrapper.vm.username = undefined;
    wrapper.vm.email = undefined;
    wrapper.vm.currentPassword = undefined;

    await flushPromises();

    expect(wrapper.vm.usernameError).toBe("this is a required field");
    expect(wrapper.vm.emailError).toBe("this is a required field");
    expect(wrapper.vm.currentPasswordError).toBe("this is a required field");
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
  // In this case, invalid password length are validated.
  //////

  it("Show validation messages", async () => {
    wrapper.vm.newPassword = invalidPasswords[0];
    await flushPromises();
    expect(wrapper.vm.newPasswordError).toBe(
      "this must be at most 30 characters"
    );

    wrapper.vm.newPassword = invalidPasswords[1];
    await flushPromises();
    expect(wrapper.vm.newPasswordError).toBe(
      "this must be at least 5 characters"
    );

  });

   //////
  // In this case, invalid password match are validated.
  //////

  it('Show validation messages', async () => {
    // todo:  do not work
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

  //////
  // In this case, valid password length are validated.
  //////

  it('Show validation messages', async () => {
    wrapper.vm.newPassword = validPasswords[0];
    await flushPromises();
    expect(wrapper.vm.newPasswordError).toBe(undefined);

    wrapper.vm.newPassword = validPasswords[1];
    await flushPromises();
    expect(wrapper.vm.newPasswordError).toBe(undefined);
  });


  //////
  // In this case, valid password match are validated.
  //////

  it('Show validation messages', async () => {
    // todo match password not work
  });

  //////
  // In this case, valid password change.
  //////

  it('Show validation messages', async () => {
    // todo compareOldNewSuccess password not work
  });

  it('Renders the template with data', async () => {
    expect(wrapper.find('[data-test="username-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="email-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="password-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="newPassword-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="confirmNewPassword-text"]').exists()).toBeTruthy();
  });
});
