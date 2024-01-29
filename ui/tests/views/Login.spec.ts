import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import Login from "../../src/views/Login.vue";
import { usersApi } from "@/api/http";
import { store, key } from "../../src/store";
import { router } from "../../src/router";
import { envVariables } from "../../src/envVariables";
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
  });

  it("disables submit button when the form is invalid", async () => {
    await wrapper.findComponent('[data-test="username-text"]').setValue("");
    await wrapper.findComponent('[data-test="password-text"]').setValue("");

    expect(wrapper.find('[data-test="login-btn"]').attributes().disabled).toBeDefined();
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
      mfa: {
        enable: false,
        validate: false,
      },
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
    expect(routerPushSpy).toHaveBeenCalledWith("/");
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
      mfa: {
        enable: true,
        validate: false,
      },
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

    expect(wrapper.findComponent('[data-test="invalid-login-alert"]').exists()).toBeTruthy();
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
});
