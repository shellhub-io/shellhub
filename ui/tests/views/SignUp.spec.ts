import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import SignUp from "../../src/views/SignUp.vue";
import { usersApi } from "@/api/http";
import { store, key } from "../../src/store";
import { router } from "../../src/router";
import { envVariables } from "../../src/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type SignUpWrapper = VueWrapper<InstanceType<typeof SignUp>>;

describe("Login", () => {
  let wrapper: SignUpWrapper;
  const vuetify = createVuetify();

  let mock: MockAdapter;

  beforeEach(() => {
    vi.useFakeTimers();

    envVariables.isCloud = true;

    // Create a mock adapter for the usersApi instance
    mock = new MockAdapter(usersApi.getAxios());

    wrapper = mount(SignUp, {
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
    expect(wrapper.find('[data-test="name-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="username-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="email-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="password-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="password-confirm-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="accept-privacy-policy-checkbox"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="accept-news-checkbox"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privacy-policy-error"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="create-account-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="login-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="accountCreated-component"]').exists()).toBe(false);
  });

  it("disables submit button when the form is invalid", async () => {
    await wrapper.findComponent('[data-test="accept-privacy-policy-checkbox"]').setValue(false);
    await nextTick();

    expect(wrapper.find('[data-test="create-account-btn"]').attributes().disabled).toBeDefined();
  });

  it("Calls the Create Account action when the button is clicked", async () => {
    const responseData = {
      name: "test",
      email: "test@test.com",
      username: "test",
      password: "test",
      emailMarketing: true,
    };

    mock.onPost("http://localhost:3000/api/register").reply(200, responseData);

    const signUpSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="name-text"]').setValue("test");
    await wrapper.findComponent('[data-test="username-text"]').setValue("test");
    await wrapper.findComponent('[data-test="email-text"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="password-text"]').setValue("test");
    await wrapper.findComponent('[data-test="password-confirm-text"]').setValue("test");
    await wrapper.findComponent('[data-test="accept-privacy-policy-checkbox"]').setValue(true);
    await wrapper.findComponent('[data-test="accept-news-checkbox"]').setValue(true);

    await wrapper.find('[data-test="create-account-btn"]').trigger("submit");

    vi.runOnlyPendingTimers();
    await flushPromises();

    expect(signUpSpy).toHaveBeenCalledWith("users/signUp", {
      name: "test",
      email: "test@test.com",
      username: "test",
      password: "test",
      confirmPassword: "test",
      emailMarketing: true,
    });
  });

  it("Handles a 400 Axios error", async () => {
    const responseData = {
      axiosError: {
        response: {
          data: [
            "username",
          ],
        },
      },
    };
    mock.onPost("http://localhost:3000/api/register").reply(400, responseData);

    await wrapper.findComponent('[data-test="name-text"]').setValue("test");
    await wrapper.findComponent('[data-test="username-text"]').setValue("test");
    await wrapper.findComponent('[data-test="email-text"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="password-text"]').setValue("test");
    await wrapper.findComponent('[data-test="password-confirm-text"]').setValue("test");
    await wrapper.findComponent('[data-test="accept-privacy-policy-checkbox"]').setValue(true);
    await wrapper.findComponent('[data-test="accept-news-checkbox"]').setValue(true);

    await wrapper.find('[data-test="create-account-btn"]').trigger("submit");

    vi.runOnlyPendingTimers();
    await flushPromises();
  });

  it("Handles a 409 Axios error", async () => {
    const responseData = {
      axiosError: {
        response: {
          data: [
            "username",
            "password",
          ],
        },
      },
    };
    mock.onPost("http://localhost:3000/api/register").reply(409, responseData);

    await wrapper.findComponent('[data-test="name-text"]').setValue("test");
    await wrapper.findComponent('[data-test="username-text"]').setValue("test");
    await wrapper.findComponent('[data-test="email-text"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="password-text"]').setValue("test");
    await wrapper.findComponent('[data-test="password-confirm-text"]').setValue("test");
    await wrapper.findComponent('[data-test="accept-privacy-policy-checkbox"]').setValue(true);
    await wrapper.findComponent('[data-test="accept-news-checkbox"]').setValue(true);

    await wrapper.find('[data-test="create-account-btn"]').trigger("submit");

    vi.runOnlyPendingTimers();
    await flushPromises();
  });
});
