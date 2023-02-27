import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createStore } from "vuex";
import SettingProfile from "../../../src/components/Setting/SettingProfile.vue";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("SettingProfile", () => {
  let wrapper: VueWrapper<InstanceType<typeof SettingProfile>>;
  const vuetify = createVuetify();

  const username = "ShellHub";
  const emailUser = "shellhub@shellhub.com";
  const tenant = "xxxxxxxx";

  // vee-validate variables bellow
  const invalidEmails = [
    "notemail",
    "notemail@",
    "not.em.ail@",
    "notemail@notemail.",
    "not email@notemail.com",
    "notemail@notemail.com.",
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

  ///////
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

  invalidEmails.forEach((email) => {
    it("Show validation messages", async () => {
      console.log("email", email);
      wrapper.vm.email = email;
      await flushPromises();
      expect(wrapper.vm.emailError).toBe("this must be a valid email");
    });
  });

  //////
  // In this case, invalid password length are validated.
  //////

  it("Show validation messages", async () => {
    const [most30Characters, less5Characters] = invalidPasswords;

    wrapper.vm.newPassword = most30Characters;
    await flushPromises();
    expect(wrapper.vm.newPasswordError).toBe(
      "this must be at most 30 characters",
    );

    wrapper.vm.newPassword = less5Characters;
    await flushPromises();
    expect(wrapper.vm.newPasswordError).toBe(
      "this must be at least 5 characters",
    );
  });

  //////
  // In this case, invalid password match are validated.
  //////

  confirmPasswordsMatchError.forEach((passwords) => {
    it("Show validation messages", async () => {
      const { new: newPassword, confirmNew: newPasswordConfirm } = passwords;
      wrapper.vm.newPassword = newPassword;
      wrapper.vm.newPasswordConfirm = newPasswordConfirm;
      await flushPromises();
      expect(wrapper.vm.newPasswordConfirmError).toBe(
        "Passwords do not match",
      );
    });
  });

  //////
  // In this case, valid email are validated.
  //////

  validEmails.forEach((email) => {
    it("Show validation messages", async () => {
      wrapper.vm.email = email;
      await flushPromises();
      expect(wrapper.vm.emailError).toBe(undefined);
    });
  });

  //////
  // In this case, valid password length are validated.
  //////

  validPasswords.forEach((password) => {
    it("Show validation messages", async () => {
      wrapper.vm.newPassword = password;
      await flushPromises();
      expect(wrapper.vm.newPasswordError).toBe(undefined);
    });
  });

  //////
  // In this case, valid password match are validated.
  //////

  confirmPasswordsMatchSuccess.forEach((passwords) => {
    it("Show validation messages", async () => {
      const { new: newPassword, confirmNew: newPasswordConfirm } = passwords;
      wrapper.vm.newPassword = newPassword;
      wrapper.vm.newPasswordConfirm = newPasswordConfirm;
      await flushPromises();
      expect(wrapper.vm.newPasswordConfirmError).toBe(undefined);
    });
  });

  //////
  // In this case, valid password change.
  //////

  compareOldNewSuccess.forEach((passwords) => {
    it("Show validation messages", async () => {
      const { old: currentPassword, new: newPassword } = passwords;
      wrapper.vm.currentPassword = currentPassword;
      wrapper.vm.newPassword = newPassword;
      await flushPromises();
      expect(wrapper.vm.currentPasswordError).toBe(undefined);
    });
  });

  it("Renders the template with data", async () => {
    expect(wrapper.find('[data-test="username-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="email-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="password-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="newPassword-text"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="confirmNewPassword-text"]').exists()).toBeTruthy();
  });
});
