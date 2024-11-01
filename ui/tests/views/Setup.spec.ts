import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import Setup from "../../src/views/Setup.vue";
import { usersApi } from "@/api/http";
import { store, key } from "../../src/store";
import { router } from "../../src/router";
import { envVariables } from "../../src/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";

type SetupWrapper = VueWrapper<InstanceType<typeof Setup>>;

describe("Setup Account", () => {
  let wrapper: SetupWrapper;
  const vuetify = createVuetify();

  let mock: MockAdapter;

  beforeEach(() => {
    vi.useFakeTimers();

    envVariables.isCloud = false;

    mock = new MockAdapter(usersApi.getAxios());

    wrapper = mount(Setup, {
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

  it("Renders the template with data", async () => {
    expect(wrapper.find('[data-test="user-status-alert"]').exists()).toBe(false);

    expect(wrapper.find('[data-test="welcome-title"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sign-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="subtitle-1"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="sign-btn"]').exists()).toBe(true);

    wrapper.vm.el = 2;

    await nextTick();

    expect(wrapper.find('[data-test="subtitle-2"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="name-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="username-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="email-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="password-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="password-confirm-text"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privacy-policy-error"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="setup-account-btn"]').exists()).toBe(true);
  });

  it("disables submit button when the form is invalid", async () => {
    wrapper.vm.el = 2;

    await nextTick();

    expect(wrapper.find('[data-test="setup-account-btn"]').attributes().disabled).toBeDefined();
  });

  it("Calls the Create Account action when the button is clicked", async () => {
    const responseData = {
      sign: "sign",
      name: "test",
      email: "test@test.com",
      username: "test",
      password: "test123",
    };

    mock.onPost("http://localhost:3000/api/setup").reply(200, responseData);

    const signUpSpy = vi.spyOn(store, "dispatch");

    await wrapper.findComponent('[data-test="sign-text"]').setValue("sign");

    wrapper.vm.el = 2;

    await nextTick();

    await wrapper.findComponent('[data-test="name-text"]').setValue("test");
    await wrapper.findComponent('[data-test="username-text"]').setValue("test");
    await wrapper.findComponent('[data-test="email-text"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="password-text"]').setValue("test123");
    await wrapper.findComponent('[data-test="password-confirm-text"]').setValue("test123");

    await wrapper.find('[data-test="setup-account-btn"]').trigger("submit");

    vi.runOnlyPendingTimers();
    await flushPromises();
    await nextTick();
    expect(signUpSpy).toHaveBeenCalledWith("users/setup", {
      sign: "sign",
      name: "test",
      email: "test@test.com",
      username: "test",
      password: "test123",
    });
  });

  it("Handles error (400)", async () => {
    mock.onPost("http://localhost:3000/api/setup").reply(400);

    await wrapper.findComponent('[data-test="sign-text"]').setValue("sign");

    await wrapper.find('[data-test="sign-btn"]').trigger("click");

    await nextTick();

    await wrapper.findComponent('[data-test="name-text"]').setValue("test");
    await wrapper.findComponent('[data-test="username-text"]').setValue("test");
    await wrapper.findComponent('[data-test="email-text"]').setValue("test@test.com");
    await wrapper.findComponent('[data-test="password-text"]').setValue("test123");
    await wrapper.findComponent('[data-test="password-confirm-text"]').setValue("test123");

    await wrapper.find('[data-test="setup-account-btn"]').trigger("submit");
    await flushPromises();
    await nextTick();

    expect(wrapper.find('[data-test="user-status-alert"]').exists());
  });
});
