import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import ForgotPassword from "@/views/ForgotPassword.vue";
import { usersApi } from "@/api/http";
import { store, key } from "../../src/store";
import { router } from "../../src/router";
import { envVariables } from "../../src/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type ForgotPasswordWrapper = VueWrapper<InstanceType<typeof ForgotPassword>>;

describe("Forgot Password", () => {
  let wrapper: ForgotPasswordWrapper;
  const vuetify = createVuetify();

  let mock: MockAdapter;

  beforeEach(() => {
    vi.useFakeTimers();

    envVariables.isCloud = true;

    // Create a mock adapter for the usersApi instance
    mock = new MockAdapter(usersApi.getAxios());

    wrapper = mount(ForgotPassword, {
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
    expect(wrapper.find('[data-test="account-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="title-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="body-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="forgotPassword-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="isCloud-card"]').exists()).toBe(true);
  });

  it("Calls the Forgot Password action when the button is clicked", async () => {
    mock.onPost("http://localhost:3000/api/user/recover_password").reply(200);

    const forgotPasswordSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="account-text"]').setValue("testuser");
    await wrapper.find('[data-test="forgotPassword-btn"]').trigger("submit");

    vi.runOnlyPendingTimers();
    await flushPromises();

    expect(forgotPasswordSpy).toHaveBeenCalledWith("users/recoverPassword", "testuser");
  });

  it("Displays success message on successful email submission", async () => {
    mock.onPost("http://localhost:3000/api/user/recover_password").reply(200);

    await wrapper.findComponent('[data-test="account-text"]').setValue("testuser");
    await wrapper.find('[data-test="forgotPassword-btn"]').trigger("submit");

    vi.runOnlyPendingTimers();
    await flushPromises();

    expect(wrapper.find('[data-test="success-text"]').exists()).toBe(true);
  });
});
