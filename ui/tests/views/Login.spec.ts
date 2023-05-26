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
    vi.useFakeTimers();

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
    vi.useRealTimers();
    vi.restoreAllMocks();
    mock.reset();
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

  it("calls the login action when the form is submitted", async () => {
    const responseData = {
      token: "fake-token",
      user: "test",
      name: "Test",
      id: "1",
      tenant: "fake-tenant",
      role: "administrator",
      email: "test@test.com",
    };

    // mock error below
    mock.onPost("http://localhost:3000/api/login").reply(200, responseData);

    const loginSpy = vi.spyOn(store, "dispatch");
    const routerPushSpy = vi.spyOn(router, "push");

    await wrapper.findComponent('[data-test="username-text"]').setValue("testuser");
    await wrapper.findComponent('[data-test="password-text"]').setValue("password");
    await wrapper.find('[data-test="login-btn"]').trigger("submit");

    vi.runOnlyPendingTimers();
    await flushPromises();

    expect(loginSpy).toHaveBeenCalledWith("auth/login", {
      username: "testuser",
      password: "password",
    });

    expect(routerPushSpy).toHaveBeenCalledWith("/");
  });

  it("shows an error message for a 401 response", async () => {
    const loginSpy = vi.spyOn(store, "dispatch");

    // mock error below
    mock.onPost("http://localhost:3000/api/login").reply(401);

    await wrapper.findComponent('[data-test="username-text"]').setValue("testuser");
    await wrapper.findComponent('[data-test="password-text"]').setValue("password");
    await wrapper.find('[data-test="login-btn"]').trigger("submit");

    vi.runOnlyPendingTimers();
    await flushPromises();

    expect(loginSpy).toHaveBeenCalledWith("auth/login", {
      username: "testuser",
      password: "password",
    });

    // Check if invalidCredentials is set to true
    expect(wrapper.vm.invalidCredentials).toBe(true);
  });

  it("redirects to ConfirmAccount route on 403 response", async () => {
    const loginSpy = vi.spyOn(store, "dispatch");
    const routerPushSpy = vi.spyOn(router, "push");

    // mock error below
    mock.onPost("http://localhost:3000/api/login").reply(403);

    await wrapper.findComponent('[data-test="username-text"]').setValue("testuser");
    await wrapper.findComponent('[data-test="password-text"]').setValue("password");
    await wrapper.find('[data-test="login-btn"]').trigger("submit");

    vi.runOnlyPendingTimers();
    await flushPromises();

    // Assert the login action dispatch
    expect(loginSpy).toHaveBeenCalledWith("auth/login", {
      username: "testuser",
      password: "password",
    });

    // Assert the redirection
    expect(routerPushSpy).toHaveBeenCalledWith({
      name: "ConfirmAccount",
      query: { username: "testuser" },
    });
  });
});
