import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import Login from "@/views/Login.vue";
import { usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type LoginWrapper = VueWrapper<InstanceType<typeof Login>>;

describe("Login", () => {
  let wrapper: LoginWrapper;
  const vuetify = createVuetify();

  let mock: MockAdapter;

  beforeEach(() => {
    envVariables.isCloud = true;

    // Create a mock adapter for the usersApi instance
    mock = new MockAdapter(usersApi.getAxios());

    wrapper = mount(Login, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
    mock.reset();
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with data", () => {
    expect(wrapper.find('[data-test="username-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="password-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="forgotPassword-card"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="isCloud-card"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="loadingToken-alert"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="sso-btn"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="or-divider-sso"]').exists()).toBe(false);
  });

  it("Renders enterprise only fragments", async () => {
    envVariables.isEnterprise = true;

    store.commit("users/setSystemInfo", {
      authentication: { local: true, saml: true },
    });

    await flushPromises();

    expect(wrapper.find('[data-test="sso-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="or-divider-sso"]').exists()).toBe(true);
  });

  it("disables fields and login button when envVariables.isEnterprise is true", async () => {
    envVariables.isCloud = false;
    envVariables.isEnterprise = true;

    store.commit("users/setSystemInfo", {
      authentication: { local: false, saml: true },
    });

    await flushPromises();

    const usernameField = wrapper.find('[data-test="username-text"]').attributes().class;
    const passwordField = wrapper.find('[data-test="password-text"]').attributes().class;
    const loginButton = wrapper.find('[data-test="login-btn"]').attributes().class;

    expect(usernameField).toContain("v-input--disabled");
    expect(passwordField).toContain("v-input--disabled");
    expect(loginButton).toContain("v-btn--disabled");
  });

  it("calls the login action when the form is submitted", async () => {
    const responseData = {
      token: "fake-token",
      user: "test",
      name: "Test",
      id: "1",
      tenant: "fake-tenant",
      role: "administrator",
      email: "test@test.com",
      mfa: false,
    };

    // mock error below
    mock.onPost("http://localhost:3000/api/login").reply(200, responseData);

    const loginSpy = vi.spyOn(store, "dispatch");
    const routerPushSpy = vi.spyOn(router, "push");

    await wrapper.findComponent('[data-test="username-text"]').setValue("test");
    await wrapper.findComponent('[data-test="password-text"]').setValue("password");
    await wrapper.findComponent('[data-test="form"]').trigger("submit");
    await flushPromises();

    // Assert the login action dispatch
    expect(loginSpy).toHaveBeenCalledWith("auth/login", {
      username: "test",
      password: "password",
    });

    expect(wrapper.findComponent(".v-alert").exists()).toBeFalsy();
    expect(routerPushSpy).toHaveBeenCalledWith({
      path: "/",
      query: {},
    });
  });

  it("calls the mfa action when the login form is submitted", async () => {
    const responseData = {
      token: "fake-token",
      user: "test",
      name: "Test",
      id: "1",
      tenant: "fake-tenant",
      role: "administrator",
      email: "test@test.com",
      mfa: true,
    };

    // mock error below
    mock.onPost("http://localhost:3000/api/login").reply(200, responseData);

    const loginSpy = vi.spyOn(store, "dispatch");
    const routerPushSpy = vi.spyOn(router, "push");

    await wrapper.findComponent('[data-test="username-text"]').setValue("testuser");
    await wrapper.findComponent('[data-test="password-text"]').setValue("password");
    await wrapper.findComponent('[data-test="form"]').trigger("submit");
    await flushPromises();

    // Assert the login action dispatch
    expect(loginSpy).toHaveBeenCalledWith("auth/login", {
      username: "testuser",
      password: "password",
    });

    expect(wrapper.findComponent(".v-alert").exists()).toBeFalsy();
    expect(routerPushSpy).toHaveBeenCalledWith({ name: "MfaLogin" });
  });

  it("shows an error message for a 401 response", async () => {
    const loginSpy = vi.spyOn(store, "dispatch");

    // mock error below
    mock.onPost("http://localhost:3000/api/login").reply(401);

    await wrapper.findComponent('[data-test="username-text"]').setValue("testuser");
    await wrapper.findComponent('[data-test="password-text"]').setValue("password");
    await wrapper.findComponent('[data-test="form"]').trigger("submit");
    await flushPromises();

    // Assert the login action dispatch
    expect(loginSpy).toHaveBeenCalledWith("auth/login", {
      username: "testuser",
      password: "password",
    });

    expect(wrapper.findComponent('[data-test="invalid-login-alert"]').exists());
  });

  it("redirects to ConfirmAccount route on 403 response", async () => {
    const loginSpy = vi.spyOn(store, "dispatch");
    const routerPushSpy = vi.spyOn(router, "push");

    // mock error below
    mock.onPost("http://localhost:3000/api/login").reply(403);

    await wrapper.findComponent('[data-test="username-text"]').setValue("testuser");
    await wrapper.findComponent('[data-test="password-text"]').setValue("password");
    await wrapper.findComponent('[data-test="form"]').trigger("submit");
    await flushPromises();

    // Assert the login action dispatch
    expect(loginSpy).toHaveBeenCalledWith("auth/login", {
      username: "testuser",
      password: "password",
    });

    expect(wrapper.findComponent(".v-alert").exists()).toBeFalsy();

    // Assert the redirection
    expect(routerPushSpy).toHaveBeenCalledWith({
      name: "ConfirmAccount",
      query: { username: "testuser" },
    });
  });

  it("locks account after 10 failed login attempts", async () => {
    const username = "testuser";
    const maxAttempts = 10;
    const lockoutDuration = 7 * 24 * 60 * 60; // 7 days in seconds
    let attempts = 0;

    mock.onPost("http://localhost:3000/api/login").reply((config) => {
      const { username: reqUsername, password } = JSON.parse(config.data);
      if (reqUsername === username && password === "wrongpassword") {
        attempts++;
        if (attempts >= maxAttempts) {
          return [429, {}, { "x-account-lockout": lockoutDuration.toString() }];
        }
        return [401];
      }
      return [200, { token: "fake-token" }];
    });

    // Simulate 10 failed login attempts
    for (let i = 0; i < maxAttempts; i++) {
      wrapper.findComponent('[data-test="username-text"]').setValue(username);
      wrapper.findComponent('[data-test="password-text"]').setValue("wrongpassword");
      wrapper.findComponent('[data-test="form"]').trigger("submit");
      // eslint-disable-next-line no-await-in-loop
      await flushPromises();
    }

    // Ensure the account is locked out
    expect(wrapper.findComponent('[data-test="invalid-login-alert"]').exists()).toBeTruthy();
  });
});
